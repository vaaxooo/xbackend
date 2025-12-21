package bootstrap

import (
	"context"
	"database/sql"
	"time"

	usersapp "github.com/vaaxooo/xbackend/internal/modules/users/application"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/common"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/link"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/login"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/profile"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/refresh"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/register"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/telegram"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/twofactor"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/verification"
	usersauth "github.com/vaaxooo/xbackend/internal/modules/users/infrastructure/auth"
	userscrypto "github.com/vaaxooo/xbackend/internal/modules/users/infrastructure/crypto"
	usersevents "github.com/vaaxooo/xbackend/internal/modules/users/infrastructure/events"
	"github.com/vaaxooo/xbackend/internal/modules/users/public"
	pdb "github.com/vaaxooo/xbackend/internal/platform/db"
	usersdb "github.com/vaaxooo/xbackend/internal/platform/db/users"
)

// Module exposes the Users bounded context public API.
// Keeping it under bootstrap package simplifies lifting the context
// into a standalone microservice without reworking imports.
type Module struct {
	Service usersapp.Service
	Auth    public.AuthPort
	Outbox  *usersevents.OutboxRepository
}

// Dependencies describes technical components required to assemble
// the Users bounded context. Only cross-cutting infrastructure goes here.
type Dependencies struct {
	DB *sql.DB
}

// Init wires all application services and adapters for the Users context.
// It can be reused by a standalone service binary, keeping the composition
// root close to the bounded context itself.
func Init(deps Dependencies, cfg public.Config) (*Module, error) {
	usersRepo := usersdb.NewUserRepo(deps.DB)
	identityRepo := usersdb.NewIdentityRepo(deps.DB)
	refreshRepo := usersdb.NewRefreshRepo(deps.DB)
	tokenRepo := usersdb.NewVerificationTokenRepo(deps.DB)
	outboxRepo := usersevents.NewOutboxRepository(deps.DB)
	uow := pdb.NewUnitOfWork(deps.DB)

	hasher := userscrypto.NewBcryptHasher(0)

	authPort, err := usersauth.NewJWTAuth(cfg.Auth.JWTSecret)
	if err != nil {
		return nil, err
	}

	eventPublisher := usersevents.NewOutboxPublisher(outboxRepo)

	registerUC := common.NewTransactionalUseCase(uow, register.New(usersRepo, identityRepo, refreshRepo, tokenRepo, hasher, authPort, eventPublisher, cfg.Auth.AccessTTL, cfg.Auth.RefreshTTL, cfg.Auth.VerificationTTL, cfg.Auth.RequireEmailConfirmation))
	loginUC := common.NewTransactionalUseCase(uow, login.New(usersRepo, identityRepo, refreshRepo, hasher, authPort, cfg.Auth.AccessTTL, cfg.Auth.RefreshTTL, cfg.Auth.RequireEmailConfirmation))
	telegramUC, err := telegram.New(usersRepo, identityRepo, refreshRepo, authPort, cfg.Telegram.BotToken, cfg.Auth.AccessTTL, cfg.Auth.RefreshTTL, cfg.Telegram.InitDataTTL)
	if err != nil {
		return nil, err
	}
	telegramTransactional := common.NewTransactionalUseCase(uow, telegramUC)
	refreshUC := common.NewTransactionalUseCase(uow, refresh.New(refreshRepo, authPort, cfg.Auth.AccessTTL, cfg.Auth.RefreshTTL))

	confirmEmailUC := common.NewTransactionalUseCase(uow, verification.NewConfirmEmailUseCase(usersRepo, identityRepo, tokenRepo, refreshRepo, authPort, cfg.Auth.AccessTTL, cfg.Auth.RefreshTTL))

	requestVerification := verification.NewRequestUseCase(identityRepo, tokenRepo, eventPublisher, cfg.Auth.VerificationTTL, cfg.Auth.PasswordResetTTL, time.Minute)
	emailVerificationUC := common.NewTransactionalUseCase(uow, funcUseCase[verification.RequestEmailInput, struct{}]{
		fn: func(ctx context.Context, cmd verification.RequestEmailInput) (struct{}, error) {
			return struct{}{}, requestVerification.RequestEmailConfirmation(ctx, cmd)
		},
	})
	passwordResetRequestUC := common.NewTransactionalUseCase(uow, funcUseCase[verification.RequestPasswordResetInput, struct{}]{
		fn: func(ctx context.Context, cmd verification.RequestPasswordResetInput) (struct{}, error) {
			return struct{}{}, requestVerification.RequestPasswordReset(ctx, cmd)
		},
	})
	resetPasswordUC := common.NewTransactionalUseCase(uow, verification.NewResetPasswordUseCase(identityRepo, tokenRepo, hasher))
	twoFactorUC := newTwoFactorHandlers(twofactor.NewUseCase(identityRepo, cfg.Auth.TwoFactorIssuer), uow)

	meUC := common.NewTransactionalUseCase(uow, profile.NewGet(usersRepo))
	profileUC := common.NewTransactionalUseCase(uow, profile.NewUpdate(usersRepo))
	linkUC := link.New(identityRepo)

	svc := usersapp.NewService(
		common.UseCaseHandler(registerUC),
		common.UseCaseHandler(loginUC),
		common.UseCaseHandler(telegramTransactional),
		common.UseCaseHandler(refreshUC),
		common.UseCaseHandler(confirmEmailUC),
		common.UseCaseHandler(emailVerificationUC),
		common.UseCaseHandler(passwordResetRequestUC),
		common.UseCaseHandler(resetPasswordUC),
		twoFactorUC.setup,
		twoFactorUC.confirm,
		twoFactorUC.disable,
		common.UseCaseHandler(meUC),
		common.UseCaseHandler(profileUC),
		common.UseCaseHandler(linkUC),
	)

	return &Module{
		Service: svc,
		Auth:    authPort,
		Outbox:  outboxRepo,
	}, nil
}

type funcUseCase[Cmd any, Resp any] struct {
	fn func(context.Context, Cmd) (Resp, error)
}

func (f funcUseCase[Cmd, Resp]) Execute(ctx context.Context, cmd Cmd) (Resp, error) {
	return f.fn(ctx, cmd)
}

type twoFactorHandlers struct {
	setup   common.Handler[twofactor.SetupInput, twofactor.SetupOutput]
	confirm common.Handler[twofactor.ConfirmInput, struct{}]
	disable common.Handler[twofactor.DisableInput, struct{}]
}

func newTwoFactorHandlers(uc *twofactor.UseCase, uow common.UnitOfWork) twoFactorHandlers {
	setup := common.NewTransactionalUseCase(uow, funcUseCase[twofactor.SetupInput, twofactor.SetupOutput]{
		fn: uc.Setup,
	})
	confirm := common.NewTransactionalUseCase(uow, funcUseCase[twofactor.ConfirmInput, struct{}]{
		fn: func(ctx context.Context, in twofactor.ConfirmInput) (struct{}, error) {
			return struct{}{}, uc.Confirm(ctx, in)
		},
	})
	disable := common.NewTransactionalUseCase(uow, funcUseCase[twofactor.DisableInput, struct{}]{
		fn: func(ctx context.Context, in twofactor.DisableInput) (struct{}, error) {
			return struct{}{}, uc.Disable(ctx, in)
		},
	})

	return twoFactorHandlers{
		setup:   common.UseCaseHandler(setup),
		confirm: common.UseCaseHandler(confirm),
		disable: common.UseCaseHandler(disable),
	}
}

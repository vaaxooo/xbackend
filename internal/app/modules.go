package app

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
	plog "github.com/vaaxooo/xbackend/internal/platform/log"
)

// Modules is a registry of all initialized bounded contexts (modules).
// Each module exposes a public facade (service) used by transports.
type Modules struct {
	Users *UsersModule
	// Billing *BillingModule
}

// ModuleDeps contains shared infrastructure dependencies that modules may need.
// Keep it small: only cross-cutting technical stuff.
type ModuleDeps struct {
	DB     *sql.DB
	Logger plog.Logger
}

// UsersModule exposes the Users bounded context public API.
type UsersModule struct {
	Service usersapp.Service
	Auth    public.AuthPort
	Outbox  *usersevents.OutboxRepository
}

type ModulesConfig struct {
	Users public.Config
}

func InitModules(deps ModuleDeps, cfg ModulesConfig) (*Modules, error) {
	users, err := initUsersModule(deps, cfg.Users)
	if err != nil {
		return nil, err
	}

	return &Modules{
		Users: users,
	}, nil
}

func initUsersModule(deps ModuleDeps, cfg public.Config) (*UsersModule, error) {
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

	return &UsersModule{
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

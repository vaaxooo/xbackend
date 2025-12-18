package app

import (
	"database/sql"

	usersapp "github.com/vaaxooo/xbackend/internal/modules/users/application"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/link"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/login"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/profile"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/refresh"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/register"
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
	outboxRepo := usersevents.NewOutboxRepository(deps.DB)
	uow := pdb.NewUnitOfWork(deps.DB)

	hasher := userscrypto.NewBcryptHasher(0)

	authPort, err := usersauth.NewJWTAuth(cfg.Auth.JWTSecret)
	if err != nil {
		return nil, err
	}

	eventPublisher := usersevents.NewOutboxPublisher(outboxRepo)

	registerUC := register.New(uow, usersRepo, identityRepo, refreshRepo, hasher, authPort, eventPublisher, cfg.Auth.AccessTTL, cfg.Auth.RefreshTTL)
	loginUC := login.New(uow, usersRepo, identityRepo, refreshRepo, hasher, authPort, cfg.Auth.AccessTTL, cfg.Auth.RefreshTTL)
	refreshUC := refresh.New(uow, refreshRepo, authPort, cfg.Auth.AccessTTL, cfg.Auth.RefreshTTL)

	meUC := profile.NewGet(usersRepo)
	profileUC := profile.NewUpdate(usersRepo)
	linkUC := link.New(identityRepo)

	svc := usersapp.NewService(registerUC, loginUC, refreshUC, meUC, profileUC, linkUC)

	return &UsersModule{
		Service: svc,
		Auth:    authPort,
		Outbox:  outboxRepo,
	}, nil
}

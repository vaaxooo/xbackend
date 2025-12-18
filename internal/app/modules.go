package app

import (
	"database/sql"
	"time"

	usersapp "github.com/vaaxooo/xbackend/internal/modules/users/application"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/link"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/login"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/profile"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/refresh"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/register"
	userscrypto "github.com/vaaxooo/xbackend/internal/modules/users/infrastructure/crypto"
	userstokens "github.com/vaaxooo/xbackend/internal/modules/users/infrastructure/tokens"
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

	// Auth config for Users module.
	AuthJWTSecret  string
	AuthAccessTTL  time.Duration
	AuthRefreshTTL time.Duration
}

// UsersModule exposes the Users bounded context public API.
type UsersModule struct {
	Service usersapp.Service
	Tokens  interface {
		Parse(token string) (string, error)
	}
}

func InitModules(deps ModuleDeps) (*Modules, error) {
	users, err := initUsersModule(deps)
	if err != nil {
		return nil, err
	}

	return &Modules{
		Users: users,
	}, nil
}

func initUsersModule(deps ModuleDeps) (*UsersModule, error) {
	usersRepo := usersdb.NewUserRepo(deps.DB)
	identityRepo := usersdb.NewIdentityRepo(deps.DB)
	refreshRepo := usersdb.NewRefreshRepo(deps.DB)
	tx := pdb.NewTransactor(deps.DB)

	hasher := userscrypto.NewBcryptHasher(0)

	tok, err := userstokens.NewHS256(deps.AuthJWTSecret)
	if err != nil {
		return nil, err
	}

	registerUC := register.New(tx, usersRepo, identityRepo, refreshRepo, hasher, tok, deps.AuthAccessTTL, deps.AuthRefreshTTL)
	loginUC := login.New(tx, usersRepo, identityRepo, refreshRepo, hasher, tok, deps.AuthAccessTTL, deps.AuthRefreshTTL)
	refreshUC := refresh.New(tx, refreshRepo, tok, deps.AuthAccessTTL, deps.AuthRefreshTTL)

	meUC := profile.NewGet(usersRepo)
	profileUC := profile.NewUpdate(usersRepo)
	linkUC := link.New(identityRepo)

	svc := usersapp.NewService(registerUC, loginUC, refreshUC, meUC, profileUC, linkUC)

	return &UsersModule{
		Service: svc,
		Tokens:  tok,
	}, nil
}

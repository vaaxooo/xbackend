package app

import (
	"database/sql"
	"time"

	usersapp "github.com/vaaxooo/xbackend/internal/modules/users/app"
	userswiring "github.com/vaaxooo/xbackend/internal/modules/users/wiring"

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
	Tokens  userswiring.TokenParser
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
	built, err := userswiring.Build(userswiring.Deps{
		DB:         deps.DB,
		JWTSecret:  deps.AuthJWTSecret,
		AccessTTL:  deps.AuthAccessTTL,
		RefreshTTL: deps.AuthRefreshTTL,
	})
	if err != nil {
		return nil, err
	}

	return &UsersModule{
		Service: built.Service,
		Tokens:  built.Tokens,
	}, nil
}

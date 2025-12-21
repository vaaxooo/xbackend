package app

import (
	"database/sql"

	usersbootstrap "github.com/vaaxooo/xbackend/internal/modules/users/bootstrap"
	"github.com/vaaxooo/xbackend/internal/modules/users/public"
	plog "github.com/vaaxooo/xbackend/internal/platform/log"
)

// Modules is a registry of all initialized bounded contexts (modules).
// Each module exposes a public facade (service) used by transports.
type Modules struct {
	Users *usersbootstrap.Module
	// Billing *BillingModule
}

// ModuleDeps contains shared infrastructure dependencies that modules may need.
// Keep it small: only cross-cutting technical stuff.
type ModuleDeps struct {
	DB     *sql.DB
	Logger plog.Logger
}

type ModulesConfig struct {
	Users public.Config
}

func InitModules(deps ModuleDeps, cfg ModulesConfig) (*Modules, error) {
	users, err := usersbootstrap.Init(usersbootstrap.Dependencies{DB: deps.DB}, cfg.Users)
	if err != nil {
		return nil, err
	}

	return &Modules{
		Users: users,
	}, nil
}

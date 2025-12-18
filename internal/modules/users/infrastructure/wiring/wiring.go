package wiring

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
	userspg "github.com/vaaxooo/xbackend/internal/modules/users/infrastructure/postgres"
	userstokens "github.com/vaaxooo/xbackend/internal/modules/users/infrastructure/tokens"
	pdb "github.com/vaaxooo/xbackend/internal/platform/db"
)

type Deps struct {
	DB         *sql.DB
	JWTSecret  string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

// TokenParser is what HTTP middleware needs.
type TokenParser interface {
	Parse(token string) (string, error)
}

type Built struct {
	Service usersapp.Service
	Tokens  TokenParser
}

func Build(deps Deps) (*Built, error) {
	usersRepo := userspg.NewUserRepo(deps.DB)
	identityRepo := userspg.NewIdentityRepo(deps.DB)
	refreshRepo := userspg.NewRefreshRepo(deps.DB)
	tx := pdb.NewTransactor(deps.DB)

	hasher := userscrypto.NewBcryptHasher(0)

	tok, err := userstokens.NewHS256(deps.JWTSecret)
	if err != nil {
		return nil, err
	}

	registerUC := register.New(tx, usersRepo, identityRepo, refreshRepo, hasher, tok, deps.AccessTTL, deps.RefreshTTL)
	loginUC := login.New(tx, usersRepo, identityRepo, refreshRepo, hasher, tok, deps.AccessTTL, deps.RefreshTTL)
	refreshUC := refresh.New(tx, refreshRepo, tok, deps.AccessTTL, deps.RefreshTTL)

	meUC := profile.NewGet(usersRepo)
	profileUC := profile.NewUpdate(usersRepo)
	linkUC := link.New(identityRepo)

	svc := usersapp.NewService(registerUC, loginUC, refreshUC, meUC, profileUC, linkUC)

	return &Built{
		Service: svc,
		Tokens:  tok,
	}, nil
}

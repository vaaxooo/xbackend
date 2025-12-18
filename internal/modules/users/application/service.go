package application

import (
	"context"

	"github.com/vaaxooo/xbackend/internal/modules/users/application/link"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/login"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/profile"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/refresh"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/register"
)

type Service interface {
	Register(ctx context.Context, in register.Input) (login.Output, error)
	Login(ctx context.Context, in login.Input) (login.Output, error)
	Refresh(ctx context.Context, in refresh.Input) (refresh.Output, error)

	GetMe(ctx context.Context, in profile.GetInput) (profile.Output, error)
	UpdateProfile(ctx context.Context, in profile.UpdateInput) (profile.Output, error)
	LinkProvider(ctx context.Context, in link.Input) (link.Output, error)
}

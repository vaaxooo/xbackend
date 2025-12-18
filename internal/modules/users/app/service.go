package app

import (
	"context"

	"github.com/vaaxooo/xbackend/internal/modules/users/app/link"
	"github.com/vaaxooo/xbackend/internal/modules/users/app/login"
	"github.com/vaaxooo/xbackend/internal/modules/users/app/profile"
	"github.com/vaaxooo/xbackend/internal/modules/users/app/refresh"
	"github.com/vaaxooo/xbackend/internal/modules/users/app/register"
)

type Service interface {
	Register(ctx context.Context, in register.Input) (login.Output, error)
	Login(ctx context.Context, in login.Input) (login.Output, error)
	Refresh(ctx context.Context, in refresh.Input) (refresh.Output, error)

	GetMe(ctx context.Context, in profile.GetInput) (profile.Output, error)
	UpdateProfile(ctx context.Context, in profile.UpdateInput) (profile.Output, error)
	LinkProvider(ctx context.Context, in link.Input) (link.Output, error)
}

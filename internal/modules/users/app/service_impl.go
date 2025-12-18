package app

import (
	"context"

	"github.com/vaaxooo/xbackend/internal/modules/users/app/link"
	"github.com/vaaxooo/xbackend/internal/modules/users/app/login"
	"github.com/vaaxooo/xbackend/internal/modules/users/app/profile"
	"github.com/vaaxooo/xbackend/internal/modules/users/app/refresh"
	"github.com/vaaxooo/xbackend/internal/modules/users/app/register"
)

type service struct {
	registerUC *register.UseCase
	loginUC    *login.UseCase
	refreshUC  *refresh.UseCase

	meUC      *profile.GetUseCase
	profileUC *profile.UpdateUseCase
	linkUC    *link.UseCase
}

func NewService(
	registerUC *register.UseCase,
	loginUC *login.UseCase,
	refreshUC *refresh.UseCase,
	meUC *profile.GetUseCase,
	profileUC *profile.UpdateUseCase,
	linkUC *link.UseCase,
) Service {
	return &service{
		registerUC: registerUC,
		loginUC:    loginUC,
		refreshUC:  refreshUC,
		meUC:       meUC,
		profileUC:  profileUC,
		linkUC:     linkUC,
	}
}

func (s *service) Register(ctx context.Context, in register.Input) (login.Output, error) {
	return s.registerUC.Execute(ctx, in)
}

func (s *service) Login(ctx context.Context, in login.Input) (login.Output, error) {
	return s.loginUC.Execute(ctx, in)
}

func (s *service) Refresh(ctx context.Context, in refresh.Input) (refresh.Output, error) {
	return s.refreshUC.Execute(ctx, in)
}

func (s *service) GetMe(ctx context.Context, in profile.GetInput) (profile.Output, error) {
	return s.meUC.Execute(ctx, in)
}

func (s *service) UpdateProfile(ctx context.Context, in profile.UpdateInput) (profile.Output, error) {
	return s.profileUC.Execute(ctx, in)
}

func (s *service) LinkProvider(ctx context.Context, in link.Input) (link.Output, error) {
	return s.linkUC.Execute(ctx, in)
}

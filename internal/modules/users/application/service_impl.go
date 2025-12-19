package application

import (
	"context"

	"github.com/vaaxooo/xbackend/internal/modules/users/application/common"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/link"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/login"
        "github.com/vaaxooo/xbackend/internal/modules/users/application/profile"
        "github.com/vaaxooo/xbackend/internal/modules/users/application/refresh"
        "github.com/vaaxooo/xbackend/internal/modules/users/application/register"
        "github.com/vaaxooo/xbackend/internal/modules/users/application/telegram"
)

type service struct {
        registerUC common.Handler[register.Input, login.Output]
        loginUC    common.Handler[login.Input, login.Output]
        telegramUC common.Handler[telegram.Input, login.Output]
        refreshUC  common.Handler[refresh.Input, refresh.Output]

	meUC      common.Handler[profile.GetInput, profile.Output]
	profileUC common.Handler[profile.UpdateInput, profile.Output]
	linkUC    common.Handler[link.Input, link.Output]
}

func NewService(
        registerUC common.Handler[register.Input, login.Output],
        loginUC common.Handler[login.Input, login.Output],
        telegramUC common.Handler[telegram.Input, login.Output],
        refreshUC common.Handler[refresh.Input, refresh.Output],
        meUC common.Handler[profile.GetInput, profile.Output],
        profileUC common.Handler[profile.UpdateInput, profile.Output],
        linkUC common.Handler[link.Input, link.Output],
) Service {
        return &service{
                registerUC: registerUC,
                loginUC:    loginUC,
                telegramUC: telegramUC,
                refreshUC:  refreshUC,
                meUC:       meUC,
                profileUC:  profileUC,
                linkUC:     linkUC,
        }
}

func (s *service) Register(ctx context.Context, in register.Input) (login.Output, error) {
	return s.registerUC.Handle(ctx, in)
}

func (s *service) Login(ctx context.Context, in login.Input) (login.Output, error) {
        return s.loginUC.Handle(ctx, in)
}

func (s *service) LoginWithTelegram(ctx context.Context, in telegram.Input) (login.Output, error) {
        return s.telegramUC.Handle(ctx, in)
}

func (s *service) Refresh(ctx context.Context, in refresh.Input) (refresh.Output, error) {
	return s.refreshUC.Handle(ctx, in)
}

func (s *service) GetMe(ctx context.Context, in profile.GetInput) (profile.Output, error) {
	return s.meUC.Handle(ctx, in)
}

func (s *service) UpdateProfile(ctx context.Context, in profile.UpdateInput) (profile.Output, error) {
	return s.profileUC.Handle(ctx, in)
}

func (s *service) LinkProvider(ctx context.Context, in link.Input) (link.Output, error) {
	return s.linkUC.Handle(ctx, in)
}

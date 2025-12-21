package application

import (
	"context"

	"github.com/vaaxooo/xbackend/internal/modules/users/application/challenge"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/link"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/login"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/password"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/profile"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/refresh"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/register"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/telegram"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/twofactor"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/verification"
)

type Service interface {
	Register(ctx context.Context, in register.Input) (login.Output, error)
	Login(ctx context.Context, in login.Input) (login.Output, error)
	LoginWithTelegram(ctx context.Context, in telegram.Input) (login.Output, error)
	Refresh(ctx context.Context, in refresh.Input) (refresh.Output, error)
	ConfirmEmail(ctx context.Context, in verification.ConfirmEmailInput) (login.Output, error)
	RequestEmailConfirmation(ctx context.Context, in verification.RequestEmailInput) error
	RequestPasswordReset(ctx context.Context, in verification.RequestPasswordResetInput) error
	ResetPassword(ctx context.Context, in verification.ResetPasswordInput) error
	SetupTwoFactor(ctx context.Context, in twofactor.SetupInput) (twofactor.SetupOutput, error)
	ConfirmTwoFactor(ctx context.Context, in twofactor.ConfirmInput) error
	DisableTwoFactor(ctx context.Context, in twofactor.DisableInput) error

	GetMe(ctx context.Context, in profile.GetInput) (profile.Output, error)
	UpdateProfile(ctx context.Context, in profile.UpdateInput) (profile.Output, error)
	ChangePassword(ctx context.Context, in password.ChangeInput) error
	LinkProvider(ctx context.Context, in link.Input) (link.Output, error)
	ChallengeStatus(ctx context.Context, in challenge.StatusInput) (login.Output, error)
	VerifyChallengeTOTP(ctx context.Context, in challenge.VerifyTOTPInput) (login.Output, error)
	ResendChallengeEmail(ctx context.Context, in challenge.ResendEmailInput) (login.Output, error)
	ConfirmChallengeEmail(ctx context.Context, in challenge.ConfirmEmailInput) (login.Output, error)
}

package application

import (
	"context"

	"github.com/vaaxooo/xbackend/internal/modules/users/application/challenge"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/common"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/link"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/login"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/password"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/profile"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/refresh"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/register"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/session"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/telegram"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/twofactor"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/verification"
)

type service struct {
	registerUC             common.Handler[register.Input, login.Output]
	loginUC                common.Handler[login.Input, login.Output]
	telegramUC             common.Handler[telegram.Input, login.Output]
	refreshUC              common.Handler[refresh.Input, refresh.Output]
	confirmEmailUC         common.Handler[verification.ConfirmEmailInput, login.Output]
	requestEmailUC         common.Handler[verification.RequestEmailInput, struct{}]
	passwordResetRequestUC common.Handler[verification.RequestPasswordResetInput, struct{}]
	resetPasswordUC        common.Handler[verification.ResetPasswordInput, struct{}]
	twoFactorSetupUC       common.Handler[twofactor.SetupInput, twofactor.SetupOutput]
	twoFactorConfirmUC     common.Handler[twofactor.ConfirmInput, struct{}]
	twoFactorDisableUC     common.Handler[twofactor.DisableInput, struct{}]
	challengeStatusUC      common.Handler[challenge.StatusInput, login.Output]
	challengeVerifyTOTP    common.Handler[challenge.VerifyTOTPInput, login.Output]
	challengeResendEmail   common.Handler[challenge.ResendEmailInput, login.Output]
	challengeConfirmEmail  common.Handler[challenge.ConfirmEmailInput, login.Output]

	meUC       common.Handler[profile.GetInput, profile.Output]
	profileUC  common.Handler[profile.UpdateInput, profile.Output]
	passwordUC common.Handler[password.ChangeInput, struct{}]
	linkUC     common.Handler[link.Input, link.Output]

	sessionsListUC  common.Handler[session.ListInput, session.Output]
	sessionRevokeUC common.Handler[session.RevokeInput, struct{}]
	sessionsPurgeUC common.Handler[session.RevokeOthersInput, struct{}]
}

func NewService(
	registerUC common.Handler[register.Input, login.Output],
	loginUC common.Handler[login.Input, login.Output],
	telegramUC common.Handler[telegram.Input, login.Output],
	refreshUC common.Handler[refresh.Input, refresh.Output],
	confirmEmailUC common.Handler[verification.ConfirmEmailInput, login.Output],
	requestEmailUC common.Handler[verification.RequestEmailInput, struct{}],
	passwordResetRequestUC common.Handler[verification.RequestPasswordResetInput, struct{}],
	resetPasswordUC common.Handler[verification.ResetPasswordInput, struct{}],
	twoFactorSetupUC common.Handler[twofactor.SetupInput, twofactor.SetupOutput],
	twoFactorConfirmUC common.Handler[twofactor.ConfirmInput, struct{}],
	twoFactorDisableUC common.Handler[twofactor.DisableInput, struct{}],
	challengeStatusUC common.Handler[challenge.StatusInput, login.Output],
	challengeVerifyTOTP common.Handler[challenge.VerifyTOTPInput, login.Output],
	challengeResendEmail common.Handler[challenge.ResendEmailInput, login.Output],
	challengeConfirmEmail common.Handler[challenge.ConfirmEmailInput, login.Output],
	meUC common.Handler[profile.GetInput, profile.Output],
	profileUC common.Handler[profile.UpdateInput, profile.Output],
	passwordUC common.Handler[password.ChangeInput, struct{}],
	linkUC common.Handler[link.Input, link.Output],
	sessionsListUC common.Handler[session.ListInput, session.Output],
	sessionRevokeUC common.Handler[session.RevokeInput, struct{}],
	sessionsPurgeUC common.Handler[session.RevokeOthersInput, struct{}],
) Service {
	return &service{
		registerUC:             registerUC,
		loginUC:                loginUC,
		telegramUC:             telegramUC,
		refreshUC:              refreshUC,
		confirmEmailUC:         confirmEmailUC,
		requestEmailUC:         requestEmailUC,
		passwordResetRequestUC: passwordResetRequestUC,
		resetPasswordUC:        resetPasswordUC,
		twoFactorSetupUC:       twoFactorSetupUC,
		twoFactorConfirmUC:     twoFactorConfirmUC,
		twoFactorDisableUC:     twoFactorDisableUC,
		challengeStatusUC:      challengeStatusUC,
		challengeVerifyTOTP:    challengeVerifyTOTP,
		challengeResendEmail:   challengeResendEmail,
		challengeConfirmEmail:  challengeConfirmEmail,
		meUC:                   meUC,
		profileUC:              profileUC,
		passwordUC:             passwordUC,
		linkUC:                 linkUC,
		sessionsListUC:         sessionsListUC,
		sessionRevokeUC:        sessionRevokeUC,
		sessionsPurgeUC:        sessionsPurgeUC,
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

func (s *service) ConfirmEmail(ctx context.Context, in verification.ConfirmEmailInput) (login.Output, error) {
	return s.confirmEmailUC.Handle(ctx, in)
}

func (s *service) RequestEmailConfirmation(ctx context.Context, in verification.RequestEmailInput) error {
	_, err := s.requestEmailUC.Handle(ctx, in)
	return err
}

func (s *service) RequestPasswordReset(ctx context.Context, in verification.RequestPasswordResetInput) error {
	_, err := s.passwordResetRequestUC.Handle(ctx, in)
	return err
}

func (s *service) ResetPassword(ctx context.Context, in verification.ResetPasswordInput) error {
	_, err := s.resetPasswordUC.Handle(ctx, in)
	return err
}

func (s *service) SetupTwoFactor(ctx context.Context, in twofactor.SetupInput) (twofactor.SetupOutput, error) {
	return s.twoFactorSetupUC.Handle(ctx, in)
}

func (s *service) ConfirmTwoFactor(ctx context.Context, in twofactor.ConfirmInput) error {
	_, err := s.twoFactorConfirmUC.Handle(ctx, in)
	return err
}

func (s *service) DisableTwoFactor(ctx context.Context, in twofactor.DisableInput) error {
	_, err := s.twoFactorDisableUC.Handle(ctx, in)
	return err
}

func (s *service) ChallengeStatus(ctx context.Context, in challenge.StatusInput) (login.Output, error) {
	return s.challengeStatusUC.Handle(ctx, in)
}

func (s *service) VerifyChallengeTOTP(ctx context.Context, in challenge.VerifyTOTPInput) (login.Output, error) {
	return s.challengeVerifyTOTP.Handle(ctx, in)
}

func (s *service) ResendChallengeEmail(ctx context.Context, in challenge.ResendEmailInput) (login.Output, error) {
	return s.challengeResendEmail.Handle(ctx, in)
}

func (s *service) ConfirmChallengeEmail(ctx context.Context, in challenge.ConfirmEmailInput) (login.Output, error) {
	return s.challengeConfirmEmail.Handle(ctx, in)
}

func (s *service) GetMe(ctx context.Context, in profile.GetInput) (profile.Output, error) {
	return s.meUC.Handle(ctx, in)
}

func (s *service) UpdateProfile(ctx context.Context, in profile.UpdateInput) (profile.Output, error) {
	return s.profileUC.Handle(ctx, in)
}

func (s *service) ChangePassword(ctx context.Context, in password.ChangeInput) error {
	_, err := s.passwordUC.Handle(ctx, in)
	return err
}

func (s *service) LinkProvider(ctx context.Context, in link.Input) (link.Output, error) {
	return s.linkUC.Handle(ctx, in)
}

func (s *service) ListSessions(ctx context.Context, in session.ListInput) (session.Output, error) {
	return s.sessionsListUC.Handle(ctx, in)
}

func (s *service) RevokeSession(ctx context.Context, in session.RevokeInput) error {
	_, err := s.sessionRevokeUC.Handle(ctx, in)
	return err
}

func (s *service) RevokeOtherSessions(ctx context.Context, in session.RevokeOthersInput) error {
	_, err := s.sessionsPurgeUC.Handle(ctx, in)
	return err
}

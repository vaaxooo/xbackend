package verification

import (
	"context"
	"time"

	"github.com/vaaxooo/xbackend/internal/modules/users/application/common"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/events"
	"github.com/vaaxooo/xbackend/internal/modules/users/domain"
)

type RequestEmailInput struct {
	Email string
}

type RequestPasswordResetInput struct {
	Email string
}

type RequestUseCase struct {
	identities domain.IdentityRepository
	tokens     domain.VerificationTokenRepository
	events     common.EventPublisher

	emailTTL          time.Duration
	passwordTTL       time.Duration
	minResendInterval time.Duration
}

func NewRequestUseCase(
	identities domain.IdentityRepository,
	tokens domain.VerificationTokenRepository,
	events common.EventPublisher,
	emailTTL time.Duration,
	passwordTTL time.Duration,
	minResendInterval time.Duration,
) *RequestUseCase {
	if emailTTL == 0 {
		emailTTL = 15 * time.Minute
	}
	if passwordTTL == 0 {
		passwordTTL = 15 * time.Minute
	}
	if minResendInterval == 0 {
		minResendInterval = time.Minute
	}
	return &RequestUseCase{
		identities:        identities,
		tokens:            tokens,
		events:            events,
		emailTTL:          emailTTL,
		passwordTTL:       passwordTTL,
		minResendInterval: minResendInterval,
	}
}

func (uc *RequestUseCase) RequestEmailConfirmation(ctx context.Context, in RequestEmailInput) error {
	return uc.request(ctx, in.Email, domain.TokenTypeEmailConfirmation)
}

func (uc *RequestUseCase) RequestPasswordReset(ctx context.Context, in RequestPasswordResetInput) error {
	return uc.request(ctx, in.Email, domain.TokenTypePasswordReset)
}

func (uc *RequestUseCase) request(ctx context.Context, emailRaw string, tokenType domain.TokenType) error {
	email, err := domain.NewEmail(emailRaw)
	if err != nil {
		return domain.ErrInvalidCredentials
	}

	ident, found, err := uc.identities.GetByProvider(ctx, email.Provider(), email.String())
	if err != nil {
		return common.NormalizeError(err)
	}
	if !found {
		return domain.ErrInvalidCredentials
	}

	latest, found, err := uc.tokens.GetLatest(ctx, ident.ID, tokenType)
	if err != nil {
		return common.NormalizeError(err)
	}
	if found && time.Since(latest.CreatedAt) < uc.minResendInterval {
		return domain.ErrTooManyRequests
	}

	now := time.Now().UTC()
	code, err := domain.GenerateNumericCode(6)
	if err != nil {
		return common.NormalizeError(err)
	}
	ttl := uc.emailTTL
	if tokenType == domain.TokenTypePasswordReset {
		ttl = uc.passwordTTL
	}
	token := domain.NewVerificationToken(ident.ID, tokenType, code, now, ttl)
	if err := uc.tokens.Create(ctx, token); err != nil {
		return common.NormalizeError(err)
	}

	switch tokenType {
	case domain.TokenTypeEmailConfirmation:
		return uc.events.PublishEmailConfirmationRequested(ctx, events.EmailConfirmationRequested{
			UserID:     ident.UserID.String(),
			IdentityID: ident.ID,
			Email:      email.String(),
			Code:       code,
			ExpiresAt:  token.ExpiresAt,
			OccurredAt: now,
		})
	case domain.TokenTypePasswordReset:
		return uc.events.PublishPasswordResetRequested(ctx, events.PasswordResetRequested{
			UserID:     ident.UserID.String(),
			IdentityID: ident.ID,
			Email:      email.String(),
			Code:       code,
			ExpiresAt:  token.ExpiresAt,
			OccurredAt: now,
		})
	default:
		return nil
	}
}

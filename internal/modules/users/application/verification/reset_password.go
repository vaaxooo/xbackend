package verification

import (
	"context"
	"time"

	"github.com/vaaxooo/xbackend/internal/modules/users/application/common"
	"github.com/vaaxooo/xbackend/internal/modules/users/domain"
)

type ResetPasswordInput struct {
	Email       string
	Code        string
	NewPassword string
}

type ResetPasswordUseCase struct {
	identities domain.IdentityRepository
	tokens     domain.VerificationTokenRepository
	hasher     domain.PasswordHasher
}

func NewResetPasswordUseCase(
	identities domain.IdentityRepository,
	tokens domain.VerificationTokenRepository,
	hasher domain.PasswordHasher,
) *ResetPasswordUseCase {
	return &ResetPasswordUseCase{
		identities: identities,
		tokens:     tokens,
		hasher:     hasher,
	}
}

func (uc *ResetPasswordUseCase) Execute(ctx context.Context, in ResetPasswordInput) error {
	email, err := domain.NewEmail(in.Email)
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

	token, found, err := uc.tokens.GetByCode(ctx, ident.ID, domain.TokenTypePasswordReset, in.Code)
	if err != nil {
		return common.NormalizeError(err)
	}
	now := time.Now().UTC()
	if !found || !token.IsValid(in.Code, now) {
		return domain.ErrInvalidCredentials
	}

	if err := uc.tokens.MarkUsed(ctx, token.ID, now); err != nil {
		return common.NormalizeError(err)
	}

	hash, err := domain.NewPasswordHash(ctx, in.NewPassword, uc.hasher)
	if err != nil {
		return common.NormalizeError(err)
	}
	ident.SecretHash = hash
	if err := uc.identities.Update(ctx, ident); err != nil {
		return common.NormalizeError(err)
	}

	return nil
}

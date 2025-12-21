package verification

import (
	"context"
	"time"

	"github.com/vaaxooo/xbackend/internal/modules/users/application/common"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/login"
	"github.com/vaaxooo/xbackend/internal/modules/users/domain"
)

type ConfirmEmailInput struct {
	Email string
	Code  string
}

type ConfirmEmailUseCase struct {
	users      domain.UserRepository
	identities domain.IdentityRepository
	tokens     domain.VerificationTokenRepository
	refresh    domain.RefreshTokenRepository
	access     common.AccessTokenIssuer

	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewConfirmEmailUseCase(
	users domain.UserRepository,
	identities domain.IdentityRepository,
	tokens domain.VerificationTokenRepository,
	refresh domain.RefreshTokenRepository,
	access common.AccessTokenIssuer,
	accessTTL time.Duration,
	refreshTTL time.Duration,
) *ConfirmEmailUseCase {
	if accessTTL == 0 {
		accessTTL = 15 * time.Minute
	}
	if refreshTTL == 0 {
		refreshTTL = 30 * 24 * time.Hour
	}
	return &ConfirmEmailUseCase{
		users:      users,
		identities: identities,
		tokens:     tokens,
		refresh:    refresh,
		access:     access,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func (uc *ConfirmEmailUseCase) Execute(ctx context.Context, in ConfirmEmailInput) (login.Output, error) {
	email, err := domain.NewEmail(in.Email)
	if err != nil {
		return login.Output{}, domain.ErrInvalidCredentials
	}

	ident, found, err := uc.identities.GetByProvider(ctx, email.Provider(), email.String())
	if err != nil {
		return login.Output{}, common.NormalizeError(err)
	}
	if !found {
		return login.Output{}, domain.ErrInvalidCredentials
	}

	token, found, err := uc.tokens.GetByCode(ctx, ident.ID, domain.TokenTypeEmailConfirmation, in.Code)
	if err != nil {
		return login.Output{}, common.NormalizeError(err)
	}
	if !found || !token.IsValid(in.Code, time.Now().UTC()) {
		return login.Output{}, domain.ErrInvalidCredentials
	}

	usedAt := time.Now().UTC()
	if err := uc.tokens.MarkUsed(ctx, token.ID, usedAt); err != nil {
		return login.Output{}, common.NormalizeError(err)
	}

	ident = ident.WithEmailVerified(usedAt)
	if err := uc.identities.Update(ctx, ident); err != nil {
		return login.Output{}, common.NormalizeError(err)
	}

	user, ok, err := uc.users.GetByID(ctx, ident.UserID)
	if err != nil {
		return login.Output{}, common.NormalizeError(err)
	}
	if !ok {
		return login.Output{}, domain.ErrInvalidCredentials
	}

	accessToken, err := uc.access.Issue(user.ID.String(), uc.accessTTL)
	if err != nil {
		return login.Output{}, common.NormalizeError(err)
	}

	refreshRaw, err := common.NewRefreshToken()
	if err != nil {
		return login.Output{}, common.NormalizeError(err)
	}
	refreshHash := common.HashToken(refreshRaw)
	refreshRecord := domain.NewRefreshTokenRecord(user.ID, refreshHash, usedAt, uc.refreshTTL)
	if err := uc.refresh.Create(ctx, refreshRecord); err != nil {
		return login.Output{}, common.NormalizeError(err)
	}

	return login.Output{
		UserID:       user.ID.String(),
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		MiddleName:   user.MiddleName,
		DisplayName:  user.DisplayName,
		AvatarURL:    user.AvatarURL,
		AccessToken:  accessToken,
		RefreshToken: refreshRaw,
	}, nil
}

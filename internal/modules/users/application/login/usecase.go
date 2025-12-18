package login

import (
	"context"
	"time"

	"github.com/vaaxooo/xbackend/internal/modules/users/application/common"
	"github.com/vaaxooo/xbackend/internal/modules/users/domain"
)

type UseCase struct {
	uow        common.UnitOfWork
	users      domain.UserRepository
	identities domain.IdentityRepository
	refresh    domain.RefreshTokenRepository
	hasher     domain.PasswordHasher

	access     common.AccessTokenIssuer
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func New(
	uow common.UnitOfWork,
	users domain.UserRepository,
	identities domain.IdentityRepository,
	refresh domain.RefreshTokenRepository,
	hasher domain.PasswordHasher,
	access common.AccessTokenIssuer,
	accessTTL time.Duration,
	refreshTTL time.Duration,
) *UseCase {
	if accessTTL == 0 {
		accessTTL = 15 * time.Minute
	}
	if refreshTTL == 0 {
		refreshTTL = 30 * 24 * time.Hour
	}
	return &UseCase{
		uow:        uow,
		users:      users,
		identities: identities,
		refresh:    refresh,
		hasher:     hasher,
		access:     access,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func (uc *UseCase) Execute(ctx context.Context, in Input) (Output, error) {
	email, err := domain.NewEmail(in.Email)
	if err != nil {
		return Output{}, domain.ErrInvalidCredentials
	}

	ident, found, err := uc.identities.GetByProvider(ctx, email.Provider(), email.String())
	if err != nil {
		return Output{}, common.NormalizeError(err)
	}
	if !found {
		return Output{}, domain.ErrInvalidCredentials
	}

	if err := ident.Authenticate(ctx, uc.hasher, in.Password); err != nil {
		return Output{}, domain.ErrInvalidCredentials
	}

	u, ok, err := uc.users.GetByID(ctx, ident.UserID)
	if err != nil {
		return Output{}, common.NormalizeError(err)
	}
	if !ok {
		return Output{}, domain.ErrInvalidCredentials
	}

	accessToken, err := uc.access.Issue(u.ID.String(), uc.accessTTL)
	if err != nil {
		return Output{}, common.NormalizeError(err)
	}

	refreshRaw, err := common.NewRefreshToken()
	if err != nil {
		return Output{}, common.NormalizeError(err)
	}
	refreshHash := common.HashToken(refreshRaw)

	now := time.Now().UTC()
	if err := uc.uow.Do(ctx, func(ctx context.Context) error {
		return common.NormalizeError(
			uc.refresh.Create(ctx, domain.NewRefreshTokenRecord(u.ID, refreshHash, now, uc.refreshTTL)),
		)
	}); err != nil {
		return Output{}, err
	}

	return Output{
		UserID:       u.ID.String(),
		FirstName:    u.FirstName,
		LastName:     u.LastName,
		MiddleName:   u.MiddleName,
		DisplayName:  u.DisplayName,
		AvatarURL:    u.AvatarURL,
		AccessToken:  accessToken,
		RefreshToken: refreshRaw,
	}, nil
}

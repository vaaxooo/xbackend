package login

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/vaaxooo/xbackend/internal/modules/users/application/common"
	"github.com/vaaxooo/xbackend/internal/modules/users/domain"
)

type UseCase struct {
	tx         common.Transactor
	users      domain.UserRepository
	identities domain.IdentityRepository
	refresh    domain.RefreshTokenRepository
	hasher     common.PasswordHasher

	access     common.AccessTokenIssuer
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func New(
	tx common.Transactor,
	users domain.UserRepository,
	identities domain.IdentityRepository,
	refresh domain.RefreshTokenRepository,
	hasher common.PasswordHasher,
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
		tx:         tx,
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
	email := common.NormalizeEmail(in.Email)
	if !common.IsValidEmail(email) {
		return Output{}, domain.ErrInvalidCredentials
	}

	ident, found, err := uc.identities.GetByProvider(ctx, "email", email)
	if err != nil {
		return Output{}, err
	}
	if !found {
		return Output{}, domain.ErrInvalidCredentials
	}

	if err := uc.hasher.Compare(ctx, ident.SecretHash, in.Password); err != nil {
		return Output{}, domain.ErrInvalidCredentials
	}

	u, ok, err := uc.users.GetByID(ctx, ident.UserID)
	if err != nil {
		return Output{}, err
	}
	if !ok {
		return Output{}, domain.ErrInvalidCredentials
	}

	accessToken, err := uc.access.Issue(u.ID, uc.accessTTL)
	if err != nil {
		return Output{}, err
	}

	refreshRaw, err := common.NewRefreshToken()
	if err != nil {
		return Output{}, err
	}
	refreshHash := common.HashToken(refreshRaw)

	now := time.Now().UTC()
	if err := uc.tx.WithinTx(ctx, func(ctx context.Context) error {
		return uc.refresh.Create(ctx, domain.RefreshToken{
			ID:        uuid.NewString(),
			UserID:    u.ID,
			TokenHash: refreshHash,
			ExpiresAt: now.Add(uc.refreshTTL),
			CreatedAt: now,
		})
	}); err != nil {
		return Output{}, err
	}

	return Output{
		UserID:       u.ID,
		FirstName:    u.FirstName,
		LastName:     u.LastName,
		MiddleName:   u.MiddleName,
		DisplayName:  u.DisplayName,
		AvatarURL:    u.AvatarURL,
		AccessToken:  accessToken,
		RefreshToken: refreshRaw,
	}, nil
}

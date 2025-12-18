package register

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/vaaxooo/xbackend/internal/modules/users/app/common"
	"github.com/vaaxooo/xbackend/internal/modules/users/app/login"
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

func (uc *UseCase) Execute(ctx context.Context, in Input) (login.Output, error) {
	email := common.NormalizeEmail(in.Email)
	if !common.IsValidEmail(email) {
		return login.Output{}, domain.ErrInvalidEmail
	}
	if !common.IsStrongPassword(in.Password) {
		return login.Output{}, domain.ErrWeakPassword
	}

	if _, found, err := uc.identities.GetByProvider(ctx, "email", email); err != nil {
		return login.Output{}, err
	} else if found {
		return login.Output{}, domain.ErrEmailAlreadyUsed
	}

	hash, err := uc.hasher.Hash(ctx, in.Password)
	if err != nil {
		return login.Output{}, err
	}

	userID := uuid.NewString()
	now := time.Now().UTC()

	accessToken, err := uc.access.Issue(userID, uc.accessTTL)
	if err != nil {
		return login.Output{}, err
	}

	refreshRaw, err := common.NewRefreshToken()
	if err != nil {
		return login.Output{}, err
	}
	refreshHash := common.HashToken(refreshRaw)

	if err := uc.tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := uc.users.Create(ctx, domain.User{
			ID:                userID,
			DisplayName:       in.DisplayName,
			AvatarURL:         "",
			ProfileCustomized: false,
			CreatedAt:         now,
		}); err != nil {
			return err
		}

		if err := uc.identities.Create(ctx, domain.Identity{
			ID:             uuid.NewString(),
			UserID:         userID,
			Provider:       "email",
			ProviderUserID: email,
			SecretHash:     hash,
			CreatedAt:      now,
		}); err != nil {
			return err
		}

		if err := uc.refresh.Create(ctx, domain.RefreshToken{
			ID:        uuid.NewString(),
			UserID:    userID,
			TokenHash: refreshHash,
			ExpiresAt: now.Add(uc.refreshTTL),
			CreatedAt: now,
		}); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return login.Output{}, err
	}

	return login.Output{
		UserID:       userID,
		DisplayName:  in.DisplayName,
		AvatarURL:    "",
		AccessToken:  accessToken,
		RefreshToken: refreshRaw,
	}, nil
}

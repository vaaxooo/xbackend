package register

import (
	"context"
	"time"

	"github.com/vaaxooo/xbackend/internal/modules/users/app/common"
	"github.com/vaaxooo/xbackend/internal/modules/users/app/login"
	"github.com/vaaxooo/xbackend/internal/modules/users/domain"
)

type UseCase struct {
	tx         common.Transactor
	users      domain.UserRepository
	identities domain.IdentityRepository
	refresh    domain.RefreshTokenRepository
	hasher     domain.PasswordHasher

	access     common.AccessTokenIssuer
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func New(
	tx common.Transactor,
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
	email, err := domain.NewEmail(in.Email)
	if err != nil {
		return login.Output{}, err
	}
	if err := email.EnsureUnique(ctx, uc.identities); err != nil {
		return login.Output{}, err
	}

	hash, err := domain.NewPasswordHash(ctx, in.Password, uc.hasher)
	if err != nil {
		return login.Output{}, err
	}

	userID := domain.NewUserID()
	now := time.Now().UTC()
	user := domain.NewUser(userID, in.DisplayName, now)
	identity := domain.NewEmailIdentity(userID, email, hash, now)

	accessToken, err := uc.access.Issue(userID.String(), uc.accessTTL)
	if err != nil {
		return login.Output{}, err
	}

	refreshRaw, err := common.NewRefreshToken()
	if err != nil {
		return login.Output{}, err
	}
	refreshHash := common.HashToken(refreshRaw)
	refreshRecord := domain.NewRefreshTokenRecord(userID, refreshHash, now, uc.refreshTTL)

	if err := uc.tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := uc.users.Create(ctx, user); err != nil {
			return err
		}

		if err := uc.identities.Create(ctx, identity); err != nil {
			return err
		}

		if err := uc.refresh.Create(ctx, refreshRecord); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return login.Output{}, err
	}

	return login.Output{
		UserID:       userID.String(),
		DisplayName:  in.DisplayName,
		AvatarURL:    "",
		AccessToken:  accessToken,
		RefreshToken: refreshRaw,
	}, nil
}

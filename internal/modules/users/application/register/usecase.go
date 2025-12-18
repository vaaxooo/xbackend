package register

import (
	"context"
	"time"

	"github.com/vaaxooo/xbackend/internal/modules/users/application/common"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/events"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/login"
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

	events common.EventPublisher
}

func New(
	uow common.UnitOfWork,
	users domain.UserRepository,
	identities domain.IdentityRepository,
	refresh domain.RefreshTokenRepository,
	hasher domain.PasswordHasher,
	access common.AccessTokenIssuer,
	events common.EventPublisher,
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
		events:     eventsOrNop(events),
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
		return login.Output{}, common.NormalizeError(err)
	}

	displayName, err := domain.NewDisplayName(in.DisplayName)
	if err != nil {
		return login.Output{}, common.NormalizeError(err)
	}

	hash, err := domain.NewPasswordHash(ctx, in.Password, uc.hasher)
	if err != nil {
		return login.Output{}, common.NormalizeError(err)
	}

	userID := domain.NewUserID()
	now := time.Now().UTC()
	user := domain.NewUser(userID, displayName, now)
	identity := domain.NewEmailIdentity(userID, email, hash, now)

	accessToken, err := uc.access.Issue(userID.String(), uc.accessTTL)
	if err != nil {
		return login.Output{}, common.NormalizeError(err)
	}

	refreshRaw, err := common.NewRefreshToken()
	if err != nil {
		return login.Output{}, common.NormalizeError(err)
	}
	refreshHash := common.HashToken(refreshRaw)
	refreshRecord := domain.NewRefreshTokenRecord(userID, refreshHash, now, uc.refreshTTL)

	event := events.UserRegistered{
		UserID:      userID.String(),
		Email:       email.String(),
		DisplayName: displayName.String(),
		OccurredAt:  now,
	}

	if err := uc.uow.Do(ctx, func(ctx context.Context) error {
		if err := uc.users.Create(ctx, user); err != nil {
			return common.NormalizeError(err)
		}

		if err := uc.identities.Create(ctx, identity); err != nil {
			return common.NormalizeError(err)
		}

		if err := uc.refresh.Create(ctx, refreshRecord); err != nil {
			return common.NormalizeError(err)
		}

		if err := uc.events.PublishUserRegistered(ctx, event); err != nil {
			return common.NormalizeError(err)
		}

		return nil
	}); err != nil {
		return login.Output{}, err
	}

	out := login.Output{
		UserID:       userID.String(),
		DisplayName:  displayName.String(),
		AvatarURL:    "",
		AccessToken:  accessToken,
		RefreshToken: refreshRaw,
	}

	return out, nil
}

func eventsOrNop(p common.EventPublisher) common.EventPublisher {
	if p == nil {
		return common.NopEventPublisher{}
	}
	return p
}

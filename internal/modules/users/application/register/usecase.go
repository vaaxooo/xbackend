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
	users      domain.UserRepository
	identities domain.IdentityRepository
	refresh    domain.RefreshTokenRepository
	tokens     domain.VerificationTokenRepository
	hasher     domain.PasswordHasher

	access                   common.AccessTokenIssuer
	accessTTL                time.Duration
	refreshTTL               time.Duration
	verificationTTL          time.Duration
	requireEmailConfirmation bool

	events common.EventPublisher
}

func New(
	users domain.UserRepository,
	identities domain.IdentityRepository,
	refresh domain.RefreshTokenRepository,
	tokens domain.VerificationTokenRepository,
	hasher domain.PasswordHasher,
	access common.AccessTokenIssuer,
	events common.EventPublisher,
	accessTTL time.Duration,
	refreshTTL time.Duration,
	verificationTTL time.Duration,
	requireEmailConfirmation bool,
) *UseCase {
	if accessTTL == 0 {
		accessTTL = 15 * time.Minute
	}
	if refreshTTL == 0 {
		refreshTTL = 30 * 24 * time.Hour
	}
	if verificationTTL == 0 {
		verificationTTL = 15 * time.Minute
	}
	return &UseCase{
		users:                    users,
		identities:               identities,
		refresh:                  refresh,
		tokens:                   tokens,
		hasher:                   hasher,
		access:                   access,
		events:                   eventsOrNop(events),
		accessTTL:                accessTTL,
		refreshTTL:               refreshTTL,
		verificationTTL:          verificationTTL,
		requireEmailConfirmation: requireEmailConfirmation,
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

	var accessToken string
	var refreshRaw string
	var refreshRecord domain.RefreshToken
	if !uc.requireEmailConfirmation {
		accessToken, err = uc.access.Issue(userID.String(), uc.accessTTL)
		if err != nil {
			return login.Output{}, common.NormalizeError(err)
		}

		refreshRaw, err = common.NewRefreshToken()
		if err != nil {
			return login.Output{}, common.NormalizeError(err)
		}
		refreshHash := common.HashToken(refreshRaw)
		refreshRecord = domain.NewRefreshTokenRecord(userID, refreshHash, now, uc.refreshTTL)
	}

	event := events.UserRegistered{
		UserID:      userID.String(),
		Email:       email.String(),
		DisplayName: displayName.String(),
		OccurredAt:  now,
	}

	if err := uc.users.Create(ctx, user); err != nil {
		return login.Output{}, common.NormalizeError(err)
	}

	if err := uc.identities.Create(ctx, identity); err != nil {
		return login.Output{}, common.NormalizeError(err)
	}

	if refreshRecord.ID != "" {
		if err := uc.refresh.Create(ctx, refreshRecord); err != nil {
			return login.Output{}, common.NormalizeError(err)
		}
	}

	if uc.requireEmailConfirmation {
		if err := uc.createEmailConfirmation(ctx, identity, email.String(), now); err != nil {
			return login.Output{}, common.NormalizeError(err)
		}
	}

	if err := uc.events.PublishUserRegistered(ctx, event); err != nil {
		return login.Output{}, common.NormalizeError(err)
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

func (uc *UseCase) createEmailConfirmation(ctx context.Context, identity domain.Identity, email string, now time.Time) error {
	code, err := domain.GenerateNumericCode(6)
	if err != nil {
		return err
	}
	token := domain.NewVerificationToken(identity.ID, domain.TokenTypeEmailConfirmation, code, now, uc.verificationTTL)
	if err := uc.tokens.Create(ctx, token); err != nil {
		return err
	}

	event := events.EmailConfirmationRequested{
		UserID:     identity.UserID.String(),
		IdentityID: identity.ID,
		Email:      email,
		Code:       code,
		ExpiresAt:  token.ExpiresAt,
		OccurredAt: now,
	}
	return uc.events.PublishEmailConfirmationRequested(ctx, event)
}

func eventsOrNop(p common.EventPublisher) common.EventPublisher {
	if p == nil {
		return common.NopEventPublisher{}
	}
	return p
}

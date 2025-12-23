package login

import (
	"context"
	"strings"
	"time"

	"github.com/vaaxooo/xbackend/internal/modules/users/application/common"
	"github.com/vaaxooo/xbackend/internal/modules/users/domain"
)

type UseCase struct {
	users      domain.UserRepository
	identities domain.IdentityRepository
	refresh    domain.RefreshTokenRepository
	challenges domain.ChallengeRepository
	hasher     domain.PasswordHasher

	access                   common.AccessTokenIssuer
	accessTTL                time.Duration
	refreshTTL               time.Duration
	requireEmailVerification bool

	challengeTTL       time.Duration
	totpAttempts       int
	totpLockDuration   time.Duration
	requestEmailVerify func(context.Context, domain.Identity) error
}

func New(
	users domain.UserRepository,
	identities domain.IdentityRepository,
	refresh domain.RefreshTokenRepository,
	challenges domain.ChallengeRepository,
	hasher domain.PasswordHasher,
	access common.AccessTokenIssuer,
	accessTTL time.Duration,
	refreshTTL time.Duration,
	requireEmailVerification bool,
	challengeTTL time.Duration,
	totpAttempts int,
	totpLockDuration time.Duration,
	requestEmailVerify func(context.Context, domain.Identity) error,
) *UseCase {
	if accessTTL == 0 {
		accessTTL = 15 * time.Minute
	}
	if refreshTTL == 0 {
		refreshTTL = 30 * 24 * time.Hour
	}
	if challengeTTL == 0 {
		challengeTTL = 5 * time.Minute
	}
	if totpAttempts <= 0 {
		totpAttempts = 3
	}
	if totpLockDuration == 0 {
		totpLockDuration = 5 * time.Minute
	}
	return &UseCase{
		users:                    users,
		identities:               identities,
		refresh:                  refresh,
		challenges:               challenges,
		hasher:                   hasher,
		access:                   access,
		accessTTL:                accessTTL,
		refreshTTL:               refreshTTL,
		requireEmailVerification: requireEmailVerification,
		challengeTTL:             challengeTTL,
		totpAttempts:             totpAttempts,
		totpLockDuration:         totpLockDuration,
		requestEmailVerify:       requestEmailVerify,
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

	requiredSteps := make([]domain.ChallengeStep, 0)
	now := time.Now().UTC()
	if u.Suspended || (u.BlockedUntil != nil && u.BlockedUntil.After(now)) {
		requiredSteps = append(requiredSteps, domain.ChallengeStepAccountBlocked)
	}

	if uc.requireEmailVerification && ident.Provider == "email" && !ident.IsEmailVerified() {
		requiredSteps = append(requiredSteps, domain.ChallengeStepEmailVerification)
		if uc.requestEmailVerify != nil {
			_ = uc.requestEmailVerify(ctx, ident)
		}
	}

	if ident.IsTwoFactorEnabled() {
		requiredSteps = append(requiredSteps, domain.ChallengeStepTOTP)
	}

	if len(requiredSteps) > 0 {
		challenge := domain.NewChallenge(u.ID, "auth_challenge", requiredSteps, now.Add(uc.challengeTTL))
		challenge.AttemptsLeft = uc.totpAttempts
		if len(requiredSteps) == 1 && requiredSteps[0] == domain.ChallengeStepAccountBlocked {
			challenge.Status = domain.ChallengeStatusBlocked
		}
		if err := uc.challenges.Create(ctx, challenge); err != nil {
			return Output{}, common.NormalizeError(err)
		}
		return Output{
			UserID:      u.ID.String(),
			Email:       ident.ProviderUserID,
			FirstName:   u.FirstName,
			LastName:    u.LastName,
			MiddleName:  u.MiddleName,
			DisplayName: u.DisplayName,
			AvatarURL:   u.AvatarURL,
			Status:      "challenge_required",
			Challenge: &ChallengeInfo{
				ID:             challenge.ID,
				Type:           challenge.Type,
				RequiredSteps:  stepsToString(challenge.RequiredSteps),
				CompletedSteps: stepsToString(challenge.CompletedSteps),
				Status:         string(challenge.Status),
				ExpiresIn:      int64(challenge.ExpiresAt.Sub(now).Seconds()),
				AttemptsLeft:   challenge.AttemptsLeft,
				LockUntil:      challenge.LockUntil,
				MaskedEmail:    maskEmail(ident.ProviderUserID),
			},
		}, nil
	}

	refreshRaw, err := common.NewRefreshToken()
	if err != nil {
		return Output{}, common.NormalizeError(err)
	}
	refreshHash := common.HashToken(refreshRaw)

	now = time.Now().UTC()
	refreshRecord, reuse, err := common.PrepareRefreshRecord(ctx, uc.refresh, u.ID, refreshHash, now, uc.refreshTTL)
	if err != nil {
		return Output{}, common.NormalizeError(err)
	}
	accessToken, err := uc.access.Issue(u.ID.String(), refreshRecord.ID, uc.accessTTL)
	if err != nil {
		return Output{}, common.NormalizeError(err)
	}

	if reuse {
		if err := uc.refresh.Update(ctx, refreshRecord); err != nil {
			return Output{}, common.NormalizeError(err)
		}
	} else if err := uc.refresh.Create(ctx, refreshRecord); err != nil {
		return Output{}, common.NormalizeError(err)
	}

	return Output{
		UserID:       u.ID.String(),
		Email:        ident.ProviderUserID,
		FirstName:    u.FirstName,
		LastName:     u.LastName,
		MiddleName:   u.MiddleName,
		DisplayName:  u.DisplayName,
		AvatarURL:    u.AvatarURL,
		AccessToken:  accessToken,
		RefreshToken: refreshRaw,
	}, nil
}

func stepsToString(steps []domain.ChallengeStep) []string {
	result := make([]string, 0, len(steps))
	for _, step := range steps {
		result = append(result, string(step))
	}
	return result
}

func maskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email
	}
	local := parts[0]
	if len(local) > 2 {
		local = local[:1] + strings.Repeat("*", len(local)-2) + local[len(local)-1:]
	} else {
		local = strings.Repeat("*", len(local))
	}
	return local + "@" + parts[1]
}

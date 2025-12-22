package challenge

import (
	"context"
	"strings"
	"time"

	"github.com/vaaxooo/xbackend/internal/modules/users/application/common"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/login"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/totp"
	"github.com/vaaxooo/xbackend/internal/modules/users/domain"
)

type UseCase struct {
	challenges domain.ChallengeRepository
	identities domain.IdentityRepository
	users      domain.UserRepository
	refresh    domain.RefreshTokenRepository
	tokens     domain.VerificationTokenRepository
	access     common.AccessTokenIssuer

	accessTTL      time.Duration
	refreshTTL     time.Duration
	totpAttempts   int
	totpLock       time.Duration
	requestEmailFn func(context.Context, domain.Identity) error
}

type VerifyTOTPInput struct {
	ChallengeID string
	Code        string
}

type ResendEmailInput struct {
	ChallengeID string
}

type StatusInput struct {
	ChallengeID string
}

type ConfirmEmailInput struct {
	ChallengeID string
	Token       string
}

type Output = login.Output

func NewUseCase(challenges domain.ChallengeRepository, identities domain.IdentityRepository, users domain.UserRepository, refresh domain.RefreshTokenRepository, tokens domain.VerificationTokenRepository, access common.AccessTokenIssuer, accessTTL, refreshTTL time.Duration, totpAttempts int, totpLock time.Duration, requestEmailFn func(context.Context, domain.Identity) error) *UseCase {
	if accessTTL == 0 {
		accessTTL = 15 * time.Minute
	}
	if refreshTTL == 0 {
		refreshTTL = 30 * 24 * time.Hour
	}
	if totpAttempts <= 0 {
		totpAttempts = 3
	}
	if totpLock == 0 {
		totpLock = 5 * time.Minute
	}
	return &UseCase{
		challenges:     challenges,
		identities:     identities,
		users:          users,
		refresh:        refresh,
		tokens:         tokens,
		access:         access,
		accessTTL:      accessTTL,
		refreshTTL:     refreshTTL,
		totpAttempts:   totpAttempts,
		totpLock:       totpLock,
		requestEmailFn: requestEmailFn,
	}
}

func (uc *UseCase) Status(ctx context.Context, in StatusInput) (Output, error) {
	challenge, ok, err := uc.challenges.GetByID(ctx, in.ChallengeID)
	if err != nil || !ok {
		return Output{}, domain.ErrUnauthorized
	}
	return uc.challengeResponse(ctx, challenge, nil)
}

func (uc *UseCase) ResendEmail(ctx context.Context, in ResendEmailInput) (Output, error) {
	challenge, ok, err := uc.challenges.GetByID(ctx, in.ChallengeID)
	if err != nil || !ok {
		return Output{}, domain.ErrUnauthorized
	}
	if !challenge.NeedsStep(domain.ChallengeStepEmailVerification) {
		return uc.challengeResponse(ctx, challenge, nil)
	}
	ident, err := uc.identityForUser(ctx, challenge.UserID)
	if err != nil {
		return Output{}, err
	}
	if uc.requestEmailFn != nil {
		_ = uc.requestEmailFn(ctx, ident)
	}
	return uc.challengeResponse(ctx, challenge, &ident)
}

func (uc *UseCase) VerifyTOTP(ctx context.Context, in VerifyTOTPInput) (Output, error) {
	challenge, ok, err := uc.challenges.GetByID(ctx, in.ChallengeID)
	if err != nil || !ok {
		return Output{}, domain.ErrUnauthorized
	}
	now := time.Now().UTC()
	if challenge.IsExpired(now) {
		challenge = challenge.WithStatus(domain.ChallengeStatusExpired, now)
		_ = uc.challenges.Update(ctx, challenge)
		return uc.challengeResponse(ctx, challenge, nil)
	}
	if challenge.LockUntil != nil && challenge.LockUntil.After(now) {
		return uc.challengeResponse(ctx, challenge, nil)
	}
	if challenge.LockUntil != nil && challenge.LockUntil.Before(now) {
		challenge = challenge.WithLockUntil(nil, now)
		challenge = challenge.WithAttemptsLeft(uc.totpAttempts, now)
		_ = uc.challenges.Update(ctx, challenge)
	}
	if !challenge.NeedsStep(domain.ChallengeStepTOTP) {
		return uc.challengeResponse(ctx, challenge, nil)
	}
	ident, err := uc.identityForUser(ctx, challenge.UserID)
	if err != nil {
		return Output{}, err
	}
	if !totp.Validate(in.Code, ident.TOTPSecret) {
		left := challenge.AttemptsLeft - 1
		if left < 0 {
			left = 0
		}
		challenge = challenge.WithAttemptsLeft(left, now)
		if left == 0 {
			lock := now.Add(uc.totpLock)
			challenge = challenge.WithLockUntil(&lock, now)
		}
		_ = uc.challenges.Update(ctx, challenge)
		return uc.challengeResponse(ctx, challenge, &ident)
	}
	challenge = challenge.WithCompleted(domain.ChallengeStepTOTP, now)
	challenge = challenge.WithAttemptsLeft(uc.totpAttempts, now)
	if err := uc.challenges.Update(ctx, challenge); err != nil {
		return Output{}, common.NormalizeError(err)
	}
	return uc.challengeResponse(ctx, challenge, &ident)
}

func (uc *UseCase) ConfirmEmail(ctx context.Context, in ConfirmEmailInput) (Output, error) {
	challenge, ok, err := uc.challenges.GetByID(ctx, in.ChallengeID)
	if err != nil || !ok {
		return Output{}, domain.ErrUnauthorized
	}
	now := time.Now().UTC()
	if challenge.IsExpired(now) {
		challenge = challenge.WithStatus(domain.ChallengeStatusExpired, now)
		_ = uc.challenges.Update(ctx, challenge)
		return uc.challengeResponse(ctx, challenge, nil)
	}
	ident, err := uc.identityForUser(ctx, challenge.UserID)
	if err != nil {
		return Output{}, err
	}
	token, found, err := uc.tokens.GetByCode(ctx, ident.ID, domain.TokenTypeEmailConfirmation, in.Token)
	if err != nil || !found || !token.IsValid(in.Token, now) {
		return Output{}, domain.ErrUnauthorized
	}
	if err := uc.tokens.MarkUsed(ctx, token.ID, now); err != nil {
		return Output{}, common.NormalizeError(err)
	}
	ident = ident.WithEmailVerified(now)
	if err := uc.identities.Update(ctx, ident); err != nil {
		return Output{}, common.NormalizeError(err)
	}
	challenge = challenge.WithCompleted(domain.ChallengeStepEmailVerification, now)
	if err := uc.challenges.Update(ctx, challenge); err != nil {
		return Output{}, common.NormalizeError(err)
	}
	return uc.challengeResponse(ctx, challenge, &ident)
}

func (uc *UseCase) challengeResponse(ctx context.Context, challenge domain.Challenge, ident *domain.Identity) (Output, error) {
	user, ok, err := uc.users.GetByID(ctx, challenge.UserID)
	if err != nil {
		return Output{}, common.NormalizeError(err)
	}
	if !ok {
		return Output{}, domain.ErrUnauthorized
	}
	info := &login.ChallengeInfo{
		ID:             challenge.ID,
		Type:           challenge.Type,
		RequiredSteps:  stepsToString(challenge.RequiredSteps),
		CompletedSteps: stepsToString(challenge.CompletedSteps),
		Status:         string(challenge.Status),
		ExpiresIn:      int64(challenge.ExpiresAt.Sub(time.Now().UTC()).Seconds()),
		AttemptsLeft:   challenge.AttemptsLeft,
		LockUntil:      challenge.LockUntil,
	}
	if ident != nil {
		info.MaskedEmail = maskEmail(ident.ProviderUserID)
	}
	out := Output{
		UserID:      user.ID.String(),
		Email:       user.Email,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		MiddleName:  user.MiddleName,
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarURL,
		Status:      "challenge_required",
		Challenge:   info,
	}
	if challenge.Status == domain.ChallengeStatusCompleted {
		accessToken, refreshToken, err := uc.issueTokens(ctx, user)
		if err != nil {
			return Output{}, err
		}
		out.AccessToken = accessToken
		out.RefreshToken = refreshToken
		out.Status = string(domain.ChallengeStatusCompleted)
	}
	return out, nil
}

func (uc *UseCase) issueTokens(ctx context.Context, user domain.User) (string, string, error) {
	refreshRaw, err := common.NewRefreshToken()
	if err != nil {
		return "", "", common.NormalizeError(err)
	}
	refreshHash := common.HashToken(refreshRaw)
	now := time.Now().UTC()
	refreshRecord := common.NewRefreshRecord(ctx, user.ID, refreshHash, now, uc.refreshTTL)

	accessToken, err := uc.access.Issue(user.ID.String(), refreshRecord.ID, uc.accessTTL)
	if err != nil {
		return "", "", common.NormalizeError(err)
	}
	if err := uc.refresh.Create(ctx, refreshRecord); err != nil {
		return "", "", common.NormalizeError(err)
	}
	return accessToken, refreshRaw, nil
}

func (uc *UseCase) identityForUser(ctx context.Context, userID domain.UserID) (domain.Identity, error) {
	ident, found, err := uc.identities.GetByUserAndProvider(ctx, userID, "email")
	if err != nil {
		return domain.Identity{}, common.NormalizeError(err)
	}
	if !found {
		return domain.Identity{}, domain.ErrUnauthorized
	}
	return ident, nil
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

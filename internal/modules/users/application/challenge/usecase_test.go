package challenge

import (
	"context"
	"testing"
	"time"

	"github.com/vaaxooo/xbackend/internal/modules/users/domain"
)

func TestStatusClampsExpiredChallenge(t *testing.T) {
	expiredAt := time.Now().UTC().Add(-time.Minute)
	ch := domain.NewChallenge(domain.NewUserID(), "auth_challenge", []domain.ChallengeStep{domain.ChallengeStepTOTP}, expiredAt)
	ch.AttemptsLeft = 3

	repo := &challengeRepoMock{challenge: ch}
	uc := &UseCase{
		challenges:   repo,
		identities:   &identityRepoMock{},
		users:        &userRepoMock{user: domain.NewUser(ch.UserID, "", mustDisplayName(t, "user"), time.Now().UTC())},
		refresh:      &refreshRepoMock{},
		tokens:       &verificationRepoMock{},
		access:       &accessIssuerMock{},
		accessTTL:    time.Minute,
		refreshTTL:   time.Hour,
		totpAttempts: 3,
		totpLock:     time.Minute,
	}

	out, err := uc.Status(context.Background(), StatusInput{ChallengeID: ch.ID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Status != string(domain.ChallengeStatusExpired) {
		t.Fatalf("expected status expired, got %s", out.Status)
	}
	if out.Challenge == nil || out.Challenge.AttemptsLeft != 0 {
		t.Fatalf("expected attempts left 0, got %+v", out.Challenge)
	}
	if out.Challenge.ExpiresIn != 0 {
		t.Fatalf("expected expires_in to be clamped to 0, got %d", out.Challenge.ExpiresIn)
	}
	if updated := repo.lastUpdated; updated.AttemptsLeft != 0 || updated.Status != domain.ChallengeStatusExpired {
		t.Fatalf("expected repo update with expired status, got %+v", updated)
	}
}

// --- test doubles ---

type challengeRepoMock struct {
	challenge   domain.Challenge
	lastUpdated domain.Challenge
}

func (m *challengeRepoMock) Create(ctx context.Context, challenge domain.Challenge) error { return nil }
func (m *challengeRepoMock) Update(ctx context.Context, challenge domain.Challenge) error {
	m.lastUpdated = challenge
	m.challenge = challenge
	return nil
}
func (m *challengeRepoMock) GetByID(ctx context.Context, id string) (domain.Challenge, bool, error) {
	if m.challenge.ID == id {
		return m.challenge, true, nil
	}
	return domain.Challenge{}, false, nil
}
func (m *challengeRepoMock) GetPendingByUser(ctx context.Context, userID domain.UserID) (domain.Challenge, bool, error) {
	return domain.Challenge{}, false, nil
}

type identityRepoMock struct{}

func (identityRepoMock) Create(context.Context, domain.Identity) error { return nil }
func (identityRepoMock) GetByProvider(context.Context, string, string) (domain.Identity, bool, error) {
	return domain.Identity{}, false, nil
}
func (identityRepoMock) GetByUserAndProvider(context.Context, domain.UserID, string) (domain.Identity, bool, error) {
	return domain.Identity{}, false, nil
}
func (identityRepoMock) Update(context.Context, domain.Identity) error { return nil }

type userRepoMock struct{ user domain.User }

func (m *userRepoMock) Create(context.Context, domain.User) error { return nil }
func (m *userRepoMock) GetByID(context.Context, domain.UserID) (domain.User, bool, error) {
	return m.user, true, nil
}
func (m *userRepoMock) UpdateProfile(context.Context, domain.User) (domain.User, error) {
	return domain.User{}, nil
}

type refreshRepoMock struct{}

func (refreshRepoMock) Create(context.Context, domain.RefreshToken) error { return nil }
func (refreshRepoMock) GetByHash(context.Context, string) (domain.RefreshToken, bool, error) {
	return domain.RefreshToken{}, false, nil
}
func (refreshRepoMock) GetByID(context.Context, string) (domain.RefreshToken, bool, error) {
	return domain.RefreshToken{}, false, nil
}
func (refreshRepoMock) ListByUser(context.Context, domain.UserID) ([]domain.RefreshToken, error) {
	return nil, nil
}
func (refreshRepoMock) Revoke(context.Context, string) error                           { return nil }
func (refreshRepoMock) RevokeAllExcept(context.Context, domain.UserID, []string) error { return nil }

type verificationRepoMock struct{}

func (verificationRepoMock) Create(context.Context, domain.VerificationToken) error { return nil }
func (verificationRepoMock) GetLatest(context.Context, string, domain.TokenType) (domain.VerificationToken, bool, error) {
	return domain.VerificationToken{}, false, nil
}
func (verificationRepoMock) GetByCode(context.Context, string, domain.TokenType, string) (domain.VerificationToken, bool, error) {
	return domain.VerificationToken{}, false, nil
}
func (verificationRepoMock) MarkUsed(context.Context, string, time.Time) error { return nil }

type accessIssuerMock struct{}

func (accessIssuerMock) Issue(userID, sessionID string, ttl time.Duration) (string, error) {
	return "", nil
}

func mustDisplayName(t *testing.T, v string) domain.DisplayName {
	t.Helper()
	name, err := domain.NewDisplayName(v)
	if err != nil {
		t.Fatalf("display name err: %v", err)
	}
	return name
}

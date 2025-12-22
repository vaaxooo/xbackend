package register

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/vaaxooo/xbackend/internal/modules/users/application/common"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/events"
	"github.com/vaaxooo/xbackend/internal/modules/users/domain"
)

func TestUseCase_PublishesEventOnSuccess(t *testing.T) {
	uow := &stubUoW{}
	users := &stubUserRepo{}
	identities := &stubIdentityRepo{}
	refresh := &stubRefreshRepo{}
	tokens := &stubVerificationTokenRepo{}
	hasher := &stubHasher{}
	tokenIssuer := &stubTokenIssuer{}
	publisher := &stubEventPublisher{}

	uc := common.NewTransactionalUseCase(uow, New(users, identities, refresh, tokens, hasher, tokenIssuer, publisher, time.Minute, time.Hour, time.Minute, true))

	_, err := uc.Execute(context.Background(), Input{Email: "john@example.com", Password: "verystrong", DisplayName: "John"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !publisher.called {
		t.Fatalf("expected event to be published")
	}
	if publisher.event.UserID == "" || publisher.event.Email == "" {
		t.Fatalf("expected event to contain identifiers: %#v", publisher.event)
	}
}

func TestUseCase_InvalidDisplayName(t *testing.T) {
	uow := &stubUoW{}
	users := &stubUserRepo{}
	identities := &stubIdentityRepo{}
	refresh := &stubRefreshRepo{}
	tokens := &stubVerificationTokenRepo{}
	hasher := &stubHasher{}
	tokenIssuer := &stubTokenIssuer{}
	publisher := &stubEventPublisher{}

	uc := common.NewTransactionalUseCase(uow, New(users, identities, refresh, tokens, hasher, tokenIssuer, publisher, time.Minute, time.Hour, time.Minute, true))

	_, err := uc.Execute(context.Background(), Input{Email: "john@example.com", Password: "verystrong", DisplayName: " "})
	if !errors.Is(err, domain.ErrInvalidDisplayName) {
		t.Fatalf("expected invalid display name error, got %v", err)
	}
}

// --- test doubles ---

type stubUoW struct{ called bool }

func (s *stubUoW) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	s.called = true
	return fn(ctx)
}

type stubUserRepo struct{}

func (stubUserRepo) Create(context.Context, domain.User) error { return nil }
func (stubUserRepo) GetByID(context.Context, domain.UserID) (domain.User, bool, error) {
	return domain.User{}, false, errors.New("not implemented")
}
func (stubUserRepo) UpdateProfile(context.Context, domain.User) (domain.User, error) {
	return domain.User{}, errors.New("not implemented")
}

type stubIdentityRepo struct{}

func (stubIdentityRepo) Create(context.Context, domain.Identity) error { return nil }
func (stubIdentityRepo) GetByProvider(context.Context, string, string) (domain.Identity, bool, error) {
	return domain.Identity{}, false, nil
}
func (stubIdentityRepo) GetByUserAndProvider(context.Context, domain.UserID, string) (domain.Identity, bool, error) {
	return domain.Identity{}, false, nil
}
func (stubIdentityRepo) Update(context.Context, domain.Identity) error { return nil }

type stubRefreshRepo struct{}

func (stubRefreshRepo) Create(context.Context, domain.RefreshToken) error { return nil }
func (stubRefreshRepo) GetByHash(context.Context, string) (domain.RefreshToken, bool, error) {
	return domain.RefreshToken{}, false, errors.New("not implemented")
}
func (stubRefreshRepo) GetByID(context.Context, string) (domain.RefreshToken, bool, error) {
	return domain.RefreshToken{}, false, errors.New("not implemented")
}
func (stubRefreshRepo) ListByUser(context.Context, domain.UserID) ([]domain.RefreshToken, error) {
	return nil, errors.New("not implemented")
}
func (stubRefreshRepo) Revoke(context.Context, string) error { return errors.New("not implemented") }
func (stubRefreshRepo) RevokeAllExcept(context.Context, domain.UserID, []string) error {
	return errors.New("not implemented")
}

type stubHasher struct{}

func (stubHasher) Hash(context.Context, string) (string, error)  { return "hash", nil }
func (stubHasher) Compare(context.Context, string, string) error { return nil }

type stubTokenIssuer struct{}

func (stubTokenIssuer) Issue(string, string, time.Duration) (string, error) { return "token", nil }

type stubEventPublisher struct {
	called bool
	event  events.UserRegistered
}

func (s *stubEventPublisher) PublishUserRegistered(_ context.Context, e events.UserRegistered) error {
	s.called = true
	s.event = e
	return nil
}

func (stubEventPublisher) PublishEmailConfirmationRequested(context.Context, events.EmailConfirmationRequested) error {
	return nil
}

func (stubEventPublisher) PublishPasswordResetRequested(context.Context, events.PasswordResetRequested) error {
	return nil
}

type stubVerificationTokenRepo struct{}

func (stubVerificationTokenRepo) Create(context.Context, domain.VerificationToken) error { return nil }
func (stubVerificationTokenRepo) GetLatest(context.Context, string, domain.TokenType) (domain.VerificationToken, bool, error) {
	return domain.VerificationToken{}, false, nil
}
func (stubVerificationTokenRepo) GetByCode(context.Context, string, domain.TokenType, string) (domain.VerificationToken, bool, error) {
	return domain.VerificationToken{}, false, nil
}
func (stubVerificationTokenRepo) MarkUsed(context.Context, string, time.Time) error { return nil }

var _ common.EventPublisher = (*stubEventPublisher)(nil)

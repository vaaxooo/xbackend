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
	hasher := &stubHasher{}
	tokenIssuer := &stubTokenIssuer{}
	publisher := &stubEventPublisher{}

	uc := common.NewTransactionalUseCase(uow, New(users, identities, refresh, hasher, tokenIssuer, publisher, time.Minute, time.Hour))

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
	hasher := &stubHasher{}
	tokenIssuer := &stubTokenIssuer{}
	publisher := &stubEventPublisher{}

	uc := common.NewTransactionalUseCase(uow, New(users, identities, refresh, hasher, tokenIssuer, publisher, time.Minute, time.Hour))

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

type stubRefreshRepo struct{}

func (stubRefreshRepo) Create(context.Context, domain.RefreshToken) error { return nil }
func (stubRefreshRepo) GetByHash(context.Context, string) (domain.RefreshToken, bool, error) {
	return domain.RefreshToken{}, false, errors.New("not implemented")
}
func (stubRefreshRepo) Revoke(context.Context, string) error { return errors.New("not implemented") }

type stubHasher struct{}

func (stubHasher) Hash(context.Context, string) (string, error)  { return "hash", nil }
func (stubHasher) Compare(context.Context, string, string) error { return nil }

type stubTokenIssuer struct{}

func (stubTokenIssuer) Issue(string, time.Duration) (string, error) { return "token", nil }

type stubEventPublisher struct {
	called bool
	event  events.UserRegistered
}

func (s *stubEventPublisher) PublishUserRegistered(_ context.Context, e events.UserRegistered) error {
	s.called = true
	s.event = e
	return nil
}

var _ common.EventPublisher = (*stubEventPublisher)(nil)

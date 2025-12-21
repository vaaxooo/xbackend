package password

import (
	"context"
	"errors"
	"testing"

	"github.com/vaaxooo/xbackend/internal/modules/users/domain"
)

type stubIdentityRepo struct {
	identity  domain.Identity
	found     bool
	updated   domain.Identity
	updateErr error
}

func (s *stubIdentityRepo) Create(context.Context, domain.Identity) error { return nil }

func (s *stubIdentityRepo) GetByProvider(context.Context, string, string) (domain.Identity, bool, error) {
	return domain.Identity{}, false, nil
}

func (s *stubIdentityRepo) GetByUserAndProvider(context.Context, domain.UserID, string) (domain.Identity, bool, error) {
	return s.identity, s.found, nil
}

func (s *stubIdentityRepo) Update(_ context.Context, identity domain.Identity) error {
	s.updated = identity
	return s.updateErr
}

type stubPasswordHasher struct{}

func (stubPasswordHasher) Hash(_ context.Context, password string) (string, error) {
	return "hashed:" + password, nil
}

func (stubPasswordHasher) Compare(_ context.Context, hash string, password string) error {
	if hash != "hashed:"+password {
		return errors.New("invalid")
	}
	return nil
}

func TestChangePasswordSuccess(t *testing.T) {
	userID := domain.NewUserID()
	repo := &stubIdentityRepo{identity: domain.Identity{UserID: userID, Provider: "email", SecretHash: "hashed:old"}, found: true}
	uc := NewChange(repo, stubPasswordHasher{})

	if _, err := uc.Execute(context.Background(), ChangeInput{UserID: userID.String(), CurrentPassword: "old", NewPassword: "newpassword"}); err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	if repo.updated.SecretHash.String() != "hashed:newpassword" {
		t.Fatalf("expected new hash saved, got %s", repo.updated.SecretHash)
	}
}

func TestChangePasswordInvalidCurrent(t *testing.T) {
	userID := domain.NewUserID()
	repo := &stubIdentityRepo{identity: domain.Identity{UserID: userID, Provider: "email", SecretHash: "hashed:old"}, found: true}
	uc := NewChange(repo, stubPasswordHasher{})

	if _, err := uc.Execute(context.Background(), ChangeInput{UserID: userID.String(), CurrentPassword: "wrong", NewPassword: "newpassword"}); !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestChangePasswordWeak(t *testing.T) {
	userID := domain.NewUserID()
	repo := &stubIdentityRepo{identity: domain.Identity{UserID: userID, Provider: "email", SecretHash: "hashed:old"}, found: true}
	uc := NewChange(repo, stubPasswordHasher{})

	if _, err := uc.Execute(context.Background(), ChangeInput{UserID: userID.String(), CurrentPassword: "old", NewPassword: "weak"}); !errors.Is(err, domain.ErrWeakPassword) {
		t.Fatalf("expected ErrWeakPassword, got %v", err)
	}
}

func TestChangePasswordUnauthorized(t *testing.T) {
	repo := &stubIdentityRepo{found: false}
	uc := NewChange(repo, stubPasswordHasher{})

	if _, err := uc.Execute(context.Background(), ChangeInput{UserID: " ", CurrentPassword: "old", NewPassword: "newpassword"}); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

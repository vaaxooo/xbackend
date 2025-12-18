package domain

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNewExternalIdentity(t *testing.T) {
	id := UserID("user-1")
	_, err := NewExternalIdentity(id, "  ", "uid", time.Now())
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}

	ident, err := NewExternalIdentity(id, "google", " user ", time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ident.Provider != "google" || ident.ProviderUserID != "user" {
		t.Fatalf("unexpected identity: %+v", ident)
	}
}

func TestEnsureIdentityAvailable(t *testing.T) {
	userID := UserID("user-1")
	repo := &fakeIdentityRepo{byUser: map[string]Identity{userID.String() + "|github": {}}}

	if err := EnsureIdentityAvailable(context.Background(), repo, userID, "github", "gh-1"); !errors.Is(err, ErrIdentityAlreadyLinked) {
		t.Fatalf("expected ErrIdentityAlreadyLinked, got %v", err)
	}

	repo.byUser = nil
	repo.byProvider = map[string]Identity{"github|gh-1": {}}
	if err := EnsureIdentityAvailable(context.Background(), repo, userID, "github", "gh-1"); !errors.Is(err, ErrIdentityAlreadyLinked) {
		t.Fatalf("expected ErrIdentityAlreadyLinked by provider, got %v", err)
	}

	repo.byProvider = nil
	if err := EnsureIdentityAvailable(context.Background(), repo, userID, "github", "gh-1"); err != nil {
		t.Fatalf("expected available, got %v", err)
	}
}

func TestIdentityAuthenticate(t *testing.T) {
	hasher := &stubHasher{compareErr: errors.New("nope")}
	ident := Identity{SecretHash: ""}
	if err := ident.Authenticate(context.Background(), hasher, "pw"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials for missing secret, got %v", err)
	}

	ident.SecretHash = "hash"
	if err := ident.Authenticate(context.Background(), hasher, "pw"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials when compare fails, got %v", err)
	}

	hasher.compareErr = nil
	if err := ident.Authenticate(context.Background(), hasher, "pw"); err != nil {
		t.Fatalf("expected successful auth, got %v", err)
	}
}

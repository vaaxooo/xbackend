package domain

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type stubHasher struct {
	hash        string
	hashErr     error
	compareErr  error
	hashInputs  []string
	compareArgs [][2]string
}

func (s *stubHasher) Hash(_ context.Context, password string) (string, error) {
	s.hashInputs = append(s.hashInputs, password)
	return s.hash, s.hashErr
}

func (s *stubHasher) Compare(_ context.Context, hash string, password string) error {
	s.compareArgs = append(s.compareArgs, [2]string{hash, password})
	return s.compareErr
}

func TestNewEmail(t *testing.T) {
	email, err := NewEmail("  USER@Example.com ")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got := email.String(); got != "user@example.com" {
		t.Fatalf("expected normalized email, got %s", got)
	}
}

func TestNewEmail_Invalid(t *testing.T) {
	if _, err := NewEmail("bad"); !errors.Is(err, ErrInvalidEmail) {
		t.Fatalf("expected ErrInvalidEmail, got %v", err)
	}
}

func TestEmailEnsureUnique(t *testing.T) {
	repo := &fakeIdentityRepo{byProvider: map[string]Identity{"email|user@example.com": {}}}
	email, _ := NewEmail("user@example.com")
	if err := email.EnsureUnique(context.Background(), repo); !errors.Is(err, ErrEmailAlreadyUsed) {
		t.Fatalf("expected ErrEmailAlreadyUsed, got %v", err)
	}

	repo.byProvider = map[string]Identity{}
	if err := email.EnsureUnique(context.Background(), repo); err != nil {
		t.Fatalf("expected unique email, got %v", err)
	}
}

func TestNewPasswordHash(t *testing.T) {
	hasher := &stubHasher{hash: "secure"}
	hash, err := NewPasswordHash(context.Background(), "strongpass", hasher)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if hash.String() != "secure" {
		t.Fatalf("unexpected hash %s", hash)
	}
	if len(hasher.hashInputs) != 1 || hasher.hashInputs[0] != "strongpass" {
		t.Fatalf("expected hasher called with password")
	}
}

func TestNewPasswordHash_Weak(t *testing.T) {
	if _, err := NewPasswordHash(context.Background(), "short", &stubHasher{}); !errors.Is(err, ErrWeakPassword) {
		t.Fatalf("expected ErrWeakPassword, got %v", err)
	}
}

func TestPasswordHashCompare(t *testing.T) {
	hasher := &stubHasher{compareErr: errors.New("boom")}
	if err := PasswordHash("abc").Compare(context.Background(), hasher, "pw"); !errors.Is(err, hasher.compareErr) {
		t.Fatalf("expected compare error, got %v", err)
	}
	if len(hasher.compareArgs) != 1 {
		t.Fatalf("expected compare called once")
	}
	if hasher.compareArgs[0] != [2]string{"abc", "pw"} {
		t.Fatalf("unexpected compare args %#v", hasher.compareArgs[0])
	}
}

func TestNewDisplayName(t *testing.T) {
	name, err := NewDisplayName("  Alice ")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got := name.String(); got != "Alice" {
		t.Fatalf("expected normalized display name, got %s", got)
	}
}

func TestNewDisplayName_Invalid(t *testing.T) {
	tooShort := " "
	if _, err := NewDisplayName(tooShort); !errors.Is(err, ErrInvalidDisplayName) {
		t.Fatalf("expected ErrInvalidDisplayName, got %v", err)
	}
	tooLong := strings.Repeat("a", displayNameMaxLength+1)
	if _, err := NewDisplayName(tooLong); !errors.Is(err, ErrInvalidDisplayName) {
		t.Fatalf("expected ErrInvalidDisplayName for long name, got %v", err)
	}
}

func TestNewAvatarURL(t *testing.T) {
	avatar, err := NewAvatarURL(" https://example.com/a.png ")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if avatar.String() != "https://example.com/a.png" {
		t.Fatalf("expected normalized url, got %s", avatar.String())
	}

	empty, err := NewAvatarURL("   ")
	if err != nil {
		t.Fatalf("expected empty avatar allowed, got %v", err)
	}
	if empty.String() != "" {
		t.Fatalf("expected empty avatar, got %s", empty.String())
	}
}

func TestNewAvatarURL_Invalid(t *testing.T) {
	if _, err := NewAvatarURL("ftp://example.com"); !errors.Is(err, ErrInvalidAvatarURL) {
		t.Fatalf("expected ErrInvalidAvatarURL, got %v", err)
	}
	tooLong := strings.Repeat("h", avatarURLMaxLength+1)
	if _, err := NewAvatarURL(tooLong); !errors.Is(err, ErrInvalidAvatarURL) {
		t.Fatalf("expected ErrInvalidAvatarURL for long url, got %v", err)
	}
}

func TestParseUserID(t *testing.T) {
	if _, err := ParseUserID("   "); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}

	id, err := ParseUserID("  user-123  ")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id.String() != "user-123" {
		t.Fatalf("expected trimmed id, got %s", id)
	}
}

// fakeIdentityRepo provides minimal behaviour for tests.
type fakeIdentityRepo struct {
	byProvider map[string]Identity
	byUser     map[string]Identity
	err        error
}

func (f *fakeIdentityRepo) Create(_ context.Context, _ Identity) error { return nil }

func (f *fakeIdentityRepo) GetByProvider(_ context.Context, provider string, providerUserID string) (Identity, bool, error) {
	if f.err != nil {
		return Identity{}, false, f.err
	}
	if i, ok := f.byProvider[provider+"|"+providerUserID]; ok {
		return i, true, nil
	}
	return Identity{}, false, nil
}

func (f *fakeIdentityRepo) GetByUserAndProvider(_ context.Context, userID UserID, provider string) (Identity, bool, error) {
	if f.err != nil {
		return Identity{}, false, f.err
	}
	if i, ok := f.byUser[userID.String()+"|"+provider]; ok {
		return i, true, nil
	}
	return Identity{}, false, nil
}

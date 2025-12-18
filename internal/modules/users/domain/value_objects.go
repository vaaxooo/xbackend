package domain

import (
	"context"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

const (
	displayNameMinLength = 2
	displayNameMaxLength = 64
	avatarURLMaxLength   = 2048
)

type PasswordHasher interface {
	Hash(ctx context.Context, password string) (string, error)
	Compare(ctx context.Context, hash string, password string) error
}

type Email struct {
	value string
}

func NewEmail(raw string) (Email, error) {
	normalized := strings.TrimSpace(strings.ToLower(raw))
	if !isValidEmail(normalized) {
		return Email{}, ErrInvalidEmail
	}
	return Email{value: normalized}, nil
}

func (e Email) String() string {
	return e.value
}

func (e Email) Provider() string {
	return "email"
}

func (e Email) EnsureUnique(ctx context.Context, identities IdentityRepository) error {
	if _, found, err := identities.GetByProvider(ctx, e.Provider(), e.String()); err != nil {
		return err
	} else if found {
		return ErrEmailAlreadyUsed
	}
	return nil
}

type PasswordHash string

func NewPasswordHash(ctx context.Context, raw string, hasher PasswordHasher) (PasswordHash, error) {
	if !isStrongPassword(raw) {
		return "", ErrWeakPassword
	}
	hashed, err := hasher.Hash(ctx, raw)
	if err != nil {
		return "", err
	}
	return PasswordHash(hashed), nil
}

func (p PasswordHash) String() string {
	return string(p)
}

func (p PasswordHash) Compare(ctx context.Context, hasher PasswordHasher, password string) error {
	return hasher.Compare(ctx, p.String(), password)
}

type UserID string

func NewUserID() UserID {
	return UserID(uuid.NewString())
}

func ParseUserID(raw string) (UserID, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", ErrUnauthorized
	}
	return UserID(trimmed), nil
}

func (id UserID) String() string {
	return string(id)
}

func isValidEmail(email string) bool {
	if len(email) < 6 {
		return false
	}
	at := strings.IndexByte(email, '@')
	if at <= 0 || at == len(email)-1 {
		return false
	}
	dot := strings.LastIndexByte(email, '.')
	if dot < at+2 || dot >= len(email)-1 {
		return false
	}
	return true
}

func isStrongPassword(pw string) bool {
	return len(pw) >= 8
}

type DisplayName struct {
	value string
}

func NewDisplayName(raw string) (DisplayName, error) {
	normalized := strings.TrimSpace(raw)
	if l := len(normalized); l < displayNameMinLength || l > displayNameMaxLength {
		return DisplayName{}, ErrInvalidDisplayName
	}
	return DisplayName{value: normalized}, nil
}

func (d DisplayName) String() string {
	return d.value
}

type AvatarURL struct {
	value string
}

func NewAvatarURL(raw string) (AvatarURL, error) {
	normalized := strings.TrimSpace(raw)
	if normalized == "" {
		return AvatarURL{value: ""}, nil
	}
	if len(normalized) > avatarURLMaxLength {
		return AvatarURL{}, ErrInvalidAvatarURL
	}
	parsed, err := url.Parse(normalized)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return AvatarURL{}, ErrInvalidAvatarURL
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return AvatarURL{}, ErrInvalidAvatarURL
	}
	return AvatarURL{value: normalized}, nil
}

func (a AvatarURL) String() string {
	return a.value
}

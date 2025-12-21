package domain

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Identity struct {
	ID              string
	UserID          UserID
	Provider        string
	ProviderUserID  string
	SecretHash      PasswordHash
	EmailVerifiedAt *time.Time
	TOTPSecret      string
	TOTPConfirmedAt *time.Time
	CreatedAt       time.Time
}

func NewEmailIdentity(userID UserID, email Email, password PasswordHash, createdAt time.Time) Identity {
	return Identity{
		ID:              uuid.NewString(),
		UserID:          userID,
		Provider:        email.Provider(),
		ProviderUserID:  email.String(),
		SecretHash:      password,
		EmailVerifiedAt: nil,
		CreatedAt:       createdAt,
	}
}

func NewExternalIdentity(userID UserID, provider string, providerUserID string, createdAt time.Time) (Identity, error) {
	provider = strings.TrimSpace(provider)
	providerUserID = strings.TrimSpace(providerUserID)
	if provider == "" || providerUserID == "" {
		return Identity{}, ErrInvalidCredentials
	}
	return Identity{
		ID:             uuid.NewString(),
		UserID:         userID,
		Provider:       provider,
		ProviderUserID: providerUserID,
		CreatedAt:      createdAt,
	}, nil
}

func EnsureIdentityAvailable(ctx context.Context, identities IdentityRepository, userID UserID, provider string, providerUserID string) error {
	if _, found, err := identities.GetByUserAndProvider(ctx, userID, provider); err != nil {
		return err
	} else if found {
		return ErrIdentityAlreadyLinked
	}

	if _, found, err := identities.GetByProvider(ctx, provider, providerUserID); err != nil {
		return err
	} else if found {
		return ErrIdentityAlreadyLinked
	}

	return nil
}

func (i Identity) Authenticate(ctx context.Context, hasher PasswordHasher, password string) error {
	if i.SecretHash == "" {
		return ErrInvalidCredentials
	}
	if err := i.SecretHash.Compare(ctx, hasher, password); err != nil {
		return ErrInvalidCredentials
	}
	return nil
}

func (i Identity) WithEmailVerified(at time.Time) Identity {
	i.EmailVerifiedAt = &at
	return i
}

func (i Identity) IsEmailVerified() bool {
	return i.EmailVerifiedAt != nil
}

func (i Identity) WithTOTPSecret(secret string) Identity {
	i.TOTPSecret = strings.TrimSpace(secret)
	return i
}

func (i Identity) WithTOTPConfirmed(at time.Time) Identity {
	i.TOTPConfirmedAt = &at
	return i
}

func (i Identity) ClearTOTP() Identity {
	i.TOTPSecret = ""
	i.TOTPConfirmedAt = nil
	return i
}

func (i Identity) IsTwoFactorEnabled() bool {
	return strings.TrimSpace(i.TOTPSecret) != "" && i.TOTPConfirmedAt != nil
}

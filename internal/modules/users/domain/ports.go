package domain

import (
	"context"
	"time"
)

type User struct {
	ID                string
	FirstName         string
	LastName          string
	MiddleName        string
	DisplayName       string
	AvatarURL         string
	ProfileCustomized bool
	CreatedAt         time.Time
}

type UserRepository interface {
	Create(ctx context.Context, user User) error
	GetByID(ctx context.Context, userID string) (User, bool, error)
	UpdateProfile(ctx context.Context, in User) (User, error)
}

type Identity struct {
	ID             string
	UserID         string
	Provider       string
	ProviderUserID string
	SecretHash     string
	CreatedAt      time.Time
}

type IdentityRepository interface {
	Create(ctx context.Context, identity Identity) error
	GetByProvider(ctx context.Context, provider string, providerUserID string) (Identity, bool, error)
	GetByUserAndProvider(ctx context.Context, userID string, provider string) (Identity, bool, error)
}

type RefreshToken struct {
	ID        string
	UserID    string
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
	UserAgent string
	IP        string
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, t RefreshToken) error
	GetByHash(ctx context.Context, tokenHash string) (RefreshToken, bool, error)
	Revoke(ctx context.Context, tokenID string) error
}

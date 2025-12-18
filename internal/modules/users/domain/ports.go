package domain

import "context"

type UserRepository interface {
	Create(ctx context.Context, user User) error
	GetByID(ctx context.Context, userID UserID) (User, bool, error)
	UpdateProfile(ctx context.Context, in User) (User, error)
}

type IdentityRepository interface {
	Create(ctx context.Context, identity Identity) error
	GetByProvider(ctx context.Context, provider string, providerUserID string) (Identity, bool, error)
	GetByUserAndProvider(ctx context.Context, userID UserID, provider string) (Identity, bool, error)
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, t RefreshToken) error
	GetByHash(ctx context.Context, tokenHash string) (RefreshToken, bool, error)
	Revoke(ctx context.Context, tokenID string) error
}

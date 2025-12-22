package domain

import (
	"context"
	"time"
)

type UserRepository interface {
	Create(ctx context.Context, user User) error
	GetByID(ctx context.Context, userID UserID) (User, bool, error)
	UpdateProfile(ctx context.Context, in User) (User, error)
}

type IdentityRepository interface {
	Create(ctx context.Context, identity Identity) error
	GetByProvider(ctx context.Context, provider string, providerUserID string) (Identity, bool, error)
	GetByUserAndProvider(ctx context.Context, userID UserID, provider string) (Identity, bool, error)
	Update(ctx context.Context, identity Identity) error
}

type RefreshTokenRepository interface {
        Create(ctx context.Context, t RefreshToken) error
        GetByHash(ctx context.Context, tokenHash string) (RefreshToken, bool, error)
        GetByID(ctx context.Context, tokenID string) (RefreshToken, bool, error)
        ListByUser(ctx context.Context, userID UserID) ([]RefreshToken, error)
        Revoke(ctx context.Context, tokenID string) error
        RevokeAllExcept(ctx context.Context, userID UserID, keepIDs []string) error
}

type VerificationTokenRepository interface {
        Create(ctx context.Context, token VerificationToken) error
        GetLatest(ctx context.Context, identityID string, tokenType TokenType) (VerificationToken, bool, error)
        GetByCode(ctx context.Context, identityID string, tokenType TokenType, code string) (VerificationToken, bool, error)
        MarkUsed(ctx context.Context, tokenID string, usedAt time.Time) error
}

type ChallengeRepository interface {
        Create(ctx context.Context, challenge Challenge) error
        Update(ctx context.Context, challenge Challenge) error
        GetByID(ctx context.Context, id string) (Challenge, bool, error)
        GetPendingByUser(ctx context.Context, userID UserID) (Challenge, bool, error)
}

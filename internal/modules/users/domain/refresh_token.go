package domain

import (
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	ID        string
	UserID    UserID
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
	UserAgent string
	IP        string
}

func NewRefreshTokenRecord(userID UserID, tokenHash string, createdAt time.Time, ttl time.Duration) RefreshToken {
	return RefreshToken{
		ID:        uuid.NewString(),
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: createdAt.Add(ttl),
		CreatedAt: createdAt,
	}
}

func (t RefreshToken) IsValid(now time.Time) bool {
	if t.RevokedAt != nil {
		return false
	}
	return !now.After(t.ExpiresAt)
}

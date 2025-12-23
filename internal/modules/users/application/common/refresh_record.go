package common

import (
	"context"
	"time"

	"github.com/vaaxooo/xbackend/internal/modules/users/domain"
)

// NewRefreshRecord builds a refresh token record enriched with request metadata
// (user agent, IP) when available in the context.
func NewRefreshRecord(ctx context.Context, userID domain.UserID, tokenHash string, now time.Time, ttl time.Duration) domain.RefreshToken {
	meta, _ := RequestMetaFromContext(ctx)
	record := domain.NewRefreshTokenRecord(userID, tokenHash, now, ttl)
	record.UserAgent = meta.UserAgent
	record.IP = meta.IP
	return record
}

// PrepareRefreshRecord returns a refresh record and whether it should reuse an existing session row.
// It reuses an active session when the same user agent and IP are already stored.
func PrepareRefreshRecord(
	ctx context.Context,
	repo domain.RefreshTokenRepository,
	userID domain.UserID,
	tokenHash string,
	now time.Time,
	ttl time.Duration,
) (domain.RefreshToken, bool, error) {
	record := NewRefreshRecord(ctx, userID, tokenHash, now, ttl)
	if record.UserAgent == "" && record.IP == "" {
		return record, false, nil
	}

	existing, found, err := repo.FindActiveByFingerprint(ctx, userID, record.UserAgent, record.IP, now)
	if err != nil {
		return record, false, err
	}
	if found {
		record.ID = existing.ID
		return record, true, nil
	}
	return record, false, nil
}

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

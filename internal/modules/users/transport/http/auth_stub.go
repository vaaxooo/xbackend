package http

import (
	"context"

	"github.com/vaaxooo/xbackend/internal/modules/users/transport/httpctx"
)

// UserIDFromContext extracts the authenticated user id from the request context.
// The value is set by users transport middleware.
func UserIDFromContext(ctx context.Context) (string, bool) {
	return httpctx.UserIDFromContext(ctx)
}

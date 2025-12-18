package httpctx

import "context"

// ctxKey is unexported to avoid collisions with other packages.
type ctxKey string

const userIDKey ctxKey = "user_id"

// WithUserID stores the authenticated user id in the context.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// UserIDFromContext extracts the authenticated user id from the context.
func UserIDFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(userIDKey)
	s, ok := v.(string)
	return s, ok && s != ""
}

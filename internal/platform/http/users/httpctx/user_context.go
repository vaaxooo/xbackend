package httpctx

import "context"

// ctxKey is unexported to avoid collisions with other packages.
type ctxKey string

const (
	userIDKey    ctxKey = "user_id"
	sessionIDKey ctxKey = "session_id"
)

// WithUserID stores the authenticated user id in the context.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// WithSessionID stores the session id extracted from the access token in the context.
func WithSessionID(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, sessionIDKey, sessionID)
}

// UserIDFromContext extracts the authenticated user id from the context.
func UserIDFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(userIDKey)
	s, ok := v.(string)
	return s, ok && s != ""
}

// SessionIDFromContext extracts the session id from the context.
func SessionIDFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(sessionIDKey)
	s, ok := v.(string)
	return s, ok && s != ""
}

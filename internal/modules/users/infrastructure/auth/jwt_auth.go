package auth

import (
	"time"

	"github.com/vaaxooo/xbackend/internal/modules/users/infrastructure/tokens"
	"github.com/vaaxooo/xbackend/internal/modules/users/public"
)

// JWTAuth adapts the JWT driver to the public AuthPort interface,
// hiding the concrete token implementation from consumers.
type JWTAuth struct {
	issuer *tokens.HS256
}

func NewJWTAuth(secret string) (*JWTAuth, error) {
	issuer, err := tokens.NewHS256(secret)
	if err != nil {
		return nil, err
	}
	return &JWTAuth{issuer: issuer}, nil
}

func (a *JWTAuth) Issue(userID string, ttl time.Duration) (string, error) {
	return a.issuer.Issue(userID, ttl)
}

func (a *JWTAuth) Verify(token string) (public.AuthContext, error) {
	uid, err := a.issuer.Parse(token)
	if err != nil {
		return public.AuthContext{}, err
	}
	return public.AuthContext{UserID: uid}, nil
}

var _ public.AuthPort = (*JWTAuth)(nil)

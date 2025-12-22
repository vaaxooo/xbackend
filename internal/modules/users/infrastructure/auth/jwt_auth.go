package auth

import (
	"context"
	"errors"
	"time"

	"github.com/vaaxooo/xbackend/internal/modules/users/domain"
	"github.com/vaaxooo/xbackend/internal/modules/users/infrastructure/tokens"
	"github.com/vaaxooo/xbackend/internal/modules/users/public"
)

// JWTAuth adapts the JWT driver to the public AuthPort interface,
// hiding the concrete token implementation from consumers.
type JWTAuth struct {
	issuer  *tokens.HS256
	refresh domain.RefreshTokenRepository
}

func NewJWTAuth(secret string, refresh domain.RefreshTokenRepository) (*JWTAuth, error) {
	issuer, err := tokens.NewHS256(secret)
	if err != nil {
		return nil, err
	}
	return &JWTAuth{issuer: issuer, refresh: refresh}, nil
}

func (a *JWTAuth) Issue(userID, sessionID string, ttl time.Duration) (string, error) {
	return a.issuer.Issue(userID, sessionID, ttl)
}

func (a *JWTAuth) Verify(token string) (public.AuthContext, error) {
	claims, err := a.issuer.Parse(token)
	if err != nil {
		return public.AuthContext{}, err
	}

	session, found, err := a.refresh.GetByID(context.Background(), claims.SessionID)
	if err != nil {
		return public.AuthContext{}, err
	}
	if !found || session.UserID.String() != claims.UserID || !session.IsValid(time.Now().UTC()) {
		return public.AuthContext{}, errors.New("session revoked")
	}

	return public.AuthContext{UserID: claims.UserID, SessionID: claims.SessionID}, nil
}

var _ public.AuthPort = (*JWTAuth)(nil)

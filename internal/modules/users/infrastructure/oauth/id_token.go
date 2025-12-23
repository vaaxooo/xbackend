package oauth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// IDTokenVerifier validates JWT-based ID tokens against a remote JWKS.
type IDTokenVerifier struct {
	issuer   string
	audience string
	keys     *remoteKeySet
}

func NewIDTokenVerifier(issuer, audience, jwksURL string) *IDTokenVerifier {
	return &IDTokenVerifier{
		issuer:   strings.TrimSpace(issuer),
		audience: strings.TrimSpace(audience),
		keys:     newRemoteKeySet(jwksURL, 10*time.Minute),
	}
}

func (v *IDTokenVerifier) Verify(ctx context.Context, raw string, claims jwt.Claims) error {
	if strings.TrimSpace(raw) == "" {
		return errors.New("token is required")
	}
	token, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodRS256 {
			return nil, errors.New("unexpected signing method")
		}
		kid, _ := t.Header["kid"].(string)
		return v.keys.key(ctx, kid)
	}, jwt.WithLeeway(time.Minute))
	if err != nil || token == nil || !token.Valid {
		return errors.New("invalid token")
	}

	audOK, err := claims.GetAudience()
	if err != nil || len(audOK) == 0 {
		return errors.New("missing audience")
	}
	if v.audience != "" && !audienceMatches(audOK, v.audience) {
		return errors.New("audience mismatch")
	}
	iss, err := claims.GetIssuer()
	if err != nil {
		return errors.New("missing issuer")
	}
	if v.issuer != "" && iss != v.issuer {
		return errors.New("issuer mismatch")
	}
	if exp, err := claims.GetExpirationTime(); err != nil || exp == nil || !exp.After(time.Now()) {
		return errors.New("token expired")
	}
	return nil
}

func audienceMatches(list jwt.ClaimStrings, expected string) bool {
	for _, aud := range list {
		if aud == expected {
			return true
		}
	}
	return false
}

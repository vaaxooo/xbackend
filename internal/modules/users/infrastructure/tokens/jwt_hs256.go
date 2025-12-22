package tokens

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type HS256 struct {
	secret []byte
}

func NewHS256(secret string) (*HS256, error) {
	if len(secret) < 32 {
		return nil, errors.New("jwt secret too short (min 32 chars)")
	}
	return &HS256{secret: []byte(secret)}, nil
}

type Claims struct {
	UserID    string `json:"uid"`
	SessionID string `json:"sid"`
	jwt.RegisteredClaims
}

func (c *HS256) Issue(userID, sessionID string, ttl time.Duration) (string, error) {
	now := time.Now().UTC()

	claims := Claims{
		UserID:    userID,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(c.secret)
}

func (c *HS256) Parse(token string) (Claims, error) {
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return c.secret, nil
	})
	if err != nil {
		return Claims{}, err
	}

	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return Claims{}, errors.New("invalid token")
	}
	if claims.UserID == "" || claims.SessionID == "" {
		return Claims{}, errors.New("missing uid or sid")
	}

	return *claims, nil
}

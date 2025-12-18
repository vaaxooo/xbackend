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
	UserID string `json:"uid"`
	jwt.RegisteredClaims
}

func (c *HS256) Issue(userID string, ttl time.Duration) (string, error) {
	now := time.Now().UTC()

	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(c.secret)
}

func (c *HS256) Parse(token string) (string, error) {
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return c.secret, nil
	})
	if err != nil {
		return "", err
	}

	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return "", errors.New("invalid token")
	}
	if claims.UserID == "" {
		return "", errors.New("missing uid")
	}

	return claims.UserID, nil
}

package common

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"strings"
	"time"
)

type AccessTokenIssuer interface {
	Issue(userID string, ttl time.Duration) (string, error)
}

func NormalizeEmail(email string) string {
	return strings.TrimSpace(strings.ToLower(email))
}

func IsValidEmail(email string) bool {
	if len(email) < 6 {
		return false
	}
	at := strings.IndexByte(email, '@')
	if at <= 0 || at == len(email)-1 {
		return false
	}
	dot := strings.LastIndexByte(email, '.')
	if dot < at+2 || dot >= len(email)-1 {
		return false
	}
	return true
}

func IsStrongPassword(pw string) bool {
	return len(pw) >= 8
}

func NewRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func HashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

package domain

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
)

type TokenType string

const (
	TokenTypeEmailConfirmation TokenType = "email_confirmation"
	TokenTypePasswordReset     TokenType = "password_reset"
)

type VerificationToken struct {
	ID         string
	IdentityID string
	Type       TokenType
	Code       string
	ExpiresAt  time.Time
	UsedAt     *time.Time
	CreatedAt  time.Time
}

func NewVerificationToken(identityID string, tokenType TokenType, code string, issuedAt time.Time, ttl time.Duration) VerificationToken {
	return VerificationToken{
		ID:         uuid.NewString(),
		IdentityID: identityID,
		Type:       tokenType,
		Code:       code,
		ExpiresAt:  issuedAt.Add(ttl),
		CreatedAt:  issuedAt,
	}
}

func (t VerificationToken) IsValid(code string, now time.Time) bool {
	if t.UsedAt != nil {
		return false
	}
	if now.After(t.ExpiresAt) {
		return false
	}
	return t.Code == code
}

func (t VerificationToken) MarkUsed(at time.Time) VerificationToken {
	t.UsedAt = &at
	return t
}

func GenerateNumericCode(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("code length must be positive")
	}
	digits := make([]byte, length)
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		digits[i] = byte('0') + byte(n.Int64())
	}
	return string(digits), nil
}

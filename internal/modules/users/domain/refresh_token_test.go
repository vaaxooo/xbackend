package domain

import (
	"testing"
	"time"
)

func TestRefreshTokenValidity(t *testing.T) {
	now := time.Now().UTC()
	token := RefreshToken{ExpiresAt: now.Add(time.Hour)}
	if !token.IsValid(now) {
		t.Fatalf("expected token to be valid")
	}

	revokedAt := now
	token.RevokedAt = &revokedAt
	if token.IsValid(now) {
		t.Fatalf("expected revoked token to be invalid")
	}

	token.RevokedAt = nil
	token.ExpiresAt = now.Add(-time.Minute)
	if token.IsValid(now) {
		t.Fatalf("expected expired token to be invalid")
	}
}

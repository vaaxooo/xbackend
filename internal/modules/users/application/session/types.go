package session

import "time"

type ListInput struct {
	UserID              string
	CurrentRefreshToken string
}

type RevokeInput struct {
	UserID    string
	SessionID string
}

type RevokeOthersInput struct {
	UserID              string
	CurrentRefreshToken string
}

type Session struct {
	ID        string
	UserAgent string
	IP        string
	CreatedAt time.Time
	ExpiresAt time.Time
	RevokedAt *time.Time
	Current   bool
}

type Output struct {
	Sessions []Session
}

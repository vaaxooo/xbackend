package dto

import "time"

type SessionResponse struct {
	ID        string     `json:"id"`
	UserAgent string     `json:"user_agent"`
	IP        string     `json:"ip"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt time.Time  `json:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
	Current   bool       `json:"current"`
}

type SessionsResponse struct {
	Sessions []SessionResponse `json:"sessions"`
}

type RevokeSessionRequest struct {
	SessionID string `json:"session_id"`
}

type RevokeOtherSessionsRequest struct {
	CurrentRefreshToken string `json:"current_refresh_token"`
}

package events

import "time"

// UserRegistered is an integration event emitted after a new user is created.
// It is intentionally transport-agnostic so it can be forwarded to any broker.
type UserRegistered struct {
	UserID      string    `json:"user_id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	OccurredAt  time.Time `json:"occurred_at"`
}

package events

type EventType string

const (
	EventTypeUserRegistered             EventType = "users.user_registered"
	EventTypeEmailConfirmationRequested EventType = "users.email_confirmation_requested"
	EventTypePasswordResetRequested     EventType = "users.password_reset_requested"
)

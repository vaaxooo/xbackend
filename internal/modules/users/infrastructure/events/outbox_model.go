package events

import (
	"time"

	"github.com/google/uuid"
)

// OutboxMessage represents a stored integration event waiting for delivery.
// It is persisted by application use-cases inside the same transaction as
// domain changes so it can be retried later by a background worker.
type OutboxMessage struct {
	ID          uuid.UUID
	EventType   string
	Payload     []byte
	OccurredAt  time.Time
	CreatedAt   time.Time
	PublishedAt *time.Time
	Attempts    int
	LastError   *string
}

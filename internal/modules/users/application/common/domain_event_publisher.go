package common

import "context"

// DomainEventPublisher describes an idempotent delivery channel for domain or
// integration events. Implementations must guarantee that publishing the same
// eventID multiple times does not lead to duplicated deliveries downstream.
//
// The transport (Kafka, HTTP, logs, etc.) is intentionally abstracted away so
// the application layer can record events without depending on a particular
// infrastructure.
type DomainEventPublisher interface {
	Publish(ctx context.Context, eventID string, eventType string, payload []byte) error
}

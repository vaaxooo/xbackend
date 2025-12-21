package events

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	userevents "github.com/vaaxooo/xbackend/internal/modules/users/application/events"
)

type recordingOutboxRepo struct {
	msg OutboxMessage
}

func (r *recordingOutboxRepo) Add(_ context.Context, msg OutboxMessage) error {
	r.msg = msg
	return nil
}

func TestOutboxPublisherPublishUserRegistered(t *testing.T) {
	repo := &recordingOutboxRepo{}
	publisher := NewOutboxPublisher(repo)

	occurredAt := time.Now().UTC()
	evt := userevents.UserRegistered{
		UserID:      "user-id",
		Email:       "test@example.com",
		DisplayName: "Tester",
		OccurredAt:  occurredAt,
	}

	if err := publisher.PublishUserRegistered(context.Background(), evt); err != nil {
		t.Fatalf("publish failed: %v", err)
	}

	if repo.msg.EventType != string(EventTypeUserRegistered) {
		t.Fatalf("unexpected event type: %s", repo.msg.EventType)
	}

	if !repo.msg.OccurredAt.Equal(occurredAt) {
		t.Fatalf("occurredAt not preserved: %s", repo.msg.OccurredAt)
	}

	var decoded userevents.UserRegistered
	if err := json.Unmarshal(repo.msg.Payload, &decoded); err != nil {
		t.Fatalf("failed to decode payload: %v", err)
	}

	if decoded != evt {
		t.Fatalf("payload mismatch: %#v", decoded)
	}
}

package events

import (
	"context"

	"github.com/vaaxooo/xbackend/internal/modules/users/application/common"
)

// MultiDomainPublisher fans out outbox deliveries to multiple destinations.
type MultiDomainPublisher struct {
	publishers []common.DomainEventPublisher
}

func NewMultiDomainPublisher(publishers ...common.DomainEventPublisher) *MultiDomainPublisher {
	return &MultiDomainPublisher{publishers: publishers}
}

func (p *MultiDomainPublisher) Publish(ctx context.Context, eventID string, eventType string, payload []byte) error {
	for _, pub := range p.publishers {
		if err := pub.Publish(ctx, eventID, eventType, payload); err != nil {
			return err
		}
	}
	return nil
}

var _ common.DomainEventPublisher = (*MultiDomainPublisher)(nil)

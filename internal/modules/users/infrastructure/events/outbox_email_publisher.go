package events

import (
	"context"
	"encoding/json"

	"github.com/vaaxooo/xbackend/internal/modules/users/application/common"
	userevents "github.com/vaaxooo/xbackend/internal/modules/users/application/events"
	plog "github.com/vaaxooo/xbackend/internal/platform/log"
)

// OutboxEmailPublisher consumes raw outbox messages and delivers user-facing
// emails via the configured mailer. Unknown event types are ignored.
type OutboxEmailPublisher struct {
	mailer    Mailer
	logger    plog.Logger
	templates emailTemplates
}

func NewOutboxEmailPublisher(mailer Mailer, logger plog.Logger) *OutboxEmailPublisher {
	return &OutboxEmailPublisher{mailer: mailer, logger: logger, templates: defaultEmailTemplates}
}

func (p *OutboxEmailPublisher) Publish(ctx context.Context, _ string, eventType string, payload []byte) error {
	switch eventType {
	case string(EventTypeEmailConfirmationRequested):
		var evt userevents.EmailConfirmationRequested
		if err := json.Unmarshal(payload, &evt); err != nil {
			return err
		}
		text, html, err := p.templates.renderConfirmation(evt)
		if err != nil {
			return err
		}
		return p.mailer.Send(ctx, evt.Email, "Подтверждение почты", text, html)
	case string(EventTypePasswordResetRequested):
		var evt userevents.PasswordResetRequested
		if err := json.Unmarshal(payload, &evt); err != nil {
			return err
		}
		text, html, err := p.templates.renderPasswordReset(evt)
		if err != nil {
			return err
		}
		return p.mailer.Send(ctx, evt.Email, "Сброс пароля", text, html)
	default:
		p.logger.Debug(ctx, "outbox event ignored", "event_type", eventType)
		return nil
	}
}

var _ common.DomainEventPublisher = (*OutboxEmailPublisher)(nil)

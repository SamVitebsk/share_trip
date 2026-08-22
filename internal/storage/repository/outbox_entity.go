package repository

import (
	"encoding/json"
	"time"

	"share_trip/internal/outbox"

	"github.com/google/uuid"
)

type outboxEventEntity struct {
	ID          uuid.UUID
	EventName   string
	AggregateID uuid.UUID
	Payload     json.RawMessage
	CreatedAt   time.Time
}

func toOutboxEventEntity(event outbox.Event) outboxEventEntity {
	return outboxEventEntity{
		ID:          event.ID,
		EventName:   event.EventName,
		AggregateID: event.AggregateID,
		Payload:     event.Payload,
	}
}

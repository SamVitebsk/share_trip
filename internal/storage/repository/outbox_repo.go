package repository

import (
	"context"
	"fmt"

	"share_trip/internal/outbox"
)

func (r *TripRepoTx) CreateOutboxEvent(ctx context.Context, event outbox.Event) error {
	eventEntity := toOutboxEventEntity(event)
	_, err := r.tx.Exec(
		ctx,
		`INSERT INTO outbox_event(
                  id,
                  event_name,
                  aggregate_id,
                  payload
			  ) VALUES ($1, $2, $3, $4::jsonb)`,
		eventEntity.ID,
		eventEntity.EventName,
		eventEntity.AggregateID,
		string(eventEntity.Payload),
	)
	if err != nil {
		return fmt.Errorf("append outbox event: %w", mapPostgresError(err))
	}

	return nil
}

package repository

import (
	"context"
	"fmt"
	"share_trip/internal/outbox"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OutboxRelayRepo struct {
	pool *pgxpool.Pool
}

func NewOutboxRelayRepo(pool *pgxpool.Pool) *OutboxRelayRepo {
	return &OutboxRelayRepo{pool: pool}
}

func (r *OutboxRelayRepo) LockPending(ctx context.Context, limit int) ([]outbox.Event, error) {
	query := `
		SELECT id, event_name, aggregate_id, payload 
		FROM outbox_event 
		WHERE status = 'pending' 
		ORDER BY created_at ASC 
		LIMIT $1 
		FOR UPDATE SKIP LOCKED
	`

	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to lock pending events: %w", err)
	}
	defer rows.Close()

	var events []outbox.Event
	for rows.Next() {
		var entity outboxEventEntity
		if err := rows.Scan(&entity.ID, &entity.EventName, &entity.AggregateID, &entity.Payload); err != nil {
			return nil, fmt.Errorf("failed to scan outbox event: %w", err)
		}

		events = append(events, outbox.Event{
			ID:          entity.ID,
			EventName:   entity.EventName,
			AggregateID: entity.AggregateID,
			Payload:     entity.Payload,
		})
	}

	return events, rows.Err()
}

func (r *OutboxRelayRepo) MarkSent(ctx context.Context, eventID uuid.UUID) error {
	query := `
		UPDATE outbox_event 
		SET status = 'sent', sent_at = NOW() 
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, eventID)
	if err != nil {
		return fmt.Errorf("failed to mark event %s as sent: %w", eventID, err)
	}
	return nil
}

func (r *OutboxRelayRepo) MarkFailed(ctx context.Context, eventID uuid.UUID, lastErr error) error {
	query := `
		UPDATE outbox_event 
		SET 
			attempts = attempts + 1,
			last_error = $2,
			status = CASE 
				WHEN attempts + 1 >= 10 THEN 'failed' 
				ELSE 'pending' 
			END
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, eventID, lastErr.Error())
	if err != nil {
		return fmt.Errorf("failed to mark event %s as failed: %w", eventID, err)
	}
	return nil
}

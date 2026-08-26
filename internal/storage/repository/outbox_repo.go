package repository

import (
	"context"
	"fmt"
	"log/slog"

	"share_trip/internal/observability/logctx"
	"share_trip/internal/outbox"
)

func (r *TripRepoTx) CreateOutboxEvent(ctx context.Context, event outbox.Event) error {
	eventEntity := toOutboxEventEntity(event)
	logger := logctx.Logger(ctx).With(
		slog.String("layer", "repository"),
		slog.String("repository", "OutboxRepository"),
		slog.String("operation", "CreateOutboxEvent"),
		slog.String("outbox_event_id", eventEntity.ID.String()),
		slog.String("event_name", eventEntity.EventName),
		slog.String("aggregate_id", eventEntity.AggregateID.String()),
	)

	logger.InfoContext(ctx, "insert outbox-события начат")

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
		logger.ErrorContext(
			ctx,
			"insert outbox-события не выполнен",
			slog.Any("error", err),
		)
		return fmt.Errorf("append outbox event: %w", mapPostgresError(err))
	}

	logger.InfoContext(ctx, "insert outbox-события выполнен")

	return nil
}

package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	config "share_trip/configs"
	"share_trip/internal/clients/kafka"
	"share_trip/internal/outbox"
	"time"

	"github.com/google/uuid"
)

type RelayRepo interface {
	LockPending(ctx context.Context, limit int) ([]outbox.Event, error)
	MarkSent(ctx context.Context, eventID uuid.UUID) error
	MarkFailed(ctx context.Context, eventID uuid.UUID, lastErr error) error
}

type KafkaPublisher interface {
	SendEvent(ctx context.Context, event kafka.TripPublished) error
}

type Publisher struct {
	repo   RelayRepo
	kafka  KafkaPublisher
	logger *slog.Logger
	cfg    config.OutboxConfig
}

func NewPublisher(repo RelayRepo, kv KafkaPublisher, logger *slog.Logger, cfg config.OutboxConfig) *Publisher {
	return &Publisher{
		repo:   repo,
		kafka:  kv,
		logger: logger,
		cfg:    cfg,
	}
}

func (p *Publisher) Run(ctx context.Context) error {
	ticker := time.NewTicker(p.cfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := p.publishBatch(ctx); err != nil {
				p.logger.Error("ошибка при публикации пачки outbox-событий", slog.Any("error", err))
			}
		}
	}
}

func (p *Publisher) publishBatch(ctx context.Context) error {
	events, err := p.repo.LockPending(ctx, p.cfg.BatchSize)
	if err != nil {
		return err
	}

	for _, event := range events {
		var kEvent kafka.TripPublished
		if err := json.Unmarshal(event.Payload, &kEvent); err != nil {
			p.logger.Error("невозможно распарсить payload", slog.Any("error", err), slog.String("event_id", event.ID.String()))
			if markErr := p.repo.MarkFailed(ctx, event.ID, err); markErr != nil {
				return fmt.Errorf("fatal db error while marking parse error: %w", markErr)
			}
			continue
		}

		if err := p.kafka.SendEvent(ctx, kEvent); err != nil {
			p.logger.Warn("не удалось отправить событие в Kafka", slog.Any("error", err), slog.String("event_id", event.ID.String()))
			if markErr := p.repo.MarkFailed(ctx, event.ID, err); markErr != nil {
				return fmt.Errorf("fatal db error while marking kafka error: %w", markErr)
			}
			continue
		}

		if err := p.repo.MarkSent(ctx, event.ID); err != nil {
			return fmt.Errorf("fatal db error while marking event sent: %w", err)
		}
	}

	return nil
}

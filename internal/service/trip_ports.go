package service

import (
	"context"

	"share_trip/internal/domain"
	"share_trip/internal/outbox"

	"github.com/google/uuid"
)

type TripRepository interface {
	Create(ctx context.Context, trip domain.Trip, history domain.TripHistory) error
	GetByID(ctx context.Context, tripID uuid.UUID) (domain.Trip, error)
}

type TripRepositoryTx interface {
	GetForUpdateByID(ctx context.Context, tripID uuid.UUID) (domain.Trip, error)
	UpdateStatus(ctx context.Context, tripID uuid.UUID, status domain.TripStatus) (domain.Trip, error)
	CreateHistory(ctx context.Context, history domain.TripHistory) error
	CreateOutboxEvent(ctx context.Context, event outbox.Event) error
}

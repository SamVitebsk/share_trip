package service

import (
	"context"

	"share_trip/internal/domain"

	"github.com/google/uuid"
)

type TripRepository interface {
	Create(ctx context.Context, trip domain.Trip, history domain.TripHistory) error
	GetByID(ctx context.Context, tripID uuid.UUID) (domain.Trip, error)
}

type TripRepositoryTx interface {
	GetForUpdateByID(ctx context.Context, tripID uuid.UUID) (domain.Trip, error)
	UpdateStatus(ctx context.Context, tripID uuid.UUID, status domain.TripStatus) error
	AppendHistory(ctx context.Context, history domain.TripHistory) error
}

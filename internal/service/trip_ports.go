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

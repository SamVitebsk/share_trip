package domain

import (
	"context"
	"github.com/google/uuid"
)

type TripRepositoryTx interface {
	GetForUpdateByID(ctx context.Context, tripID uuid.UUID) (Trip, error)
	UpdateStatus(ctx context.Context, tripID uuid.UUID, status TripStatus) (Trip, error)
	CreateHistory(ctx context.Context, history TripHistory) error
}

package service

import (
	"context"
	"share_trip/internal/clients/kafka"

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

type CheckServiceRequest struct {
	DriverID    string
	ServiceCode string
}

//go:generate go run go.uber.org/mock/mockgen@latest -source=trip_ports.go -destination=mocks/trip_ports_mocks.go -package=mocks
type ContractChecker interface {
	CheckService(ctx context.Context, req CheckServiceRequest) (CheckResult, error)
}

type CheckResult struct {
	Allowed bool
	Reason  string
}

type EventPublisher interface {
	SendEvent(ctx context.Context, event kafka.TripPublished) error
}

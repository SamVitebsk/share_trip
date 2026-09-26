package service

import (
	"context"
	"share_trip/internal/observability/metrics"
)

type TripTxRunner func(ctx context.Context, fn func(ctx context.Context, trips TripRepositoryTx) error) error

type TripService struct {
	tripRepository  TripRepository
	runTripTx       TripTxRunner
	metrics         *metrics.Metrics
	contractChecker ContractChecker
	eventPublisher  EventPublisher
}

func NewTripService(
	trips TripRepository,
	runTripTx TripTxRunner,
	metrics *metrics.Metrics,
	contractChecker ContractChecker,
	eventPublisher EventPublisher,
) *TripService {
	return &TripService{
		tripRepository:  trips,
		runTripTx:       runTripTx,
		metrics:         metrics,
		contractChecker: contractChecker,
		eventPublisher:  eventPublisher,
	}
}

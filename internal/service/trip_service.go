package service

import "context"

type TripTxRunner func(ctx context.Context, fn func(ctx context.Context, trips TripRepositoryTx) error) error

type TripService struct {
	trips     TripRepository
	runTripTx TripTxRunner
}

func NewTripService(trips TripRepository, runTripTx TripTxRunner) *TripService {
	return &TripService{
		trips:     trips,
		runTripTx: runTripTx,
	}
}

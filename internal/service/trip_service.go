package service

type TripService struct {
	trips TripRepository
}

func NewTripService(trips TripRepository) *TripService {
	return &TripService{trips: trips}
}

package service

import (
	"fmt"

	"share_trip/internal/domain"
)

var tripStatusTransitions = map[domain.TripStatus][]domain.TripStatus{
	domain.TripStatusDraft: {
		domain.TripStatusPublished,
		domain.TripStatusCanceled,
	},
	domain.TripStatusPublished: {
		domain.TripStatusCanceled,
		domain.TripStatusCompleted,
	},
}

func ValidateTripStatusTransition(from, to domain.TripStatus) error {
	for _, allowedStatus := range tripStatusTransitions[from] {
		if allowedStatus == to {
			return nil
		}
	}

	return Unprocessable(fmt.Sprintf("недопустимый переход статуса поездки: %s -> %s", from, to))
}

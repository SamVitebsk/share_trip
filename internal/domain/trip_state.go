package domain

var tripStatusTransitions = map[TripStatus]map[TripStatus]struct{}{
	TripStatusDraft: {
		TripStatusPublished: {},
		TripStatusCanceled:  {},
	},
	TripStatusPublished: {
		TripStatusCanceled:  {},
		TripStatusCompleted: {},
	},
}

func CanTransitionTripStatus(from, to TripStatus) bool {
	allowedTransitions, ok := tripStatusTransitions[from]
	if !ok {
		return false
	}

	_, ok = allowedTransitions[to]
	return ok
}

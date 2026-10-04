package domain

var tripStatusTransitions = map[TripStatus]map[TripStatus]struct{}{
	TripStatusDraft: {
		TripStatusPublished: {},
		TripStatusCanceled:  {},
	},
	TripStatusPublished: {
		TripStatusStarted:   {},
		TripStatusCanceled:  {},
		TripStatusCompleted: {},
	},
	TripStatusStarted: {
		TripStatusCompleted: {},
	},
}

func canTransitionTripStatus(from, to TripStatus) bool {
	allowedTransitions, ok := tripStatusTransitions[from]
	if !ok {
		return false
	}

	_, ok = allowedTransitions[to]
	return ok
}

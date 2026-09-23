package outbox

import (
	"encoding/json"

	"github.com/google/uuid"
)

const (
	EventNameTripPublished = "trip_published"
	EventNameTripStarted   = "trip_started"
)

type Event struct {
	ID          uuid.UUID
	EventName   string
	AggregateID uuid.UUID
	Payload     json.RawMessage
}

type TripEventPayload struct {
	TripID uuid.UUID `json:"trip_id"`
}

func newTripEvent(tripID uuid.UUID, eventName string) (Event, error) {
	payload, err := json.Marshal(TripEventPayload{TripID: tripID})
	if err != nil {
		return Event{}, err
	}

	return Event{
		ID:          uuid.New(),
		EventName:   eventName,
		AggregateID: tripID,
		Payload:     payload,
	}, nil
}

func NewTripPublishedEvent(tripID uuid.UUID) (Event, error) {
	return newTripEvent(tripID, EventNameTripPublished)
}

func NewTripStartedEvent(tripID uuid.UUID) (Event, error) {
	return newTripEvent(tripID, EventNameTripStarted)
}

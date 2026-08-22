package outbox

import (
	"encoding/json"

	"github.com/google/uuid"
)

const EventNameTripPublished = "trip_published"

type Event struct {
	ID          uuid.UUID
	EventName   string
	AggregateID uuid.UUID
	Payload     json.RawMessage
}

func NewTripPublishedEvent(tripID uuid.UUID) (Event, error) {
	payload, err := json.Marshal(struct {
		TripID uuid.UUID `json:"trip_id"`
	}{
		TripID: tripID,
	})
	if err != nil {
		return Event{}, err
	}

	return Event{
		ID:          uuid.New(),
		EventName:   EventNameTripPublished,
		AggregateID: tripID,
		Payload:     payload,
	}, nil
}

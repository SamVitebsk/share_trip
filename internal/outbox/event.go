package outbox

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	EventNameTripPublished = "TripPublished"
	EventNameTripStarted   = "TripStarted"
)

type Event struct {
	ID          uuid.UUID
	EventName   string
	AggregateID uuid.UUID
	Payload     json.RawMessage
}

type TripEventPayload struct {
	EventID    string    `json:"event_id"`
	EventType  string    `json:"event_type"`
	TripID     string    `json:"trip_id"`
	DriverID   string    `json:"driver_id"`
	OccurredAt time.Time `json:"occurred_at"`
}

func newTripEvent(tripID, driverID uuid.UUID, eventName string) (Event, error) {
	idempotencyKey := tripID.String() + "_" + eventName
	eventID := uuid.NewSHA1(uuid.NameSpaceOID, []byte(idempotencyKey))

	kafkaEvent := TripEventPayload{
		EventID:    eventID.String(),
		EventType:  eventName,
		TripID:     tripID.String(),
		DriverID:   driverID.String(),
		OccurredAt: time.Now(),
	}

	payload, err := json.Marshal(kafkaEvent)
	if err != nil {
		return Event{}, err
	}

	return Event{
		ID:          eventID,
		EventName:   eventName,
		AggregateID: tripID,
		Payload:     payload,
	}, nil
}

func NewTripPublishedEvent(tripID, driverID uuid.UUID) (Event, error) {
	return newTripEvent(tripID, driverID, EventNameTripPublished)
}

func NewTripStartedEvent(tripID, driverID uuid.UUID) (Event, error) {
	return newTripEvent(tripID, driverID, EventNameTripStarted)
}

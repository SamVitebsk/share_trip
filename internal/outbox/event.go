package outbox

import (
	"go.opentelemetry.io/otel/trace"

	"context"

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
	TraceID    string    `json:"trace_id,omitempty"`
	SpanID     string    `json:"span_id,omitempty"`
}

type TripEventRequest struct {
	TripID   uuid.UUID
	DriverID uuid.UUID
}

func newTripEvent(ctx context.Context, req TripEventRequest, eventName string) (Event, error) {
	idempotencyKey := req.TripID.String() + "_" + eventName
	eventID := uuid.NewSHA1(uuid.NameSpaceOID, []byte(idempotencyKey))

	spanCtx := trace.SpanContextFromContext(ctx)
	var traceID, spanID string
	if spanCtx.HasTraceID() {
		traceID = spanCtx.TraceID().String()
	}
	if spanCtx.HasSpanID() {
		spanID = spanCtx.SpanID().String()
	}

	kafkaEvent := TripEventPayload{
		EventID:    eventID.String(),
		EventType:  eventName,
		TripID:     req.TripID.String(),
		DriverID:   req.DriverID.String(),
		OccurredAt: time.Now(),
		TraceID:    traceID,
		SpanID:     spanID,
	}

	payload, err := json.Marshal(kafkaEvent)
	if err != nil {
		return Event{}, err
	}

	return Event{
		ID:          eventID,
		EventName:   eventName,
		AggregateID: req.TripID,
		Payload:     payload,
	}, nil
}

func NewTripPublishedEvent(ctx context.Context, req TripEventRequest) (Event, error) {
	return newTripEvent(ctx, req, EventNameTripPublished)
}

func NewTripStartedEvent(ctx context.Context, req TripEventRequest) (Event, error) {
	return newTripEvent(ctx, req, EventNameTripStarted)
}

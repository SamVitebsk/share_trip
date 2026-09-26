package api_test

import (
	"context"
	"share_trip/internal/clients/kafka"
)

type mockEventPublisher struct{}

func (m *mockEventPublisher) SendEvent(ctx context.Context, event kafka.TripPublished) error {
	return nil
}

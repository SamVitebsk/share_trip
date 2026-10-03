package service_test

import (
	"context"
	"testing"
	"time"

	"share_trip/internal/domain"
	"share_trip/internal/observability/metrics"
	"share_trip/internal/service"
	"share_trip/internal/service/mocks"
	"share_trip/internal/storage/repository"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func setupService(ctrl *gomock.Controller) (*service.TripService, service.TripRepository, *mocks.MockContractChecker) {
	registry := prometheus.NewRegistry()
	appMetrics := metrics.New(registry)

	repo := repository.NewRepoPg(testPool, appMetrics)

	runTripTx := func(ctx context.Context, fn func(context.Context, service.TripRepositoryTx) error) error {
		return repo.WithinTripTx(ctx, func(ctx context.Context, trips *repository.TripRepoTx) error {
			return fn(ctx, trips)
		})
	}

	contractChecker := mocks.NewMockContractChecker(ctrl)

	svc := service.NewTripService(repo, runTripTx, appMetrics, contractChecker, nil)
	return svc, repo, contractChecker
}

func insertTestTrip(t *testing.T, ctx context.Context, repo service.TripRepository, tripID, driverID uuid.UUID) domain.Trip {
	trip := domain.Trip{
		ID:            tripID,
		DriverID:      driverID,
		FromPoint:     "Point A",
		ToPoint:       "Point B",
		DepartureTime: time.Now().Add(24 * time.Hour),
		Seats:         4,
		Status:        domain.TripStatusPublished,
		CreatedAt:     time.Now(),
	}

	history := domain.TripHistory{
		ID:        uuid.New(),
		TripID:    tripID,
		ToStatus:  domain.TripStatusPublished,
		CreatedAt: time.Now(),
	}

	err := repo.Create(ctx, trip, history)
	require.NoError(t, err)
	return trip
}

func TestService_StartTrip_Allowed(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, repo, contractChecker := setupService(ctrl)
	ctx := context.Background()

	tripID := uuid.New()
	driverID := uuid.New()

	insertedTrip := insertTestTrip(t, ctx, repo, tripID, driverID)

	contractChecker.EXPECT().
		CheckService(gomock.Any(), driverID.String(), "tripCreation").
		Return(service.CheckResult{Allowed: true, Reason: "service_allowed"}, nil)

	response, err := svc.StartTrip(ctx, service.StartTripRequest{
		TripID:   tripID,
		DriverID: driverID,
	})

	require.NoError(t, err)
	require.NotNil(t, response)

	expectedResponse := &service.StartTripResponse{
		ID:            insertedTrip.ID,
		DriverID:      insertedTrip.DriverID,
		FromPoint:     insertedTrip.FromPoint,
		ToPoint:       insertedTrip.ToPoint,
		DepartureTime: insertedTrip.DepartureTime,
		Seats:         insertedTrip.Seats,
		Status:        domain.TripStatusStarted,
		CreatedAt:     insertedTrip.CreatedAt,
	}

	require.WithinDuration(t, expectedResponse.DepartureTime, response.DepartureTime, time.Millisecond)
	require.WithinDuration(t, expectedResponse.CreatedAt, response.CreatedAt, time.Millisecond)

	expectedResponse.DepartureTime = response.DepartureTime
	expectedResponse.CreatedAt = response.CreatedAt

	require.Equal(t, expectedResponse, response)
}

func TestService_StartTrip_Denied(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, repo, contractChecker := setupService(ctrl)
	ctx := context.Background()

	tripID := uuid.New()
	driverID := uuid.New()

	insertTestTrip(t, ctx, repo, tripID, driverID)

	contractChecker.EXPECT().
		CheckService(gomock.Any(), driverID.String(), "tripCreation").
		Return(service.CheckResult{Allowed: false, Reason: "service_not_allowed"}, nil)

	response, err := svc.StartTrip(context.Background(), service.StartTripRequest{
		TripID:   tripID,
		DriverID: driverID,
	})

	require.Error(t, err)
	require.Nil(t, response)

	var appErr *service.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, service.CodeForbidden, appErr.Code)
}

func TestService_StartTrip_Timeout(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, repo, contractChecker := setupService(ctrl)
	ctx := context.Background()

	tripID := uuid.New()
	driverID := uuid.New()

	insertTestTrip(t, ctx, repo, tripID, driverID)

	contractChecker.EXPECT().
		CheckService(gomock.Any(), driverID.String(), "tripCreation").
		Return(service.CheckResult{}, context.DeadlineExceeded)

	response, err := svc.StartTrip(context.Background(), service.StartTripRequest{
		TripID:   tripID,
		DriverID: driverID,
	})

	require.Error(t, err)
	require.Nil(t, response)

	require.ErrorContains(t, err, "не удалось проверить права на старт поездки")
}

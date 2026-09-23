package service_test

import (
	"context"
	"share_trip/internal/domain"
	"share_trip/internal/service"
	"share_trip/internal/service/mocks"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestService_StartTrip_Allowed(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tripID := uuid.New()
	driverID := uuid.New()

	contractChecker := mocks.NewMockContractChecker(ctrl)
	repository := mocks.NewMockTripRepository(ctrl)
	repositoryTx := mocks.NewMockTripRepositoryTx(ctrl)

	trip := domain.Trip{
		ID:       tripID,
		DriverID: driverID,
		Status:   domain.TripStatusPublished,
	}
	repository.EXPECT().GetByID(gomock.Any(), tripID).Return(trip, nil)

	contractChecker.EXPECT().
		CheckService(gomock.Any(), driverID.String(), "tripCreation").
		Return(service.CheckResult{Allowed: true, Reason: "service_allowed"}, nil)

	repositoryTx.EXPECT().GetForUpdateByID(gomock.Any(), tripID).Return(trip, nil)
	repositoryTx.EXPECT().
		UpdateStatus(gomock.Any(), tripID, domain.TripStatusStarted).
		Return(domain.Trip{Status: domain.TripStatusStarted}, nil)
	repositoryTx.EXPECT().CreateHistory(gomock.Any(), gomock.Any()).Return(nil)
	repositoryTx.EXPECT().CreateOutboxEvent(gomock.Any(), gomock.Any()).Return(nil)

	runTripTx := func(ctx context.Context, fn func(context.Context, service.TripRepositoryTx) error) error {
		return fn(ctx, repositoryTx)
	}
	svc := service.NewTripService(repository, runTripTx, nil, contractChecker)

	response, err := svc.StartTrip(context.Background(), service.StartTripRequest{
		TripID:   tripID,
		DriverID: driverID,
	})

	require.NoError(t, err)
	require.NotNil(t, response)
	require.Equal(t, domain.TripStatusStarted, response.Status)
}

func TestService_StartTrip_Denied(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tripID := uuid.New()
	driverID := uuid.New()

	contractChecker := mocks.NewMockContractChecker(ctrl)
	repository := mocks.NewMockTripRepository(ctrl)
	repositoryTx := mocks.NewMockTripRepositoryTx(ctrl)

	trip := domain.Trip{
		ID:       tripID,
		DriverID: driverID,
		Status:   domain.TripStatusPublished,
	}

	repository.EXPECT().GetByID(gomock.Any(), tripID).Return(trip, nil)

	contractChecker.EXPECT().
		CheckService(gomock.Any(), driverID.String(), "tripCreation").
		Return(service.CheckResult{Allowed: false, Reason: "service_not_allowed"}, nil)

	runTripTx := func(ctx context.Context, fn func(context.Context, service.TripRepositoryTx) error) error {
		return fn(ctx, repositoryTx)
	}
	svc := service.NewTripService(repository, runTripTx, nil, contractChecker)

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

	tripID := uuid.New()
	driverID := uuid.New()

	contractChecker := mocks.NewMockContractChecker(ctrl)
	repository := mocks.NewMockTripRepository(ctrl)
	repositoryTx := mocks.NewMockTripRepositoryTx(ctrl)

	trip := domain.Trip{
		ID:       tripID,
		DriverID: driverID,
		Status:   domain.TripStatusPublished,
	}

	repository.EXPECT().GetByID(gomock.Any(), tripID).Return(trip, nil)

	contractChecker.EXPECT().
		CheckService(gomock.Any(), driverID.String(), "tripCreation").
		Return(service.CheckResult{}, context.DeadlineExceeded)

	runTripTx := func(ctx context.Context, fn func(context.Context, service.TripRepositoryTx) error) error {
		return fn(ctx, repositoryTx)
	}
	svc := service.NewTripService(repository, runTripTx, nil, contractChecker)

	response, err := svc.StartTrip(context.Background(), service.StartTripRequest{
		TripID:   tripID,
		DriverID: driverID,
	})

	require.Error(t, err)
	require.Nil(t, response)

	require.ErrorContains(t, err, "не удалось проверить права на старт поездки")
}

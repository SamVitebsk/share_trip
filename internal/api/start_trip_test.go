package api_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"share_trip/internal/api"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestServer_StartTrip(t *testing.T) {
	t.Run("успешный старт поездки", func(t *testing.T) {
		t.Parallel()

		tripID, driverID := insertTrip(t, "published")

		startResp := startTrip(t, tripID, driverID)
		defer func() {
			if err := startResp.Body.Close(); err != nil {
				t.Errorf("close start response body: %v", err)
			}
		}()

		require.Equal(t, http.StatusOK, startResp.StatusCode)

		actual := decodeStartTripResponse(t, startResp)

		expected := api.StartTripResponse{
			ID:            tripID.String(),
			DriverID:      driverID.String(),
			FromPoint:     "Минск",
			ToPoint:       "Гродно",
			DepartureTime: actual.DepartureTime,
			Seats:         3,
			Status:        "started",
			CreatedAt:     actual.CreatedAt,
		}

		require.Equal(t, expected, actual)
		require.Equal(t, 1, countTripEvents(t, tripID, "trip_started"))
	})

	t.Run("ошибка 403: доступ запрещен (другой водитель)", func(t *testing.T) {
		t.Parallel()

		tripID, _ := insertTrip(t, "published")
		startResp := startTrip(t, tripID, uuid.New())
		defer func() {
			if err := startResp.Body.Close(); err != nil {
				t.Errorf("close start response body: %v", err)
			}
		}()

		require.Equal(t, http.StatusForbidden, startResp.StatusCode)
		require.Equal(t, "published", getTripStatus(t, tripID))
		require.Equal(t, 0, countTripEvents(t, tripID, "trip_started"))
	})

	t.Run("ошибка 409: статус не позволяет начать поездку", func(t *testing.T) {
		t.Parallel()

		tripID, driverID := insertTrip(t, "draft")
		startResp := startTrip(t, tripID, driverID)
		defer func() {
			if err := startResp.Body.Close(); err != nil {
				t.Errorf("close start response body: %v", err)
			}
		}()

		require.Equal(t, http.StatusConflict, startResp.StatusCode)
		require.Equal(t, "draft", getTripStatus(t, tripID))
		require.Equal(t, 0, countTripEvents(t, tripID, "trip_started"))
	})

	t.Run("успех: поездка уже начата", func(t *testing.T) {
		t.Parallel()

		tripID, driverID := insertTrip(t, "started")

		_, err := testDB.Exec(
			`INSERT INTO outbox_event(
				id,
				event_name,
				aggregate_id,
				payload
			) VALUES ($1::uuid, 'trip_started', $2::uuid, jsonb_build_object('trip_id', $3::text))`,
			uuid.NewString(),
			tripID.String(),
			tripID.String(),
		)
		require.NoError(t, err)

		startResp := startTrip(t, tripID, driverID)
		defer func() {
			if err := startResp.Body.Close(); err != nil {
				t.Errorf("close start response body: %v", err)
			}
		}()

		require.Equal(t, http.StatusOK, startResp.StatusCode)
		actual := decodeStartTripResponse(t, startResp)

		expected := api.StartTripResponse{
			ID:            tripID.String(),
			DriverID:      driverID.String(),
			FromPoint:     "Минск",
			ToPoint:       "Гродно",
			DepartureTime: actual.DepartureTime,
			Seats:         3,
			Status:        "started",
			CreatedAt:     actual.CreatedAt,
		}

		require.Equal(t, expected, actual)
		require.Equal(t, "started", getTripStatus(t, tripID))
		require.Equal(t, 1, countTripEvents(t, tripID, "trip_started"))
	})
}

func startTrip(t *testing.T, tripID uuid.UUID, driverID uuid.UUID) *http.Response {
	t.Helper()

	startReq, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("/api/trip/%s/start", tripID),
		nil,
	)
	require.NoError(t, err)
	startReq.Header.Set(testAuthSubjectHeader, driverID.String())

	startResp, err := testApp.Test(startReq, -1)
	require.NoError(t, err)

	return startResp
}

func decodeStartTripResponse(t *testing.T, resp *http.Response) api.StartTripResponse {
	t.Helper()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var got api.StartTripResponse
	err = json.Unmarshal(respBody, &got)
	require.NoError(t, err)

	return got
}

func countTripEvents(t *testing.T, tripID uuid.UUID, eventName string) int {
	t.Helper()

	var count int
	err := testDB.QueryRow(
		`SELECT count(*)
		 FROM outbox_event
		 WHERE event_name = $1
		   AND aggregate_id = $2::uuid
		   AND payload->>'trip_id' = $3`,
		eventName,
		tripID.String(),
		tripID.String(),
	).Scan(&count)
	require.NoError(t, err)

	return count
}

func TestServer_StartTrip_Validation(t *testing.T) {
	t.Run("ошибка 400: пустой ID поездки", func(t *testing.T) {
		t.Parallel()

		startResp := startTrip(t, uuid.Nil, uuid.New())
		defer func() {
			if err := startResp.Body.Close(); err != nil {
				t.Errorf("close start response body: %v", err)
			}
		}()

		require.Equal(t, http.StatusBadRequest, startResp.StatusCode)
	})

	t.Run("ошибка 400: пустой ID водителя", func(t *testing.T) {
		t.Parallel()

		startResp := startTrip(t, uuid.New(), uuid.Nil)
		defer func() {
			if err := startResp.Body.Close(); err != nil {
				t.Errorf("close start response body: %v", err)
			}
		}()

		require.Equal(t, http.StatusBadRequest, startResp.StatusCode)
	})
}

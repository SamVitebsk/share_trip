package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"share_trip/internal/api"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestServer_PublishTrip(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()

		tripID, driverID := insertTrip(t, "draft")

		publishPayload := api.PublishTripRequest{
			DriverID: driverID.String(),
		}

		publishResp := publishTrip(t, tripID, publishPayload)
		defer func() {
			if err := publishResp.Body.Close(); err != nil {
				t.Errorf("close publish response body: %v", err)
			}
		}()

		require.Equal(t, http.StatusOK, publishResp.StatusCode)

		tripGot := decodePublishTripResponse(t, publishResp)

		require.Equal(t, tripID.String(), tripGot.ID)
		require.Equal(t, driverID.String(), tripGot.DriverID)
		require.Equal(t, "published", tripGot.Status)
		require.Equal(t, 1, countTripPublishedEvents(t, tripID))
	})

	t.Run("forbidden when driver mismatch", func(t *testing.T) {
		t.Parallel()

		tripID, _ := insertTrip(t, "draft")
		publishResp := publishTrip(t, tripID, api.PublishTripRequest{
			DriverID: uuid.NewString(),
		})
		defer func() {
			if err := publishResp.Body.Close(); err != nil {
				t.Errorf("close publish response body: %v", err)
			}
		}()

		require.Equal(t, http.StatusForbidden, publishResp.StatusCode)
		require.Equal(t, "draft", getTripStatus(t, tripID))
		require.Equal(t, 0, countTripPublishedEvents(t, tripID))
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()

		tripID := uuid.New()
		publishResp := publishTrip(t, tripID, api.PublishTripRequest{
			DriverID: uuid.NewString(),
		})
		defer func() {
			if err := publishResp.Body.Close(); err != nil {
				t.Errorf("close publish response body: %v", err)
			}
		}()

		require.Equal(t, http.StatusNotFound, publishResp.StatusCode)
		require.Equal(t, 0, countTripPublishedEvents(t, tripID))
	})

	t.Run("conflict when status does not allow publishing", func(t *testing.T) {
		t.Parallel()

		tripID, driverID := insertTrip(t, "canceled")
		publishResp := publishTrip(t, tripID, api.PublishTripRequest{
			DriverID: driverID.String(),
		})
		defer func() {
			if err := publishResp.Body.Close(); err != nil {
				t.Errorf("close publish response body: %v", err)
			}
		}()

		require.Equal(t, http.StatusConflict, publishResp.StatusCode)
		require.Equal(t, "canceled", getTripStatus(t, tripID))
		require.Equal(t, 0, countTripPublishedEvents(t, tripID))
	})

	t.Run("already published", func(t *testing.T) {
		t.Parallel()

		tripID, driverID := insertTrip(t, "published")

		_, err := testDB.Exec(
			`INSERT INTO outbox_event(
				id,
				event_name,
				aggregate_id,
				payload
			) VALUES ($1::uuid, 'trip_published', $2::uuid, jsonb_build_object('trip_id', $3::text))`,
			uuid.NewString(),
			tripID.String(),
			tripID.String(),
		)
		require.NoError(t, err)

		publishPayload := api.PublishTripRequest{
			DriverID: driverID.String(),
		}

		publishResp := publishTrip(t, tripID, publishPayload)
		defer func() {
			if err := publishResp.Body.Close(); err != nil {
				t.Errorf("close publish response body: %v", err)
			}
		}()

		require.Equal(t, http.StatusOK, publishResp.StatusCode)
		tripGot := decodePublishTripResponse(t, publishResp)

		require.Equal(t, tripID.String(), tripGot.ID)
		require.Equal(t, driverID.String(), tripGot.DriverID)
		require.Equal(t, "published", tripGot.Status)
		require.Equal(t, "published", getTripStatus(t, tripID))
		require.Equal(t, 1, countTripPublishedEvents(t, tripID))
	})
}

func publishTrip(t *testing.T, tripID uuid.UUID, payload api.PublishTripRequest) *http.Response {
	t.Helper()

	publishBody, err := json.Marshal(payload)
	require.NoError(t, err)

	publishReq, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("/api/trip/%s/publish", tripID),
		bytes.NewReader(publishBody),
	)
	require.NoError(t, err)
	publishReq.Header.Set("Content-Type", "application/json")

	publishResp, err := testApp.Test(publishReq, -1)
	require.NoError(t, err)

	return publishResp
}

func decodePublishTripResponse(t *testing.T, resp *http.Response) api.PublishTripResponse {
	t.Helper()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var got api.PublishTripResponse
	err = json.Unmarshal(respBody, &got)
	require.NoError(t, err)

	return got
}

func insertTrip(t *testing.T, status string) (uuid.UUID, uuid.UUID) {
	t.Helper()

	driverID := uuid.New()
	tripID := uuid.New()
	_, err := testDB.Exec(
		`INSERT INTO trips(
			id,
			driver_id,
			from_point,
			to_point,
			departure_time,
			seats,
			status
		) VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7)`,
		tripID.String(),
		driverID.String(),
		"Минск",
		"Гродно",
		time.Now().Add(24*time.Hour).UTC(),
		3,
		status,
	)
	require.NoError(t, err)

	return tripID, driverID
}

func getTripStatus(t *testing.T, tripID uuid.UUID) string {
	t.Helper()

	var status string
	err := testDB.QueryRow(
		`SELECT status
		 FROM trips
		 WHERE id = $1::uuid`,
		tripID.String(),
	).Scan(&status)
	require.NoError(t, err)

	return status
}

func countTripPublishedEvents(t *testing.T, tripID uuid.UUID) int {
	t.Helper()

	var count int
	err := testDB.QueryRow(
		`SELECT count(*)
		 FROM outbox_event
		 WHERE event_name = 'trip_published'
		   AND aggregate_id = $1::uuid
		   AND payload->>'trip_id' = $2`,
		tripID.String(),
		tripID.String(),
	).Scan(&count)
	require.NoError(t, err)

	return count
}

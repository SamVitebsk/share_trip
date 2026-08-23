package api_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"share_trip/internal/api"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestServer_CreateTrip(t *testing.T) {
	t.Run("success - создание поездки", func(t *testing.T) {
		truncateTestTables(t)

		driverID := uuid.NewString()
		payload := api.CreateTripRequest{
			DriverID:      driverID,
			FromPoint:     "Минск",
			ToPoint:       "Витебск",
			DepartureTime: time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339),
			Seats:         3,
		}

		body, err := json.Marshal(payload)
		require.NoError(t, err)

		req, err := http.NewRequest(
			http.MethodPost,
			"/api/trip/create",
			bytes.NewReader(body),
		)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := testApp.Test(req, -1)
		require.NoError(t, err)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("close response body: %v", err)
			}
		}()

		require.Equal(t, http.StatusCreated, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var got api.CreateTripResponse
		err = json.Unmarshal(respBody, &got)
		require.NoError(t, err)

		require.NotEmpty(t, got.ID)

		_, err = uuid.Parse(got.ID)
		require.NoError(t, err)
	})
}

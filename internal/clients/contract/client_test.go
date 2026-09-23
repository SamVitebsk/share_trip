package contract_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	config "share_trip/configs"
	"share_trip/internal/clients/contract"
	"share_trip/internal/clients/contract/gen"
	"share_trip/internal/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestClient_CheckService(t *testing.T) {
	t.Run("успешный ответ", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := contractgen.CheckServiceAvailabilityResponse{
				Allowed: true,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		cfg := config.ContractConfig{
			BaseURL:    server.URL,
			Timeout:    1 * time.Second,
			RetryCount: 0,
		}
		client, err := contract.NewClient(cfg)
		require.NoError(t, err)

		res, err := client.CheckService(context.Background(), uuid.New().String(), "tripCreation")
		require.NoError(t, err)

		expected := service.CheckResult{
			Allowed: true,
			Reason:  "",
		}
		require.Equal(t, expected, res)
	})

	t.Run("ошибка 403: доступ запрещен", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reasonStr := "contractExpired"
			reasonEnum := contractgen.CheckServiceAvailabilityResponseReason(reasonStr)

			resp := contractgen.CheckServiceAvailabilityResponse{
				Allowed: false,
				Reason:  &reasonEnum,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		cfg := config.ContractConfig{
			BaseURL:    server.URL,
			Timeout:    1 * time.Second,
			RetryCount: 0,
		}
		client, err := contract.NewClient(cfg)
		require.NoError(t, err)

		res, err := client.CheckService(context.Background(), uuid.New().String(), "tripCreation")
		require.NoError(t, err)

		expected := service.CheckResult{
			Allowed: false,
			Reason:  "contractExpired",
		}
		require.Equal(t, expected, res)
	})

	t.Run("внутренняя ошибка 500 и проверка Retry", func(t *testing.T) {
		t.Parallel()

		var attempts int

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempts++
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		cfg := config.ContractConfig{
			BaseURL:    server.URL,
			Timeout:    2 * time.Second,
			RetryCount: 2,
		}
		client, err := contract.NewClient(cfg)
		require.NoError(t, err)

		_, err = client.CheckService(context.Background(), uuid.New().String(), "tripCreation")

		require.Error(t, err)
		require.Equal(t, 3, attempts)
	})
}

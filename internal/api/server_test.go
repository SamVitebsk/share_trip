package api_test

import (
	"context"
	"database/sql"
	"log"
	"os"
	"testing"

	"share_trip/internal/api"
	"share_trip/internal/api/middleware"
	"share_trip/internal/observability/metrics"
	"share_trip/internal/service"
	"share_trip/internal/storage/repository"
	"share_trip/internal/test"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
)

const testKeycloakClientID = "sharetrip-api"
const testAuthSubjectHeader = "X-Test-Auth-Subject"

var (
	testCtx  context.Context
	testPool *pgxpool.Pool
	testDB   *sql.DB
	testApp  *fiber.App
)

func TestMain(m *testing.M) {
	testCtx = context.Background()

	var err error
	var teardown func()

	testPool, testDB, teardown, err = test.SetupTestDB(testCtx, "../../migrations")
	if err != nil {
		log.Fatalf("failed to setup test database: %v", err)
	}
	defer teardown()

	registry := prometheus.NewRegistry()
	appMetrics := metrics.New(registry)
	repo := repository.NewRepoPg(testPool, appMetrics)
	runTripTx := func(ctx context.Context, fn func(context.Context, service.TripRepositoryTx) error) error {
		return repo.WithinTripTx(ctx, func(ctx context.Context, trips *repository.TripRepoTx) error {
			return fn(ctx, trips)
		})
	}
	mockChecker := &mockContractChecker{allowed: true}
	tripService := service.NewTripService(repo, runTripTx, appMetrics, mockChecker, &mockEventPublisher{})
	tripHandler := api.NewTripHandler(tripService)
	readyHandler := api.NewReadyHandler(repo)
	server := api.NewServer(tripHandler, readyHandler)

	testApp = fiber.New()
	server.Route(testApp.Group("/api"), testAuthMiddleware, testKeycloakClientID)

	os.Exit(m.Run())
}

func testAuthMiddleware(c *fiber.Ctx) error {
	subject := c.Get(testAuthSubjectHeader)
	if subject == "" {
		subject = uuid.NewString()
	}

	c.Locals(middleware.KeycloakClaimsKey, &middleware.KeycloakClaims{
		Subject: subject,
		ResourceAccess: map[string]struct {
			Roles []string `json:"roles"`
		}{
			testKeycloakClientID: {
				Roles: []string{"client"},
			},
		},
	})

	return c.Next()
}

type mockContractChecker struct {
	allowed bool
	err     error
	reason  string
}

func (m *mockContractChecker) CheckService(_ context.Context, _ string, _ string) (service.CheckResult, error) {
	return service.CheckResult{
		Allowed: m.allowed,
		Reason:  m.reason,
	}, m.err
}

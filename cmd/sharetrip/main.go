package main

import (
	"context"
	"log"
	"os"
	"share_trip/internal/app"
	"share_trip/internal/observability/metrics"
	"share_trip/internal/observability/tracing"
	"time"

	config "share_trip/configs"
	"share_trip/internal/api"
	"share_trip/internal/api/middleware"
	"share_trip/internal/service"
	"share_trip/internal/storage/postgres"
	"share_trip/internal/storage/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	logger, logFile, err := app.NewLogger()
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := logFile.Close(); err != nil {
			logger.Error("закрытие файла логов не выполнено", "error", err)
		}
	}()

	ctx := context.Background()

	tracerProvider, err := tracing.NewProvider(ctx, tracing.Config{
		ServiceName:    "share-trip",
		ServiceVersion: "1.0.0",
		Environment:    "local",
		Endpoint:       "localhost:4319",
	})
	if err != nil {
		logger.Error("инициализация трассировки не выполнена", "error", err)
		os.Exit(1)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancel()

		if err := tracerProvider.Shutdown(shutdownCtx); err != nil {
			logger.Error("shutdown tracing failed", "error", err)
		}
	}()

	cfg := config.PostgresConfig{
		Host:     config.Env("DB_HOST", "localhost"),
		Port:     config.EnvInt("DB_PORT", 6543),
		User:     config.Env("DB_USER", "postgres"),
		Password: config.Env("DB_PASSWORD", "admin"),
		DBName:   config.Env("DB_NAME", "share_trip"),
		SSLMode:  config.Env("DB_SSLMODE", "disable"),
	}

	pool, err := postgres.NewPool(ctx, cfg.DSN())
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	registry := prometheus.NewRegistry()
	appMetrics := metrics.New(registry)

	repo := repository.NewRepoPg(pool, appMetrics)
	runTripTx := func(ctx context.Context, fn func(context.Context, service.TripRepositoryTx) error) error {
		return repo.WithinTripTx(ctx, func(ctx context.Context, trips *repository.TripRepoTx) error {
			return fn(ctx, trips)
		})
	}
	tripService := service.NewTripService(repo, runTripTx, appMetrics)
	tripHandler := api.NewTripHandler(tripService)
	readyHandler := api.NewReadyHandler(repo)

	server := api.NewServer(tripHandler, readyHandler)

	fiberApp := fiber.New()

	fiberApp.Use(middleware.Correlation(logger))
	fiberApp.Use(tracing.NewFiberMiddleware())
	fiberApp.Use(middleware.NewHTTPMetricsMiddleware(appMetrics))

	keycloakClientID := config.Env("KEYCLOAK_CLIENT_ID", "sharetrip-api")
	keycloakAuthMiddleware := middleware.KeycloakRefreshTokenMiddleware(
		middleware.KeycloakConfig{
			Issuer:       config.Env("KEYCLOAK_ISSUER", "http://localhost:8087/realms/sharetrip"),
			ClientID:     keycloakClientID,
			ClientSecret: config.Env("KEYCLOAK_CLIENT_SECRET", "kcTclgACcVx4ozusKmvvihUqARRE4OnI"),
		},
	)

	fiberApp.Get("/metrics", adaptor.HTTPHandler(promhttp.HandlerFor(registry, promhttp.HandlerOpts{})))
	server.Route(fiberApp.Group("/api"), keycloakAuthMiddleware, keycloakClientID)

	err = fiberApp.Listen(config.Env("SERVER_PORT", ":9090"))
	if err != nil {
		log.Fatal(err)
	}
}

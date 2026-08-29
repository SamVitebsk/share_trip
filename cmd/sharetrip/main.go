package main

import (
	"context"
	"log"
	"share_trip/internal/app"
	"share_trip/internal/observability/metrics"

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
	fiberApp.Use(middleware.NewHTTPMetricsMiddleware(appMetrics))

	fiberApp.Get("/metrics", adaptor.HTTPHandler(promhttp.HandlerFor(registry, promhttp.HandlerOpts{})))
	server.Route(fiberApp.Group("/api"))

	err = fiberApp.Listen(config.Env("SERVER_PORT", ":9090"))
	if err != nil {
		log.Fatal(err)
	}
}

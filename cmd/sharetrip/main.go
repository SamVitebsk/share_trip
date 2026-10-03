package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"os"
	"share_trip/internal/app"
	"share_trip/internal/clients/contract"
	"share_trip/internal/clients/kafka"
	"share_trip/internal/observability/metrics"
	"share_trip/internal/observability/tracing"
	"time"

	config "share_trip/configs"
	"share_trip/internal/api"
	"share_trip/internal/api/middleware"
	"share_trip/internal/outbox/publisher"
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

	cfg, err := config.Load()
	if err != nil {
		logger.Error("ошибка инициализации конфигурации", "error", err)
		os.Exit(1)
	}

	contractClient, err := contract.NewClient(cfg.Contract)
	if err != nil {
		log.Fatal(err)
	}

	pool, err := postgres.NewPool(ctx, cfg.Database.DSN)
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
	producer := kafka.NewProducer(cfg.Kafka.Brokers, cfg.Kafka.Topic)
	tripService := service.NewTripService(repo, runTripTx, appMetrics, contractClient)
	tripHandler := api.NewTripHandler(tripService)
	readyHandler := api.NewReadyHandler(repo)

	outboxConfig := cfg.Outbox
	relayRepo := repository.NewOutboxRelayRepo(pool)
	outboxPublisher := publisher.NewPublisher(relayRepo, producer, logger, outboxConfig)
	go func() {
		logger.Info("запуск фонового Outbox Publisher relay...")
		if err := outboxPublisher.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
			logger.Error("фоновый Outbox Publisher relay завершил работу с ошибкой", slog.Any("error", err))
		} else {
			logger.Info("фоновый Outbox Publisher relay успешно остановлен")
		}
	}()

	server := api.NewServer(tripHandler, readyHandler)

	fiberApp := fiber.New()

	fiberApp.Use(middleware.Correlation(logger))
	fiberApp.Use(tracing.NewFiberMiddleware())
	fiberApp.Use(middleware.NewHTTPMetricsMiddleware(appMetrics))

	keycloakClientID := cfg.KeycloakClientID
	keycloakAuthMiddleware := middleware.KeycloakRefreshTokenMiddleware(
		middleware.KeycloakConfig{
			Issuer:       cfg.KeycloakIssuer,
			ClientID:     keycloakClientID,
			ClientSecret: cfg.KeycloakClientSecret,
		},
	)

	fiberApp.Get("/metrics", adaptor.HTTPHandler(promhttp.HandlerFor(registry, promhttp.HandlerOpts{})))
	server.Route(fiberApp.Group("/api"), keycloakAuthMiddleware, keycloakClientID)

	err = fiberApp.Listen(":" + cfg.HTTPPort)
	if err != nil {
		log.Fatal(err)
	}
}

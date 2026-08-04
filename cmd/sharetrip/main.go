package main

import (
	"context"
	"log"
	config "share_trip/configs"
	"share_trip/internal/api"
	"share_trip/internal/storage/postgres"
	"share_trip/internal/storage/repository"

	"github.com/gofiber/fiber/v2"
)

func main() {
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

	repo := repository.NewRepoPg(pool)

	server := api.NewServer(repo)

	app := fiber.New()
	server.Route(app.Group("/api"))

	err = app.Listen(config.Env("SERVER_PORT", ":9090"))
	if err != nil {
		log.Fatal(err)
	}
}

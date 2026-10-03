package test

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func SetupTestDB(ctx context.Context, migrationsPath string) (*pgxpool.Pool, *sql.DB, func(), error) {
	container, err := postgres.Run(
		ctx,
		"postgres:16",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("password"),
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("start postgres container: %w", err)
	}

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, nil, nil, fmt.Errorf("get connection string: %w", err)
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, nil, nil, fmt.Errorf("open sql db: %w", err)
	}

	if err := waitReady(db); err != nil {
		_ = db.Close()
		_ = container.Terminate(ctx)
		return nil, nil, nil, fmt.Errorf("database not ready: %w", err)
	}

	if err = goose.SetDialect("postgres"); err != nil {
		_ = db.Close()
		_ = container.Terminate(ctx)
		return nil, nil, nil, fmt.Errorf("set goose dialect: %w", err)
	}

	if err = goose.Up(db, migrationsPath); err != nil {
		_ = db.Close()
		_ = container.Terminate(ctx)
		return nil, nil, nil, fmt.Errorf("run migrations: %w", err)
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		_ = db.Close()
		_ = container.Terminate(ctx)
		return nil, nil, nil, fmt.Errorf("create pgx pool: %w", err)
	}

	teardown := func() {
		pool.Close()
		_ = db.Close()
		_ = container.Terminate(context.Background())
	}

	return pool, db, teardown, nil
}

func waitReady(db *sql.DB) error {
	deadline := time.Now().Add(30 * time.Second)

	for time.Now().Before(deadline) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		err := db.PingContext(ctx)
		cancel()

		if err == nil {
			return nil
		}

		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("database is not ready after timeout")
}

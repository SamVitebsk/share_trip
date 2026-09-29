package config

import (
	"fmt"
	"os"
)

type PostgresConfig struct {
	DSN string
}

func LoadPostgresConfig() (PostgresConfig, error) {
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		return PostgresConfig{}, fmt.Errorf("DATABASE_DSN is required")
	}

	return PostgresConfig{
		DSN: dsn,
	}, nil
}

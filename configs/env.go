package config

import (
	"fmt"
	"os"
	"strconv"
)

func getenv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getenvInt(key string, fallback string) (int, error) {
	strVal := getenv(key, fallback)
	n, err := strconv.Atoi(strVal)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	return n, nil
}

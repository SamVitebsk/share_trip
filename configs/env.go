package config

import (
	"os"
	"strconv"
	"time"
)

func Env(key, def string) string {
	val := os.Getenv(key)
	if val == "" {
		return def
	}

	return val
}

func EnvInt(key string, def int) int {
	val := os.Getenv(key)
	if val == "" {
		return def
	}

	n, err := strconv.Atoi(val)
	if err != nil {
		return def
	}

	return n
}

func EnvDuration(key string, def time.Duration) time.Duration {
	val := os.Getenv(key)
	if val == "" {
		return def
	}

	d, err := time.ParseDuration(val)
	if err != nil {
		return def
	}

	return d
}

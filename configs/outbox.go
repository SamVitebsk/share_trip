package config

import (
	"os"
	"strconv"
	"time"
)

type OutboxConfig struct {
	PollInterval time.Duration
	BatchSize    int
}

func LoadOutboxConfig() OutboxConfig {
	intervalStr := os.Getenv("OUTBOX_POLL_INTERVAL_MS")
	pollMs, err := strconv.Atoi(intervalStr)
	if err != nil || pollMs <= 0 {
		pollMs = 2000
	}

	batchStr := os.Getenv("OUTBOX_BATCH_SIZE")
	batchSize, err := strconv.Atoi(batchStr)
	if err != nil || batchSize <= 0 {
		batchSize = 100
	}

	return OutboxConfig{
		PollInterval: time.Duration(pollMs) * time.Millisecond,
		BatchSize:    batchSize,
	}
}

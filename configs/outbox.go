package config

import (
	"time"
)

type OutboxConfig struct {
	PollInterval time.Duration
	BatchSize    int
}

func LoadOutboxConfig() (OutboxConfig, error) {
	pollIntervalMS, err := getenvInt("OUTBOX_POLL_INTERVAL_MS", "2000")
	if err != nil {
		return OutboxConfig{}, err
	}

	batchSize, err := getenvInt("OUTBOX_BATCH_SIZE", "100")
	if err != nil {
		return OutboxConfig{}, err
	}

	return OutboxConfig{
		PollInterval: time.Duration(pollIntervalMS) * time.Millisecond,
		BatchSize:    batchSize,
	}, nil
}

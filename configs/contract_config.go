package config

import "time"

type ContractConfig struct {
	BaseURL    string
	Timeout    time.Duration
	RetryCount int
}

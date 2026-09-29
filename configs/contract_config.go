package config

import (
	"fmt"
	"os"
	"time"
)

type ContractConfig struct {
	BaseURL    string
	Timeout    time.Duration
	RetryCount int
}

func LoadContractConfig() (ContractConfig, error) {
	url := os.Getenv("CONTRACT_SERVICE_URL")
	if url == "" {
		return ContractConfig{}, fmt.Errorf("CONTRACT_SERVICE_URL is required")
	}

	timeoutMS, err := getenvInt("REQUEST_TIMEOUT_MS", "1500")
	if err != nil {
		return ContractConfig{}, err
	}

	retryCount, err := getenvInt("RETRY_ATTEMPTS", "2")
	if err != nil {
		return ContractConfig{}, err
	}

	return ContractConfig{
		BaseURL:    url,
		Timeout:    time.Duration(timeoutMS) * time.Millisecond,
		RetryCount: retryCount,
	}, nil
}

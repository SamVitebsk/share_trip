package repository

import (
	"errors"
)

const (
	repositoryMetricResultSuccess  = "success"
	repositoryMetricResultNotFound = "not_found"
	repositoryMetricResultError    = "error"
)

func repositoryMetricResultFromError(err error) string {
	if err == nil {
		return repositoryMetricResultSuccess
	}
	if errors.Is(err, ErrNotFound) {
		return repositoryMetricResultNotFound
	}
	return repositoryMetricResultError
}

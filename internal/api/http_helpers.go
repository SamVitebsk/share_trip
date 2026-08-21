package api

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

func parseUUIDIfPresent(value string, invalidErr error) (uuid.UUID, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return uuid.Nil, nil
	}

	parsed, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, invalidErr
	}

	return parsed, nil
}

func parseRFC3339TimeIfPresent(value string, invalidErr error) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, nil
	}

	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, invalidErr
	}

	return parsed, nil
}

func formatHTTPTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}

	return value.Format(time.RFC3339Nano)
}

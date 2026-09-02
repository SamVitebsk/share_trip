package api

import (
	"strings"
	"time"
)

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

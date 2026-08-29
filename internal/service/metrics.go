package service

import "errors"

const (
	metricResultSuccess         = "success"
	metricResultValidationError = "validation_error"
	metricResultNotFound        = "not_found"
	metricResultForbidden       = "forbidden"
	metricResultConflict        = "conflict"
	metricResultInternalError   = "internal_error"
)

func metricResultFromError(err error) string {
	if err == nil {
		return metricResultSuccess
	}

	var validationErr *ValidationError
	if errors.As(err, &validationErr) {
		return metricResultValidationError
	}

	var appErr *AppError
	if !errors.As(err, &appErr) {
		return metricResultInternalError
	}

	switch appErr.Code {
	case CodeValidation:
		return metricResultValidationError
	case CodeNotFound:
		return metricResultNotFound
	case CodeForbidden:
		return metricResultForbidden
	case CodeConflict:
		return metricResultConflict
	default:
		return metricResultInternalError
	}
}

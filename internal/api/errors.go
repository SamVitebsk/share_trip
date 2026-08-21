package api

import (
	"errors"

	"share_trip/internal/service"

	"github.com/gofiber/fiber/v2"
)

type ErrorCode string

const (
	ErrorCodeValidation    ErrorCode = "VALIDATION_ERROR"
	ErrorCodeNotFound      ErrorCode = "NOT_FOUND"
	ErrorCodeForbidden     ErrorCode = "FORBIDDEN"
	ErrorCodeConflict      ErrorCode = "CONFLICT"
	ErrorCodeUnprocessable ErrorCode = "UNPROCESSABLE"
	ErrorCodeInternal      ErrorCode = "INTERNAL_ERROR"
)

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    ErrorCode    `json:"code"`
	Message string       `json:"message"`
	Fields  []FieldError `json:"fields,omitempty"`
}

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func writeError(c *fiber.Ctx, code ErrorCode, message string) error {
	return writeErrorWithFields(c, code, message, nil)
}

func writeErrorWithFields(c *fiber.Ctx, code ErrorCode, message string, fields []FieldError) error {
	return c.Status(statusByErrorCode(code)).JSON(ErrorResponse{
		Error: ErrorBody{
			Code:    code,
			Message: message,
			Fields:  fields,
		},
	})
}

func writeServiceError(c *fiber.Ctx, err error) error {
	var validationErr *service.ValidationError
	if errors.As(err, &validationErr) {
		return writeErrorWithFields(
			c,
			ErrorCodeValidation,
			validationErr.Error(),
			fieldErrorsFromValidationError(validationErr),
		)
	}

	var appErr *service.AppError
	if errors.As(err, &appErr) {
		code := errorCodeFromServiceCode(appErr.Code)
		if code == ErrorCodeInternal {
			return writeError(c, code, "внутренняя ошибка сервера")
		}

		return writeError(c, code, appErr.Error())
	}

	return writeError(c, ErrorCodeInternal, "внутренняя ошибка сервера")
}

func errorCodeFromServiceCode(code service.ErrorCode) ErrorCode {
	switch code {
	case service.CodeValidation:
		return ErrorCodeValidation
	case service.CodeNotFound:
		return ErrorCodeNotFound
	case service.CodeForbidden:
		return ErrorCodeForbidden
	case service.CodeConflict:
		return ErrorCodeConflict
	case service.CodeUnprocessable:
		return ErrorCodeUnprocessable
	default:
		return ErrorCodeInternal
	}
}

func fieldErrorsFromValidationError(validationErr *service.ValidationError) []FieldError {
	fields := make([]FieldError, 0, len(validationErr.Fields))
	for _, field := range validationErr.Fields {
		fields = append(fields, FieldError{
			Field:   field.Field,
			Message: field.Message,
		})
	}

	return fields
}

func statusByErrorCode(code ErrorCode) int {
	switch code {
	case ErrorCodeValidation:
		return fiber.StatusBadRequest
	case ErrorCodeNotFound:
		return fiber.StatusNotFound
	case ErrorCodeForbidden:
		return fiber.StatusForbidden
	case ErrorCodeConflict:
		return fiber.StatusConflict
	case ErrorCodeUnprocessable:
		return fiber.StatusUnprocessableEntity
	case ErrorCodeInternal:
		return fiber.StatusInternalServerError
	default:
		return fiber.StatusInternalServerError
	}
}

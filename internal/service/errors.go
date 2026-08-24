package service

import (
	"fmt"
)

type ErrorCode int

const (
	CodeUnknown       ErrorCode = 0
	CodeValidation    ErrorCode = 1
	CodeNotFound      ErrorCode = 2
	CodeForbidden     ErrorCode = 3
	CodeConflict      ErrorCode = 4
	CodeUnprocessable ErrorCode = 5
)

const validationErrorMessage = "ошибка валидации"

type AppError struct {
	Code    ErrorCode
	Message string
}

type FieldError struct {
	Field   string
	Message string
}

type ValidationError struct {
	Fields []FieldError
}

func (e *AppError) Error() string {
	return e.Message
}

func (e FieldError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func (e *ValidationError) Error() string {
	return validationErrorMessage
}

func NewAppError(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

func Validation(message string) *AppError {
	return NewAppError(CodeValidation, message)
}

func ValidationWithFields(fields ...FieldError) *ValidationError {
	return &ValidationError{
		Fields: fields,
	}
}

func NotFound(message string) *AppError {
	return NewAppError(CodeNotFound, message)
}

func Forbidden(message string) *AppError {
	return NewAppError(CodeForbidden, message)
}

func Conflict(message string) *AppError {
	return NewAppError(CodeConflict, message)
}

func Unprocessable(message string) *AppError {
	return NewAppError(CodeUnprocessable, message)
}

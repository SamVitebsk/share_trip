package domain

import "errors"

var (
	ErrTripDriverMismatch          = errors.New("водитель не является владельцем поездки")
	ErrTripInvalidStatusTransition = errors.New("недопустимый переход статуса поездки")
)

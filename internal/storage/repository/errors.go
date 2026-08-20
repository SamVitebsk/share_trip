package repository

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

const (
	postgresUniqueViolation     = "23505"
	postgresForeignKeyViolation = "23503"
	postgresNotNullViolation    = "23502"
	postgresCheckViolation      = "23514"
)

var (
	ErrNotFound            = errors.New("not found")
	ErrAlreadyExists       = errors.New("already exists")
	ErrInvalidReference    = errors.New("invalid reference")
	ErrConstraintViolation = errors.New("constraint violation")
)

func mapPostgresError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}

	switch pgErr.Code {
	case postgresUniqueViolation:
		return errors.Join(ErrAlreadyExists, err)
	case postgresForeignKeyViolation:
		return errors.Join(ErrInvalidReference, err)
	case postgresNotNullViolation, postgresCheckViolation:
		return errors.Join(ErrConstraintViolation, err)
	default:
		return err
	}
}

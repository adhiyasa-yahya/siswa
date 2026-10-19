// Package repository contains the PostgreSQL data access layer.
package repository

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// ErrNotFound is returned when a query matches no row. Services translate it
// into the public 1002 / HTTP 404 error.
var ErrNotFound = errors.New("record not found")

// PostgreSQL error codes used by the repositories.
const (
	pgUniqueViolation     = "23505"
	pgForeignKeyViolation = "23503"
	pgCheckViolation      = "23514"
)

func pgErrorCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}

func pgConstraint(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.ConstraintName
	}
	return ""
}

// IsUniqueViolation reports whether err is a unique index violation, optionally
// restricted to a specific constraint name.
func IsUniqueViolation(err error, constraint string) bool {
	if pgErrorCode(err) != pgUniqueViolation {
		return false
	}
	return constraint == "" || pgConstraint(err) == constraint
}

// IsForeignKeyViolation reports whether err is a foreign key violation.
func IsForeignKeyViolation(err error) bool {
	return pgErrorCode(err) == pgForeignKeyViolation
}

// IsCheckViolation reports whether err is a CHECK constraint violation (for
// example the stocks.qty >= 0 guard).
func IsCheckViolation(err error, constraint string) bool {
	if pgErrorCode(err) != pgCheckViolation {
		return false
	}
	return constraint == "" || pgConstraint(err) == constraint
}

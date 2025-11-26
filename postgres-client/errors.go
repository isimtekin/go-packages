package postgresclient

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Common errors
var (
	ErrNilConfig          = errors.New("config cannot be nil")
	ErrClientNotConnected = errors.New("client is not connected")
	ErrNoRows             = pgx.ErrNoRows
	ErrTxClosed           = pgx.ErrTxClosed
	ErrTxCommitRollback   = pgx.ErrTxCommitRollback
)

// PostgreSQL error codes
const (
	// Class 23 - Integrity Constraint Violation
	UniqueViolation     = "23505"
	ForeignKeyViolation = "23503"
	NotNullViolation    = "23502"
	CheckViolation      = "23514"

	// Class 42 - Syntax Error or Access Rule Violation
	UndefinedTable  = "42P01"
	UndefinedColumn = "42703"

	// Class 53 - Insufficient Resources
	DiskFull           = "53100"
	OutOfMemory        = "53200"
	TooManyConnections = "53300"
)

// IsNoRows checks if the error is a "no rows" error
func IsNoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

// IsUniqueViolation checks if the error is a unique constraint violation
func IsUniqueViolation(err error) bool {
	return isPgError(err, UniqueViolation)
}

// IsForeignKeyViolation checks if the error is a foreign key constraint violation
func IsForeignKeyViolation(err error) bool {
	return isPgError(err, ForeignKeyViolation)
}

// IsNotNullViolation checks if the error is a NOT NULL constraint violation
func IsNotNullViolation(err error) bool {
	return isPgError(err, NotNullViolation)
}

// IsCheckViolation checks if the error is a CHECK constraint violation
func IsCheckViolation(err error) bool {
	return isPgError(err, CheckViolation)
}

// IsConnectionError checks if the error is a connection-related error
func IsConnectionError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		// Class 08 - Connection Exception
		return len(pgErr.Code) >= 2 && pgErr.Code[:2] == "08"
	}
	return false
}

// IsTooManyConnections checks if the error is due to too many connections
func IsTooManyConnections(err error) bool {
	return isPgError(err, TooManyConnections)
}

// GetPgErrorCode returns the PostgreSQL error code if available
func GetPgErrorCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}

// GetPgErrorMessage returns the PostgreSQL error message if available
func GetPgErrorMessage(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Message
	}
	return err.Error()
}

// GetPgErrorDetail returns the PostgreSQL error detail if available
func GetPgErrorDetail(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Detail
	}
	return ""
}

// isPgError checks if the error matches a specific PostgreSQL error code
func isPgError(err error, code string) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == code
	}
	return false
}

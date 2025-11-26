package postgresclient

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestIsNoRows(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "no rows error",
			err:      pgx.ErrNoRows,
			expected: true,
		},
		{
			name:     "wrapped no rows error",
			err:      errors.New("wrapped: " + pgx.ErrNoRows.Error()),
			expected: false,
		},
		{
			name:     "other error",
			err:      errors.New("some other error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNoRows(tt.err)
			if result != tt.expected {
				t.Errorf("IsNoRows() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestIsUniqueViolation(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name: "unique violation error",
			err: &pgconn.PgError{
				Code: UniqueViolation,
			},
			expected: true,
		},
		{
			name: "other pg error",
			err: &pgconn.PgError{
				Code: "42000",
			},
			expected: false,
		},
		{
			name:     "non-pg error",
			err:      errors.New("some error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsUniqueViolation(tt.err)
			if result != tt.expected {
				t.Errorf("IsUniqueViolation() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestIsForeignKeyViolation(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name: "foreign key violation error",
			err: &pgconn.PgError{
				Code: ForeignKeyViolation,
			},
			expected: true,
		},
		{
			name: "other pg error",
			err: &pgconn.PgError{
				Code: "42000",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsForeignKeyViolation(tt.err)
			if result != tt.expected {
				t.Errorf("IsForeignKeyViolation() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestIsNotNullViolation(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name: "not null violation error",
			err: &pgconn.PgError{
				Code: NotNullViolation,
			},
			expected: true,
		},
		{
			name: "other pg error",
			err: &pgconn.PgError{
				Code: "42000",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNotNullViolation(tt.err)
			if result != tt.expected {
				t.Errorf("IsNotNullViolation() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestIsCheckViolation(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name: "check violation error",
			err: &pgconn.PgError{
				Code: CheckViolation,
			},
			expected: true,
		},
		{
			name: "other pg error",
			err: &pgconn.PgError{
				Code: "42000",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsCheckViolation(tt.err)
			if result != tt.expected {
				t.Errorf("IsCheckViolation() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestIsConnectionError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name: "connection error",
			err: &pgconn.PgError{
				Code: "08000", // Connection Exception
			},
			expected: true,
		},
		{
			name: "connection does not exist",
			err: &pgconn.PgError{
				Code: "08003",
			},
			expected: true,
		},
		{
			name: "non-connection error",
			err: &pgconn.PgError{
				Code: "42000",
			},
			expected: false,
		},
		{
			name:     "non-pg error",
			err:      errors.New("some error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsConnectionError(tt.err)
			if result != tt.expected {
				t.Errorf("IsConnectionError() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestIsTooManyConnections(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name: "too many connections error",
			err: &pgconn.PgError{
				Code: TooManyConnections,
			},
			expected: true,
		},
		{
			name: "other pg error",
			err: &pgconn.PgError{
				Code: "42000",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsTooManyConnections(tt.err)
			if result != tt.expected {
				t.Errorf("IsTooManyConnections() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestGetPgErrorCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name: "pg error with code",
			err: &pgconn.PgError{
				Code: "23505",
			},
			expected: "23505",
		},
		{
			name:     "non-pg error",
			err:      errors.New("some error"),
			expected: "",
		},
		{
			name:     "nil error",
			err:      nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetPgErrorCode(tt.err)
			if result != tt.expected {
				t.Errorf("GetPgErrorCode() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestGetPgErrorMessage(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name: "pg error with message",
			err: &pgconn.PgError{
				Message: "duplicate key value violates unique constraint",
			},
			expected: "duplicate key value violates unique constraint",
		},
		{
			name:     "non-pg error",
			err:      errors.New("some error"),
			expected: "some error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetPgErrorMessage(tt.err)
			if result != tt.expected {
				t.Errorf("GetPgErrorMessage() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestGetPgErrorDetail(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name: "pg error with detail",
			err: &pgconn.PgError{
				Detail: "Key (id)=(1) already exists.",
			},
			expected: "Key (id)=(1) already exists.",
		},
		{
			name:     "non-pg error",
			err:      errors.New("some error"),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetPgErrorDetail(tt.err)
			if result != tt.expected {
				t.Errorf("GetPgErrorDetail() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestErrorConstants(t *testing.T) {
	// Verify error codes are correct
	if UniqueViolation != "23505" {
		t.Errorf("UniqueViolation = %q, expected 23505", UniqueViolation)
	}
	if ForeignKeyViolation != "23503" {
		t.Errorf("ForeignKeyViolation = %q, expected 23503", ForeignKeyViolation)
	}
	if NotNullViolation != "23502" {
		t.Errorf("NotNullViolation = %q, expected 23502", NotNullViolation)
	}
	if CheckViolation != "23514" {
		t.Errorf("CheckViolation = %q, expected 23514", CheckViolation)
	}
	if TooManyConnections != "53300" {
		t.Errorf("TooManyConnections = %q, expected 53300", TooManyConnections)
	}
}

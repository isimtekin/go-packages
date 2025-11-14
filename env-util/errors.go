package envutil

import "errors"

var (
	// ErrVariableNotSet is returned when a required environment variable is not set
	ErrVariableNotSet = errors.New("environment variable is not set")
	
	// ErrInvalidValue is returned when an environment variable has an invalid value
	ErrInvalidValue = errors.New("invalid environment variable value")
	
	// ErrInvalidType is returned when type conversion fails
	ErrInvalidType = errors.New("invalid type conversion")
	
	// ErrFileNotFound is returned when .env file is not found
	ErrFileNotFound = errors.New("environment file not found")
	
	// ErrInvalidFormat is returned when .env file has invalid format
	ErrInvalidFormat = errors.New("invalid environment file format")
)

// IsNotSet returns true if the error indicates a variable is not set
func IsNotSet(err error) bool {
	return errors.Is(err, ErrVariableNotSet)
}

// IsInvalidValue returns true if the error indicates an invalid value
func IsInvalidValue(err error) bool {
	return errors.Is(err, ErrInvalidValue)
}
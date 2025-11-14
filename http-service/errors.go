package httpservice

import (
	"errors"
	"fmt"
)

// Common errors
var (
	// ErrServiceClosed is returned when operating on a closed service
	ErrServiceClosed = errors.New("http service is closed")

	// ErrAlreadyClosed is returned when closing an already closed service
	ErrAlreadyClosed = errors.New("http service is already closed")

	// ErrInvalidHandler is returned when handler is invalid
	ErrInvalidHandler = errors.New("invalid handler")

	// ErrInvalidRoute is returned when route pattern is invalid
	ErrInvalidRoute = errors.New("invalid route pattern")

	// ErrValidationFailed is returned when request validation fails
	ErrValidationFailed = errors.New("validation failed")

	// ErrInvalidRequest is returned when request is malformed
	ErrInvalidRequest = errors.New("invalid request")
)

// HTTPError represents an HTTP error with status code
type HTTPError struct {
	Code    int                    `json:"-"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
	Err     error                  `json:"-"`
}

// Error implements the error interface
func (e *HTTPError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *HTTPError) Unwrap() error {
	return e.Err
}

// NewHTTPError creates a new HTTP error
func NewHTTPError(code int, message string, err error) *HTTPError {
	return &HTTPError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// WithDetails adds details to the error
func (e *HTTPError) WithDetails(details map[string]interface{}) *HTTPError {
	e.Details = details
	return e
}

// HTTP error constructors

// BadRequest returns a 400 error
func BadRequest(message string) *HTTPError {
	return &HTTPError{
		Code:    400,
		Message: message,
	}
}

// BadRequestf returns a 400 error with formatted message
func BadRequestf(format string, args ...interface{}) *HTTPError {
	return &HTTPError{
		Code:    400,
		Message: fmt.Sprintf(format, args...),
	}
}

// Unauthorized returns a 401 error
func Unauthorized(message string) *HTTPError {
	return &HTTPError{
		Code:    401,
		Message: message,
	}
}

// Forbidden returns a 403 error
func Forbidden(message string) *HTTPError {
	return &HTTPError{
		Code:    403,
		Message: message,
	}
}

// NotFound returns a 404 error
func NotFound(message string) *HTTPError {
	return &HTTPError{
		Code:    404,
		Message: message,
	}
}

// NotFoundf returns a 404 error with formatted message
func NotFoundf(format string, args ...interface{}) *HTTPError {
	return &HTTPError{
		Code:    404,
		Message: fmt.Sprintf(format, args...),
	}
}

// Conflict returns a 409 error
func Conflict(message string) *HTTPError {
	return &HTTPError{
		Code:    409,
		Message: message,
	}
}

// UnprocessableEntity returns a 422 error
func UnprocessableEntity(message string) *HTTPError {
	return &HTTPError{
		Code:    422,
		Message: message,
	}
}

// InternalServerError returns a 500 error
func InternalServerError(message string) *HTTPError {
	return &HTTPError{
		Code:    500,
		Message: message,
	}
}

// InternalServerErrorf returns a 500 error with formatted message
func InternalServerErrorf(format string, args ...interface{}) *HTTPError {
	return &HTTPError{
		Code:    500,
		Message: fmt.Sprintf(format, args...),
	}
}

// ServiceUnavailable returns a 503 error
func ServiceUnavailable(message string) *HTTPError {
	return &HTTPError{
		Code:    503,
		Message: message,
	}
}

// ValidationError represents validation errors
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Tag     string `json:"tag,omitempty"`
}

// NewValidationError creates a validation error
func NewValidationError(field, message, tag string) ValidationError {
	return ValidationError{
		Field:   field,
		Message: message,
		Tag:     tag,
	}
}

// IsHTTPError checks if error is HTTPError
func IsHTTPError(err error) bool {
	var httpErr *HTTPError
	return errors.As(err, &httpErr)
}

// GetHTTPError extracts HTTPError from error
func GetHTTPError(err error) *HTTPError {
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		return httpErr
	}
	return nil
}

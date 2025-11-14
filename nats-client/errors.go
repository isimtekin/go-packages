package natsclient

import "errors"

var (
	// ErrClientClosed is returned when operating on a closed client
	ErrClientClosed = errors.New("client is closed")

	// ErrAlreadyClosed is returned when closing an already closed client
	ErrAlreadyClosed = errors.New("client is already closed")

	// ErrConnectionFailed is returned when connection fails
	ErrConnectionFailed = errors.New("connection failed")

	// ErrTimeout is returned when an operation times out
	ErrTimeout = errors.New("operation timeout")

	// ErrNoResponders is returned when there are no responders
	ErrNoResponders = errors.New("no responders available for request")

	// ErrInvalidSubject is returned when the subject is invalid
	ErrInvalidSubject = errors.New("invalid subject")

	// ErrSlowConsumer is returned when consumer is too slow
	ErrSlowConsumer = errors.New("slow consumer, messages dropped")
)

// IsConnectionError returns true if the error is connection related
func IsConnectionError(err error) bool {
	return errors.Is(err, ErrConnectionFailed) ||
		errors.Is(err, ErrClientClosed)
}

// IsTimeoutError returns true if the error is timeout related
func IsTimeoutError(err error) bool {
	return errors.Is(err, ErrTimeout)
}

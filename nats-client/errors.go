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

	// ErrJetStreamNotEnabled is returned when JetStream operations are attempted without enabling JetStream
	ErrJetStreamNotEnabled = errors.New("JetStream is not enabled in configuration")

	// ErrJetStreamNotInitialized is returned when JetStream context is not initialized
	ErrJetStreamNotInitialized = errors.New("JetStream context is not initialized")

	// ErrStreamNotFound is returned when a stream is not found
	ErrStreamNotFound = errors.New("stream not found")

	// ErrConsumerNotFound is returned when a consumer is not found
	ErrConsumerNotFound = errors.New("consumer not found")

	// ErrMessageNotFound is returned when a message is not found
	ErrMessageNotFound = errors.New("message not found")
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

// IsJetStreamError returns true if the error is JetStream related
func IsJetStreamError(err error) bool {
	return errors.Is(err, ErrJetStreamNotEnabled) ||
		errors.Is(err, ErrJetStreamNotInitialized) ||
		errors.Is(err, ErrStreamNotFound) ||
		errors.Is(err, ErrConsumerNotFound) ||
		errors.Is(err, ErrMessageNotFound)
}

// IsNotFoundError returns true if the error is a not found error
func IsNotFoundError(err error) bool {
	return errors.Is(err, ErrStreamNotFound) ||
		errors.Is(err, ErrConsumerNotFound) ||
		errors.Is(err, ErrMessageNotFound)
}

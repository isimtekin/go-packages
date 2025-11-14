package kafkaclient

import "errors"

var (
	// ErrClientClosed is returned when operating on a closed client
	ErrClientClosed = errors.New("kafka client is closed")

	// ErrAlreadyClosed is returned when closing an already closed client
	ErrAlreadyClosed = errors.New("kafka client is already closed")

	// ErrConnectionFailed is returned when connection to Kafka fails
	ErrConnectionFailed = errors.New("failed to connect to Kafka brokers")

	// ErrTimeout is returned when an operation times out
	ErrTimeout = errors.New("kafka operation timeout")

	// ErrProducerClosed is returned when the producer is closed
	ErrProducerClosed = errors.New("kafka producer is closed")

	// ErrConsumerClosed is returned when the consumer is closed
	ErrConsumerClosed = errors.New("kafka consumer is closed")

	// ErrNoConsumer is returned when attempting consumer operations without a consumer
	ErrNoConsumer = errors.New("no consumer configured - specify topics in config")

	// ErrNoProducer is returned when attempting producer operations without a producer
	ErrNoProducer = errors.New("no producer configured")

	// ErrInvalidTopic is returned when an invalid topic is specified
	ErrInvalidTopic = errors.New("invalid topic name")

	// ErrInvalidPartition is returned when an invalid partition is specified
	ErrInvalidPartition = errors.New("invalid partition")

	// ErrMessageTooLarge is returned when message exceeds max size
	ErrMessageTooLarge = errors.New("message too large")

	// ErrInvalidConfig is returned when configuration is invalid
	ErrInvalidConfig = errors.New("invalid configuration")
)

// IsConnectionError returns true if the error is connection related
func IsConnectionError(err error) bool {
	return errors.Is(err, ErrConnectionFailed) ||
		errors.Is(err, ErrClientClosed) ||
		errors.Is(err, ErrProducerClosed) ||
		errors.Is(err, ErrConsumerClosed)
}

// IsTimeoutError returns true if the error is timeout related
func IsTimeoutError(err error) bool {
	return errors.Is(err, ErrTimeout)
}

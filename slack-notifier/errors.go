package slacknotifier

import "errors"

var (
	// ErrClientClosed is returned when operating on a closed client
	ErrClientClosed = errors.New("slack notifier client is closed")

	// ErrAlreadyClosed is returned when closing an already closed client
	ErrAlreadyClosed = errors.New("slack notifier client is already closed")

	// ErrConnectionFailed is returned when connection to Slack fails
	ErrConnectionFailed = errors.New("failed to connect to Slack webhook")

	// ErrTimeout is returned when an operation times out
	ErrTimeout = errors.New("slack notification timeout")

	// ErrInvalidResponse is returned when the response from Slack is invalid
	ErrInvalidResponse = errors.New("invalid response from Slack")

	// ErrEmptyWebhookURL is returned when webhook URL is empty
	ErrEmptyWebhookURL = errors.New("webhook URL cannot be empty")

	// ErrInvalidMessage is returned when message is invalid
	ErrInvalidMessage = errors.New("invalid message")
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

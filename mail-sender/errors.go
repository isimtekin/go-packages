package mailsender

import "errors"

var (
	// ErrMissingFrom is returned when the From field is empty.
	ErrMissingFrom = errors.New("missing sender email address")

	// ErrMissingRecipients is returned when there are no recipients.
	ErrMissingRecipients = errors.New("missing recipients")

	// ErrMissingSubject is returned when the Subject field is empty.
	ErrMissingSubject = errors.New("missing email subject")

	// ErrMissingContent is returned when both PlainText and HTML are empty.
	ErrMissingContent = errors.New("missing email content (plain text or HTML)")

	// ErrInvalidProvider is returned when an unsupported provider is specified.
	ErrInvalidProvider = errors.New("invalid email provider")

	// ErrMissingAPIKey is returned when the API key is empty.
	ErrMissingAPIKey = errors.New("missing API key")

	// ErrSendFailed is returned when sending an email fails.
	ErrSendFailed = errors.New("failed to send email")
)

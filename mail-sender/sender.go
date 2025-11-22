package mailsender

import (
	"context"
)

// EmailSender defines the interface that all email providers must implement.
type EmailSender interface {
	// Send sends an email message using the configured provider.
	Send(ctx context.Context, message *EmailMessage) error

	// Close closes the sender and releases any resources.
	Close() error
}

// EmailMessage represents an email message to be sent.
type EmailMessage struct {
	// From is the sender email address.
	From string

	// FromName is the sender name (optional).
	FromName string

	// To is the list of recipient email addresses.
	To []string

	// Cc is the list of carbon copy recipient email addresses (optional).
	Cc []string

	// Bcc is the list of blind carbon copy recipient email addresses (optional).
	Bcc []string

	// Subject is the email subject line.
	Subject string

	// PlainText is the plain text body of the email.
	PlainText string

	// HTML is the HTML body of the email.
	HTML string

	// ReplyTo is the reply-to email address (optional).
	ReplyTo string
}

// Validate validates the email message fields.
func (m *EmailMessage) Validate() error {
	if m.From == "" {
		return ErrMissingFrom
	}

	if len(m.To) == 0 {
		return ErrMissingRecipients
	}

	if m.Subject == "" {
		return ErrMissingSubject
	}

	if m.PlainText == "" && m.HTML == "" {
		return ErrMissingContent
	}

	return nil
}

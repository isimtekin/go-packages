package mailsender

import (
	"context"
	"fmt"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// SendGridSender implements the EmailSender interface using SendGrid.
type SendGridSender struct {
	client          *sendgrid.Client
	defaultFrom     string
	defaultFromName string
}

// NewSendGrid creates a new SendGrid email sender with the provided configuration.
func NewSendGrid(config *Config) (*SendGridSender, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	if config.Provider != ProviderSendGrid {
		return nil, fmt.Errorf("%w: expected sendgrid, got %s", ErrInvalidProvider, config.Provider)
	}

	client := sendgrid.NewSendClient(config.APIKey)

	return &SendGridSender{
		client:          client,
		defaultFrom:     config.DefaultFrom,
		defaultFromName: config.DefaultFromName,
	}, nil
}

// NewSendGridWithOptions creates a new SendGrid email sender with functional options.
func NewSendGridWithOptions(opts ...Option) (*SendGridSender, error) {
	config := DefaultConfig()
	config.Provider = ProviderSendGrid

	for _, opt := range opts {
		opt(config)
	}

	return NewSendGrid(config)
}

// Send sends an email using SendGrid.
func (s *SendGridSender) Send(ctx context.Context, message *EmailMessage) error {
	// Apply defaults if not set
	if message.From == "" {
		message.From = s.defaultFrom
	}
	if message.FromName == "" {
		message.FromName = s.defaultFromName
	}

	// Validate message
	if err := message.Validate(); err != nil {
		return fmt.Errorf("invalid message: %w", err)
	}

	// Build SendGrid message
	sgMessage := s.buildMessage(message)

	// Send the email
	response, err := s.client.SendWithContext(ctx, sgMessage)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSendFailed, err)
	}

	// Check response status
	if response.StatusCode >= 400 {
		return fmt.Errorf("%w: status code %d, body: %s", ErrSendFailed, response.StatusCode, response.Body)
	}

	return nil
}

// Close closes the SendGrid sender and releases any resources.
func (s *SendGridSender) Close() error {
	// SendGrid client doesn't require explicit cleanup
	return nil
}

// buildMessage builds a SendGrid mail message from an EmailMessage.
func (s *SendGridSender) buildMessage(message *EmailMessage) *mail.SGMailV3 {
	from := mail.NewEmail(message.FromName, message.From)

	// Build personalization with all recipients
	personalization := mail.NewPersonalization()

	// Add To recipients
	for _, to := range message.To {
		personalization.AddTos(mail.NewEmail("", to))
	}

	// Add Cc recipients
	for _, cc := range message.Cc {
		personalization.AddCCs(mail.NewEmail("", cc))
	}

	// Add Bcc recipients
	for _, bcc := range message.Bcc {
		personalization.AddBCCs(mail.NewEmail("", bcc))
	}

	// Create mail object
	m := mail.NewV3Mail()
	m.SetFrom(from)
	m.Subject = message.Subject
	m.AddPersonalizations(personalization)

	// Add content
	if message.PlainText != "" {
		m.AddContent(mail.NewContent("text/plain", message.PlainText))
	}
	if message.HTML != "" {
		m.AddContent(mail.NewContent("text/html", message.HTML))
	}

	// Add reply-to if set
	if message.ReplyTo != "" {
		m.SetReplyTo(mail.NewEmail("", message.ReplyTo))
	}

	return m
}

package mailsender

import (
	envutil "github.com/isimtekin/go-packages/env-util"
)

// NewFromEnv creates a new email sender from environment variables with the specified prefix.
// Environment variables:
//   - <PREFIX>PROVIDER: Email service provider (default: "sendgrid")
//   - <PREFIX>API_KEY: API key for the email service provider (required)
//   - <PREFIX>DEFAULT_FROM: Default sender email address (optional)
//   - <PREFIX>DEFAULT_FROM_NAME: Default sender name (optional)
func NewFromEnv(prefix string) (EmailSender, error) {
	env := envutil.NewWithOptions(
		envutil.WithPrefix(prefix),
	)

	config := DefaultConfig()

	// Load provider
	providerStr := env.GetString("PROVIDER", string(ProviderSendGrid))
	config.Provider = Provider(providerStr)

	// Load API key (required)
	config.APIKey = env.GetString("API_KEY", "")

	// Load defaults
	config.DefaultFrom = env.GetString("DEFAULT_FROM", "")
	config.DefaultFromName = env.GetString("DEFAULT_FROM_NAME", "")

	// Create sender based on provider
	switch config.Provider {
	case ProviderSendGrid:
		return NewSendGrid(config)
	default:
		return nil, ErrInvalidProvider
	}
}

// NewSendGridFromEnv creates a new SendGrid sender from environment variables with default prefix "SENDGRID_".
// Environment variables:
//   - SENDGRID_API_KEY: SendGrid API key (required)
//   - SENDGRID_DEFAULT_FROM: Default sender email address (optional)
//   - SENDGRID_DEFAULT_FROM_NAME: Default sender name (optional)
func NewSendGridFromEnv() (*SendGridSender, error) {
	sender, err := NewFromEnv("SENDGRID_")
	if err != nil {
		return nil, err
	}

	// Type assert to SendGridSender
	sgSender, ok := sender.(*SendGridSender)
	if !ok {
		return nil, ErrInvalidProvider
	}

	return sgSender, nil
}

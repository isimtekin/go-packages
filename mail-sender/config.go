package mailsender

import "fmt"

// Provider represents the email service provider type.
type Provider string

const (
	// ProviderSendGrid represents the SendGrid email service.
	ProviderSendGrid Provider = "sendgrid"
)

// Config holds the configuration for the email sender.
type Config struct {
	// Provider specifies which email service provider to use.
	Provider Provider

	// APIKey is the API key for the email service provider.
	APIKey string

	// DefaultFrom is the default sender email address.
	// If set, it will be used when EmailMessage.From is empty.
	DefaultFrom string

	// DefaultFromName is the default sender name.
	// If set, it will be used when EmailMessage.FromName is empty.
	DefaultFromName string
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c.Provider == "" {
		return ErrInvalidProvider
	}

	if c.Provider != ProviderSendGrid {
		return fmt.Errorf("%w: %s", ErrInvalidProvider, c.Provider)
	}

	if c.APIKey == "" {
		return ErrMissingAPIKey
	}

	return nil
}

// DefaultConfig returns a new Config with default values.
func DefaultConfig() *Config {
	return &Config{
		Provider: ProviderSendGrid,
	}
}

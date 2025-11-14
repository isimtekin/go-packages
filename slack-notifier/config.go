package slacknotifier

import (
	"fmt"
	"time"
)

// Config holds the configuration for Slack notifier
type Config struct {
	// WebhookURL is the Slack incoming webhook URL
	WebhookURL string

	// DefaultChannel overrides the default channel from webhook config
	DefaultChannel string

	// DefaultUsername sets the bot username
	DefaultUsername string

	// DefaultIconEmoji sets the bot icon emoji (e.g., ":robot_face:")
	DefaultIconEmoji string

	// DefaultIconURL sets the bot icon URL
	DefaultIconURL string

	// Timeout for HTTP requests
	Timeout time.Duration

	// MaxRetries for failed requests
	MaxRetries int

	// RetryDelay between retries
	RetryDelay time.Duration

	// EnableDebug enables debug logging
	EnableDebug bool

	// ThreadTS for threading messages (optional)
	ThreadTS string
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		WebhookURL:      "",
		DefaultChannel:  "",
		DefaultUsername: "Slack Notifier",
		DefaultIconEmoji: ":robot_face:",
		Timeout:         30 * time.Second,
		MaxRetries:      3,
		RetryDelay:      time.Second,
		EnableDebug:     false,
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.WebhookURL == "" {
		return fmt.Errorf("webhook URL cannot be empty")
	}

	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	if c.MaxRetries < 0 {
		return fmt.Errorf("max retries cannot be negative")
	}

	if c.RetryDelay < 0 {
		return fmt.Errorf("retry delay cannot be negative")
	}

	return nil
}

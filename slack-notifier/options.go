package slacknotifier

import "time"

// Option is a functional option for configuring the client
type Option func(*Config)

// WithWebhookURL sets the Slack webhook URL
func WithWebhookURL(url string) Option {
	return func(c *Config) {
		c.WebhookURL = url
	}
}

// WithChannel sets the default channel
func WithChannel(channel string) Option {
	return func(c *Config) {
		c.DefaultChannel = channel
	}
}

// WithUsername sets the default username
func WithUsername(username string) Option {
	return func(c *Config) {
		c.DefaultUsername = username
	}
}

// WithIconEmoji sets the default icon emoji
func WithIconEmoji(emoji string) Option {
	return func(c *Config) {
		c.DefaultIconEmoji = emoji
	}
}

// WithIconURL sets the default icon URL
func WithIconURL(url string) Option {
	return func(c *Config) {
		c.DefaultIconURL = url
	}
}

// WithTimeout sets the HTTP timeout
func WithTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

// WithMaxRetries sets the maximum number of retries
func WithMaxRetries(retries int) Option {
	return func(c *Config) {
		c.MaxRetries = retries
	}
}

// WithRetryDelay sets the delay between retries
func WithRetryDelay(delay time.Duration) Option {
	return func(c *Config) {
		c.RetryDelay = delay
	}
}

// WithDebug enables or disables debug mode
func WithDebug(enable bool) Option {
	return func(c *Config) {
		c.EnableDebug = enable
	}
}

// WithThreadTS sets the thread timestamp for threading messages
func WithThreadTS(threadTS string) Option {
	return func(c *Config) {
		c.ThreadTS = threadTS
	}
}

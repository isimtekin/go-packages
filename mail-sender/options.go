package mailsender

// Option is a function that modifies the Config.
type Option func(*Config)

// WithProvider sets the email service provider.
func WithProvider(provider Provider) Option {
	return func(c *Config) {
		c.Provider = provider
	}
}

// WithAPIKey sets the API key for the email service provider.
func WithAPIKey(apiKey string) Option {
	return func(c *Config) {
		c.APIKey = apiKey
	}
}

// WithDefaultFrom sets the default sender email address.
func WithDefaultFrom(from string) Option {
	return func(c *Config) {
		c.DefaultFrom = from
	}
}

// WithDefaultFromName sets the default sender name.
func WithDefaultFromName(fromName string) Option {
	return func(c *Config) {
		c.DefaultFromName = fromName
	}
}

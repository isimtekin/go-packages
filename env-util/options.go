package envutil

// Option is a functional option for configuring the Client
type Option func(*Config)

// WithLogger sets a custom logger
func WithLogger(logger Logger) Option {
	return func(c *Config) {
		c.Logger = logger
	}
}

// WithSilent disables all logging
func WithSilent(silent bool) Option {
	return func(c *Config) {
		c.Silent = silent
	}
}

// WithPrefix sets a prefix for all environment variable keys
func WithPrefix(prefix string) Option {
	return func(c *Config) {
		c.EnvPrefix = prefix
	}
}

// WithEnvFile loads environment variables from a file
func WithEnvFile(filename string) Option {
	return func(c *Config) {
		c.EnvFile = filename
	}
}

// WithRequired sets required environment variables that must exist
func WithRequired(keys ...string) Option {
	return func(c *Config) {
		c.Required = append(c.Required, keys...)
	}
}

// NewWithOptions creates a new Client with functional options
func NewWithOptions(opts ...Option) *Client {
	config := &Config{}
	
	for _, opt := range opts {
		opt(config)
	}
	
	return New(config)
}
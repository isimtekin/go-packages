package mongoclient

import "time"

// Option is a functional option for configuring the MongoDB client
type Option func(*Config)

// WithURI sets the MongoDB connection URI
func WithURI(uri string) Option {
	return func(c *Config) {
		c.URI = uri
	}
}

// WithDatabase sets the database name
func WithDatabase(database string) Option {
	return func(c *Config) {
		c.Database = database
	}
}

// WithMaxPoolSize sets the maximum connection pool size
func WithMaxPoolSize(size uint64) Option {
	return func(c *Config) {
		c.MaxPoolSize = size
	}
}

// WithMinPoolSize sets the minimum connection pool size
func WithMinPoolSize(size uint64) Option {
	return func(c *Config) {
		c.MinPoolSize = size
	}
}

// WithMaxConnIdleTime sets the maximum connection idle time
func WithMaxConnIdleTime(duration time.Duration) Option {
	return func(c *Config) {
		c.MaxConnIdleTime = duration
	}
}

// WithConnectTimeout sets the connection timeout
func WithConnectTimeout(duration time.Duration) Option {
	return func(c *Config) {
		c.ConnectTimeout = duration
	}
}

// WithSocketTimeout sets the socket timeout
func WithSocketTimeout(duration time.Duration) Option {
	return func(c *Config) {
		c.SocketTimeout = duration
	}
}

// WithServerSelectionTimeout sets the server selection timeout
func WithServerSelectionTimeout(duration time.Duration) Option {
	return func(c *Config) {
		c.ServerSelectionTimeout = duration
	}
}

// WithOperationTimeout sets the default operation timeout
func WithOperationTimeout(duration time.Duration) Option {
	return func(c *Config) {
		c.OperationTimeout = duration
	}
}

// WithRetryWrites enables or disables retry writes
func WithRetryWrites(enabled bool) Option {
	return func(c *Config) {
		c.RetryWrites = enabled
	}
}

// WithRetryReads enables or disables retry reads
func WithRetryReads(enabled bool) Option {
	return func(c *Config) {
		c.RetryReads = enabled
	}
}

package postgresclient

import "time"

// Option is a functional option for configuring the PostgreSQL client
type Option func(*Config)

// WithHost sets the PostgreSQL host
func WithHost(host string) Option {
	return func(c *Config) {
		c.Host = host
	}
}

// WithPort sets the PostgreSQL port
func WithPort(port int) Option {
	return func(c *Config) {
		c.Port = port
	}
}

// WithDatabase sets the database name
func WithDatabase(database string) Option {
	return func(c *Config) {
		c.Database = database
	}
}

// WithUsername sets the username
func WithUsername(username string) Option {
	return func(c *Config) {
		c.Username = username
	}
}

// WithPassword sets the password
func WithPassword(password string) Option {
	return func(c *Config) {
		c.Password = password
	}
}

// WithSSLMode sets the SSL mode
func WithSSLMode(sslMode string) Option {
	return func(c *Config) {
		c.SSLMode = sslMode
	}
}

// WithMaxConns sets the maximum number of connections
func WithMaxConns(maxConns int32) Option {
	return func(c *Config) {
		c.MaxConns = maxConns
	}
}

// WithMinConns sets the minimum number of connections
func WithMinConns(minConns int32) Option {
	return func(c *Config) {
		c.MinConns = minConns
	}
}

// WithMaxConnLifetime sets the maximum connection lifetime
func WithMaxConnLifetime(duration time.Duration) Option {
	return func(c *Config) {
		c.MaxConnLifetime = duration
	}
}

// WithMaxConnIdleTime sets the maximum connection idle time
func WithMaxConnIdleTime(duration time.Duration) Option {
	return func(c *Config) {
		c.MaxConnIdleTime = duration
	}
}

// WithHealthCheckPeriod sets the health check period
func WithHealthCheckPeriod(duration time.Duration) Option {
	return func(c *Config) {
		c.HealthCheckPeriod = duration
	}
}

// WithConnectTimeout sets the connection timeout
func WithConnectTimeout(duration time.Duration) Option {
	return func(c *Config) {
		c.ConnectTimeout = duration
	}
}

// WithQueryTimeout sets the default query timeout
func WithQueryTimeout(duration time.Duration) Option {
	return func(c *Config) {
		c.QueryTimeout = duration
	}
}

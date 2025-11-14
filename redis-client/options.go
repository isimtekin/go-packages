package redisclient

import "time"

// Option is a functional option for configuring the Redis client
type Option func(*Config)

// WithAddr sets the Redis server address (host:port)
func WithAddr(addr string) Option {
	return func(c *Config) {
		c.Addr = addr
	}
}

// WithPassword sets the Redis password
func WithPassword(password string) Option {
	return func(c *Config) {
		c.Password = password
	}
}

// WithDB sets the Redis database number
func WithDB(db int) Option {
	return func(c *Config) {
		c.DB = db
	}
}

// WithMaxRetries sets the maximum number of retries
func WithMaxRetries(retries int) Option {
	return func(c *Config) {
		c.MaxRetries = retries
	}
}

// WithPoolSize sets the connection pool size
func WithPoolSize(size int) Option {
	return func(c *Config) {
		c.PoolSize = size
	}
}

// WithMinIdleConns sets the minimum number of idle connections
func WithMinIdleConns(min int) Option {
	return func(c *Config) {
		c.MinIdleConns = min
	}
}

// WithMaxIdleConns sets the maximum number of idle connections
func WithMaxIdleConns(max int) Option {
	return func(c *Config) {
		c.MaxIdleConns = max
	}
}

// WithPoolTimeout sets the pool timeout
func WithPoolTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.PoolTimeout = timeout
	}
}

// WithConnMaxIdleTime sets the maximum idle time for connections
func WithConnMaxIdleTime(d time.Duration) Option {
	return func(c *Config) {
		c.ConnMaxIdleTime = d
	}
}

// WithConnMaxLifetime sets the maximum lifetime for connections
func WithConnMaxLifetime(d time.Duration) Option {
	return func(c *Config) {
		c.ConnMaxLifetime = d
	}
}

// WithDialTimeout sets the dial timeout
func WithDialTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.DialTimeout = timeout
	}
}

// WithReadTimeout sets the read timeout
func WithReadTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.ReadTimeout = timeout
	}
}

// WithWriteTimeout sets the write timeout
func WithWriteTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.WriteTimeout = timeout
	}
}

// WithTLS enables TLS/SSL
func WithTLS(certFile, keyFile, caFile string) Option {
	return func(c *Config) {
		c.TLSEnabled = true
		c.TLSCertFile = certFile
		c.TLSKeyFile = keyFile
		c.TLSCAFile = caFile
	}
}

// WithDatabaseNames sets the database name mappings
// Example: WithDatabaseNames(map[string]int{"cache": 0, "session": 1, "queue": 2})
func WithDatabaseNames(names map[string]int) Option {
	return func(c *Config) {
		c.DatabaseNames = names
	}
}

// WithDatabaseName adds a single database name mapping
// Example: WithDatabaseName("cache", 0)
func WithDatabaseName(name string, dbNum int) Option {
	return func(c *Config) {
		if c.DatabaseNames == nil {
			c.DatabaseNames = make(map[string]int)
		}
		c.DatabaseNames[name] = dbNum
	}
}

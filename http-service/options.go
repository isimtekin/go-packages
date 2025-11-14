package httpservice

import "time"

// Option is a functional option for configuring the service
type Option func(*Config)

// WithTitle sets the service title
func WithTitle(title string) Option {
	return func(c *Config) {
		c.Title = title
	}
}

// WithServiceDescription sets the service description
func WithServiceDescription(desc string) Option {
	return func(c *Config) {
		c.Description = desc
	}
}

// WithVersion sets the service version
func WithVersion(version string) Option {
	return func(c *Config) {
		c.Version = version
	}
}

// WithHost sets the host
func WithHost(host string) Option {
	return func(c *Config) {
		c.Host = host
	}
}

// WithPort sets the port
func WithPort(port int) Option {
	return func(c *Config) {
		c.Port = port
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

// WithIdleTimeout sets the idle timeout
func WithIdleTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.IdleTimeout = timeout
	}
}

// WithMaxRequestBodySize sets the maximum request body size
func WithMaxRequestBodySize(size int) Option {
	return func(c *Config) {
		c.MaxRequestBodySize = size
	}
}

// WithDocs enables or disables documentation endpoints
func WithDocs(enable bool) Option {
	return func(c *Config) {
		c.EnableDocs = enable
		c.EnableOpenAPI = enable
	}
}

// WithHealthCheck enables or disables health check endpoint
func WithHealthCheck(enable bool) Option {
	return func(c *Config) {
		c.EnableHealthCheck = enable
	}
}

// WithMetrics enables or disables metrics endpoint
func WithMetrics(enable bool) Option {
	return func(c *Config) {
		c.EnableMetrics = enable
	}
}

// WithCORS enables or disables CORS
func WithCORS(enable bool) Option {
	return func(c *Config) {
		c.EnableCORS = enable
	}
}

// WithCORSOrigins sets allowed CORS origins
func WithCORSOrigins(origins ...string) Option {
	return func(c *Config) {
		c.CORSAllowOrigins = origins
	}
}

// WithRequestID enables or disables request ID middleware
func WithRequestID(enable bool) Option {
	return func(c *Config) {
		c.EnableRequestID = enable
	}
}

// WithLogger enables or disables logging middleware
func WithLogger(enable bool) Option {
	return func(c *Config) {
		c.EnableLogger = enable
	}
}

// WithRecovery enables or disables recovery middleware
func WithRecovery(enable bool) Option {
	return func(c *Config) {
		c.EnableRecovery = enable
	}
}

// WithCompression enables or disables response compression
func WithCompression(enable bool) Option {
	return func(c *Config) {
		c.EnableCompression = enable
	}
}

// WithValidation enables or disables request validation
func WithValidation(enable bool) Option {
	return func(c *Config) {
		c.EnableValidation = enable
	}
}

// WithRateLimiting enables or disables rate limiting
func WithRateLimiting(enable bool, requests int, window time.Duration) Option {
	return func(c *Config) {
		c.EnableRateLimiting = enable
		c.RateLimitRequests = requests
		c.RateLimitWindow = window
	}
}

// WithDebug enables or disables debug mode
func WithDebug(enable bool) Option {
	return func(c *Config) {
		c.EnableDebug = enable
	}
}

package natsclient

import "time"

// Option is a functional option for configuring the NATS client
type Option func(*Config)

// WithURL sets the NATS server URL
func WithURL(url string) Option {
	return func(c *Config) {
		c.URL = url
	}
}

// WithName sets the client name
func WithName(name string) Option {
	return func(c *Config) {
		c.Name = name
	}
}

// WithUsername sets the username for authentication
func WithUsername(username string) Option {
	return func(c *Config) {
		c.Username = username
	}
}

// WithPassword sets the password for authentication
func WithPassword(password string) Option {
	return func(c *Config) {
		c.Password = password
	}
}

// WithToken sets the token for authentication
func WithToken(token string) Option {
	return func(c *Config) {
		c.Token = token
	}
}

// WithMaxReconnects sets the maximum number of reconnect attempts
func WithMaxReconnects(max int) Option {
	return func(c *Config) {
		c.MaxReconnects = max
	}
}

// WithReconnectWait sets the wait time between reconnect attempts
func WithReconnectWait(wait time.Duration) Option {
	return func(c *Config) {
		c.ReconnectWait = wait
	}
}

// WithTimeout sets the connection timeout
func WithTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

// WithPingInterval sets the ping interval
func WithPingInterval(interval time.Duration) Option {
	return func(c *Config) {
		c.PingInterval = interval
	}
}

// WithAllowReconnect enables or disables automatic reconnection
func WithAllowReconnect(allow bool) Option {
	return func(c *Config) {
		c.AllowReconnect = allow
	}
}

// WithNoEcho enables or disables echo messages
func WithNoEcho(noEcho bool) Option {
	return func(c *Config) {
		c.NoEcho = noEcho
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

// WithJetStream enables JetStream
func WithJetStream(enable bool) Option {
	return func(c *Config) {
		c.EnableJetStream = enable
	}
}

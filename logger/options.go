package logger

import "time"

// Option is a function that modifies a Config
type Option func(*Config)

// WithEnabled sets whether logging is enabled
func WithEnabled(enabled bool) Option {
	return func(c *Config) {
		c.Enabled = enabled
	}
}

// WithLevel sets the log level
func WithLevel(level string) Option {
	return func(c *Config) {
		c.Level = level
	}
}

// WithOutput sets the output type
func WithOutput(output string) Option {
	return func(c *Config) {
		c.Output = output
	}
}

// WithFormat sets the log format
func WithFormat(format string) Option {
	return func(c *Config) {
		c.Format = format
	}
}

// WithConsoleColors enables/disables console colors
func WithConsoleColors(enabled bool) Option {
	return func(c *Config) {
		c.ConsoleColors = enabled
	}
}

// WithLogstash configures Logstash output
func WithLogstash(host string, port int, protocol string) Option {
	return func(c *Config) {
		c.LogstashHost = host
		c.LogstashPort = port
		c.LogstashProtocol = protocol
	}
}

// WithLogstashTimeout sets the Logstash connection timeout
func WithLogstashTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.LogstashTimeout = timeout
	}
}

// WithLogstashRetries sets the number of retry attempts
func WithLogstashRetries(retries int) Option {
	return func(c *Config) {
		c.LogstashRetries = retries
	}
}

// WithFile configures file output
func WithFile(path string) Option {
	return func(c *Config) {
		c.FilePath = path
	}
}

// WithFileRotation configures file rotation settings
func WithFileRotation(maxSize, maxAge, maxBackups int) Option {
	return func(c *Config) {
		c.FileMaxSize = maxSize
		c.FileMaxAge = maxAge
		c.FileMaxBackups = maxBackups
	}
}

// WithServiceName sets the service name
func WithServiceName(name string) Option {
	return func(c *Config) {
		c.ServiceName = name
	}
}

// WithEnvironment sets the environment
func WithEnvironment(env string) Option {
	return func(c *Config) {
		c.Environment = env
	}
}

// WithVersion sets the application version
func WithVersion(version string) Option {
	return func(c *Config) {
		c.Version = version
	}
}

// WithHost sets the host name
func WithHost(host string) Option {
	return func(c *Config) {
		c.Host = host
	}
}

// WithReportCaller enables/disables caller reporting
func WithReportCaller(enabled bool) Option {
	return func(c *Config) {
		c.ReportCaller = enabled
	}
}

// WithMultiOutputs sets multiple output configurations
func WithMultiOutputs(outputs []OutputConfig) Option {
	return func(c *Config) {
		c.MultiOutputs = outputs
	}
}

// NewWithOptions creates a new logger with functional options
func NewWithOptions(opts ...Option) (*Logger, error) {
	config := DefaultConfig()
	for _, opt := range opts {
		opt(config)
	}
	return New(config)
}

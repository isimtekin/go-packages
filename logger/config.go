package logger

import (
	"fmt"
	"time"
)

// Config holds the logger configuration
type Config struct {
	// Enabled controls whether logging is active
	// If false, all log calls become no-ops
	Enabled bool

	// Level sets the minimum log level
	// Options: "trace", "debug", "info", "warn", "error", "fatal", "panic"
	Level string

	// Output determines where logs are written
	// Options: "console", "logstash", "file", "multi"
	Output string

	// Format specifies the log output format
	// Options: "json", "text"
	Format string

	// Console output settings
	ConsoleColors bool // Enable colored output for console

	// Logstash settings
	LogstashHost     string        // Logstash host address
	LogstashPort     int           // Logstash port (default: 5000)
	LogstashProtocol string        // Protocol: "tcp" or "udp"
	LogstashTimeout  time.Duration // Connection timeout
	LogstashRetries  int           // Number of retry attempts

	// File output settings
	FilePath       string // Path to log file
	FileMaxSize    int    // Maximum file size in MB before rotation
	FileMaxAge     int    // Maximum number of days to retain old log files
	FileMaxBackups int    // Maximum number of old log files to retain

	// Application metadata (added to all logs)
	ServiceName string // Service/application name
	Environment string // Environment (dev, staging, production)
	Version     string // Application version
	Host        string // Hostname

	// Performance settings
	ReportCaller bool // Include caller information (file:line)

	// Multi-output configurations
	MultiOutputs []OutputConfig
}

// OutputConfig represents a single output configuration for multi-output mode
type OutputConfig struct {
	Type     string // "console", "logstash", or "file"
	Level    string // Minimum level for this output
	Format   string // Format for this output
	Enabled  bool   // Whether this output is enabled
	FilePath string // For file output
	Host     string // For logstash output
	Port     int    // For logstash output
}

// DefaultConfig returns a logger configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Enabled:          true,
		Level:            "info",
		Output:           "console",
		Format:           "text",
		ConsoleColors:    true,
		LogstashPort:     5000,
		LogstashProtocol: "tcp",
		LogstashTimeout:  5 * time.Second,
		LogstashRetries:  3,
		FileMaxSize:      100, // 100 MB
		FileMaxAge:       30,  // 30 days
		FileMaxBackups:   10,  // 10 files
		ReportCaller:     false,
		MultiOutputs:     []OutputConfig{},
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Check output type
	validOutputs := map[string]bool{
		"console":  true,
		"logstash": true,
		"file":     true,
		"multi":    true,
	}
	if !validOutputs[c.Output] {
		return fmt.Errorf("invalid output type: %s (must be: console, logstash, file, or multi)", c.Output)
	}

	// Check log level
	validLevels := map[string]bool{
		"trace": true,
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
		"fatal": true,
		"panic": true,
	}
	if !validLevels[c.Level] {
		return fmt.Errorf("invalid log level: %s", c.Level)
	}

	// Check format
	if c.Format != "json" && c.Format != "text" {
		return fmt.Errorf("invalid format: %s (must be: json or text)", c.Format)
	}

	// Validate logstash settings if output is logstash
	if c.Output == "logstash" {
		if c.LogstashHost == "" {
			return fmt.Errorf("logstash host is required when output is logstash")
		}
		if c.LogstashPort <= 0 {
			return fmt.Errorf("logstash port must be positive")
		}
		if c.LogstashProtocol != "tcp" && c.LogstashProtocol != "udp" {
			return fmt.Errorf("logstash protocol must be tcp or udp")
		}
	}

	// Validate file settings if output is file
	if c.Output == "file" {
		if c.FilePath == "" {
			return fmt.Errorf("file path is required when output is file")
		}
	}

	// Validate multi-output settings
	if c.Output == "multi" && len(c.MultiOutputs) == 0 {
		return fmt.Errorf("at least one output must be configured for multi-output mode")
	}

	return nil
}

package httpservice

import (
	"fmt"
	"time"
)

// Config holds the configuration for the HTTP service
type Config struct {
	// Server settings
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Host        string `json:"host"`
	Port        int    `json:"port"`

	// Timeouts
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`

	// Limits
	MaxRequestBodySize int `json:"max_request_body_size"` // in bytes

	// Features
	EnableDocs          bool   `json:"enable_docs"`           // Enable /docs endpoint
	EnableOpenAPI       bool   `json:"enable_openapi"`        // Enable /openapi.json endpoint
	EnableHealthCheck   bool   `json:"enable_health_check"`   // Enable /health endpoint
	EnableMetrics       bool   `json:"enable_metrics"`        // Enable /metrics endpoint
	EnableCORS          bool   `json:"enable_cors"`           // Enable CORS
	EnableRequestID     bool   `json:"enable_request_id"`     // Enable request ID middleware
	EnableLogger        bool   `json:"enable_logger"`         // Enable logging middleware
	EnableRecovery      bool   `json:"enable_recovery"`       // Enable recovery middleware
	EnableCompression   bool   `json:"enable_compression"`    // Enable response compression
	EnableValidation    bool   `json:"enable_validation"`     // Enable request validation
	EnableRateLimiting  bool   `json:"enable_rate_limiting"`  // Enable rate limiting

	// CORS settings
	CORSAllowOrigins     []string `json:"cors_allow_origins"`
	CORSAllowMethods     []string `json:"cors_allow_methods"`
	CORSAllowHeaders     []string `json:"cors_allow_headers"`
	CORSExposeHeaders    []string `json:"cors_expose_headers"`
	CORSAllowCredentials bool     `json:"cors_allow_credentials"`
	CORSMaxAge           int      `json:"cors_max_age"`

	// Rate limiting
	RateLimitRequests int           `json:"rate_limit_requests"` // requests per window
	RateLimitWindow   time.Duration `json:"rate_limit_window"`   // time window

	// Debug
	EnableDebug bool `json:"enable_debug"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		// Server settings
		Title:       "HTTP Service",
		Description: "HTTP Service built with http-service",
		Version:     "1.0.0",
		Host:        "0.0.0.0",
		Port:        8080,

		// Timeouts
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,

		// Limits
		MaxRequestBodySize: 10 * 1024 * 1024, // 10MB

		// Features (defaults)
		EnableDocs:         true,
		EnableOpenAPI:      true,
		EnableHealthCheck:  true,
		EnableMetrics:      false,
		EnableCORS:         true,
		EnableRequestID:    true,
		EnableLogger:       true,
		EnableRecovery:     true,
		EnableCompression:  true,
		EnableValidation:   true,
		EnableRateLimiting: false,

		// CORS defaults
		CORSAllowOrigins:     []string{"*"},
		CORSAllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		CORSAllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		CORSExposeHeaders:    []string{"Content-Length"},
		CORSAllowCredentials: false,
		CORSMaxAge:           3600,

		// Rate limiting defaults
		RateLimitRequests: 100,
		RateLimitWindow:   time.Minute,

		// Debug
		EnableDebug: false,
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Title == "" {
		return fmt.Errorf("title cannot be empty")
	}

	if c.Version == "" {
		return fmt.Errorf("version cannot be empty")
	}

	if c.Host == "" {
		return fmt.Errorf("host cannot be empty")
	}

	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}

	if c.ReadTimeout <= 0 {
		return fmt.Errorf("read timeout must be positive")
	}

	if c.WriteTimeout <= 0 {
		return fmt.Errorf("write timeout must be positive")
	}

	if c.MaxRequestBodySize <= 0 {
		return fmt.Errorf("max request body size must be positive")
	}

	if c.EnableRateLimiting {
		if c.RateLimitRequests <= 0 {
			return fmt.Errorf("rate limit requests must be positive")
		}
		if c.RateLimitWindow <= 0 {
			return fmt.Errorf("rate limit window must be positive")
		}
	}

	return nil
}

// Addr returns the server address (host:port)
func (c *Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

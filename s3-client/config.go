package s3client

import (
	"fmt"
	"time"
)

// Config holds the configuration for S3 client
type Config struct {
	// AWS Region (e.g., "us-east-1", "eu-west-1")
	Region string `json:"region" yaml:"region"`

	// S3 Bucket name
	Bucket string `json:"bucket" yaml:"bucket"`

	// AWS Access Key ID (optional if using IAM role or environment)
	AccessKeyID string `json:"access_key_id" yaml:"access_key_id"`

	// AWS Secret Access Key (optional if using IAM role or environment)
	SecretAccessKey string `json:"secret_access_key" yaml:"secret_access_key"`

	// Custom endpoint URL (for S3-compatible services like MinIO, LocalStack)
	Endpoint string `json:"endpoint" yaml:"endpoint"`

	// Use path-style addressing (required for some S3-compatible services)
	UsePathStyle bool `json:"use_path_style" yaml:"use_path_style"`

	// Operation timeout
	Timeout time.Duration `json:"timeout" yaml:"timeout"`

	// Maximum number of retries for failed operations
	MaxRetries int `json:"max_retries" yaml:"max_retries"`

	// Threshold for multipart upload (files larger than this use multipart)
	MultipartThreshold int64 `json:"multipart_threshold" yaml:"multipart_threshold"`

	// Part size for multipart uploads
	PartSize int64 `json:"part_size" yaml:"part_size"`

	// Maximum concurrent uploads for multipart
	MaxConcurrentUploads int `json:"max_concurrent_uploads" yaml:"max_concurrent_uploads"`

	// Enable debug logging
	EnableDebug bool `json:"enable_debug" yaml:"enable_debug"`

	// Disable SSL verification (not recommended for production)
	DisableSSL bool `json:"disable_ssl" yaml:"disable_ssl"`
}

const (
	// DefaultRegion is the default AWS region
	DefaultRegion = "us-east-1"

	// DefaultTimeout is the default operation timeout
	DefaultTimeout = 30 * time.Second

	// DefaultMaxRetries is the default number of retries
	DefaultMaxRetries = 3

	// DefaultMultipartThreshold is 5MB - AWS minimum for multipart
	DefaultMultipartThreshold = 5 * 1024 * 1024

	// DefaultPartSize is 5MB - minimum part size for multipart
	DefaultPartSize = 5 * 1024 * 1024

	// DefaultMaxConcurrentUploads for parallel part uploads
	DefaultMaxConcurrentUploads = 5
)

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Region:               DefaultRegion,
		Timeout:              DefaultTimeout,
		MaxRetries:           DefaultMaxRetries,
		MultipartThreshold:   DefaultMultipartThreshold,
		PartSize:             DefaultPartSize,
		MaxConcurrentUploads: DefaultMaxConcurrentUploads,
		UsePathStyle:         false,
		EnableDebug:          false,
		DisableSSL:           false,
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Region == "" {
		return fmt.Errorf("%w: region cannot be empty", ErrInvalidConfig)
	}

	if c.Bucket == "" {
		return fmt.Errorf("%w: bucket cannot be empty", ErrInvalidConfig)
	}

	if c.Timeout <= 0 {
		return fmt.Errorf("%w: timeout must be positive", ErrInvalidConfig)
	}

	if c.MaxRetries < 0 {
		return fmt.Errorf("%w: max_retries cannot be negative", ErrInvalidConfig)
	}

	if c.MultipartThreshold < DefaultMultipartThreshold {
		return fmt.Errorf("%w: multipart_threshold must be at least %d bytes", ErrInvalidConfig, DefaultMultipartThreshold)
	}

	if c.PartSize < DefaultPartSize {
		return fmt.Errorf("%w: part_size must be at least %d bytes", ErrInvalidConfig, DefaultPartSize)
	}

	if c.MaxConcurrentUploads <= 0 {
		return fmt.Errorf("%w: max_concurrent_uploads must be positive", ErrInvalidConfig)
	}

	return nil
}

// Clone returns a deep copy of the configuration
func (c *Config) Clone() *Config {
	return &Config{
		Region:               c.Region,
		Bucket:               c.Bucket,
		AccessKeyID:          c.AccessKeyID,
		SecretAccessKey:      c.SecretAccessKey,
		Endpoint:             c.Endpoint,
		UsePathStyle:         c.UsePathStyle,
		Timeout:              c.Timeout,
		MaxRetries:           c.MaxRetries,
		MultipartThreshold:   c.MultipartThreshold,
		PartSize:             c.PartSize,
		MaxConcurrentUploads: c.MaxConcurrentUploads,
		EnableDebug:          c.EnableDebug,
		DisableSSL:           c.DisableSSL,
	}
}

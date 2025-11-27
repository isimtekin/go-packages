package s3client

import "time"

// Option is a functional option for configuring the client
type Option func(*Config)

// WithRegion sets the AWS region
func WithRegion(region string) Option {
	return func(c *Config) {
		c.Region = region
	}
}

// WithBucket sets the S3 bucket name
func WithBucket(bucket string) Option {
	return func(c *Config) {
		c.Bucket = bucket
	}
}

// WithCredentials sets the AWS credentials
func WithCredentials(accessKeyID, secretAccessKey string) Option {
	return func(c *Config) {
		c.AccessKeyID = accessKeyID
		c.SecretAccessKey = secretAccessKey
	}
}

// WithEndpoint sets a custom endpoint URL (for S3-compatible services)
func WithEndpoint(endpoint string) Option {
	return func(c *Config) {
		c.Endpoint = endpoint
	}
}

// WithPathStyle enables path-style addressing
func WithPathStyle(enable bool) Option {
	return func(c *Config) {
		c.UsePathStyle = enable
	}
}

// WithTimeout sets the operation timeout
func WithTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

// WithMaxRetries sets the maximum number of retries
func WithMaxRetries(retries int) Option {
	return func(c *Config) {
		c.MaxRetries = retries
	}
}

// WithMultipartThreshold sets the threshold for multipart uploads
func WithMultipartThreshold(threshold int64) Option {
	return func(c *Config) {
		c.MultipartThreshold = threshold
	}
}

// WithPartSize sets the part size for multipart uploads
func WithPartSize(size int64) Option {
	return func(c *Config) {
		c.PartSize = size
	}
}

// WithMaxConcurrentUploads sets the maximum concurrent uploads for multipart
func WithMaxConcurrentUploads(max int) Option {
	return func(c *Config) {
		c.MaxConcurrentUploads = max
	}
}

// WithDebug enables or disables debug mode
func WithDebug(enable bool) Option {
	return func(c *Config) {
		c.EnableDebug = enable
	}
}

// WithDisableSSL disables SSL verification (not recommended for production)
func WithDisableSSL(disable bool) Option {
	return func(c *Config) {
		c.DisableSSL = disable
	}
}

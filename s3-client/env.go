package s3client

import (
	"context"

	envutil "github.com/isimtekin/go-packages/env-util"
)

// NewFromEnvWithDefaults creates a new client using environment variables with S3_ prefix
func NewFromEnvWithDefaults(ctx context.Context) (*Client, error) {
	return NewFromEnv(ctx, "S3_")
}

// NewFromEnv creates a new client using environment variables with a custom prefix
//
// Environment variables:
//   - {PREFIX}REGION: AWS region (default: us-east-1)
//   - {PREFIX}BUCKET: S3 bucket name (required)
//   - {PREFIX}ACCESS_KEY_ID: AWS access key ID
//   - {PREFIX}SECRET_ACCESS_KEY: AWS secret access key
//   - {PREFIX}ENDPOINT: Custom endpoint URL (for S3-compatible services)
//   - {PREFIX}USE_PATH_STYLE: Use path-style addressing (true/false)
//   - {PREFIX}TIMEOUT: Operation timeout (e.g., "30s", "1m")
//   - {PREFIX}MAX_RETRIES: Maximum number of retries
//   - {PREFIX}MULTIPART_THRESHOLD: Threshold for multipart uploads in bytes
//   - {PREFIX}PART_SIZE: Part size for multipart uploads in bytes
//   - {PREFIX}MAX_CONCURRENT_UPLOADS: Maximum concurrent uploads for multipart
//   - {PREFIX}DEBUG: Enable debug logging (true/false)
//   - {PREFIX}DISABLE_SSL: Disable SSL verification (true/false)
func NewFromEnv(ctx context.Context, prefix string) (*Client, error) {
	env := envutil.NewWithOptions(
		envutil.WithPrefix(prefix),
	)

	cfg := DefaultConfig()

	// Load configuration from environment variables
	cfg.Region = env.GetString("REGION", cfg.Region)
	cfg.Bucket = env.GetString("BUCKET", cfg.Bucket)
	cfg.AccessKeyID = env.GetString("ACCESS_KEY_ID", cfg.AccessKeyID)
	cfg.SecretAccessKey = env.GetString("SECRET_ACCESS_KEY", cfg.SecretAccessKey)
	cfg.Endpoint = env.GetString("ENDPOINT", cfg.Endpoint)
	cfg.UsePathStyle = env.GetBool("USE_PATH_STYLE", cfg.UsePathStyle)
	cfg.Timeout = env.GetDuration("TIMEOUT", cfg.Timeout)
	cfg.MaxRetries = env.GetInt("MAX_RETRIES", cfg.MaxRetries)
	cfg.MultipartThreshold = env.GetInt64("MULTIPART_THRESHOLD", cfg.MultipartThreshold)
	cfg.PartSize = env.GetInt64("PART_SIZE", cfg.PartSize)
	cfg.MaxConcurrentUploads = env.GetInt("MAX_CONCURRENT_UPLOADS", cfg.MaxConcurrentUploads)
	cfg.EnableDebug = env.GetBool("DEBUG", cfg.EnableDebug)
	cfg.DisableSSL = env.GetBool("DISABLE_SSL", cfg.DisableSSL)

	return New(cfg)
}

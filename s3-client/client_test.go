package s3client

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Region != DefaultRegion {
		t.Errorf("Region = %v, want %v", cfg.Region, DefaultRegion)
	}

	if cfg.Timeout != DefaultTimeout {
		t.Errorf("Timeout = %v, want %v", cfg.Timeout, DefaultTimeout)
	}

	if cfg.MaxRetries != DefaultMaxRetries {
		t.Errorf("MaxRetries = %v, want %v", cfg.MaxRetries, DefaultMaxRetries)
	}

	if cfg.MultipartThreshold != DefaultMultipartThreshold {
		t.Errorf("MultipartThreshold = %v, want %v", cfg.MultipartThreshold, DefaultMultipartThreshold)
	}

	if cfg.PartSize != DefaultPartSize {
		t.Errorf("PartSize = %v, want %v", cfg.PartSize, DefaultPartSize)
	}

	if cfg.MaxConcurrentUploads != DefaultMaxConcurrentUploads {
		t.Errorf("MaxConcurrentUploads = %v, want %v", cfg.MaxConcurrentUploads, DefaultMaxConcurrentUploads)
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr error
	}{
		{
			name: "valid config",
			config: &Config{
				Region:               "us-east-1",
				Bucket:               "test-bucket",
				Timeout:              30 * time.Second,
				MaxRetries:           3,
				MultipartThreshold:   DefaultMultipartThreshold,
				PartSize:             DefaultPartSize,
				MaxConcurrentUploads: 5,
			},
			wantErr: nil,
		},
		{
			name: "empty region",
			config: &Config{
				Region:               "",
				Bucket:               "test-bucket",
				Timeout:              30 * time.Second,
				MaxRetries:           3,
				MultipartThreshold:   DefaultMultipartThreshold,
				PartSize:             DefaultPartSize,
				MaxConcurrentUploads: 5,
			},
			wantErr: ErrInvalidConfig,
		},
		{
			name: "empty bucket",
			config: &Config{
				Region:               "us-east-1",
				Bucket:               "",
				Timeout:              30 * time.Second,
				MaxRetries:           3,
				MultipartThreshold:   DefaultMultipartThreshold,
				PartSize:             DefaultPartSize,
				MaxConcurrentUploads: 5,
			},
			wantErr: ErrInvalidConfig,
		},
		{
			name: "zero timeout",
			config: &Config{
				Region:               "us-east-1",
				Bucket:               "test-bucket",
				Timeout:              0,
				MaxRetries:           3,
				MultipartThreshold:   DefaultMultipartThreshold,
				PartSize:             DefaultPartSize,
				MaxConcurrentUploads: 5,
			},
			wantErr: ErrInvalidConfig,
		},
		{
			name: "negative timeout",
			config: &Config{
				Region:               "us-east-1",
				Bucket:               "test-bucket",
				Timeout:              -1 * time.Second,
				MaxRetries:           3,
				MultipartThreshold:   DefaultMultipartThreshold,
				PartSize:             DefaultPartSize,
				MaxConcurrentUploads: 5,
			},
			wantErr: ErrInvalidConfig,
		},
		{
			name: "negative max retries",
			config: &Config{
				Region:               "us-east-1",
				Bucket:               "test-bucket",
				Timeout:              30 * time.Second,
				MaxRetries:           -1,
				MultipartThreshold:   DefaultMultipartThreshold,
				PartSize:             DefaultPartSize,
				MaxConcurrentUploads: 5,
			},
			wantErr: ErrInvalidConfig,
		},
		{
			name: "multipart threshold too small",
			config: &Config{
				Region:               "us-east-1",
				Bucket:               "test-bucket",
				Timeout:              30 * time.Second,
				MaxRetries:           3,
				MultipartThreshold:   1024, // Less than 5MB
				PartSize:             DefaultPartSize,
				MaxConcurrentUploads: 5,
			},
			wantErr: ErrInvalidConfig,
		},
		{
			name: "part size too small",
			config: &Config{
				Region:               "us-east-1",
				Bucket:               "test-bucket",
				Timeout:              30 * time.Second,
				MaxRetries:           3,
				MultipartThreshold:   DefaultMultipartThreshold,
				PartSize:             1024, // Less than 5MB
				MaxConcurrentUploads: 5,
			},
			wantErr: ErrInvalidConfig,
		},
		{
			name: "zero concurrent uploads",
			config: &Config{
				Region:               "us-east-1",
				Bucket:               "test-bucket",
				Timeout:              30 * time.Second,
				MaxRetries:           3,
				MultipartThreshold:   DefaultMultipartThreshold,
				PartSize:             DefaultPartSize,
				MaxConcurrentUploads: 0,
			},
			wantErr: ErrInvalidConfig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("Validate() error = nil, want %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Validate() error = %v, want %v", err, tt.wantErr)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() error = %v, want nil", err)
				}
			}
		})
	}
}

func TestConfig_Clone(t *testing.T) {
	original := &Config{
		Region:               "eu-west-1",
		Bucket:               "my-bucket",
		AccessKeyID:          "AKIATEST",
		SecretAccessKey:      "secret123",
		Endpoint:             "http://localhost:9000",
		UsePathStyle:         true,
		Timeout:              60 * time.Second,
		MaxRetries:           5,
		MultipartThreshold:   10 * 1024 * 1024,
		PartSize:             10 * 1024 * 1024,
		MaxConcurrentUploads: 10,
		EnableDebug:          true,
		DisableSSL:           true,
	}

	cloned := original.Clone()

	// Verify all fields are copied
	if cloned.Region != original.Region {
		t.Errorf("Region = %v, want %v", cloned.Region, original.Region)
	}
	if cloned.Bucket != original.Bucket {
		t.Errorf("Bucket = %v, want %v", cloned.Bucket, original.Bucket)
	}
	if cloned.AccessKeyID != original.AccessKeyID {
		t.Errorf("AccessKeyID = %v, want %v", cloned.AccessKeyID, original.AccessKeyID)
	}
	if cloned.SecretAccessKey != original.SecretAccessKey {
		t.Errorf("SecretAccessKey = %v, want %v", cloned.SecretAccessKey, original.SecretAccessKey)
	}
	if cloned.Endpoint != original.Endpoint {
		t.Errorf("Endpoint = %v, want %v", cloned.Endpoint, original.Endpoint)
	}
	if cloned.UsePathStyle != original.UsePathStyle {
		t.Errorf("UsePathStyle = %v, want %v", cloned.UsePathStyle, original.UsePathStyle)
	}
	if cloned.Timeout != original.Timeout {
		t.Errorf("Timeout = %v, want %v", cloned.Timeout, original.Timeout)
	}
	if cloned.MaxRetries != original.MaxRetries {
		t.Errorf("MaxRetries = %v, want %v", cloned.MaxRetries, original.MaxRetries)
	}
	if cloned.MultipartThreshold != original.MultipartThreshold {
		t.Errorf("MultipartThreshold = %v, want %v", cloned.MultipartThreshold, original.MultipartThreshold)
	}
	if cloned.PartSize != original.PartSize {
		t.Errorf("PartSize = %v, want %v", cloned.PartSize, original.PartSize)
	}
	if cloned.MaxConcurrentUploads != original.MaxConcurrentUploads {
		t.Errorf("MaxConcurrentUploads = %v, want %v", cloned.MaxConcurrentUploads, original.MaxConcurrentUploads)
	}
	if cloned.EnableDebug != original.EnableDebug {
		t.Errorf("EnableDebug = %v, want %v", cloned.EnableDebug, original.EnableDebug)
	}
	if cloned.DisableSSL != original.DisableSSL {
		t.Errorf("DisableSSL = %v, want %v", cloned.DisableSSL, original.DisableSSL)
	}

	// Verify it's a deep copy (modifying original doesn't affect clone)
	original.Region = "ap-southeast-1"
	if cloned.Region == original.Region {
		t.Error("Clone should be independent of original")
	}
}

func TestOptions(t *testing.T) {
	cfg := DefaultConfig()

	// Apply all options
	opts := []Option{
		WithRegion("eu-central-1"),
		WithBucket("test-bucket"),
		WithCredentials("AKIATEST", "secret123"),
		WithEndpoint("http://localhost:9000"),
		WithPathStyle(true),
		WithTimeout(60 * time.Second),
		WithMaxRetries(5),
		WithMultipartThreshold(10 * 1024 * 1024),
		WithPartSize(10 * 1024 * 1024),
		WithMaxConcurrentUploads(10),
		WithDebug(true),
		WithDisableSSL(true),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.Region != "eu-central-1" {
		t.Errorf("WithRegion: got %v, want eu-central-1", cfg.Region)
	}
	if cfg.Bucket != "test-bucket" {
		t.Errorf("WithBucket: got %v, want test-bucket", cfg.Bucket)
	}
	if cfg.AccessKeyID != "AKIATEST" {
		t.Errorf("WithCredentials: AccessKeyID = %v, want AKIATEST", cfg.AccessKeyID)
	}
	if cfg.SecretAccessKey != "secret123" {
		t.Errorf("WithCredentials: SecretAccessKey = %v, want secret123", cfg.SecretAccessKey)
	}
	if cfg.Endpoint != "http://localhost:9000" {
		t.Errorf("WithEndpoint: got %v, want http://localhost:9000", cfg.Endpoint)
	}
	if !cfg.UsePathStyle {
		t.Error("WithPathStyle: got false, want true")
	}
	if cfg.Timeout != 60*time.Second {
		t.Errorf("WithTimeout: got %v, want 60s", cfg.Timeout)
	}
	if cfg.MaxRetries != 5 {
		t.Errorf("WithMaxRetries: got %v, want 5", cfg.MaxRetries)
	}
	if cfg.MultipartThreshold != 10*1024*1024 {
		t.Errorf("WithMultipartThreshold: got %v, want 10MB", cfg.MultipartThreshold)
	}
	if cfg.PartSize != 10*1024*1024 {
		t.Errorf("WithPartSize: got %v, want 10MB", cfg.PartSize)
	}
	if cfg.MaxConcurrentUploads != 10 {
		t.Errorf("WithMaxConcurrentUploads: got %v, want 10", cfg.MaxConcurrentUploads)
	}
	if !cfg.EnableDebug {
		t.Error("WithDebug: got false, want true")
	}
	if !cfg.DisableSSL {
		t.Error("WithDisableSSL: got false, want true")
	}
}

func TestNew_InvalidConfig(t *testing.T) {
	// Test with invalid config (missing bucket)
	cfg := DefaultConfig()
	cfg.Bucket = ""

	_, err := New(cfg)
	if err == nil {
		t.Error("New() with invalid config should return error")
	}
	if !errors.Is(err, ErrInvalidConfig) {
		t.Errorf("New() error = %v, want %v", err, ErrInvalidConfig)
	}
}

func TestNewWithOptions_InvalidConfig(t *testing.T) {
	// Test with invalid options (missing bucket)
	_, err := NewWithOptions(
		WithRegion("us-east-1"),
		// No bucket set
	)
	if err == nil {
		t.Error("NewWithOptions() without bucket should return error")
	}
}

func TestErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		checkFn  func(error) bool
		expected bool
	}{
		{
			name:     "IsNotFoundError with ErrObjectNotFound",
			err:      ErrObjectNotFound,
			checkFn:  IsNotFoundError,
			expected: true,
		},
		{
			name:     "IsNotFoundError with ErrBucketNotFound",
			err:      ErrBucketNotFound,
			checkFn:  IsNotFoundError,
			expected: true,
		},
		{
			name:     "IsNotFoundError with other error",
			err:      ErrClientClosed,
			checkFn:  IsNotFoundError,
			expected: false,
		},
		{
			name:     "IsAccessDeniedError with ErrAccessDenied",
			err:      ErrAccessDenied,
			checkFn:  IsAccessDeniedError,
			expected: true,
		},
		{
			name:     "IsAccessDeniedError with other error",
			err:      ErrClientClosed,
			checkFn:  IsAccessDeniedError,
			expected: false,
		},
		{
			name:     "IsConfigError with ErrInvalidConfig",
			err:      ErrInvalidConfig,
			checkFn:  IsConfigError,
			expected: true,
		},
		{
			name:     "IsConfigError with other error",
			err:      ErrClientClosed,
			checkFn:  IsConfigError,
			expected: false,
		},
		{
			name:     "IsClientClosedError with ErrClientClosed",
			err:      ErrClientClosed,
			checkFn:  IsClientClosedError,
			expected: true,
		},
		{
			name:     "IsClientClosedError with ErrAlreadyClosed",
			err:      ErrAlreadyClosed,
			checkFn:  IsClientClosedError,
			expected: true,
		},
		{
			name:     "IsClientClosedError with other error",
			err:      ErrAccessDenied,
			checkFn:  IsClientClosedError,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.checkFn(tt.err)
			if got != tt.expected {
				t.Errorf("%s(%v) = %v, want %v", tt.name, tt.err, got, tt.expected)
			}
		})
	}
}

func TestDetectContentType(t *testing.T) {
	tests := []struct {
		key      string
		expected string
	}{
		{"file.html", "text/html"},
		{"file.htm", "text/html"},
		{"file.css", "text/css"},
		{"file.js", "application/javascript"},
		{"file.json", "application/json"},
		{"file.xml", "application/xml"},
		{"file.txt", "text/plain"},
		{"file.pdf", "application/pdf"},
		{"file.zip", "application/zip"},
		{"file.gz", "application/gzip"},
		{"file.tar", "application/x-tar"},
		{"file.png", "image/png"},
		{"file.jpg", "image/jpeg"},
		{"file.jpeg", "image/jpeg"},
		{"file.gif", "image/gif"},
		{"file.webp", "image/webp"},
		{"file.svg", "image/svg+xml"},
		{"file.ico", "image/x-icon"},
		{"file.mp3", "audio/mpeg"},
		{"file.mp4", "video/mp4"},
		{"file.webm", "video/webm"},
		{"file.woff", "font/woff"},
		{"file.woff2", "font/woff2"},
		{"file.ttf", "font/ttf"},
		{"file.otf", "font/otf"},
		{"file.unknown", "application/octet-stream"},
		{"file", "application/octet-stream"},
		{"path/to/file.PNG", "image/png"}, // Test case insensitivity
		{"path/to/file.JSON", "application/json"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := detectContentType(tt.key)
			if got != tt.expected {
				t.Errorf("detectContentType(%q) = %q, want %q", tt.key, got, tt.expected)
			}
		})
	}
}

func TestUploadInput_Validation(t *testing.T) {
	// These tests verify the validation logic that happens in the Upload method
	// Since we can't easily test without a real S3 connection, we test the struct
	input := &UploadInput{
		Key:         "test-key",
		Body:        bytes.NewReader([]byte("test data")),
		ContentType: "text/plain",
		Metadata:    map[string]string{"custom": "value"},
		ACL:         "public-read",
	}

	if input.Key != "test-key" {
		t.Errorf("Key = %v, want test-key", input.Key)
	}
	if input.ContentType != "text/plain" {
		t.Errorf("ContentType = %v, want text/plain", input.ContentType)
	}
	if input.ACL != "public-read" {
		t.Errorf("ACL = %v, want public-read", input.ACL)
	}
	if input.Metadata["custom"] != "value" {
		t.Errorf("Metadata[custom] = %v, want value", input.Metadata["custom"])
	}
}

func TestListInput_Defaults(t *testing.T) {
	input := &ListInput{}

	if input.Prefix != "" {
		t.Errorf("Default Prefix = %v, want empty", input.Prefix)
	}
	if input.Delimiter != "" {
		t.Errorf("Default Delimiter = %v, want empty", input.Delimiter)
	}
	if input.MaxKeys != 0 {
		t.Errorf("Default MaxKeys = %v, want 0", input.MaxKeys)
	}
	if input.StartAfter != "" {
		t.Errorf("Default StartAfter = %v, want empty", input.StartAfter)
	}
}

func TestObjectInfo(t *testing.T) {
	now := time.Now()
	info := ObjectInfo{
		Key:          "test/file.txt",
		Size:         1024,
		LastModified: now,
		ETag:         "\"abc123\"",
		ContentType:  "text/plain",
		StorageClass: "STANDARD",
	}

	if info.Key != "test/file.txt" {
		t.Errorf("Key = %v, want test/file.txt", info.Key)
	}
	if info.Size != 1024 {
		t.Errorf("Size = %v, want 1024", info.Size)
	}
	if !info.LastModified.Equal(now) {
		t.Errorf("LastModified = %v, want %v", info.LastModified, now)
	}
	if info.ETag != "\"abc123\"" {
		t.Errorf("ETag = %v, want \"abc123\"", info.ETag)
	}
	if info.ContentType != "text/plain" {
		t.Errorf("ContentType = %v, want text/plain", info.ContentType)
	}
	if info.StorageClass != "STANDARD" {
		t.Errorf("StorageClass = %v, want STANDARD", info.StorageClass)
	}
}

func TestUploadOutput(t *testing.T) {
	output := UploadOutput{
		Key:       "test/file.txt",
		Location:  "https://bucket.s3.amazonaws.com/test/file.txt",
		ETag:      "\"abc123\"",
		VersionID: "v1",
	}

	if output.Key != "test/file.txt" {
		t.Errorf("Key = %v, want test/file.txt", output.Key)
	}
	if output.Location != "https://bucket.s3.amazonaws.com/test/file.txt" {
		t.Errorf("Location = %v, want https://bucket.s3.amazonaws.com/test/file.txt", output.Location)
	}
	if output.ETag != "\"abc123\"" {
		t.Errorf("ETag = %v, want \"abc123\"", output.ETag)
	}
	if output.VersionID != "v1" {
		t.Errorf("VersionID = %v, want v1", output.VersionID)
	}
}

func TestListOutput(t *testing.T) {
	output := ListOutput{
		Objects: []ObjectInfo{
			{Key: "file1.txt", Size: 100},
			{Key: "file2.txt", Size: 200},
		},
		CommonPrefixes: []string{"folder1/", "folder2/"},
		IsTruncated:    true,
		NextMarker:     "file2.txt",
	}

	if len(output.Objects) != 2 {
		t.Errorf("Objects count = %v, want 2", len(output.Objects))
	}
	if len(output.CommonPrefixes) != 2 {
		t.Errorf("CommonPrefixes count = %v, want 2", len(output.CommonPrefixes))
	}
	if !output.IsTruncated {
		t.Error("IsTruncated = false, want true")
	}
	if output.NextMarker != "file2.txt" {
		t.Errorf("NextMarker = %v, want file2.txt", output.NextMarker)
	}
}

// Mock tests for client methods that require S3 connection
// These test the validation and error paths that don't require actual S3 calls

func TestClient_ClosedState(t *testing.T) {
	// Create a minimal valid config for testing
	cfg := &Config{
		Region:               "us-east-1",
		Bucket:               "test-bucket",
		Timeout:              30 * time.Second,
		MaxRetries:           3,
		MultipartThreshold:   DefaultMultipartThreshold,
		PartSize:             DefaultPartSize,
		MaxConcurrentUploads: 5,
	}

	client := &Client{
		cfg:    cfg,
		closed: true,
	}

	ctx := context.Background()

	// Test all methods return ErrClientClosed when client is closed
	if err := client.Ping(ctx); !errors.Is(err, ErrClientClosed) {
		t.Errorf("Ping() on closed client: got %v, want ErrClientClosed", err)
	}

	if _, err := client.Upload(ctx, &UploadInput{Key: "test", Body: bytes.NewReader([]byte{})}); !errors.Is(err, ErrClientClosed) {
		t.Errorf("Upload() on closed client: got %v, want ErrClientClosed", err)
	}

	if _, err := client.UploadBytes(ctx, "test", []byte{}, "text/plain"); !errors.Is(err, ErrClientClosed) {
		t.Errorf("UploadBytes() on closed client: got %v, want ErrClientClosed", err)
	}

	if _, err := client.UploadFile(ctx, "test", bytes.NewReader([]byte{})); !errors.Is(err, ErrClientClosed) {
		t.Errorf("UploadFile() on closed client: got %v, want ErrClientClosed", err)
	}

	if _, err := client.Download(ctx, "test"); !errors.Is(err, ErrClientClosed) {
		t.Errorf("Download() on closed client: got %v, want ErrClientClosed", err)
	}

	if _, err := client.DownloadToWriter(ctx, "test", &bytes.Buffer{}); !errors.Is(err, ErrClientClosed) {
		t.Errorf("DownloadToWriter() on closed client: got %v, want ErrClientClosed", err)
	}

	if err := client.Delete(ctx, "test"); !errors.Is(err, ErrClientClosed) {
		t.Errorf("Delete() on closed client: got %v, want ErrClientClosed", err)
	}

	if err := client.DeleteMultiple(ctx, []string{"test"}); !errors.Is(err, ErrClientClosed) {
		t.Errorf("DeleteMultiple() on closed client: got %v, want ErrClientClosed", err)
	}

	if _, err := client.Exists(ctx, "test"); !errors.Is(err, ErrClientClosed) {
		t.Errorf("Exists() on closed client: got %v, want ErrClientClosed", err)
	}

	if _, err := client.GetInfo(ctx, "test"); !errors.Is(err, ErrClientClosed) {
		t.Errorf("GetInfo() on closed client: got %v, want ErrClientClosed", err)
	}

	if _, err := client.List(ctx, nil); !errors.Is(err, ErrClientClosed) {
		t.Errorf("List() on closed client: got %v, want ErrClientClosed", err)
	}

	if _, err := client.ListAll(ctx, ""); !errors.Is(err, ErrClientClosed) {
		t.Errorf("ListAll() on closed client: got %v, want ErrClientClosed", err)
	}

	if err := client.Copy(ctx, "src", "dst"); !errors.Is(err, ErrClientClosed) {
		t.Errorf("Copy() on closed client: got %v, want ErrClientClosed", err)
	}

	if _, err := client.GetPresignedURL(ctx, "test", time.Hour); !errors.Is(err, ErrClientClosed) {
		t.Errorf("GetPresignedURL() on closed client: got %v, want ErrClientClosed", err)
	}

	if _, err := client.GetPresignedUploadURL(ctx, "test", time.Hour); !errors.Is(err, ErrClientClosed) {
		t.Errorf("GetPresignedUploadURL() on closed client: got %v, want ErrClientClosed", err)
	}
}

func TestClient_EmptyKeyValidation(t *testing.T) {
	cfg := &Config{
		Region:               "us-east-1",
		Bucket:               "test-bucket",
		Timeout:              30 * time.Second,
		MaxRetries:           3,
		MultipartThreshold:   DefaultMultipartThreshold,
		PartSize:             DefaultPartSize,
		MaxConcurrentUploads: 5,
	}

	client := &Client{
		cfg:    cfg,
		closed: false,
	}

	ctx := context.Background()

	// Test methods that require non-empty key
	if _, err := client.Upload(ctx, &UploadInput{Key: "", Body: bytes.NewReader([]byte{})}); !errors.Is(err, ErrEmptyKey) {
		t.Errorf("Upload() with empty key: got %v, want ErrEmptyKey", err)
	}

	if _, err := client.Download(ctx, ""); !errors.Is(err, ErrEmptyKey) {
		t.Errorf("Download() with empty key: got %v, want ErrEmptyKey", err)
	}

	if _, err := client.DownloadToWriter(ctx, "", &bytes.Buffer{}); !errors.Is(err, ErrEmptyKey) {
		t.Errorf("DownloadToWriter() with empty key: got %v, want ErrEmptyKey", err)
	}

	if err := client.Delete(ctx, ""); !errors.Is(err, ErrEmptyKey) {
		t.Errorf("Delete() with empty key: got %v, want ErrEmptyKey", err)
	}

	if _, err := client.Exists(ctx, ""); !errors.Is(err, ErrEmptyKey) {
		t.Errorf("Exists() with empty key: got %v, want ErrEmptyKey", err)
	}

	if _, err := client.GetInfo(ctx, ""); !errors.Is(err, ErrEmptyKey) {
		t.Errorf("GetInfo() with empty key: got %v, want ErrEmptyKey", err)
	}

	if err := client.Copy(ctx, "", "dst"); !errors.Is(err, ErrEmptyKey) {
		t.Errorf("Copy() with empty source key: got %v, want ErrEmptyKey", err)
	}

	if err := client.Copy(ctx, "src", ""); !errors.Is(err, ErrEmptyKey) {
		t.Errorf("Copy() with empty dest key: got %v, want ErrEmptyKey", err)
	}

	if _, err := client.GetPresignedURL(ctx, "", time.Hour); !errors.Is(err, ErrEmptyKey) {
		t.Errorf("GetPresignedURL() with empty key: got %v, want ErrEmptyKey", err)
	}

	if _, err := client.GetPresignedUploadURL(ctx, "", time.Hour); !errors.Is(err, ErrEmptyKey) {
		t.Errorf("GetPresignedUploadURL() with empty key: got %v, want ErrEmptyKey", err)
	}
}

func TestClient_NilReaderWriter(t *testing.T) {
	cfg := &Config{
		Region:               "us-east-1",
		Bucket:               "test-bucket",
		Timeout:              30 * time.Second,
		MaxRetries:           3,
		MultipartThreshold:   DefaultMultipartThreshold,
		PartSize:             DefaultPartSize,
		MaxConcurrentUploads: 5,
	}

	client := &Client{
		cfg:    cfg,
		closed: false,
	}

	ctx := context.Background()

	// Test nil reader
	if _, err := client.Upload(ctx, &UploadInput{Key: "test", Body: nil}); !errors.Is(err, ErrNilReader) {
		t.Errorf("Upload() with nil reader: got %v, want ErrNilReader", err)
	}

	// Test nil writer
	if _, err := client.DownloadToWriter(ctx, "test", nil); !errors.Is(err, ErrNilWriter) {
		t.Errorf("DownloadToWriter() with nil writer: got %v, want ErrNilWriter", err)
	}
}

func TestClient_DeleteMultipleEmpty(t *testing.T) {
	cfg := &Config{
		Region:               "us-east-1",
		Bucket:               "test-bucket",
		Timeout:              30 * time.Second,
		MaxRetries:           3,
		MultipartThreshold:   DefaultMultipartThreshold,
		PartSize:             DefaultPartSize,
		MaxConcurrentUploads: 5,
	}

	client := &Client{
		cfg:    cfg,
		closed: false,
	}

	ctx := context.Background()

	// DeleteMultiple with empty slice should succeed
	if err := client.DeleteMultiple(ctx, []string{}); err != nil {
		t.Errorf("DeleteMultiple() with empty slice: got %v, want nil", err)
	}
}

func TestClient_BucketAndRegion(t *testing.T) {
	cfg := &Config{
		Region:               "eu-west-1",
		Bucket:               "my-test-bucket",
		Timeout:              30 * time.Second,
		MaxRetries:           3,
		MultipartThreshold:   DefaultMultipartThreshold,
		PartSize:             DefaultPartSize,
		MaxConcurrentUploads: 5,
	}

	client := &Client{
		cfg:    cfg,
		closed: false,
	}

	if client.Bucket() != "my-test-bucket" {
		t.Errorf("Bucket() = %v, want my-test-bucket", client.Bucket())
	}

	if client.Region() != "eu-west-1" {
		t.Errorf("Region() = %v, want eu-west-1", client.Region())
	}
}

func TestClient_Close(t *testing.T) {
	cfg := &Config{
		Region:               "us-east-1",
		Bucket:               "test-bucket",
		Timeout:              30 * time.Second,
		MaxRetries:           3,
		MultipartThreshold:   DefaultMultipartThreshold,
		PartSize:             DefaultPartSize,
		MaxConcurrentUploads: 5,
	}

	client := &Client{
		cfg:    cfg,
		closed: false,
	}

	// First close should succeed
	if err := client.Close(); err != nil {
		t.Errorf("First Close() failed: %v", err)
	}

	// Second close should return ErrAlreadyClosed
	if err := client.Close(); !errors.Is(err, ErrAlreadyClosed) {
		t.Errorf("Second Close() = %v, want ErrAlreadyClosed", err)
	}
}

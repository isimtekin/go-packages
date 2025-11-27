package s3client

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// Client represents the S3 client
type Client struct {
	cfg      *Config
	s3Client *s3.Client
	uploader *manager.Uploader

	mu     sync.RWMutex
	closed bool
}

// ObjectInfo contains metadata about an S3 object
type ObjectInfo struct {
	Key          string
	Size         int64
	LastModified time.Time
	ETag         string
	ContentType  string
	StorageClass string
}

// UploadInput contains parameters for upload operations
type UploadInput struct {
	Key         string
	Body        io.Reader
	ContentType string
	Metadata    map[string]string
	ACL         string
}

// UploadOutput contains the result of an upload operation
type UploadOutput struct {
	Key      string
	Location string
	ETag     string
	VersionID string
}

// ListInput contains parameters for list operations
type ListInput struct {
	Prefix     string
	Delimiter  string
	MaxKeys    int32
	StartAfter string
}

// ListOutput contains the result of a list operation
type ListOutput struct {
	Objects       []ObjectInfo
	CommonPrefixes []string
	IsTruncated   bool
	NextMarker    string
}

// New creates a new S3 client with the given configuration
func New(cfg *Config) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	client := &Client{
		cfg:    cfg.Clone(),
		closed: false,
	}

	if err := client.connect(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return client, nil
}

// NewWithOptions creates a new S3 client with functional options
func NewWithOptions(opts ...Option) (*Client, error) {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	return New(cfg)
}

// connect initializes the AWS S3 client
func (c *Client) connect(ctx context.Context) error {
	var optFns []func(*config.LoadOptions) error

	// Set region
	optFns = append(optFns, config.WithRegion(c.cfg.Region))

	// Set credentials if provided
	if c.cfg.AccessKeyID != "" && c.cfg.SecretAccessKey != "" {
		optFns = append(optFns, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				c.cfg.AccessKeyID,
				c.cfg.SecretAccessKey,
				"",
			),
		))
	}

	// Set retry configuration
	optFns = append(optFns, config.WithRetryMaxAttempts(c.cfg.MaxRetries))

	// Load AWS configuration
	awsCfg, err := config.LoadDefaultConfig(ctx, optFns...)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client options
	s3OptFns := []func(*s3.Options){}

	// Set custom endpoint if provided
	if c.cfg.Endpoint != "" {
		s3OptFns = append(s3OptFns, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(c.cfg.Endpoint)
			o.UsePathStyle = c.cfg.UsePathStyle
		})
	}

	// Create S3 client
	c.s3Client = s3.NewFromConfig(awsCfg, s3OptFns...)

	// Create uploader with multipart configuration
	c.uploader = manager.NewUploader(c.s3Client, func(u *manager.Uploader) {
		u.PartSize = c.cfg.PartSize
		u.Concurrency = c.cfg.MaxConcurrentUploads
	})

	return nil
}

// Close closes the client and releases resources
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return ErrAlreadyClosed
	}

	c.closed = true
	return nil
}

// Ping checks if the S3 bucket is accessible
func (c *Client) Ping(ctx context.Context) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return ErrClientClosed
	}

	ctx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
	defer cancel()

	_, err := c.s3Client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(c.cfg.Bucket),
	})
	if err != nil {
		return c.wrapError(err, "ping failed")
	}

	return nil
}

// Upload uploads data to S3
func (c *Client) Upload(ctx context.Context, input *UploadInput) (*UploadOutput, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, ErrClientClosed
	}

	if input.Key == "" {
		return nil, ErrEmptyKey
	}

	if input.Body == nil {
		return nil, ErrNilReader
	}

	ctx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
	defer cancel()

	uploadInput := &s3.PutObjectInput{
		Bucket: aws.String(c.cfg.Bucket),
		Key:    aws.String(input.Key),
		Body:   input.Body,
	}

	if input.ContentType != "" {
		uploadInput.ContentType = aws.String(input.ContentType)
	}

	if len(input.Metadata) > 0 {
		uploadInput.Metadata = input.Metadata
	}

	if input.ACL != "" {
		uploadInput.ACL = types.ObjectCannedACL(input.ACL)
	}

	result, err := c.uploader.Upload(ctx, uploadInput)
	if err != nil {
		return nil, c.wrapError(err, "upload failed")
	}

	output := &UploadOutput{
		Key:      input.Key,
		Location: result.Location,
	}

	if result.ETag != nil {
		output.ETag = *result.ETag
	}

	if result.VersionID != nil {
		output.VersionID = *result.VersionID
	}

	return output, nil
}

// UploadBytes uploads byte data to S3
func (c *Client) UploadBytes(ctx context.Context, key string, data []byte, contentType string) (*UploadOutput, error) {
	return c.Upload(ctx, &UploadInput{
		Key:         key,
		Body:        bytes.NewReader(data),
		ContentType: contentType,
	})
}

// UploadFile uploads a file to S3 with automatic content type detection
func (c *Client) UploadFile(ctx context.Context, key string, reader io.Reader) (*UploadOutput, error) {
	contentType := detectContentType(key)
	return c.Upload(ctx, &UploadInput{
		Key:         key,
		Body:        reader,
		ContentType: contentType,
	})
}

// Download downloads an object from S3
func (c *Client) Download(ctx context.Context, key string) ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, ErrClientClosed
	}

	if key == "" {
		return nil, ErrEmptyKey
	}

	ctx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
	defer cancel()

	result, err := c.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.cfg.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, c.wrapError(err, "download failed")
	}
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to read body: %v", ErrDownloadFailed, err)
	}

	return data, nil
}

// DownloadToWriter downloads an object from S3 to the provided writer
func (c *Client) DownloadToWriter(ctx context.Context, key string, writer io.Writer) (int64, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return 0, ErrClientClosed
	}

	if key == "" {
		return 0, ErrEmptyKey
	}

	if writer == nil {
		return 0, ErrNilWriter
	}

	ctx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
	defer cancel()

	result, err := c.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.cfg.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return 0, c.wrapError(err, "download failed")
	}
	defer result.Body.Close()

	written, err := io.Copy(writer, result.Body)
	if err != nil {
		return written, fmt.Errorf("%w: failed to write: %v", ErrDownloadFailed, err)
	}

	return written, nil
}

// Delete deletes an object from S3
func (c *Client) Delete(ctx context.Context, key string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return ErrClientClosed
	}

	if key == "" {
		return ErrEmptyKey
	}

	ctx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
	defer cancel()

	_, err := c.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.cfg.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return c.wrapError(err, "delete failed")
	}

	return nil
}

// DeleteMultiple deletes multiple objects from S3
func (c *Client) DeleteMultiple(ctx context.Context, keys []string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return ErrClientClosed
	}

	if len(keys) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
	defer cancel()

	objects := make([]types.ObjectIdentifier, len(keys))
	for i, key := range keys {
		objects[i] = types.ObjectIdentifier{
			Key: aws.String(key),
		}
	}

	_, err := c.s3Client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
		Bucket: aws.String(c.cfg.Bucket),
		Delete: &types.Delete{
			Objects: objects,
			Quiet:   aws.Bool(true),
		},
	})
	if err != nil {
		return c.wrapError(err, "delete multiple failed")
	}

	return nil
}

// Exists checks if an object exists in S3
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return false, ErrClientClosed
	}

	if key == "" {
		return false, ErrEmptyKey
	}

	ctx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
	defer cancel()

	_, err := c.s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(c.cfg.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if isNotFoundError(err) {
			return false, nil
		}
		return false, c.wrapError(err, "exists check failed")
	}

	return true, nil
}

// GetInfo returns metadata about an object
func (c *Client) GetInfo(ctx context.Context, key string) (*ObjectInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, ErrClientClosed
	}

	if key == "" {
		return nil, ErrEmptyKey
	}

	ctx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
	defer cancel()

	result, err := c.s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(c.cfg.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, c.wrapError(err, "get info failed")
	}

	info := &ObjectInfo{
		Key:  key,
		Size: aws.ToInt64(result.ContentLength),
	}

	if result.LastModified != nil {
		info.LastModified = *result.LastModified
	}

	if result.ETag != nil {
		info.ETag = *result.ETag
	}

	if result.ContentType != nil {
		info.ContentType = *result.ContentType
	}

	if result.StorageClass != "" {
		info.StorageClass = string(result.StorageClass)
	}

	return info, nil
}

// List lists objects in the bucket
func (c *Client) List(ctx context.Context, input *ListInput) (*ListOutput, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, ErrClientClosed
	}

	if input == nil {
		input = &ListInput{}
	}

	ctx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
	defer cancel()

	listInput := &s3.ListObjectsV2Input{
		Bucket: aws.String(c.cfg.Bucket),
	}

	if input.Prefix != "" {
		listInput.Prefix = aws.String(input.Prefix)
	}

	if input.Delimiter != "" {
		listInput.Delimiter = aws.String(input.Delimiter)
	}

	if input.MaxKeys > 0 {
		listInput.MaxKeys = aws.Int32(input.MaxKeys)
	}

	if input.StartAfter != "" {
		listInput.StartAfter = aws.String(input.StartAfter)
	}

	result, err := c.s3Client.ListObjectsV2(ctx, listInput)
	if err != nil {
		return nil, c.wrapError(err, "list failed")
	}

	output := &ListOutput{
		Objects:        make([]ObjectInfo, 0, len(result.Contents)),
		CommonPrefixes: make([]string, 0, len(result.CommonPrefixes)),
		IsTruncated:    aws.ToBool(result.IsTruncated),
	}

	for _, obj := range result.Contents {
		info := ObjectInfo{
			Key:  aws.ToString(obj.Key),
			Size: aws.ToInt64(obj.Size),
		}

		if obj.LastModified != nil {
			info.LastModified = *obj.LastModified
		}

		if obj.ETag != nil {
			info.ETag = *obj.ETag
		}

		if obj.StorageClass != "" {
			info.StorageClass = string(obj.StorageClass)
		}

		output.Objects = append(output.Objects, info)
	}

	for _, prefix := range result.CommonPrefixes {
		output.CommonPrefixes = append(output.CommonPrefixes, aws.ToString(prefix.Prefix))
	}

	if result.NextContinuationToken != nil {
		output.NextMarker = *result.NextContinuationToken
	}

	return output, nil
}

// ListAll lists all objects with the given prefix (handles pagination)
func (c *Client) ListAll(ctx context.Context, prefix string) ([]ObjectInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, ErrClientClosed
	}

	var allObjects []ObjectInfo
	var continuationToken *string

	for {
		listInput := &s3.ListObjectsV2Input{
			Bucket: aws.String(c.cfg.Bucket),
		}

		if prefix != "" {
			listInput.Prefix = aws.String(prefix)
		}

		if continuationToken != nil {
			listInput.ContinuationToken = continuationToken
		}

		result, err := c.s3Client.ListObjectsV2(ctx, listInput)
		if err != nil {
			return nil, c.wrapError(err, "list all failed")
		}

		for _, obj := range result.Contents {
			info := ObjectInfo{
				Key:  aws.ToString(obj.Key),
				Size: aws.ToInt64(obj.Size),
			}

			if obj.LastModified != nil {
				info.LastModified = *obj.LastModified
			}

			if obj.ETag != nil {
				info.ETag = *obj.ETag
			}

			allObjects = append(allObjects, info)
		}

		if !aws.ToBool(result.IsTruncated) {
			break
		}

		continuationToken = result.NextContinuationToken
	}

	return allObjects, nil
}

// Copy copies an object within the bucket or across buckets
func (c *Client) Copy(ctx context.Context, sourceKey, destKey string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return ErrClientClosed
	}

	if sourceKey == "" || destKey == "" {
		return ErrEmptyKey
	}

	ctx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
	defer cancel()

	copySource := fmt.Sprintf("%s/%s", c.cfg.Bucket, sourceKey)

	_, err := c.s3Client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(c.cfg.Bucket),
		CopySource: aws.String(copySource),
		Key:        aws.String(destKey),
	})
	if err != nil {
		return c.wrapError(err, "copy failed")
	}

	return nil
}

// GetPresignedURL generates a presigned URL for downloading an object
func (c *Client) GetPresignedURL(ctx context.Context, key string, expiration time.Duration) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return "", ErrClientClosed
	}

	if key == "" {
		return "", ErrEmptyKey
	}

	presignClient := s3.NewPresignClient(c.s3Client)

	presignedReq, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.cfg.Bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiration))
	if err != nil {
		return "", c.wrapError(err, "presign failed")
	}

	return presignedReq.URL, nil
}

// GetPresignedUploadURL generates a presigned URL for uploading an object
func (c *Client) GetPresignedUploadURL(ctx context.Context, key string, expiration time.Duration) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return "", ErrClientClosed
	}

	if key == "" {
		return "", ErrEmptyKey
	}

	presignClient := s3.NewPresignClient(c.s3Client)

	presignedReq, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(c.cfg.Bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiration))
	if err != nil {
		return "", c.wrapError(err, "presign upload failed")
	}

	return presignedReq.URL, nil
}

// Bucket returns the configured bucket name
func (c *Client) Bucket() string {
	return c.cfg.Bucket
}

// Region returns the configured region
func (c *Client) Region() string {
	return c.cfg.Region
}

// wrapError wraps AWS errors with appropriate package errors
func (c *Client) wrapError(err error, operation string) error {
	if err == nil {
		return nil
	}

	// Check for HTTP response errors
	var respErr *awshttp.ResponseError
	if ok := errors.As(err, &respErr); ok {
		switch respErr.HTTPStatusCode() {
		case http.StatusNotFound:
			return fmt.Errorf("%w: %s: %v", ErrObjectNotFound, operation, err)
		case http.StatusForbidden:
			return fmt.Errorf("%w: %s: %v", ErrAccessDenied, operation, err)
		}
	}

	return fmt.Errorf("%s: %w", operation, err)
}

// isNotFoundError checks if the error is a 404 not found error
func isNotFoundError(err error) bool {
	var respErr *awshttp.ResponseError
	if ok := errors.As(err, &respErr); ok {
		return respErr.HTTPStatusCode() == http.StatusNotFound
	}
	return false
}

// detectContentType detects content type from file extension
func detectContentType(key string) string {
	ext := strings.ToLower(filepath.Ext(key))
	contentTypes := map[string]string{
		".html": "text/html",
		".htm":  "text/html",
		".css":  "text/css",
		".js":   "application/javascript",
		".json": "application/json",
		".xml":  "application/xml",
		".txt":  "text/plain",
		".pdf":  "application/pdf",
		".zip":  "application/zip",
		".gz":   "application/gzip",
		".tar":  "application/x-tar",
		".png":  "image/png",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".gif":  "image/gif",
		".webp": "image/webp",
		".svg":  "image/svg+xml",
		".ico":  "image/x-icon",
		".mp3":  "audio/mpeg",
		".mp4":  "video/mp4",
		".webm": "video/webm",
		".woff": "font/woff",
		".woff2": "font/woff2",
		".ttf":  "font/ttf",
		".otf":  "font/otf",
	}

	if ct, ok := contentTypes[ext]; ok {
		return ct
	}
	return "application/octet-stream"
}

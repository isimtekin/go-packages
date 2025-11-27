# s3-client

A Go client library for AWS S3 and S3-compatible storage services with support for file uploads, downloads, multipart uploads, and presigned URLs.

## Installation

```bash
go get github.com/isimtekin/go-packages/s3-client@v0.0.1
```

## Features

- Upload files to S3 (simple and multipart)
- Download files from S3
- Delete single or multiple objects
- List objects with pagination support
- Check object existence
- Get object metadata
- Copy objects
- Generate presigned URLs for downloads and uploads
- Support for S3-compatible services (MinIO, LocalStack, etc.)
- Automatic content type detection
- Environment variable configuration
- Thread-safe operations

## Quick Start

```go
package main

import (
    "context"
    "log"

    s3client "github.com/isimtekin/go-packages/s3-client"
)

func main() {
    // Create client with configuration
    client, err := s3client.NewWithOptions(
        s3client.WithRegion("us-east-1"),
        s3client.WithBucket("my-bucket"),
        s3client.WithCredentials("AKIAIOSFODNN7EXAMPLE", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    ctx := context.Background()

    // Upload a file
    output, err := client.UploadBytes(ctx, "hello.txt", []byte("Hello, World!"), "text/plain")
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Uploaded to: %s", output.Location)

    // Download a file
    data, err := client.Download(ctx, "hello.txt")
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Downloaded: %s", string(data))

    // Delete a file
    if err := client.Delete(ctx, "hello.txt"); err != nil {
        log.Fatal(err)
    }
}
```

## Configuration

### Using Environment Variables

```go
// Using default prefix S3_
client, err := s3client.NewFromEnvWithDefaults(context.Background())

// Using custom prefix
client, err := s3client.NewFromEnv(context.Background(), "AWS_S3_")
```

Environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `{PREFIX}REGION` | AWS region | `us-east-1` |
| `{PREFIX}BUCKET` | S3 bucket name | (required) |
| `{PREFIX}ACCESS_KEY_ID` | AWS access key ID | (optional) |
| `{PREFIX}SECRET_ACCESS_KEY` | AWS secret access key | (optional) |
| `{PREFIX}ENDPOINT` | Custom endpoint URL | (optional) |
| `{PREFIX}USE_PATH_STYLE` | Use path-style addressing | `false` |
| `{PREFIX}TIMEOUT` | Operation timeout | `30s` |
| `{PREFIX}MAX_RETRIES` | Maximum retry attempts | `3` |
| `{PREFIX}MULTIPART_THRESHOLD` | Multipart threshold (bytes) | `5242880` (5MB) |
| `{PREFIX}PART_SIZE` | Multipart part size (bytes) | `5242880` (5MB) |
| `{PREFIX}MAX_CONCURRENT_UPLOADS` | Max concurrent uploads | `5` |
| `{PREFIX}DEBUG` | Enable debug logging | `false` |
| `{PREFIX}DISABLE_SSL` | Disable SSL verification | `false` |

### Using Functional Options

```go
client, err := s3client.NewWithOptions(
    s3client.WithRegion("eu-west-1"),
    s3client.WithBucket("my-bucket"),
    s3client.WithCredentials("access-key", "secret-key"),
    s3client.WithEndpoint("http://localhost:9000"),
    s3client.WithPathStyle(true),
    s3client.WithTimeout(60*time.Second),
    s3client.WithMaxRetries(5),
    s3client.WithMultipartThreshold(10*1024*1024), // 10MB
    s3client.WithPartSize(10*1024*1024),           // 10MB
    s3client.WithMaxConcurrentUploads(10),
    s3client.WithDebug(true),
)
```

### Using Config Struct

```go
cfg := &s3client.Config{
    Region:               "us-east-1",
    Bucket:               "my-bucket",
    AccessKeyID:          "access-key",
    SecretAccessKey:      "secret-key",
    Timeout:              30 * time.Second,
    MaxRetries:           3,
    MultipartThreshold:   5 * 1024 * 1024,
    PartSize:             5 * 1024 * 1024,
    MaxConcurrentUploads: 5,
}

client, err := s3client.New(cfg)
```

## API Reference

### Upload Operations

#### Upload with Custom Input

```go
output, err := client.Upload(ctx, &s3client.UploadInput{
    Key:         "path/to/file.txt",
    Body:        reader,
    ContentType: "text/plain",
    Metadata:    map[string]string{"custom": "value"},
    ACL:         "public-read",
})
```

#### Upload Bytes

```go
output, err := client.UploadBytes(ctx, "file.txt", []byte("content"), "text/plain")
```

#### Upload File with Auto Content-Type

```go
file, _ := os.Open("image.png")
defer file.Close()
output, err := client.UploadFile(ctx, "images/photo.png", file)
```

### Download Operations

#### Download to Memory

```go
data, err := client.Download(ctx, "file.txt")
```

#### Download to Writer

```go
file, _ := os.Create("local-file.txt")
defer file.Close()
written, err := client.DownloadToWriter(ctx, "file.txt", file)
```

### Delete Operations

#### Delete Single Object

```go
err := client.Delete(ctx, "file.txt")
```

#### Delete Multiple Objects

```go
err := client.DeleteMultiple(ctx, []string{"file1.txt", "file2.txt", "file3.txt"})
```

### List Operations

#### List with Options

```go
output, err := client.List(ctx, &s3client.ListInput{
    Prefix:    "images/",
    Delimiter: "/",
    MaxKeys:   100,
})

for _, obj := range output.Objects {
    fmt.Printf("Key: %s, Size: %d\n", obj.Key, obj.Size)
}

for _, prefix := range output.CommonPrefixes {
    fmt.Printf("Folder: %s\n", prefix)
}
```

#### List All (Handles Pagination)

```go
objects, err := client.ListAll(ctx, "images/")
for _, obj := range objects {
    fmt.Printf("%s (%d bytes)\n", obj.Key, obj.Size)
}
```

### Other Operations

#### Check Existence

```go
exists, err := client.Exists(ctx, "file.txt")
```

#### Get Object Info

```go
info, err := client.GetInfo(ctx, "file.txt")
fmt.Printf("Size: %d, ContentType: %s, LastModified: %v\n",
    info.Size, info.ContentType, info.LastModified)
```

#### Copy Object

```go
err := client.Copy(ctx, "source.txt", "destination.txt")
```

#### Generate Presigned Download URL

```go
url, err := client.GetPresignedURL(ctx, "file.txt", 1*time.Hour)
fmt.Println("Download URL:", url)
```

#### Generate Presigned Upload URL

```go
url, err := client.GetPresignedUploadURL(ctx, "uploads/new-file.txt", 15*time.Minute)
fmt.Println("Upload URL:", url)
```

## Using with MinIO

MinIO is a high-performance, S3-compatible object storage. This client fully supports MinIO.

### Quick Start with Docker

```bash
# Start MinIO container
docker run -d \
  --name minio \
  -p 9000:9000 \
  -p 9001:9001 \
  -e "MINIO_ROOT_USER=minioadmin" \
  -e "MINIO_ROOT_PASSWORD=minioadmin" \
  minio/minio server /data --console-address ":9001"

# Create a bucket using MinIO CLI (mc)
docker exec minio mc alias set local http://localhost:9000 minioadmin minioadmin
docker exec minio mc mb local/my-bucket
```

### Docker Compose

```yaml
version: '3.8'
services:
  minio:
    image: minio/minio:latest
    ports:
      - "9000:9000"
      - "9001:9001"
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin
    command: server /data --console-address ":9001"
    volumes:
      - minio_data:/data

volumes:
  minio_data:
```

### Code Configuration

```go
client, err := s3client.NewWithOptions(
    s3client.WithEndpoint("http://localhost:9000"),
    s3client.WithBucket("my-bucket"),
    s3client.WithCredentials("minioadmin", "minioadmin"),
    s3client.WithPathStyle(true),  // Required for MinIO
    s3client.WithRegion("us-east-1"),
)
if err != nil {
    log.Fatal(err)
}
defer client.Close()
```

### Environment Variables Configuration

```bash
# .env file for MinIO
S3_ENDPOINT=http://localhost:9000
S3_BUCKET=my-bucket
S3_ACCESS_KEY_ID=minioadmin
S3_SECRET_ACCESS_KEY=minioadmin
S3_USE_PATH_STYLE=true
S3_REGION=us-east-1
```

```go
// Load from environment variables
client, err := s3client.NewFromEnvWithDefaults(ctx)
```

### Complete MinIO Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"

    s3client "github.com/isimtekin/go-packages/s3-client"
)

func main() {
    // Create MinIO client
    client, err := s3client.NewWithOptions(
        s3client.WithEndpoint("http://localhost:9000"),
        s3client.WithBucket("my-bucket"),
        s3client.WithCredentials("minioadmin", "minioadmin"),
        s3client.WithPathStyle(true),
        s3client.WithRegion("us-east-1"),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    ctx := context.Background()

    // Check connection
    if err := client.Ping(ctx); err != nil {
        log.Fatal("MinIO connection failed:", err)
    }
    fmt.Println("Connected to MinIO!")

    // Upload a text file
    _, err = client.UploadBytes(ctx, "documents/hello.txt", []byte("Hello MinIO!"), "text/plain")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Uploaded: documents/hello.txt")

    // Upload an image file
    imageFile, _ := os.Open("photo.jpg")
    defer imageFile.Close()
    _, err = client.UploadFile(ctx, "images/photo.jpg", imageFile)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Uploaded: images/photo.jpg")

    // List all files
    objects, _ := client.ListAll(ctx, "")
    fmt.Println("\nAll files in bucket:")
    for _, obj := range objects {
        fmt.Printf("  - %s (%d bytes)\n", obj.Key, obj.Size)
    }

    // Generate presigned URL (valid for 1 hour)
    url, _ := client.GetPresignedURL(ctx, "documents/hello.txt", 1*time.Hour)
    fmt.Printf("\nPresigned URL: %s\n", url)

    // Download file
    data, _ := client.Download(ctx, "documents/hello.txt")
    fmt.Printf("\nContent: %s\n", string(data))
}
```

### MinIO Configuration Notes

| Option | Description | Required |
|--------|-------------|----------|
| `WithEndpoint()` | MinIO server URL (e.g., `http://localhost:9000`) | Yes |
| `WithPathStyle(true)` | Must be `true` for MinIO | Yes |
| `WithCredentials()` | MinIO access key and secret key | Yes |
| `WithRegion()` | Any valid region string (MinIO ignores this) | Yes |
| `WithDisableSSL(true)` | Set if using HTTP instead of HTTPS | Optional |

### MinIO Console

Access MinIO web console at `http://localhost:9001` to:
- Browse buckets and objects
- Manage access policies
- View server metrics
- Create/delete buckets

## Using with LocalStack

```go
client, err := s3client.NewWithOptions(
    s3client.WithEndpoint("http://localhost:4566"),
    s3client.WithBucket("test-bucket"),
    s3client.WithCredentials("test", "test"),
    s3client.WithPathStyle(true),
    s3client.WithRegion("us-east-1"),
)
```

## Error Handling

```go
output, err := client.Download(ctx, "non-existent.txt")
if err != nil {
    if s3client.IsNotFoundError(err) {
        // Object doesn't exist
        log.Println("File not found")
    } else if s3client.IsAccessDeniedError(err) {
        // Permission denied
        log.Println("Access denied")
    } else if s3client.IsClientClosedError(err) {
        // Client was closed
        log.Println("Client is closed")
    } else {
        // Other error
        log.Printf("Error: %v", err)
    }
}
```

### Error Types

| Error | Description |
|-------|-------------|
| `ErrInvalidConfig` | Configuration is invalid |
| `ErrClientClosed` | Client has been closed |
| `ErrAlreadyClosed` | Client is already closed |
| `ErrBucketNotFound` | Bucket does not exist |
| `ErrObjectNotFound` | Object does not exist |
| `ErrAccessDenied` | Access is denied |
| `ErrEmptyKey` | Object key is empty |
| `ErrNilReader` | Reader is nil |
| `ErrNilWriter` | Writer is nil |
| `ErrUploadFailed` | Upload operation failed |
| `ErrDownloadFailed` | Download operation failed |
| `ErrDeleteFailed` | Delete operation failed |

## Testing

Run tests:

```bash
go test -v ./...
```

Run tests with coverage:

```bash
go test -v -cover -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

Run tests with race detector:

```bash
go test -v -race ./...
```

## Dependencies

- [aws-sdk-go-v2](https://github.com/aws/aws-sdk-go-v2) - AWS SDK for Go v2
- [env-util](https://github.com/isimtekin/go-packages/env-util) - Environment variable utilities

## License

MIT License

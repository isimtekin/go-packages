package s3client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aws/smithy-go"
)

// fakeS3 answers every request with status and, when code is set, an S3 XML error body.
func fakeS3(t *testing.T, status int, code string) *Client {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if code != "" && r.Method != http.MethodHead {
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(status)
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><Error><Code>` + code +
				`</Code><Message>m</Message><RequestId>r</RequestId></Error>`))
			return
		}
		w.WriteHeader(status)
	}))
	t.Cleanup(server.Close)
	cfg := presignTestConfig()
	cfg.Endpoint = server.URL
	cfg.MaxRetries = 1
	client, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func TestWrapErrorNoSuchBucket(t *testing.T) {
	client := fakeS3(t, http.StatusNotFound, "NoSuchBucket")
	_, err := client.Download(context.Background(), "k")
	if !errors.Is(err, ErrBucketNotFound) {
		t.Fatalf("Download error = %v, want ErrBucketNotFound", err)
	}
	if errors.Is(err, ErrObjectNotFound) {
		t.Errorf("a missing bucket must not read as a missing object: %v", err)
	}
	var apiErr smithy.APIError
	if !errors.As(err, &apiErr) || apiErr.ErrorCode() != "NoSuchBucket" {
		t.Errorf("errors.As(smithy.APIError) = %v, %v", apiErr, err)
	}
}

func TestWrapErrorNoSuchBucketOnList(t *testing.T) {
	client := fakeS3(t, http.StatusNotFound, "NoSuchBucket")
	_, err := client.List(context.Background(), &ListInput{})
	if !errors.Is(err, ErrBucketNotFound) {
		t.Fatalf("List error = %v, want ErrBucketNotFound", err)
	}
}

func TestWrapErrorNoSuchKey(t *testing.T) {
	client := fakeS3(t, http.StatusNotFound, "NoSuchKey")
	_, err := client.Download(context.Background(), "k")
	if !errors.Is(err, ErrObjectNotFound) || errors.Is(err, ErrBucketNotFound) {
		t.Fatalf("Download error = %v, want ErrObjectNotFound only", err)
	}
	var apiErr smithy.APIError
	if !errors.As(err, &apiErr) || apiErr.ErrorCode() != "NoSuchKey" {
		t.Errorf("errors.As(smithy.APIError) = %v, %v", apiErr, err)
	}
}

func TestWrapErrorPlain404OnObjectOperation(t *testing.T) {
	client := fakeS3(t, http.StatusNotFound, "")
	_, err := client.GetInfo(context.Background(), "k")
	if !errors.Is(err, ErrObjectNotFound) {
		t.Fatalf("GetInfo error = %v, want ErrObjectNotFound", err)
	}
}

func TestPingMissingBucket(t *testing.T) {
	client := fakeS3(t, http.StatusNotFound, "")
	err := client.Ping(context.Background())
	if !errors.Is(err, ErrBucketNotFound) || errors.Is(err, ErrObjectNotFound) {
		t.Fatalf("Ping error = %v, want ErrBucketNotFound only", err)
	}
}

func TestWrapErrorAccessDeniedKeepsChain(t *testing.T) {
	client := fakeS3(t, http.StatusForbidden, "AccessDenied")
	_, err := client.Download(context.Background(), "k")
	if !errors.Is(err, ErrAccessDenied) {
		t.Fatalf("Download error = %v, want ErrAccessDenied", err)
	}
	var apiErr smithy.APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("errors.As(smithy.APIError) failed: %v", err)
	}
}

func TestPresignUploadWithOptionsSignsHeaders(t *testing.T) {
	client, err := New(presignTestConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	raw, err := client.GetPresignedUploadURLWithOptions(context.Background(), "a/k.png", time.Minute,
		PresignPutOptions{ContentType: "image/png", ContentLength: 1234})
	q := parsedPresign(t, raw, err).Query()
	signed := strings.Split(q.Get("X-Amz-SignedHeaders"), ";")
	for _, want := range []string{"content-length", "content-type"} {
		if !contains(signed, want) {
			t.Errorf("X-Amz-SignedHeaders = %q, missing %s", q.Get("X-Amz-SignedHeaders"), want)
		}
	}
}

func TestPresignUploadWithoutOptionsSignsNoBodyHeaders(t *testing.T) {
	client, err := New(presignTestConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	for name, call := range map[string]func() (string, error){
		"plain": func() (string, error) {
			return client.GetPresignedUploadURL(context.Background(), "k", time.Minute)
		},
		"zero options": func() (string, error) {
			return client.GetPresignedUploadURLWithOptions(context.Background(), "k", time.Minute, PresignPutOptions{})
		},
	} {
		t.Run(name, func(t *testing.T) {
			raw, err := call()
			q := parsedPresign(t, raw, err).Query()
			signed := strings.Split(q.Get("X-Amz-SignedHeaders"), ";")
			if contains(signed, "content-length") || contains(signed, "content-type") {
				t.Errorf("X-Amz-SignedHeaders = %q", q.Get("X-Amz-SignedHeaders"))
			}
		})
	}
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

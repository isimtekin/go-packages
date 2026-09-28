package s3client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func presignTestConfig() *Config {
	cfg := DefaultConfig()
	cfg.Bucket = "b"
	cfg.Endpoint = "http://minio.internal:9000"
	cfg.AccessKeyID = "key"
	cfg.SecretAccessKey = "secret"
	cfg.UsePathStyle = true
	return cfg
}

func parsedPresign(t *testing.T, raw string, err error) *url.URL {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	return u
}

func TestPresignUsesThePublicEndpoint(t *testing.T) {
	cfg := presignTestConfig()
	cfg.PublicEndpoint = "https://files.example"
	client, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	for name, call := range map[string]func() (string, error){
		"download": func() (string, error) { return client.GetPresignedURL(context.Background(), "a/k.png", time.Minute) },
		"upload": func() (string, error) {
			return client.GetPresignedUploadURL(context.Background(), "a/k.png", time.Minute)
		},
	} {
		t.Run(name, func(t *testing.T) {
			raw, err := call()
			u := parsedPresign(t, raw, err)
			if u.Host != "files.example" || u.Scheme != "https" || u.Path != "/b/a/k.png" {
				t.Errorf("presigned URL = %s", u)
			}
		})
	}
}

func TestPresignFallsBackToEndpoint(t *testing.T) {
	client, err := New(presignTestConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	raw, err := client.GetPresignedURL(context.Background(), "k", time.Minute)
	u := parsedPresign(t, raw, err)
	if u.Host != "minio.internal:9000" {
		t.Errorf("host = %q", u.Host)
	}
}

func TestPresignPinsResponseHeaders(t *testing.T) {
	client, err := New(presignTestConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	raw, err := client.GetPresignedURLWithOptions(context.Background(), "k", time.Minute, PresignGetOptions{
		ResponseContentDisposition: "attachment", ResponseContentType: "image/png",
	})
	u := parsedPresign(t, raw, err)
	q := u.Query()
	if q.Get("response-content-disposition") != "attachment" || q.Get("response-content-type") != "image/png" {
		t.Errorf("response overrides = %v", q)
	}
	if q.Get("X-Amz-SignedHeaders") == "" || q.Get("X-Amz-Signature") == "" {
		t.Errorf("signature missing: %v", q)
	}
}

func TestPresignWithoutOptionsHasNoOverrides(t *testing.T) {
	client, err := New(presignTestConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	raw, err := client.GetPresignedURL(context.Background(), "k", time.Minute)
	q := parsedPresign(t, raw, err).Query()
	if q.Has("response-content-disposition") || q.Has("response-content-type") {
		t.Errorf("unexpected response overrides: %v", q)
	}
}

func TestPublicEndpointDoesNotChangeOrdinaryCalls(t *testing.T) {
	hits := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if r.Method != http.MethodHead || r.URL.Path != "/b/k" {
			t.Errorf("request = %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	cfg := presignTestConfig()
	cfg.Endpoint = server.URL
	cfg.PublicEndpoint = "http://127.0.0.1:1"
	client, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	if exists, err := client.Exists(context.Background(), "k"); err != nil || !exists {
		t.Fatalf("Exists = %v, %v", exists, err)
	}
	if hits != 1 {
		t.Errorf("server received %d requests, want 1", hits)
	}
}

func TestConfigCloneKeepsPublicEndpoint(t *testing.T) {
	cfg := presignTestConfig()
	cfg.PublicEndpoint = "https://files.example"
	if got := cfg.Clone().PublicEndpoint; got != cfg.PublicEndpoint {
		t.Errorf("clone endpoint = %q", got)
	}
}

func TestWithPublicEndpoint(t *testing.T) {
	cfg := DefaultConfig()
	WithPublicEndpoint("https://files.example")(cfg)
	if cfg.PublicEndpoint != "https://files.example" {
		t.Errorf("public endpoint = %q", cfg.PublicEndpoint)
	}
}

func TestEnvPublicEndpoint(t *testing.T) {
	t.Setenv("S3_BUCKET", "b")
	t.Setenv("S3_ACCESS_KEY_ID", "key")
	t.Setenv("S3_SECRET_ACCESS_KEY", "secret")
	t.Setenv("S3_PUBLIC_ENDPOINT", "https://files.example")
	client, err := NewFromEnv(context.Background(), "S3_")
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	if got := client.cfg.PublicEndpoint; got != "https://files.example" {
		t.Errorf("public endpoint = %q", got)
	}
}

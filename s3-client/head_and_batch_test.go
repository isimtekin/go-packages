package s3client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// routedS3 answers a bucket-level HEAD (path "/b") with bucketStatus and every
// object request with objectStatus (path-style: "/b/<key>"); a POST ?delete
// answers deleteBody.
func routedS3(t *testing.T, bucketStatus, objectStatus int, deleteBody string) *Client {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Query().Has("delete") {
			w.Header().Set("Content-Type", "application/xml")
			_, _ = w.Write([]byte(deleteBody))
			return
		}
		if strings.Trim(r.URL.Path, "/") == "b" {
			w.WriteHeader(bucketStatus)
			return
		}
		w.WriteHeader(objectStatus)
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

// A HEAD on an object answers a bodiless 404 for a missing bucket too: the
// client asks the bucket before calling it a missing object.
func TestHeadNotFoundTellsBucketFromObject(t *testing.T) {
	gone := routedS3(t, http.StatusNotFound, http.StatusNotFound, "")
	if _, err := gone.GetInfo(context.Background(), "k"); !errors.Is(err, ErrBucketNotFound) {
		t.Errorf("GetInfo on a missing bucket = %v, want ErrBucketNotFound", err)
	}
	if ok, err := gone.Exists(context.Background(), "k"); ok || !errors.Is(err, ErrBucketNotFound) {
		t.Errorf("Exists on a missing bucket = %v, %v, want false, ErrBucketNotFound", ok, err)
	}

	here := routedS3(t, http.StatusOK, http.StatusNotFound, "")
	if _, err := here.GetInfo(context.Background(), "k"); !errors.Is(err, ErrObjectNotFound) || errors.Is(err, ErrBucketNotFound) {
		t.Errorf("GetInfo of a missing object = %v, want ErrObjectNotFound only", err)
	}
	if ok, err := here.Exists(context.Background(), "k"); ok || err != nil {
		t.Errorf("Exists of a missing object = %v, %v, want false, nil", ok, err)
	}
}

// DeleteMultiple reads the per-key errors of DeleteObjects: a key the store
// refused is reported; a missing key is not an error (already gone).
func TestDeleteMultipleReportsRefusedKeys(t *testing.T) {
	refused := `<?xml version="1.0" encoding="UTF-8"?><DeleteResult><Error><Key>a</Key><Code>AccessDenied</Code><Message>m</Message></Error><Error><Key>b</Key><Code>NoSuchKey</Code><Message>m</Message></Error></DeleteResult>`
	err := routedS3(t, http.StatusOK, http.StatusOK, refused).DeleteMultiple(context.Background(), []string{"a", "b", "c"})
	if !errors.Is(err, ErrDeleteIncomplete) {
		t.Fatalf("DeleteMultiple = %v, want ErrDeleteIncomplete", err)
	}
	var de *DeleteMultipleError
	if !errors.As(err, &de) || len(de.Failed) != 1 || de.Failed[0] != "a" {
		t.Fatalf("failed keys = %+v, want [a]", de)
	}

	ok := `<?xml version="1.0" encoding="UTF-8"?><DeleteResult><Error><Key>b</Key><Code>NoSuchKey</Code><Message>m</Message></Error></DeleteResult>`
	if err := routedS3(t, http.StatusOK, http.StatusOK, ok).DeleteMultiple(context.Background(), []string{"b"}); err != nil {
		t.Fatalf("a missing key failed DeleteMultiple: %v", err)
	}
}

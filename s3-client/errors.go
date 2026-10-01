package s3client

import (
	"errors"
	"fmt"
)

var (
	// ErrInvalidConfig indicates configuration is invalid
	ErrInvalidConfig = errors.New("invalid configuration")

	// ErrClientClosed indicates the client has been closed
	ErrClientClosed = errors.New("client is closed")

	// ErrAlreadyClosed indicates the client is already closed
	ErrAlreadyClosed = errors.New("client is already closed")

	// ErrBucketNotFound indicates the bucket does not exist
	ErrBucketNotFound = errors.New("bucket not found")

	// ErrDeleteIncomplete indicates a DeleteMultiple the store refused for
	// some keys (see DeleteMultipleError)
	ErrDeleteIncomplete = errors.New("some objects were not deleted")

	// ErrObjectNotFound indicates the object does not exist
	ErrObjectNotFound = errors.New("object not found")

	// ErrAccessDenied indicates access is denied
	ErrAccessDenied = errors.New("access denied")

	// ErrInvalidKey indicates the object key is invalid
	ErrInvalidKey = errors.New("invalid object key")

	// ErrEmptyKey indicates the object key is empty
	ErrEmptyKey = errors.New("object key cannot be empty")

	// ErrUploadFailed indicates the upload operation failed
	ErrUploadFailed = errors.New("upload failed")

	// ErrDownloadFailed indicates the download operation failed
	ErrDownloadFailed = errors.New("download failed")

	// ErrDeleteFailed indicates the delete operation failed
	ErrDeleteFailed = errors.New("delete failed")

	// ErrMultipartUploadFailed indicates multipart upload failed
	ErrMultipartUploadFailed = errors.New("multipart upload failed")

	// ErrInvalidContentType indicates invalid content type
	ErrInvalidContentType = errors.New("invalid content type")

	// ErrFileTooLarge indicates file exceeds maximum size
	ErrFileTooLarge = errors.New("file too large")

	// ErrNilReader indicates reader is nil
	ErrNilReader = errors.New("reader cannot be nil")

	// ErrNilWriter indicates writer is nil
	ErrNilWriter = errors.New("writer cannot be nil")
)

// IsNotFoundError checks if the error is a not found error
func IsNotFoundError(err error) bool {
	return errors.Is(err, ErrObjectNotFound) || errors.Is(err, ErrBucketNotFound)
}

// IsAccessDeniedError checks if the error is an access denied error
func IsAccessDeniedError(err error) bool {
	return errors.Is(err, ErrAccessDenied)
}

// IsConfigError checks if the error is a configuration error
func IsConfigError(err error) bool {
	return errors.Is(err, ErrInvalidConfig)
}

// IsClientClosedError checks if the error is a client closed error
func IsClientClosedError(err error) bool {
	return errors.Is(err, ErrClientClosed) || errors.Is(err, ErrAlreadyClosed)
}

// DeleteMultipleError lists the keys a DeleteMultiple could not delete, with
// the store's error code for each. A key that did not exist is not listed.
type DeleteMultipleError struct {
	Failed []string
	Codes  []string
}

func (e *DeleteMultipleError) Error() string {
	return fmt.Sprintf("%d of the objects were not deleted", len(e.Failed))
}

// Is makes errors.Is(err, ErrDeleteIncomplete) hold.
func (e *DeleteMultipleError) Is(target error) bool { return target == ErrDeleteIncomplete }

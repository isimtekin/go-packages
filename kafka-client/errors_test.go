package kafkaclient

import (
	"errors"
	"fmt"
	"testing"
)

func TestErrors_Definitions(t *testing.T) {
	errorList := []struct {
		name string
		err  error
	}{
		{"ErrClientClosed", ErrClientClosed},
		{"ErrAlreadyClosed", ErrAlreadyClosed},
		{"ErrConnectionFailed", ErrConnectionFailed},
		{"ErrTimeout", ErrTimeout},
		{"ErrProducerClosed", ErrProducerClosed},
		{"ErrConsumerClosed", ErrConsumerClosed},
		{"ErrNoConsumer", ErrNoConsumer},
		{"ErrNoProducer", ErrNoProducer},
		{"ErrInvalidTopic", ErrInvalidTopic},
		{"ErrInvalidPartition", ErrInvalidPartition},
		{"ErrMessageTooLarge", ErrMessageTooLarge},
		{"ErrInvalidConfig", ErrInvalidConfig},
	}

	for _, tt := range errorList {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Errorf("%s should not be nil", tt.name)
			}
			if tt.err.Error() == "" {
				t.Errorf("%s should have error message", tt.name)
			}
		})
	}
}

func TestErrors_IsConnectionError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "ErrConnectionFailed",
			err:      ErrConnectionFailed,
			expected: true,
		},
		{
			name:     "ErrClientClosed",
			err:      ErrClientClosed,
			expected: true,
		},
		{
			name:     "ErrProducerClosed",
			err:      ErrProducerClosed,
			expected: true,
		},
		{
			name:     "ErrConsumerClosed",
			err:      ErrConsumerClosed,
			expected: true,
		},
		{
			name:     "ErrTimeout",
			err:      ErrTimeout,
			expected: false,
		},
		{
			name:     "ErrInvalidTopic",
			err:      ErrInvalidTopic,
			expected: false,
		},
		{
			name:     "wrapped connection error",
			err:      fmt.Errorf("operation failed: %w", ErrConnectionFailed),
			expected: true,
		},
		{
			name:     "wrapped client closed error",
			err:      fmt.Errorf("cannot proceed: %w", ErrClientClosed),
			expected: true,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "generic error",
			err:      errors.New("some error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsConnectionError(tt.err)
			if result != tt.expected {
				t.Errorf("IsConnectionError(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestErrors_IsTimeoutError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "ErrTimeout",
			err:      ErrTimeout,
			expected: true,
		},
		{
			name:     "wrapped timeout error",
			err:      fmt.Errorf("request failed: %w", ErrTimeout),
			expected: true,
		},
		{
			name:     "ErrConnectionFailed",
			err:      ErrConnectionFailed,
			expected: false,
		},
		{
			name:     "ErrClientClosed",
			err:      ErrClientClosed,
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "generic error",
			err:      errors.New("some error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsTimeoutError(tt.err)
			if result != tt.expected {
				t.Errorf("IsTimeoutError(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestErrors_ErrorMessages(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		wantContains string
	}{
		{
			name:         "client closed",
			err:          ErrClientClosed,
			wantContains: "closed",
		},
		{
			name:         "connection failed",
			err:          ErrConnectionFailed,
			wantContains: "connect",
		},
		{
			name:         "timeout",
			err:          ErrTimeout,
			wantContains: "timeout",
		},
		{
			name:         "invalid topic",
			err:          ErrInvalidTopic,
			wantContains: "topic",
		},
		{
			name:         "invalid config",
			err:          ErrInvalidConfig,
			wantContains: "configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errMsg := tt.err.Error()
			if errMsg == "" {
				t.Error("Error message should not be empty")
			}
		})
	}
}

func TestErrors_ErrorWrapping(t *testing.T) {
	tests := []struct {
		name      string
		baseErr   error
		wrapErr   error
		shouldBe  bool
	}{
		{
			name:      "wrapped connection error",
			baseErr:   ErrConnectionFailed,
			wrapErr:   fmt.Errorf("failed to send: %w", ErrConnectionFailed),
			shouldBe:  true,
		},
		{
			name:      "wrapped timeout error",
			baseErr:   ErrTimeout,
			wrapErr:   fmt.Errorf("operation timed out: %w", ErrTimeout),
			shouldBe:  true,
		},
		{
			name:      "double wrapped error",
			baseErr:   ErrClientClosed,
			wrapErr:   fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", ErrClientClosed)),
			shouldBe:  true,
		},
		{
			name:      "non-wrapped error",
			baseErr:   ErrInvalidTopic,
			wrapErr:   errors.New("different error"),
			shouldBe:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := errors.Is(tt.wrapErr, tt.baseErr)
			if result != tt.shouldBe {
				t.Errorf("errors.Is(%v, %v) = %v, want %v", tt.wrapErr, tt.baseErr, result, tt.shouldBe)
			}
		})
	}
}

func TestErrors_ClientErrors(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		isClient      bool
		isConnection  bool
		isTimeout     bool
	}{
		{
			name:          "client closed",
			err:           ErrClientClosed,
			isClient:      true,
			isConnection:  true,
			isTimeout:     false,
		},
		{
			name:          "already closed",
			err:           ErrAlreadyClosed,
			isClient:      true,
			isConnection:  false,
			isTimeout:     false,
		},
		{
			name:          "connection failed",
			err:           ErrConnectionFailed,
			isClient:      true,
			isConnection:  true,
			isTimeout:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.isConnection != IsConnectionError(tt.err) {
				t.Errorf("IsConnectionError mismatch for %v", tt.err)
			}
			if tt.isTimeout != IsTimeoutError(tt.err) {
				t.Errorf("IsTimeoutError mismatch for %v", tt.err)
			}
		})
	}
}

func TestErrors_ProducerErrors(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		isConnection bool
	}{
		{
			name:         "producer closed",
			err:          ErrProducerClosed,
			isConnection: true,
		},
		{
			name:         "no producer",
			err:          ErrNoProducer,
			isConnection: false,
		},
		{
			name:         "invalid topic",
			err:          ErrInvalidTopic,
			isConnection: false,
		},
		{
			name:         "message too large",
			err:          ErrMessageTooLarge,
			isConnection: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.isConnection != IsConnectionError(tt.err) {
				t.Errorf("IsConnectionError(%v) = %v, want %v", tt.err, IsConnectionError(tt.err), tt.isConnection)
			}
		})
	}
}

func TestErrors_ConsumerErrors(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		isConnection bool
	}{
		{
			name:         "consumer closed",
			err:          ErrConsumerClosed,
			isConnection: true,
		},
		{
			name:         "no consumer",
			err:          ErrNoConsumer,
			isConnection: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.isConnection != IsConnectionError(tt.err) {
				t.Errorf("IsConnectionError(%v) = %v, want %v", tt.err, IsConnectionError(tt.err), tt.isConnection)
			}
		})
	}
}

func TestErrors_ValidationErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"invalid config", ErrInvalidConfig},
		{"invalid topic", ErrInvalidTopic},
		{"invalid partition", ErrInvalidPartition},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Error("Validation error should not be nil")
			}
			if IsConnectionError(tt.err) {
				t.Error("Validation errors should not be connection errors")
			}
			if IsTimeoutError(tt.err) {
				t.Error("Validation errors should not be timeout errors")
			}
		})
	}
}

func TestErrors_ErrorComparison(t *testing.T) {
	// Test that errors are comparable
	if ErrClientClosed == ErrAlreadyClosed {
		t.Error("Different errors should not be equal")
	}

	if ErrProducerClosed == ErrConsumerClosed {
		t.Error("Different errors should not be equal")
	}

	// Test error identity
	err1 := ErrClientClosed
	err2 := ErrClientClosed

	if err1 != err2 {
		t.Error("Same error should be equal")
	}

	if !errors.Is(err1, err2) {
		t.Error("errors.Is should return true for same error")
	}
}

func TestErrors_ComplexWrapping(t *testing.T) {
	// Test multiple levels of wrapping
	baseErr := ErrConnectionFailed
	wrapped1 := fmt.Errorf("level 1: %w", baseErr)
	wrapped2 := fmt.Errorf("level 2: %w", wrapped1)
	wrapped3 := fmt.Errorf("level 3: %w", wrapped2)

	if !errors.Is(wrapped3, baseErr) {
		t.Error("Multiple levels of wrapping should preserve error identity")
	}

	if !IsConnectionError(wrapped3) {
		t.Error("IsConnectionError should work with multiple wrap levels")
	}
}

func TestErrors_NilHandling(t *testing.T) {
	// Test that helper functions handle nil correctly
	if IsConnectionError(nil) {
		t.Error("IsConnectionError should return false for nil")
	}

	if IsTimeoutError(nil) {
		t.Error("IsTimeoutError should return false for nil")
	}
}

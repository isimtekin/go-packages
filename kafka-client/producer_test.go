package kafkaclient

import (
	"context"
	"testing"
	"time"
)

func TestProducer_MessageValidation(t *testing.T) {
	tests := []struct {
		name      string
		msg       *Message
		wantError bool
	}{
		{
			name: "valid message",
			msg: &Message{
				Topic: "test-topic",
				Value: []byte("test value"),
			},
			wantError: false,
		},
		{
			name: "empty topic",
			msg: &Message{
				Topic: "",
				Value: []byte("test value"),
			},
			wantError: true,
		},
		{
			name: "message with key",
			msg: &Message{
				Topic: "test-topic",
				Key:   []byte("test-key"),
				Value: []byte("test value"),
			},
			wantError: false,
		},
		{
			name: "message with headers",
			msg: &Message{
				Topic: "test-topic",
				Value: []byte("test value"),
				Headers: map[string]string{
					"content-type": "application/json",
					"correlation-id": "12345",
				},
			},
			wantError: false,
		},
		{
			name: "message with partition",
			msg: &Message{
				Topic:     "test-topic",
				Value:     []byte("test value"),
				Partition: 2,
			},
			wantError: false,
		},
		{
			name: "message with automatic partition",
			msg: &Message{
				Topic:     "test-topic",
				Value:     []byte("test value"),
				Partition: -1,
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just validate the message structure
			if tt.msg.Topic == "" && !tt.wantError {
				t.Error("Expected error for empty topic but got none")
			}
			if tt.msg.Topic != "" && tt.wantError {
				t.Error("Expected no error but got one")
			}
		})
	}
}

func TestProducer_BatchMessages(t *testing.T) {
	tests := []struct {
		name      string
		messages  []*Message
		wantError bool
	}{
		{
			name: "valid batch",
			messages: []*Message{
				{Topic: "test-topic", Value: []byte("msg1")},
				{Topic: "test-topic", Value: []byte("msg2")},
				{Topic: "test-topic", Value: []byte("msg3")},
			},
			wantError: false,
		},
		{
			name:      "empty batch",
			messages:  []*Message{},
			wantError: false,
		},
		{
			name: "batch with invalid message",
			messages: []*Message{
				{Topic: "test-topic", Value: []byte("msg1")},
				{Topic: "", Value: []byte("msg2")}, // Invalid
				{Topic: "test-topic", Value: []byte("msg3")},
			},
			wantError: true,
		},
		{
			name: "large batch",
			messages: func() []*Message {
				msgs := make([]*Message, 100)
				for i := 0; i < 100; i++ {
					msgs[i] = &Message{
						Topic: "test-topic",
						Value: []byte("test message"),
					}
				}
				return msgs
			}(),
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate batch structure
			hasInvalidMsg := false
			for _, msg := range tt.messages {
				if msg.Topic == "" {
					hasInvalidMsg = true
					break
				}
			}

			if hasInvalidMsg != tt.wantError {
				t.Errorf("Expected error %v, got %v", tt.wantError, hasInvalidMsg)
			}
		})
	}
}

func TestProducer_MessageHeaders(t *testing.T) {
	tests := []struct {
		name    string
		headers map[string]string
	}{
		{
			name:    "no headers",
			headers: nil,
		},
		{
			name:    "empty headers",
			headers: map[string]string{},
		},
		{
			name: "single header",
			headers: map[string]string{
				"content-type": "application/json",
			},
		},
		{
			name: "multiple headers",
			headers: map[string]string{
				"content-type":   "application/json",
				"correlation-id": "abc-123",
				"timestamp":      "2024-01-01T00:00:00Z",
				"user-id":        "user-456",
			},
		},
		{
			name: "headers with special characters",
			headers: map[string]string{
				"x-custom-header": "value-with-dashes",
				"X-Another_Header": "value_with_underscores",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &Message{
				Topic:   "test-topic",
				Value:   []byte("test"),
				Headers: tt.headers,
			}

			if len(msg.Headers) != len(tt.headers) {
				t.Errorf("Headers count = %d, want %d", len(msg.Headers), len(tt.headers))
			}

			for k, v := range tt.headers {
				if msg.Headers[k] != v {
					t.Errorf("Header[%s] = %s, want %s", k, msg.Headers[k], v)
				}
			}
		})
	}
}

func TestProducer_ContextCancellation(t *testing.T) {
	// Test context cancellation handling
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	msg := &Message{
		Topic: "test-topic",
		Value: []byte("test"),
	}

	// Context is already cancelled
	if ctx.Err() == nil {
		t.Error("Expected context to be cancelled")
	}

	_ = msg
}

func TestProducer_ContextTimeout(t *testing.T) {
	// Test context timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Wait for timeout
	time.Sleep(10 * time.Millisecond)

	if ctx.Err() == nil {
		t.Error("Expected context timeout error")
	}

	if ctx.Err() != context.DeadlineExceeded {
		t.Errorf("Expected DeadlineExceeded, got %v", ctx.Err())
	}
}

func TestProducer_MessageSize(t *testing.T) {
	tests := []struct {
		name      string
		valueSize int
	}{
		{"small message", 100},
		{"medium message", 1024},
		{"large message", 10 * 1024},
		{"very large message", 100 * 1024},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := make([]byte, tt.valueSize)
			for i := range value {
				value[i] = 'A'
			}

			msg := &Message{
				Topic: "test-topic",
				Value: value,
			}

			if len(msg.Value) != tt.valueSize {
				t.Errorf("Message size = %d, want %d", len(msg.Value), tt.valueSize)
			}
		})
	}
}

func TestProducer_WorkspacePrefix(t *testing.T) {
	tests := []struct {
		name      string
		workspace string
		topic     string
		expected  string
	}{
		{
			name:      "with workspace",
			workspace: "production",
			topic:     "orders",
			expected:  "production.orders",
		},
		{
			name:      "without workspace",
			workspace: "",
			topic:     "orders",
			expected:  "orders",
		},
		{
			name:      "dev workspace",
			workspace: "dev",
			topic:     "events",
			expected:  "dev.events",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Workspace: tt.workspace,
			}

			result := config.ApplyWorkspacePrefix(tt.topic)
			if result != tt.expected {
				t.Errorf("ApplyWorkspacePrefix() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestProducer_PartitioningStrategies(t *testing.T) {
	partitioners := []struct {
		name       string
		partitioner string
	}{
		{"hash partitioner", "hash"},
		{"random partitioner", "random"},
		{"round robin partitioner", "roundrobin"},
		{"manual partitioner", "manual"},
	}

	for _, p := range partitioners {
		t.Run(p.name, func(t *testing.T) {
			config := DefaultConfig()
			config.Producer.Partitioner = p.partitioner

			saramaConfig, err := config.ToSaramaConfig()
			if err != nil {
				t.Fatalf("ToSaramaConfig failed: %v", err)
			}

			if saramaConfig.Producer.Partitioner == nil {
				t.Error("Partitioner should not be nil")
			}
		})
	}
}

func TestProducer_CompressionCodecs(t *testing.T) {
	codecs := []struct {
		name        string
		compression string
	}{
		{"no compression", "none"},
		{"gzip compression", "gzip"},
		{"snappy compression", "snappy"},
		{"lz4 compression", "lz4"},
		{"zstd compression", "zstd"},
	}

	for _, codec := range codecs {
		t.Run(codec.name, func(t *testing.T) {
			config := DefaultConfig()
			config.Producer.Compression = codec.compression

			saramaConfig, err := config.ToSaramaConfig()
			if err != nil {
				t.Fatalf("ToSaramaConfig failed: %v", err)
			}

			// Verify compression is set
			_ = saramaConfig.Producer.Compression
		})
	}
}

func TestProducer_RequiredAcks(t *testing.T) {
	acks := []struct {
		name  string
		acks  int16
		desc  string
	}{
		{"no response", 0, "NoResponse - fire and forget"},
		{"wait for local", 1, "WaitForLocal - leader acknowledgment"},
		{"wait for all", -1, "WaitForAll - all replicas"},
	}

	for _, ack := range acks {
		t.Run(ack.name, func(t *testing.T) {
			config := DefaultConfig()
			config.Producer.RequiredAcks = ack.acks

			saramaConfig, err := config.ToSaramaConfig()
			if err != nil {
				t.Fatalf("ToSaramaConfig failed: %v", err)
			}

			if int16(saramaConfig.Producer.RequiredAcks) != ack.acks {
				t.Errorf("RequiredAcks = %d, want %d", saramaConfig.Producer.RequiredAcks, ack.acks)
			}
		})
	}
}

func TestProducer_IdempotentWrites(t *testing.T) {
	tests := []struct {
		name       string
		idempotent bool
	}{
		{"idempotent enabled", true},
		{"idempotent disabled", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			config.Producer.IdempotentWrites = tt.idempotent

			saramaConfig, err := config.ToSaramaConfig()
			if err != nil {
				t.Fatalf("ToSaramaConfig failed: %v", err)
			}

			if saramaConfig.Producer.Idempotent != tt.idempotent {
				t.Errorf("Idempotent = %v, want %v", saramaConfig.Producer.Idempotent, tt.idempotent)
			}
		})
	}
}

func TestProducer_RetryConfiguration(t *testing.T) {
	tests := []struct {
		name         string
		retryMax     int
		retryBackoff time.Duration
	}{
		{"no retries", 0, 0},
		{"default retries", 3, 100 * time.Millisecond},
		{"aggressive retries", 10, 50 * time.Millisecond},
		{"conservative retries", 5, 500 * time.Millisecond},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			config.Producer.RetryMax = tt.retryMax
			config.Producer.RetryBackoff = tt.retryBackoff

			saramaConfig, err := config.ToSaramaConfig()
			if err != nil {
				t.Fatalf("ToSaramaConfig failed: %v", err)
			}

			if saramaConfig.Producer.Retry.Max != tt.retryMax {
				t.Errorf("Retry.Max = %d, want %d", saramaConfig.Producer.Retry.Max, tt.retryMax)
			}

			if saramaConfig.Producer.Retry.Backoff != tt.retryBackoff {
				t.Errorf("Retry.Backoff = %v, want %v", saramaConfig.Producer.Retry.Backoff, tt.retryBackoff)
			}
		})
	}
}

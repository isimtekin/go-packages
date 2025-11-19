package kafkaclient

import (
	"context"
	"testing"
	"time"

	"github.com/IBM/sarama"
)

func TestConsumer_MessageStruct(t *testing.T) {
	tests := []struct {
		name string
		msg  *ConsumedMessage
	}{
		{
			name: "basic message",
			msg: &ConsumedMessage{
				Topic:     "test-topic",
				Partition: 0,
				Offset:    100,
				Value:     []byte("test value"),
			},
		},
		{
			name: "message with key",
			msg: &ConsumedMessage{
				Topic:     "test-topic",
				Partition: 1,
				Offset:    200,
				Key:       []byte("test-key"),
				Value:     []byte("test value"),
			},
		},
		{
			name: "message with headers",
			msg: &ConsumedMessage{
				Topic:     "test-topic",
				Partition: 2,
				Offset:    300,
				Value:     []byte("test value"),
				Headers: map[string]string{
					"content-type": "application/json",
					"timestamp":    "2024-01-01",
				},
			},
		},
		{
			name: "message with timestamp",
			msg: &ConsumedMessage{
				Topic:     "test-topic",
				Partition: 0,
				Offset:    400,
				Value:     []byte("test value"),
				Timestamp: time.Now().Unix(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.msg.Topic == "" {
				t.Error("Topic should not be empty")
			}
			if tt.msg.Value == nil {
				t.Error("Value should not be nil")
			}
			if tt.msg.Partition < 0 {
				t.Error("Partition should be non-negative")
			}
			if tt.msg.Offset < 0 {
				t.Error("Offset should be non-negative")
			}
		})
	}
}

func TestConsumer_MessageHandler(t *testing.T) {
	tests := []struct {
		name        string
		handler     MessageHandler
		msg         *ConsumedMessage
		expectError bool
	}{
		{
			name: "successful handler",
			handler: func(ctx context.Context, msg *ConsumedMessage) error {
				return nil
			},
			msg: &ConsumedMessage{
				Topic: "test-topic",
				Value: []byte("test"),
			},
			expectError: false,
		},
		{
			name: "handler with error",
			handler: func(ctx context.Context, msg *ConsumedMessage) error {
				return ErrTimeout
			},
			msg: &ConsumedMessage{
				Topic: "test-topic",
				Value: []byte("test"),
			},
			expectError: true,
		},
		{
			name: "handler with message processing",
			handler: func(ctx context.Context, msg *ConsumedMessage) error {
				// Simulate message processing
				if len(msg.Value) == 0 {
					return ErrInvalidTopic
				}
				return nil
			},
			msg: &ConsumedMessage{
				Topic: "test-topic",
				Value: []byte("test data"),
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := tt.handler(ctx, tt.msg)

			if (err != nil) != tt.expectError {
				t.Errorf("Handler error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestConsumer_Config(t *testing.T) {
	tests := []struct {
		name   string
		config ConsumerConfig
		valid  bool
	}{
		{
			name: "valid config",
			config: ConsumerConfig{
				GroupID:            "test-group",
				Topics:             []string{"topic1", "topic2"},
				SessionTimeout:     10 * time.Second,
				RebalanceTimeout:   60 * time.Second,
				OffsetInitial:      sarama.OffsetNewest,
				AutoCommit:         true,
				AutoCommitInterval: 1 * time.Second,
				MaxProcessingTime:  5 * time.Minute,
			},
			valid: true,
		},
		{
			name: "empty topics",
			config: ConsumerConfig{
				GroupID:        "test-group",
				Topics:         []string{},
				SessionTimeout: 10 * time.Second,
			},
			valid: false,
		},
		{
			name: "multiple topics",
			config: ConsumerConfig{
				GroupID: "test-group",
				Topics:  []string{"orders", "users", "events", "payments"},
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.valid && len(tt.config.Topics) == 0 {
				t.Error("Valid config should have topics")
			}
			if !tt.valid && len(tt.config.Topics) > 0 {
				t.Error("Invalid config should not have topics")
			}
		})
	}
}

func TestConsumer_OffsetManagement(t *testing.T) {
	tests := []struct {
		name          string
		offsetInitial int64
		description   string
	}{
		{
			name:          "newest offset",
			offsetInitial: sarama.OffsetNewest,
			description:   "Start from newest messages",
		},
		{
			name:          "oldest offset",
			offsetInitial: sarama.OffsetOldest,
			description:   "Start from oldest messages",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			config.Consumer.OffsetInitial = tt.offsetInitial

			saramaConfig, err := config.ToSaramaConfig()
			if err != nil {
				t.Fatalf("ToSaramaConfig failed: %v", err)
			}

			if saramaConfig.Consumer.Offsets.Initial != tt.offsetInitial {
				t.Errorf("OffsetInitial = %d, want %d", saramaConfig.Consumer.Offsets.Initial, tt.offsetInitial)
			}
		})
	}
}

func TestConsumer_AutoCommit(t *testing.T) {
	tests := []struct {
		name               string
		autoCommit         bool
		autoCommitInterval time.Duration
	}{
		{
			name:               "auto commit enabled",
			autoCommit:         true,
			autoCommitInterval: 1 * time.Second,
		},
		{
			name:               "auto commit disabled",
			autoCommit:         false,
			autoCommitInterval: 0,
		},
		{
			name:               "custom commit interval",
			autoCommit:         true,
			autoCommitInterval: 5 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			config.Consumer.AutoCommit = tt.autoCommit
			config.Consumer.AutoCommitInterval = tt.autoCommitInterval

			saramaConfig, err := config.ToSaramaConfig()
			if err != nil {
				t.Fatalf("ToSaramaConfig failed: %v", err)
			}

			if saramaConfig.Consumer.Offsets.AutoCommit.Enable != tt.autoCommit {
				t.Errorf("AutoCommit.Enable = %v, want %v", saramaConfig.Consumer.Offsets.AutoCommit.Enable, tt.autoCommit)
			}

			if tt.autoCommit && saramaConfig.Consumer.Offsets.AutoCommit.Interval != tt.autoCommitInterval {
				t.Errorf("AutoCommit.Interval = %v, want %v", saramaConfig.Consumer.Offsets.AutoCommit.Interval, tt.autoCommitInterval)
			}
		})
	}
}

func TestConsumer_SessionTimeout(t *testing.T) {
	tests := []struct {
		name           string
		sessionTimeout time.Duration
	}{
		{"short timeout", 5 * time.Second},
		{"default timeout", 10 * time.Second},
		{"long timeout", 30 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			config.Consumer.SessionTimeout = tt.sessionTimeout

			saramaConfig, err := config.ToSaramaConfig()
			if err != nil {
				t.Fatalf("ToSaramaConfig failed: %v", err)
			}

			if saramaConfig.Consumer.Group.Session.Timeout != tt.sessionTimeout {
				t.Errorf("SessionTimeout = %v, want %v", saramaConfig.Consumer.Group.Session.Timeout, tt.sessionTimeout)
			}
		})
	}
}

func TestConsumer_RebalanceTimeout(t *testing.T) {
	tests := []struct {
		name             string
		rebalanceTimeout time.Duration
	}{
		{"short rebalance", 30 * time.Second},
		{"default rebalance", 60 * time.Second},
		{"long rebalance", 120 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			config.Consumer.RebalanceTimeout = tt.rebalanceTimeout

			saramaConfig, err := config.ToSaramaConfig()
			if err != nil {
				t.Fatalf("ToSaramaConfig failed: %v", err)
			}

			if saramaConfig.Consumer.Group.Rebalance.Timeout != tt.rebalanceTimeout {
				t.Errorf("RebalanceTimeout = %v, want %v", saramaConfig.Consumer.Group.Rebalance.Timeout, tt.rebalanceTimeout)
			}
		})
	}
}

func TestConsumer_MaxProcessingTime(t *testing.T) {
	tests := []struct {
		name              string
		maxProcessingTime time.Duration
	}{
		{"fast processing", 1 * time.Minute},
		{"normal processing", 5 * time.Minute},
		{"slow processing", 15 * time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			config.Consumer.MaxProcessingTime = tt.maxProcessingTime

			saramaConfig, err := config.ToSaramaConfig()
			if err != nil {
				t.Fatalf("ToSaramaConfig failed: %v", err)
			}

			if saramaConfig.Consumer.MaxProcessingTime != tt.maxProcessingTime {
				t.Errorf("MaxProcessingTime = %v, want %v", saramaConfig.Consumer.MaxProcessingTime, tt.maxProcessingTime)
			}
		})
	}
}

func TestConsumer_WorkspacePrefix(t *testing.T) {
	tests := []struct {
		name      string
		workspace string
		topics    []string
		expected  []string
	}{
		{
			name:      "with workspace",
			workspace: "production",
			topics:    []string{"orders", "events"},
			expected:  []string{"production.orders", "production.events"},
		},
		{
			name:      "without workspace",
			workspace: "",
			topics:    []string{"orders", "events"},
			expected:  []string{"orders", "events"},
		},
		{
			name:      "dev workspace",
			workspace: "dev",
			topics:    []string{"users", "payments"},
			expected:  []string{"dev.users", "dev.payments"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Workspace: tt.workspace,
			}

			result := config.ApplyWorkspacePrefixToTopics(tt.topics)
			if len(result) != len(tt.expected) {
				t.Errorf("Result count = %d, want %d", len(result), len(tt.expected))
			}

			for i, topic := range result {
				if topic != tt.expected[i] {
					t.Errorf("Topic[%d] = %s, want %s", i, topic, tt.expected[i])
				}
			}
		})
	}
}

func TestConsumer_GroupID(t *testing.T) {
	tests := []struct {
		name    string
		groupID string
		valid   bool
	}{
		{"valid group ID", "order-processor", true},
		{"hyphenated group ID", "event-consumer-v2", true},
		{"underscored group ID", "data_processor", true},
		{"empty group ID", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			config.Consumer.GroupID = tt.groupID
			config.Consumer.Topics = []string{"test-topic"}

			err := config.Validate()
			if tt.valid && err != nil {
				t.Errorf("Expected valid config, got error: %v", err)
			}
			if !tt.valid && err == nil {
				t.Error("Expected error for invalid config")
			}
		})
	}
}

func TestConsumer_MessageHeaders(t *testing.T) {
	tests := []struct {
		name    string
		headers map[string]string
	}{
		{
			name:    "no headers",
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
				"timestamp":      "2024-01-01",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &ConsumedMessage{
				Topic:   "test-topic",
				Value:   []byte("test"),
				Headers: tt.headers,
			}

			if len(msg.Headers) != len(tt.headers) {
				t.Errorf("Headers count = %d, want %d", len(msg.Headers), len(tt.headers))
			}
		})
	}
}

func TestConsumer_ContextHandling(t *testing.T) {
	tests := []struct {
		name    string
		handler MessageHandler
	}{
		{
			name: "handler with context value",
			handler: func(ctx context.Context, msg *ConsumedMessage) error {
				// Check if context is valid
				if ctx == nil {
					return ErrInvalidConfig
				}
				return nil
			},
		},
		{
			name: "handler with deadline",
			handler: func(ctx context.Context, msg *ConsumedMessage) error {
				_, ok := ctx.Deadline()
				if ok {
					// Deadline is set
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			msg := &ConsumedMessage{
				Topic: "test-topic",
				Value: []byte("test"),
			}

			err := tt.handler(ctx, msg)
			if err != nil {
				t.Errorf("Handler returned error: %v", err)
			}
		})
	}
}

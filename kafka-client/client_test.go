package kafkaclient

import (
	"testing"
	"time"

	"github.com/IBM/sarama"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config == nil {
		t.Fatal("DefaultConfig returned nil")
	}

	if len(config.Brokers) == 0 {
		t.Error("Default config should have at least one broker")
	}

	if config.ClientID == "" {
		t.Error("ClientID should not be empty")
	}

	if config.Timeout <= 0 {
		t.Error("Timeout should be positive")
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		wantError bool
	}{
		{
			name:      "valid config",
			config:    DefaultConfig(),
			wantError: false,
		},
		{
			name: "empty brokers",
			config: &Config{
				Brokers:  []string{},
				ClientID: "test",
				Timeout:  30 * time.Second,
			},
			wantError: true,
		},
		{
			name: "empty client ID",
			config: &Config{
				Brokers:  []string{"localhost:9092"},
				ClientID: "",
				Timeout:  30 * time.Second,
			},
			wantError: true,
		},
		{
			name: "zero timeout",
			config: &Config{
				Brokers:  []string{"localhost:9092"},
				ClientID: "test",
				Timeout:  0,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestToSaramaConfig(t *testing.T) {
	config := DefaultConfig()
	saramaConfig, err := config.ToSaramaConfig()

	if err != nil {
		t.Fatalf("ToSaramaConfig() failed: %v", err)
	}

	if saramaConfig == nil {
		t.Fatal("ToSaramaConfig() returned nil")
	}

	if saramaConfig.ClientID != config.ClientID {
		t.Errorf("ClientID mismatch: got %s, want %s", saramaConfig.ClientID, config.ClientID)
	}
}

func TestMessageStruct(t *testing.T) {
	msg := &Message{
		Topic:     "test-topic",
		Key:       []byte("key"),
		Value:     []byte("value"),
		Headers:   map[string]string{"header1": "value1"},
		Partition: 0,
	}

	if msg.Topic != "test-topic" {
		t.Errorf("Topic = %s, want test-topic", msg.Topic)
	}

	if string(msg.Key) != "key" {
		t.Errorf("Key = %s, want key", string(msg.Key))
	}
}

func TestCompressionTypes(t *testing.T) {
	compressionTypes := []string{"none", "gzip", "snappy", "lz4", "zstd"}

	for _, compression := range compressionTypes {
		config := DefaultConfig()
		config.Producer.Compression = compression

		saramaConfig, err := config.ToSaramaConfig()
		if err != nil {
			t.Fatalf("ToSaramaConfig() failed for %s: %v", compression, err)
		}

		// Just verify it doesn't panic
		_ = saramaConfig.Producer.Compression
	}
}

func TestPartitionerTypes(t *testing.T) {
	partitioners := []string{"hash", "random", "roundrobin"}

	for _, partitioner := range partitioners {
		config := DefaultConfig()
		config.Producer.Partitioner = partitioner

		saramaConfig, err := config.ToSaramaConfig()
		if err != nil {
			t.Fatalf("ToSaramaConfig() failed for %s: %v", partitioner, err)
		}

		// Just verify it doesn't panic
		_ = saramaConfig.Producer.Partitioner
	}
}

func TestWithOptions(t *testing.T) {
	config := DefaultConfig()

	// Apply various options
	WithBrokers([]string{"broker1:9092", "broker2:9092"})(config)
	WithClientID("test-client")(config)
	WithTimeout(60 * time.Second)(config)
	WithCompression("gzip")(config)
	WithRequiredAcks(-1)(config)

	if len(config.Brokers) != 2 {
		t.Errorf("Brokers count = %d, want 2", len(config.Brokers))
	}

	if config.ClientID != "test-client" {
		t.Errorf("ClientID = %s, want test-client", config.ClientID)
	}

	if config.Timeout != 60*time.Second {
		t.Errorf("Timeout = %v, want 60s", config.Timeout)
	}
}

func TestOffsetInitial(t *testing.T) {
	config := DefaultConfig()

	// Test newest offset
	WithOffsetInitial(sarama.OffsetNewest)(config)
	if config.Consumer.OffsetInitial != sarama.OffsetNewest {
		t.Errorf("OffsetInitial = %d, want %d (OffsetNewest)", config.Consumer.OffsetInitial, sarama.OffsetNewest)
	}

	// Test oldest offset
	WithOffsetInitial(sarama.OffsetOldest)(config)
	if config.Consumer.OffsetInitial != sarama.OffsetOldest {
		t.Errorf("OffsetInitial = %d, want %d (OffsetOldest)", config.Consumer.OffsetInitial, sarama.OffsetOldest)
	}
}

func TestErrorHelpers(t *testing.T) {
	// Test IsConnectionError
	if !IsConnectionError(ErrConnectionFailed) {
		t.Error("IsConnectionError should return true for ErrConnectionFailed")
	}

	if !IsConnectionError(ErrClientClosed) {
		t.Error("IsConnectionError should return true for ErrClientClosed")
	}

	if IsConnectionError(ErrTimeout) {
		t.Error("IsConnectionError should return false for ErrTimeout")
	}

	// Test IsTimeoutError
	if !IsTimeoutError(ErrTimeout) {
		t.Error("IsTimeoutError should return true for ErrTimeout")
	}

	if IsTimeoutError(ErrConnectionFailed) {
		t.Error("IsTimeoutError should return false for ErrConnectionFailed")
	}
}

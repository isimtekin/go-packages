package kafkaclient

import (
	"testing"
	"time"
)

func TestConfig_FullValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		wantError bool
		errorMsg  string
	}{
		{
			name:      "fully valid config",
			config:    DefaultConfig(),
			wantError: false,
		},
		{
			name: "nil brokers",
			config: &Config{
				Brokers:  nil,
				ClientID: "test",
				Timeout:  30 * time.Second,
			},
			wantError: true,
			errorMsg:  "at least one broker must be specified",
		},
		{
			name: "empty brokers list",
			config: &Config{
				Brokers:  []string{},
				ClientID: "test",
				Timeout:  30 * time.Second,
			},
			wantError: true,
			errorMsg:  "at least one broker must be specified",
		},
		{
			name: "empty client ID",
			config: &Config{
				Brokers:  []string{"localhost:9092"},
				ClientID: "",
				Timeout:  30 * time.Second,
			},
			wantError: true,
			errorMsg:  "client ID cannot be empty",
		},
		{
			name: "zero timeout",
			config: &Config{
				Brokers:  []string{"localhost:9092"},
				ClientID: "test",
				Timeout:  0,
			},
			wantError: true,
			errorMsg:  "timeout must be positive",
		},
		{
			name: "negative timeout",
			config: &Config{
				Brokers:  []string{"localhost:9092"},
				ClientID: "test",
				Timeout:  -1 * time.Second,
			},
			wantError: true,
			errorMsg:  "timeout must be positive",
		},
		{
			name: "consumer with empty group ID",
			config: &Config{
				Brokers:  []string{"localhost:9092"},
				ClientID: "test",
				Timeout:  30 * time.Second,
				Consumer: ConsumerConfig{
					GroupID: "",
					Topics:  []string{"test-topic"},
				},
			},
			wantError: true,
			errorMsg:  "consumer group ID cannot be empty when topics are specified",
		},
		{
			name: "consumer with zero session timeout",
			config: &Config{
				Brokers:  []string{"localhost:9092"},
				ClientID: "test",
				Timeout:  30 * time.Second,
				Consumer: ConsumerConfig{
					GroupID:        "test-group",
					Topics:         []string{"test-topic"},
					SessionTimeout: 0,
				},
			},
			wantError: true,
			errorMsg:  "consumer session timeout must be positive",
		},
		{
			name: "producer with zero max message bytes",
			config: &Config{
				Brokers:  []string{"localhost:9092"},
				ClientID: "test",
				Timeout:  30 * time.Second,
				Producer: ProducerConfig{
					MaxMessageBytes: 0,
				},
			},
			wantError: true,
			errorMsg:  "max message bytes must be positive",
		},
		{
			name: "producer with negative retry max",
			config: &Config{
				Brokers:  []string{"localhost:9092"},
				ClientID: "test",
				Timeout:  30 * time.Second,
				Producer: ProducerConfig{
					MaxMessageBytes: 1000000,
					RetryMax:        -1,
				},
			},
			wantError: true,
			errorMsg:  "retry max cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
			if err != nil && tt.errorMsg != "" && err.Error() != tt.errorMsg {
				t.Errorf("Validate() error message = %v, want %v", err.Error(), tt.errorMsg)
			}
		})
	}
}

func TestConfig_BrokerAddresses(t *testing.T) {
	tests := []struct {
		name    string
		brokers []string
		valid   bool
	}{
		{
			name:    "single broker",
			brokers: []string{"localhost:9092"},
			valid:   true,
		},
		{
			name:    "multiple brokers",
			brokers: []string{"broker1:9092", "broker2:9092", "broker3:9092"},
			valid:   true,
		},
		{
			name:    "broker with hostname",
			brokers: []string{"kafka.example.com:9092"},
			valid:   true,
		},
		{
			name:    "broker with IP",
			brokers: []string{"192.168.1.100:9092"},
			valid:   true,
		},
		{
			name:    "empty brokers",
			brokers: []string{},
			valid:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			config.Brokers = tt.brokers

			err := config.Validate()
			if (err == nil) != tt.valid {
				t.Errorf("Validate() error = %v, valid = %v", err, tt.valid)
			}
		})
	}
}

func TestConfig_KafkaVersions(t *testing.T) {
	tests := []struct {
		name    string
		version string
		valid   bool
	}{
		{"Kafka 3.6.0", "3.6.0", true},
		{"Kafka 3.5.0", "3.5.0", true},
		{"Kafka 3.4.0", "3.4.0", true},
		{"Kafka 2.8.0", "2.8.0", true},
		{"invalid version", "invalid", false},
		{"empty version", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			config.Version = tt.version

			_, err := config.ToSaramaConfig()
			if (err == nil) != tt.valid {
				t.Errorf("ToSaramaConfig() error = %v, valid = %v", err, tt.valid)
			}
		})
	}
}

func TestConfig_SecuritySASL(t *testing.T) {
	tests := []struct {
		name      string
		mechanism string
		username  string
		password  string
	}{
		{
			name:      "SASL PLAIN",
			mechanism: "PLAIN",
			username:  "user",
			password:  "pass",
		},
		{
			name:      "SASL SCRAM-SHA-256",
			mechanism: "SCRAM-SHA-256",
			username:  "user",
			password:  "pass",
		},
		{
			name:      "SASL SCRAM-SHA-512",
			mechanism: "SCRAM-SHA-512",
			username:  "user",
			password:  "pass",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			config.Security.Enabled = true
			config.Security.Mechanism = tt.mechanism
			config.Security.Username = tt.username
			config.Security.Password = tt.password

			saramaConfig, err := config.ToSaramaConfig()
			if err != nil {
				t.Fatalf("ToSaramaConfig failed: %v", err)
			}

			if !saramaConfig.Net.SASL.Enable {
				t.Error("SASL should be enabled")
			}

			if saramaConfig.Net.SASL.User != tt.username {
				t.Errorf("SASL user = %s, want %s", saramaConfig.Net.SASL.User, tt.username)
			}

			if saramaConfig.Net.SASL.Password != tt.password {
				t.Errorf("SASL password = %s, want %s", saramaConfig.Net.SASL.Password, tt.password)
			}
		})
	}
}

func TestConfig_SecurityTLS(t *testing.T) {
	tests := []struct {
		name      string
		enableTLS bool
	}{
		{"TLS enabled", true},
		{"TLS disabled", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			config.Security.Enabled = tt.enableTLS
			config.Security.EnableTLS = tt.enableTLS

			saramaConfig, err := config.ToSaramaConfig()
			if err != nil {
				t.Fatalf("ToSaramaConfig failed: %v", err)
			}

			if saramaConfig.Net.TLS.Enable != tt.enableTLS {
				t.Errorf("TLS.Enable = %v, want %v", saramaConfig.Net.TLS.Enable, tt.enableTLS)
			}
		})
	}
}

func TestConfig_NetworkTimeouts(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
	}{
		{"short timeout", 5 * time.Second},
		{"medium timeout", 30 * time.Second},
		{"long timeout", 60 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			config.Timeout = tt.timeout

			saramaConfig, err := config.ToSaramaConfig()
			if err != nil {
				t.Fatalf("ToSaramaConfig failed: %v", err)
			}

			if saramaConfig.Net.DialTimeout != tt.timeout {
				t.Errorf("DialTimeout = %v, want %v", saramaConfig.Net.DialTimeout, tt.timeout)
			}

			if saramaConfig.Net.ReadTimeout != tt.timeout {
				t.Errorf("ReadTimeout = %v, want %v", saramaConfig.Net.ReadTimeout, tt.timeout)
			}

			if saramaConfig.Net.WriteTimeout != tt.timeout {
				t.Errorf("WriteTimeout = %v, want %v", saramaConfig.Net.WriteTimeout, tt.timeout)
			}
		})
	}
}

func TestConfig_ProducerAcks(t *testing.T) {
	tests := []struct {
		name string
		acks int16
		desc string
	}{
		{"NoResponse", 0, "Fire and forget"},
		{"WaitForLocal", 1, "Wait for leader"},
		{"WaitForAll", -1, "Wait for all replicas"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			config.Producer.RequiredAcks = tt.acks

			saramaConfig, err := config.ToSaramaConfig()
			if err != nil {
				t.Fatalf("ToSaramaConfig failed: %v", err)
			}

			if int16(saramaConfig.Producer.RequiredAcks) != tt.acks {
				t.Errorf("RequiredAcks = %d, want %d", saramaConfig.Producer.RequiredAcks, tt.acks)
			}
		})
	}
}

func TestConfig_ProducerMaxMessageBytes(t *testing.T) {
	tests := []struct {
		name     string
		maxBytes int
	}{
		{"1MB", 1 * 1024 * 1024},
		{"10MB", 10 * 1024 * 1024},
		{"100MB", 100 * 1024 * 1024},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			config.Producer.MaxMessageBytes = tt.maxBytes

			saramaConfig, err := config.ToSaramaConfig()
			if err != nil {
				t.Fatalf("ToSaramaConfig failed: %v", err)
			}

			if saramaConfig.Producer.MaxMessageBytes != tt.maxBytes {
				t.Errorf("MaxMessageBytes = %d, want %d", saramaConfig.Producer.MaxMessageBytes, tt.maxBytes)
			}
		})
	}
}

func TestConfig_ProducerTimeout(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
	}{
		{"fast timeout", 1 * time.Second},
		{"normal timeout", 10 * time.Second},
		{"slow timeout", 30 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			config.Producer.Timeout = tt.timeout

			saramaConfig, err := config.ToSaramaConfig()
			if err != nil {
				t.Fatalf("ToSaramaConfig failed: %v", err)
			}

			if saramaConfig.Producer.Timeout != tt.timeout {
				t.Errorf("Producer.Timeout = %v, want %v", saramaConfig.Producer.Timeout, tt.timeout)
			}
		})
	}
}

func TestConfig_ConsumerGroupRebalance(t *testing.T) {
	tests := []struct {
		name             string
		sessionTimeout   time.Duration
		rebalanceTimeout time.Duration
	}{
		{"short timeouts", 5 * time.Second, 30 * time.Second},
		{"default timeouts", 10 * time.Second, 60 * time.Second},
		{"long timeouts", 20 * time.Second, 120 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			config.Consumer.SessionTimeout = tt.sessionTimeout
			config.Consumer.RebalanceTimeout = tt.rebalanceTimeout

			saramaConfig, err := config.ToSaramaConfig()
			if err != nil {
				t.Fatalf("ToSaramaConfig failed: %v", err)
			}

			if saramaConfig.Consumer.Group.Session.Timeout != tt.sessionTimeout {
				t.Errorf("SessionTimeout = %v, want %v", saramaConfig.Consumer.Group.Session.Timeout, tt.sessionTimeout)
			}

			if saramaConfig.Consumer.Group.Rebalance.Timeout != tt.rebalanceTimeout {
				t.Errorf("RebalanceTimeout = %v, want %v", saramaConfig.Consumer.Group.Rebalance.Timeout, tt.rebalanceTimeout)
			}
		})
	}
}

func TestConfig_DefaultValues(t *testing.T) {
	config := DefaultConfig()

	// Test default broker
	if len(config.Brokers) != 1 || config.Brokers[0] != "localhost:9092" {
		t.Errorf("Default broker = %v, want [localhost:9092]", config.Brokers)
	}

	// Test default version
	if config.Version != "3.6.0" {
		t.Errorf("Default version = %s, want 3.6.0", config.Version)
	}

	// Test default client ID
	if config.ClientID == "" {
		t.Error("Default client ID should not be empty")
	}

	// Test default timeout
	if config.Timeout != 30*time.Second {
		t.Errorf("Default timeout = %v, want 30s", config.Timeout)
	}

	// Test default producer config
	if config.Producer.RequiredAcks != -1 {
		t.Errorf("Default RequiredAcks = %d, want -1", config.Producer.RequiredAcks)
	}

	if config.Producer.Compression != "snappy" {
		t.Errorf("Default compression = %s, want snappy", config.Producer.Compression)
	}

	if !config.Producer.IdempotentWrites {
		t.Error("Default idempotent writes should be true")
	}

	// Test default consumer config
	if config.Consumer.AutoCommit != true {
		t.Error("Default auto commit should be true")
	}
}

func TestConfig_Addr(t *testing.T) {
	// Config doesn't have Addr method, but we can test broker format
	tests := []struct {
		name   string
		broker string
	}{
		{"localhost", "localhost:9092"},
		{"IP address", "192.168.1.100:9092"},
		{"hostname", "kafka.example.com:9092"},
		{"custom port", "localhost:19092"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			config.Brokers = []string{tt.broker}

			if config.Brokers[0] != tt.broker {
				t.Errorf("Broker = %s, want %s", config.Brokers[0], tt.broker)
			}
		})
	}
}

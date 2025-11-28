package natsclient

import (
	"context"
	"testing"
	"time"
)

func TestStreamConfig_toNatsStreamConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    *StreamConfig
		wantName  string
		wantSubjs int
	}{
		{
			name: "basic config",
			config: &StreamConfig{
				Name:     "TEST",
				Subjects: []string{"test.>"},
			},
			wantName:  "TEST",
			wantSubjs: 1,
		},
		{
			name: "full config",
			config: &StreamConfig{
				Name:         "ORDERS",
				Description:  "Order stream",
				Subjects:     []string{"orders.>", "order.*"},
				Retention:    WorkQueuePolicy,
				MaxConsumers: 10,
				MaxMsgs:      1000,
				MaxBytes:     1024 * 1024,
				MaxAge:       24 * time.Hour,
				MaxMsgSize:   1024,
				Storage:      MemoryStorage,
				Replicas:     1,
				NoAck:        false,
				Duplicates:   2 * time.Minute,
			},
			wantName:  "ORDERS",
			wantSubjs: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.config.toNatsStreamConfig()
			if cfg.Name != tt.wantName {
				t.Errorf("Name = %s, want %s", cfg.Name, tt.wantName)
			}
			if len(cfg.Subjects) != tt.wantSubjs {
				t.Errorf("Subjects count = %d, want %d", len(cfg.Subjects), tt.wantSubjs)
			}
		})
	}
}

func TestConsumerConfig_toNatsConsumerConfig(t *testing.T) {
	startTime := time.Now()

	tests := []struct {
		name        string
		config      *ConsumerConfig
		wantDurable string
	}{
		{
			name: "durable consumer",
			config: &ConsumerConfig{
				Durable:       "my-consumer",
				DeliverPolicy: DeliverAll,
				AckPolicy:     AckExplicit,
				AckWait:       30 * time.Second,
				MaxDeliver:    5,
			},
			wantDurable: "my-consumer",
		},
		{
			name: "ephemeral consumer",
			config: &ConsumerConfig{
				DeliverPolicy:     DeliverNew,
				AckPolicy:         AckNone,
				InactiveThreshold: 5 * time.Minute,
			},
			wantDurable: "",
		},
		{
			name: "consumer with start time",
			config: &ConsumerConfig{
				Durable:       "time-consumer",
				DeliverPolicy: DeliverByStartTime,
				OptStartTime:  &startTime,
				AckPolicy:     AckAll,
			},
			wantDurable: "time-consumer",
		},
		{
			name: "consumer with start sequence",
			config: &ConsumerConfig{
				Durable:       "seq-consumer",
				DeliverPolicy: DeliverByStartSequence,
				OptStartSeq:   100,
				AckPolicy:     AckExplicit,
			},
			wantDurable: "seq-consumer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.config.toNatsConsumerConfig()
			if cfg.Durable != tt.wantDurable {
				t.Errorf("Durable = %s, want %s", cfg.Durable, tt.wantDurable)
			}
		})
	}
}

func TestRetentionPolicy(t *testing.T) {
	tests := []struct {
		policy   RetentionPolicy
		expected int
	}{
		{LimitsPolicy, 0},
		{InterestPolicy, 1},
		{WorkQueuePolicy, 2},
	}

	for _, tt := range tests {
		if int(tt.policy) != tt.expected {
			t.Errorf("RetentionPolicy %d != expected %d", tt.policy, tt.expected)
		}
	}
}

func TestStorageType(t *testing.T) {
	tests := []struct {
		storage  StorageType
		expected int
	}{
		{FileStorage, 0},
		{MemoryStorage, 1},
	}

	for _, tt := range tests {
		if int(tt.storage) != tt.expected {
			t.Errorf("StorageType %d != expected %d", tt.storage, tt.expected)
		}
	}
}

func TestDeliverPolicy(t *testing.T) {
	tests := []struct {
		policy   DeliverPolicy
		expected int
	}{
		{DeliverAll, 0},
		{DeliverLast, 1},
		{DeliverNew, 2},
		{DeliverByStartSequence, 3},
		{DeliverByStartTime, 4},
		{DeliverLastPerSubject, 5},
	}

	for _, tt := range tests {
		if int(tt.policy) != tt.expected {
			t.Errorf("DeliverPolicy %d != expected %d", tt.policy, tt.expected)
		}
	}
}

func TestAckPolicy(t *testing.T) {
	tests := []struct {
		policy   AckPolicy
		expected int
	}{
		{AckNone, 0},
		{AckAll, 1},
		{AckExplicit, 2},
	}

	for _, tt := range tests {
		if int(tt.policy) != tt.expected {
			t.Errorf("AckPolicy %d != expected %d", tt.policy, tt.expected)
		}
	}
}

func TestReplayPolicy(t *testing.T) {
	tests := []struct {
		policy   ReplayPolicy
		expected int
	}{
		{ReplayInstant, 0},
		{ReplayOriginal, 1},
	}

	for _, tt := range tests {
		if int(tt.policy) != tt.expected {
			t.Errorf("ReplayPolicy %d != expected %d", tt.policy, tt.expected)
		}
	}
}

func TestClient_getJetStream_Errors(t *testing.T) {
	t.Run("returns error when client is closed", func(t *testing.T) {
		client := &Client{
			config: DefaultConfig(),
			closed: true,
		}

		_, err := client.getJetStream()
		if err != ErrClientClosed {
			t.Errorf("getJetStream() error = %v, want %v", err, ErrClientClosed)
		}
	})

	t.Run("returns error when JetStream not enabled", func(t *testing.T) {
		config := DefaultConfig()
		config.EnableJetStream = false
		client := &Client{
			config: config,
			closed: false,
		}

		_, err := client.getJetStream()
		if err != ErrJetStreamNotEnabled {
			t.Errorf("getJetStream() error = %v, want %v", err, ErrJetStreamNotEnabled)
		}
	})

	t.Run("returns error when JetStream not initialized", func(t *testing.T) {
		config := DefaultConfig()
		config.EnableJetStream = true
		client := &Client{
			config: config,
			closed: false,
			js:     nil,
		}

		_, err := client.getJetStream()
		if err != ErrJetStreamNotInitialized {
			t.Errorf("getJetStream() error = %v, want %v", err, ErrJetStreamNotInitialized)
		}
	})
}

func TestClient_JetStreamOperations_ClosedClient(t *testing.T) {
	ctx := context.Background()
	client := &Client{
		config: DefaultConfig(),
		closed: true,
	}

	t.Run("CreateStream returns error when closed", func(t *testing.T) {
		_, err := client.CreateStream(ctx, &StreamConfig{Name: "TEST"})
		if err != ErrClientClosed {
			t.Errorf("CreateStream() error = %v, want %v", err, ErrClientClosed)
		}
	})

	t.Run("UpdateStream returns error when closed", func(t *testing.T) {
		_, err := client.UpdateStream(ctx, &StreamConfig{Name: "TEST"})
		if err != ErrClientClosed {
			t.Errorf("UpdateStream() error = %v, want %v", err, ErrClientClosed)
		}
	})

	t.Run("DeleteStream returns error when closed", func(t *testing.T) {
		err := client.DeleteStream(ctx, "TEST")
		if err != ErrClientClosed {
			t.Errorf("DeleteStream() error = %v, want %v", err, ErrClientClosed)
		}
	})

	t.Run("GetStream returns error when closed", func(t *testing.T) {
		_, err := client.GetStream(ctx, "TEST")
		if err != ErrClientClosed {
			t.Errorf("GetStream() error = %v, want %v", err, ErrClientClosed)
		}
	})

	t.Run("ListStreams returns error when closed", func(t *testing.T) {
		_, err := client.ListStreams(ctx)
		if err != ErrClientClosed {
			t.Errorf("ListStreams() error = %v, want %v", err, ErrClientClosed)
		}
	})

	t.Run("StreamNames returns error when closed", func(t *testing.T) {
		_, err := client.StreamNames(ctx)
		if err != ErrClientClosed {
			t.Errorf("StreamNames() error = %v, want %v", err, ErrClientClosed)
		}
	})

	t.Run("PurgeStream returns error when closed", func(t *testing.T) {
		err := client.PurgeStream(ctx, "TEST")
		if err != ErrClientClosed {
			t.Errorf("PurgeStream() error = %v, want %v", err, ErrClientClosed)
		}
	})

	t.Run("CreateConsumer returns error when closed", func(t *testing.T) {
		_, err := client.CreateConsumer(ctx, "TEST", &ConsumerConfig{Durable: "test"})
		if err != ErrClientClosed {
			t.Errorf("CreateConsumer() error = %v, want %v", err, ErrClientClosed)
		}
	})

	t.Run("UpdateConsumer returns error when closed", func(t *testing.T) {
		_, err := client.UpdateConsumer(ctx, "TEST", &ConsumerConfig{Durable: "test"})
		if err != ErrClientClosed {
			t.Errorf("UpdateConsumer() error = %v, want %v", err, ErrClientClosed)
		}
	})

	t.Run("DeleteConsumer returns error when closed", func(t *testing.T) {
		err := client.DeleteConsumer(ctx, "TEST", "consumer")
		if err != ErrClientClosed {
			t.Errorf("DeleteConsumer() error = %v, want %v", err, ErrClientClosed)
		}
	})

	t.Run("GetConsumer returns error when closed", func(t *testing.T) {
		_, err := client.GetConsumer(ctx, "TEST", "consumer")
		if err != ErrClientClosed {
			t.Errorf("GetConsumer() error = %v, want %v", err, ErrClientClosed)
		}
	})

	t.Run("ListConsumers returns error when closed", func(t *testing.T) {
		_, err := client.ListConsumers(ctx, "TEST")
		if err != ErrClientClosed {
			t.Errorf("ListConsumers() error = %v, want %v", err, ErrClientClosed)
		}
	})

	t.Run("ConsumerNames returns error when closed", func(t *testing.T) {
		_, err := client.ConsumerNames(ctx, "TEST")
		if err != ErrClientClosed {
			t.Errorf("ConsumerNames() error = %v, want %v", err, ErrClientClosed)
		}
	})

	t.Run("PublishToStream returns error when closed", func(t *testing.T) {
		_, err := client.PublishToStream(ctx, "test.subject", []byte("data"))
		if err != ErrClientClosed {
			t.Errorf("PublishToStream() error = %v, want %v", err, ErrClientClosed)
		}
	})

	t.Run("PublishToStreamAsync returns error when closed", func(t *testing.T) {
		_, err := client.PublishToStreamAsync("test.subject", []byte("data"))
		if err != ErrClientClosed {
			t.Errorf("PublishToStreamAsync() error = %v, want %v", err, ErrClientClosed)
		}
	})

	t.Run("PublishMsgToStream returns error when closed", func(t *testing.T) {
		_, err := client.PublishMsgToStream(ctx, nil)
		if err != ErrClientClosed {
			t.Errorf("PublishMsgToStream() error = %v, want %v", err, ErrClientClosed)
		}
	})

	t.Run("PullSubscribe returns error when closed", func(t *testing.T) {
		_, err := client.PullSubscribe("test.subject", "durable")
		if err != ErrClientClosed {
			t.Errorf("PullSubscribe() error = %v, want %v", err, ErrClientClosed)
		}
	})

	t.Run("SubscribeToStream returns error when closed", func(t *testing.T) {
		_, err := client.SubscribeToStream("test.subject", nil)
		if err != ErrClientClosed {
			t.Errorf("SubscribeToStream() error = %v, want %v", err, ErrClientClosed)
		}
	})

	t.Run("QueueSubscribeToStream returns error when closed", func(t *testing.T) {
		_, err := client.QueueSubscribeToStream("test.subject", "queue", nil)
		if err != ErrClientClosed {
			t.Errorf("QueueSubscribeToStream() error = %v, want %v", err, ErrClientClosed)
		}
	})

	t.Run("ChanSubscribeToStream returns error when closed", func(t *testing.T) {
		_, err := client.ChanSubscribeToStream("test.subject", nil)
		if err != ErrClientClosed {
			t.Errorf("ChanSubscribeToStream() error = %v, want %v", err, ErrClientClosed)
		}
	})

	t.Run("GetMsg returns error when closed", func(t *testing.T) {
		_, err := client.GetMsg(ctx, "TEST", 1)
		if err != ErrClientClosed {
			t.Errorf("GetMsg() error = %v, want %v", err, ErrClientClosed)
		}
	})

	t.Run("GetLastMsg returns error when closed", func(t *testing.T) {
		_, err := client.GetLastMsg(ctx, "TEST", "subject")
		if err != ErrClientClosed {
			t.Errorf("GetLastMsg() error = %v, want %v", err, ErrClientClosed)
		}
	})

	t.Run("DeleteMsg returns error when closed", func(t *testing.T) {
		err := client.DeleteMsg(ctx, "TEST", 1)
		if err != ErrClientClosed {
			t.Errorf("DeleteMsg() error = %v, want %v", err, ErrClientClosed)
		}
	})

	t.Run("SecureDeleteMsg returns error when closed", func(t *testing.T) {
		err := client.SecureDeleteMsg(ctx, "TEST", 1)
		if err != ErrClientClosed {
			t.Errorf("SecureDeleteMsg() error = %v, want %v", err, ErrClientClosed)
		}
	})
}

func TestClient_JetStreamOperations_NotEnabled(t *testing.T) {
	ctx := context.Background()
	config := DefaultConfig()
	config.EnableJetStream = false
	client := &Client{
		config: config,
		closed: false,
	}

	t.Run("CreateStream returns error when not enabled", func(t *testing.T) {
		_, err := client.CreateStream(ctx, &StreamConfig{Name: "TEST"})
		if err != ErrJetStreamNotEnabled {
			t.Errorf("CreateStream() error = %v, want %v", err, ErrJetStreamNotEnabled)
		}
	})

	t.Run("UpdateStream returns error when not enabled", func(t *testing.T) {
		_, err := client.UpdateStream(ctx, &StreamConfig{Name: "TEST"})
		if err != ErrJetStreamNotEnabled {
			t.Errorf("UpdateStream() error = %v, want %v", err, ErrJetStreamNotEnabled)
		}
	})

	t.Run("DeleteStream returns error when not enabled", func(t *testing.T) {
		err := client.DeleteStream(ctx, "TEST")
		if err != ErrJetStreamNotEnabled {
			t.Errorf("DeleteStream() error = %v, want %v", err, ErrJetStreamNotEnabled)
		}
	})

	t.Run("GetStream returns error when not enabled", func(t *testing.T) {
		_, err := client.GetStream(ctx, "TEST")
		if err != ErrJetStreamNotEnabled {
			t.Errorf("GetStream() error = %v, want %v", err, ErrJetStreamNotEnabled)
		}
	})

	t.Run("ListStreams returns error when not enabled", func(t *testing.T) {
		_, err := client.ListStreams(ctx)
		if err != ErrJetStreamNotEnabled {
			t.Errorf("ListStreams() error = %v, want %v", err, ErrJetStreamNotEnabled)
		}
	})

	t.Run("StreamNames returns error when not enabled", func(t *testing.T) {
		_, err := client.StreamNames(ctx)
		if err != ErrJetStreamNotEnabled {
			t.Errorf("StreamNames() error = %v, want %v", err, ErrJetStreamNotEnabled)
		}
	})

	t.Run("PurgeStream returns error when not enabled", func(t *testing.T) {
		err := client.PurgeStream(ctx, "TEST")
		if err != ErrJetStreamNotEnabled {
			t.Errorf("PurgeStream() error = %v, want %v", err, ErrJetStreamNotEnabled)
		}
	})

	t.Run("CreateConsumer returns error when not enabled", func(t *testing.T) {
		_, err := client.CreateConsumer(ctx, "TEST", &ConsumerConfig{})
		if err != ErrJetStreamNotEnabled {
			t.Errorf("CreateConsumer() error = %v, want %v", err, ErrJetStreamNotEnabled)
		}
	})

	t.Run("UpdateConsumer returns error when not enabled", func(t *testing.T) {
		_, err := client.UpdateConsumer(ctx, "TEST", &ConsumerConfig{})
		if err != ErrJetStreamNotEnabled {
			t.Errorf("UpdateConsumer() error = %v, want %v", err, ErrJetStreamNotEnabled)
		}
	})

	t.Run("DeleteConsumer returns error when not enabled", func(t *testing.T) {
		err := client.DeleteConsumer(ctx, "TEST", "consumer")
		if err != ErrJetStreamNotEnabled {
			t.Errorf("DeleteConsumer() error = %v, want %v", err, ErrJetStreamNotEnabled)
		}
	})

	t.Run("GetConsumer returns error when not enabled", func(t *testing.T) {
		_, err := client.GetConsumer(ctx, "TEST", "consumer")
		if err != ErrJetStreamNotEnabled {
			t.Errorf("GetConsumer() error = %v, want %v", err, ErrJetStreamNotEnabled)
		}
	})

	t.Run("ListConsumers returns error when not enabled", func(t *testing.T) {
		_, err := client.ListConsumers(ctx, "TEST")
		if err != ErrJetStreamNotEnabled {
			t.Errorf("ListConsumers() error = %v, want %v", err, ErrJetStreamNotEnabled)
		}
	})

	t.Run("ConsumerNames returns error when not enabled", func(t *testing.T) {
		_, err := client.ConsumerNames(ctx, "TEST")
		if err != ErrJetStreamNotEnabled {
			t.Errorf("ConsumerNames() error = %v, want %v", err, ErrJetStreamNotEnabled)
		}
	})

	t.Run("PublishToStream returns error when not enabled", func(t *testing.T) {
		_, err := client.PublishToStream(ctx, "test.subject", []byte("data"))
		if err != ErrJetStreamNotEnabled {
			t.Errorf("PublishToStream() error = %v, want %v", err, ErrJetStreamNotEnabled)
		}
	})

	t.Run("PublishToStreamAsync returns error when not enabled", func(t *testing.T) {
		_, err := client.PublishToStreamAsync("test.subject", []byte("data"))
		if err != ErrJetStreamNotEnabled {
			t.Errorf("PublishToStreamAsync() error = %v, want %v", err, ErrJetStreamNotEnabled)
		}
	})

	t.Run("PublishMsgToStream returns error when not enabled", func(t *testing.T) {
		_, err := client.PublishMsgToStream(ctx, nil)
		if err != ErrJetStreamNotEnabled {
			t.Errorf("PublishMsgToStream() error = %v, want %v", err, ErrJetStreamNotEnabled)
		}
	})

	t.Run("PullSubscribe returns error when not enabled", func(t *testing.T) {
		_, err := client.PullSubscribe("test", "durable")
		if err != ErrJetStreamNotEnabled {
			t.Errorf("PullSubscribe() error = %v, want %v", err, ErrJetStreamNotEnabled)
		}
	})

	t.Run("SubscribeToStream returns error when not enabled", func(t *testing.T) {
		_, err := client.SubscribeToStream("test", nil)
		if err != ErrJetStreamNotEnabled {
			t.Errorf("SubscribeToStream() error = %v, want %v", err, ErrJetStreamNotEnabled)
		}
	})

	t.Run("QueueSubscribeToStream returns error when not enabled", func(t *testing.T) {
		_, err := client.QueueSubscribeToStream("test", "queue", nil)
		if err != ErrJetStreamNotEnabled {
			t.Errorf("QueueSubscribeToStream() error = %v, want %v", err, ErrJetStreamNotEnabled)
		}
	})

	t.Run("ChanSubscribeToStream returns error when not enabled", func(t *testing.T) {
		_, err := client.ChanSubscribeToStream("test", nil)
		if err != ErrJetStreamNotEnabled {
			t.Errorf("ChanSubscribeToStream() error = %v, want %v", err, ErrJetStreamNotEnabled)
		}
	})

	t.Run("GetMsg returns error when not enabled", func(t *testing.T) {
		_, err := client.GetMsg(ctx, "TEST", 1)
		if err != ErrJetStreamNotEnabled {
			t.Errorf("GetMsg() error = %v, want %v", err, ErrJetStreamNotEnabled)
		}
	})

	t.Run("GetLastMsg returns error when not enabled", func(t *testing.T) {
		_, err := client.GetLastMsg(ctx, "TEST", "subject")
		if err != ErrJetStreamNotEnabled {
			t.Errorf("GetLastMsg() error = %v, want %v", err, ErrJetStreamNotEnabled)
		}
	})

	t.Run("DeleteMsg returns error when not enabled", func(t *testing.T) {
		err := client.DeleteMsg(ctx, "TEST", 1)
		if err != ErrJetStreamNotEnabled {
			t.Errorf("DeleteMsg() error = %v, want %v", err, ErrJetStreamNotEnabled)
		}
	})

	t.Run("SecureDeleteMsg returns error when not enabled", func(t *testing.T) {
		err := client.SecureDeleteMsg(ctx, "TEST", 1)
		if err != ErrJetStreamNotEnabled {
			t.Errorf("SecureDeleteMsg() error = %v, want %v", err, ErrJetStreamNotEnabled)
		}
	})
}

func TestClient_JetStreamOperations_NotInitialized(t *testing.T) {
	ctx := context.Background()
	config := DefaultConfig()
	config.EnableJetStream = true
	client := &Client{
		config: config,
		closed: false,
		js:     nil, // JetStream enabled but not initialized
	}

	t.Run("CreateStream returns error when not initialized", func(t *testing.T) {
		_, err := client.CreateStream(ctx, &StreamConfig{Name: "TEST"})
		if err != ErrJetStreamNotInitialized {
			t.Errorf("CreateStream() error = %v, want %v", err, ErrJetStreamNotInitialized)
		}
	})

	t.Run("UpdateStream returns error when not initialized", func(t *testing.T) {
		_, err := client.UpdateStream(ctx, &StreamConfig{Name: "TEST"})
		if err != ErrJetStreamNotInitialized {
			t.Errorf("UpdateStream() error = %v, want %v", err, ErrJetStreamNotInitialized)
		}
	})

	t.Run("DeleteStream returns error when not initialized", func(t *testing.T) {
		err := client.DeleteStream(ctx, "TEST")
		if err != ErrJetStreamNotInitialized {
			t.Errorf("DeleteStream() error = %v, want %v", err, ErrJetStreamNotInitialized)
		}
	})

	t.Run("GetStream returns error when not initialized", func(t *testing.T) {
		_, err := client.GetStream(ctx, "TEST")
		if err != ErrJetStreamNotInitialized {
			t.Errorf("GetStream() error = %v, want %v", err, ErrJetStreamNotInitialized)
		}
	})

	t.Run("ListStreams returns error when not initialized", func(t *testing.T) {
		_, err := client.ListStreams(ctx)
		if err != ErrJetStreamNotInitialized {
			t.Errorf("ListStreams() error = %v, want %v", err, ErrJetStreamNotInitialized)
		}
	})

	t.Run("StreamNames returns error when not initialized", func(t *testing.T) {
		_, err := client.StreamNames(ctx)
		if err != ErrJetStreamNotInitialized {
			t.Errorf("StreamNames() error = %v, want %v", err, ErrJetStreamNotInitialized)
		}
	})

	t.Run("PurgeStream returns error when not initialized", func(t *testing.T) {
		err := client.PurgeStream(ctx, "TEST")
		if err != ErrJetStreamNotInitialized {
			t.Errorf("PurgeStream() error = %v, want %v", err, ErrJetStreamNotInitialized)
		}
	})

	t.Run("CreateConsumer returns error when not initialized", func(t *testing.T) {
		_, err := client.CreateConsumer(ctx, "TEST", &ConsumerConfig{})
		if err != ErrJetStreamNotInitialized {
			t.Errorf("CreateConsumer() error = %v, want %v", err, ErrJetStreamNotInitialized)
		}
	})

	t.Run("UpdateConsumer returns error when not initialized", func(t *testing.T) {
		_, err := client.UpdateConsumer(ctx, "TEST", &ConsumerConfig{})
		if err != ErrJetStreamNotInitialized {
			t.Errorf("UpdateConsumer() error = %v, want %v", err, ErrJetStreamNotInitialized)
		}
	})

	t.Run("DeleteConsumer returns error when not initialized", func(t *testing.T) {
		err := client.DeleteConsumer(ctx, "TEST", "consumer")
		if err != ErrJetStreamNotInitialized {
			t.Errorf("DeleteConsumer() error = %v, want %v", err, ErrJetStreamNotInitialized)
		}
	})

	t.Run("GetConsumer returns error when not initialized", func(t *testing.T) {
		_, err := client.GetConsumer(ctx, "TEST", "consumer")
		if err != ErrJetStreamNotInitialized {
			t.Errorf("GetConsumer() error = %v, want %v", err, ErrJetStreamNotInitialized)
		}
	})

	t.Run("ListConsumers returns error when not initialized", func(t *testing.T) {
		_, err := client.ListConsumers(ctx, "TEST")
		if err != ErrJetStreamNotInitialized {
			t.Errorf("ListConsumers() error = %v, want %v", err, ErrJetStreamNotInitialized)
		}
	})

	t.Run("ConsumerNames returns error when not initialized", func(t *testing.T) {
		_, err := client.ConsumerNames(ctx, "TEST")
		if err != ErrJetStreamNotInitialized {
			t.Errorf("ConsumerNames() error = %v, want %v", err, ErrJetStreamNotInitialized)
		}
	})

	t.Run("PublishToStream returns error when not initialized", func(t *testing.T) {
		_, err := client.PublishToStream(ctx, "test.subject", []byte("data"))
		if err != ErrJetStreamNotInitialized {
			t.Errorf("PublishToStream() error = %v, want %v", err, ErrJetStreamNotInitialized)
		}
	})

	t.Run("PublishToStreamAsync returns error when not initialized", func(t *testing.T) {
		_, err := client.PublishToStreamAsync("test.subject", []byte("data"))
		if err != ErrJetStreamNotInitialized {
			t.Errorf("PublishToStreamAsync() error = %v, want %v", err, ErrJetStreamNotInitialized)
		}
	})

	t.Run("PublishMsgToStream returns error when not initialized", func(t *testing.T) {
		_, err := client.PublishMsgToStream(ctx, nil)
		if err != ErrJetStreamNotInitialized {
			t.Errorf("PublishMsgToStream() error = %v, want %v", err, ErrJetStreamNotInitialized)
		}
	})

	t.Run("PullSubscribe returns error when not initialized", func(t *testing.T) {
		_, err := client.PullSubscribe("test", "durable")
		if err != ErrJetStreamNotInitialized {
			t.Errorf("PullSubscribe() error = %v, want %v", err, ErrJetStreamNotInitialized)
		}
	})

	t.Run("SubscribeToStream returns error when not initialized", func(t *testing.T) {
		_, err := client.SubscribeToStream("test", nil)
		if err != ErrJetStreamNotInitialized {
			t.Errorf("SubscribeToStream() error = %v, want %v", err, ErrJetStreamNotInitialized)
		}
	})

	t.Run("QueueSubscribeToStream returns error when not initialized", func(t *testing.T) {
		_, err := client.QueueSubscribeToStream("test", "queue", nil)
		if err != ErrJetStreamNotInitialized {
			t.Errorf("QueueSubscribeToStream() error = %v, want %v", err, ErrJetStreamNotInitialized)
		}
	})

	t.Run("ChanSubscribeToStream returns error when not initialized", func(t *testing.T) {
		_, err := client.ChanSubscribeToStream("test", nil)
		if err != ErrJetStreamNotInitialized {
			t.Errorf("ChanSubscribeToStream() error = %v, want %v", err, ErrJetStreamNotInitialized)
		}
	})

	t.Run("GetMsg returns error when not initialized", func(t *testing.T) {
		_, err := client.GetMsg(ctx, "TEST", 1)
		if err != ErrJetStreamNotInitialized {
			t.Errorf("GetMsg() error = %v, want %v", err, ErrJetStreamNotInitialized)
		}
	})

	t.Run("GetLastMsg returns error when not initialized", func(t *testing.T) {
		_, err := client.GetLastMsg(ctx, "TEST", "subject")
		if err != ErrJetStreamNotInitialized {
			t.Errorf("GetLastMsg() error = %v, want %v", err, ErrJetStreamNotInitialized)
		}
	})

	t.Run("DeleteMsg returns error when not initialized", func(t *testing.T) {
		err := client.DeleteMsg(ctx, "TEST", 1)
		if err != ErrJetStreamNotInitialized {
			t.Errorf("DeleteMsg() error = %v, want %v", err, ErrJetStreamNotInitialized)
		}
	})

	t.Run("SecureDeleteMsg returns error when not initialized", func(t *testing.T) {
		err := client.SecureDeleteMsg(ctx, "TEST", 1)
		if err != ErrJetStreamNotInitialized {
			t.Errorf("SecureDeleteMsg() error = %v, want %v", err, ErrJetStreamNotInitialized)
		}
	})
}

func TestClient_PublishAckFutureWait(t *testing.T) {
	t.Run("returns closed channel when js is nil", func(t *testing.T) {
		client := &Client{
			config: DefaultConfig(),
			closed: false,
			js:     nil,
		}

		ch := client.PublishAckFutureWait()
		select {
		case <-ch:
			// Expected - channel should be closed immediately
		case <-time.After(100 * time.Millisecond):
			t.Error("PublishAckFutureWait() should return closed channel immediately")
		}
	})
}

func TestIsJetStreamError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"JetStreamNotEnabled", ErrJetStreamNotEnabled, true},
		{"JetStreamNotInitialized", ErrJetStreamNotInitialized, true},
		{"StreamNotFound", ErrStreamNotFound, true},
		{"ConsumerNotFound", ErrConsumerNotFound, true},
		{"MessageNotFound", ErrMessageNotFound, true},
		{"ClientClosed", ErrClientClosed, false},
		{"ConnectionFailed", ErrConnectionFailed, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsJetStreamError(tt.err); got != tt.want {
				t.Errorf("IsJetStreamError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsNotFoundError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"StreamNotFound", ErrStreamNotFound, true},
		{"ConsumerNotFound", ErrConsumerNotFound, true},
		{"MessageNotFound", ErrMessageNotFound, true},
		{"JetStreamNotEnabled", ErrJetStreamNotEnabled, false},
		{"ClientClosed", ErrClientClosed, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNotFoundError(tt.err); got != tt.want {
				t.Errorf("IsNotFoundError() = %v, want %v", got, tt.want)
			}
		})
	}
}

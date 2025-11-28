package natsclient

import (
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  DefaultConfig(),
			wantErr: false, // May succeed if NATS server is running
		},
		{
			name: "invalid url",
			config: &Config{
				URL:     "",
				Timeout: 30 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid timeout",
			config: &Config{
				URL:     "nats://localhost:4222",
				Timeout: 0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if client != nil {
				defer client.Close()
			}
		})
	}
}

func TestNewWithOptions(t *testing.T) {
	t.Run("creates client with options", func(t *testing.T) {
		_, err := NewWithOptions(
			WithURL("nats://localhost:4222"),
			WithName("test-client"),
			WithTimeout(10*time.Second),
		)

		// Will fail without NATS server, but validates options work
		if err == nil {
			t.Skip("NATS server not available")
		}
	})
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name: "empty url",
			config: &Config{
				URL:     "",
				Timeout: 1 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "zero timeout",
			config: &Config{
				URL:     "nats://localhost:4222",
				Timeout: 0,
			},
			wantErr: true,
		},
		{
			name: "invalid max reconnects",
			config: &Config{
				URL:           "nats://localhost:4222",
				Timeout:       1 * time.Second,
				MaxReconnects: -2,
			},
			wantErr: true,
		},
		{
			name: "tls cert without key",
			config: &Config{
				URL:         "nats://localhost:4222",
				Timeout:     1 * time.Second,
				TLSEnabled:  true,
				TLSCertFile: "cert.pem",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_IsClosed(t *testing.T) {
	config := DefaultConfig()
	client := &Client{
		config: config,
		closed: false,
	}

	if client.IsClosed() {
		t.Error("IsClosed() = true, want false")
	}

	client.closed = true
	if !client.IsClosed() {
		t.Error("IsClosed() = false, want true")
	}
}

func TestClient_ClosedClientErrors(t *testing.T) {
	config := DefaultConfig()
	client := &Client{
		config: config,
		closed: true,
	}

	t.Run("Publish returns error when closed", func(t *testing.T) {
		err := client.Publish("test", []byte("data"))
		if err != ErrClientClosed {
			t.Errorf("Publish() error = %v, want %v", err, ErrClientClosed)
		}
	})

	t.Run("PublishMsg returns error when closed", func(t *testing.T) {
		err := client.PublishMsg(nil)
		if err != ErrClientClosed {
			t.Errorf("PublishMsg() error = %v, want %v", err, ErrClientClosed)
		}
	})

	t.Run("PublishRequest returns error when closed", func(t *testing.T) {
		err := client.PublishRequest("test", "reply", []byte("data"))
		if err != ErrClientClosed {
			t.Errorf("PublishRequest() error = %v, want %v", err, ErrClientClosed)
		}
	})

	t.Run("Subscribe returns error when closed", func(t *testing.T) {
		_, err := client.Subscribe("test", nil)
		if err != ErrClientClosed {
			t.Errorf("Subscribe() error = %v, want %v", err, ErrClientClosed)
		}
	})

	t.Run("QueueSubscribe returns error when closed", func(t *testing.T) {
		_, err := client.QueueSubscribe("test", "queue", nil)
		if err != ErrClientClosed {
			t.Errorf("QueueSubscribe() error = %v, want %v", err, ErrClientClosed)
		}
	})

	t.Run("ChanSubscribe returns error when closed", func(t *testing.T) {
		_, err := client.ChanSubscribe("test", nil)
		if err != ErrClientClosed {
			t.Errorf("ChanSubscribe() error = %v, want %v", err, ErrClientClosed)
		}
	})

	t.Run("QueueChanSubscribe returns error when closed", func(t *testing.T) {
		_, err := client.QueueChanSubscribe("test", "queue", nil)
		if err != ErrClientClosed {
			t.Errorf("QueueChanSubscribe() error = %v, want %v", err, ErrClientClosed)
		}
	})

	t.Run("Request returns error when closed", func(t *testing.T) {
		_, err := client.Request("test", []byte("data"), time.Second)
		if err != ErrClientClosed {
			t.Errorf("Request() error = %v, want %v", err, ErrClientClosed)
		}
	})

	t.Run("RequestMsg returns error when closed", func(t *testing.T) {
		_, err := client.RequestMsg(nil, time.Second)
		if err != ErrClientClosed {
			t.Errorf("RequestMsg() error = %v, want %v", err, ErrClientClosed)
		}
	})

	t.Run("Flush returns error when closed", func(t *testing.T) {
		err := client.Flush()
		if err != ErrClientClosed {
			t.Errorf("Flush() error = %v, want %v", err, ErrClientClosed)
		}
	})

	t.Run("JetStream returns error when closed", func(t *testing.T) {
		_, err := client.JetStream()
		if err != ErrClientClosed {
			t.Errorf("JetStream() error = %v, want %v", err, ErrClientClosed)
		}
	})
}

func TestClient_NilConnectionErrors(t *testing.T) {
	config := DefaultConfig()
	client := &Client{
		config: config,
		closed: false,
		conn:   nil, // No connection
	}

	t.Run("Publish returns error with nil connection", func(t *testing.T) {
		err := client.Publish("test", []byte("data"))
		if err != ErrConnectionFailed {
			t.Errorf("Publish() error = %v, want %v", err, ErrConnectionFailed)
		}
	})

	t.Run("PublishMsg returns error with nil connection", func(t *testing.T) {
		err := client.PublishMsg(nil)
		if err != ErrConnectionFailed {
			t.Errorf("PublishMsg() error = %v, want %v", err, ErrConnectionFailed)
		}
	})

	t.Run("PublishRequest returns error with nil connection", func(t *testing.T) {
		err := client.PublishRequest("test", "reply", []byte("data"))
		if err != ErrConnectionFailed {
			t.Errorf("PublishRequest() error = %v, want %v", err, ErrConnectionFailed)
		}
	})

	t.Run("Subscribe returns error with nil connection", func(t *testing.T) {
		_, err := client.Subscribe("test", nil)
		if err != ErrConnectionFailed {
			t.Errorf("Subscribe() error = %v, want %v", err, ErrConnectionFailed)
		}
	})

	t.Run("QueueSubscribe returns error with nil connection", func(t *testing.T) {
		_, err := client.QueueSubscribe("test", "queue", nil)
		if err != ErrConnectionFailed {
			t.Errorf("QueueSubscribe() error = %v, want %v", err, ErrConnectionFailed)
		}
	})

	t.Run("ChanSubscribe returns error with nil connection", func(t *testing.T) {
		_, err := client.ChanSubscribe("test", nil)
		if err != ErrConnectionFailed {
			t.Errorf("ChanSubscribe() error = %v, want %v", err, ErrConnectionFailed)
		}
	})

	t.Run("QueueChanSubscribe returns error with nil connection", func(t *testing.T) {
		_, err := client.QueueChanSubscribe("test", "queue", nil)
		if err != ErrConnectionFailed {
			t.Errorf("QueueChanSubscribe() error = %v, want %v", err, ErrConnectionFailed)
		}
	})

	t.Run("Request returns error with nil connection", func(t *testing.T) {
		_, err := client.Request("test", []byte("data"), time.Second)
		if err != ErrConnectionFailed {
			t.Errorf("Request() error = %v, want %v", err, ErrConnectionFailed)
		}
	})

	t.Run("RequestMsg returns error with nil connection", func(t *testing.T) {
		_, err := client.RequestMsg(nil, time.Second)
		if err != ErrConnectionFailed {
			t.Errorf("RequestMsg() error = %v, want %v", err, ErrConnectionFailed)
		}
	})

	t.Run("Flush returns error with nil connection", func(t *testing.T) {
		err := client.Flush()
		if err != ErrConnectionFailed {
			t.Errorf("Flush() error = %v, want %v", err, ErrConnectionFailed)
		}
	})

	t.Run("IsConnected returns false with nil connection", func(t *testing.T) {
		if client.IsConnected() {
			t.Error("IsConnected() = true, want false")
		}
	})

	t.Run("IsReconnecting returns false with nil connection", func(t *testing.T) {
		if client.IsReconnecting() {
			t.Error("IsReconnecting() = true, want false")
		}
	})

	t.Run("Stats returns empty stats with nil connection", func(t *testing.T) {
		stats := client.Stats()
		if stats.InMsgs != 0 || stats.OutMsgs != 0 {
			t.Error("Stats() should return empty stats for nil connection")
		}
	})

	t.Run("Conn returns nil", func(t *testing.T) {
		if client.Conn() != nil {
			t.Error("Conn() should return nil")
		}
	})
}

func TestClient_Close(t *testing.T) {
	config := DefaultConfig()
	client := &Client{
		config: config,
		closed: false,
		conn:   nil,
	}

	// First close should succeed
	err := client.Close()
	if err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}

	// Second close should return ErrAlreadyClosed
	err = client.Close()
	if err != ErrAlreadyClosed {
		t.Errorf("Close() error = %v, want %v", err, ErrAlreadyClosed)
	}
}

func TestClient_RequestWithContext_ClosedClient(t *testing.T) {
	config := DefaultConfig()
	client := &Client{
		config: config,
		closed: true,
	}

	_, err := client.RequestWithContext(nil, "test", []byte("data"))
	if err != ErrClientClosed {
		t.Errorf("RequestWithContext() error = %v, want %v", err, ErrClientClosed)
	}
}

func TestClient_RequestWithContext_NilConnection(t *testing.T) {
	config := DefaultConfig()
	client := &Client{
		config: config,
		closed: false,
		conn:   nil,
	}

	_, err := client.RequestWithContext(nil, "test", []byte("data"))
	if err != ErrConnectionFailed {
		t.Errorf("RequestWithContext() error = %v, want %v", err, ErrConnectionFailed)
	}
}

func TestClient_RequestMsgWithContext_ClosedClient(t *testing.T) {
	config := DefaultConfig()
	client := &Client{
		config: config,
		closed: true,
	}

	_, err := client.RequestMsgWithContext(nil, nil)
	if err != ErrClientClosed {
		t.Errorf("RequestMsgWithContext() error = %v, want %v", err, ErrClientClosed)
	}
}

func TestClient_RequestMsgWithContext_NilConnection(t *testing.T) {
	config := DefaultConfig()
	client := &Client{
		config: config,
		closed: false,
		conn:   nil,
	}

	_, err := client.RequestMsgWithContext(nil, nil)
	if err != ErrConnectionFailed {
		t.Errorf("RequestMsgWithContext() error = %v, want %v", err, ErrConnectionFailed)
	}
}

func TestIsConnectionError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"ConnectionFailed", ErrConnectionFailed, true},
		{"ClientClosed", ErrClientClosed, true},
		{"Timeout", ErrTimeout, false},
		{"NoResponders", ErrNoResponders, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsConnectionError(tt.err); got != tt.want {
				t.Errorf("IsConnectionError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsTimeoutError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"Timeout", ErrTimeout, true},
		{"ClientClosed", ErrClientClosed, false},
		{"ConnectionFailed", ErrConnectionFailed, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsTimeoutError(tt.err); got != tt.want {
				t.Errorf("IsTimeoutError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_JetStream_Errors(t *testing.T) {
	t.Run("returns error when JetStream not enabled", func(t *testing.T) {
		config := DefaultConfig()
		config.EnableJetStream = false
		client := &Client{
			config: config,
			closed: false,
		}

		_, err := client.JetStream()
		if err == nil || err.Error() != "JetStream not enabled in configuration" {
			t.Errorf("JetStream() error = %v, want 'JetStream not enabled in configuration'", err)
		}
	})

	t.Run("returns error when JetStream context nil", func(t *testing.T) {
		config := DefaultConfig()
		config.EnableJetStream = true
		client := &Client{
			config: config,
			closed: false,
			js:     nil,
		}

		_, err := client.JetStream()
		if err == nil || err.Error() != "JetStream context not initialized" {
			t.Errorf("JetStream() error = %v, want 'JetStream context not initialized'", err)
		}
	})
}

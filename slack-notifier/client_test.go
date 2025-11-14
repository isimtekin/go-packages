package slacknotifier

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.WebhookURL != "" {
		t.Errorf("Default webhook URL should be empty")
	}

	if config.DefaultUsername != "Slack Notifier" {
		t.Errorf("Expected default username 'Slack Notifier', got %s", config.DefaultUsername)
	}

	if config.DefaultIconEmoji != ":robot_face:" {
		t.Errorf("Expected default icon emoji ':robot_face:', got %s", config.DefaultIconEmoji)
	}

	if config.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", config.Timeout)
	}

	if config.MaxRetries != 3 {
		t.Errorf("Expected max retries 3, got %d", config.MaxRetries)
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				WebhookURL: "https://hooks.slack.com/services/test",
				Timeout:    30 * time.Second,
				MaxRetries: 3,
				RetryDelay: time.Second,
			},
			wantErr: false,
		},
		{
			name: "empty webhook URL",
			config: &Config{
				WebhookURL: "",
				Timeout:    30 * time.Second,
				MaxRetries: 3,
			},
			wantErr: true,
		},
		{
			name: "zero timeout",
			config: &Config{
				WebhookURL: "https://hooks.slack.com/services/test",
				Timeout:    0,
				MaxRetries: 3,
			},
			wantErr: true,
		},
		{
			name: "negative max retries",
			config: &Config{
				WebhookURL: "https://hooks.slack.com/services/test",
				Timeout:    30 * time.Second,
				MaxRetries: -1,
			},
			wantErr: true,
		},
		{
			name: "negative retry delay",
			config: &Config{
				WebhookURL: "https://hooks.slack.com/services/test",
				Timeout:    30 * time.Second,
				MaxRetries: 3,
				RetryDelay: -1 * time.Second,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewClient(t *testing.T) {
	config := &Config{
		WebhookURL: "https://hooks.slack.com/services/test",
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		RetryDelay: time.Second,
	}

	client, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	if client == nil {
		t.Fatal("Expected client to be non-nil")
	}

	if client.IsClosed() {
		t.Error("New client should not be closed")
	}
}

func TestNewWithOptions(t *testing.T) {
	client, err := NewWithOptions(
		WithWebhookURL("https://hooks.slack.com/services/test"),
		WithChannel("#general"),
		WithUsername("Test Bot"),
		WithIconEmoji(":ghost:"),
		WithTimeout(10*time.Second),
		WithMaxRetries(5),
		WithRetryDelay(2*time.Second),
		WithDebug(true),
	)

	if err != nil {
		t.Fatalf("NewWithOptions() failed: %v", err)
	}

	if client == nil {
		t.Fatal("Expected client to be non-nil")
	}

	if client.config.WebhookURL != "https://hooks.slack.com/services/test" {
		t.Errorf("Expected webhook URL to be set")
	}

	if client.config.DefaultChannel != "#general" {
		t.Errorf("Expected channel '#general', got %s", client.config.DefaultChannel)
	}

	if client.config.DefaultUsername != "Test Bot" {
		t.Errorf("Expected username 'Test Bot', got %s", client.config.DefaultUsername)
	}

	if client.config.Timeout != 10*time.Second {
		t.Errorf("Expected timeout 10s, got %v", client.config.Timeout)
	}

	if client.config.MaxRetries != 5 {
		t.Errorf("Expected max retries 5, got %d", client.config.MaxRetries)
	}
}

func TestClientClose(t *testing.T) {
	client, err := NewWithOptions(
		WithWebhookURL("https://hooks.slack.com/services/test"),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// First close should succeed
	if err := client.Close(); err != nil {
		t.Errorf("First Close() failed: %v", err)
	}

	if !client.IsClosed() {
		t.Error("Client should be closed after Close()")
	}

	// Second close should return ErrAlreadyClosed
	if err := client.Close(); err != ErrAlreadyClosed {
		t.Errorf("Second Close() error = %v, want %v", err, ErrAlreadyClosed)
	}
}

func TestClientSendWithMockServer(t *testing.T) {
	// Create mock Slack server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json")
		}

		var msg Message
		if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
			t.Errorf("Failed to decode message: %v", err)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	client, err := NewWithOptions(
		WithWebhookURL(server.URL),
		WithMaxRetries(0),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	tests := []struct {
		name    string
		message *Message
	}{
		{
			name: "simple text message",
			message: &Message{
				Text: "Test message",
			},
		},
		{
			name: "message with channel",
			message: &Message{
				Text:    "Test message",
				Channel: "#test-channel",
			},
		},
		{
			name: "message with username",
			message: &Message{
				Text:     "Test message",
				Username: "Test Bot",
			},
		},
		{
			name: "message with attachment",
			message: &Message{
				Attachments: []Attachment{
					{
						Fallback: "Test",
						Text:     "Attachment text",
						Color:    ColorGood,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.Send(ctx, tt.message)
			if err != nil {
				t.Errorf("Send() failed: %v", err)
			}
		})
	}
}

func TestClientSendText(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	client, err := NewWithOptions(
		WithWebhookURL(server.URL),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	if err := client.SendText(ctx, "Simple text"); err != nil {
		t.Errorf("SendText() failed: %v", err)
	}
}

func TestClientSendWithColorMethods(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	client, err := NewWithOptions(
		WithWebhookURL(server.URL),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	tests := []struct {
		name   string
		method func(context.Context, string) error
	}{
		{"SendSuccess", client.SendSuccess},
		{"SendWarning", client.SendWarning},
		{"SendError", client.SendError},
		{"SendInfo", client.SendInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.method(ctx, "Test message"); err != nil {
				t.Errorf("%s() failed: %v", tt.name, err)
			}
		})
	}
}

func TestClientRetryLogic(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("error"))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		}
	}))
	defer server.Close()

	client, err := NewWithOptions(
		WithWebhookURL(server.URL),
		WithMaxRetries(3),
		WithRetryDelay(10*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	err = client.SendText(ctx, "Test retry")
	if err != nil {
		t.Errorf("Send with retries failed: %v", err)
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestClientMaxRetriesExceeded(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error"))
	}))
	defer server.Close()

	client, err := NewWithOptions(
		WithWebhookURL(server.URL),
		WithMaxRetries(2),
		WithRetryDelay(10*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	err = client.SendText(ctx, "Test max retries")
	if err == nil {
		t.Error("Expected error when max retries exceeded")
	}
}

func TestClientClosedError(t *testing.T) {
	client, err := NewWithOptions(
		WithWebhookURL("https://hooks.slack.com/services/test"),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Close the client
	client.Close()

	// Try to send a message
	ctx := context.Background()
	err = client.SendText(ctx, "Test")

	if err != ErrClientClosed {
		t.Errorf("Expected ErrClientClosed, got %v", err)
	}
}

func TestClientContextTimeout(t *testing.T) {
	// Create a server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	client, err := NewWithOptions(
		WithWebhookURL(server.URL),
		WithMaxRetries(0),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err = client.SendText(ctx, "Test timeout")
	if err == nil {
		t.Error("Expected timeout error")
	}
}

func TestApplyDefaults(t *testing.T) {
	config := &Config{
		WebhookURL:       "https://hooks.slack.com/services/test",
		DefaultChannel:   "#default-channel",
		DefaultUsername:  "Default Bot",
		DefaultIconEmoji: ":robot_face:",
		DefaultIconURL:   "https://example.com/icon.png",
		ThreadTS:         "1234567890.123456",
		Timeout:          30 * time.Second,
	}

	client, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	message := &Message{
		Text: "Test message",
	}

	client.applyDefaults(message)

	if message.Channel != "#default-channel" {
		t.Errorf("Expected channel '#default-channel', got %s", message.Channel)
	}

	if message.Username != "Default Bot" {
		t.Errorf("Expected username 'Default Bot', got %s", message.Username)
	}

	if message.IconEmoji != ":robot_face:" {
		t.Errorf("Expected icon emoji ':robot_face:', got %s", message.IconEmoji)
	}

	if message.IconURL != "https://example.com/icon.png" {
		t.Errorf("Expected icon URL, got %s", message.IconURL)
	}

	if message.ThreadTS != "1234567890.123456" {
		t.Errorf("Expected thread TS, got %s", message.ThreadTS)
	}
}

func TestInvalidResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid"))
	}))
	defer server.Close()

	client, err := NewWithOptions(
		WithWebhookURL(server.URL),
		WithMaxRetries(0),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	err = client.SendText(ctx, "Test")

	if err == nil {
		t.Error("Expected error for invalid response")
	}

	// The error should contain ErrInvalidResponse (may be wrapped)
	if err != nil && err.Error() == "" {
		t.Error("Error message should not be empty")
	}
}

func TestErrorHelpers(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		isConn   bool
		isTimeout bool
	}{
		{
			name:     "connection error",
			err:      ErrConnectionFailed,
			isConn:   true,
			isTimeout: false,
		},
		{
			name:     "client closed error",
			err:      ErrClientClosed,
			isConn:   true,
			isTimeout: false,
		},
		{
			name:     "timeout error",
			err:      ErrTimeout,
			isConn:   false,
			isTimeout: true,
		},
		{
			name:     "invalid response",
			err:      ErrInvalidResponse,
			isConn:   false,
			isTimeout: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if IsConnectionError(tt.err) != tt.isConn {
				t.Errorf("IsConnectionError() = %v, want %v", IsConnectionError(tt.err), tt.isConn)
			}
			if IsTimeoutError(tt.err) != tt.isTimeout {
				t.Errorf("IsTimeoutError() = %v, want %v", IsTimeoutError(tt.err), tt.isTimeout)
			}
		})
	}
}

func TestSendWithAttachments(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var msg Message
		json.NewDecoder(r.Body).Decode(&msg)

		if len(msg.Attachments) == 0 {
			t.Error("Expected attachments in message")
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	client, err := NewWithOptions(
		WithWebhookURL(server.URL),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	attachments := []Attachment{
		{
			Fallback: "Test attachment",
			Text:     "Attachment text",
			Color:    ColorGood,
			Fields: []AttachmentField{
				{Title: "Field1", Value: "Value1", Short: true},
			},
		},
	}

	err = client.SendWithAttachments(ctx, "Main text", attachments)
	if err != nil {
		t.Errorf("SendWithAttachments() failed: %v", err)
	}
}

func TestSendWithBlocks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var msg Message
		json.NewDecoder(r.Body).Decode(&msg)

		if len(msg.Blocks) == 0 {
			t.Error("Expected blocks in message")
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	client, err := NewWithOptions(
		WithWebhookURL(server.URL),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	blocks := []Block{
		NewHeaderBlock("Test Header"),
		NewSectionBlock("Section text"),
		NewDividerBlock(),
	}

	err = client.SendWithBlocks(ctx, blocks)
	if err != nil {
		t.Errorf("SendWithBlocks() failed: %v", err)
	}
}

func ExampleClient_SendText() {
	client, err := NewWithOptions(
		WithWebhookURL("https://hooks.slack.com/services/YOUR/WEBHOOK/URL"),
	)
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		return
	}
	defer client.Close()

	ctx := context.Background()
	if err := client.SendText(ctx, "Hello, Slack!"); err != nil {
		fmt.Printf("Failed to send message: %v\n", err)
	}
}

func ExampleClient_SendSuccess() {
	client, err := NewWithOptions(
		WithWebhookURL("https://hooks.slack.com/services/YOUR/WEBHOOK/URL"),
	)
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		return
	}
	defer client.Close()

	ctx := context.Background()
	if err := client.SendSuccess(ctx, "Deployment completed successfully!"); err != nil {
		fmt.Printf("Failed to send message: %v\n", err)
	}
}

func ExampleNewMessage() {
	message := NewMessage("Hello from message builder").
		Channel("#general").
		Username("Custom Bot").
		IconEmoji(":tada:").
		AddAttachment(Attachment{
			Fallback: "Important notification",
			Text:     "This is an important update",
			Color:    ColorWarning,
		}).
		Build()

	fmt.Printf("Message: %s\n", message.Text)
}

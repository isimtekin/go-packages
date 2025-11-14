package slacknotifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// Client represents the Slack notifier client
type Client struct {
	config     *Config
	httpClient *http.Client

	mu     sync.RWMutex
	closed bool
}

// New creates a new Slack notifier client
func New(config *Config) (*Client, error) {
	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	client := &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		closed: false,
	}

	return client, nil
}

// NewWithOptions creates a new client with functional options
func NewWithOptions(opts ...Option) (*Client, error) {
	config := DefaultConfig()

	for _, opt := range opts {
		opt(config)
	}

	return New(config)
}

// Send sends a message to Slack
func (c *Client) Send(ctx context.Context, message *Message) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return ErrClientClosed
	}

	// Apply default values from config
	c.applyDefaults(message)

	// Marshal message to JSON
	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Send with retries
	return c.sendWithRetry(ctx, payload)
}

// SendText sends a simple text message
func (c *Client) SendText(ctx context.Context, text string) error {
	return c.Send(ctx, &Message{Text: text})
}

// SendSuccess sends a success message (green)
func (c *Client) SendSuccess(ctx context.Context, text string) error {
	return c.Send(ctx, NewSuccessMessage(text).Build())
}

// SendWarning sends a warning message (yellow)
func (c *Client) SendWarning(ctx context.Context, text string) error {
	return c.Send(ctx, NewWarningMessage(text).Build())
}

// SendError sends an error message (red)
func (c *Client) SendError(ctx context.Context, text string) error {
	return c.Send(ctx, NewErrorMessage(text).Build())
}

// SendInfo sends an info message (blue)
func (c *Client) SendInfo(ctx context.Context, text string) error {
	return c.Send(ctx, NewInfoMessage(text).Build())
}

// SendWithAttachments sends a message with attachments
func (c *Client) SendWithAttachments(ctx context.Context, text string, attachments []Attachment) error {
	return c.Send(ctx, &Message{
		Text:        text,
		Attachments: attachments,
	})
}

// SendWithBlocks sends a message with Block Kit blocks
func (c *Client) SendWithBlocks(ctx context.Context, blocks []Block) error {
	return c.Send(ctx, &Message{
		Blocks: blocks,
	})
}

// applyDefaults applies default values from config to message
func (c *Client) applyDefaults(message *Message) {
	if message.Channel == "" && c.config.DefaultChannel != "" {
		message.Channel = c.config.DefaultChannel
	}

	if message.Username == "" && c.config.DefaultUsername != "" {
		message.Username = c.config.DefaultUsername
	}

	if message.IconEmoji == "" && c.config.DefaultIconEmoji != "" {
		message.IconEmoji = c.config.DefaultIconEmoji
	}

	if message.IconURL == "" && c.config.DefaultIconURL != "" {
		message.IconURL = c.config.DefaultIconURL
	}

	if message.ThreadTS == "" && c.config.ThreadTS != "" {
		message.ThreadTS = c.config.ThreadTS
	}
}

// sendWithRetry sends the payload with retry logic
func (c *Client) sendWithRetry(ctx context.Context, payload []byte) error {
	var lastErr error

	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry
			select {
			case <-ctx.Done():
				return fmt.Errorf("%w: %v", ErrTimeout, ctx.Err())
			case <-time.After(c.config.RetryDelay * time.Duration(attempt)):
			}
		}

		if c.config.EnableDebug {
			fmt.Printf("[slack-notifier] Attempt %d/%d: Sending message\n", attempt+1, c.config.MaxRetries+1)
		}

		err := c.sendRequest(ctx, payload)
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't retry on context cancellation
		if ctx.Err() != nil {
			return fmt.Errorf("%w: %v", ErrTimeout, ctx.Err())
		}
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

// sendRequest sends a single HTTP request to Slack
func (c *Client) sendRequest(ctx context.Context, payload []byte) error {
	req, err := http.NewRequestWithContext(ctx, "POST", c.config.WebhookURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// Slack webhook returns "ok" on success
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: status %d, body: %s", ErrInvalidResponse, resp.StatusCode, string(body))
	}

	if string(body) != "ok" {
		return fmt.Errorf("%w: %s", ErrInvalidResponse, string(body))
	}

	if c.config.EnableDebug {
		fmt.Printf("[slack-notifier] Message sent successfully\n")
	}

	return nil
}

// Close closes the client and releases resources
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return ErrAlreadyClosed
	}

	c.httpClient.CloseIdleConnections()
	c.closed = true

	return nil
}

// IsClosed returns true if the client is closed
func (c *Client) IsClosed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.closed
}

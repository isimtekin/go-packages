package mailsender

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithProvider(t *testing.T) {
	config := DefaultConfig()
	opt := WithProvider("test-provider")
	opt(config)
	assert.Equal(t, Provider("test-provider"), config.Provider)
}

func TestWithAPIKey(t *testing.T) {
	config := DefaultConfig()
	opt := WithAPIKey("test-api-key")
	opt(config)
	assert.Equal(t, "test-api-key", config.APIKey)
}

func TestWithDefaultFrom(t *testing.T) {
	config := DefaultConfig()
	opt := WithDefaultFrom("sender@example.com")
	opt(config)
	assert.Equal(t, "sender@example.com", config.DefaultFrom)
}

func TestWithDefaultFromName(t *testing.T) {
	config := DefaultConfig()
	opt := WithDefaultFromName("Test Sender")
	opt(config)
	assert.Equal(t, "Test Sender", config.DefaultFromName)
}

func TestOptions_Combined(t *testing.T) {
	config := DefaultConfig()

	opts := []Option{
		WithProvider(ProviderSendGrid),
		WithAPIKey("test-api-key"),
		WithDefaultFrom("sender@example.com"),
		WithDefaultFromName("Test Sender"),
	}

	for _, opt := range opts {
		opt(config)
	}

	assert.Equal(t, ProviderSendGrid, config.Provider)
	assert.Equal(t, "test-api-key", config.APIKey)
	assert.Equal(t, "sender@example.com", config.DefaultFrom)
	assert.Equal(t, "Test Sender", config.DefaultFromName)
}

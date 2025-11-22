package mailsender

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFromEnv(t *testing.T) {
	// Save original env vars
	originalAPIKey := os.Getenv("TEST_API_KEY")
	originalProvider := os.Getenv("TEST_PROVIDER")
	originalFrom := os.Getenv("TEST_DEFAULT_FROM")
	originalFromName := os.Getenv("TEST_DEFAULT_FROM_NAME")

	// Restore after test
	defer func() {
		os.Setenv("TEST_API_KEY", originalAPIKey)
		os.Setenv("TEST_PROVIDER", originalProvider)
		os.Setenv("TEST_DEFAULT_FROM", originalFrom)
		os.Setenv("TEST_DEFAULT_FROM_NAME", originalFromName)
	}()

	tests := []struct {
		name        string
		prefix      string
		envVars     map[string]string
		wantErr     bool
		checkSender func(t *testing.T, sender EmailSender)
	}{
		{
			name:   "valid sendgrid config",
			prefix: "TEST_",
			envVars: map[string]string{
				"TEST_PROVIDER":           "sendgrid",
				"TEST_API_KEY":            "test-api-key",
				"TEST_DEFAULT_FROM":       "sender@example.com",
				"TEST_DEFAULT_FROM_NAME":  "Test Sender",
			},
			wantErr: false,
			checkSender: func(t *testing.T, sender EmailSender) {
				assert.NotNil(t, sender)
				sgSender, ok := sender.(*SendGridSender)
				assert.True(t, ok)
				assert.Equal(t, "sender@example.com", sgSender.defaultFrom)
				assert.Equal(t, "Test Sender", sgSender.defaultFromName)
			},
		},
		{
			name:   "sendgrid without defaults",
			prefix: "TEST_",
			envVars: map[string]string{
				"TEST_PROVIDER": "sendgrid",
				"TEST_API_KEY":  "test-api-key",
			},
			wantErr: false,
			checkSender: func(t *testing.T, sender EmailSender) {
				assert.NotNil(t, sender)
			},
		},
		{
			name:   "missing api key",
			prefix: "TEST_",
			envVars: map[string]string{
				"TEST_PROVIDER": "sendgrid",
			},
			wantErr: true,
		},
		{
			name:   "invalid provider",
			prefix: "TEST_",
			envVars: map[string]string{
				"TEST_PROVIDER": "invalid-provider",
				"TEST_API_KEY":  "test-api-key",
			},
			wantErr: true,
		},
		{
			name:   "default provider (sendgrid)",
			prefix: "TEST_",
			envVars: map[string]string{
				"TEST_API_KEY": "test-api-key",
			},
			wantErr: false,
			checkSender: func(t *testing.T, sender EmailSender) {
				assert.NotNil(t, sender)
				_, ok := sender.(*SendGridSender)
				assert.True(t, ok)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear env vars
			os.Unsetenv("TEST_PROVIDER")
			os.Unsetenv("TEST_API_KEY")
			os.Unsetenv("TEST_DEFAULT_FROM")
			os.Unsetenv("TEST_DEFAULT_FROM_NAME")

			// Set env vars for this test
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			sender, err := NewFromEnv(tt.prefix)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, sender)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, sender)
				if tt.checkSender != nil {
					tt.checkSender(t, sender)
				}
				if sender != nil {
					sender.Close()
				}
			}
		})
	}
}

func TestNewSendGridFromEnv(t *testing.T) {
	// Save original env vars
	originalAPIKey := os.Getenv("SENDGRID_API_KEY")
	originalFrom := os.Getenv("SENDGRID_DEFAULT_FROM")
	originalFromName := os.Getenv("SENDGRID_DEFAULT_FROM_NAME")

	// Restore after test
	defer func() {
		os.Setenv("SENDGRID_API_KEY", originalAPIKey)
		os.Setenv("SENDGRID_DEFAULT_FROM", originalFrom)
		os.Setenv("SENDGRID_DEFAULT_FROM_NAME", originalFromName)
	}()

	tests := []struct {
		name    string
		envVars map[string]string
		wantErr bool
	}{
		{
			name: "valid config",
			envVars: map[string]string{
				"SENDGRID_API_KEY":            "test-api-key",
				"SENDGRID_DEFAULT_FROM":       "sender@example.com",
				"SENDGRID_DEFAULT_FROM_NAME":  "Test Sender",
			},
			wantErr: false,
		},
		{
			name: "minimal config",
			envVars: map[string]string{
				"SENDGRID_API_KEY": "test-api-key",
			},
			wantErr: false,
		},
		{
			name: "missing api key",
			envVars: map[string]string{
				"SENDGRID_DEFAULT_FROM": "sender@example.com",
			},
			wantErr: true,
		},
		{
			name: "empty config - all env vars empty",
			envVars: map[string]string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear env vars
			os.Unsetenv("SENDGRID_API_KEY")
			os.Unsetenv("SENDGRID_DEFAULT_FROM")
			os.Unsetenv("SENDGRID_DEFAULT_FROM_NAME")

			// Set env vars for this test
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			sender, err := NewSendGridFromEnv()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, sender)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, sender)
				assert.NotNil(t, sender.client)
				sender.Close()
			}
		})
	}
}

func TestNewFromEnv_InvalidProvider(t *testing.T) {
	// Clear env vars
	os.Unsetenv("TEST_PROVIDER")
	os.Unsetenv("TEST_API_KEY")

	// Set invalid provider
	os.Setenv("TEST_PROVIDER", "invalid-provider")
	os.Setenv("TEST_API_KEY", "test-key")

	sender, err := NewFromEnv("TEST_")
	assert.Error(t, err)
	assert.Nil(t, sender)
}

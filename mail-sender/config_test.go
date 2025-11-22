package mailsender

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr error
	}{
		{
			name: "valid sendgrid config",
			config: &Config{
				Provider: ProviderSendGrid,
				APIKey:   "test-api-key",
			},
			wantErr: nil,
		},
		{
			name: "valid config with defaults",
			config: &Config{
				Provider:        ProviderSendGrid,
				APIKey:          "test-api-key",
				DefaultFrom:     "sender@example.com",
				DefaultFromName: "Test Sender",
			},
			wantErr: nil,
		},
		{
			name: "missing provider",
			config: &Config{
				APIKey: "test-api-key",
			},
			wantErr: ErrInvalidProvider,
		},
		{
			name: "invalid provider",
			config: &Config{
				Provider: "invalid",
				APIKey:   "test-api-key",
			},
			wantErr: ErrInvalidProvider,
		},
		{
			name: "missing api key",
			config: &Config{
				Provider: ProviderSendGrid,
			},
			wantErr: ErrMissingAPIKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	assert.NotNil(t, config)
	assert.Equal(t, ProviderSendGrid, config.Provider)
	assert.Empty(t, config.APIKey)
	assert.Empty(t, config.DefaultFrom)
	assert.Empty(t, config.DefaultFromName)
}

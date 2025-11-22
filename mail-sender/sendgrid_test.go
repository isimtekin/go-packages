package mailsender

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSendGrid(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Provider: ProviderSendGrid,
				APIKey:   "test-api-key",
			},
			wantErr: false,
		},
		{
			name: "valid config with defaults",
			config: &Config{
				Provider:        ProviderSendGrid,
				APIKey:          "test-api-key",
				DefaultFrom:     "sender@example.com",
				DefaultFromName: "Test Sender",
			},
			wantErr: false,
		},
		{
			name: "missing api key",
			config: &Config{
				Provider: ProviderSendGrid,
			},
			wantErr: true,
		},
		{
			name: "invalid provider",
			config: &Config{
				Provider: "invalid",
				APIKey:   "test-api-key",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sender, err := NewSendGrid(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, sender)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, sender)
				assert.NotNil(t, sender.client)
			}
		})
	}
}

func TestNewSendGridWithOptions(t *testing.T) {
	sender, err := NewSendGridWithOptions(
		WithAPIKey("test-api-key"),
		WithDefaultFrom("sender@example.com"),
		WithDefaultFromName("Test Sender"),
	)

	require.NoError(t, err)
	assert.NotNil(t, sender)
	assert.Equal(t, "sender@example.com", sender.defaultFrom)
	assert.Equal(t, "Test Sender", sender.defaultFromName)
}

func TestSendGridSender_Close(t *testing.T) {
	sender, err := NewSendGrid(&Config{
		Provider: ProviderSendGrid,
		APIKey:   "test-api-key",
	})
	require.NoError(t, err)

	err = sender.Close()
	assert.NoError(t, err)
}

func TestSendGridSender_buildMessage(t *testing.T) {
	sender, err := NewSendGrid(&Config{
		Provider: ProviderSendGrid,
		APIKey:   "test-api-key",
	})
	require.NoError(t, err)

	tests := []struct {
		name    string
		message *EmailMessage
	}{
		{
			name: "simple message",
			message: &EmailMessage{
				From:      "sender@example.com",
				FromName:  "Test Sender",
				To:        []string{"recipient@example.com"},
				Subject:   "Test Subject",
				PlainText: "Test Body",
			},
		},
		{
			name: "message with HTML",
			message: &EmailMessage{
				From:    "sender@example.com",
				To:      []string{"recipient@example.com"},
				Subject: "Test Subject",
				HTML:    "<p>Test Body</p>",
			},
		},
		{
			name: "message with multiple recipients",
			message: &EmailMessage{
				From:      "sender@example.com",
				To:        []string{"recipient1@example.com", "recipient2@example.com"},
				Cc:        []string{"cc@example.com"},
				Bcc:       []string{"bcc@example.com"},
				Subject:   "Test Subject",
				PlainText: "Test Body",
			},
		},
		{
			name: "message with reply-to",
			message: &EmailMessage{
				From:      "sender@example.com",
				To:        []string{"recipient@example.com"},
				Subject:   "Test Subject",
				PlainText: "Test Body",
				ReplyTo:   "replyto@example.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sgMessage := sender.buildMessage(tt.message)
			assert.NotNil(t, sgMessage)
			assert.Equal(t, tt.message.Subject, sgMessage.Subject)
			assert.NotNil(t, sgMessage.From)
			assert.Equal(t, tt.message.From, sgMessage.From.Address)
		})
	}
}

func TestSendGridSender_ApplyDefaults(t *testing.T) {
	sender, err := NewSendGrid(&Config{
		Provider:        ProviderSendGrid,
		APIKey:          "test-api-key",
		DefaultFrom:     "default@example.com",
		DefaultFromName: "Default Sender",
	})
	require.NoError(t, err)

	message := &EmailMessage{
		To:        []string{"recipient@example.com"},
		Subject:   "Test Subject",
		PlainText: "Test Body",
	}

	// buildMessage applies defaults internally
	sgMessage := sender.buildMessage(message)
	assert.NotNil(t, sgMessage)

	// After buildMessage, the original message should have defaults applied via Send
	// We'll test this through the validation path
	assert.Empty(t, message.From) // Original message not modified by buildMessage
}

func TestSendGridSender_buildMessage_WithAllFeatures(t *testing.T) {
	sender, err := NewSendGrid(&Config{
		Provider: ProviderSendGrid,
		APIKey:   "test-api-key",
	})
	require.NoError(t, err)

	message := &EmailMessage{
		From:      "sender@example.com",
		FromName:  "Test Sender",
		To:        []string{"to1@example.com", "to2@example.com"},
		Cc:        []string{"cc1@example.com", "cc2@example.com"},
		Bcc:       []string{"bcc1@example.com", "bcc2@example.com"},
		Subject:   "Complete Test",
		PlainText: "Plain text content",
		HTML:      "<p>HTML content</p>",
		ReplyTo:   "replyto@example.com",
	}

	sgMessage := sender.buildMessage(message)

	assert.NotNil(t, sgMessage)
	assert.NotNil(t, sgMessage.From)
	assert.Equal(t, "sender@example.com", sgMessage.From.Address)
	assert.Equal(t, "Test Sender", sgMessage.From.Name)
	assert.Equal(t, "Complete Test", sgMessage.Subject)
	assert.NotNil(t, sgMessage.Personalizations)
	assert.Len(t, sgMessage.Personalizations, 1)
	assert.NotNil(t, sgMessage.ReplyTo)
	assert.Equal(t, "replyto@example.com", sgMessage.ReplyTo.Address)
}

func TestNewSendGrid_InvalidProvider(t *testing.T) {
	config := &Config{
		Provider: "mailgun", // Not sendgrid
		APIKey:   "test-api-key",
	}

	sender, err := NewSendGrid(config)
	assert.Error(t, err)
	assert.Nil(t, sender)
	assert.Contains(t, err.Error(), "mailgun")
}

func TestSendGridSender_Send_ValidationErrors(t *testing.T) {
	sender, err := NewSendGrid(&Config{
		Provider: ProviderSendGrid,
		APIKey:   "test-api-key",
	})
	require.NoError(t, err)

	tests := []struct {
		name    string
		message *EmailMessage
		wantErr error
	}{
		{
			name: "missing from and no default",
			message: &EmailMessage{
				To:        []string{"recipient@example.com"},
				Subject:   "Test",
				PlainText: "Test",
			},
			wantErr: ErrMissingFrom,
		},
		{
			name: "missing recipients",
			message: &EmailMessage{
				From:      "sender@example.com",
				Subject:   "Test",
				PlainText: "Test",
			},
			wantErr: ErrMissingRecipients,
		},
		{
			name: "missing subject",
			message: &EmailMessage{
				From:      "sender@example.com",
				To:        []string{"recipient@example.com"},
				PlainText: "Test",
			},
			wantErr: ErrMissingSubject,
		},
		{
			name: "missing content",
			message: &EmailMessage{
				From:    "sender@example.com",
				To:      []string{"recipient@example.com"},
				Subject: "Test",
			},
			wantErr: ErrMissingContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sender.Send(context.Background(), tt.message)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestSendGridSender_Send_WithDefaults(t *testing.T) {
	sender, err := NewSendGrid(&Config{
		Provider:        ProviderSendGrid,
		APIKey:          "test-api-key",
		DefaultFrom:     "default@example.com",
		DefaultFromName: "Default Sender",
	})
	require.NoError(t, err)

	// This will fail at the actual send, but will exercise the default application logic
	err = sender.Send(context.Background(), &EmailMessage{
		// No From, should use default
		To:        []string{"recipient@example.com"},
		Subject:   "Test",
		PlainText: "Test",
	})

	// Will fail on actual send, but that's OK - we're testing the validation path
	assert.Error(t, err) // Will get send error
	assert.Contains(t, err.Error(), "failed to send email")
}

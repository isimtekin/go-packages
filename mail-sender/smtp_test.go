package mailsender

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSMTPConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *SMTPConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: &SMTPConfig{
				Host: "smtp.example.com",
				Port: 587,
			},
			wantErr: false,
		},
		{
			name: "missing host",
			config: &SMTPConfig{
				Port: 587,
			},
			wantErr: true,
			errMsg:  "host is required",
		},
		{
			name: "missing port",
			config: &SMTPConfig{
				Host: "smtp.example.com",
				Port: 0,
			},
			wantErr: true,
			errMsg:  "port must be positive",
		},
		{
			name: "negative port",
			config: &SMTPConfig{
				Host: "smtp.example.com",
				Port: -1,
			},
			wantErr: true,
			errMsg:  "port must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDefaultSMTPConfig(t *testing.T) {
	config := DefaultSMTPConfig()

	assert.Equal(t, 587, config.Port)
	assert.True(t, config.UseTLS)
	assert.Equal(t, 30*time.Second, config.Timeout)
	assert.Equal(t, "localhost", config.LocalName)
}

func TestNewSMTP(t *testing.T) {
	tests := []struct {
		name    string
		config  *SMTPConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: &SMTPConfig{
				Host: "smtp.example.com",
				Port: 587,
			},
			wantErr: false,
		},
		{
			name:    "nil config uses defaults but fails validation",
			config:  nil,
			wantErr: true, // Because DefaultSMTPConfig has no host
		},
		{
			name: "invalid config",
			config: &SMTPConfig{
				Host: "",
				Port: 587,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sender, err := NewSMTP(tt.config)
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, sender)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, sender)
			}
		})
	}
}

func TestNewSMTPWithOptions(t *testing.T) {
	sender, err := NewSMTPWithOptions(
		WithSMTPHost("smtp.example.com"),
		WithSMTPPort(465),
		WithSMTPAuth("user", "pass"),
		WithSMTPTLS(true),
		WithSMTPInsecureSkipVerify(true),
		WithSMTPTimeout(60*time.Second),
		WithSMTPDefaultFrom("default@example.com", "Default Sender"),
		WithSMTPLocalName("myhost.local"),
	)

	require.NoError(t, err)
	assert.NotNil(t, sender)
	assert.Equal(t, "smtp.example.com", sender.config.Host)
	assert.Equal(t, 465, sender.config.Port)
	assert.Equal(t, "user", sender.config.Username)
	assert.Equal(t, "pass", sender.config.Password)
	assert.True(t, sender.config.UseTLS)
	assert.True(t, sender.config.InsecureSkipVerify)
	assert.Equal(t, 60*time.Second, sender.config.Timeout)
	assert.Equal(t, "default@example.com", sender.config.DefaultFrom)
	assert.Equal(t, "Default Sender", sender.config.DefaultFromName)
	assert.Equal(t, "myhost.local", sender.config.LocalName)
}

func TestSMTPSender_Close(t *testing.T) {
	sender, err := NewSMTPWithOptions(
		WithSMTPHost("smtp.example.com"),
		WithSMTPPort(587),
	)
	require.NoError(t, err)

	err = sender.Close()
	require.NoError(t, err)
}

func TestSMTPSender_buildEmail_PlainText(t *testing.T) {
	sender, err := NewSMTPWithOptions(
		WithSMTPHost("smtp.example.com"),
		WithSMTPPort(587),
	)
	require.NoError(t, err)

	message := &EmailMessage{
		From:      "sender@example.com",
		FromName:  "Sender Name",
		To:        []string{"recipient@example.com"},
		Subject:   "Test Subject",
		PlainText: "Hello, this is a test email.",
	}

	email := sender.buildEmail(message)
	emailStr := string(email)

	assert.Contains(t, emailStr, "From: Sender Name <sender@example.com>")
	assert.Contains(t, emailStr, "To: recipient@example.com")
	assert.Contains(t, emailStr, "Subject: Test Subject")
	assert.Contains(t, emailStr, "Content-Type: text/plain")
	assert.Contains(t, emailStr, "Hello, this is a test email.")
}

func TestSMTPSender_buildEmail_HTML(t *testing.T) {
	sender, err := NewSMTPWithOptions(
		WithSMTPHost("smtp.example.com"),
		WithSMTPPort(587),
	)
	require.NoError(t, err)

	message := &EmailMessage{
		From:    "sender@example.com",
		To:      []string{"recipient@example.com"},
		Subject: "Test Subject",
		HTML:    "<html><body><h1>Hello</h1></body></html>",
	}

	email := sender.buildEmail(message)
	emailStr := string(email)

	assert.Contains(t, emailStr, "From: sender@example.com")
	assert.Contains(t, emailStr, "Content-Type: text/html")
	assert.Contains(t, emailStr, "<html><body><h1>Hello</h1></body></html>")
}

func TestSMTPSender_buildEmail_Multipart(t *testing.T) {
	sender, err := NewSMTPWithOptions(
		WithSMTPHost("smtp.example.com"),
		WithSMTPPort(587),
	)
	require.NoError(t, err)

	message := &EmailMessage{
		From:      "sender@example.com",
		To:        []string{"recipient@example.com"},
		Subject:   "Test Subject",
		PlainText: "Plain text version",
		HTML:      "<html><body><h1>HTML version</h1></body></html>",
	}

	email := sender.buildEmail(message)
	emailStr := string(email)

	assert.Contains(t, emailStr, "Content-Type: multipart/alternative")
	assert.Contains(t, emailStr, "text/plain")
	assert.Contains(t, emailStr, "text/html")
	assert.Contains(t, emailStr, "Plain text version")
	assert.Contains(t, emailStr, "<html><body><h1>HTML version</h1></body></html>")
}

func TestSMTPSender_buildEmail_WithCc(t *testing.T) {
	sender, err := NewSMTPWithOptions(
		WithSMTPHost("smtp.example.com"),
		WithSMTPPort(587),
	)
	require.NoError(t, err)

	message := &EmailMessage{
		From:      "sender@example.com",
		To:        []string{"recipient@example.com"},
		Cc:        []string{"cc1@example.com", "cc2@example.com"},
		Subject:   "Test Subject",
		PlainText: "Test",
	}

	email := sender.buildEmail(message)
	emailStr := string(email)

	assert.Contains(t, emailStr, "Cc: cc1@example.com, cc2@example.com")
}

func TestSMTPSender_buildEmail_WithReplyTo(t *testing.T) {
	sender, err := NewSMTPWithOptions(
		WithSMTPHost("smtp.example.com"),
		WithSMTPPort(587),
	)
	require.NoError(t, err)

	message := &EmailMessage{
		From:      "sender@example.com",
		To:        []string{"recipient@example.com"},
		ReplyTo:   "replyto@example.com",
		Subject:   "Test Subject",
		PlainText: "Test",
	}

	email := sender.buildEmail(message)
	emailStr := string(email)

	assert.Contains(t, emailStr, "Reply-To: replyto@example.com")
}

func TestSMTPSender_getAllRecipients(t *testing.T) {
	sender, err := NewSMTPWithOptions(
		WithSMTPHost("smtp.example.com"),
		WithSMTPPort(587),
	)
	require.NoError(t, err)

	message := &EmailMessage{
		To:  []string{"to1@example.com", "to2@example.com"},
		Cc:  []string{"cc@example.com"},
		Bcc: []string{"bcc@example.com"},
	}

	recipients := sender.getAllRecipients(message)

	assert.Len(t, recipients, 4)
	assert.Contains(t, recipients, "to1@example.com")
	assert.Contains(t, recipients, "to2@example.com")
	assert.Contains(t, recipients, "cc@example.com")
	assert.Contains(t, recipients, "bcc@example.com")
}

func TestSMTPSender_Send_ValidationError(t *testing.T) {
	sender, err := NewSMTPWithOptions(
		WithSMTPHost("smtp.example.com"),
		WithSMTPPort(587),
	)
	require.NoError(t, err)

	// Missing required fields
	message := &EmailMessage{
		Subject: "Test",
	}

	err = sender.Send(context.Background(), message)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid message")
}

func TestSMTPSender_Send_AppliesDefaults(t *testing.T) {
	sender, err := NewSMTPWithOptions(
		WithSMTPHost("smtp.example.com"),
		WithSMTPPort(587),
		WithSMTPDefaultFrom("default@example.com", "Default Name"),
	)
	require.NoError(t, err)

	// Message without From
	message := &EmailMessage{
		To:        []string{"to@example.com"},
		Subject:   "Test",
		PlainText: "Hello",
	}

	// This will fail at the connection stage, but we can verify defaults were applied
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_ = sender.Send(ctx, message)

	// Verify defaults were applied to the message
	assert.Equal(t, "default@example.com", message.From)
	assert.Equal(t, "Default Name", message.FromName)
}

func TestSMTPSender_Send_ContextCancellation(t *testing.T) {
	sender, err := NewSMTPWithOptions(
		WithSMTPHost("smtp.example.com"),
		WithSMTPPort(587),
		WithSMTPTimeout(5*time.Second),
	)
	require.NoError(t, err)

	// Create already cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	message := &EmailMessage{
		From:      "from@example.com",
		To:        []string{"to@example.com"},
		Subject:   "Test",
		PlainText: "Hello",
	}

	err = sender.Send(ctx, message)
	require.Error(t, err)
	// The error is wrapped, so check the error message contains context canceled
	assert.Contains(t, err.Error(), "context canceled")
}

func TestSMTPSender_buildEmail_NoFromName(t *testing.T) {
	sender, err := NewSMTPWithOptions(
		WithSMTPHost("smtp.example.com"),
		WithSMTPPort(587),
	)
	require.NoError(t, err)

	message := &EmailMessage{
		From:      "sender@example.com",
		To:        []string{"recipient@example.com"},
		Subject:   "Test Subject",
		PlainText: "Hello",
	}

	email := sender.buildEmail(message)
	emailStr := string(email)

	// Should not have the name format
	assert.Contains(t, emailStr, "From: sender@example.com\r\n")
	assert.NotContains(t, emailStr, "<sender@example.com>")
}

func TestSMTPSender_buildEmail_MultipleRecipients(t *testing.T) {
	sender, err := NewSMTPWithOptions(
		WithSMTPHost("smtp.example.com"),
		WithSMTPPort(587),
	)
	require.NoError(t, err)

	message := &EmailMessage{
		From:      "sender@example.com",
		To:        []string{"r1@example.com", "r2@example.com", "r3@example.com"},
		Subject:   "Test Subject",
		PlainText: "Hello",
	}

	email := sender.buildEmail(message)
	emailStr := string(email)

	assert.Contains(t, emailStr, "To: r1@example.com, r2@example.com, r3@example.com")
}

func TestSMTPOptions(t *testing.T) {
	config := DefaultSMTPConfig()

	WithSMTPHost("custom.host.com")(config)
	assert.Equal(t, "custom.host.com", config.Host)

	WithSMTPPort(2525)(config)
	assert.Equal(t, 2525, config.Port)

	WithSMTPAuth("myuser", "mypass")(config)
	assert.Equal(t, "myuser", config.Username)
	assert.Equal(t, "mypass", config.Password)

	WithSMTPTLS(false)(config)
	assert.False(t, config.UseTLS)

	WithSMTPInsecureSkipVerify(true)(config)
	assert.True(t, config.InsecureSkipVerify)

	WithSMTPTimeout(45 * time.Second)(config)
	assert.Equal(t, 45*time.Second, config.Timeout)

	WithSMTPDefaultFrom("from@test.com", "From Test")(config)
	assert.Equal(t, "from@test.com", config.DefaultFrom)
	assert.Equal(t, "From Test", config.DefaultFromName)

	WithSMTPLocalName("mail.myhost.com")(config)
	assert.Equal(t, "mail.myhost.com", config.LocalName)
}

func TestSMTPSender_buildEmail_MIMEVersion(t *testing.T) {
	sender, err := NewSMTPWithOptions(
		WithSMTPHost("smtp.example.com"),
		WithSMTPPort(587),
	)
	require.NoError(t, err)

	message := &EmailMessage{
		From:      "sender@example.com",
		To:        []string{"recipient@example.com"},
		Subject:   "Test Subject",
		PlainText: "Hello",
	}

	email := sender.buildEmail(message)
	emailStr := string(email)

	assert.Contains(t, emailStr, "MIME-Version: 1.0")
}

func TestSMTPSender_buildEmail_CharsetUTF8(t *testing.T) {
	sender, err := NewSMTPWithOptions(
		WithSMTPHost("smtp.example.com"),
		WithSMTPPort(587),
	)
	require.NoError(t, err)

	message := &EmailMessage{
		From:      "sender@example.com",
		To:        []string{"recipient@example.com"},
		Subject:   "Test Subject",
		PlainText: "Hello",
	}

	email := sender.buildEmail(message)
	emailStr := string(email)

	assert.True(t, strings.Contains(emailStr, "charset=\"utf-8\"") || strings.Contains(emailStr, "charset=utf-8"))
}

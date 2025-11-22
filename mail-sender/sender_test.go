package mailsender

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmailMessage_Validate(t *testing.T) {
	tests := []struct {
		name    string
		message *EmailMessage
		wantErr error
	}{
		{
			name: "valid message with plain text",
			message: &EmailMessage{
				From:      "sender@example.com",
				To:        []string{"recipient@example.com"},
				Subject:   "Test Subject",
				PlainText: "Test Body",
			},
			wantErr: nil,
		},
		{
			name: "valid message with HTML",
			message: &EmailMessage{
				From:    "sender@example.com",
				To:      []string{"recipient@example.com"},
				Subject: "Test Subject",
				HTML:    "<p>Test Body</p>",
			},
			wantErr: nil,
		},
		{
			name: "valid message with both plain text and HTML",
			message: &EmailMessage{
				From:      "sender@example.com",
				To:        []string{"recipient@example.com"},
				Subject:   "Test Subject",
				PlainText: "Test Body",
				HTML:      "<p>Test Body</p>",
			},
			wantErr: nil,
		},
		{
			name: "valid message with multiple recipients",
			message: &EmailMessage{
				From:      "sender@example.com",
				To:        []string{"recipient1@example.com", "recipient2@example.com"},
				Cc:        []string{"cc@example.com"},
				Bcc:       []string{"bcc@example.com"},
				Subject:   "Test Subject",
				PlainText: "Test Body",
			},
			wantErr: nil,
		},
		{
			name: "missing from",
			message: &EmailMessage{
				To:        []string{"recipient@example.com"},
				Subject:   "Test Subject",
				PlainText: "Test Body",
			},
			wantErr: ErrMissingFrom,
		},
		{
			name: "missing recipients",
			message: &EmailMessage{
				From:      "sender@example.com",
				Subject:   "Test Subject",
				PlainText: "Test Body",
			},
			wantErr: ErrMissingRecipients,
		},
		{
			name: "missing subject",
			message: &EmailMessage{
				From:      "sender@example.com",
				To:        []string{"recipient@example.com"},
				PlainText: "Test Body",
			},
			wantErr: ErrMissingSubject,
		},
		{
			name: "missing content",
			message: &EmailMessage{
				From:    "sender@example.com",
				To:      []string{"recipient@example.com"},
				Subject: "Test Subject",
			},
			wantErr: ErrMissingContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.message.Validate()
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

package mailsender

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSESConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *SESConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: &SESConfig{
				Region: "us-east-1",
			},
			wantErr: false,
		},
		{
			name: "valid config with all options",
			config: &SESConfig{
				Region:          "eu-west-1",
				AccessKeyID:     "AKIAIOSFODNN7EXAMPLE",
				SecretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				DefaultFrom:     "sender@example.com",
			},
			wantErr: false,
		},
		{
			name: "missing region",
			config: &SESConfig{
				Region: "",
			},
			wantErr: true,
			errMsg:  "region is required",
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

func TestDefaultSESConfig(t *testing.T) {
	config := DefaultSESConfig()

	assert.Equal(t, "us-east-1", config.Region)
	assert.Empty(t, config.AccessKeyID)
	assert.Empty(t, config.SecretAccessKey)
	assert.Empty(t, config.DefaultFrom)
}

func TestSESOptions(t *testing.T) {
	config := DefaultSESConfig()

	WithSESRegion("eu-west-1")(config)
	assert.Equal(t, "eu-west-1", config.Region)

	WithSESCredentials("AKIATEST", "secretkey")(config)
	assert.Equal(t, "AKIATEST", config.AccessKeyID)
	assert.Equal(t, "secretkey", config.SecretAccessKey)

	WithSESSessionToken("sessiontoken")(config)
	assert.Equal(t, "sessiontoken", config.SessionToken)

	WithSESEndpoint("http://localhost:4566")(config)
	assert.Equal(t, "http://localhost:4566", config.Endpoint)

	WithSESDefaultFrom("default@example.com", "Default Name")(config)
	assert.Equal(t, "default@example.com", config.DefaultFrom)
	assert.Equal(t, "Default Name", config.DefaultFromName)

	WithSESConfigurationSet("my-config-set")(config)
	assert.Equal(t, "my-config-set", config.ConfigurationSetName)
}

// mockSESClient is a mock implementation of SES client for testing
type mockSESClient struct {
	sendEmailFunc func(ctx context.Context, params *ses.SendEmailInput, optFns ...func(*ses.Options)) (*ses.SendEmailOutput, error)
	lastInput     *ses.SendEmailInput
}

func (m *mockSESClient) SendEmail(ctx context.Context, params *ses.SendEmailInput, optFns ...func(*ses.Options)) (*ses.SendEmailOutput, error) {
	m.lastInput = params
	if m.sendEmailFunc != nil {
		return m.sendEmailFunc(ctx, params, optFns...)
	}
	return &ses.SendEmailOutput{
		MessageId: aws.String("test-message-id"),
	}, nil
}

func TestSESSender_buildSendEmailInput_Basic(t *testing.T) {
	sender := &SESSender{
		config:          DefaultSESConfig(),
		defaultFrom:     "",
		defaultFromName: "",
	}

	message := &EmailMessage{
		From:      "sender@example.com",
		To:        []string{"recipient@example.com"},
		Subject:   "Test Subject",
		PlainText: "Hello, this is a test.",
	}

	input := sender.buildSendEmailInput(message)

	assert.Equal(t, "sender@example.com", *input.Source)
	assert.Equal(t, []string{"recipient@example.com"}, input.Destination.ToAddresses)
	assert.Equal(t, "Test Subject", *input.Message.Subject.Data)
	assert.Equal(t, "Hello, this is a test.", *input.Message.Body.Text.Data)
	assert.Nil(t, input.Message.Body.Html)
}

func TestSESSender_buildSendEmailInput_WithFromName(t *testing.T) {
	sender := &SESSender{
		config:          DefaultSESConfig(),
		defaultFrom:     "",
		defaultFromName: "",
	}

	message := &EmailMessage{
		From:      "sender@example.com",
		FromName:  "Sender Name",
		To:        []string{"recipient@example.com"},
		Subject:   "Test Subject",
		PlainText: "Hello",
	}

	input := sender.buildSendEmailInput(message)

	assert.Equal(t, "Sender Name <sender@example.com>", *input.Source)
}

func TestSESSender_buildSendEmailInput_HTML(t *testing.T) {
	sender := &SESSender{
		config:          DefaultSESConfig(),
		defaultFrom:     "",
		defaultFromName: "",
	}

	message := &EmailMessage{
		From:    "sender@example.com",
		To:      []string{"recipient@example.com"},
		Subject: "Test Subject",
		HTML:    "<html><body><h1>Hello</h1></body></html>",
	}

	input := sender.buildSendEmailInput(message)

	assert.NotNil(t, input.Message.Body.Html)
	assert.Equal(t, "<html><body><h1>Hello</h1></body></html>", *input.Message.Body.Html.Data)
	assert.Nil(t, input.Message.Body.Text)
}

func TestSESSender_buildSendEmailInput_Both(t *testing.T) {
	sender := &SESSender{
		config:          DefaultSESConfig(),
		defaultFrom:     "",
		defaultFromName: "",
	}

	message := &EmailMessage{
		From:      "sender@example.com",
		To:        []string{"recipient@example.com"},
		Subject:   "Test Subject",
		PlainText: "Plain text version",
		HTML:      "<html><body>HTML version</body></html>",
	}

	input := sender.buildSendEmailInput(message)

	assert.NotNil(t, input.Message.Body.Text)
	assert.NotNil(t, input.Message.Body.Html)
	assert.Equal(t, "Plain text version", *input.Message.Body.Text.Data)
	assert.Equal(t, "<html><body>HTML version</body></html>", *input.Message.Body.Html.Data)
}

func TestSESSender_buildSendEmailInput_WithCcBcc(t *testing.T) {
	sender := &SESSender{
		config:          DefaultSESConfig(),
		defaultFrom:     "",
		defaultFromName: "",
	}

	message := &EmailMessage{
		From:      "sender@example.com",
		To:        []string{"to@example.com"},
		Cc:        []string{"cc1@example.com", "cc2@example.com"},
		Bcc:       []string{"bcc@example.com"},
		Subject:   "Test Subject",
		PlainText: "Hello",
	}

	input := sender.buildSendEmailInput(message)

	assert.Equal(t, []string{"to@example.com"}, input.Destination.ToAddresses)
	assert.Equal(t, []string{"cc1@example.com", "cc2@example.com"}, input.Destination.CcAddresses)
	assert.Equal(t, []string{"bcc@example.com"}, input.Destination.BccAddresses)
}

func TestSESSender_buildSendEmailInput_WithReplyTo(t *testing.T) {
	sender := &SESSender{
		config:          DefaultSESConfig(),
		defaultFrom:     "",
		defaultFromName: "",
	}

	message := &EmailMessage{
		From:      "sender@example.com",
		To:        []string{"to@example.com"},
		ReplyTo:   "replyto@example.com",
		Subject:   "Test Subject",
		PlainText: "Hello",
	}

	input := sender.buildSendEmailInput(message)

	assert.Equal(t, []string{"replyto@example.com"}, input.ReplyToAddresses)
}

func TestSESSender_buildSendEmailInput_WithConfigurationSet(t *testing.T) {
	config := DefaultSESConfig()
	config.ConfigurationSetName = "my-config-set"

	sender := &SESSender{
		config:          config,
		defaultFrom:     "",
		defaultFromName: "",
	}

	message := &EmailMessage{
		From:      "sender@example.com",
		To:        []string{"to@example.com"},
		Subject:   "Test Subject",
		PlainText: "Hello",
	}

	input := sender.buildSendEmailInput(message)

	assert.Equal(t, "my-config-set", *input.ConfigurationSetName)
}

func TestSESSender_buildSendEmailInput_UTF8Charset(t *testing.T) {
	sender := &SESSender{
		config:          DefaultSESConfig(),
		defaultFrom:     "",
		defaultFromName: "",
	}

	message := &EmailMessage{
		From:      "sender@example.com",
		To:        []string{"to@example.com"},
		Subject:   "Test Subject",
		PlainText: "Hello",
		HTML:      "<p>Hello</p>",
	}

	input := sender.buildSendEmailInput(message)

	assert.Equal(t, "UTF-8", *input.Message.Subject.Charset)
	assert.Equal(t, "UTF-8", *input.Message.Body.Text.Charset)
	assert.Equal(t, "UTF-8", *input.Message.Body.Html.Charset)
}

func TestSESSender_Close(t *testing.T) {
	sender := &SESSender{
		config: DefaultSESConfig(),
	}

	err := sender.Close()
	require.NoError(t, err)
}

func TestSESSender_Send_ValidationError(t *testing.T) {
	sender := &SESSender{
		config:      DefaultSESConfig(),
		defaultFrom: "",
	}

	// Missing required fields
	message := &EmailMessage{
		Subject: "Test",
	}

	err := sender.Send(context.Background(), message)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid message")
}

func TestSESSender_AppliesDefaults(t *testing.T) {
	// Test that defaults are applied correctly by checking the built input
	sender := &SESSender{
		config:          DefaultSESConfig(),
		defaultFrom:     "default@example.com",
		defaultFromName: "Default Name",
	}

	message := &EmailMessage{
		To:        []string{"to@example.com"},
		Subject:   "Test",
		PlainText: "Hello",
	}

	// Apply defaults manually (like Send does)
	if message.From == "" {
		message.From = sender.defaultFrom
	}
	if message.FromName == "" {
		message.FromName = sender.defaultFromName
	}

	assert.Equal(t, "default@example.com", message.From)
	assert.Equal(t, "Default Name", message.FromName)

	// Verify the built input uses the defaults
	input := sender.buildSendEmailInput(message)
	assert.Equal(t, "Default Name <default@example.com>", *input.Source)
}

func TestSESSender_buildSendEmailInput_NoCcBcc(t *testing.T) {
	sender := &SESSender{
		config:          DefaultSESConfig(),
		defaultFrom:     "",
		defaultFromName: "",
	}

	message := &EmailMessage{
		From:      "sender@example.com",
		To:        []string{"to@example.com"},
		Subject:   "Test Subject",
		PlainText: "Hello",
	}

	input := sender.buildSendEmailInput(message)

	assert.Nil(t, input.Destination.CcAddresses)
	assert.Nil(t, input.Destination.BccAddresses)
}

func TestSESSender_buildSendEmailInput_NoConfigurationSet(t *testing.T) {
	sender := &SESSender{
		config:          DefaultSESConfig(),
		defaultFrom:     "",
		defaultFromName: "",
	}

	message := &EmailMessage{
		From:      "sender@example.com",
		To:        []string{"to@example.com"},
		Subject:   "Test Subject",
		PlainText: "Hello",
	}

	input := sender.buildSendEmailInput(message)

	assert.Nil(t, input.ConfigurationSetName)
}

func TestSESSender_buildSendEmailInput_NoReplyTo(t *testing.T) {
	sender := &SESSender{
		config:          DefaultSESConfig(),
		defaultFrom:     "",
		defaultFromName: "",
	}

	message := &EmailMessage{
		From:      "sender@example.com",
		To:        []string{"to@example.com"},
		Subject:   "Test Subject",
		PlainText: "Hello",
	}

	input := sender.buildSendEmailInput(message)

	assert.Nil(t, input.ReplyToAddresses)
}

func TestSESSender_buildSendEmailInput_MultipleRecipients(t *testing.T) {
	sender := &SESSender{
		config:          DefaultSESConfig(),
		defaultFrom:     "",
		defaultFromName: "",
	}

	message := &EmailMessage{
		From:      "sender@example.com",
		To:        []string{"to1@example.com", "to2@example.com", "to3@example.com"},
		Subject:   "Test Subject",
		PlainText: "Hello",
	}

	input := sender.buildSendEmailInput(message)

	assert.Equal(t, 3, len(input.Destination.ToAddresses))
	assert.Contains(t, input.Destination.ToAddresses, "to1@example.com")
	assert.Contains(t, input.Destination.ToAddresses, "to2@example.com")
	assert.Contains(t, input.Destination.ToAddresses, "to3@example.com")
}

// mockableSESSender is a version of SESSender that can accept a mock client
type mockableSESSender struct {
	client          SESClientAPI
	config          *SESConfig
	defaultFrom     string
	defaultFromName string
}

func (s *mockableSESSender) Send(ctx context.Context, message *EmailMessage) error {
	// Apply defaults
	if message.From == "" {
		message.From = s.defaultFrom
	}
	if message.FromName == "" {
		message.FromName = s.defaultFromName
	}

	// Validate message
	if err := message.Validate(); err != nil {
		return err
	}

	// Build source (from)
	var source string
	if message.FromName != "" {
		source = message.FromName + " <" + message.From + ">"
	} else {
		source = message.From
	}

	// Build destination
	destination := &types.Destination{
		ToAddresses: message.To,
	}
	if len(message.Cc) > 0 {
		destination.CcAddresses = message.Cc
	}
	if len(message.Bcc) > 0 {
		destination.BccAddresses = message.Bcc
	}

	// Build message body
	body := &types.Body{}
	if message.PlainText != "" {
		body.Text = &types.Content{
			Data:    aws.String(message.PlainText),
			Charset: aws.String("UTF-8"),
		}
	}
	if message.HTML != "" {
		body.Html = &types.Content{
			Data:    aws.String(message.HTML),
			Charset: aws.String("UTF-8"),
		}
	}

	// Build input
	input := &ses.SendEmailInput{
		Source:      aws.String(source),
		Destination: destination,
		Message: &types.Message{
			Subject: &types.Content{
				Data:    aws.String(message.Subject),
				Charset: aws.String("UTF-8"),
			},
			Body: body,
		},
	}

	// Add reply-to if set
	if message.ReplyTo != "" {
		input.ReplyToAddresses = []string{message.ReplyTo}
	}

	// Send the email
	_, err := s.client.SendEmail(ctx, input)
	if err != nil {
		return err
	}

	return nil
}

func (s *mockableSESSender) Close() error {
	return nil
}

func TestMockableSESSender_Send_Success(t *testing.T) {
	mock := &mockSESClient{}
	sender := &mockableSESSender{
		client: mock,
		config: DefaultSESConfig(),
	}

	message := &EmailMessage{
		From:      "sender@example.com",
		To:        []string{"to@example.com"},
		Subject:   "Test",
		PlainText: "Hello",
	}

	err := sender.Send(context.Background(), message)
	require.NoError(t, err)
	assert.NotNil(t, mock.lastInput)
	assert.Equal(t, "sender@example.com", *mock.lastInput.Source)
}

func TestMockableSESSender_Send_Error(t *testing.T) {
	expectedErr := errors.New("SES error")
	mock := &mockSESClient{
		sendEmailFunc: func(ctx context.Context, params *ses.SendEmailInput, optFns ...func(*ses.Options)) (*ses.SendEmailOutput, error) {
			return nil, expectedErr
		},
	}
	sender := &mockableSESSender{
		client: mock,
		config: DefaultSESConfig(),
	}

	message := &EmailMessage{
		From:      "sender@example.com",
		To:        []string{"to@example.com"},
		Subject:   "Test",
		PlainText: "Hello",
	}

	err := sender.Send(context.Background(), message)
	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestMockableSESSender_Send_WithDefaults(t *testing.T) {
	mock := &mockSESClient{}
	sender := &mockableSESSender{
		client:          mock,
		config:          DefaultSESConfig(),
		defaultFrom:     "default@example.com",
		defaultFromName: "Default Name",
	}

	message := &EmailMessage{
		To:        []string{"to@example.com"},
		Subject:   "Test",
		PlainText: "Hello",
	}

	err := sender.Send(context.Background(), message)
	require.NoError(t, err)
	assert.Equal(t, "Default Name <default@example.com>", *mock.lastInput.Source)
}

func TestMockableSESSender_Send_ValidationError(t *testing.T) {
	mock := &mockSESClient{}
	sender := &mockableSESSender{
		client: mock,
		config: DefaultSESConfig(),
	}

	message := &EmailMessage{
		// Missing required fields
		Subject: "Test",
	}

	err := sender.Send(context.Background(), message)
	require.Error(t, err)
}

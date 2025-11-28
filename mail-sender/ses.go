package mailsender

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
)

// SESConfig holds the configuration for the AWS SES sender.
type SESConfig struct {
	// Region is the AWS region (e.g., "us-east-1").
	Region string

	// AccessKeyID is the AWS access key ID.
	// If empty, the SDK will use the default credential chain.
	AccessKeyID string

	// SecretAccessKey is the AWS secret access key.
	// If empty, the SDK will use the default credential chain.
	SecretAccessKey string

	// SessionToken is the AWS session token (optional, for temporary credentials).
	SessionToken string

	// Endpoint is a custom endpoint URL (optional, for LocalStack or testing).
	Endpoint string

	// DefaultFrom is the default sender email address.
	DefaultFrom string

	// DefaultFromName is the default sender name.
	DefaultFromName string

	// ConfigurationSetName is the SES configuration set name (optional).
	ConfigurationSetName string
}

// Validate validates the SES configuration.
func (c *SESConfig) Validate() error {
	if c.Region == "" {
		return fmt.Errorf("AWS region is required")
	}
	return nil
}

// DefaultSESConfig returns a new SESConfig with default values.
func DefaultSESConfig() *SESConfig {
	return &SESConfig{
		Region: "us-east-1",
	}
}

// SESSender implements the EmailSender interface using AWS SES.
type SESSender struct {
	client          *ses.Client
	config          *SESConfig
	defaultFrom     string
	defaultFromName string
}

// NewSES creates a new AWS SES email sender with the provided configuration.
func NewSES(ctx context.Context, sesConfig *SESConfig) (*SESSender, error) {
	if sesConfig == nil {
		sesConfig = DefaultSESConfig()
	}

	if err := sesConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid SES config: %w", err)
	}

	// Build AWS config options
	var opts []func(*config.LoadOptions) error
	opts = append(opts, config.WithRegion(sesConfig.Region))

	// Use explicit credentials if provided
	if sesConfig.AccessKeyID != "" && sesConfig.SecretAccessKey != "" {
		opts = append(opts, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				sesConfig.AccessKeyID,
				sesConfig.SecretAccessKey,
				sesConfig.SessionToken,
			),
		))
	}

	// Load AWS config
	awsConfig, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create SES client options
	var clientOpts []func(*ses.Options)
	if sesConfig.Endpoint != "" {
		clientOpts = append(clientOpts, func(o *ses.Options) {
			o.BaseEndpoint = aws.String(sesConfig.Endpoint)
		})
	}

	client := ses.NewFromConfig(awsConfig, clientOpts...)

	return &SESSender{
		client:          client,
		config:          sesConfig,
		defaultFrom:     sesConfig.DefaultFrom,
		defaultFromName: sesConfig.DefaultFromName,
	}, nil
}

// SESOption is a functional option for configuring SESSender.
type SESOption func(*SESConfig)

// WithSESRegion sets the AWS region.
func WithSESRegion(region string) SESOption {
	return func(c *SESConfig) {
		c.Region = region
	}
}

// WithSESCredentials sets the AWS credentials.
func WithSESCredentials(accessKeyID, secretAccessKey string) SESOption {
	return func(c *SESConfig) {
		c.AccessKeyID = accessKeyID
		c.SecretAccessKey = secretAccessKey
	}
}

// WithSESSessionToken sets the AWS session token.
func WithSESSessionToken(token string) SESOption {
	return func(c *SESConfig) {
		c.SessionToken = token
	}
}

// WithSESEndpoint sets a custom endpoint URL.
func WithSESEndpoint(endpoint string) SESOption {
	return func(c *SESConfig) {
		c.Endpoint = endpoint
	}
}

// WithSESDefaultFrom sets the default from address.
func WithSESDefaultFrom(from, fromName string) SESOption {
	return func(c *SESConfig) {
		c.DefaultFrom = from
		c.DefaultFromName = fromName
	}
}

// WithSESConfigurationSet sets the SES configuration set name.
func WithSESConfigurationSet(name string) SESOption {
	return func(c *SESConfig) {
		c.ConfigurationSetName = name
	}
}

// NewSESWithOptions creates a new SES sender with functional options.
func NewSESWithOptions(ctx context.Context, opts ...SESOption) (*SESSender, error) {
	cfg := DefaultSESConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	return NewSES(ctx, cfg)
}

// Send sends an email using AWS SES.
func (s *SESSender) Send(ctx context.Context, message *EmailMessage) error {
	// Apply defaults
	if message.From == "" {
		message.From = s.defaultFrom
	}
	if message.FromName == "" {
		message.FromName = s.defaultFromName
	}

	// Validate message
	if err := message.Validate(); err != nil {
		return fmt.Errorf("invalid message: %w", err)
	}

	// Build SES input
	input := s.buildSendEmailInput(message)

	// Send the email
	_, err := s.client.SendEmail(ctx, input)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSendFailed, err)
	}

	return nil
}

// Close closes the SES sender.
func (s *SESSender) Close() error {
	// AWS SDK clients don't require explicit cleanup
	return nil
}

// buildSendEmailInput builds the SES SendEmailInput from an EmailMessage.
func (s *SESSender) buildSendEmailInput(message *EmailMessage) *ses.SendEmailInput {
	// Build source (from)
	var source string
	if message.FromName != "" {
		source = fmt.Sprintf("%s <%s>", message.FromName, message.From)
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

	// Add configuration set if configured
	if s.config.ConfigurationSetName != "" {
		input.ConfigurationSetName = aws.String(s.config.ConfigurationSetName)
	}

	return input
}

// SESClientAPI defines the interface for SES client operations used by SESSender.
// This interface is useful for testing with mock clients.
type SESClientAPI interface {
	SendEmail(ctx context.Context, params *ses.SendEmailInput, optFns ...func(*ses.Options)) (*ses.SendEmailOutput, error)
}

// testableSesSender is used internally for testing with mock clients.
type testableSesSender struct {
	clientAPI       SESClientAPI
	config          *SESConfig
	defaultFrom     string
	defaultFromName string
}

// newTestableSesSender creates a new testable SES sender with a mock client.
func newTestableSesSender(client SESClientAPI, cfg *SESConfig) *testableSesSender {
	if cfg == nil {
		cfg = DefaultSESConfig()
	}
	return &testableSesSender{
		clientAPI:       client,
		config:          cfg,
		defaultFrom:     cfg.DefaultFrom,
		defaultFromName: cfg.DefaultFromName,
	}
}

// Send sends an email using the mock SES client.
func (s *testableSesSender) Send(ctx context.Context, message *EmailMessage) error {
	// Apply defaults
	if message.From == "" {
		message.From = s.defaultFrom
	}
	if message.FromName == "" {
		message.FromName = s.defaultFromName
	}

	// Validate message
	if err := message.Validate(); err != nil {
		return fmt.Errorf("invalid message: %w", err)
	}

	// Build source (from)
	var source string
	if message.FromName != "" {
		source = fmt.Sprintf("%s <%s>", message.FromName, message.From)
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

	// Add configuration set if configured
	if s.config.ConfigurationSetName != "" {
		input.ConfigurationSetName = aws.String(s.config.ConfigurationSetName)
	}

	// Send the email
	_, err := s.clientAPI.SendEmail(ctx, input)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSendFailed, err)
	}

	return nil
}

// Close closes the testable SES sender.
func (s *testableSesSender) Close() error {
	return nil
}

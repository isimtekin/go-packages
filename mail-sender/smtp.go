package mailsender

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"
)

// SMTPConfig holds the configuration for the SMTP sender.
type SMTPConfig struct {
	// Host is the SMTP server hostname.
	Host string

	// Port is the SMTP server port.
	Port int

	// Username is the SMTP authentication username.
	Username string

	// Password is the SMTP authentication password.
	Password string

	// UseTLS enables TLS encryption.
	UseTLS bool

	// InsecureSkipVerify skips TLS certificate verification.
	InsecureSkipVerify bool

	// Timeout is the connection timeout.
	Timeout time.Duration

	// DefaultFrom is the default sender email address.
	DefaultFrom string

	// DefaultFromName is the default sender name.
	DefaultFromName string

	// LocalName is the hostname sent to the SMTP server with HELO/EHLO.
	// If empty, "localhost" is used.
	LocalName string
}

// Validate validates the SMTP configuration.
func (c *SMTPConfig) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("SMTP host is required")
	}
	if c.Port <= 0 {
		return fmt.Errorf("SMTP port must be positive")
	}
	return nil
}

// DefaultSMTPConfig returns a new SMTPConfig with default values.
func DefaultSMTPConfig() *SMTPConfig {
	return &SMTPConfig{
		Port:      587,
		UseTLS:    true,
		Timeout:   30 * time.Second,
		LocalName: "localhost",
	}
}

// SMTPSender implements the EmailSender interface using SMTP.
type SMTPSender struct {
	config *SMTPConfig
}

// NewSMTP creates a new SMTP email sender with the provided configuration.
func NewSMTP(config *SMTPConfig) (*SMTPSender, error) {
	if config == nil {
		config = DefaultSMTPConfig()
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid SMTP config: %w", err)
	}

	return &SMTPSender{
		config: config,
	}, nil
}

// SMTPOption is a functional option for configuring SMTPSender.
type SMTPOption func(*SMTPConfig)

// WithSMTPHost sets the SMTP server host.
func WithSMTPHost(host string) SMTPOption {
	return func(c *SMTPConfig) {
		c.Host = host
	}
}

// WithSMTPPort sets the SMTP server port.
func WithSMTPPort(port int) SMTPOption {
	return func(c *SMTPConfig) {
		c.Port = port
	}
}

// WithSMTPAuth sets the SMTP authentication credentials.
func WithSMTPAuth(username, password string) SMTPOption {
	return func(c *SMTPConfig) {
		c.Username = username
		c.Password = password
	}
}

// WithSMTPTLS enables or disables TLS.
func WithSMTPTLS(useTLS bool) SMTPOption {
	return func(c *SMTPConfig) {
		c.UseTLS = useTLS
	}
}

// WithSMTPInsecureSkipVerify sets whether to skip TLS certificate verification.
func WithSMTPInsecureSkipVerify(skip bool) SMTPOption {
	return func(c *SMTPConfig) {
		c.InsecureSkipVerify = skip
	}
}

// WithSMTPTimeout sets the connection timeout.
func WithSMTPTimeout(timeout time.Duration) SMTPOption {
	return func(c *SMTPConfig) {
		c.Timeout = timeout
	}
}

// WithSMTPDefaultFrom sets the default from address.
func WithSMTPDefaultFrom(from, fromName string) SMTPOption {
	return func(c *SMTPConfig) {
		c.DefaultFrom = from
		c.DefaultFromName = fromName
	}
}

// WithSMTPLocalName sets the local hostname for HELO/EHLO.
func WithSMTPLocalName(name string) SMTPOption {
	return func(c *SMTPConfig) {
		c.LocalName = name
	}
}

// NewSMTPWithOptions creates a new SMTP sender with functional options.
func NewSMTPWithOptions(opts ...SMTPOption) (*SMTPSender, error) {
	config := DefaultSMTPConfig()
	for _, opt := range opts {
		opt(config)
	}
	return NewSMTP(config)
}

// Send sends an email using SMTP.
func (s *SMTPSender) Send(ctx context.Context, message *EmailMessage) error {
	// Apply defaults
	if message.From == "" {
		message.From = s.config.DefaultFrom
	}
	if message.FromName == "" {
		message.FromName = s.config.DefaultFromName
	}

	// Validate message
	if err := message.Validate(); err != nil {
		return fmt.Errorf("invalid message: %w", err)
	}

	// Build the email
	emailBody := s.buildEmail(message)

	// Get all recipients
	recipients := s.getAllRecipients(message)

	// Send the email
	if err := s.sendMail(ctx, message.From, recipients, emailBody); err != nil {
		return fmt.Errorf("%w: %v", ErrSendFailed, err)
	}

	return nil
}

// Close closes the SMTP sender.
func (s *SMTPSender) Close() error {
	// SMTP connections are not persistent, nothing to close
	return nil
}

// sendMail sends the email via SMTP with context support.
func (s *SMTPSender) sendMail(ctx context.Context, from string, to []string, body []byte) error {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	// Create a channel for the result
	done := make(chan error, 1)

	go func() {
		done <- s.doSendMail(addr, from, to, body)
	}()

	// Wait for completion or context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	}
}

// doSendMail performs the actual SMTP sending.
func (s *SMTPSender) doSendMail(addr string, from string, to []string, body []byte) error {
	// Create connection with timeout
	dialer := &net.Dialer{
		Timeout: s.config.Timeout,
	}

	var conn net.Conn
	var err error

	if s.config.UseTLS && s.config.Port == 465 {
		// Implicit TLS (SMTPS)
		tlsConfig := &tls.Config{
			ServerName:         s.config.Host,
			InsecureSkipVerify: s.config.InsecureSkipVerify,
		}
		conn, err = tls.DialWithDialer(dialer, "tcp", addr, tlsConfig)
	} else {
		conn, err = dialer.Dial("tcp", addr)
	}

	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	// Create SMTP client
	client, err := smtp.NewClient(conn, s.config.Host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	// Set local name
	localName := s.config.LocalName
	if localName == "" {
		localName = "localhost"
	}
	if err := client.Hello(localName); err != nil {
		return fmt.Errorf("HELO failed: %w", err)
	}

	// Start TLS if enabled and not already using implicit TLS
	if s.config.UseTLS && s.config.Port != 465 {
		if ok, _ := client.Extension("STARTTLS"); ok {
			tlsConfig := &tls.Config{
				ServerName:         s.config.Host,
				InsecureSkipVerify: s.config.InsecureSkipVerify,
			}
			if err := client.StartTLS(tlsConfig); err != nil {
				return fmt.Errorf("STARTTLS failed: %w", err)
			}
		}
	}

	// Authenticate if credentials provided
	if s.config.Username != "" && s.config.Password != "" {
		auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}
	}

	// Set sender
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("MAIL FROM failed: %w", err)
	}

	// Set recipients
	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("RCPT TO failed for %s: %w", recipient, err)
		}
	}

	// Send body
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA failed: %w", err)
	}

	_, err = writer.Write(body)
	if err != nil {
		return fmt.Errorf("failed to write email body: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	// Quit
	return client.Quit()
}

// buildEmail builds the raw email content.
func (s *SMTPSender) buildEmail(message *EmailMessage) []byte {
	var builder strings.Builder

	// From header
	if message.FromName != "" {
		builder.WriteString(fmt.Sprintf("From: %s <%s>\r\n", message.FromName, message.From))
	} else {
		builder.WriteString(fmt.Sprintf("From: %s\r\n", message.From))
	}

	// To header
	builder.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(message.To, ", ")))

	// Cc header
	if len(message.Cc) > 0 {
		builder.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(message.Cc, ", ")))
	}

	// Reply-To header
	if message.ReplyTo != "" {
		builder.WriteString(fmt.Sprintf("Reply-To: %s\r\n", message.ReplyTo))
	}

	// Subject header
	builder.WriteString(fmt.Sprintf("Subject: %s\r\n", message.Subject))

	// MIME headers
	builder.WriteString("MIME-Version: 1.0\r\n")

	// Content based on what's provided
	if message.HTML != "" && message.PlainText != "" {
		// Multipart message
		boundary := "boundary-" + fmt.Sprintf("%d", time.Now().UnixNano())
		builder.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n", boundary))
		builder.WriteString("\r\n")

		// Plain text part
		builder.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		builder.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
		builder.WriteString("Content-Transfer-Encoding: quoted-printable\r\n")
		builder.WriteString("\r\n")
		builder.WriteString(message.PlainText)
		builder.WriteString("\r\n")

		// HTML part
		builder.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		builder.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n")
		builder.WriteString("Content-Transfer-Encoding: quoted-printable\r\n")
		builder.WriteString("\r\n")
		builder.WriteString(message.HTML)
		builder.WriteString("\r\n")

		// End boundary
		builder.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	} else if message.HTML != "" {
		// HTML only
		builder.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n")
		builder.WriteString("\r\n")
		builder.WriteString(message.HTML)
	} else {
		// Plain text only
		builder.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
		builder.WriteString("\r\n")
		builder.WriteString(message.PlainText)
	}

	return []byte(builder.String())
}

// getAllRecipients returns all recipients (To, Cc, Bcc).
func (s *SMTPSender) getAllRecipients(message *EmailMessage) []string {
	recipients := make([]string, 0, len(message.To)+len(message.Cc)+len(message.Bcc))
	recipients = append(recipients, message.To...)
	recipients = append(recipients, message.Cc...)
	recipients = append(recipients, message.Bcc...)
	return recipients
}

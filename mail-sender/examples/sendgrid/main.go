package main

import (
	"context"
	"fmt"
	"log"

	mailsender "github.com/isimtekin/go-packages/mail-sender"
)

func main() {
	// Example 1: Using NewSendGridWithOptions
	sender, err := mailsender.NewSendGridWithOptions(
		mailsender.WithAPIKey("your-sendgrid-api-key"),
		mailsender.WithDefaultFrom("sender@example.com"),
		mailsender.WithDefaultFromName("Example Sender"),
	)
	if err != nil {
		log.Fatalf("Failed to create SendGrid sender: %v", err)
	}
	defer sender.Close()

	// Example 2: Send plain text email
	err = sender.Send(context.Background(), &mailsender.EmailMessage{
		To:        []string{"recipient@example.com"},
		Subject:   "Plain Text Email",
		PlainText: "This is a plain text email message.",
	})
	if err != nil {
		log.Printf("Failed to send plain text email: %v", err)
	} else {
		fmt.Println("Plain text email sent successfully!")
	}

	// Example 3: Send HTML email
	err = sender.Send(context.Background(), &mailsender.EmailMessage{
		To:      []string{"recipient@example.com"},
		Subject: "HTML Email",
		HTML:    "<h1>Hello World</h1><p>This is an HTML email.</p>",
	})
	if err != nil {
		log.Printf("Failed to send HTML email: %v", err)
	} else {
		fmt.Println("HTML email sent successfully!")
	}

	// Example 4: Send email with both plain text and HTML
	err = sender.Send(context.Background(), &mailsender.EmailMessage{
		To:        []string{"recipient@example.com"},
		Subject:   "Multi-part Email",
		PlainText: "This is the plain text version.",
		HTML:      "<h1>This is the HTML version</h1>",
	})
	if err != nil {
		log.Printf("Failed to send multi-part email: %v", err)
	} else {
		fmt.Println("Multi-part email sent successfully!")
	}

	// Example 5: Send email with multiple recipients
	err = sender.Send(context.Background(), &mailsender.EmailMessage{
		To:        []string{"recipient1@example.com", "recipient2@example.com"},
		Cc:        []string{"cc@example.com"},
		Bcc:       []string{"bcc@example.com"},
		Subject:   "Email to Multiple Recipients",
		PlainText: "This email goes to multiple recipients.",
	})
	if err != nil {
		log.Printf("Failed to send email to multiple recipients: %v", err)
	} else {
		fmt.Println("Email to multiple recipients sent successfully!")
	}

	// Example 6: Send email with custom reply-to
	err = sender.Send(context.Background(), &mailsender.EmailMessage{
		To:        []string{"recipient@example.com"},
		Subject:   "Email with Reply-To",
		PlainText: "Please reply to the support address.",
		ReplyTo:   "support@example.com",
	})
	if err != nil {
		log.Printf("Failed to send email with reply-to: %v", err)
	} else {
		fmt.Println("Email with reply-to sent successfully!")
	}

	// Example 7: Using environment variables
	senderFromEnv, err := mailsender.NewSendGridFromEnv()
	if err != nil {
		log.Printf("Failed to create sender from env: %v", err)
		return
	}
	defer senderFromEnv.Close()

	err = senderFromEnv.Send(context.Background(), &mailsender.EmailMessage{
		To:        []string{"recipient@example.com"},
		Subject:   "Email from Environment Config",
		PlainText: "This email was sent using environment variable configuration.",
	})
	if err != nil {
		log.Printf("Failed to send email from env: %v", err)
	} else {
		fmt.Println("Email from environment config sent successfully!")
	}
}

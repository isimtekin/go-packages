package main

import (
	"context"
	"fmt"
	"log"

	mailsender "github.com/isimtekin/go-packages/mail-sender"
)

func main() {
	// Create SendGrid sender
	sender, err := mailsender.NewSendGridWithOptions(
		mailsender.WithAPIKey("your-sendgrid-api-key"),
		mailsender.WithDefaultFrom("sender@example.com"),
		mailsender.WithDefaultFromName("Example Sender"),
	)
	if err != nil {
		log.Fatalf("Failed to create SendGrid sender: %v", err)
	}
	defer sender.Close()

	// Example 1: Simple HTML template
	htmlTemplate := `
		<html>
			<body>
				<h1>Hello {{.Name}}!</h1>
				<p>Thank you for signing up for our service.</p>
				<p>Your account has been created successfully.</p>
			</body>
		</html>
	`

	data := map[string]string{
		"Name": "John Doe",
	}

	htmlContent, err := mailsender.RenderHTMLTemplate(htmlTemplate, data)
	if err != nil {
		log.Fatalf("Failed to render HTML template: %v", err)
	}

	err = sender.Send(context.Background(), &mailsender.EmailMessage{
		To:      []string{"recipient@example.com"},
		Subject: "Welcome to Our Service",
		HTML:    htmlContent,
	})
	if err != nil {
		log.Printf("Failed to send HTML template email: %v", err)
	} else {
		fmt.Println("HTML template email sent successfully!")
	}

	// Example 2: Plain text template
	textTemplate := `
Hello {{.Name}},

Thank you for your purchase!

Order Summary:
{{range .Items}}
- {{.}}
{{end}}

Total: ${{.Total}}

Best regards,
The Team
	`

	orderData := map[string]interface{}{
		"Name":  "Jane Smith",
		"Items": []string{"Product A", "Product B", "Product C"},
		"Total": "99.99",
	}

	textContent, err := mailsender.RenderTextTemplate(textTemplate, orderData)
	if err != nil {
		log.Fatalf("Failed to render text template: %v", err)
	}

	err = sender.Send(context.Background(), &mailsender.EmailMessage{
		To:        []string{"recipient@example.com"},
		Subject:   "Order Confirmation",
		PlainText: textContent,
	})
	if err != nil {
		log.Printf("Failed to send text template email: %v", err)
	} else {
		fmt.Println("Text template email sent successfully!")
	}

	// Example 3: Combined HTML and text templates
	htmlWelcomeTemplate := `
		<html>
			<head>
				<style>
					body { font-family: Arial, sans-serif; }
					.header { background-color: #4CAF50; color: white; padding: 20px; }
					.content { padding: 20px; }
					.footer { background-color: #f1f1f1; padding: 10px; text-align: center; }
				</style>
			</head>
			<body>
				<div class="header">
					<h1>Welcome {{.Name}}!</h1>
				</div>
				<div class="content">
					<p>Your verification code is: <strong>{{.Code}}</strong></p>
					<p>This code will expire in {{.ExpiryMinutes}} minutes.</p>
				</div>
				<div class="footer">
					<p>&copy; 2024 Example Company</p>
				</div>
			</body>
		</html>
	`

	textWelcomeTemplate := `
Welcome {{.Name}}!

Your verification code is: {{.Code}}

This code will expire in {{.ExpiryMinutes}} minutes.

---
© 2024 Example Company
	`

	welcomeData := map[string]interface{}{
		"Name":          "Alice Johnson",
		"Code":          "ABC123",
		"ExpiryMinutes": 15,
	}

	htmlWelcome, err := mailsender.RenderHTMLTemplate(htmlWelcomeTemplate, welcomeData)
	if err != nil {
		log.Fatalf("Failed to render HTML welcome template: %v", err)
	}

	textWelcome, err := mailsender.RenderTextTemplate(textWelcomeTemplate, welcomeData)
	if err != nil {
		log.Fatalf("Failed to render text welcome template: %v", err)
	}

	err = sender.Send(context.Background(), &mailsender.EmailMessage{
		To:        []string{"recipient@example.com"},
		Subject:   "Email Verification",
		PlainText: textWelcome,
		HTML:      htmlWelcome,
	})
	if err != nil {
		log.Printf("Failed to send combined template email: %v", err)
	} else {
		fmt.Println("Combined template email sent successfully!")
	}

	// Example 4: Template with conditionals
	conditionalTemplate := `
Hello {{.Name}},

{{if .Premium}}
Thank you for being a premium member! You have access to all our exclusive features.
{{else}}
Consider upgrading to premium for exclusive features and benefits.
{{end}}

{{if gt .MessageCount 0}}
You have {{.MessageCount}} new messages.
{{else}}
You have no new messages.
{{end}}

Best regards,
The Team
	`

	conditionalData := map[string]interface{}{
		"Name":         "Bob Wilson",
		"Premium":      true,
		"MessageCount": 5,
	}

	conditionalContent, err := mailsender.RenderTextTemplate(conditionalTemplate, conditionalData)
	if err != nil {
		log.Fatalf("Failed to render conditional template: %v", err)
	}

	err = sender.Send(context.Background(), &mailsender.EmailMessage{
		To:        []string{"recipient@example.com"},
		Subject:   "Account Update",
		PlainText: conditionalContent,
	})
	if err != nil {
		log.Printf("Failed to send conditional template email: %v", err)
	} else {
		fmt.Println("Conditional template email sent successfully!")
	}
}

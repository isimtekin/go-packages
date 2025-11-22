package main

import (
	"context"
	"fmt"
	"log"
	"time"

	mailsender "github.com/isimtekin/go-packages/mail-sender"
)

func main() {
	// Create a SendGrid sender
	sender, err := mailsender.NewSendGridWithOptions(
		mailsender.WithAPIKey("your-sendgrid-api-key"),
		mailsender.WithDefaultFrom("sender@example.com"),
		mailsender.WithDefaultFromName("Async Sender Demo"),
	)
	if err != nil {
		log.Fatalf("Failed to create SendGrid sender: %v", err)
	}

	// Example 1: Basic async sender
	fmt.Println("=== Example 1: Basic Async Sender ===")
	asyncSender := mailsender.NewAsyncSender(sender,
		mailsender.WithWorkers(3),
		mailsender.WithQueueSize(100),
	)

	// Send emails asynchronously (non-blocking)
	for i := 0; i < 5; i++ {
		err := asyncSender.SendAsync(context.Background(), &mailsender.EmailMessage{
			To:        []string{fmt.Sprintf("user%d@example.com", i)},
			Subject:   fmt.Sprintf("Email #%d", i),
			PlainText: fmt.Sprintf("This is email number %d", i),
		})
		if err != nil {
			log.Printf("Failed to queue email: %v", err)
		} else {
			fmt.Printf("Email #%d queued successfully\n", i)
		}
	}

	// Check stats
	time.Sleep(1 * time.Second)
	stats := asyncSender.Stats()
	fmt.Printf("Stats: Sent=%d, Failed=%d, Pending=%d\n\n",
		stats.Sent, stats.Failed, stats.Pending)

	asyncSender.Close()

	// Example 2: Async sender with event handlers
	fmt.Println("=== Example 2: Event Handlers ===")
	asyncSenderWithEvents := mailsender.NewAsyncSender(sender,
		mailsender.WithWorkers(2),
		mailsender.WithOnSuccess(func(msg *mailsender.EmailMessage) {
			fmt.Printf("✓ Email sent successfully to: %v\n", msg.To)
		}),
		mailsender.WithOnFailure(func(msg *mailsender.EmailMessage, err error) {
			fmt.Printf("✗ Failed to send email to %v: %v\n", msg.To, err)
		}),
	)

	// Send some emails
	for i := 0; i < 3; i++ {
		asyncSenderWithEvents.SendAsync(context.Background(), &mailsender.EmailMessage{
			To:        []string{fmt.Sprintf("recipient%d@example.com", i)},
			Subject:   "Event Handler Demo",
			PlainText: "This email demonstrates event handlers.",
		})
	}

	time.Sleep(1 * time.Second)
	asyncSenderWithEvents.Close()
	fmt.Println()

	// Example 3: Async sender with retry logic
	fmt.Println("=== Example 3: Retry Logic ===")
	asyncSenderWithRetry := mailsender.NewAsyncSender(sender,
		mailsender.WithWorkers(1),
		mailsender.WithRetry(3, 500*time.Millisecond), // 3 retries with 500ms delay
		mailsender.WithOnRetry(func(msg *mailsender.EmailMessage, attempt int, err error) {
			fmt.Printf("⟳ Retrying email to %v (attempt %d): %v\n", msg.To, attempt, err)
		}),
		mailsender.WithOnSuccess(func(msg *mailsender.EmailMessage) {
			fmt.Printf("✓ Email sent successfully to: %v\n", msg.To)
		}),
		mailsender.WithOnFailure(func(msg *mailsender.EmailMessage, err error) {
			fmt.Printf("✗ Failed permanently to send email to %v: %v\n", msg.To, err)
		}),
	)

	asyncSenderWithRetry.SendAsync(context.Background(), &mailsender.EmailMessage{
		To:        []string{"retry@example.com"},
		Subject:   "Retry Demo",
		PlainText: "This demonstrates retry logic.",
	})

	time.Sleep(3 * time.Second)
	stats = asyncSenderWithRetry.Stats()
	fmt.Printf("Stats: Sent=%d, Failed=%d, Retried=%d\n\n",
		stats.Sent, stats.Failed, stats.Retried)

	asyncSenderWithRetry.Close()

	// Example 4: Bulk email sending
	fmt.Println("=== Example 4: Bulk Email Sending ===")
	bulkSender := mailsender.NewAsyncSender(sender,
		mailsender.WithWorkers(5),
		mailsender.WithQueueSize(1000),
		mailsender.WithOnSuccess(func(msg *mailsender.EmailMessage) {
			// Silent success
		}),
	)

	// Send 50 emails in bulk
	startTime := time.Now()
	for i := 0; i < 50; i++ {
		err := bulkSender.SendAsync(context.Background(), &mailsender.EmailMessage{
			To:        []string{fmt.Sprintf("bulk%d@example.com", i)},
			Subject:   "Bulk Email",
			PlainText: "This is a bulk email.",
		})
		if err != nil {
			log.Printf("Failed to queue bulk email: %v", err)
		}
	}

	fmt.Println("All emails queued (non-blocking)")
	fmt.Printf("Time to queue: %v\n", time.Since(startTime))

	// Wait for completion
	time.Sleep(2 * time.Second)

	stats = bulkSender.Stats()
	fmt.Printf("Bulk Stats: Sent=%d, Failed=%d, Pending=%d\n\n",
		stats.Sent, stats.Failed, stats.Pending)

	bulkSender.Close()

	// Example 5: Graceful shutdown
	fmt.Println("=== Example 5: Graceful Shutdown ===")
	gracefulSender := mailsender.NewAsyncSender(sender,
		mailsender.WithWorkers(1),
		mailsender.WithQueueSize(10),
	)

	// Queue several emails
	for i := 0; i < 5; i++ {
		gracefulSender.SendAsync(context.Background(), &mailsender.EmailMessage{
			To:        []string{fmt.Sprintf("graceful%d@example.com", i)},
			Subject:   "Graceful Shutdown Demo",
			PlainText: "Testing graceful shutdown.",
		})
	}

	// Close waits for all queued emails to be sent
	fmt.Println("Closing gracefully (waiting for all emails to be sent)...")
	startClose := time.Now()
	err = gracefulSender.Close()
	if err != nil {
		log.Printf("Close error: %v", err)
	}
	fmt.Printf("Closed after %v\n\n", time.Since(startClose))

	// Example 6: Close with timeout
	fmt.Println("=== Example 6: Close with Timeout ===")
	timeoutSender := mailsender.NewAsyncSender(sender,
		mailsender.WithWorkers(1),
		mailsender.WithQueueSize(100),
	)

	// Queue many emails
	for i := 0; i < 20; i++ {
		timeoutSender.SendAsync(context.Background(), &mailsender.EmailMessage{
			To:        []string{fmt.Sprintf("timeout%d@example.com", i)},
			Subject:   "Timeout Demo",
			PlainText: "Testing close with timeout.",
		})
	}

	// Close with short timeout
	fmt.Println("Closing with 500ms timeout...")
	err = timeoutSender.CloseWithTimeout(500 * time.Millisecond)
	if err != nil {
		fmt.Printf("Close timeout error: %v\n", err)
	}

	stats = timeoutSender.Stats()
	fmt.Printf("Final Stats: Sent=%d, Failed=%d, Pending=%d\n\n",
		stats.Sent, stats.Failed, stats.Pending)

	// Example 7: Using all event handlers together
	fmt.Println("=== Example 7: Complete Event Handling ===")
	completeSender := mailsender.NewAsyncSender(sender,
		mailsender.WithWorkers(2),
		mailsender.WithQueueSize(50),
		mailsender.WithRetry(2, 200*time.Millisecond),
		mailsender.WithEventHandlers(&mailsender.EventHandlers{
			OnSuccess: func(msg *mailsender.EmailMessage) {
				fmt.Printf("✓ SUCCESS: Email to %v\n", msg.To)
			},
			OnFailure: func(msg *mailsender.EmailMessage, err error) {
				fmt.Printf("✗ FAILURE: Email to %v - %v\n", msg.To, err)
			},
			OnRetry: func(msg *mailsender.EmailMessage, attempt int, err error) {
				fmt.Printf("⟳ RETRY: Email to %v (attempt %d) - %v\n", msg.To, attempt, err)
			},
		}),
	)

	// Send emails
	for i := 0; i < 5; i++ {
		completeSender.SendAsync(context.Background(), &mailsender.EmailMessage{
			To:        []string{fmt.Sprintf("complete%d@example.com", i)},
			Subject:   "Complete Demo",
			PlainText: "Demonstrating all event handlers.",
		})
	}

	time.Sleep(2 * time.Second)

	stats = completeSender.Stats()
	fmt.Printf("\nFinal Stats:\n")
	fmt.Printf("  Sent:    %d\n", stats.Sent)
	fmt.Printf("  Failed:  %d\n", stats.Failed)
	fmt.Printf("  Retried: %d\n", stats.Retried)
	fmt.Printf("  Pending: %d\n", stats.Pending)

	completeSender.Close()

	fmt.Println("\n=== All Examples Complete ===")
}

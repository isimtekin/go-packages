package mailsender

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// EventHandlers contains callback functions for email sending events.
type EventHandlers struct {
	// OnSuccess is called when an email is sent successfully.
	OnSuccess func(message *EmailMessage)

	// OnFailure is called when an email fails to send.
	OnFailure func(message *EmailMessage, err error)

	// OnRetry is called when an email send is being retried.
	OnRetry func(message *EmailMessage, attempt int, err error)
}

// AsyncConfig holds configuration for the async sender.
type AsyncConfig struct {
	// Workers is the number of concurrent workers sending emails.
	Workers int

	// QueueSize is the size of the email queue buffer.
	QueueSize int

	// EventHandlers contains event callback functions.
	EventHandlers *EventHandlers

	// RetryAttempts is the number of retry attempts for failed sends (0 = no retry).
	RetryAttempts int

	// RetryDelay is the delay between retry attempts.
	RetryDelay time.Duration
}

// AsyncSender wraps an EmailSender to provide non-blocking, event-based email sending.
type AsyncSender struct {
	sender    EmailSender
	config    *AsyncConfig
	queue     chan *emailTask
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
	closed    atomic.Bool
	stats     *AsyncStats
	startOnce sync.Once
}

// AsyncStats holds statistics about the async sender.
type AsyncStats struct {
	sent    atomic.Int64
	failed  atomic.Int64
	pending atomic.Int64
	retried atomic.Int64
}

// emailTask represents an email to be sent with its context.
type emailTask struct {
	ctx     context.Context
	message *EmailMessage
	attempt int
}

// NewAsyncSender creates a new async email sender.
func NewAsyncSender(sender EmailSender, opts ...AsyncOption) *AsyncSender {
	config := defaultAsyncConfig()
	for _, opt := range opts {
		opt(config)
	}

	ctx, cancel := context.WithCancel(context.Background())

	as := &AsyncSender{
		sender: sender,
		config: config,
		queue:  make(chan *emailTask, config.QueueSize),
		ctx:    ctx,
		cancel: cancel,
		stats:  &AsyncStats{},
	}

	return as
}

// Start starts the worker pool. This is called automatically on first SendAsync.
func (as *AsyncSender) Start() {
	as.startOnce.Do(func() {
		for i := 0; i < as.config.Workers; i++ {
			as.wg.Add(1)
			go as.worker(i)
		}
	})
}

// worker processes email tasks from the queue.
func (as *AsyncSender) worker(id int) {
	defer as.wg.Done()

	for {
		select {
		case <-as.ctx.Done():
			return
		case task, ok := <-as.queue:
			if !ok {
				return
			}
			as.processTask(task)
		}
	}
}

// processTask sends an email and handles retries and events.
func (as *AsyncSender) processTask(task *emailTask) {
	as.stats.pending.Add(-1)

	err := as.sender.Send(task.ctx, task.message)

	if err != nil {
		// Check if we should retry
		if as.config.RetryAttempts > 0 && task.attempt < as.config.RetryAttempts {
			task.attempt++
			as.stats.retried.Add(1)

			// Call OnRetry handler
			if as.config.EventHandlers != nil && as.config.EventHandlers.OnRetry != nil {
				as.config.EventHandlers.OnRetry(task.message, task.attempt, err)
			}

			// Wait before retry
			if as.config.RetryDelay > 0 {
				select {
				case <-time.After(as.config.RetryDelay):
				case <-as.ctx.Done():
					as.stats.failed.Add(1)
					if as.config.EventHandlers != nil && as.config.EventHandlers.OnFailure != nil {
						as.config.EventHandlers.OnFailure(task.message, err)
					}
					return
				}
			}

			// Re-queue for retry
			// Check if sender is closed before re-queueing
			if as.closed.Load() {
				as.stats.failed.Add(1)
				if as.config.EventHandlers != nil && as.config.EventHandlers.OnFailure != nil {
					as.config.EventHandlers.OnFailure(task.message, err)
				}
				return
			}

			as.stats.pending.Add(1)
			select {
			case as.queue <- task:
			case <-as.ctx.Done():
				as.stats.pending.Add(-1)
				as.stats.failed.Add(1)
				if as.config.EventHandlers != nil && as.config.EventHandlers.OnFailure != nil {
					as.config.EventHandlers.OnFailure(task.message, err)
				}
			}
			return
		}

		// Failed after all retries
		as.stats.failed.Add(1)
		if as.config.EventHandlers != nil && as.config.EventHandlers.OnFailure != nil {
			as.config.EventHandlers.OnFailure(task.message, err)
		}
		return
	}

	// Success
	as.stats.sent.Add(1)
	if as.config.EventHandlers != nil && as.config.EventHandlers.OnSuccess != nil {
		as.config.EventHandlers.OnSuccess(task.message)
	}
}

// SendAsync queues an email for asynchronous sending.
// Returns an error only if the sender is closed or the queue is full (non-blocking).
func (as *AsyncSender) SendAsync(ctx context.Context, message *EmailMessage) error {
	if as.closed.Load() {
		return fmt.Errorf("async sender is closed")
	}

	// Start workers on first send
	as.Start()

	task := &emailTask{
		ctx:     ctx,
		message: message,
		attempt: 0,
	}

	as.stats.pending.Add(1)

	select {
	case as.queue <- task:
		return nil
	default:
		as.stats.pending.Add(-1)
		return fmt.Errorf("queue is full")
	}
}

// SendAsyncBlocking queues an email for asynchronous sending, blocking if queue is full.
func (as *AsyncSender) SendAsyncBlocking(ctx context.Context, message *EmailMessage) error {
	if as.closed.Load() {
		return fmt.Errorf("async sender is closed")
	}

	// Start workers on first send
	as.Start()

	task := &emailTask{
		ctx:     ctx,
		message: message,
		attempt: 0,
	}

	as.stats.pending.Add(1)

	select {
	case as.queue <- task:
		return nil
	case <-ctx.Done():
		as.stats.pending.Add(-1)
		return ctx.Err()
	case <-as.ctx.Done():
		as.stats.pending.Add(-1)
		return fmt.Errorf("async sender is closed")
	}
}

// Close gracefully shuts down the async sender, waiting for all queued emails to be sent.
func (as *AsyncSender) Close() error {
	if !as.closed.CompareAndSwap(false, true) {
		return nil // Already closed
	}

	// Close the queue to prevent new items
	close(as.queue)

	// Wait for all workers to finish
	as.wg.Wait()

	// Cancel context
	as.cancel()

	// Close the underlying sender
	return as.sender.Close()
}

// CloseWithTimeout closes the async sender with a timeout.
// If the timeout is reached, it force-closes and returns the remaining pending count.
func (as *AsyncSender) CloseWithTimeout(timeout time.Duration) error {
	if !as.closed.CompareAndSwap(false, true) {
		return nil // Already closed
	}

	// Close the queue to prevent new items
	close(as.queue)

	// Wait for workers with timeout
	done := make(chan struct{})
	go func() {
		as.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// All workers finished
	case <-time.After(timeout):
		// Timeout reached, force close
		as.cancel()
		return fmt.Errorf("close timeout reached, %d emails may not be sent", as.stats.pending.Load())
	}

	as.cancel()
	return as.sender.Close()
}

// Stats returns the current statistics of the async sender.
func (as *AsyncSender) Stats() AsyncStatsSnapshot {
	return AsyncStatsSnapshot{
		Sent:    as.stats.sent.Load(),
		Failed:  as.stats.failed.Load(),
		Pending: as.stats.pending.Load(),
		Retried: as.stats.retried.Load(),
	}
}

// AsyncStatsSnapshot represents a snapshot of async sender statistics.
type AsyncStatsSnapshot struct {
	Sent    int64
	Failed  int64
	Pending int64
	Retried int64
}

// defaultAsyncConfig returns the default async configuration.
func defaultAsyncConfig() *AsyncConfig {
	return &AsyncConfig{
		Workers:       3,
		QueueSize:     100,
		RetryAttempts: 0,
		RetryDelay:    time.Second,
	}
}

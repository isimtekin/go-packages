package mailsender

import "time"

// AsyncOption is a function that modifies the AsyncConfig.
type AsyncOption func(*AsyncConfig)

// WithWorkers sets the number of concurrent workers.
func WithWorkers(workers int) AsyncOption {
	return func(c *AsyncConfig) {
		if workers > 0 {
			c.Workers = workers
		}
	}
}

// WithQueueSize sets the size of the email queue buffer.
func WithQueueSize(size int) AsyncOption {
	return func(c *AsyncConfig) {
		if size > 0 {
			c.QueueSize = size
		}
	}
}

// WithEventHandlers sets the event handlers for the async sender.
func WithEventHandlers(handlers *EventHandlers) AsyncOption {
	return func(c *AsyncConfig) {
		c.EventHandlers = handlers
	}
}

// WithRetry sets the retry configuration.
func WithRetry(attempts int, delay time.Duration) AsyncOption {
	return func(c *AsyncConfig) {
		c.RetryAttempts = attempts
		c.RetryDelay = delay
	}
}

// WithOnSuccess sets the OnSuccess event handler.
func WithOnSuccess(handler func(*EmailMessage)) AsyncOption {
	return func(c *AsyncConfig) {
		if c.EventHandlers == nil {
			c.EventHandlers = &EventHandlers{}
		}
		c.EventHandlers.OnSuccess = handler
	}
}

// WithOnFailure sets the OnFailure event handler.
func WithOnFailure(handler func(*EmailMessage, error)) AsyncOption {
	return func(c *AsyncConfig) {
		if c.EventHandlers == nil {
			c.EventHandlers = &EventHandlers{}
		}
		c.EventHandlers.OnFailure = handler
	}
}

// WithOnRetry sets the OnRetry event handler.
func WithOnRetry(handler func(*EmailMessage, int, error)) AsyncOption {
	return func(c *AsyncConfig) {
		if c.EventHandlers == nil {
			c.EventHandlers = &EventHandlers{}
		}
		c.EventHandlers.OnRetry = handler
	}
}

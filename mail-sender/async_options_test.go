package mailsender

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWithWorkers(t *testing.T) {
	config := defaultAsyncConfig()
	opt := WithWorkers(10)
	opt(config)
	assert.Equal(t, 10, config.Workers)

	// Test with invalid value (should not change)
	opt = WithWorkers(0)
	opt(config)
	assert.Equal(t, 10, config.Workers) // Should remain unchanged

	opt = WithWorkers(-1)
	opt(config)
	assert.Equal(t, 10, config.Workers) // Should remain unchanged
}

func TestWithQueueSize(t *testing.T) {
	config := defaultAsyncConfig()
	opt := WithQueueSize(500)
	opt(config)
	assert.Equal(t, 500, config.QueueSize)

	// Test with invalid value (should not change)
	opt = WithQueueSize(0)
	opt(config)
	assert.Equal(t, 500, config.QueueSize) // Should remain unchanged

	opt = WithQueueSize(-1)
	opt(config)
	assert.Equal(t, 500, config.QueueSize) // Should remain unchanged
}

func TestWithEventHandlers(t *testing.T) {
	config := defaultAsyncConfig()
	assert.Nil(t, config.EventHandlers)

	var successCalled bool
	var failureCalled bool
	var retryCalled bool

	handlers := &EventHandlers{
		OnSuccess: func(msg *EmailMessage) {
			successCalled = true
		},
		OnFailure: func(msg *EmailMessage, err error) {
			failureCalled = true
		},
		OnRetry: func(msg *EmailMessage, attempt int, err error) {
			retryCalled = true
		},
	}

	opt := WithEventHandlers(handlers)
	opt(config)

	assert.NotNil(t, config.EventHandlers)
	assert.NotNil(t, config.EventHandlers.OnSuccess)
	assert.NotNil(t, config.EventHandlers.OnFailure)
	assert.NotNil(t, config.EventHandlers.OnRetry)

	// Test handlers are callable
	config.EventHandlers.OnSuccess(&EmailMessage{})
	assert.True(t, successCalled)

	config.EventHandlers.OnFailure(&EmailMessage{}, assert.AnError)
	assert.True(t, failureCalled)

	config.EventHandlers.OnRetry(&EmailMessage{}, 1, assert.AnError)
	assert.True(t, retryCalled)
}

func TestWithRetry(t *testing.T) {
	config := defaultAsyncConfig()
	opt := WithRetry(5, 2*time.Second)
	opt(config)
	assert.Equal(t, 5, config.RetryAttempts)
	assert.Equal(t, 2*time.Second, config.RetryDelay)
}

func TestWithOnSuccess(t *testing.T) {
	config := defaultAsyncConfig()
	assert.Nil(t, config.EventHandlers)

	var called bool
	opt := WithOnSuccess(func(msg *EmailMessage) {
		called = true
	})
	opt(config)

	assert.NotNil(t, config.EventHandlers)
	assert.NotNil(t, config.EventHandlers.OnSuccess)

	// Test handler is callable
	config.EventHandlers.OnSuccess(&EmailMessage{})
	assert.True(t, called)
}

func TestWithOnFailure(t *testing.T) {
	config := defaultAsyncConfig()
	assert.Nil(t, config.EventHandlers)

	var called bool
	opt := WithOnFailure(func(msg *EmailMessage, err error) {
		called = true
	})
	opt(config)

	assert.NotNil(t, config.EventHandlers)
	assert.NotNil(t, config.EventHandlers.OnFailure)

	// Test handler is callable
	config.EventHandlers.OnFailure(&EmailMessage{}, assert.AnError)
	assert.True(t, called)
}

func TestWithOnRetry(t *testing.T) {
	config := defaultAsyncConfig()
	assert.Nil(t, config.EventHandlers)

	var called bool
	opt := WithOnRetry(func(msg *EmailMessage, attempt int, err error) {
		called = true
	})
	opt(config)

	assert.NotNil(t, config.EventHandlers)
	assert.NotNil(t, config.EventHandlers.OnRetry)

	// Test handler is callable
	config.EventHandlers.OnRetry(&EmailMessage{}, 1, assert.AnError)
	assert.True(t, called)
}

func TestAsyncOptions_Combined(t *testing.T) {
	config := defaultAsyncConfig()

	var successCalled bool
	var failureCalled bool
	var retryCalled bool

	opts := []AsyncOption{
		WithWorkers(10),
		WithQueueSize(500),
		WithRetry(3, time.Second),
		WithOnSuccess(func(msg *EmailMessage) {
			successCalled = true
		}),
		WithOnFailure(func(msg *EmailMessage, err error) {
			failureCalled = true
		}),
		WithOnRetry(func(msg *EmailMessage, attempt int, err error) {
			retryCalled = true
		}),
	}

	for _, opt := range opts {
		opt(config)
	}

	assert.Equal(t, 10, config.Workers)
	assert.Equal(t, 500, config.QueueSize)
	assert.Equal(t, 3, config.RetryAttempts)
	assert.Equal(t, time.Second, config.RetryDelay)
	assert.NotNil(t, config.EventHandlers)
	assert.NotNil(t, config.EventHandlers.OnSuccess)
	assert.NotNil(t, config.EventHandlers.OnFailure)
	assert.NotNil(t, config.EventHandlers.OnRetry)

	// Test all handlers are callable
	config.EventHandlers.OnSuccess(&EmailMessage{})
	assert.True(t, successCalled)

	config.EventHandlers.OnFailure(&EmailMessage{}, assert.AnError)
	assert.True(t, failureCalled)

	config.EventHandlers.OnRetry(&EmailMessage{}, 1, assert.AnError)
	assert.True(t, retryCalled)
}

func TestAsyncOptions_OverwritingHandlers(t *testing.T) {
	config := defaultAsyncConfig()

	var firstCalled bool
	var secondCalled bool

	// Set first handler
	opt1 := WithOnSuccess(func(msg *EmailMessage) {
		firstCalled = true
	})
	opt1(config)

	// Overwrite with second handler
	opt2 := WithOnSuccess(func(msg *EmailMessage) {
		secondCalled = true
	})
	opt2(config)

	// Only second handler should be called
	config.EventHandlers.OnSuccess(&EmailMessage{})
	assert.False(t, firstCalled)
	assert.True(t, secondCalled)
}

package mailsender

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// mockSender is a mock email sender for testing.
type mockSender struct {
	mu          sync.Mutex
	sent        []*EmailMessage
	sendDelay   time.Duration
	shouldFail  bool
	failCount   int
	currentFail int
}

func newMockSender() *mockSender {
	return &mockSender{
		sent: make([]*EmailMessage, 0),
	}
}

func (m *mockSender) Send(ctx context.Context, message *EmailMessage) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.sendDelay > 0 {
		time.Sleep(m.sendDelay)
	}

	if m.shouldFail {
		if m.failCount == 0 || m.currentFail < m.failCount {
			m.currentFail++
			return errors.New("mock send failed")
		}
	}

	m.sent = append(m.sent, message)
	return nil
}

func (m *mockSender) Close() error {
	return nil
}

func (m *mockSender) getSentCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.sent)
}

func (m *mockSender) getSentMessages() []*EmailMessage {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]*EmailMessage{}, m.sent...)
}

func TestNewAsyncSender(t *testing.T) {
	mock := newMockSender()
	as := NewAsyncSender(mock)

	assert.NotNil(t, as)
	assert.NotNil(t, as.sender)
	assert.NotNil(t, as.queue)
	assert.NotNil(t, as.config)
	assert.Equal(t, 3, as.config.Workers)
	assert.Equal(t, 100, as.config.QueueSize)
}

func TestNewAsyncSender_WithOptions(t *testing.T) {
	mock := newMockSender()
	var successCalled bool
	var failureCalled bool

	as := NewAsyncSender(mock,
		WithWorkers(5),
		WithQueueSize(50),
		WithRetry(3, 10*time.Millisecond),
		WithOnSuccess(func(msg *EmailMessage) {
			successCalled = true
		}),
		WithOnFailure(func(msg *EmailMessage, err error) {
			failureCalled = true
		}),
	)

	assert.Equal(t, 5, as.config.Workers)
	assert.Equal(t, 50, as.config.QueueSize)
	assert.Equal(t, 3, as.config.RetryAttempts)
	assert.Equal(t, 10*time.Millisecond, as.config.RetryDelay)
	assert.NotNil(t, as.config.EventHandlers)
	assert.NotNil(t, as.config.EventHandlers.OnSuccess)
	assert.NotNil(t, as.config.EventHandlers.OnFailure)

	// Test handlers are callable
	as.config.EventHandlers.OnSuccess(&EmailMessage{})
	assert.True(t, successCalled)

	as.config.EventHandlers.OnFailure(&EmailMessage{}, errors.New("test"))
	assert.True(t, failureCalled)
}

func TestAsyncSender_SendAsync(t *testing.T) {
	mock := newMockSender()
	as := NewAsyncSender(mock, WithWorkers(2))
	defer as.Close()

	message := &EmailMessage{
		From:      "sender@example.com",
		To:        []string{"recipient@example.com"},
		Subject:   "Test",
		PlainText: "Test body",
	}

	err := as.SendAsync(context.Background(), message)
	assert.NoError(t, err)

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, 1, mock.getSentCount())
	stats := as.Stats()
	assert.Equal(t, int64(1), stats.Sent)
	assert.Equal(t, int64(0), stats.Failed)
	assert.Equal(t, int64(0), stats.Pending)
}

func TestAsyncSender_SendAsync_Multiple(t *testing.T) {
	mock := newMockSender()
	as := NewAsyncSender(mock, WithWorkers(3))
	defer as.Close()

	count := 10
	for i := 0; i < count; i++ {
		message := &EmailMessage{
			From:      "sender@example.com",
			To:        []string{"recipient@example.com"},
			Subject:   "Test",
			PlainText: "Test body",
		}
		err := as.SendAsync(context.Background(), message)
		assert.NoError(t, err)
	}

	// Wait for processing
	time.Sleep(200 * time.Millisecond)

	assert.Equal(t, count, mock.getSentCount())
	stats := as.Stats()
	assert.Equal(t, int64(count), stats.Sent)
	assert.Equal(t, int64(0), stats.Failed)
}

func TestAsyncSender_EventHandlers(t *testing.T) {
	mock := newMockSender()

	var successMu sync.Mutex
	var successCount int
	var failureMu sync.Mutex
	var failureCount int

	as := NewAsyncSender(mock,
		WithWorkers(2),
		WithOnSuccess(func(msg *EmailMessage) {
			successMu.Lock()
			successCount++
			successMu.Unlock()
		}),
		WithOnFailure(func(msg *EmailMessage, err error) {
			failureMu.Lock()
			failureCount++
			failureMu.Unlock()
		}),
	)
	defer as.Close()

	// Send successful email
	err := as.SendAsync(context.Background(), &EmailMessage{
		From:      "sender@example.com",
		To:        []string{"recipient@example.com"},
		Subject:   "Success Test",
		PlainText: "Test body",
	})
	assert.NoError(t, err)

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	successMu.Lock()
	assert.Equal(t, 1, successCount)
	successMu.Unlock()

	failureMu.Lock()
	assert.Equal(t, 0, failureCount)
	failureMu.Unlock()

	// Test failure
	mock.shouldFail = true
	err = as.SendAsync(context.Background(), &EmailMessage{
		From:      "sender@example.com",
		To:        []string{"recipient@example.com"},
		Subject:   "Failure Test",
		PlainText: "Test body",
	})
	assert.NoError(t, err)

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	failureMu.Lock()
	assert.Equal(t, 1, failureCount)
	failureMu.Unlock()
}

func TestAsyncSender_Retry(t *testing.T) {
	mock := newMockSender()
	mock.shouldFail = true
	mock.failCount = 2 // Fail first 2 attempts, succeed on 3rd

	var retryMu sync.Mutex
	var retryCount int
	var successMu sync.Mutex
	var successCount int

	as := NewAsyncSender(mock,
		WithWorkers(1),
		WithRetry(3, 10*time.Millisecond),
		WithOnRetry(func(msg *EmailMessage, attempt int, err error) {
			retryMu.Lock()
			retryCount++
			retryMu.Unlock()
		}),
		WithOnSuccess(func(msg *EmailMessage) {
			successMu.Lock()
			successCount++
			successMu.Unlock()
		}),
	)
	defer as.Close()

	err := as.SendAsync(context.Background(), &EmailMessage{
		From:      "sender@example.com",
		To:        []string{"recipient@example.com"},
		Subject:   "Retry Test",
		PlainText: "Test body",
	})
	assert.NoError(t, err)

	// Wait for processing with retries
	time.Sleep(500 * time.Millisecond)

	retryMu.Lock()
	assert.Equal(t, 2, retryCount)
	retryMu.Unlock()

	successMu.Lock()
	assert.Equal(t, 1, successCount)
	successMu.Unlock()

	stats := as.Stats()
	assert.Equal(t, int64(1), stats.Sent)
	assert.Equal(t, int64(2), stats.Retried)
	assert.Equal(t, int64(0), stats.Failed)
}

func TestAsyncSender_RetryExhausted(t *testing.T) {
	mock := newMockSender()
	mock.shouldFail = true

	var failureMu sync.Mutex
	var failureCount int

	as := NewAsyncSender(mock,
		WithWorkers(1),
		WithRetry(2, 10*time.Millisecond),
		WithOnFailure(func(msg *EmailMessage, err error) {
			failureMu.Lock()
			failureCount++
			failureMu.Unlock()
		}),
	)
	defer as.Close()

	err := as.SendAsync(context.Background(), &EmailMessage{
		From:      "sender@example.com",
		To:        []string{"recipient@example.com"},
		Subject:   "Retry Exhausted Test",
		PlainText: "Test body",
	})
	assert.NoError(t, err)

	// Wait for processing with retries
	time.Sleep(500 * time.Millisecond)

	failureMu.Lock()
	assert.Equal(t, 1, failureCount)
	failureMu.Unlock()

	stats := as.Stats()
	assert.Equal(t, int64(0), stats.Sent)
	assert.Equal(t, int64(1), stats.Failed)
	assert.Equal(t, int64(2), stats.Retried)
}

func TestAsyncSender_QueueFull(t *testing.T) {
	mock := newMockSender()
	mock.sendDelay = 100 * time.Millisecond // Slow down processing

	as := NewAsyncSender(mock,
		WithWorkers(1),
		WithQueueSize(2),
	)
	defer as.Close()

	// Fill the queue
	err1 := as.SendAsync(context.Background(), &EmailMessage{
		From: "sender@example.com", To: []string{"1@example.com"},
		Subject: "1", PlainText: "1",
	})
	assert.NoError(t, err1)

	err2 := as.SendAsync(context.Background(), &EmailMessage{
		From: "sender@example.com", To: []string{"2@example.com"},
		Subject: "2", PlainText: "2",
	})
	assert.NoError(t, err2)

	// This should fail because queue is full
	err3 := as.SendAsync(context.Background(), &EmailMessage{
		From: "sender@example.com", To: []string{"3@example.com"},
		Subject: "3", PlainText: "3",
	})
	assert.Error(t, err3)
	assert.Contains(t, err3.Error(), "queue is full")
}

func TestAsyncSender_SendAsyncBlocking(t *testing.T) {
	mock := newMockSender()
	as := NewAsyncSender(mock, WithWorkers(2))
	defer as.Close()

	message := &EmailMessage{
		From:      "sender@example.com",
		To:        []string{"recipient@example.com"},
		Subject:   "Blocking Test",
		PlainText: "Test body",
	}

	err := as.SendAsyncBlocking(context.Background(), message)
	assert.NoError(t, err)

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, 1, mock.getSentCount())
}

func TestAsyncSender_Close(t *testing.T) {
	mock := newMockSender()
	as := NewAsyncSender(mock, WithWorkers(2))

	// Send some emails
	for i := 0; i < 5; i++ {
		_ = as.SendAsync(context.Background(), &EmailMessage{
			From: "sender@example.com", To: []string{"recipient@example.com"},
			Subject: "Close Test", PlainText: "Test body",
		})
	}

	// Close should wait for all to be sent
	err := as.Close()
	assert.NoError(t, err)

	// All emails should be sent
	assert.Equal(t, 5, mock.getSentCount())

	// Sending after close should fail
	err = as.SendAsync(context.Background(), &EmailMessage{
		From: "sender@example.com", To: []string{"recipient@example.com"},
		Subject: "After Close", PlainText: "Test body",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "closed")
}

func TestAsyncSender_CloseWithTimeout(t *testing.T) {
	mock := newMockSender()
	mock.sendDelay = 500 * time.Millisecond // Very slow

	as := NewAsyncSender(mock, WithWorkers(1))

	// Send emails
	for i := 0; i < 5; i++ {
		_ = as.SendAsync(context.Background(), &EmailMessage{
			From: "sender@example.com", To: []string{"recipient@example.com"},
			Subject: "Timeout Test", PlainText: "Test body",
		})
	}

	// Close with short timeout
	err := as.CloseWithTimeout(100 * time.Millisecond)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")

	// Not all emails should be sent
	assert.Less(t, mock.getSentCount(), 5)
}

func TestAsyncSender_Stats(t *testing.T) {
	mock := newMockSender()
	as := NewAsyncSender(mock, WithWorkers(2))
	defer as.Close()

	// Initial stats
	stats := as.Stats()
	assert.Equal(t, int64(0), stats.Sent)
	assert.Equal(t, int64(0), stats.Failed)
	assert.Equal(t, int64(0), stats.Pending)

	// Send some emails
	for i := 0; i < 3; i++ {
		_ = as.SendAsync(context.Background(), &EmailMessage{
			From: "sender@example.com", To: []string{"recipient@example.com"},
			Subject: "Stats Test", PlainText: "Test body",
		})
	}

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	stats = as.Stats()
	assert.Equal(t, int64(3), stats.Sent)
	assert.Equal(t, int64(0), stats.Failed)
}

func TestAsyncSender_ConcurrentSends(t *testing.T) {
	mock := newMockSender()
	as := NewAsyncSender(mock, WithWorkers(5), WithQueueSize(100))
	defer as.Close()

	count := 50
	var wg sync.WaitGroup

	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_ = as.SendAsyncBlocking(context.Background(), &EmailMessage{
				From: "sender@example.com", To: []string{"recipient@example.com"},
				Subject: "Concurrent Test", PlainText: "Test body",
			})
		}(i)
	}

	wg.Wait()

	// Wait for processing
	time.Sleep(300 * time.Millisecond)

	assert.Equal(t, count, mock.getSentCount())
	stats := as.Stats()
	assert.Equal(t, int64(count), stats.Sent)
}

func TestAsyncSender_SendAsyncBlocking_ContextCancellation(t *testing.T) {
	mock := newMockSender()
	mock.sendDelay = 500 * time.Millisecond

	as := NewAsyncSender(mock, WithWorkers(1), WithQueueSize(1))
	defer as.Close()

	// Fill the queue
	_ = as.SendAsync(context.Background(), &EmailMessage{
		From: "sender@example.com", To: []string{"recipient@example.com"},
		Subject: "Fill Queue", PlainText: "Test",
	})

	// Try to send with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := as.SendAsyncBlocking(ctx, &EmailMessage{
		From: "sender@example.com", To: []string{"recipient@example.com"},
		Subject: "Cancelled", PlainText: "Test",
	})

	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestAsyncSender_SendAsyncBlocking_Timeout(t *testing.T) {
	mock := newMockSender()
	mock.sendDelay = 200 * time.Millisecond

	as := NewAsyncSender(mock, WithWorkers(1), WithQueueSize(1))
	defer as.Close()

	// Fill the queue and start processing
	_ = as.SendAsync(context.Background(), &EmailMessage{
		From: "sender@example.com", To: []string{"recipient@example.com"},
		Subject: "Fill Queue", PlainText: "Test",
	})

	// Wait a bit to ensure queue is being processed
	time.Sleep(50 * time.Millisecond)

	// Fill the queue again
	_ = as.SendAsync(context.Background(), &EmailMessage{
		From: "sender@example.com", To: []string{"recipient2@example.com"},
		Subject: "Fill Queue 2", PlainText: "Test",
	})

	// Try to send with timeout context (queue should be full)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := as.SendAsyncBlocking(ctx, &EmailMessage{
		From: "sender@example.com", To: []string{"recipient3@example.com"},
		Subject: "Timeout", PlainText: "Test",
	})

	assert.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
}

func TestAsyncSender_MultipleClose(t *testing.T) {
	mock := newMockSender()
	as := NewAsyncSender(mock, WithWorkers(2))

	// First close should succeed
	err := as.Close()
	assert.NoError(t, err)

	// Second close should also succeed (idempotent)
	err = as.Close()
	assert.NoError(t, err)

	// Third close should also succeed (idempotent)
	err = as.Close()
	assert.NoError(t, err)
}

func TestAsyncSender_CloseWithTimeout_Success(t *testing.T) {
	mock := newMockSender()
	as := NewAsyncSender(mock, WithWorkers(2))

	// Send a few emails
	for i := 0; i < 3; i++ {
		_ = as.SendAsync(context.Background(), &EmailMessage{
			From: "sender@example.com", To: []string{"recipient@example.com"},
			Subject: "Test", PlainText: "Test",
		})
	}

	// Close with generous timeout (should succeed)
	err := as.CloseWithTimeout(2 * time.Second)
	assert.NoError(t, err)

	// All emails should be sent
	assert.Equal(t, 3, mock.getSentCount())
}

func TestAsyncSender_RetryWithContextCancellation(t *testing.T) {
	mock := newMockSender()
	mock.shouldFail = true

	var failureCalled bool
	var failureMu sync.Mutex

	as := NewAsyncSender(mock,
		WithWorkers(1),
		WithRetry(5, 100*time.Millisecond),
		WithOnFailure(func(msg *EmailMessage, err error) {
			failureMu.Lock()
			failureCalled = true
			failureMu.Unlock()
		}),
	)

	// Start sending
	_ = as.SendAsync(context.Background(), &EmailMessage{
		From: "sender@example.com", To: []string{"recipient@example.com"},
		Subject: "Test", PlainText: "Test",
	})

	// Close immediately (cancels context)
	time.Sleep(50 * time.Millisecond)
	as.Close()

	// Wait a bit
	time.Sleep(100 * time.Millisecond)

	// Should have called OnFailure
	failureMu.Lock()
	assert.True(t, failureCalled)
	failureMu.Unlock()
}

func TestAsyncSender_Start_MultipleCallsIdempotent(t *testing.T) {
	mock := newMockSender()
	as := NewAsyncSender(mock, WithWorkers(3))
	defer as.Close()

	// Call Start multiple times
	as.Start()
	as.Start()
	as.Start()

	// Send an email to verify it works
	err := as.SendAsync(context.Background(), &EmailMessage{
		From: "sender@example.com", To: []string{"recipient@example.com"},
		Subject: "Test", PlainText: "Test",
	})
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 1, mock.getSentCount())
}

func TestAsyncSender_DefaultAsyncConfig(t *testing.T) {
	config := defaultAsyncConfig()
	assert.Equal(t, 3, config.Workers)
	assert.Equal(t, 100, config.QueueSize)
	assert.Equal(t, 0, config.RetryAttempts)
	assert.Equal(t, time.Second, config.RetryDelay)
	assert.Nil(t, config.EventHandlers)
}

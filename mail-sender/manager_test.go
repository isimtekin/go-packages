package mailsender

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// managerMockSender is a mock implementation of EmailSender for testing SenderManager.
type managerMockSender struct {
	sendFunc  func(ctx context.Context, message *EmailMessage) error
	closeFunc func() error
	closed    bool
	mu        sync.Mutex
}

func newManagerMockSender() *managerMockSender {
	return &managerMockSender{}
}

func (m *managerMockSender) Send(ctx context.Context, message *EmailMessage) error {
	if m.sendFunc != nil {
		return m.sendFunc(ctx, message)
	}
	return nil
}

func (m *managerMockSender) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

func (m *managerMockSender) isClosed() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.closed
}

func TestNewSenderManager(t *testing.T) {
	manager := NewSenderManager()
	assert.NotNil(t, manager)
	assert.Equal(t, 0, manager.Count())
	assert.Equal(t, "", manager.DefaultSenderName())
}

func TestSenderManager_Register(t *testing.T) {
	tests := []struct {
		name        string
		senderName  string
		sender      EmailSender
		wantErr     bool
		errContains string
	}{
		{
			name:       "valid registration",
			senderName: "test",
			sender:     newManagerMockSender(),
			wantErr:    false,
		},
		{
			name:        "empty name",
			senderName:  "",
			sender:      newManagerMockSender(),
			wantErr:     true,
			errContains: "cannot be empty",
		},
		{
			name:        "nil sender",
			senderName:  "test",
			sender:      nil,
			wantErr:     true,
			errContains: "cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewSenderManager()
			err := manager.Register(tt.senderName, tt.sender)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
				assert.Equal(t, 1, manager.Count())
			}
		})
	}
}

func TestSenderManager_Register_Duplicate(t *testing.T) {
	manager := NewSenderManager()

	err := manager.Register("test", newManagerMockSender())
	require.NoError(t, err)

	err = manager.Register("test", newManagerMockSender())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}

func TestSenderManager_Register_FirstBecomesDefault(t *testing.T) {
	manager := NewSenderManager()

	err := manager.Register("first", newManagerMockSender())
	require.NoError(t, err)
	assert.Equal(t, "first", manager.DefaultSenderName())

	err = manager.Register("second", newManagerMockSender())
	require.NoError(t, err)
	assert.Equal(t, "first", manager.DefaultSenderName())
}

func TestSenderManager_Unregister(t *testing.T) {
	manager := NewSenderManager()
	err := manager.Register("test", newManagerMockSender())
	require.NoError(t, err)

	err = manager.Unregister("test")
	require.NoError(t, err)
	assert.Equal(t, 0, manager.Count())
}

func TestSenderManager_Unregister_NotFound(t *testing.T) {
	manager := NewSenderManager()

	err := manager.Unregister("nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestSenderManager_Unregister_DefaultSender(t *testing.T) {
	manager := NewSenderManager()

	err := manager.Register("first", newManagerMockSender())
	require.NoError(t, err)
	err = manager.Register("second", newManagerMockSender())
	require.NoError(t, err)

	// Remove the default sender
	err = manager.Unregister("first")
	require.NoError(t, err)

	// Default should now be "second"
	assert.Equal(t, "second", manager.DefaultSenderName())
}

func TestSenderManager_Unregister_LastSender(t *testing.T) {
	manager := NewSenderManager()

	err := manager.Register("only", newManagerMockSender())
	require.NoError(t, err)

	err = manager.Unregister("only")
	require.NoError(t, err)

	assert.Equal(t, "", manager.DefaultSenderName())
}

func TestSenderManager_SetDefault(t *testing.T) {
	manager := NewSenderManager()
	err := manager.Register("first", newManagerMockSender())
	require.NoError(t, err)
	err = manager.Register("second", newManagerMockSender())
	require.NoError(t, err)

	err = manager.SetDefault("second")
	require.NoError(t, err)
	assert.Equal(t, "second", manager.DefaultSenderName())
}

func TestSenderManager_SetDefault_NotFound(t *testing.T) {
	manager := NewSenderManager()

	err := manager.SetDefault("nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestSenderManager_SetDefaultFrom(t *testing.T) {
	manager := NewSenderManager()
	manager.SetDefaultFrom("test@example.com", "Test User")

	// Verify defaults are applied when sending
	var capturedMessage *EmailMessage
	mock := newManagerMockSender()
	mock.sendFunc = func(ctx context.Context, msg *EmailMessage) error {
		capturedMessage = msg
		return nil
	}

	err := manager.Register("test", mock)
	require.NoError(t, err)

	err = manager.Send(context.Background(), "test", &EmailMessage{
		To:        []string{"to@example.com"},
		Subject:   "Test",
		PlainText: "Hello",
	})
	require.NoError(t, err)

	assert.Equal(t, "test@example.com", capturedMessage.From)
	assert.Equal(t, "Test User", capturedMessage.FromName)
}

func TestSenderManager_GetSender(t *testing.T) {
	manager := NewSenderManager()
	mock := newManagerMockSender()
	err := manager.Register("test", mock)
	require.NoError(t, err)

	sender, err := manager.GetSender("test")
	require.NoError(t, err)
	assert.Equal(t, mock, sender)
}

func TestSenderManager_GetSender_NotFound(t *testing.T) {
	manager := NewSenderManager()

	_, err := manager.GetSender("nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestSenderManager_GetDefault(t *testing.T) {
	manager := NewSenderManager()
	mock := newManagerMockSender()
	err := manager.Register("test", mock)
	require.NoError(t, err)

	sender, err := manager.GetDefault()
	require.NoError(t, err)
	assert.Equal(t, mock, sender)
}

func TestSenderManager_GetDefault_NoDefault(t *testing.T) {
	manager := NewSenderManager()

	_, err := manager.GetDefault()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no default sender")
}

func TestSenderManager_Send(t *testing.T) {
	manager := NewSenderManager()

	sendCalled := false
	mock := newManagerMockSender()
	mock.sendFunc = func(ctx context.Context, msg *EmailMessage) error {
		sendCalled = true
		return nil
	}

	err := manager.Register("test", mock)
	require.NoError(t, err)

	err = manager.Send(context.Background(), "test", &EmailMessage{
		From:      "from@example.com",
		To:        []string{"to@example.com"},
		Subject:   "Test",
		PlainText: "Hello",
	})
	require.NoError(t, err)
	assert.True(t, sendCalled)
}

func TestSenderManager_Send_NotFound(t *testing.T) {
	manager := NewSenderManager()

	err := manager.Send(context.Background(), "nonexistent", &EmailMessage{
		From:      "from@example.com",
		To:        []string{"to@example.com"},
		Subject:   "Test",
		PlainText: "Hello",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestSenderManager_Send_Error(t *testing.T) {
	manager := NewSenderManager()

	expectedErr := errors.New("send failed")
	mock := newManagerMockSender()
	mock.sendFunc = func(ctx context.Context, msg *EmailMessage) error {
		return expectedErr
	}

	err := manager.Register("test", mock)
	require.NoError(t, err)

	err = manager.Send(context.Background(), "test", &EmailMessage{
		From:      "from@example.com",
		To:        []string{"to@example.com"},
		Subject:   "Test",
		PlainText: "Hello",
	})
	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestSenderManager_SendDefault(t *testing.T) {
	manager := NewSenderManager()

	sendCalled := false
	mock := newManagerMockSender()
	mock.sendFunc = func(ctx context.Context, msg *EmailMessage) error {
		sendCalled = true
		return nil
	}

	err := manager.Register("test", mock)
	require.NoError(t, err)

	err = manager.SendDefault(context.Background(), &EmailMessage{
		From:      "from@example.com",
		To:        []string{"to@example.com"},
		Subject:   "Test",
		PlainText: "Hello",
	})
	require.NoError(t, err)
	assert.True(t, sendCalled)
}

func TestSenderManager_SendDefault_NoDefault(t *testing.T) {
	manager := NewSenderManager()

	err := manager.SendDefault(context.Background(), &EmailMessage{
		From:      "from@example.com",
		To:        []string{"to@example.com"},
		Subject:   "Test",
		PlainText: "Hello",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no default sender")
}

func TestSenderManager_List(t *testing.T) {
	manager := NewSenderManager()

	err := manager.Register("first", newManagerMockSender())
	require.NoError(t, err)
	err = manager.Register("second", newManagerMockSender())
	require.NoError(t, err)

	list := manager.List()
	assert.Len(t, list, 2)
	assert.Contains(t, list, "first")
	assert.Contains(t, list, "second")
}

func TestSenderManager_Close(t *testing.T) {
	manager := NewSenderManager()

	mock1 := newManagerMockSender()
	mock2 := newManagerMockSender()

	err := manager.Register("first", mock1)
	require.NoError(t, err)
	err = manager.Register("second", mock2)
	require.NoError(t, err)

	err = manager.Close()
	require.NoError(t, err)

	assert.True(t, mock1.isClosed())
	assert.True(t, mock2.isClosed())
	assert.Equal(t, 0, manager.Count())
	assert.Equal(t, "", manager.DefaultSenderName())
}

func TestSenderManager_Close_WithError(t *testing.T) {
	manager := NewSenderManager()

	mock := newManagerMockSender()
	mock.closeFunc = func() error {
		return errors.New("close error")
	}

	err := manager.Register("test", mock)
	require.NoError(t, err)

	err = manager.Close()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "close error")
}

func TestSenderManager_ConcurrentAccess(t *testing.T) {
	manager := NewSenderManager()

	// Pre-register a sender
	err := manager.Register("shared", newManagerMockSender())
	require.NoError(t, err)

	var wg sync.WaitGroup
	numGoroutines := 100

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = manager.GetSender("shared")
			_ = manager.List()
			_ = manager.Count()
			_ = manager.DefaultSenderName()
		}()
	}

	wg.Wait()
}

func TestSenderManager_MultipleSenders(t *testing.T) {
	manager := NewSenderManager()

	// Create multiple mock senders with different behaviors
	sendgridCalled := false
	smtpCalled := false
	sesCalled := false

	sendgridMock := newManagerMockSender()
	sendgridMock.sendFunc = func(ctx context.Context, msg *EmailMessage) error {
		sendgridCalled = true
		return nil
	}

	smtpMock := newManagerMockSender()
	smtpMock.sendFunc = func(ctx context.Context, msg *EmailMessage) error {
		smtpCalled = true
		return nil
	}

	sesMock := newManagerMockSender()
	sesMock.sendFunc = func(ctx context.Context, msg *EmailMessage) error {
		sesCalled = true
		return nil
	}

	// Register all senders
	err := manager.Register("sendgrid", sendgridMock)
	require.NoError(t, err)
	err = manager.Register("smtp", smtpMock)
	require.NoError(t, err)
	err = manager.Register("ses", sesMock)
	require.NoError(t, err)

	msg := &EmailMessage{
		From:      "from@example.com",
		To:        []string{"to@example.com"},
		Subject:   "Test",
		PlainText: "Hello",
	}

	// Send via each provider
	err = manager.Send(context.Background(), "sendgrid", msg)
	require.NoError(t, err)
	assert.True(t, sendgridCalled)

	err = manager.Send(context.Background(), "smtp", msg)
	require.NoError(t, err)
	assert.True(t, smtpCalled)

	err = manager.Send(context.Background(), "ses", msg)
	require.NoError(t, err)
	assert.True(t, sesCalled)
}

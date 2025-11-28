package mailsender

import (
	"context"
	"fmt"
	"sync"
)

// SenderManager manages multiple email providers and allows sending through a specific provider.
type SenderManager struct {
	mu             sync.RWMutex
	senders        map[string]EmailSender
	defaultSender  string
	defaultFrom    string
	defaultFromName string
}

// NewSenderManager creates a new sender manager.
func NewSenderManager() *SenderManager {
	return &SenderManager{
		senders: make(map[string]EmailSender),
	}
}

// Register registers an email sender with a name.
// If this is the first sender registered, it becomes the default.
func (m *SenderManager) Register(name string, sender EmailSender) error {
	if name == "" {
		return fmt.Errorf("sender name cannot be empty")
	}
	if sender == nil {
		return fmt.Errorf("sender cannot be nil")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.senders[name]; exists {
		return fmt.Errorf("sender %q already registered", name)
	}

	m.senders[name] = sender

	// Set as default if this is the first sender
	if m.defaultSender == "" {
		m.defaultSender = name
	}

	return nil
}

// Unregister removes a sender from the manager.
func (m *SenderManager) Unregister(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.senders[name]; !exists {
		return fmt.Errorf("sender %q not found", name)
	}

	delete(m.senders, name)

	// If the default sender was removed, clear it
	if m.defaultSender == name {
		m.defaultSender = ""
		// Set another sender as default if available
		for n := range m.senders {
			m.defaultSender = n
			break
		}
	}

	return nil
}

// SetDefault sets the default sender by name.
func (m *SenderManager) SetDefault(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.senders[name]; !exists {
		return fmt.Errorf("sender %q not found", name)
	}

	m.defaultSender = name
	return nil
}

// SetDefaultFrom sets the default from address for all senders.
func (m *SenderManager) SetDefaultFrom(from, fromName string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.defaultFrom = from
	m.defaultFromName = fromName
}

// GetSender returns a sender by name.
func (m *SenderManager) GetSender(name string) (EmailSender, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sender, exists := m.senders[name]
	if !exists {
		return nil, fmt.Errorf("sender %q not found", name)
	}

	return sender, nil
}

// GetDefault returns the default sender.
func (m *SenderManager) GetDefault() (EmailSender, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.defaultSender == "" {
		return nil, fmt.Errorf("no default sender configured")
	}

	return m.senders[m.defaultSender], nil
}

// Send sends an email using the specified sender.
func (m *SenderManager) Send(ctx context.Context, senderName string, message *EmailMessage) error {
	m.mu.RLock()
	sender, exists := m.senders[senderName]
	defaultFrom := m.defaultFrom
	defaultFromName := m.defaultFromName
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("sender %q not found", senderName)
	}

	// Apply defaults
	if message.From == "" {
		message.From = defaultFrom
	}
	if message.FromName == "" {
		message.FromName = defaultFromName
	}

	return sender.Send(ctx, message)
}

// SendDefault sends an email using the default sender.
func (m *SenderManager) SendDefault(ctx context.Context, message *EmailMessage) error {
	m.mu.RLock()
	defaultSender := m.defaultSender
	defaultFrom := m.defaultFrom
	defaultFromName := m.defaultFromName
	m.mu.RUnlock()

	if defaultSender == "" {
		return fmt.Errorf("no default sender configured")
	}

	// Apply defaults
	if message.From == "" {
		message.From = defaultFrom
	}
	if message.FromName == "" {
		message.FromName = defaultFromName
	}

	return m.Send(ctx, defaultSender, message)
}

// List returns a list of all registered sender names.
func (m *SenderManager) List() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.senders))
	for name := range m.senders {
		names = append(names, name)
	}
	return names
}

// Close closes all registered senders.
func (m *SenderManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error
	for name, sender := range m.senders {
		if err := sender.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close sender %q: %w", name, err))
		}
	}

	m.senders = make(map[string]EmailSender)
	m.defaultSender = ""

	if len(errs) > 0 {
		return fmt.Errorf("errors closing senders: %v", errs)
	}

	return nil
}

// Count returns the number of registered senders.
func (m *SenderManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.senders)
}

// DefaultSenderName returns the name of the default sender.
func (m *SenderManager) DefaultSenderName() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.defaultSender
}

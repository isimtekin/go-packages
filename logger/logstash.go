package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"
)

// LogstashHook is a hook for sending logs to Logstash via TCP/UDP
type LogstashHook struct {
	config       *Config
	conn         net.Conn
	mu           sync.Mutex
	buffer       [][]byte
	maxBuffer    int
	reconnecting bool
}

// LogstashWriter implements io.Writer for Logstash output
type LogstashWriter struct {
	hook *LogstashHook
}

// Write implements io.Writer
func (w *LogstashWriter) Write(p []byte) (n int, err error) {
	return w.hook.Write(p)
}

// setupLogstashOutput configures Logstash as the output destination
func (l *Logger) setupLogstashOutput() error {
	hook, err := NewLogstashHook(l.config)
	if err != nil {
		return fmt.Errorf("failed to create logstash hook: %w", err)
	}

	writer := &LogstashWriter{hook: hook}
	l.logger.SetOutput(writer)

	return nil
}

// NewLogstashHook creates a new Logstash hook
func NewLogstashHook(config *Config) (*LogstashHook, error) {
	hook := &LogstashHook{
		config:    config,
		buffer:    make([][]byte, 0),
		maxBuffer: 1000, // Buffer up to 1000 messages
	}

	// Establish initial connection
	if err := hook.connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to logstash: %w", err)
	}

	return hook, nil
}

// connect establishes a connection to Logstash
func (h *LogstashHook) connect() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	address := fmt.Sprintf("%s:%d", h.config.LogstashHost, h.config.LogstashPort)

	var conn net.Conn
	var err error

	for attempt := 0; attempt < h.config.LogstashRetries; attempt++ {
		if h.config.LogstashProtocol == "tcp" {
			conn, err = net.DialTimeout("tcp", address, h.config.LogstashTimeout)
		} else {
			conn, err = net.DialTimeout("udp", address, h.config.LogstashTimeout)
		}

		if err == nil {
			h.conn = conn
			h.reconnecting = false
			return nil
		}

		// Wait before retrying
		time.Sleep(time.Second * time.Duration(attempt+1))
	}

	return fmt.Errorf("failed to connect after %d attempts: %w", h.config.LogstashRetries, err)
}

// Write sends log data to Logstash
func (h *LogstashHook) Write(p []byte) (n int, err error) {
	// Parse the log entry to add @version and @timestamp if not present
	var logEntry map[string]interface{}
	if err := json.Unmarshal(p, &logEntry); err == nil {
		// Add Logstash-specific fields
		if _, ok := logEntry["@version"]; !ok {
			logEntry["@version"] = "1"
		}
		if timestamp, ok := logEntry["@timestamp"]; ok {
			// Keep existing timestamp
			logEntry["@timestamp"] = timestamp
		} else {
			logEntry["@timestamp"] = time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
		}

		// Re-marshal with added fields
		p, err = json.Marshal(logEntry)
		if err != nil {
			return 0, err
		}
	}

	// Add newline if not present
	if len(p) > 0 && p[len(p)-1] != '\n' {
		p = append(p, '\n')
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	// Try to send
	if h.conn != nil {
		if h.config.LogstashProtocol == "tcp" {
			// Set write deadline for TCP
			h.conn.SetWriteDeadline(time.Now().Add(h.config.LogstashTimeout))
		}

		n, err = h.conn.Write(p)
		if err == nil {
			return n, nil
		}

		// Connection error, try to reconnect
		h.conn.Close()
		h.conn = nil
	}

	// Buffer the message if we can't send it
	if len(h.buffer) < h.maxBuffer {
		msgCopy := make([]byte, len(p))
		copy(msgCopy, p)
		h.buffer = append(h.buffer, msgCopy)
	}

	// Try to reconnect in background
	if !h.reconnecting {
		h.reconnecting = true
		go h.reconnect()
	}

	return len(p), nil
}

// reconnect attempts to reconnect to Logstash and flush buffer
func (h *LogstashHook) reconnect() {
	for {
		time.Sleep(5 * time.Second)

		if err := h.connect(); err != nil {
			continue
		}

		// Connection successful, flush buffer
		h.mu.Lock()
		buffered := h.buffer
		h.buffer = make([][]byte, 0)
		h.mu.Unlock()

		// Send buffered messages
		for _, msg := range buffered {
			h.conn.Write(msg)
		}

		return
	}
}

// Close closes the Logstash connection
func (h *LogstashHook) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.conn != nil {
		return h.conn.Close()
	}
	return nil
}

// setupFileOutput configures file as the output destination
func (l *Logger) setupFileOutput() error {
	file, err := openLogFile(l.config.FilePath)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	l.logger.SetOutput(file)
	return nil
}

// openLogFile opens or creates a log file
func openLogFile(path string) (io.Writer, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	return file, nil
}

// setupMultiOutput configures multiple outputs
func (l *Logger) setupMultiOutput() error {
	writers := make([]io.Writer, 0)

	for _, output := range l.config.MultiOutputs {
		if !output.Enabled {
			continue
		}

		switch output.Type {
		case "console":
			writers = append(writers, os.Stdout)
		case "file":
			if output.FilePath != "" {
				file, err := openLogFile(output.FilePath)
				if err != nil {
					return fmt.Errorf("failed to open file %s: %w", output.FilePath, err)
				}
				writers = append(writers, file)
			}
		case "logstash":
			if output.Host != "" && output.Port > 0 {
				// Create a temporary config for this logstash output
				logstashConfig := *l.config
				logstashConfig.LogstashHost = output.Host
				logstashConfig.LogstashPort = output.Port

				hook, err := NewLogstashHook(&logstashConfig)
				if err != nil {
					return fmt.Errorf("failed to create logstash hook for %s:%d: %w", output.Host, output.Port, err)
				}
				writers = append(writers, &LogstashWriter{hook: hook})
			}
		}
	}

	if len(writers) == 0 {
		return fmt.Errorf("no valid outputs configured for multi-output mode")
	}

	// Use MultiWriter to write to all outputs
	multiWriter := io.MultiWriter(writers...)
	l.logger.SetOutput(multiWriter)

	return nil
}

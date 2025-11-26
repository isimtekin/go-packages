package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	t.Run("creates logger with default config", func(t *testing.T) {
		config := DefaultConfig()
		logger, err := New(config)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if logger == nil {
			t.Fatal("expected logger, got nil")
		}
		if !logger.IsEnabled() {
			t.Error("expected logger to be enabled")
		}
	})

	t.Run("validates invalid config", func(t *testing.T) {
		config := DefaultConfig()
		config.Output = "invalid"
		_, err := New(config)
		if err == nil {
			t.Error("expected validation error")
		}
	})

	t.Run("creates logger with JSON format", func(t *testing.T) {
		config := DefaultConfig()
		config.Format = "json"
		logger, err := New(config)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if logger == nil {
			t.Fatal("expected logger, got nil")
		}
	})
}

func TestLogger_Disabled(t *testing.T) {
	t.Run("disabled logger does not log", func(t *testing.T) {
		config := DefaultConfig()
		config.Enabled = false

		logger, err := New(config)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if logger.IsEnabled() {
			t.Error("expected logger to be disabled")
		}

		// These should not panic or log anything
		logger.Info("test message")
		logger.Debug("debug message")
		logger.Error("error message")
	})
}

func TestLogger_Levels(t *testing.T) {
	tests := []struct {
		name  string
		level string
	}{
		{"trace level", "trace"},
		{"debug level", "debug"},
		{"info level", "info"},
		{"warn level", "warn"},
		{"error level", "error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			config.Level = tt.level
			logger, err := New(config)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if logger == nil {
				t.Fatal("expected logger, got nil")
			}
		})
	}
}

func TestLogger_Fields(t *testing.T) {
	t.Run("adds fields to logger", func(t *testing.T) {
		config := DefaultConfig()
		logger, err := New(config)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		newLogger := logger.WithFields(
			String("key1", "value1"),
			Int("key2", 42),
		)

		if len(newLogger.fields) < 2 {
			t.Error("expected fields to be added")
		}
	})

	t.Run("WithField adds single field", func(t *testing.T) {
		config := DefaultConfig()
		logger, err := New(config)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		newLogger := logger.WithField("test", "value")
		if len(newLogger.fields) == 0 {
			t.Error("expected field to be added")
		}
	})

	t.Run("WithContext extracts context values", func(t *testing.T) {
		config := DefaultConfig()
		logger, err := New(config)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		ctx := context.WithValue(context.Background(), "request_id", "test-123")
		newLogger := logger.WithContext(ctx)

		if _, ok := newLogger.fields["request_id"]; !ok {
			t.Error("expected request_id field to be added")
		}
	})
}

func TestLogger_IsDebugEnabled(t *testing.T) {
	t.Run("debug enabled when level is debug", func(t *testing.T) {
		config := DefaultConfig()
		config.Level = "debug"
		logger, err := New(config)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if !logger.IsDebugEnabled() {
			t.Error("expected debug to be enabled")
		}
	})

	t.Run("debug disabled when level is info", func(t *testing.T) {
		config := DefaultConfig()
		config.Level = "info"
		logger, err := New(config)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if logger.IsDebugEnabled() {
			t.Error("expected debug to be disabled")
		}
	})
}

func TestFieldHelpers(t *testing.T) {
	t.Run("String field", func(t *testing.T) {
		field := String("key", "value")
		if field.Key != "key" || field.Value != "value" {
			t.Error("String field not created correctly")
		}
	})

	t.Run("Int field", func(t *testing.T) {
		field := Int("count", 42)
		if field.Key != "count" || field.Value != 42 {
			t.Error("Int field not created correctly")
		}
	})

	t.Run("Int64 field", func(t *testing.T) {
		field := Int64("count", int64(42))
		if field.Key != "count" || field.Value != int64(42) {
			t.Error("Int64 field not created correctly")
		}
	})

	t.Run("Float64 field", func(t *testing.T) {
		field := Float64("value", 3.14)
		if field.Key != "value" || field.Value != 3.14 {
			t.Error("Float64 field not created correctly")
		}
	})

	t.Run("Bool field", func(t *testing.T) {
		field := Bool("flag", true)
		if field.Key != "flag" || field.Value != true {
			t.Error("Bool field not created correctly")
		}
	})

	t.Run("Err field", func(t *testing.T) {
		err := io.EOF
		field := Err(err)
		if field.Key != "error" || field.Value != err.Error() {
			t.Error("Err field not created correctly")
		}
	})

	t.Run("Any field", func(t *testing.T) {
		value := map[string]string{"key": "value"}
		field := Any("data", value)
		if field.Key != "data" {
			t.Error("Any field not created correctly")
		}
	})
}

func TestGlobalLogger(t *testing.T) {
	t.Run("SetGlobal and GetGlobal", func(t *testing.T) {
		config := DefaultConfig()
		config.ServiceName = "test-service"
		logger, err := New(config)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		SetGlobal(logger)
		global := GetGlobal()

		if global.config.ServiceName != "test-service" {
			t.Error("global logger not set correctly")
		}
	})

	t.Run("global logger functions work", func(t *testing.T) {
		config := DefaultConfig()
		config.Enabled = false // Disable to avoid output
		logger, _ := New(config)
		SetGlobal(logger)

		// These should not panic
		Info("test")
		Debug("test")
		Warn("test")
		Error("test")
		Trace("test")
	})
}

func TestLogger_OutputFormats(t *testing.T) {
	t.Run("JSON format output", func(t *testing.T) {
		var buf bytes.Buffer

		config := DefaultConfig()
		config.Format = "json"
		logger, err := New(config)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Redirect output to buffer
		logger.logger.SetOutput(&buf)

		logger.Info("test message", String("key", "value"))

		output := buf.String()
		if output == "" {
			t.Error("expected output, got empty string")
		}

		// Verify it's valid JSON
		var jsonOutput map[string]interface{}
		if err := json.Unmarshal([]byte(output), &jsonOutput); err != nil {
			t.Errorf("expected valid JSON, got error: %v", err)
		}
	})

	t.Run("Text format output", func(t *testing.T) {
		var buf bytes.Buffer

		config := DefaultConfig()
		config.Format = "text"
		logger, err := New(config)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		logger.logger.SetOutput(&buf)
		logger.Info("test message")

		output := buf.String()
		if !strings.Contains(output, "test message") {
			t.Error("expected message in output")
		}
	})
}

func TestLogger_Metadata(t *testing.T) {
	t.Run("adds service metadata", func(t *testing.T) {
		config := DefaultConfig()
		config.ServiceName = "my-service"
		config.Environment = "production"
		config.Version = "1.0.0"

		logger, err := New(config)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if logger.fields["service"] != "my-service" {
			t.Error("service name not added to fields")
		}
		if logger.fields["environment"] != "production" {
			t.Error("environment not added to fields")
		}
		if logger.fields["version"] != "1.0.0" {
			t.Error("version not added to fields")
		}
	})

	t.Run("adds hostname", func(t *testing.T) {
		config := DefaultConfig()
		logger, err := New(config)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if _, ok := logger.fields["host"]; !ok {
			t.Error("hostname not added to fields")
		}
	})
}

func TestConfig_Validate(t *testing.T) {
	t.Run("validates output type", func(t *testing.T) {
		config := DefaultConfig()
		config.Output = "invalid"
		if err := config.Validate(); err == nil {
			t.Error("expected validation error for invalid output")
		}
	})

	t.Run("validates log level", func(t *testing.T) {
		config := DefaultConfig()
		config.Level = "invalid"
		if err := config.Validate(); err == nil {
			t.Error("expected validation error for invalid level")
		}
	})

	t.Run("validates format", func(t *testing.T) {
		config := DefaultConfig()
		config.Format = "invalid"
		if err := config.Validate(); err == nil {
			t.Error("expected validation error for invalid format")
		}
	})

	t.Run("validates logstash config", func(t *testing.T) {
		config := DefaultConfig()
		config.Output = "logstash"
		config.LogstashHost = ""
		if err := config.Validate(); err == nil {
			t.Error("expected validation error for missing logstash host")
		}
	})

	t.Run("validates file config", func(t *testing.T) {
		config := DefaultConfig()
		config.Output = "file"
		config.FilePath = ""
		if err := config.Validate(); err == nil {
			t.Error("expected validation error for missing file path")
		}
	})

	t.Run("validates multi config", func(t *testing.T) {
		config := DefaultConfig()
		config.Output = "multi"
		config.MultiOutputs = []OutputConfig{}
		if err := config.Validate(); err == nil {
			t.Error("expected validation error for empty multi outputs")
		}
	})
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if !config.Enabled {
		t.Error("expected Enabled to be true")
	}
	if config.Level != "info" {
		t.Errorf("expected Level to be 'info', got %s", config.Level)
	}
	if config.Output != "console" {
		t.Errorf("expected Output to be 'console', got %s", config.Output)
	}
	if config.Format != "text" {
		t.Errorf("expected Format to be 'text', got %s", config.Format)
	}
	if !config.ConsoleColors {
		t.Error("expected ConsoleColors to be true")
	}
	if config.LogstashPort != 5000 {
		t.Errorf("expected LogstashPort to be 5000, got %d", config.LogstashPort)
	}
}

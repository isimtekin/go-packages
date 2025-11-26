package logger

import (
	"testing"
	"time"
)

func TestOptions(t *testing.T) {
	t.Run("WithEnabled", func(t *testing.T) {
		config := DefaultConfig()
		WithEnabled(false)(config)
		if config.Enabled {
			t.Error("expected Enabled to be false")
		}
	})

	t.Run("WithLevel", func(t *testing.T) {
		config := DefaultConfig()
		WithLevel("debug")(config)
		if config.Level != "debug" {
			t.Errorf("expected Level 'debug', got %s", config.Level)
		}
	})

	t.Run("WithOutput", func(t *testing.T) {
		config := DefaultConfig()
		WithOutput("file")(config)
		if config.Output != "file" {
			t.Errorf("expected Output 'file', got %s", config.Output)
		}
	})

	t.Run("WithFormat", func(t *testing.T) {
		config := DefaultConfig()
		WithFormat("json")(config)
		if config.Format != "json" {
			t.Errorf("expected Format 'json', got %s", config.Format)
		}
	})

	t.Run("WithConsoleColors", func(t *testing.T) {
		config := DefaultConfig()
		WithConsoleColors(false)(config)
		if config.ConsoleColors {
			t.Error("expected ConsoleColors to be false")
		}
	})

	t.Run("WithLogstash", func(t *testing.T) {
		config := DefaultConfig()
		WithLogstash("logstash.example.com", 5044, "tcp")(config)
		if config.LogstashHost != "logstash.example.com" {
			t.Errorf("expected LogstashHost 'logstash.example.com', got %s", config.LogstashHost)
		}
		if config.LogstashPort != 5044 {
			t.Errorf("expected LogstashPort 5044, got %d", config.LogstashPort)
		}
		if config.LogstashProtocol != "tcp" {
			t.Errorf("expected LogstashProtocol 'tcp', got %s", config.LogstashProtocol)
		}
	})

	t.Run("WithLogstashTimeout", func(t *testing.T) {
		config := DefaultConfig()
		timeout := 10 * time.Second
		WithLogstashTimeout(timeout)(config)
		if config.LogstashTimeout != timeout {
			t.Errorf("expected LogstashTimeout %v, got %v", timeout, config.LogstashTimeout)
		}
	})

	t.Run("WithLogstashRetries", func(t *testing.T) {
		config := DefaultConfig()
		WithLogstashRetries(5)(config)
		if config.LogstashRetries != 5 {
			t.Errorf("expected LogstashRetries 5, got %d", config.LogstashRetries)
		}
	})

	t.Run("WithFile", func(t *testing.T) {
		config := DefaultConfig()
		WithFile("/var/log/app.log")(config)
		if config.FilePath != "/var/log/app.log" {
			t.Errorf("expected FilePath '/var/log/app.log', got %s", config.FilePath)
		}
	})

	t.Run("WithFileRotation", func(t *testing.T) {
		config := DefaultConfig()
		WithFileRotation(200, 60, 20)(config)
		if config.FileMaxSize != 200 {
			t.Errorf("expected FileMaxSize 200, got %d", config.FileMaxSize)
		}
		if config.FileMaxAge != 60 {
			t.Errorf("expected FileMaxAge 60, got %d", config.FileMaxAge)
		}
		if config.FileMaxBackups != 20 {
			t.Errorf("expected FileMaxBackups 20, got %d", config.FileMaxBackups)
		}
	})

	t.Run("WithServiceName", func(t *testing.T) {
		config := DefaultConfig()
		WithServiceName("my-service")(config)
		if config.ServiceName != "my-service" {
			t.Errorf("expected ServiceName 'my-service', got %s", config.ServiceName)
		}
	})

	t.Run("WithEnvironment", func(t *testing.T) {
		config := DefaultConfig()
		WithEnvironment("production")(config)
		if config.Environment != "production" {
			t.Errorf("expected Environment 'production', got %s", config.Environment)
		}
	})

	t.Run("WithVersion", func(t *testing.T) {
		config := DefaultConfig()
		WithVersion("1.2.3")(config)
		if config.Version != "1.2.3" {
			t.Errorf("expected Version '1.2.3', got %s", config.Version)
		}
	})

	t.Run("WithHost", func(t *testing.T) {
		config := DefaultConfig()
		WithHost("server-01")(config)
		if config.Host != "server-01" {
			t.Errorf("expected Host 'server-01', got %s", config.Host)
		}
	})

	t.Run("WithReportCaller", func(t *testing.T) {
		config := DefaultConfig()
		WithReportCaller(true)(config)
		if !config.ReportCaller {
			t.Error("expected ReportCaller to be true")
		}
	})

	t.Run("WithMultiOutputs", func(t *testing.T) {
		config := DefaultConfig()
		outputs := []OutputConfig{
			{Type: "console", Enabled: true},
			{Type: "file", Enabled: true, FilePath: "/tmp/log.txt"},
		}
		WithMultiOutputs(outputs)(config)
		if len(config.MultiOutputs) != 2 {
			t.Errorf("expected 2 MultiOutputs, got %d", len(config.MultiOutputs))
		}
	})
}

func TestNewWithOptions(t *testing.T) {
	t.Run("creates logger with multiple options", func(t *testing.T) {
		logger, err := NewWithOptions(
			WithEnabled(true),
			WithLevel("debug"),
			WithFormat("json"),
			WithServiceName("test-service"),
			WithEnvironment("test"),
			WithVersion("1.0.0"),
		)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if !logger.IsEnabled() {
			t.Error("expected logger to be enabled")
		}
		if logger.config.Level != "debug" {
			t.Error("expected debug level")
		}
		if logger.config.Format != "json" {
			t.Error("expected json format")
		}
		if logger.config.ServiceName != "test-service" {
			t.Error("expected service name to be set")
		}
	})

	t.Run("creates logger with console output", func(t *testing.T) {
		logger, err := NewWithOptions(
			WithOutput("console"),
			WithConsoleColors(true),
		)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if logger.config.Output != "console" {
			t.Error("expected console output")
		}
		if !logger.config.ConsoleColors {
			t.Error("expected console colors to be enabled")
		}
	})

	t.Run("fails with invalid options", func(t *testing.T) {
		_, err := NewWithOptions(
			WithOutput("invalid"),
		)

		if err == nil {
			t.Error("expected error with invalid output")
		}
	})
}

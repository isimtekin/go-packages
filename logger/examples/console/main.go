package main

import (
	"errors"
	"time"

	"github.com/isimtekin/go-packages/logger"
)

func main() {
	// Create logger with console output
	log, err := logger.NewWithOptions(
		logger.WithEnabled(true),
		logger.WithLevel("debug"),
		logger.WithFormat("text"),
		logger.WithConsoleColors(true),
		logger.WithServiceName("console-example"),
		logger.WithEnvironment("development"),
		logger.WithVersion("1.0.0"),
	)
	if err != nil {
		panic(err)
	}

	// Basic logging
	log.Info("Application started")
	log.Debug("Debug information")
	log.Warn("This is a warning")

	// Structured logging with fields
	log.Info("User logged in",
		logger.String("user_id", "12345"),
		logger.String("username", "johndoe"),
		logger.String("ip", "192.168.1.100"),
		logger.Int("login_attempts", 1),
	)

	// Error logging
	err = errors.New("database connection failed")
	log.Error("Failed to connect to database",
		logger.Err(err),
		logger.String("host", "localhost"),
		logger.Int("port", 5432),
		logger.Int("retry_count", 3),
	)

	// Different field types
	log.Info("Order processed",
		logger.String("order_id", "ORD-12345"),
		logger.Float64("amount", 99.99),
		logger.Int("quantity", 2),
		logger.Bool("express_shipping", true),
		logger.Any("metadata", map[string]string{
			"source":  "web",
			"payment": "credit_card",
		}),
	)

	// Using logger with fields
	userLogger := log.WithFields(
		logger.String("user_id", "12345"),
		logger.String("session_id", "sess-abc-123"),
	)

	userLogger.Info("User viewed product")
	userLogger.Info("User added to cart")
	userLogger.Info("User completed checkout")

	// Check debug mode
	if log.IsDebugEnabled() {
		log.Debug("Debug mode is enabled",
			logger.String("timestamp", time.Now().Format(time.RFC3339)),
		)
	}

	// JSON format example
	jsonLog, _ := logger.NewWithOptions(
		logger.WithFormat("json"),
		logger.WithLevel("info"),
	)

	jsonLog.Info("JSON formatted log",
		logger.String("format", "json"),
		logger.Int("example", 1),
	)

	log.Info("Application finished")
}

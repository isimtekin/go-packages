package main

import (
	"context"
	"errors"
	"time"

	"github.com/isimtekin/go-packages/logger"
)

func main() {
	// Create logger with JSON format for structured logging
	log, err := logger.NewWithOptions(
		logger.WithFormat("json"),
		logger.WithLevel("debug"),
		logger.WithServiceName("structured-example"),
		logger.WithEnvironment("production"),
		logger.WithVersion("2.1.0"),
	)
	if err != nil {
		panic(err)
	}

	// Simple structured log
	log.Info("Application started",
		logger.String("startup_mode", "normal"),
		logger.Int("worker_count", 10),
	)

	// Complex structured data
	log.Info("User registration",
		logger.String("user_id", "usr_12345"),
		logger.String("email", "user@example.com"),
		logger.String("registration_source", "web"),
		logger.Bool("email_verified", true),
		logger.Int("age", 28),
		logger.Any("preferences", map[string]interface{}{
			"newsletter": true,
			"language":   "en",
			"timezone":   "UTC",
		}),
	)

	// HTTP request logging
	log.Info("HTTP request received",
		logger.String("method", "POST"),
		logger.String("path", "/api/users"),
		logger.String("remote_addr", "192.168.1.100"),
		logger.String("user_agent", "Mozilla/5.0"),
		logger.Int("status_code", 201),
		logger.Float64("duration_ms", 45.23),
	)

	// Database operation logging
	log.Info("Database query executed",
		logger.String("query", "SELECT * FROM users WHERE id = ?"),
		logger.Int("rows_affected", 1),
		logger.Float64("execution_time_ms", 12.5),
		logger.String("database", "main"),
	)

	// Error with context
	dbErr := errors.New("connection timeout")
	log.Error("Database operation failed",
		logger.Err(dbErr),
		logger.String("operation", "INSERT"),
		logger.String("table", "orders"),
		logger.Int("retry_attempt", 3),
		logger.Float64("timeout_seconds", 30.0),
	)

	// Business metrics
	log.Info("Order processed",
		logger.String("order_id", "ORD-2025-001"),
		logger.String("customer_id", "CUST-12345"),
		logger.Float64("order_total", 299.99),
		logger.String("currency", "USD"),
		logger.Int("item_count", 3),
		logger.String("payment_method", "credit_card"),
		logger.String("shipping_address", "New York, NY"),
		logger.Float64("processing_time_ms", 156.78),
	)

	// Context-aware logging
	ctx := context.WithValue(context.Background(), "request_id", "req-abc-123")
	ctx = context.WithValue(ctx, "user_id", "usr-456")

	requestLogger := log.WithContext(ctx)
	requestLogger.Info("Processing user request")
	requestLogger.Debug("Validating input")
	requestLogger.Info("Request completed",
		logger.Float64("total_time_ms", 234.56),
	)

	// Logger with persistent fields
	serviceLogger := log.WithFields(
		logger.String("service", "payment-processor"),
		logger.String("version", "2.1.0"),
	)

	serviceLogger.Info("Payment initiated",
		logger.String("transaction_id", "txn-789"),
		logger.Float64("amount", 99.99),
	)

	serviceLogger.Info("Payment authorized",
		logger.String("transaction_id", "txn-789"),
		logger.String("auth_code", "AUTH123456"),
	)

	serviceLogger.Info("Payment captured",
		logger.String("transaction_id", "txn-789"),
		logger.String("status", "completed"),
	)

	// Time-series data
	for i := 0; i < 3; i++ {
		log.Info("Metric collected",
			logger.String("metric_name", "cpu_usage"),
			logger.Float64("value", 45.5+float64(i)*2.3),
			logger.String("unit", "percent"),
			logger.String("timestamp", time.Now().Format(time.RFC3339)),
		)
		time.Sleep(100 * time.Millisecond)
	}

	log.Info("Application shutdown",
		logger.String("reason", "normal"),
		logger.Int("uptime_seconds", 3600),
	)
}

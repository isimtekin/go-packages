package main

import (
	"os"

	"github.com/isimtekin/go-packages/logger"
)

func main() {
	// Set environment variables
	os.Setenv("LOG_ENABLED", "true")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("LOG_OUTPUT", "console")
	os.Setenv("LOG_FORMAT", "json")
	os.Setenv("LOG_SERVICE_NAME", "env-example")
	os.Setenv("LOG_ENVIRONMENT", "development")
	os.Setenv("LOG_VERSION", "1.0.0")

	// Create logger from environment variables
	log, err := logger.NewFromEnvWithDefaults()
	if err != nil {
		panic(err)
	}

	log.Info("Logger created from environment variables")
	log.Debug("Configuration loaded from env",
		logger.String("prefix", "LOG_"),
	)

	// Using custom prefix
	os.Setenv("APP_LOG_LEVEL", "info")
	os.Setenv("APP_LOG_SERVICE_NAME", "custom-prefix-example")

	customLog, err := logger.NewFromEnv("APP_LOG_")
	if err != nil {
		panic(err)
	}

	customLog.Info("Logger with custom prefix",
		logger.String("prefix", "APP_LOG_"),
	)

	// Set global logger from environment
	err = logger.SetGlobalFromEnvWithDefaults()
	if err != nil {
		panic(err)
	}

	// Use global logger
	logger.Info("Using global logger from environment")
	logger.Debug("Global logger is configured",
		logger.String("source", "environment"),
	)

	// Load configuration only
	config, err := logger.LoadConfigFromEnvWithDefaults()
	if err != nil {
		panic(err)
	}

	log.Info("Configuration loaded",
		logger.String("level", config.Level),
		logger.String("output", config.Output),
		logger.String("format", config.Format),
		logger.String("service", config.ServiceName),
	)
}

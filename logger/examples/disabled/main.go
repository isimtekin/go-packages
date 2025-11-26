package main

import (
	"time"

	"github.com/isimtekin/go-packages/logger"
)

// This example demonstrates the enable/disable feature
func main() {
	// Create a disabled logger
	log, err := logger.NewWithOptions(
		logger.WithEnabled(false), // Logging is disabled
		logger.WithLevel("debug"),
		logger.WithServiceName("disabled-example"),
	)
	if err != nil {
		panic(err)
	}

	println("Logger is disabled, these messages won't be logged:")

	// These calls are no-ops (zero overhead)
	log.Info("This won't be logged")
	log.Debug("This won't be logged either")
	log.Error("Even errors won't be logged")

	// Check if logging is enabled
	if log.IsEnabled() {
		println("Logging is enabled")
	} else {
		println("Logging is disabled - confirmed!")
	}

	// Expensive operation that should only run when debug is enabled
	if log.IsDebugEnabled() {
		// This won't execute since debug is not enabled
		complexData := gatherComplexDebugData()
		log.Debug("Complex data", logger.Any("data", complexData))
	} else {
		println("Debug is disabled, skipping expensive operation")
	}

	// Now enable logging
	println("\nEnabling logger...")
	enabledLog, _ := logger.NewWithOptions(
		logger.WithEnabled(true),
		logger.WithLevel("debug"),
		logger.WithServiceName("enabled-example"),
	)

	enabledLog.Info("Now logging is enabled!")
	enabledLog.Debug("Debug messages are visible")

	if enabledLog.IsEnabled() {
		println("Logging is now enabled - confirmed!")
	}

	if enabledLog.IsDebugEnabled() {
		enabledLog.Debug("Debug is enabled, running expensive operation")
		complexData := gatherComplexDebugData()
		enabledLog.Debug("Complex data collected",
			logger.Int("data_size", len(complexData)),
		)
	}
}

// Simulates an expensive operation
func gatherComplexDebugData() map[string]interface{} {
	println("Gathering complex debug data (expensive operation)...")
	time.Sleep(100 * time.Millisecond)
	return map[string]interface{}{
		"timestamp": time.Now(),
		"stats": map[string]int{
			"requests": 1000,
			"errors":   5,
		},
	}
}

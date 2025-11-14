package envutil

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// Package-level default client for standalone functions
var defaultClient = NewDefault()

// GetEnv gets a string environment variable with default value
func GetEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// GetEnvBool gets a boolean environment variable with default value
func GetEnvBool(key string, defaultVal bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	
	// Handle yes/no values
	valLower := strings.ToLower(val)
	if valLower == "yes" || valLower == "y" {
		return true
	}
	if valLower == "no" || valLower == "n" {
		return false
	}
	
	parsed, err := strconv.ParseBool(val)
	if err != nil {
		log.Printf("Invalid boolean for %s: %s", key, val)
		return defaultVal
	}
	return parsed
}

// GetEnvInt gets an integer environment variable with default value
func GetEnvInt(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	parsed, err := strconv.Atoi(val)
	if err != nil {
		log.Printf("Invalid int for %s: %s", key, val)
		return defaultVal
	}
	return parsed
}

// GetEnvInt64 gets an int64 environment variable with default value
func GetEnvInt64(key string, defaultVal int64) int64 {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	parsed, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		log.Printf("Invalid int64 for %s: %s", key, val)
		return defaultVal
	}
	return parsed
}

// GetEnvFloat64 gets a float64 environment variable with default value
func GetEnvFloat64(key string, defaultVal float64) float64 {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	parsed, err := strconv.ParseFloat(val, 64)
	if err != nil {
		log.Printf("Invalid float64 for %s: %s", key, val)
		return defaultVal
	}
	return parsed
}

// GetEnvDuration gets a duration environment variable with default value
func GetEnvDuration(key string, defaultVal time.Duration) time.Duration {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	
	// Check if value contains a duration suffix
	if strings.ContainsAny(val, "smhµu") {
		duration, err := time.ParseDuration(val)
		if err != nil {
			log.Printf("Invalid duration for %s: %s", key, val)
			return defaultVal
		}
		return duration
	}
	
	// Otherwise parse as integer
	intVal, err := strconv.Atoi(val)
	if err != nil || intVal < 0 {
		log.Printf("Invalid duration for %s: %s", key, val)
		return defaultVal
	}
	
	// Guess unit based on key name
	keyLower := strings.ToLower(key)
	if strings.Contains(keyLower, "_ms") || strings.HasSuffix(keyLower, "ms") {
		return time.Duration(intVal) * time.Millisecond
	} else if strings.Contains(keyLower, "_us") || strings.HasSuffix(keyLower, "us") {
		return time.Duration(intVal) * time.Microsecond
	} else if strings.Contains(keyLower, "_min") || strings.HasSuffix(keyLower, "min") {
		return time.Duration(intVal) * time.Minute
	} else if strings.Contains(keyLower, "_hour") || strings.HasSuffix(keyLower, "hour") {
		return time.Duration(intVal) * time.Hour
	}
	
	// Default to seconds
	return time.Duration(intVal) * time.Second
}

// GetEnvStringSlice gets a string slice from comma-separated environment variable
func GetEnvStringSlice(key string, defaultVal []string) []string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	
	// Parse comma-separated values
	parts := strings.Split(val, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	
	if len(result) == 0 {
		return defaultVal
	}
	return result
}

// GetEnvIntSlice gets an int slice from comma-separated environment variable
func GetEnvIntSlice(key string, defaultVal []int) []int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	
	// Parse comma-separated values
	parts := strings.Split(val, ",")
	result := make([]int, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		
		num, err := strconv.Atoi(trimmed)
		if err != nil {
			log.Printf("Invalid int in slice for %s: %s", key, trimmed)
			continue
		}
		result = append(result, num)
	}
	
	if len(result) == 0 {
		return defaultVal
	}
	return result
}

// GetEnvURL gets a URL from environment variable
func GetEnvURL(key string, defaultVal string) *url.URL {
	val := GetEnv(key, defaultVal)
	if val == "" {
		return nil
	}
	
	parsed, err := url.Parse(val)
	if err != nil {
		log.Printf("Invalid URL for %s: %s, error: %v", key, val, err)
		if defaultVal != "" {
			parsed, _ = url.Parse(defaultVal)
		}
		return parsed
	}
	return parsed
}

// MustGetEnv gets a string environment variable or panics if not set
func MustGetEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic("Required environment variable " + key + " is not set")
	}
	return val
}

// MustGetEnvInt gets an integer environment variable or panics if not set or invalid
func MustGetEnvInt(key string) int {
	val := os.Getenv(key)
	if val == "" {
		panic("Required environment variable " + key + " is not set")
	}
	
	parsed, err := strconv.Atoi(val)
	if err != nil {
		panic("Invalid integer value for " + key + ": " + val)
	}
	return parsed
}

// MustGetEnvBool gets a boolean environment variable or panics if not set or invalid
func MustGetEnvBool(key string) bool {
	val := os.Getenv(key)
	if val == "" {
		panic("Required environment variable " + key + " is not set")
	}
	
	// Handle yes/no values
	valLower := strings.ToLower(val)
	if valLower == "yes" || valLower == "y" {
		return true
	}
	if valLower == "no" || valLower == "n" {
		return false
	}
	
	parsed, err := strconv.ParseBool(val)
	if err != nil {
		panic("Invalid boolean value for " + key + ": " + val)
	}
	return parsed
}

// IsEnvSet checks if an environment variable is set
func IsEnvSet(key string) bool {
	_, exists := os.LookupEnv(key)
	return exists
}

// GetEnvOrDefault returns the environment variable value or a default based on type inference
func GetEnvOrDefault(key string, defaultVal interface{}) interface{} {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	
	// Try to convert to the same type as default
	switch v := defaultVal.(type) {
	case string:
		return val
	case int:
		if parsed, err := strconv.Atoi(val); err == nil {
			return parsed
		}
		return v
	case int64:
		if parsed, err := strconv.ParseInt(val, 10, 64); err == nil {
			return parsed
		}
		return v
	case float64:
		if parsed, err := strconv.ParseFloat(val, 64); err == nil {
			return parsed
		}
		return v
	case bool:
		if parsed, err := strconv.ParseBool(val); err == nil {
			return parsed
		}
		return v
	case time.Duration:
		if parsed, err := time.ParseDuration(val); err == nil {
			return parsed
		}
		// Try integer as seconds
		if intVal, err := strconv.Atoi(val); err == nil {
			return time.Duration(intVal) * time.Second
		}
		return v
	default:
		return val
	}
}

// LoadEnvFile loads environment variables from a .env file (standalone function)
func LoadEnvFile(filename string) error {
	return defaultClient.LoadEnvFile(filename)
}

// ValidateRequired validates that all required environment variables are set
func ValidateRequired(keys ...string) error {
	missing := defaultClient.ValidateRequired(keys)
	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %v", missing)
	}
	return nil
}

// GetEnvWithFallback tries multiple environment variable keys in order
func GetEnvWithFallback(keys []string, defaultVal string) string {
	for _, key := range keys {
		if val := os.Getenv(key); val != "" {
			return val
		}
	}
	return defaultVal
}

// SetEnvIfNotSet sets an environment variable only if it's not already set
func SetEnvIfNotSet(key, value string) error {
	if !IsEnvSet(key) {
		return os.Setenv(key, value)
	}
	return nil
}

// ExpandEnv expands environment variables in a string (like shell $VAR or ${VAR})
func ExpandEnv(s string) string {
	return os.ExpandEnv(s)
}

// GetAllEnvWithPrefix returns all environment variables with a specific prefix
func GetAllEnvWithPrefix(prefix string) map[string]string {
	result := make(map[string]string)
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if len(pair) == 2 && strings.HasPrefix(pair[0], prefix) {
			result[pair[0]] = pair[1]
		}
	}
	return result
}

// GetEnvPort gets a port number from environment variable with validation
func GetEnvPort(key string, defaultVal int) int {
	val := GetEnvInt(key, defaultVal)
	if val < 1 || val > 65535 {
		log.Printf("Invalid port number for %s: %d, using default: %d", key, val, defaultVal)
		return defaultVal
	}
	return val
}
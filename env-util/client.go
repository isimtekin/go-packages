package envutil

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Logger interface for custom logging
type Logger interface {
	Printf(format string, v ...interface{})
}

// defaultLogger uses standard log package
type defaultLogger struct{}

func (d defaultLogger) Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

// Config holds configuration for env-util
type Config struct {
	Logger      Logger
	Silent      bool   // Don't log warnings for invalid values
	EnvPrefix   string // Optional prefix for all env variables
	EnvFile     string // Optional .env file path to load
	Required    []string // List of required environment variables
}

// Client provides environment variable utilities
type Client struct {
	config *Config
	logger Logger
	cache  map[string]string // Cache for loaded values
}

// New creates a new env-util client
func New(config *Config) *Client {
	if config == nil {
		config = &Config{}
	}
	
	client := &Client{
		config: config,
		cache:  make(map[string]string),
	}
	
	// Set default logger
	if config.Logger == nil && !config.Silent {
		client.logger = defaultLogger{}
	} else if !config.Silent {
		client.logger = config.Logger
	}
	
	// Load .env file if specified
	if config.EnvFile != "" {
		if err := client.LoadEnvFile(config.EnvFile); err != nil {
			client.logf("Failed to load env file %s: %v", config.EnvFile, err)
		}
	}
	
	// Validate required variables
	if len(config.Required) > 0 {
		missing := client.ValidateRequired(config.Required)
		if len(missing) > 0 {
			panic(fmt.Sprintf("Missing required environment variables: %v", missing))
		}
	}
	
	return client
}

// NewDefault creates a client with default configuration
func NewDefault() *Client {
	return New(nil)
}

// logf logs a message if logger is set
func (c *Client) logf(format string, v ...interface{}) {
	if c.logger != nil {
		c.logger.Printf(format, v...)
	}
}

// prefixKey adds configured prefix to key
func (c *Client) prefixKey(key string) string {
	if c.config.EnvPrefix != "" {
		return c.config.EnvPrefix + key
	}
	return key
}

// GetString gets a string environment variable with default value
func (c *Client) GetString(key, defaultVal string) string {
	key = c.prefixKey(key)
	
	// Check cache first
	if val, ok := c.cache[key]; ok {
		return val
	}
	
	if val := os.Getenv(key); val != "" {
		c.cache[key] = val
		return val
	}
	return defaultVal
}

// GetBool gets a boolean environment variable with default value
func (c *Client) GetBool(key string, defaultVal bool) bool {
	key = c.prefixKey(key)
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
		c.logf("Invalid boolean for %s: %s, using default: %v", key, val, defaultVal)
		return defaultVal
	}
	return parsed
}

// GetInt gets an integer environment variable with default value
func (c *Client) GetInt(key string, defaultVal int) int {
	key = c.prefixKey(key)
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	
	parsed, err := strconv.Atoi(val)
	if err != nil {
		c.logf("Invalid int for %s: %s, using default: %d", key, val, defaultVal)
		return defaultVal
	}
	return parsed
}

// GetInt64 gets an int64 environment variable with default value
func (c *Client) GetInt64(key string, defaultVal int64) int64 {
	key = c.prefixKey(key)
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	
	parsed, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		c.logf("Invalid int64 for %s: %s, using default: %d", key, val, defaultVal)
		return defaultVal
	}
	return parsed
}

// GetFloat64 gets a float64 environment variable with default value
func (c *Client) GetFloat64(key string, defaultVal float64) float64 {
	key = c.prefixKey(key)
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	
	parsed, err := strconv.ParseFloat(val, 64)
	if err != nil {
		c.logf("Invalid float64 for %s: %s, using default: %f", key, val, defaultVal)
		return defaultVal
	}
	return parsed
}

// GetDuration gets a duration environment variable with default value
func (c *Client) GetDuration(key string, defaultVal time.Duration) time.Duration {
	key = c.prefixKey(key)
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	
	// Check if value contains a duration suffix
	if strings.ContainsAny(val, "smhµu") {
		duration, err := time.ParseDuration(val)
		if err != nil {
			c.logf("Invalid duration for %s: %s, using default: %v", key, val, defaultVal)
			return defaultVal
		}
		return duration
	}
	
	// Otherwise parse as integer
	intVal, err := strconv.Atoi(val)
	if err != nil || intVal < 0 {
		c.logf("Invalid duration for %s: %s, using default: %v", key, val, defaultVal)
		return defaultVal
	}
	
	// Guess unit based on key name
	keyLower := strings.ToLower(key)
	if strings.Contains(keyLower, "_ms") || strings.HasSuffix(keyLower, "ms") {
		return time.Duration(intVal) * time.Millisecond
	} else if strings.Contains(keyLower, "_us") || strings.HasSuffix(keyLower, "us") {
		return time.Duration(intVal) * time.Microsecond
	} else if strings.Contains(keyLower, "_ns") || strings.HasSuffix(keyLower, "ns") {
		return time.Duration(intVal) * time.Nanosecond
	} else if strings.Contains(keyLower, "_min") || strings.HasSuffix(keyLower, "min") {
		return time.Duration(intVal) * time.Minute
	} else if strings.Contains(keyLower, "_hour") || strings.HasSuffix(keyLower, "hour") {
		return time.Duration(intVal) * time.Hour
	}
	
	// Default to seconds
	return time.Duration(intVal) * time.Second
}

// GetStringSlice gets a string slice from comma-separated environment variable
func (c *Client) GetStringSlice(key string, defaultVal []string) []string {
	key = c.prefixKey(key)
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

// GetIntSlice gets an int slice from comma-separated environment variable
func (c *Client) GetIntSlice(key string, defaultVal []int) []int {
	key = c.prefixKey(key)
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
			c.logf("Invalid int in slice for %s: %s", key, trimmed)
			continue
		}
		result = append(result, num)
	}
	
	if len(result) == 0 {
		return defaultVal
	}
	return result
}

// GetURL gets a URL from environment variable
func (c *Client) GetURL(key string, defaultVal *url.URL) *url.URL {
	key = c.prefixKey(key)
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	
	parsed, err := url.Parse(val)
	if err != nil {
		c.logf("Invalid URL for %s: %s, error: %v", key, val, err)
		return defaultVal
	}
	return parsed
}

// GetFilePath gets a file path and validates it exists
func (c *Client) GetFilePath(key string, defaultVal string) string {
	key = c.prefixKey(key)
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	
	// Expand home directory
	if strings.HasPrefix(val, "~") {
		home, err := os.UserHomeDir()
		if err == nil {
			val = strings.Replace(val, "~", home, 1)
		}
	}
	
	// Make absolute path
	absPath, err := filepath.Abs(val)
	if err != nil {
		c.logf("Invalid file path for %s: %s, error: %v", key, val, err)
		return defaultVal
	}
	
	// Check if file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		c.logf("File does not exist for %s: %s", key, absPath)
		return defaultVal
	}
	
	return absPath
}

// GetJSON unmarshals JSON from environment variable
func (c *Client) GetJSON(key string, target interface{}) error {
	key = c.prefixKey(key)
	val := os.Getenv(key)
	if val == "" {
		return fmt.Errorf("environment variable %s is not set", key)
	}
	
	if err := json.Unmarshal([]byte(val), target); err != nil {
		return fmt.Errorf("failed to unmarshal JSON from %s: %w", key, err)
	}
	return nil
}

// MustGetString gets a string environment variable or panics if not set
func (c *Client) MustGetString(key string) string {
	key = c.prefixKey(key)
	val := os.Getenv(key)
	if val == "" {
		panic(fmt.Sprintf("Required environment variable %s is not set", key))
	}
	return val
}

// MustGetInt gets an integer environment variable or panics if not set or invalid
func (c *Client) MustGetInt(key string) int {
	key = c.prefixKey(key)
	val := os.Getenv(key)
	if val == "" {
		panic(fmt.Sprintf("Required environment variable %s is not set", key))
	}
	
	parsed, err := strconv.Atoi(val)
	if err != nil {
		panic(fmt.Sprintf("Invalid integer value for %s: %s", key, val))
	}
	return parsed
}

// IsSet checks if an environment variable is set
func (c *Client) IsSet(key string) bool {
	key = c.prefixKey(key)
	_, exists := os.LookupEnv(key)
	return exists
}

// ValidateRequired checks if all required environment variables are set
func (c *Client) ValidateRequired(keys []string) []string {
	var missing []string
	for _, key := range keys {
		if !c.IsSet(key) {
			missing = append(missing, c.prefixKey(key))
		}
	}
	return missing
}

// LoadEnvFile loads environment variables from a file
func (c *Client) LoadEnvFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open env file: %w", err)
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	lineNum := 0
	
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		// Parse KEY=VALUE format
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			c.logf("Invalid line %d in %s: %s", lineNum, filename, line)
			continue
		}
		
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		// Remove quotes if present
		if len(value) >= 2 {
			if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
				(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
				value = value[1 : len(value)-1]
			}
		}
		
		// Set environment variable if not already set
		if _, exists := os.LookupEnv(key); !exists {
			if err := os.Setenv(key, value); err != nil {
				c.logf("Failed to set %s: %v", key, err)
			} else {
				c.cache[key] = value
			}
		}
	}
	
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading env file: %w", err)
	}
	
	return nil
}

// Export exports all environment variables to a map
func (c *Client) Export() map[string]string {
	env := make(map[string]string)
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if len(pair) == 2 {
			// Apply prefix filter if configured
			if c.config.EnvPrefix != "" {
				if strings.HasPrefix(pair[0], c.config.EnvPrefix) {
					env[pair[0]] = pair[1]
				}
			} else {
				env[pair[0]] = pair[1]
			}
		}
	}
	return env
}

// SetEnv sets an environment variable (mainly for testing)
func (c *Client) SetEnv(key, value string) error {
	key = c.prefixKey(key)
	c.cache[key] = value
	return os.Setenv(key, value)
}

// UnsetEnv unsets an environment variable
func (c *Client) UnsetEnv(key string) error {
	key = c.prefixKey(key)
	delete(c.cache, key)
	return os.Unsetenv(key)
}

// ClearCache clears the internal cache
func (c *Client) ClearCache() {
	c.cache = make(map[string]string)
}
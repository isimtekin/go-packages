package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	envutil "github.com/isimtekin/go-packages/env-util"
)

// Config represents application configuration
type Config struct {
	// Server
	Host         string
	Port         int
	MetricsPort  int
	
	// Database
	DatabaseURL       string
	MaxConnections    int
	ConnectionTimeout time.Duration
	
	// Redis
	RedisHost     string
	RedisPort     int
	RedisPassword string
	RedisDB       int
	
	// API
	APIKey      string
	APIEndpoint string
	APITimeout  time.Duration
	RateLimit   int
	
	// Features
	Debug        bool
	LogLevel     string
	EnableCache  bool
	EnableNewUI  bool
	
	// Lists
	AllowedOrigins []string
	KafkaBrokers   []string
}

func main() {
	// Example 1: Simple standalone usage
	fmt.Println("=== Example 1: Standalone Functions ===")
	simpleExample()
	
	// Example 2: Using client with prefix
	fmt.Println("\n=== Example 2: Client with Prefix ===")
	clientExample()
	
	// Example 3: Loading from .env file
	fmt.Println("\n=== Example 3: Loading from .env ===")
	envFileExample()
	
	// Example 4: Complete application config
	fmt.Println("\n=== Example 4: Full Application Config ===")
	fullConfigExample()
}

func simpleExample() {
	// Get various types with defaults
	host := envutil.GetEnv("HOST", "localhost")
	port := envutil.GetEnvInt("PORT", 8080)
	debug := envutil.GetEnvBool("DEBUG", false)
	timeout := envutil.GetEnvDuration("TIMEOUT", 30*time.Second)
	
	fmt.Printf("Server: %s:%d (debug=%v, timeout=%v)\n", host, port, debug, timeout)
	
	// Check if variable exists
	if envutil.IsEnvSet("CUSTOM_CONFIG") {
		fmt.Println("Custom config is set")
	} else {
		fmt.Println("Custom config not set")
	}
	
	// Get with fallback
	dbHost := envutil.GetEnvWithFallback(
		[]string{"DATABASE_HOST", "DB_HOST", "POSTGRES_HOST"},
		"localhost",
	)
	fmt.Printf("Database host: %s\n", dbHost)
	
	// Get all with prefix
	appVars := envutil.GetAllEnvWithPrefix("APP_")
	fmt.Printf("Found %d APP_ variables\n", len(appVars))
}

func clientExample() {
	// Create client with prefix
	client := envutil.NewWithOptions(
		envutil.WithPrefix("MYAPP_"),
		envutil.WithSilent(false), // Enable logging
	)
	
	// Set some test variables
	client.SetEnv("DATABASE_URL", "postgres://localhost/myapp")
	client.SetEnv("CACHE_ENABLED", "true")
	client.SetEnv("MAX_WORKERS", "10")
	
	// Get with prefix (automatically prepends MYAPP_)
	dbURL := client.GetString("DATABASE_URL", "postgres://localhost/db")
	cacheEnabled := client.GetBool("CACHE_ENABLED", false)
	maxWorkers := client.GetInt("MAX_WORKERS", 5)
	
	fmt.Printf("Database URL: %s\n", dbURL)
	fmt.Printf("Cache Enabled: %v\n", cacheEnabled)
	fmt.Printf("Max Workers: %d\n", maxWorkers)
	
	// Export all with prefix
	exported := client.Export()
	fmt.Printf("Exported %d variables\n", len(exported))
}

func envFileExample() {
	// Create .env content for demo
	envContent := `
# Demo .env file
APP_NAME=DemoApp
APP_PORT=3000
APP_DEBUG=true
DATABASE_URL=postgres://user:pass@localhost/demo
REDIS_HOST=redis.local
API_KEYS=key1,key2,key3
REQUEST_TIMEOUT=15s
`
	
	// Write temporary .env file
	if err := os.WriteFile(".env.demo", []byte(envContent), 0644); err != nil {
		log.Printf("Failed to create demo .env: %v", err)
		return
	}
	defer os.Remove(".env.demo")
	
	// Load from file
	if err := envutil.LoadEnvFile(".env.demo"); err != nil {
		log.Printf("Failed to load .env: %v", err)
		return
	}
	
	// Now use the loaded variables
	appName := envutil.GetEnv("APP_NAME", "Unknown")
	appPort := envutil.GetEnvInt("APP_PORT", 8080)
	apiKeys := envutil.GetEnvStringSlice("API_KEYS", nil)
	
	fmt.Printf("Loaded from .env:\n")
	fmt.Printf("  App Name: %s\n", appName)
	fmt.Printf("  App Port: %d\n", appPort)
	fmt.Printf("  API Keys: %v\n", apiKeys)
}

func fullConfigExample() {
	// Create a comprehensive config loader
	config := loadConfig()
	
	// Display loaded configuration
	fmt.Printf("Configuration loaded:\n")
	fmt.Printf("  Server: %s:%d\n", config.Host, config.Port)
	fmt.Printf("  Debug: %v, Log Level: %s\n", config.Debug, config.LogLevel)
	fmt.Printf("  Database: %s (max connections: %d)\n", 
		maskPassword(config.DatabaseURL), config.MaxConnections)
	fmt.Printf("  Redis: %s:%d (db: %d)\n", 
		config.RedisHost, config.RedisPort, config.RedisDB)
	fmt.Printf("  API Endpoint: %s\n", config.APIEndpoint)
	fmt.Printf("  Rate Limit: %d requests\n", config.RateLimit)
	fmt.Printf("  Features: Cache=%v, NewUI=%v\n", 
		config.EnableCache, config.EnableNewUI)
	
	// Validate configuration
	if err := validateConfig(config); err != nil {
		log.Printf("Configuration validation failed: %v", err)
	} else {
		fmt.Println("Configuration validated successfully!")
	}
}

func loadConfig() *Config {
	// Create client with options
	client := envutil.NewWithOptions(
		envutil.WithPrefix("APP_"),
		envutil.WithSilent(true),
	)
	
	// Set some defaults for demo
	client.SetEnv("HOST", "0.0.0.0")
	client.SetEnv("PORT", "8080")
	client.SetEnv("DATABASE_URL", "postgres://user:password@localhost/app")
	client.SetEnv("REDIS_HOST", "localhost")
	client.SetEnv("API_ENDPOINT", "https://api.example.com")
	
	return &Config{
		// Server
		Host:        client.GetString("HOST", "localhost"),
		Port:        client.GetInt("PORT", 8080),
		MetricsPort: client.GetInt("METRICS_PORT", 9090),
		
		// Database
		DatabaseURL:       client.GetString("DATABASE_URL", ""),
		MaxConnections:    client.GetInt("DB_MAX_CONN", 25),
		ConnectionTimeout: client.GetDuration("DB_TIMEOUT", 5*time.Second),
		
		// Redis
		RedisHost:     client.GetString("REDIS_HOST", "localhost"),
		RedisPort:     client.GetInt("REDIS_PORT", 6379),
		RedisPassword: client.GetString("REDIS_PASSWORD", ""),
		RedisDB:       client.GetInt("REDIS_DB", 0),
		
		// API
		APIKey:      client.GetString("API_KEY", ""),
		APIEndpoint: client.GetString("API_ENDPOINT", ""),
		APITimeout:  client.GetDuration("API_TIMEOUT", 30*time.Second),
		RateLimit:   client.GetInt("RATE_LIMIT", 100),
		
		// Features
		Debug:       client.GetBool("DEBUG", false),
		LogLevel:    client.GetString("LOG_LEVEL", "info"),
		EnableCache: client.GetBool("ENABLE_CACHE", true),
		EnableNewUI: client.GetBool("ENABLE_NEW_UI", false),
		
		// Lists
		AllowedOrigins: client.GetStringSlice("ALLOWED_ORIGINS", []string{"*"}),
		KafkaBrokers:   client.GetStringSlice("KAFKA_BROKERS", []string{"localhost:9092"}),
	}
}

func validateConfig(cfg *Config) error {
	// Validate required fields
	if cfg.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	
	if cfg.Port < 1 || cfg.Port > 65535 {
		return fmt.Errorf("invalid port number: %d", cfg.Port)
	}
	
	if cfg.MaxConnections < 1 {
		return fmt.Errorf("max connections must be at least 1")
	}
	
	return nil
}

func maskPassword(url string) string {
	// Simple password masking for display
	if idx := strings.Index(url, "://"); idx != -1 {
		prefix := url[:idx+3]
		remainder := url[idx+3:]
		if atIdx := strings.Index(remainder, "@"); atIdx != -1 {
			return prefix + "****:****@" + remainder[atIdx+1:]
		}
	}
	return url
}

// Example of using Must functions (these would panic if not set)
func mustExample() {
	// These will panic if environment variables are not set
	apiKey := envutil.MustGetEnv("API_KEY")
	dbURL := envutil.MustGetEnv("DATABASE_URL")
	port := envutil.MustGetEnvInt("PORT")
	
	fmt.Printf("Required: API Key=%s, DB=%s, Port=%d\n", 
		apiKey, dbURL, port)
}
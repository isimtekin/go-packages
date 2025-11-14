# env-util

[![Zero Dependencies](https://img.shields.io/badge/dependencies-zero-brightgreen)](https://github.com/isimtekin/go-packages/tree/main/env-util)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

A comprehensive, zero-dependency Go package for managing environment variables with type safety, validation, and convenience functions. Perfect for 12-factor apps and containerized applications.

## ✨ Features

- 🔒 **Type-safe parsing** - String, int, int64, float64, bool, duration, URL, port, slices
- 🎯 **Zero dependencies** - Pure Go standard library only
- ⚙️ **Flexible usage** - Standalone functions or client-based with configuration
- 📁 **.env file support** - Load variables from files
- ✅ **Validation** - Required variable checking with clear error messages
- 🚀 **Smart duration parsing** - Automatic unit detection based on key names
- 📋 **List parsing** - Comma-separated values to slices
- 💾 **Caching** - Built-in value caching for performance
- 🔧 **Functional options** - Clean configuration with prefix, silent mode, custom logger
- 🧪 **Testing helpers** - SetEnv/UnsetEnv for easy test setup

## 🚀 Quick Start

```bash
# Installation (after first release)
go get github.com/isimtekin/go-packages/env-util

# For development/testing from local
go get github.com/isimtekin/go-packages/env-util@latest
```

## 📋 Implementation Guide for AI

When implementing this package, follow this structured approach:

### Step 1: Import the Package

```go
import envutil "github.com/isimtekin/go-packages/env-util"
```

### Step 2: Choose Your Usage Pattern

#### Pattern A: Simple Standalone Functions (Most Common)
Use when you need quick access to environment variables without complex configuration.

```go
// Direct usage - no setup required
host := envutil.GetEnv("HOST", "localhost")
port := envutil.GetEnvInt("PORT", 8080)
debug := envutil.GetEnvBool("DEBUG", false)
```

#### Pattern B: Client with Configuration
Use when you need prefixes, .env files, or caching.

```go
// Create configured client
client := envutil.NewWithOptions(
    envutil.WithPrefix("APP_"),              // All keys prefixed with APP_
    envutil.WithEnvFile(".env"),             // Load from .env file
    envutil.WithSilent(true),                // Disable warnings
    envutil.WithRequired("API_KEY", "DB_URL"), // Panic if missing
)

// Use the client (prefix is automatic)
dbUrl := client.GetString("DATABASE_URL", "postgres://localhost/db")
```

**Available Options:**
- `WithPrefix(prefix)` - Add prefix to all keys
- `WithEnvFile(filename)` - Load .env file on creation
- `WithSilent(bool)` - Enable/disable logging
- `WithRequired(keys...)` - Validate required variables (panics if missing)
- `WithLogger(logger)` - Set custom logger (implements Logger interface)

### Step 3: Type-Safe Value Retrieval

The package provides type-specific getters with automatic parsing and validation:

```go
// String - returns raw value or default
name := envutil.GetEnv("APP_NAME", "MyApp")

// Integer - parses or returns default
workers := envutil.GetEnvInt("WORKERS", 4)

// Boolean - supports true/false, 1/0, yes/no, y/n
enabled := envutil.GetEnvBool("FEATURE_FLAG", false)

// Float
rate := envutil.GetEnvFloat64("SAMPLE_RATE", 0.1)

// Duration - supports Go duration syntax or intelligent parsing
timeout := envutil.GetEnvDuration("TIMEOUT", 30*time.Second)
// Supports: "30s", "5m", "2h" or plain numbers with context

// Lists - comma-separated values
hosts := envutil.GetEnvStringSlice("REDIS_HOSTS", []string{"localhost"})
ports := envutil.GetEnvIntSlice("PORTS", []int{8080, 8081})

// URL - parses and validates
apiUrl := envutil.GetEnvURL("API_ENDPOINT", "https://api.example.com")

// Port - validates range 1-65535
port := envutil.GetEnvPort("PORT", 8080)
```

## 📚 Core API Functions

### Basic Getters

| Function | Purpose | Example Usage | Default Behavior |
|----------|---------|---------------|------------------|
| `GetEnv(key, default)` | Get string value | `GetEnv("HOST", "localhost")` | Returns default if not set |
| `GetEnvInt(key, default)` | Get integer | `GetEnvInt("PORT", 8080)` | Returns default if not set or invalid |
| `GetEnvBool(key, default)` | Get boolean | `GetEnvBool("DEBUG", false)` | Supports multiple formats |
| `GetEnvFloat64(key, default)` | Get float | `GetEnvFloat64("RATE", 0.5)` | Returns default if invalid |
| `GetEnvInt64(key, default)` | Get int64 | `GetEnvInt64("BIG_NUM", 0)` | Returns default if invalid |
| `GetEnvDuration(key, default)` | Get time.Duration | `GetEnvDuration("TIMEOUT", 5*time.Second)` | Smart parsing with units |
| `GetEnvPort(key, default)` | Get port (1-65535) | `GetEnvPort("PORT", 8080)` | Validates port range |

### Required Variables (Panic on Missing)

| Function | Purpose | Example | Behavior |
|----------|---------|---------|----------|
| `MustGetEnv(key)` | Required string | `MustGetEnv("API_KEY")` | Panics if not set |
| `MustGetEnvInt(key)` | Required integer | `MustGetEnvInt("PORT")` | Panics if not set or invalid |
| `MustGetEnvBool(key)` | Required boolean | `MustGetEnvBool("PRODUCTION")` | Panics if not set or invalid |

### Validation & Utilities

| Function | Purpose | Example | Return |
|----------|---------|---------|--------|
| `IsEnvSet(key)` | Check if exists | `IsEnvSet("CUSTOM_CONFIG")` | `bool` |
| `ValidateRequired(keys...)` | Validate multiple | `ValidateRequired("DB_URL", "API_KEY")` | `error` or `nil` |
| `GetAllEnvWithPrefix(prefix)` | Get matching vars | `GetAllEnvWithPrefix("APP_")` | `map[string]string` |
| `GetEnvWithFallback(keys, default)` | Try multiple keys | `GetEnvWithFallback([]string{"DB_URL", "DATABASE_URL"}, "")` | First match or default |
| `LoadEnvFile(filename)` | Load .env file | `LoadEnvFile(".env")` | `error` or `nil` |
| `ExpandEnv(s)` | Expand $VAR syntax | `ExpandEnv("$HOME/config")` | Expanded string |

## 🎯 Implementation Examples

### Example 1: Basic Configuration Structure

```go
package main

import (
    "log"
    "time"
    envutil "github.com/isimtekin/go-packages/env-util"
)

type Config struct {
    // Server
    Host string
    Port int
    
    // Database
    DatabaseURL    string
    MaxConnections int
    Timeout        time.Duration
    
    // Features
    Debug      bool
    RateLimit  float64
}

func LoadConfig() (*Config, error) {
    // Validate required variables first
    if err := envutil.ValidateRequired("DATABASE_URL"); err != nil {
        return nil, err
    }
    
    return &Config{
        // Server
        Host: envutil.GetEnv("HOST", "0.0.0.0"),
        Port: envutil.GetEnvPort("PORT", 8080),
        
        // Database (required)
        DatabaseURL:    envutil.MustGetEnv("DATABASE_URL"),
        MaxConnections: envutil.GetEnvInt("DB_MAX_CONN", 25),
        Timeout:        envutil.GetEnvDuration("DB_TIMEOUT", 5*time.Second),
        
        // Features
        Debug:     envutil.GetEnvBool("DEBUG", false),
        RateLimit: envutil.GetEnvFloat64("RATE_LIMIT", 100.0),
    }, nil
}

func main() {
    config, err := LoadConfig()
    if err != nil {
        log.Fatal("Configuration error:", err)
    }
    
    log.Printf("Server: %s:%d", config.Host, config.Port)
    log.Printf("Database: %s (max: %d)", config.DatabaseURL, config.MaxConnections)
}
```

### Example 2: Using Client with Prefix

```go
package main

import (
    "fmt"
    envutil "github.com/isimtekin/go-packages/env-util"
)

func main() {
    // Create client with APP_ prefix
    client := envutil.NewWithOptions(
        envutil.WithPrefix("APP_"),
        envutil.WithEnvFile(".env"),
        envutil.WithRequired("NAME", "VERSION"),
    )
    
    // These will look for APP_NAME and APP_VERSION
    name := client.GetString("NAME", "Unknown")
    version := client.GetString("VERSION", "0.0.0")
    
    fmt.Printf("%s v%s\n", name, version)
}
```

### Example 3: Loading from .env File

```go
// .env file content:
// DATABASE_URL=postgres://user:pass@localhost/mydb
// REDIS_HOST=localhost
// REDIS_PORT=6379
// API_KEYS=key1,key2,key3
// DEBUG=true

func main() {
    // Load .env file
    if err := envutil.LoadEnvFile(".env"); err != nil {
        log.Printf("Warning: .env file not loaded: %v", err)
    }
    
    // Now use normally
    dbUrl := envutil.GetEnv("DATABASE_URL", "")
    redisHost := envutil.GetEnv("REDIS_HOST", "localhost")
    apiKeys := envutil.GetEnvStringSlice("API_KEYS", nil)
    
    fmt.Printf("Database: %s\n", dbUrl)
    fmt.Printf("Redis: %s\n", redisHost)
    fmt.Printf("API Keys: %v\n", apiKeys)
}
```

## 🔧 Duration Parsing Intelligence

The package intelligently parses durations based on the format and key name:

```go
// Standard Go duration format
envutil.GetEnvDuration("TIMEOUT", 0)  // "30s" → 30 seconds

// Integer with smart detection based on key name
REQUEST_TIMEOUT_MS=500  // → 500 milliseconds (detected from _MS or ms suffix)
CACHE_TTL_MIN=15       // → 15 minutes (detected from _MIN or min suffix)
DB_TIMEOUT_HOUR=2      // → 2 hours (detected from _HOUR or hour suffix)
TIMEOUT=30             // → 30 seconds (default for plain integers)

// Supported suffixes: _ms/ms, _us/us, _ns/ns, _min/min, _hour/hour
```

## 🎨 Boolean Value Formats

The package accepts multiple boolean formats for flexibility:

```go
// All these are TRUE:
"true", "TRUE", "True", "1", "yes", "YES", "Yes", "y", "Y"

// All these are FALSE:
"false", "FALSE", "False", "0", "no", "NO", "No", "n", "N"

// Any other value returns the default
```

## 🏗️ Advanced Client Features

### Client Methods

The Client type provides all the same functionality as standalone functions, plus additional features:

**All type getters available:**
- `client.GetString(key, default)` - String values
- `client.GetBool(key, default)` - Boolean values
- `client.GetInt(key, default)` - Integer values
- `client.GetInt64(key, default)` - Int64 values
- `client.GetFloat64(key, default)` - Float64 values
- `client.GetDuration(key, default)` - Duration values
- `client.GetStringSlice(key, default)` - String slices
- `client.GetIntSlice(key, default)` - Integer slices
- `client.GetURL(key, default)` - URL parsing
- `client.GetFilePath(key, default)` - File path validation
- `client.GetJSON(key, target)` - JSON unmarshaling

**Must methods (panic on missing):**
- `client.MustGetString(key)`
- `client.MustGetInt(key)`

**Utility methods:**
- `client.IsSet(key)` - Check if variable exists
- `client.SetEnv(key, value)` - Set variable (for testing)
- `client.UnsetEnv(key)` - Unset variable
- `client.ClearCache()` - Clear value cache
- `client.Export()` - Export all matching variables
- `client.LoadEnvFile(filename)` - Load .env file
- `client.ValidateRequired(keys)` - Validate required keys

### Caching

The client caches values for performance:

```go
client := envutil.NewDefault()

// First call reads from environment
val1 := client.GetString("KEY", "default")  // Reads from env

// Subsequent calls use cache
val2 := client.GetString("KEY", "default")  // Uses cached value

// Clear cache if needed
client.ClearCache()
```

### Prefix Namespacing

Use prefixes to avoid conflicts:

```go
// Different services with their own namespaces
apiClient := envutil.NewWithOptions(envutil.WithPrefix("API_"))
dbClient := envutil.NewWithOptions(envutil.WithPrefix("DB_"))

// API_HOST vs DB_HOST - no conflicts
apiHost := apiClient.GetString("HOST", "api.example.com")
dbHost := dbClient.GetString("HOST", "localhost")
```

### Required Variables Validation

Ensure critical variables exist:

```go
// Method 1: Panic on creation
client := envutil.NewWithOptions(
    envutil.WithRequired("DATABASE_URL", "API_KEY"),
)

// Method 2: Validate and handle error
err := envutil.ValidateRequired("DATABASE_URL", "API_KEY", "SECRET")
if err != nil {
    log.Fatal("Missing required configuration:", err)
}

// Method 3: Use Must* functions
apiKey := envutil.MustGetEnv("API_KEY")  // Panics if not set
```

## 📝 Testing Support

### Mock Environment in Tests

```go
func TestMyFunction(t *testing.T) {
    // Save original value
    original := os.Getenv("API_URL")
    
    // Set test value
    os.Setenv("API_URL", "http://test.local")
    defer os.Setenv("API_URL", original)
    
    // Run test
    result := MyFunction()
    assert.Equal(t, "expected", result)
}
```

### Using Client for Testing

```go
func TestWithClient(t *testing.T) {
    client := envutil.NewWithOptions(
        envutil.WithSilent(true),  // Disable logs in tests
    )
    
    // Set test values
    client.SetEnv("TEST_VAR", "test_value")
    
    // Test with known values
    val := client.GetString("TEST_VAR", "default")
    assert.Equal(t, "test_value", val)
}
```

## 🚨 Common Pitfalls & Solutions

### Pitfall 1: Wrong Type Conversion
```go
// BAD: Will return default on invalid input
port := envutil.GetEnvInt("PORT", 8080)  // "abc" → 8080

// GOOD: Validate critical values
port := envutil.MustGetEnvInt("PORT")  // Panics on invalid
```

### Pitfall 2: Missing Required Variables
```go
// BAD: Might fail at runtime
url := envutil.GetEnv("DATABASE_URL", "")

// GOOD: Fail fast
url := envutil.MustGetEnv("DATABASE_URL")
```

### Pitfall 3: Forgetting Prefix
```go
// BAD: Looking for wrong key
client := envutil.NewWithOptions(envutil.WithPrefix("APP_"))
host := os.Getenv("HOST")  // Wrong! Should use client

// GOOD: Use client methods
host := client.GetString("HOST", "localhost")  // Looks for APP_HOST
```

## 📊 Performance Considerations

- **Caching**: Client caches values after first read
- **Zero Dependencies**: No external libraries = faster builds
- **Lazy Loading**: .env files loaded only when specified
- **Efficient Parsing**: Type conversions only when requested

## 🔒 Security Best Practices

1. **Never commit .env files**: Add to `.gitignore`
2. **Use secrets management**: For production, use proper secret stores
3. **Validate inputs**: Use `Must*` functions for critical values
4. **Limit access**: Use minimal permissions for env files

## 📄 License

MIT License - See [LICENSE](LICENSE) file for details.

## 🤝 Contributing

Contributions welcome! Please feel free to submit a Pull Request.

## 🔗 Links

- [GitHub Repository](https://github.com/isimtekin/go-packages)
- [Package Directory](https://github.com/isimtekin/go-packages/tree/main/env-util)
- [Go Package Documentation](https://pkg.go.dev/github.com/isimtekin/go-packages/env-util) (after first release)
- [Report Issues](https://github.com/isimtekin/go-packages/issues)
- [Example Code](./examples/main.go)

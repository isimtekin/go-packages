# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Context

This is a **Go packages monorepo** containing multiple independent, reusable Go packages. Each package:
- Has its own `go.mod` and independent versioning
- Uses semantic versioning with format `package-name/vX.Y.Z`
- Can depend on other packages in the monorepo
- Is released and tagged independently

**Current Packages:**
- `env-util`: Environment variable loading and configuration utilities
- `mongo-client`: MongoDB client wrapper with auto-timestamps
- `redis-client`: Redis client wrapper with multi-database support

## Monorepo Structure

```
go-packages/
├── package-name/         # Each package in its own directory
│   ├── go.mod           # Independent module definition
│   ├── go.sum           # Dependencies
│   ├── *.go             # Source files
│   ├── *_test.go        # Test files
│   ├── examples/        # Example code (optional)
│   ├── README.md        # Package documentation
│   ├── Makefile         # Package-specific commands (optional)
│   └── .env.example     # Environment variable examples (optional)
├── scripts/             # Monorepo automation scripts
│   ├── create-package.sh # Package template generator
│   ├── release.sh       # Package release script
│   ├── bump-version.sh  # Version increment script
│   └── update-deps.sh   # Dependency update script
├── .github/workflows/   # CI/CD workflows
├── Makefile             # Monorepo-level commands
├── PACKAGE_TEMPLATE.md  # Package template documentation
├── CLAUDE.md            # This file
└── README.md            # Monorepo documentation
```

## Development Workflow

### Working on a Specific Package

```bash
# Navigate to the package directory
cd package-name

# Run package-specific tests
go test -v ./...
go test -v -race ./...

# Or use package Makefile if available
make test
make coverage

# Format code
go fmt ./...
# Or: make fmt

# Run linter (if golangci-lint installed)
golangci-lint run
# Or: make lint
```

### Creating a New Package

```bash
# From monorepo root
make create-package
# Enter package name when prompted

# Or use script directly
./scripts/create-package.sh my-new-package

# Quick start (create + initial release)
make quick-start
```

This generates a complete package structure with:
- Client implementation with connection management
- Configuration with validation and functional options
- Error definitions
- Unit tests
- Complete README and documentation
- .gitignore

See `PACKAGE_TEMPLATE.md` for template details.

### Releasing a Package

```bash
# From monorepo root

# Release specific version
./scripts/release.sh package-name 1.2.3

# Auto-increment version
./scripts/bump-version.sh package-name patch  # 1.0.0 → 1.0.1
./scripts/bump-version.sh package-name minor  # 1.0.1 → 1.1.0
./scripts/bump-version.sh package-name major  # 1.1.0 → 2.0.0

# Or use Makefile
make release PACKAGE=package-name VERSION=1.2.3
make bump PACKAGE=package-name TYPE=patch
```

### Monorepo Commands

```bash
# List all package versions
make list-versions

# Update dependencies for all packages
make update-deps

# Format all packages
make fmt-all

# Test all packages
make test-all

# Clean all packages
make clean-all
```

## Local Package Dependencies

Packages in this monorepo can depend on each other. Use the `replace` directive in `go.mod`:

```go
module github.com/isimtekin/go-packages/my-package

require (
    github.com/isimtekin/go-packages/env-util v1.0.0
)

// Use relative path for local development
replace github.com/isimtekin/go-packages/env-util => ../env-util
```

**Important:**
- Always use `replace` directive for local dependencies
- The version in `require` can be any valid semver (it's replaced)
- Path in `replace` must be relative to the package's directory

## Common Patterns Across Packages

### Configuration
Most packages follow this pattern:

```go
// Config struct with validation
type Config struct {
    // Fields...
}

func (c *Config) Validate() error {
    // Validation logic
}

// Default configuration
func DefaultConfig() *Config {
    return &Config{
        // Defaults...
    }
}
```

### Functional Options Pattern
```go
type Option func(*Config)

func WithField(value string) Option {
    return func(c *Config) {
        c.Field = value
    }
}

func NewWithOptions(opts ...Option) (*Client, error) {
    config := DefaultConfig()
    for _, opt := range opts {
        opt(config)
    }
    return New(config)
}
```

### Environment Variable Loading
Most packages use `env-util` for configuration:

```go
import envutil "github.com/isimtekin/go-packages/env-util"

func NewFromEnvWithDefaults(ctx context.Context) (*Client, error) {
    return NewFromEnv(ctx, "PREFIX_")
}

func NewFromEnv(ctx context.Context, prefix string) (*Client, error) {
    config := DefaultConfig()
    env := envutil.NewWithOptions(
        envutil.WithPrefix(prefix),
    )

    // Load from environment
    config.Field = env.GetString("FIELD", config.Field)

    return New(config)
}
```

### Testing
- All packages use standard Go testing
- Tests should not require external services by default
- Use table-driven tests where appropriate
- Include race detection: `go test -race`

### Documentation
- Each package has comprehensive README.md
- Code examples in `examples/` directory
- .env.example for environment variables
- Inline comments for exported functions

## Code Style and Conventions

1. **Error Handling**
   ```go
   // Wrap errors with context
   if err != nil {
       return nil, fmt.Errorf("operation failed: %w", err)
   }
   ```

2. **Context Usage**
   ```go
   // All operations accept context as first parameter
   func (c *Client) Operation(ctx context.Context, params...) error
   ```

3. **Naming**
   - Use clear, descriptive names
   - Follow Go naming conventions
   - Acronyms should be all caps: `ID`, `URL`, `HTTP`

4. **Package Organization**
   ```
   package-name/
   ├── client.go      # Main client/implementation
   ├── config.go      # Configuration
   ├── options.go     # Functional options
   ├── errors.go      # Error definitions
   ├── env.go         # Environment variable loading
   ├── helpers.go     # Utility functions
   ├── *_test.go      # Tests alongside implementation
   └── examples/      # Example programs
   ```

## Git Workflow

### Tagging and Releases
Tags use the format: `package-name/vX.Y.Z`

```bash
# Example tags
env-util/v1.0.0
mongo-client/v1.2.3
redis-client/v2.0.0
```

### Committing Changes
```bash
# Work in package directory
cd package-name

# Make changes
# ... edit files ...

# Test changes
make test

# Commit (from monorepo root)
cd ..
git add package-name/
git commit -m "package-name: description of changes"
```

### Branch Strategy
- `main`: stable, released code
- Feature branches: `feature/package-name-description`
- Fix branches: `fix/package-name-issue`

## Testing Strategy

### Unit Tests
```bash
# Run tests for specific package
cd package-name
go test -v ./...

# With race detection
go test -v -race ./...

# With coverage
go test -v -cover -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Integration Tests
Some packages have integration tests requiring external services:

```bash
# Start required services (if Makefile available)
make docker-up

# Run integration tests
make test-integration

# Or manually
go test -v -tags=integration ./...

# Stop services
make docker-down
```

## Environment Variables

Most packages support environment variable configuration:

**Common Patterns:**
- Prefix-based: `PACKAGE_FIELD` (e.g., `MONGO_URI`, `REDIS_ADDR`)
- Snake case for multi-word: `MAX_POOL_SIZE`
- Booleans: `"true"` or `"false"`
- Durations: Go duration format (`"5s"`, `"1m"`, `"2h"`)

**Example `.env` file:**
```bash
# Package prefix
PACKAGE_HOST=localhost
PACKAGE_PORT=27017
PACKAGE_DATABASE=mydb

# Connection settings
PACKAGE_MAX_POOL_SIZE=100
PACKAGE_TIMEOUT=5s
```

## Important Notes

1. **Independent Versioning**: Each package versions independently - breaking changes in one package don't affect others

2. **No Shared Dependencies**: Avoid creating tightly coupled packages - each should be independently usable

3. **Replace Directives**: Always use `replace` for local package dependencies in `go.mod`

4. **Examples**: Example code goes in `examples/` subdirectories to avoid "main redeclared" errors

5. **Tests Without Services**: Default `go test` should work without external services - use build tags for integration tests

6. **README.md**: Each package needs comprehensive documentation with usage examples

7. **Backward Compatibility**: Follow semantic versioning - don't break APIs in minor/patch releases

## Troubleshooting

### "Module not found" errors
```bash
# Ensure replace directive exists
cd package-name
grep "replace" go.mod

# Update dependencies
go mod tidy
```

### Tests requiring services
```bash
# Start services with Docker
make docker-up

# Or check examples/README.md for setup instructions
```

### Import path issues
```bash
# Always use full import path
import "github.com/isimtekin/go-packages/package-name"

# Not relative imports
# import "../package-name"  // ❌ Wrong
```

## Getting Help

- Check package README.md for specific documentation
- See PACKAGE_TEMPLATE.md for template structure
- Review examples/ directory for usage patterns
- Each package may have its own Makefile with helpful commands

## Quick Reference

```bash
# Create new package
make create-package

# Release package
./scripts/release.sh package-name 1.0.0

# Bump version
./scripts/bump-version.sh package-name patch

# Test all packages
make test-all

# Update all dependencies
make update-deps

# List all versions
make list-versions

# Format all code
make fmt-all
```

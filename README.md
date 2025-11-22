# Go Packages

Reusable Go packages monorepo for common functionalities.

## 📑 Table of Contents

- [Getting Started](#-getting-started)
- [Repository Structure](#-repository-structure)
- [Release Process](#-release-process)
- [Available Scripts](#-available-scripts)
- [Makefile Commands](#-makefile-commands)
- [GitHub Actions](#-github-actions-automated-releases)
- [Quick Commands](#-quick-commands)
- [Package List](#-package-list)
- [Development Workflow](#-development-workflow)
- [Important Configuration](#-important-configuration)
- [Version Guidelines](#-version-guidelines)
- [Maintenance](#-maintenance-commands)
- [Troubleshooting](#-troubleshooting)

## 🚀 Getting Started

### Prerequisites
- Go 1.21 or higher
- Git

### Initial Setup
```bash
# 1. Clone the repository
git clone https://github.com/isimtekin/go-packages.git
cd go-packages

# 2. Make scripts executable
chmod +x scripts/*.sh

# 3. Verify setup
make list-versions  # Should show empty or existing versions
```

### 🎯 Create Package Template

Create a fully-featured package with one command:

```bash
# Interactive creation
make create-package
# Enter package name: redis-client
when 
# Or direct script usage
./scripts/create-package.sh mongo-client

# Quick start (create + release v0.1.0)
make quick-start
```

This generates a complete package structure with:
- ✅ Client implementation with connection management
- ✅ Configuration with validation
- ✅ Functional options pattern
- ✅ Error definitions
- ✅ Unit tests
- ✅ Complete README
- ✅ .gitignore

See [PACKAGE_TEMPLATE.md](PACKAGE_TEMPLATE.md) for details.

## 📦 Repository Structure

```
go-packages/
├── package-name/         # Each package in its own directory
│   ├── go.mod           # Module definition
│   ├── go.sum           # Dependencies
│   └── README.md        # Package documentation
├── scripts/             # Automation scripts
│   ├── create-package.sh # Package template generator
│   ├── release.sh       # Package release script
│   ├── bump-version.sh  # Version increment script
│   └── update-deps.sh   # Dependency update script
├── .github/
│   └── workflows/
│       └── release.yml  # Automated GitHub release
├── Makefile             # Common tasks automation
├── PACKAGE_TEMPLATE.md  # Package template documentation
└── README.md            # This file
```

## 🚀 Creating a New Package

```bash
# Create package directory
mkdir my-package
cd my-package

# Initialize Go module
go mod init github.com/isimtekin/go-packages/my-package

# Create your package files
touch main.go README.md
```

## 🔖 Release Process

### Version Format

Each package uses independent semantic versioning:
- Format: `package-name/vX.Y.Z`
- Example: `logger/v1.0.0`, `redis-client/v2.1.3`

## 📜 Available Scripts

### 1. create-package.sh - Generate Package from Template (NEW!)

Creates a complete package structure with boilerplate code.

```bash
# Usage
./scripts/create-package.sh <package-name>

# Example
./scripts/create-package.sh redis-client
```

**What it creates:**
- Complete client implementation with connection management
- Configuration with validation
- Functional options pattern
- Error handling utilities
- Unit test structure
- Comprehensive README
- .gitignore file

### 2. release.sh - Create a New Release

Validates, tests, and releases a package with a specific version.

```bash
# Usage
./scripts/release.sh <package-name> <version>

# Examples
./scripts/release.sh mongo-client 1.0.0
./scripts/release.sh redis-client 2.1.3
```

**What it does:**
- Checks if package directory exists
- Verifies tag doesn't already exist
- Runs package tests
- Updates package README.md with new version
- Updates root README.md package table with new version
- Commits README changes
- Creates Git tag with format `package-name/vX.Y.Z`
- Pushes tag to GitHub
- Shows installation command for users

### 3. bump-version.sh - Auto-increment Version

Automatically increments version and triggers release.

```bash
# Usage
./scripts/bump-version.sh <package-name> <major|minor|patch>

# Examples
./scripts/bump-version.sh logger patch    # 1.0.0 → 1.0.1
./scripts/bump-version.sh logger minor    # 1.0.1 → 1.1.0
./scripts/bump-version.sh logger major    # 1.1.0 → 2.0.0
```

**What it does:**
- Finds latest version tag for the package
- Increments version based on type (major/minor/patch)
- Automatically calls release.sh with new version
- If no previous version exists, starts from v0.0.0

### 4. update-deps.sh - Update All Dependencies

Updates Go dependencies for all packages in the repository.

```bash
# Usage
./scripts/update-deps.sh

# No parameters needed - updates all packages
```

**What it does:**
- Finds all packages with go.mod files
- Updates all dependencies to latest versions
- Runs `go mod tidy` to clean up

## 📋 Makefile Commands

The repository includes an enhanced Makefile with many useful commands:

### Package Management
```bash
make help           # Show all available commands
make create-package # Create new package from template (interactive)
make quick-start    # Create package and release v0.1.0
make list-packages  # List all packages in repo
```

### Release Commands
```bash
make release        # Release a specific package (interactive)
make bump-version   # Auto-increment version (interactive)
make release-all    # Release all packages with same version
make list-versions  # List all package versions
```

### Development Commands
```bash
make test           # Test specific package (interactive)
make test-all       # Test all packages
make fmt            # Format all Go code
make lint           # Run linter on all packages
make clean          # Clean build artifacts and caches
make update-deps    # Update all dependencies
```

## 🚀 Quick Release Guide

### First Time Setup

```bash
# Clone repository
git clone https://github.com/isimtekin/go-packages.git
cd go-packages

# Make scripts executable
chmod +x scripts/*.sh
```

### Creating Your First Release

```bash
# Option 1: Manual version
./scripts/release.sh my-package 1.0.0

# Option 2: Auto-increment from v0.0.0
./scripts/bump-version.sh my-package minor  # Creates v0.1.0
```

### Subsequent Releases

```bash
# Bug fix release (v1.0.0 → v1.0.1)
./scripts/bump-version.sh my-package patch

# New feature release (v1.0.1 → v1.1.0)
./scripts/bump-version.sh my-package minor

# Breaking change release (v1.1.0 → v2.0.0)
./scripts/bump-version.sh my-package major
```

## 📋 Version Management

### List All Versions

```bash
# List all tags
git tag -l

# List specific package versions
git tag -l "logger/*"
git tag -l "redis-client/*"
```

### Delete a Version (if needed)

```bash
# Delete local tag
git tag -d package-name/v1.0.0

# Delete remote tag
git push origin --delete package-name/v1.0.0
```

## 🎯 Using Packages in Your Projects

### Install a Package

```bash
# Install latest version
go get github.com/isimtekin/go-packages/logger@latest

# Install specific version
go get github.com/isimtekin/go-packages/logger@v1.0.0

# Update package
go get -u github.com/isimtekin/go-packages/logger
```

### Import in Code

```go
import (
"github.com/isimtekin/go-packages/logger"
redisclient "github.com/isimtekin/go-packages/redis-client"
)
```

## 🔄 GitHub Actions (Automated Releases)

The repository includes automated release workflow that triggers when you push a version tag:

### Workflow: `.github/workflows/release.yml`

**Triggers on:** Push of tags matching `*/v*` pattern (e.g., `logger/v1.0.0`)

**What it does:**
1. Sets up Go 1.21 environment
2. Extracts package name and version from tag
3. Runs tests for the package
4. Creates GitHub Release automatically

### How It Works

When you run:
```bash
./scripts/release.sh logger 1.0.0
```

The script pushes tag `logger/v1.0.0`, which triggers the GitHub Action to:
- Run tests one more time in CI
- Create a GitHub Release with release notes
- Make the package available for download

### Manual Trigger
You can also manually create and push tags:
```bash
git tag -a "package-name/v1.0.0" -m "Release message"
git push origin "package-name/v1.0.0"
# GitHub Actions will automatically create the release
```

## 📝 Package List

Current packages available:

| Package | Description | Latest Version | Install |
|---------|-------------|----------------|---------|
| [env-util](./env-util) | Zero-dependency environment variable utilities with type safety | v0.0.2 | `go get github.com/isimtekin/go-packages/env-util@v0.0.2` |
| [mongo-client](./mongo-client) | High-level MongoDB client wrapper with CRUD helpers and transactions | v0.0.2 | `go get github.com/isimtekin/go-packages/mongo-client@v0.0.2` |
| [redis-client](./redis-client) | Redis client wrapper with multi-database support and connection pooling | v0.0.2 | `go get github.com/isimtekin/go-packages/redis-client@v0.0.2` |
| [nats-client](./nats-client) | NATS client wrapper with pub/sub, request/reply, and JetStream support | v0.0.2 | `go get github.com/isimtekin/go-packages/nats-client@v0.0.2` |
| [crypto-utils](./crypto-utils) | Comprehensive cryptography utilities with AES, RSA, ECDSA, ECDH, hashing, and key derivation | v0.0.1 | `go get github.com/isimtekin/go-packages/crypto-utils@v0.0.1` |
| [kafka-client](./kafka-client) | Robust Kafka client with producer, consumer, and admin operations | v0.1.0 | `go get github.com/isimtekin/go-packages/kafka-client@v0.1.0` |
| [slack-notifier](./slack-notifier) | Easy-to-use Slack webhook notifier with retry logic and Block Kit support | v0.0.1 | `go get github.com/isimtekin/go-packages/slack-notifier@v0.0.1` |
| [http-service](./http-service) | FastAPI-inspired HTTP framework with auto OpenAPI docs and request validation | v0.0.2 | `go get github.com/isimtekin/go-packages/http-service@v0.0.2` |
| [mail-sender](./mail-sender) | Flexible email sending library with SendGrid support, async sending, and template rendering | v0.1.0 | `go get github.com/isimtekin/go-packages/mail-sender@v0.1.0` |

**env-util** features:
- 🔒 Type-safe environment variable parsing (string, int, bool, duration, URL, etc.)
- 🎯 Zero dependencies - pure Go stdlib
- ⚙️ Functional options pattern with prefix support
- 📁 .env file loading
- ✅ Required variable validation
- 🚀 Smart duration parsing with unit detection
- 📋 Slice/list parsing from comma-separated values
- 💾 Value caching for performance

**mongo-client** features:
- 🚀 High-level CRUD operations with simple methods
- 🔄 Transaction support with automatic rollback
- ⚙️ Connection pooling and health checks
- ⏱️ Context management with automatic timeouts
- 📊 Aggregation pipeline helpers
- 🎯 Mongoose-like BaseModel with timestamps
- 🔧 Query and update builders
- 📋 Pagination support

**redis-client** features:
- 🔧 All Redis data types (strings, hashes, lists, sets, sorted sets)
- 🗄️ Multi-database support with singleton pattern
- 🏷️ Config-based database naming for readability
- 🔄 Connection pooling per database
- 🔐 TLS/SSL support
- 📊 Pipeline and transaction support
- 🎯 Pub/Sub messaging
- 🌐 Environment variable configuration

**nats-client** features:
- 📨 Publish/subscribe messaging patterns
- 🔄 Request/reply RPC support
- 👥 Queue groups for load balancing
- 🚀 JetStream support for persistence
- 🔒 TLS/SSL encryption
- 🔐 Username/password and token authentication
- 🔁 Automatic reconnection with backoff
- 🌐 Environment variable configuration

**crypto-utils** features:
- 🔐 AES-GCM and AES-CBC encryption (128/192/256-bit)
- 🔑 RSA key generation, OAEP encryption, PSS signing
- ✍️ ECDSA signing/verification (P-256/P-384/P-521)
- 🤝 ECDH key exchange (X25519, P-256, P-384, P-521)
- 🔨 SHA-256/384/512 hashing and HMAC-SHA256/512
- 🛡️ PBKDF2 key derivation (SHA-256/SHA-512)
- 🎲 Cryptographically secure random generation
- 🔐 Secure password and PIN generation
- 🆔 Short ID and secure token generation
- 📦 Base64 encoding/decoding (standard and URL-safe)

**kafka-client** features:
- 📨 Producer with single and batch message sending
- 📥 Consumer group support with automatic offset management
- 🔧 Admin operations (create/delete/list topics, metadata)
- 🔐 SASL authentication (PLAIN, SCRAM-SHA-256, SCRAM-SHA-512)
- 🔒 TLS/SSL encryption support
- 📦 Multiple compression codecs (Snappy, GZIP, LZ4, Zstd)
- 🎯 Partitioning strategies (hash, random, round-robin)
- ⚡ Idempotent writes for exactly-once semantics
- 🔄 Context-aware operations with timeout support
- 🧵 Thread-safe for concurrent use

**slack-notifier** features:
- 📨 Simple webhook-based Slack notifications
- 🎨 Color-coded messages (success, warning, error, info)
- 🏗️ Message builder pattern for fluent API
- 🧱 Block Kit support for rich, interactive messages
- 🔄 Automatic retry logic with exponential backoff
- ⏱️ Context-aware operations with timeout support
- 🧵 Thread support for organized conversations
- ⚙️ Customizable username, icon, and channel
- 📎 Attachment support with fields and formatting
- ✅ 81.8% test coverage with comprehensive tests
- 🔒 Type-safe message structures

**http-service** features:
- 🚀 FastHTTP-powered for maximum performance
- 📝 Automatic OpenAPI 3.0 spec generation + Swagger UI
- ✅ Integrated request validation with go-playground/validator
- 🎯 Type-safe handlers with generics support
- 🔧 Built-in middleware (CORS, logging, recovery, rate limiting)
- 🏗️ Builder pattern for service and route configuration
- ⚡ Full context.Context integration for cancellation
- 🔒 Production-ready with graceful shutdown and panic recovery
- 📊 Built-in /health endpoint, optional /metrics
- 🎨 FastAPI-inspired developer experience
- 🔐 Authentication and authorization middleware support

**mail-sender** features:
- 📧 Multi-provider email sending (SendGrid supported, extensible for more)
- ⚡ Async/non-blocking email sending with worker pools
- 🎯 Event-based architecture (OnSuccess, OnFailure, OnRetry)
- 🔄 Automatic retry logic with configurable attempts and delays
- 📝 HTML and plain text template rendering with Go templates
- 🎨 Multiple recipients support (To, Cc, Bcc)
- 👥 Worker pool for concurrent email sending
- 📊 Real-time statistics (sent, failed, pending, retried)
- 🛡️ Graceful shutdown with timeout support
- ⚙️ Flexible configuration (code, functional options, or env vars)
- 🔐 Environment variable configuration support
- ✅ 91.5% test coverage with comprehensive tests

## 🔄 Development Workflow

### 1. Start New Package

```bash
# Create and setup
mkdir my-new-package
cd my-new-package
go mod init github.com/isimtekin/go-packages/my-new-package

# Add your code
echo "package mynewpackage" > client.go

# Create README
echo "# My New Package" > README.md
```

### 2. Development Cycle

```bash
# Make changes
git add .
git commit -m "feat: add awesome feature"
git push origin main

# Test locally
go test ./...

# Update dependencies if needed
cd ..
./scripts/update-deps.sh
```

### 3. Release Package

```bash
# First release
./scripts/release.sh my-new-package 0.1.0

# Future releases - patch fix
./scripts/bump-version.sh my-new-package patch

# Future releases - new feature
./scripts/bump-version.sh my-new-package minor

# Future releases - breaking change
./scripts/bump-version.sh my-new-package major
```

### 4. Using Released Package

In your other Go projects:

```go
// go.mod
require github.com/isimtekin/go-packages/my-new-package v0.1.0

// main.go
import "github.com/isimtekin/go-packages/my-new-package"
```

## ⚠️ Important Configuration

**Repository URL in Scripts:**
The `release.sh` script has a reference that needs updating (line 44):

```bash
# Edit scripts/release.sh and change:
echo "go get github.com/finbyte/go-packages/${PACKAGE}@v${VERSION}"

# To your repository:
echo "go get github.com/isimtekin/go-packages/${PACKAGE}@v${VERSION}"
```

Or use sed to fix it:
```bash
sed -i 's/finbyte/isimtekin/g' scripts/release.sh
```

## 🛠️ Quick Commands

### Using Makefile (Recommended)
```bash
# Clone and setup
git clone https://github.com/isimtekin/go-packages.git
cd go-packages
chmod +x scripts/*.sh

# Interactive release
make release

# List all versions
make list-versions

# Test everything
make test-all

# Update all dependencies
./scripts/update-deps.sh
```

### Using Scripts Directly
```bash
# Create new package
mkdir new-package && cd new-package
go mod init github.com/isimtekin/go-packages/new-package
cd ..

# Release with version bump
./scripts/bump-version.sh new-package minor

# Or release with specific version
./scripts/release.sh new-package 1.0.0

# Install in your project
go get github.com/isimtekin/go-packages/new-package@v1.0.0
```

## 📊 Version Guidelines

- **Major (X.0.0)**: Breaking changes
- **Minor (0.X.0)**: New features, backward compatible
- **Patch (0.0.X)**: Bug fixes, backward compatible

Examples:
- `v1.0.0` → `v2.0.0`: Breaking API changes
- `v1.0.0` → `v1.1.0`: New feature added
- `v1.0.0` → `v1.0.1`: Bug fix

## 🔧 Maintenance Commands

```bash
# List all package versions
git tag -l | sort -V

# List versions for specific package
git tag -l "logger/*" | sort -V

# Check latest version of a package
git tag -l "logger/*" | sort -V | tail -1

# Delete a wrong release (be careful!)
git tag -d package-name/v1.0.0              # Delete local
git push origin --delete package-name/v1.0.0 # Delete remote

# Update all package dependencies
./scripts/update-deps.sh

# Test all packages
for dir in $(find . -name go.mod -not -path "./scripts/*" -exec dirname {} \;); do
    echo "Testing $dir"
    (cd $dir && go test ./...)
done
```

## ❓ Troubleshooting

### "Package does not exist" Error
- Make sure package directory exists
- Check you're in the repository root when running scripts

### "Tag already exists" Error
- Check existing tags: `git tag -l "package-name/*"`
- Use a different version number
- Or delete the existing tag if it was a mistake

### Tests Failing on Release
- Fix the tests before releasing
- Run tests manually: `cd package-name && go test ./...`
- Skip tests temporarily (not recommended): Remove test section from release.sh

### Permission Denied on Scripts
```bash
chmod +x scripts/*.sh
```

### Module Not Found After Release
- Wait a few minutes for Go proxy to update
- Force refresh: `GOPROXY=direct go get github.com/isimtekin/go-packages/package@version`

## 📄 License

MIT License

## 👤 Author

**Ersin Işımtekin**
- GitHub: [@isimtekin](https://github.com/isimtekin)

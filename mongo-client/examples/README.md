# MongoDB Client Examples

This directory contains examples demonstrating how to use the mongo-client package.

## Available Examples

### 1. Basic Example (`basic/main.go`)

Comprehensive example showing all major features:
- Creating clients with options
- CRUD operations with auto-timestamps
- Advanced queries and operators
- Aggregation pipelines
- Transactions
- Pagination
- Health checks
- Index management

**Run:**
```bash
cd basic
go run main.go
```

**Requirements:**
- MongoDB running on localhost:27017

### 2. Environment Configuration Example (`env-config/main.go`)

Demonstrates loading configuration from environment variables:
- Using default MONGO_ prefix
- Custom prefixes (DB_, MYAPP_, etc.)
- Building URI from components
- Loading from .env files
- Different deployment patterns

**Run:**
```bash
cd env-config
go run main.go
```

**No MongoDB required** - this example shows configuration patterns without connecting.

## Quick Start

### Run Basic Example with Docker

```bash
# Start MongoDB
docker run -d --name mongo-test -p 27017:27017 mongo:latest

# Run basic example
cd basic
go run main.go

# Cleanup
docker stop mongo-test && docker rm mongo-test
```

### Run with Environment Variables

```bash
# Set environment variables
export MONGO_URI=mongodb://localhost:27017
export MONGO_DATABASE=exampledb

# Run basic example (it will use env vars if uncommented)
cd basic
go run main.go
```

### Using .env File

```bash
# Create .env file
cat > .env << EOF
MONGO_URI=mongodb://localhost:27017
MONGO_DATABASE=exampledb
EOF

# Modify basic/main.go to uncomment env loading
# Then run
cd basic
go run main.go
```

## Example Structure

```
examples/
├── README.md                 # This file
├── basic/
│   └── main.go              # Comprehensive feature examples
└── env-config/
    └── main.go              # Environment configuration examples
```

## Tips

1. **Start with env-config** - Learn configuration patterns first
2. **Then try basic** - See all features in action
3. **Check the code** - Examples are well-commented
4. **Experiment** - Modify examples to test different scenarios

## Common Issues

**"Failed to connect" error:**
- Make sure MongoDB is running on localhost:27017
- Or set MONGO_URI environment variable to your MongoDB instance

**"No such file or directory" error:**
- Make sure you're in the correct directory
- Use `cd basic` or `cd env-config` before running

**Build errors:**
- Run `go mod tidy` in the parent mongo-client directory
- Make sure all dependencies are installed

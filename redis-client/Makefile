.PHONY: test coverage lint fmt clean run-example help docker-up docker-down

# Default target
help:
	@echo "Available commands:"
	@echo "  make test            - Run all tests"
	@echo "  make coverage        - Run tests with coverage report"
	@echo "  make lint            - Run linter"
	@echo "  make fmt             - Format code"
	@echo "  make clean           - Clean build artifacts"
	@echo "  make run-example     - Run multi-db example"
	@echo "  make docker-up       - Start Redis in Docker"
	@echo "  make docker-down     - Stop Redis in Docker"

# Run tests
test:
	go test -v -race ./...

# Run tests with coverage
coverage:
	go test -v -race -cover -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linter (requires golangci-lint)
lint:
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Format code
fmt:
	go fmt ./...
	@if command -v goimports > /dev/null; then \
		goimports -w .; \
	fi

# Clean build artifacts
clean:
	rm -f coverage.out coverage.html
	go clean -cache

# Run multi-db example (requires Redis running)
run-example:
	@echo "Running multi-database example..."
	@cd examples/multi-db && go run main.go

# Install dependencies
deps:
	go mod download
	go mod tidy

# Start Redis in Docker
docker-up:
	docker run -d --name redis-test -p 6379:6379 redis:latest
	@echo "Redis started on port 6379"

# Stop Redis in Docker
docker-down:
	docker stop redis-test || true
	docker rm redis-test || true
	@echo "Redis stopped"

# Run integration tests (requires Redis)
test-integration: docker-up
	@sleep 2  # Wait for Redis to be ready
	go test -v -race ./...
	@make docker-down

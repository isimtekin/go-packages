.PHONY: test coverage lint fmt clean run-example help

# Default target
help:
	@echo "Available commands:"
	@echo "  make test         - Run all tests"
	@echo "  make coverage     - Run tests with coverage report"
	@echo "  make lint         - Run linter"
	@echo "  make fmt          - Format code"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make run-example  - Run example application"
	@echo "  make docker-up    - Start MongoDB in Docker"
	@echo "  make docker-down  - Stop MongoDB in Docker"

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

# Run example (requires MongoDB running)
run-example:
	@cd examples && go run main.go

# Install dependencies
deps:
	go mod download
	go mod tidy

# Start MongoDB in Docker
docker-up:
	docker run -d --name mongo-test -p 27017:27017 mongo:latest
	@echo "MongoDB started on port 27017"

# Stop MongoDB in Docker
docker-down:
	docker stop mongo-test || true
	docker rm mongo-test || true
	@echo "MongoDB stopped"

# Run integration tests (requires MongoDB)
test-integration: docker-up
	@sleep 2  # Wait for MongoDB to be ready
	go test -v -race ./...
	@make docker-down

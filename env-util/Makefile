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
	goimports -w .

# Clean build artifacts
clean:
	rm -f coverage.out coverage.html
	go clean -cache

# Run example
run-example:
	@cd examples && go run main.go

# Install dependencies
deps:
	go mod download
	go mod tidy

# Run benchmarks
bench:
	go test -bench=. -benchmem ./...
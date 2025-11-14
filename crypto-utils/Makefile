.PHONY: help test test-verbose test-race test-coverage bench clean fmt lint vet tidy install-tools

# Default target
.DEFAULT_GOAL := help

# Go parameters
GOCMD=go
GOTEST=$(GOCMD) test
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet
GOMOD=$(GOCMD) mod
GOBUILD=$(GOCMD) build
GOINSTALL=$(GOCMD) install

# Package name
PACKAGE=github.com/isimtekin/go-packages/crypto-utils

## help: Display this help message
help:
	@echo "crypto-utils Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make <target>"
	@echo ""
	@echo "Targets:"
	@awk '/^##/ { \
		split($$0, arr, ":"); \
		target = substr(arr[1], 4); \
		desc = substr($$0, index($$0, ":")+2); \
		printf "  \033[36m%-20s\033[0m %s\n", target, desc \
	}' $(MAKEFILE_LIST)

## test: Run all tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

## test-short: Run tests in short mode
test-short:
	@echo "Running tests in short mode..."
	$(GOTEST) -short ./...

## test-verbose: Run tests with verbose output
test-verbose:
	@echo "Running tests with verbose output..."
	$(GOTEST) -v -count=1 ./...

## test-race: Run tests with race detector
test-race:
	@echo "Running tests with race detector..."
	$(GOTEST) -race -v ./...

## test-coverage: Run tests with coverage report
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -cover -coverprofile=coverage.out ./...
	@echo ""
	@echo "Coverage summary:"
	$(GOCMD) tool cover -func=coverage.out | tail -1
	@echo ""
	@echo "To view detailed coverage report in browser:"
	@echo "  make coverage-html"

## coverage-html: Generate HTML coverage report
coverage-html: test-coverage
	@echo "Generating HTML coverage report..."
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Opening coverage report in browser..."
	@if command -v open > /dev/null; then \
		open coverage.html; \
	elif command -v xdg-open > /dev/null; then \
		xdg-open coverage.html; \
	else \
		echo "Please open coverage.html in your browser"; \
	fi

## bench: Run benchmarks
bench:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

## bench-cpu: Run CPU benchmarks
bench-cpu:
	@echo "Running CPU benchmarks..."
	$(GOTEST) -bench=. -benchmem -cpuprofile=cpu.prof ./...
	@echo "To analyze CPU profile: go tool pprof cpu.prof"

## bench-mem: Run memory benchmarks
bench-mem:
	@echo "Running memory benchmarks..."
	$(GOTEST) -bench=. -benchmem -memprofile=mem.prof ./...
	@echo "To analyze memory profile: go tool pprof mem.prof"

## fmt: Format Go code
fmt:
	@echo "Formatting Go code..."
	$(GOFMT) ./...
	@echo "Code formatted successfully!"

## vet: Run go vet
vet:
	@echo "Running go vet..."
	$(GOVET) ./...
	@echo "Vet check passed!"

## lint: Run golangci-lint (requires installation)
lint:
	@if command -v golangci-lint > /dev/null; then \
		echo "Running golangci-lint..."; \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not found. Install it with:"; \
		echo "  make install-tools"; \
	fi

## tidy: Tidy Go modules
tidy:
	@echo "Tidying Go modules..."
	$(GOMOD) tidy
	@echo "Modules tidied successfully!"

## download: Download dependencies
download:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	@echo "Dependencies downloaded!"

## verify: Verify dependencies
verify:
	@echo "Verifying dependencies..."
	$(GOMOD) verify
	@echo "Dependencies verified!"

## clean: Clean build artifacts and cache
clean:
	@echo "Cleaning build artifacts and cache..."
	$(GOCMD) clean
	rm -f coverage.out coverage.html
	rm -f cpu.prof mem.prof
	rm -f *.test
	@echo "Cleaned successfully!"

## check: Run all checks (fmt, vet, test)
check: fmt vet test
	@echo "All checks passed!"

## ci: Run CI checks (fmt, vet, lint, test with coverage)
ci: fmt vet test-race test-coverage
	@echo "CI checks completed!"

## install-tools: Install development tools
install-tools:
	@echo "Installing development tools..."
	@echo "Installing golangci-lint..."
	@if ! command -v golangci-lint > /dev/null; then \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin; \
	else \
		echo "golangci-lint already installed"; \
	fi
	@echo "Tools installed successfully!"

## update-deps: Update all dependencies
update-deps:
	@echo "Updating dependencies..."
	$(GOCMD) get -u ./...
	$(GOMOD) tidy
	@echo "Dependencies updated!"

## audit: Run security audit
audit:
	@echo "Running security audit..."
	@if command -v gosec > /dev/null; then \
		gosec ./...; \
	else \
		echo "gosec not found. Install it with: go install github.com/securego/gosec/v2/cmd/gosec@latest"; \
	fi

## stats: Show package statistics
stats:
	@echo "Package Statistics:"
	@echo ""
	@echo "Go files:"
	@find . -name "*.go" -not -path "./vendor/*" | wc -l
	@echo ""
	@echo "Test files:"
	@find . -name "*_test.go" -not -path "./vendor/*" | wc -l
	@echo ""
	@echo "Lines of code (excluding tests):"
	@find . -name "*.go" -not -name "*_test.go" -not -path "./vendor/*" -exec wc -l {} \; | awk '{sum+=$$1} END {print sum}'
	@echo ""
	@echo "Lines of test code:"
	@find . -name "*_test.go" -not -path "./vendor/*" -exec wc -l {} \; | awk '{sum+=$$1} END {print sum}'
	@echo ""
	@echo "Test coverage:"
	@$(GOTEST) -cover ./... 2>&1 | grep "coverage:" | tail -1

## watch: Watch for changes and run tests (requires entr)
watch:
	@if command -v entr > /dev/null; then \
		echo "Watching for changes..."; \
		find . -name "*.go" | entr -c make test; \
	else \
		echo "entr not found. Install it with your package manager"; \
		echo "  macOS: brew install entr"; \
		echo "  Linux: apt-get install entr"; \
	fi

## example: Run example usage
example:
	@echo "Example: AES-GCM Encryption"
	@echo ""
	@echo 'package main' > /tmp/crypto_example.go
	@echo '' >> /tmp/crypto_example.go
	@echo 'import (' >> /tmp/crypto_example.go
	@echo '    "fmt"' >> /tmp/crypto_example.go
	@echo '    cryptoutils "github.com/isimtekin/go-packages/crypto-utils"' >> /tmp/crypto_example.go
	@echo ')' >> /tmp/crypto_example.go
	@echo '' >> /tmp/crypto_example.go
	@echo 'func main() {' >> /tmp/crypto_example.go
	@echo '    key, _ := cryptoutils.GenerateRandomBytes(32)' >> /tmp/crypto_example.go
	@echo '    plaintext := []byte("Secret message")' >> /tmp/crypto_example.go
	@echo '    ciphertext, _ := cryptoutils.EncryptAESGCM(key, plaintext)' >> /tmp/crypto_example.go
	@echo '    decrypted, _ := cryptoutils.DecryptAESGCM(key, ciphertext)' >> /tmp/crypto_example.go
	@echo '    fmt.Printf("Decrypted: %s\\n", decrypted)' >> /tmp/crypto_example.go
	@echo '}' >> /tmp/crypto_example.go
	@$(GOCMD) run /tmp/crypto_example.go
	@rm /tmp/crypto_example.go

## version: Show Go version
version:
	@echo "Go version:"
	@$(GOCMD) version
	@echo ""
	@echo "Module version:"
	@grep "^module" go.mod
	@echo ""
	@echo "Go version requirement:"
	@grep "^go" go.mod

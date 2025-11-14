# Makefile in root directory

.PHONY: help create-package release bump-version list-versions test-all release-all update-deps
.PHONY: release-pkg bump-pkg test-pkg fmt-pkg clean-pkg lint-pkg

# Default target - show help
help:
	@echo "Go Packages Monorepo - Makefile Commands"
	@echo ""
	@echo "Interactive Commands:"
	@echo "  make create-package    - Create a new package from template"
	@echo "  make release          - Release a specific package (interactive)"
	@echo "  make bump-version     - Auto-increment version and release (interactive)"
	@echo "  make test             - Test a specific package (interactive)"
	@echo ""
	@echo "Direct Commands (with PACKAGE parameter):"
	@echo "  make release-pkg PACKAGE=crypto-utils VERSION=1.0.0"
	@echo "  make bump-pkg PACKAGE=crypto-utils TYPE=patch"
	@echo "  make test-pkg PACKAGE=crypto-utils"
	@echo "  make fmt-pkg PACKAGE=crypto-utils"
	@echo "  make clean-pkg PACKAGE=crypto-utils"
	@echo "  make lint-pkg PACKAGE=crypto-utils"
	@echo ""
	@echo "Global Commands:"
	@echo "  make list-versions    - List all package versions"
	@echo "  make list-packages    - List all available packages"
	@echo "  make test-all         - Test all packages"
	@echo "  make fmt              - Format all packages"
	@echo "  make clean            - Clean all packages"
	@echo "  make lint             - Lint all packages"
	@echo "  make release-all      - Release all packages with same version"
	@echo "  make update-deps      - Update dependencies for all packages"
	@echo ""
	@echo "Examples:"
	@echo "  make release-pkg PACKAGE=crypto-utils VERSION=1.0.0"
	@echo "  make bump-pkg PACKAGE=nats-client TYPE=minor"
	@echo "  make test-pkg PACKAGE=redis-client"

# Create a new package from template
create-package:
	@read -p "Enter package name (e.g., redis-client): " package; \
	./scripts/create-package.sh $$package

# Release a specific package
release:
	@read -p "Package name: " package; \
	read -p "Version (x.y.z): " version; \
	./scripts/release.sh $$package $$version

# Auto-increment version and release
bump-version:
	@read -p "Package name: " package; \
	read -p "Bump type (major/minor/patch): " bump_type; \
	./scripts/bump-version.sh $$package $$bump_type

# List all versions
list-versions:
	@echo "=== Package Versions ==="
	@for tag in $$(git tag | sort -V); do \
		echo $$tag; \
	done

# Test all packages
test-all:
	@for dir in $$(find . -name go.mod -not -path "./scripts/*" -exec dirname {} \;); do \
		echo "Testing $$dir..."; \
		(cd $$dir && go test ./...); \
	done

# Test a specific package
test:
	@read -p "Package name: " package; \
	@if [ -d "$$package" ]; then \
		echo "Testing $$package..."; \
		(cd $$package && go test -v -cover ./...); \
	else \
		echo "Package $$package not found"; \
	fi

# Release multiple packages at once
release-all:
	@read -p "Version (x.y.z): " version; \
	for dir in $$(find . -name go.mod -not -path "./scripts/*" -exec dirname {} \;); do \
		package=$$(basename $$dir); \
		./scripts/release.sh $$package $$version; \
	done

# Update all dependencies
update-deps:
	@./scripts/update-deps.sh

# List all packages
list-packages:
	@echo "=== Available Packages ==="
	@for dir in $$(find . -name go.mod -not -path "./scripts/*" -exec dirname {} \;); do \
		package=$$(basename $$dir); \
		echo "  - $$package"; \
	done

# Clean all test caches and build artifacts
clean:
	@for dir in $$(find . -name go.mod -not -path "./scripts/*" -exec dirname {} \;); do \
		echo "Cleaning $$dir..."; \
		(cd $$dir && go clean -cache -testcache); \
	done

# Format all Go code
fmt:
	@for dir in $$(find . -name go.mod -not -path "./scripts/*" -exec dirname {} \;); do \
		echo "Formatting $$dir..."; \
		(cd $$dir && go fmt ./...); \
	done

# Run linter on all packages (requires golangci-lint)
lint:
	@for dir in $$(find . -name go.mod -not -path "./scripts/*" -exec dirname {} \;); do \
		echo "Linting $$dir..."; \
		(cd $$dir && golangci-lint run); \
	done

# Quick package creation and first release
quick-start:
	@read -p "Enter package name (e.g., redis-client): " package; \
	./scripts/create-package.sh $$package && \
	echo "" && \
	echo "Package created! Now releasing v0.1.0..." && \
	./scripts/release.sh $$package 0.1.0

# ============================================================================
# Direct commands with PACKAGE parameter
# ============================================================================

# Release a specific package directly
# Usage: make release-pkg PACKAGE=crypto-utils VERSION=1.0.0
release-pkg:
ifndef PACKAGE
	@echo "Error: PACKAGE parameter is required"
	@echo "Usage: make release-pkg PACKAGE=crypto-utils VERSION=1.0.0"
	@exit 1
endif
ifndef VERSION
	@echo "Error: VERSION parameter is required"
	@echo "Usage: make release-pkg PACKAGE=crypto-utils VERSION=1.0.0"
	@exit 1
endif
	@if [ ! -d "$(PACKAGE)" ]; then \
		echo "Error: Package $(PACKAGE) not found"; \
		exit 1; \
	fi
	@echo "Releasing $(PACKAGE) version $(VERSION)..."
	@./scripts/release.sh $(PACKAGE) $(VERSION)

# Bump version for a specific package directly
# Usage: make bump-pkg PACKAGE=crypto-utils TYPE=patch
bump-pkg:
ifndef PACKAGE
	@echo "Error: PACKAGE parameter is required"
	@echo "Usage: make bump-pkg PACKAGE=crypto-utils TYPE=patch"
	@exit 1
endif
ifndef TYPE
	@echo "Error: TYPE parameter is required (major/minor/patch)"
	@echo "Usage: make bump-pkg PACKAGE=crypto-utils TYPE=patch"
	@exit 1
endif
	@if [ ! -d "$(PACKAGE)" ]; then \
		echo "Error: Package $(PACKAGE) not found"; \
		exit 1; \
	fi
	@echo "Bumping $(PACKAGE) version ($(TYPE))..."
	@./scripts/bump-version.sh $(PACKAGE) $(TYPE)

# Test a specific package directly
# Usage: make test-pkg PACKAGE=crypto-utils
test-pkg:
ifndef PACKAGE
	@echo "Error: PACKAGE parameter is required"
	@echo "Usage: make test-pkg PACKAGE=crypto-utils"
	@exit 1
endif
	@if [ ! -d "$(PACKAGE)" ]; then \
		echo "Error: Package $(PACKAGE) not found"; \
		exit 1; \
	fi
	@echo "Testing $(PACKAGE)..."
	@cd $(PACKAGE) && go test -v -cover ./...

# Format a specific package directly
# Usage: make fmt-pkg PACKAGE=crypto-utils
fmt-pkg:
ifndef PACKAGE
	@echo "Error: PACKAGE parameter is required"
	@echo "Usage: make fmt-pkg PACKAGE=crypto-utils"
	@exit 1
endif
	@if [ ! -d "$(PACKAGE)" ]; then \
		echo "Error: Package $(PACKAGE) not found"; \
		exit 1; \
	fi
	@echo "Formatting $(PACKAGE)..."
	@cd $(PACKAGE) && go fmt ./...

# Clean a specific package directly
# Usage: make clean-pkg PACKAGE=crypto-utils
clean-pkg:
ifndef PACKAGE
	@echo "Error: PACKAGE parameter is required"
	@echo "Usage: make clean-pkg PACKAGE=crypto-utils"
	@exit 1
endif
	@if [ ! -d "$(PACKAGE)" ]; then \
		echo "Error: Package $(PACKAGE) not found"; \
		exit 1; \
	fi
	@echo "Cleaning $(PACKAGE)..."
	@cd $(PACKAGE) && go clean -cache -testcache

# Lint a specific package directly
# Usage: make lint-pkg PACKAGE=crypto-utils
lint-pkg:
ifndef PACKAGE
	@echo "Error: PACKAGE parameter is required"
	@echo "Usage: make lint-pkg PACKAGE=crypto-utils"
	@exit 1
endif
	@if [ ! -d "$(PACKAGE)" ]; then \
		echo "Error: Package $(PACKAGE) not found"; \
		exit 1; \
	fi
	@echo "Linting $(PACKAGE)..."
	@if command -v golangci-lint > /dev/null; then \
		cd $(PACKAGE) && golangci-lint run; \
	else \
		echo "golangci-lint not found. Install it with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Run package-specific Makefile target
# Usage: make pkg-make PACKAGE=crypto-utils TARGET=test-coverage
pkg-make:
ifndef PACKAGE
	@echo "Error: PACKAGE parameter is required"
	@echo "Usage: make pkg-make PACKAGE=crypto-utils TARGET=test-coverage"
	@exit 1
endif
ifndef TARGET
	@echo "Error: TARGET parameter is required"
	@echo "Usage: make pkg-make PACKAGE=crypto-utils TARGET=test-coverage"
	@exit 1
endif
	@if [ ! -d "$(PACKAGE)" ]; then \
		echo "Error: Package $(PACKAGE) not found"; \
		exit 1; \
	fi
	@if [ ! -f "$(PACKAGE)/Makefile" ]; then \
		echo "Error: $(PACKAGE)/Makefile not found"; \
		exit 1; \
	fi
	@echo "Running 'make $(TARGET)' in $(PACKAGE)..."
	@cd $(PACKAGE) && make $(TARGET)

# Show stats for a specific package
# Usage: make stats-pkg PACKAGE=crypto-utils
stats-pkg:
ifndef PACKAGE
	@echo "Error: PACKAGE parameter is required"
	@echo "Usage: make stats-pkg PACKAGE=crypto-utils"
	@exit 1
endif
	@if [ ! -d "$(PACKAGE)" ]; then \
		echo "Error: Package $(PACKAGE) not found"; \
		exit 1; \
	fi
	@echo "Statistics for $(PACKAGE):"
	@echo ""
	@echo "Go files:"
	@find $(PACKAGE) -name "*.go" -not -path "*/vendor/*" | wc -l
	@echo ""
	@echo "Test files:"
	@find $(PACKAGE) -name "*_test.go" -not -path "*/vendor/*" | wc -l
	@echo ""
	@echo "Lines of code (excluding tests):"
	@find $(PACKAGE) -name "*.go" -not -name "*_test.go" -not -path "*/vendor/*" -exec wc -l {} \; | awk '{sum+=$$1} END {print sum}'
	@echo ""
	@echo "Lines of test code:"
	@find $(PACKAGE) -name "*_test.go" -not -path "*/vendor/*" -exec wc -l {} \; | awk '{sum+=$$1} END {print sum}'
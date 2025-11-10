# Makefile in root directory

.PHONY: help create-package release bump-version list-versions test-all release-all update-deps

# Default target - show help
help:
	@echo "Available commands:"
	@echo "  make create-package  - Create a new package from template"
	@echo "  make release        - Release a specific package"
	@echo "  make bump-version   - Auto-increment version and release"
	@echo "  make list-versions  - List all package versions"
	@echo "  make test-all       - Test all packages"
	@echo "  make release-all    - Release all packages with same version"
	@echo "  make update-deps    - Update dependencies for all packages"
	@echo ""
	@echo "Quick start:"
	@echo "  1. make create-package  (to create new package)"
	@echo "  2. make test-all        (to test everything)"
	@echo "  3. make release         (to release a package)"

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
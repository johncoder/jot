.PHONY: help setup build build-all clean test test-unit test-integration lint fmt vet staticcheck install deps release info

# Build configuration
APP_NAME := jot
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build flags
LDFLAGS := -s -w
LDFLAGS += -X github.com/johncoder/jot/cmd.version=$(VERSION)
LDFLAGS += -X github.com/johncoder/jot/cmd.buildTime=$(BUILD_TIME)
LDFLAGS += -X github.com/johncoder/jot/cmd.gitCommit=$(GIT_COMMIT)

# Go configuration
GO_VERSION := 1.24
COVERAGE_OUT := coverage.out

# Default target
help: ## Show this help message
	@echo "Available targets:"
	@echo ""
	@echo "Development:"
	@grep -E '^(setup|test|lint|fmt|vet|staticcheck):.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'
	@echo ""
	@echo "Building:"
	@grep -E '^(build|build-all|install|clean):.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'
	@echo ""
	@echo "Utilities:"
	@grep -E '^(deps|release|info):.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

setup: ## Setup development environment (non-interactive)
	@echo "Setting up development environment..."
	@echo "For interactive setup with detailed guidance, use: ./scripts/setup.sh"
	@echo "For troubleshooting help, use: ./scripts/setup.sh --help"
	@echo ""
	@if ! ./scripts/setup.sh --yes; then echo "Setup failed - try ./scripts/setup.sh --troubleshoot for help"; exit 1; fi

build: ## Build for current platform
	@echo "Building $(APP_NAME) v$(VERSION)..."
	@if ! go build -ldflags "$(LDFLAGS)" -o $(APP_NAME) .; then echo "Build failed"; exit 1; fi
	@echo "Build complete: ./$(APP_NAME)"

build-all: ## Build for all platforms
	@echo "Building for all platforms..."
	@if ! ./build.sh --all; then echo "Multi-platform build failed"; exit 1; fi

clean: ## Clean build artifacts and temporary files
	@if ! ./scripts/clean.sh; then echo "Clean failed"; exit 1; fi

test: ## Run all tests (unit + integration + linting)
	@if ! ./scripts/test.sh; then echo "Tests failed"; exit 1; fi

test-unit: ## Run unit tests only
	@echo "Running unit tests..."
	@go test -v -race -coverprofile=$(COVERAGE_OUT) ./...
	@echo "Coverage report: $(COVERAGE_OUT)"

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	@go test -v -tags=integration ./...

lint: ## Run all linting checks
	@if ! ./scripts/lint.sh; then echo "Linting failed"; exit 1; fi

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Code formatted successfully"

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...
	@echo "go vet passed"

staticcheck: ## Run staticcheck
	@echo "Running staticcheck..."
	@staticcheck ./...
	@echo "staticcheck passed"

install: ## Install to $GOPATH/bin
	@echo "Installing $(APP_NAME)..."
	@if ! go install -ldflags "$(LDFLAGS)" .; then echo "Install failed"; exit 1; fi
	@echo "Installation complete"

deps: ## Download and tidy dependencies
	@echo "Downloading dependencies..."
	@if ! go mod tidy; then echo "Failed to tidy modules"; exit 1; fi
	@if ! go mod download; then echo "Failed to download modules"; exit 1; fi
	@if ! go mod verify; then echo "Failed to verify modules"; exit 1; fi
	@echo "Dependencies updated successfully"

release: clean lint test build-all ## Prepare a release build
	@echo "Release build complete!"
	@echo "Version: $(VERSION)"
	@ls -la dist/

info: ## Show build information
	@echo "Build Information:"
	@echo "  App Name:    $(APP_NAME)"
	@echo "  Version:     $(VERSION)"
	@echo "  Build Time:  $(BUILD_TIME)"
	@echo "  Git Commit:  $(GIT_COMMIT)"
	@echo "  Go Version:  $(GO_VERSION)"
	@echo "  LDFLAGS:     $(LDFLAGS)"

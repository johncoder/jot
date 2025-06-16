.PHONY: build build-all clean test fmt vet install dev help

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

# Default target
help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## Build for current platform
	@echo "Building $(APP_NAME) v$(VERSION)..."
	go build -ldflags "$(LDFLAGS)" -o $(APP_NAME) .
	@echo "Build complete: ./$(APP_NAME)"

build-all: ## Build for all platforms
	@echo "Building for all platforms..."
	./build.sh --all

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -f $(APP_NAME)
	rm -rf dist/
	go clean

test: ## Run tests
	@echo "Running tests..."
	go test -v ./...

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

install: ## Install to $GOPATH/bin
	@echo "Installing $(APP_NAME)..."
	go install -ldflags "$(LDFLAGS)" .

dev: fmt vet test build ## Run development checks and build

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod tidy
	go mod download

release: clean fmt vet test build-all ## Prepare a release build

info: ## Show build information
	@echo "App Name:    $(APP_NAME)"
	@echo "Version:     $(VERSION)"
	@echo "Build Time:  $(BUILD_TIME)"
	@echo "Git Commit:  $(GIT_COMMIT)"
	@echo "LDFLAGS:     $(LDFLAGS)"

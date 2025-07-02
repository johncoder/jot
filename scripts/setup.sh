#!/bin/bash
set -e

# Ensure we exit cleanly on signals
cleanup() {
    echo ""
    echo "Setup interrupted or failed."
    echo ""
    echo "🔧 For troubleshooting help, run: $0 --troubleshoot"
    echo "📖 Or check DEVELOPMENT.md for detailed documentation"
    exit 1
}

trap cleanup INT TERM ERR

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
info() {
    echo -e "${BLUE}INFO:${NC} $1"
}

success() {
    echo -e "${GREEN}SUCCESS:${NC} $1"
}

warn() {
    echo -e "${YELLOW}WARNING:${NC} $1"
}

error() {
    echo -e "${RED}ERROR:${NC} $1"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Print what the script will do
print_overview() {
    echo "🚀 jot Development Environment Setup"
    echo ""
    echo "This will configure your development environment and ensure everything works."
    echo ""
    echo "What will be downloaded/installed:"
    echo "  • Go project dependencies (go mod download)"
    echo "  • staticcheck linting tool (if not already installed)"
    echo ""
    echo "Setup will verify Go 1.20+, run code checks, build the project, and test."
    echo "Requires internet connection and \$GOPATH/bin in PATH."
    echo ""
    
    # Skip prompt if running in non-interactive mode
    if [[ "$1" == "--yes" || "$1" == "-y" ]]; then
        echo "Running in non-interactive mode..."
        echo ""
        return
    fi
    
    read -p "Continue? [Y/n] " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Nn]$ ]]; then
        echo "Setup cancelled."
        exit 0
    fi
    echo ""
}

# Main setup function
main() {
    print_overview "$@"

    # Check Go installation
    info "Checking Go installation..."
    if ! command_exists go; then
        error "Go is not installed. Install Go 1.20+ from https://golang.org/dl/"
        exit 1
    fi

    GO_VERSION=$(go version | grep -o 'go[0-9]\+\.[0-9]\+' | sed 's/go//')
    if [[ $(echo "$GO_VERSION 1.20" | tr " " "\n" | sort -V | head -n1) != "1.20" ]]; then
        error "Go 1.20+ required. Current: $GO_VERSION. Update from https://golang.org/dl/"
        exit 1
    fi
    success "Go $GO_VERSION found"

    # Download dependencies
    info "Downloading dependencies..."
    if ! go mod tidy >/dev/null 2>&1; then
        error "Failed to tidy modules. Check internet connection or run 'go mod tidy' manually."
        exit 1
    fi
    if ! go mod download >/dev/null 2>&1; then
        error "Failed to download modules. Check internet/proxy or run 'go mod download' manually."
        exit 1
    fi
    if ! go mod verify >/dev/null 2>&1; then
        error "Failed to verify modules. Try 'go clean -modcache' and rerun."
        exit 1
    fi
    success "Dependencies ready"

    # Install development tools
    info "Installing development tools..."
    if ! command_exists staticcheck; then
        if ! go install honnef.co/go/tools/cmd/staticcheck@latest >/dev/null 2>&1; then
            error "Failed to install staticcheck. Ensure \$GOPATH/bin is writable and in PATH."
            exit 1
        fi
        success "staticcheck installed"
    else
        success "staticcheck available"
    fi

    # Verify tools work
    if ! command_exists staticcheck; then
        error "staticcheck not in PATH. Add \$GOPATH/bin to PATH: export PATH=\$PATH:\$(go env GOPATH)/bin"
        exit 1
    fi

    # Run code quality checks
    info "Running code quality checks..."
    
    # Check formatting
    UNFORMATTED=$(go fmt ./... 2>&1)
    if [ $? -ne 0 ]; then
        error "Code formatting issues. Run 'go fmt ./...' to fix."
        exit 1
    fi
    
    # Run go vet
    if ! go vet ./... >/dev/null 2>&1; then
        error "go vet found issues. Run 'go vet ./...' for details."
        exit 1
    fi
    
    # Run staticcheck
    if ! staticcheck ./... >/dev/null 2>&1; then
        error "staticcheck found issues. Run 'staticcheck ./...' for details."
        exit 1
    fi
    success "Code quality checks passed"

    # Build the project
    info "Building jot..."
    if ! go build -o jot . >/dev/null 2>&1; then
        error "Build failed. Run 'go build -v .' for details."
        exit 1
    fi
    success "Build successful"

    # Run tests
    info "Running tests..."
    if ! go test ./... >/dev/null 2>&1; then
        error "Tests failed. Run 'go test -v ./...' for details."
        exit 1
    fi
    success "Tests passed"

    echo ""
    success "✅ Development environment ready!"
    echo ""
    echo "Next steps:"
    echo "  ./jot --help         # Try the CLI"
    echo "  make help            # See all commands"
    echo "  make test            # Run tests again"
    echo ""
    echo "Happy coding! 🚀"
}

# Print troubleshooting guidance
print_troubleshooting() {
    echo "🔧 TROUBLESHOOTING GUIDE"
    echo ""
    echo "1. GO INSTALLATION:"
    echo "   'go: command not found' → Install Go from https://golang.org/dl/"
    echo "   'version too old' → Update to Go 1.20+ from official site"
    echo ""
    echo "2. DEPENDENCIES:"
    echo "   Download fails → Check internet/proxy, try: go clean -modcache"
    echo "   Corporate network → Configure GOPROXY and GOPRIVATE"
    echo ""
    echo "3. TOOLS:"
    echo "   'staticcheck not found' → Add to PATH: export PATH=\$PATH:\$(go env GOPATH)/bin"
    echo "   Add to shell profile (~/.bashrc, ~/.zshrc) and restart terminal"
    echo ""
    echo "4. BUILD/TESTS:"
    echo "   Failures → Clear caches: go clean -cache && go clean -modcache"
    echo "   Run individual commands for detailed errors"
    echo ""
    echo "5. MANUAL SETUP:"
    echo "   go mod tidy && go mod download"
    echo "   go install honnef.co/go/tools/cmd/staticcheck@latest"
    echo "   go fmt ./... && go vet ./... && staticcheck ./..."
    echo "   go build -o jot . && go test ./..."
    echo ""
    echo "For more help: Check DEVELOPMENT.md or open a GitHub issue"
}

# Handle help flag
if [[ "$1" == "--help" || "$1" == "-h" ]]; then
    echo "🚀 jot Development Environment Setup"
    echo ""
    echo "Sets up a complete development environment for the jot project."
    echo "Verifies Go 1.20+, downloads dependencies, installs tools, and validates everything works."
    echo ""
    echo "USAGE: $0 [--help] [--yes] [--troubleshoot]"
    echo ""
    echo "OPTIONS:"
    echo "  --help          Show this help"
    echo "  --yes           Non-interactive mode"
    echo "  --troubleshoot  Show troubleshooting guide"
    echo ""
    echo "REQUIREMENTS:"
    echo "  • Go 1.20+ installed and in PATH"
    echo "  • Internet connection"
    echo "  • \$GOPATH/bin in PATH (for tools)"
    echo ""
    echo "COMMON ISSUES:"
    echo "  Go not found: Install from https://golang.org/dl/"
    echo "  Version too old: Update to Go 1.20+"
    echo "  Tools not in PATH: export PATH=\$PATH:\$(go env GOPATH)/bin"
    echo "  Network issues: Check proxy/firewall settings"
    echo ""
    echo "For detailed troubleshooting: $0 --troubleshoot"
    exit 0
fi

# Handle troubleshooting flag
if [[ "$1" == "--troubleshoot" ]]; then
    print_troubleshooting
    exit 0
fi

# Run main function
main "$@"

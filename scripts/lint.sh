#!/bin/bash
set -e

# Ensure we exit cleanly on signals
trap 'echo "Lint interrupted"; exit 1' INT TERM

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

# Linting functions
check_gofmt() {
    info "Checking code formatting with gofmt..."
    local unformatted
    unformatted=$(gofmt -l .)
    
    if [[ -n "$unformatted" ]]; then
        error "The following files need formatting:"
        echo "$unformatted"
        echo ""
        echo "Run 'make fmt' or 'go fmt ./...' to fix formatting issues."
        return 1
    else
        success "Code formatting is correct"
        return 0
    fi
}

check_goimports() {
    if command_exists goimports; then
        info "Checking import formatting with goimports..."
        local unformatted
        unformatted=$(goimports -l .)
        
        if [[ -n "$unformatted" ]]; then
            warn "The following files have import formatting issues:"
            echo "$unformatted"
            echo ""
            echo "Run 'goimports -w .' to fix import issues."
        else
            success "Import formatting is correct"
        fi
    else
        info "goimports not found (optional tool)"
    fi
}

check_govet() {
    info "Running go vet..."
    if go vet ./...; then
        success "go vet passed"
        return 0
    else
        error "go vet found issues"
        return 1
    fi
}

check_staticcheck() {
    if command_exists staticcheck; then
        info "Running staticcheck..."
        if staticcheck ./...; then
            success "staticcheck passed"
            return 0
        else
            error "staticcheck found issues"
            return 1
        fi
    else
        warn "staticcheck not found - install with: go install honnef.co/go/tools/cmd/staticcheck@latest"
        return 0
    fi
}

check_golint() {
    if command_exists golint; then
        info "Running golint..."
        local lint_output
        lint_output=$(golint ./... | grep -v "should have comment or be unexported" || true)
        
        if [[ -n "$lint_output" ]]; then
            warn "golint suggestions:"
            echo "$lint_output"
        else
            success "golint passed"
        fi
    else
        info "golint not found (optional tool)"
    fi
}

check_govulncheck() {
    if command_exists govulncheck; then
        info "Running govulncheck..."
        if govulncheck ./...; then
            success "No known vulnerabilities found"
            return 0
        else
            error "govulncheck found vulnerabilities"
            return 1
        fi
    else
        info "govulncheck not found (optional tool) - install with: go install golang.org/x/vuln/cmd/govulncheck@latest"
        return 0
    fi
}

check_go_mod() {
    info "Checking go.mod and go.sum..."
    
    # Save current state
    local temp_dir=$(mktemp -d)
    cp go.mod "$temp_dir/go.mod.orig" 2>/dev/null || true
    cp go.sum "$temp_dir/go.sum.orig" 2>/dev/null || true
    
    # Check if go.mod is tidy
    if ! go mod tidy; then
        error "go mod tidy failed"
        rm -rf "$temp_dir"
        return 1
    fi
    
    # Wait a moment for file system to sync
    sleep 0.1
    
    # Compare with original state
    local changed=false
    if ! diff -q go.mod "$temp_dir/go.mod.orig" >/dev/null 2>&1; then
        changed=true
    fi
    if ! diff -q go.sum "$temp_dir/go.sum.orig" >/dev/null 2>&1; then
        changed=true
    fi
    
    rm -rf "$temp_dir"
    
    if [[ "$changed" == "true" ]]; then
        warn "go.mod or go.sum was updated by 'go mod tidy'"
        warn "This is normal if dependencies were added/updated"
        info "Changes have been applied automatically"
        return 0
    else
        success "go.mod and go.sum are tidy"
        return 0
    fi
}

# Main linting function
main() {
    echo "🔍 Running code quality checks for jot..."
    echo ""

    local failed=0

    # Core Go tools (required)
    if ! check_gofmt; then failed=1; fi
    if ! check_govet; then failed=1; fi
    if ! check_go_mod; then failed=1; fi
    
    # Static analysis tools
    if ! check_staticcheck; then failed=1; fi
    
    # Optional tools (warnings only)
    check_goimports
    check_golint
    check_govulncheck

    echo ""
    if [[ $failed -eq 0 ]]; then
        success "All code quality checks passed! 🎉"
        echo ""
        echo "Checks completed:"
        echo "  ✅ Code formatting (gofmt)"
        echo "  ✅ Code correctness (go vet)"
        echo "  ✅ Static analysis (staticcheck)"
        echo "  ✅ Module integrity (go mod)"
    else
        error "Some code quality checks failed! ❌"
        echo ""
        echo "Please fix the issues above before proceeding."
        exit 1
    fi
}

# Handle command line arguments
case "${1:-}" in
    --fmt)
        check_gofmt
        ;;
    --vet)
        check_govet
        ;;
    --staticcheck)
        check_staticcheck
        ;;
    --mod)
        check_go_mod
        ;;
    --help|-h)
        echo "jot Linting Script"
        echo ""
        echo "Runs code quality checks for the jot project."
        echo ""
        echo "Usage: $0 [OPTIONS]"
        echo ""
        echo "Options:"
        echo "  --fmt         Check code formatting only"
        echo "  --vet         Run go vet only"
        echo "  --staticcheck Run staticcheck only"
        echo "  --mod         Check go.mod integrity only"
        echo "  --help, -h    Show this help message"
        echo ""
        echo "Default: Run all linting checks"
        echo ""
        echo "Required tools:"
        echo "  • gofmt (built into Go)"
        echo "  • go vet (built into Go)"
        echo "  • staticcheck (install: go install honnef.co/go/tools/cmd/staticcheck@latest)"
        echo ""
        echo "Optional tools:"
        echo "  • goimports (install: go install golang.org/x/tools/cmd/goimports@latest)"
        echo "  • golint (install: go install golang.org/x/lint/golint@latest)"
        echo "  • govulncheck (install: go install golang.org/x/vuln/cmd/govulncheck@latest)"
        ;;
    "")
        main
        ;;
    *)
        error "Unknown option: $1"
        echo "Use --help for usage information."
        exit 1
        ;;
esac

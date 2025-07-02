#!/bin/bash
set -e

# Ensure we exit cleanly on signals
trap 'echo "Test interrupted"; exit 1' INT TERM

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
COVERAGE_FILE="coverage.out"
COVERAGE_HTML="coverage.html"

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

# Test functions
run_unit_tests() {
    info "Running unit tests with coverage..."
    if go test -race -coverprofile="$COVERAGE_FILE" ./...; then
        success "Unit tests passed"
        
        # Generate coverage report
        if [[ -f "$COVERAGE_FILE" ]]; then
            COVERAGE=$(go tool cover -func="$COVERAGE_FILE" | grep total | awk '{print $3}')
            info "Test coverage: $COVERAGE"
            
            # Generate HTML coverage report
            if ! go tool cover -html="$COVERAGE_FILE" -o "$COVERAGE_HTML"; then
                warn "Failed to generate HTML coverage report"
            else
                info "HTML coverage report: $COVERAGE_HTML"
            fi
        fi
        return 0
    else
        error "Unit tests failed"
        return 1
    fi
}

run_integration_tests() {
    info "Running integration tests..."
    if go test -tags=integration ./...; then
        success "Integration tests passed"
        return 0
    else
        error "Integration tests failed"
        return 1
    fi
}

run_linting() {
    info "Running linting checks..."
    if ! ./scripts/lint.sh; then
        error "Linting checks failed"
        return 1
    fi
    return 0
}

run_build_test() {
    info "Testing build process..."
    if go build -o /tmp/jot-test .; then
        success "Build test passed"
        rm -f /tmp/jot-test
        return 0
    else
        error "Build test failed"
        return 1
    fi
}

# Main test function
main() {
    echo "🧪 Running comprehensive test suite for jot..."
    echo ""

    local failed=0

    # Clean up previous coverage files
    rm -f "$COVERAGE_FILE" "$COVERAGE_HTML"

    # Run linting first (fast feedback)
    if ! run_linting; then
        failed=1
    fi

    # Run unit tests
    if ! run_unit_tests; then
        failed=1
    fi

    # Run integration tests
    if ! run_integration_tests; then
        failed=1
    fi

    # Test build process
    if ! run_build_test; then
        failed=1
    fi

    echo ""
    if [[ $failed -eq 0 ]]; then
        success "All tests passed! 🎉"
        echo ""
        echo "Test summary:"
        echo "  ✅ Linting checks"
        echo "  ✅ Unit tests with coverage"
        echo "  ✅ Integration tests"
        echo "  ✅ Build verification"
        
        if [[ -f "$COVERAGE_FILE" ]]; then
            echo ""
            echo "Coverage files generated:"
            echo "  • $COVERAGE_FILE (for CI/tooling)"
            echo "  • $COVERAGE_HTML (open in browser)"
        fi
    else
        error "Some tests failed! ❌"
        echo ""
        echo "Please fix the failing tests before proceeding."
        exit 1
    fi
}

# Handle command line arguments
case "${1:-}" in
    --unit)
        echo "🧪 Running unit tests only..."
        run_unit_tests
        ;;
    --integration)
        echo "🧪 Running integration tests only..."
        run_integration_tests
        ;;
    --lint)
        echo "🧪 Running linting only..."
        run_linting
        ;;
    --help|-h)
        echo "jot Test Runner"
        echo ""
        echo "Runs comprehensive tests for the jot project."
        echo ""
        echo "Usage: $0 [OPTIONS]"
        echo ""
        echo "Options:"
        echo "  --unit         Run unit tests only"
        echo "  --integration  Run integration tests only"
        echo "  --lint         Run linting checks only"
        echo "  --help, -h     Show this help message"
        echo ""
        echo "Default: Run all tests (linting + unit + integration + build)"
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

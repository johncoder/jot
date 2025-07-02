#!/bin/bash
set -e

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

# Main cleanup function
main() {
    echo "🧹 Cleaning jot build artifacts and temporary files..."
    echo ""

    # Remove build artifacts
    info "Removing build artifacts..."
    rm -f jot
    rm -rf dist/
    
    # Remove Go build cache
    info "Cleaning Go build cache..."
    go clean -cache
    go clean -modcache
    
    # Remove test artifacts
    info "Removing test artifacts..."
    rm -f coverage.out coverage.html
    find . -name "*.test" -type f -delete
    
    # Remove temporary files
    info "Removing temporary files..."
    find . -name "*.tmp" -type f -delete
    find . -name ".DS_Store" -type f -delete
    find . -name "Thumbs.db" -type f -delete
    
    # Remove log files
    info "Removing log files..."
    find . -name "*.log" -type f -delete
    
    # Remove editor backup files
    info "Removing editor backup files..."
    find . -name "*~" -type f -delete
    find . -name "*.swp" -type f -delete
    find . -name "*.swo" -type f -delete
    find . -name "#*#" -type f -delete
    
    echo ""
    success "Cleanup complete! 🎉"
    echo ""
    echo "Cleaned:"
    echo "  • Build artifacts (jot binary, dist/ directory)"
    echo "  • Go build and module cache"
    echo "  • Test artifacts (coverage files, test binaries)"
    echo "  • Temporary files"
    echo "  • Editor backup files"
}

# Handle command line arguments
case "${1:-}" in
    --all)
        info "Deep cleaning (including Go module cache)..."
        main
        ;;
    --help|-h)
        echo "jot Cleanup Script"
        echo ""
        echo "Cleans build artifacts and temporary files."
        echo ""
        echo "Usage: $0 [OPTIONS]"
        echo ""
        echo "Options:"
        echo "  --all         Deep clean including Go module cache"
        echo "  --help, -h    Show this help message"
        echo ""
        echo "Default: Clean build artifacts and temporary files"
        ;;
    "")
        # Light cleanup (preserve module cache)
        echo "🧹 Cleaning jot build artifacts..."
        echo ""
        
        info "Removing build artifacts..."
        rm -f jot
        rm -rf dist/
        
        info "Cleaning Go build cache..."
        go clean
        
        info "Removing test artifacts..."
        rm -f coverage.out coverage.html
        find . -name "*.test" -type f -delete
        
        info "Removing temporary files..."
        find . -name "*.tmp" -type f -delete
        find . -name ".DS_Store" -type f -delete
        find . -name "Thumbs.db" -type f -delete
        find . -name "*.log" -type f -delete
        find . -name "*~" -type f -delete
        find . -name "*.swp" -type f -delete
        find . -name "*.swo" -type f -delete
        find . -name "#*#" -type f -delete
        
        success "Basic cleanup complete!"
        ;;
    *)
        error "Unknown option: $1"
        echo "Use --help for usage information."
        exit 1
        ;;
esac

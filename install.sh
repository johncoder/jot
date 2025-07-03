#!/bin/bash
set -e

# jot installer script
# Usage: curl -sSL https://raw.githubusercontent.com/johncoder/jot/main/install.sh | sh

# Configuration
GITHUB_REPO="johncoder/jot"
INSTALL_DIR="$HOME/.local/bin"
BINARY_NAME="jot"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# Detect OS and architecture
detect_platform() {
    local os=""
    local arch=""
    
    # Detect OS
    case "$(uname -s)" in
        Linux*) os="linux" ;;
        Darwin*) os="darwin" ;;
        CYGWIN*|MINGW*|MSYS*) os="windows" ;;
        *) 
            error "Unsupported operating system: $(uname -s)"
            exit 1
            ;;
    esac
    
    # Detect architecture
    case "$(uname -m)" in
        x86_64|amd64) arch="amd64" ;;
        arm64|aarch64) arch="arm64" ;;
        *)
            error "Unsupported architecture: $(uname -m)"
            exit 1
            ;;
    esac
    
    echo "${os}_${arch}"
}

# Get latest release version
get_latest_version() {
    local version
    version=$(curl -sSL "https://api.github.com/repos/$GITHUB_REPO/releases/latest" | \
        grep '"tag_name":' | \
        sed -E 's/.*"([^"]+)".*/\1/')
    
    if [ -z "$version" ]; then
        error "Failed to get latest version"
        exit 1
    fi
    
    echo "$version"
}

# Check if binary is in PATH
check_path() {
    if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
        warn "$INSTALL_DIR is not in your PATH"
        echo ""
        echo "To add it, run one of these commands:"
        echo ""
        echo "For bash:"
        echo "  echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.bashrc"
        echo "  source ~/.bashrc"
        echo ""
        echo "For zsh:"
        echo "  echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.zshrc"
        echo "  source ~/.zshrc"
        echo ""
        echo "For fish:"
        echo "  fish_add_path ~/.local/bin"
        echo ""
        return 1
    fi
    return 0
}

# Main installation function
main() {
    echo "🚀 jot installer"
    echo ""
    
    # Detect platform
    info "Detecting platform..."
    PLATFORM=$(detect_platform)
    info "Platform: $PLATFORM"
    
    # Get latest version
    info "Getting latest version..."
    VERSION=$(get_latest_version)
    info "Latest version: $VERSION"
    
    # Construct download URL
    ARCHIVE_NAME="jot_${VERSION}_${PLATFORM}.tar.gz"
    if [[ "$PLATFORM" == *"windows"* ]]; then
        ARCHIVE_NAME="jot_${VERSION}_${PLATFORM}.zip"
    fi
    
    DOWNLOAD_URL="https://github.com/$GITHUB_REPO/releases/download/$VERSION/$ARCHIVE_NAME"
    
    info "Download URL: $DOWNLOAD_URL"
    
    # Create install directory
    info "Creating install directory..."
    mkdir -p "$INSTALL_DIR"
    
    # Create temporary directory
    TEMP_DIR=$(mktemp -d)
    trap "rm -rf $TEMP_DIR" EXIT
    
    # Download and extract
    info "Downloading jot..."
    cd "$TEMP_DIR"
    
    if command -v curl >/dev/null 2>&1; then
        curl -sSL -o "$ARCHIVE_NAME" "$DOWNLOAD_URL"
    elif command -v wget >/dev/null 2>&1; then
        wget -q -O "$ARCHIVE_NAME" "$DOWNLOAD_URL"
    else
        error "Neither curl nor wget found. Please install one of them."
        exit 1
    fi
    
    info "Extracting archive..."
    if [[ "$ARCHIVE_NAME" == *.zip ]]; then
        if command -v unzip >/dev/null 2>&1; then
            unzip -q "$ARCHIVE_NAME"
        else
            error "unzip not found. Please install unzip to extract Windows archives."
            exit 1
        fi
    else
        tar -xzf "$ARCHIVE_NAME"
    fi
    
    # Find the binary
    BINARY_PATH=""
    for file in jot_*; do
        if [ -x "$file" ] && [ ! -d "$file" ]; then
            BINARY_PATH="$file"
            break
        fi
    done
    
    if [ -z "$BINARY_PATH" ]; then
        error "Binary not found in archive"
        exit 1
    fi
    
    # Install binary
    info "Installing jot to $INSTALL_DIR..."
    cp "$BINARY_PATH" "$INSTALL_DIR/$BINARY_NAME"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
    
    success "✅ jot $VERSION installed successfully!"
    echo ""
    
    # Check PATH
    if check_path; then
        info "jot is ready to use!"
        echo "Try: jot --help"
    else
        warn "Please add ~/.local/bin to your PATH (see instructions above)"
        echo "Then try: jot --help"
    fi
    
    echo ""
    echo "📚 Documentation: https://github.com/$GITHUB_REPO"
    echo "🐛 Report issues: https://github.com/$GITHUB_REPO/issues"
}

# Handle command line arguments
case "${1:-}" in
    --help|-h)
        echo "jot installer"
        echo ""
        echo "Usage: $0"
        echo ""
        echo "This script installs the latest version of jot to ~/.local/bin"
        echo ""
        echo "Environment variables:"
        echo "  INSTALL_DIR   Installation directory (default: ~/.local/bin)"
        echo ""
        echo "Examples:"
        echo "  $0                    # Install to default location"
        echo "  INSTALL_DIR=~/bin $0  # Install to ~/bin"
        exit 0
        ;;
    --version|-v)
        get_latest_version
        exit 0
        ;;
esac

# Run main installation
main "$@"

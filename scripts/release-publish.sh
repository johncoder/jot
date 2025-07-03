#!/bin/bash
set -e

# Script to publish a release to GitHub and update Homebrew
# Usage: ./scripts/release-publish.sh <version>

VERSION=${1:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}

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

# Check if version is provided
if [ -z "$VERSION" ] || [ "$VERSION" = "dev" ]; then
    error "No version specified or version is 'dev'"
    echo "Usage: $0 <version>"
    echo "Example: $0 v0.9.0"
    exit 1
fi

# Check if we're in the right directory
if [ ! -f "go.mod" ] || [ ! -d "dist" ]; then
    error "Must be run from project root and dist/ must exist"
    echo "Run 'make release-build' first"
    exit 1
fi

# Check if GitHub CLI is installed
if ! command -v gh >/dev/null 2>&1; then
    error "GitHub CLI (gh) not found"
    echo "Install it with: brew install gh  (or see https://cli.github.com/)"
    exit 1
fi

# Check if we're authenticated with GitHub
if ! gh auth status >/dev/null 2>&1; then
    error "Not authenticated with GitHub"
    echo "Run: gh auth login"
    exit 1
fi

# Verify we have release artifacts
ARTIFACTS=($(ls dist/*.tar.gz dist/*.zip 2>/dev/null || true))
if [ ${#ARTIFACTS[@]} -eq 0 ]; then
    error "No release artifacts found in dist/"
    echo "Run 'make release-build' first"
    exit 1
fi

info "Publishing release $VERSION..."

# Create GitHub release
info "Creating GitHub release..."
RELEASE_NOTES="Release $VERSION

## Installation

### Shell Script (Recommended)
\`\`\`bash
curl -sSL https://raw.githubusercontent.com/johncoder/jot/main/install.sh | sh
\`\`\`

### Homebrew
\`\`\`bash
brew install johncoder/tap/jot
\`\`\`

### Go Install
\`\`\`bash
go install github.com/johncoder/jot@$VERSION
\`\`\`

### Manual Download
Download the appropriate binary for your platform from the assets below.

## Changes
- See commit history for detailed changes
"

# Create the release
gh release create "$VERSION" \
    --title "jot $VERSION" \
    --notes "$RELEASE_NOTES" \
    dist/*.tar.gz dist/*.zip dist/checksums.txt

success "GitHub release created: $VERSION"

# Update Homebrew formula (if tap exists)
info "Checking for Homebrew tap..."
if gh repo view johncoder/homebrew-tap >/dev/null 2>&1; then
    info "Updating Homebrew formula..."
    # This would need the actual tap repository and formula
    warn "Homebrew formula update not implemented yet"
    echo "Manual step: Update formula in johncoder/homebrew-tap repository"
else
    warn "Homebrew tap not found (johncoder/homebrew-tap)"
    echo "Consider creating a tap repository for easier installation"
fi

echo ""
success "🎉 Release $VERSION published successfully!"
echo ""
echo "📋 Next steps:"
echo "  • Test installation: curl -sSL https://raw.githubusercontent.com/johncoder/jot/main/install.sh | sh"
echo "  • Update documentation with new version"
echo "  • Announce the release"
echo ""
echo "📊 Release info:"
echo "  Version: $VERSION"
echo "  Artifacts: ${#ARTIFACTS[@]} files"
echo "  GitHub: https://github.com/johncoder/jot/releases/tag/$VERSION"

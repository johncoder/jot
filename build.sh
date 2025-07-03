#!/bin/bash

set -e

# Build configuration
APP_NAME="jot"
# VERSION: Uses git describe to get semver tags (e.g., v0.9.0) or commit hash
# - For releases: Set VERSION=v0.9.0 manually or ensure a proper git tag exists
# - For development: Automatically uses git describe --tags --always --dirty
VERSION=${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=${GIT_COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")}

# Build flags
LDFLAGS="-s -w"
LDFLAGS="$LDFLAGS -X github.com/johncoder/jot/cmd.version=$VERSION"
LDFLAGS="$LDFLAGS -X github.com/johncoder/jot/cmd.buildTime=$BUILD_TIME"
LDFLAGS="$LDFLAGS -X github.com/johncoder/jot/cmd.gitCommit=$GIT_COMMIT"

# Platforms to build for
PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
    "windows/arm64"
)

# Clean previous builds
echo "Cleaning previous builds..."
rm -rf dist/
mkdir -p dist/

# Build for current platform (development)
echo "Building for current platform..."
go build -ldflags "$LDFLAGS" -o "dist/$APP_NAME" .

# Function to build for specific platform
build_platform() {
    local platform=$1
    local os=$(echo $platform | cut -d'/' -f1)
    local arch=$(echo $platform | cut -d'/' -f2)
    local output_name="${APP_NAME}_${os}_${arch}"
    
    if [ "$os" = "windows" ]; then
        output_name="${output_name}.exe"
    fi
    
    echo "Building for $os/$arch..."
    GOOS=$os GOARCH=$arch go build \
        -ldflags "$LDFLAGS" \
        -o "dist/$output_name" .
}

# Build for all platforms if --all flag is provided
if [ "$1" = "--all" ]; then
    echo "Building for all platforms..."
    for platform in "${PLATFORMS[@]}"; do
        build_platform $platform
    done
    
    echo "Creating checksums..."
    cd dist
    sha256sum * > checksums.txt
    cd ..
    
    # Create archives for release
    echo "Creating release archives..."
    cd dist
    for platform in "${PLATFORMS[@]}"; do
        os=$(echo $platform | cut -d'/' -f1)
        arch=$(echo $platform | cut -d'/' -f2)
        binary_name="${APP_NAME}_${os}_${arch}"
        
        if [ "$os" = "windows" ]; then
            binary_name="${binary_name}.exe"
            archive_name="${APP_NAME}_${VERSION}_${os}_${arch}.zip"
            zip "$archive_name" "$binary_name" checksums.txt
        else
            archive_name="${APP_NAME}_${VERSION}_${os}_${arch}.tar.gz"
            tar -czf "$archive_name" "$binary_name" checksums.txt
        fi
        
        echo "Created $archive_name"
    done
    cd ..
    
    echo "Build complete! Binaries and archives are in dist/"
    ls -la dist/
else
    echo "Build complete! Binary: dist/$APP_NAME"
    echo "Use './build.sh --all' to build for all platforms"
fi

echo "Version: $VERSION"
echo "Build time: $BUILD_TIME"
echo "Git commit: $GIT_COMMIT"

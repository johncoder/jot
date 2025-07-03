# Release Process Guide

This document describes how to release new versions of jot.

## Versioning Strategy

jot follows [Semantic Versioning (SemVer)](https://semver.org/):

- **MAJOR.MINOR.PATCH** (e.g., `v1.2.3`)
- **MAJOR**: Breaking changes to CLI interface or behavior
- **MINOR**: New features, backwards compatible
- **PATCH**: Bug fixes, backwards compatible

### Version in `jot --version`

The version displayed by `jot --version` depends on the build:

- **Development builds**: `jot version dev (build: unknown, commit: unknown)`
- **Release builds**: `jot version v0.9.0` (clean semver)

### When to Update the Version

Update the version for:

- **PATCH** (v0.9.1): Bug fixes, documentation updates, minor improvements
- **MINOR** (v0.10.0): New commands, new features, configuration changes
- **MAJOR** (v1.0.0): Breaking changes to CLI interface or file formats

### How Versioning Works

1. **Git Tags**: Create tags like `v0.9.0` for releases
2. **Build Detection**: `git describe --tags --always --dirty` automatically detects version
3. **Manual Override**: Set `VERSION=v0.9.0` environment variable to override
4. **LDFLAGS Injection**: Version is compiled into the binary at build time

```bash
# Development build (no tags)
jot --version  # Shows: jot version <commit-hash> (build: ..., commit: ...)

# Release build (with v0.9.0 tag)
jot --version  # Shows: jot version v0.9.0
```

## Overview

The release process is designed to be simple and reliable:

1. **Build**: `make release-build` - Creates all release artifacts
2. **Publish**: `make release-publish` - Uploads to GitHub and updates distribution

## Prerequisites

### Required Tools

- **GitHub CLI**: `brew install gh` (or see [installation guide](https://cli.github.com/))
- **Git**: For tagging and version management
- **Go**: For building binaries

### Authentication

```bash
# Authenticate with GitHub
gh auth login
```

## Release Steps

### 1. Prepare for Release

Choose the new version number according to SemVer:

- Bug fixes: increment PATCH (v0.9.0 → v0.9.1)
- New features: increment MINOR (v0.9.0 → v0.10.0)
- Breaking changes: increment MAJOR (v0.9.0 → v1.0.0)

```bash
# Ensure everything is committed and pushed
git add .
git commit -m "Prepare for release v0.9.0"
git push origin main

# Tag the release (this enables clean version display)
git tag v0.9.0
git push origin v0.9.0

# Verify version detection works
git describe --tags --always --dirty  # Should show: v0.9.0
```

### 2. Build Release Artifacts

```bash
# This will:
# - Clean previous builds
# - Run linting and tests
# - Build for all platforms
# - Create archives and checksums
make release-build
```

The artifacts will be created in `dist/`:

- `jot_v0.9.0_linux_amd64.tar.gz`
- `jot_v0.9.0_darwin_amd64.tar.gz`
- `jot_v0.9.0_windows_amd64.zip`
- ... (all supported platforms)
- `checksums.txt`

### 3. Publish Release

```bash
# This will:
# - Create GitHub release
# - Upload all artifacts
# - Update Homebrew formula (if tap exists)
make release-publish
```

## Installation Methods

After release, users can install jot using:

### 1. Shell Script (Recommended)

```bash
curl -sSL https://raw.githubusercontent.com/johncoder/jot/main/install.sh | sh
```

### 2. Homebrew (after setting up tap)

```bash
brew install johncoder/tap/jot
```

### 3. Go Install

```bash
go install github.com/johncoder/jot@v0.9.0
```

### 4. Manual Download

Download binaries directly from [GitHub Releases](https://github.com/johncoder/jot/releases).

## Setting Up Homebrew Tap (Optional)

To enable `brew install johncoder/tap/jot`:

1. Create repository: `johncoder/homebrew-tap`
2. Copy `homebrew/jot.rb` to `Formula/jot.rb` in that repo
3. Update SHA256 checksums for each release
4. Users can then: `brew install johncoder/tap/jot`

## Versioning

- Use semantic versioning: `v0.9.0`, `v0.9.1`, `v1.0.0`
- Pre-1.0 versions indicate API may change
- Post-1.0 versions follow strict semantic versioning

## Troubleshooting

### GitHub CLI Issues

```bash
# Check authentication
gh auth status

# Re-authenticate if needed
gh auth login
```

### Missing Artifacts

```bash
# Rebuild release artifacts
make clean
make release-build
```

### Failed Upload

```bash
# Check if release already exists
gh release view v0.9.0

# Delete and recreate if needed
gh release delete v0.9.0
make release-publish
```

## File Structure

```
dist/                           # Build artifacts (generated)
├── jot_v0.9.0_linux_amd64.tar.gz
├── jot_v0.9.0_darwin_amd64.tar.gz
├── jot_v0.9.0_windows_amd64.zip
└── checksums.txt

scripts/
├── release-publish.sh          # Publishing automation

homebrew/
└── jot.rb                      # Homebrew formula template

install.sh                      # User installation script
build.sh                        # Cross-platform build script
Makefile                        # Build automation
```

## Security Notes

- All downloads are served over HTTPS
- Checksums are provided for verification
- Release artifacts are signed by GitHub
- Installation script validates checksums

## Future Improvements

- Automated formula updates
- Code signing for binaries
- Windows installer (MSI)
- Package manager support (apt, yum, etc.)
- Automated changelog generation

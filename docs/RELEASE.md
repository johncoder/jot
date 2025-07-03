# Release Process

This document describes how to create and publish releases for jot.

## Overview

jot uses a two-step release process:

1. **Build**: `make release-build` - Creates release artifacts locally
2. **Publish**: `make release-publish` - Uploads to GitHub and updates distribution channels

## Prerequisites

### Required Tools

- Git with tags
- GitHub CLI (`gh`) installed and authenticated
- Go 1.20+ for building

### Setup GitHub CLI

```bash
# Install GitHub CLI
brew install gh  # macOS
# or: curl -sSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | sudo tee /etc/apt/keyrings/githubcli-archive-keyring.gpg > /dev/null
# ...follow instructions for your platform

# Authenticate
gh auth login
```

## Release Steps

### 1. Prepare for Release

Ensure the code is ready:

```bash
# Run all checks
make test

# Verify version will be correct
make info
```

### 2. Create Git Tag

```bash
# Create and push a version tag
git tag v0.9.0
git push origin v0.9.0
```

The version format should follow semantic versioning: `vMAJOR.MINOR.PATCH`

### 3. Build Release Artifacts

```bash
make release-build
```

This will:

- Clean previous builds
- Run linting and tests
- Build binaries for all platforms
- Create compressed archives
- Generate checksums

Artifacts are created in `dist/`:

- `jot_v0.9.0_linux_amd64.tar.gz`
- `jot_v0.9.0_darwin_arm64.tar.gz`
- `jot_v0.9.0_windows_amd64.zip`
- `checksums.txt`
- etc.

### 4. Publish Release

```bash
make release-publish
```

This will:

- Create a GitHub release with the tag
- Upload all binary archives
- Generate release notes template
- Open your `$EDITOR` to customize release notes

The script will:

1. Detect the version from your git tag
2. Create a GitHub release draft
3. Upload all archives and checksums
4. Open release notes for editing
5. Publish the release

### 5. Post-Release Tasks

After publishing:

1. **Test the installer**:

   ```bash
   curl -sSL https://raw.githubusercontent.com/johncoder/jot/main/install.sh | sh
   ```

2. **Update Homebrew** (if you have a tap):

   ```bash
   # Update the formula in your homebrew tap repository
   # This is typically automated but may need manual updates
   ```

3. **Verify go install**:

   ```bash
   go install github.com/johncoder/jot@v0.9.0
   ```

4. **Announce the release**:
   - Update README.md if needed
   - Social media, etc.

## Installation Methods

After release, users can install jot via:

### Shell Script (Recommended)

```bash
curl -sSL https://raw.githubusercontent.com/johncoder/jot/main/install.sh | sh
```

### Homebrew

```bash
brew install johncoder/tap/jot
```

### Go Install

```bash
go install github.com/johncoder/jot@latest
```

### Manual Download

Download from [GitHub Releases](https://github.com/johncoder/jot/releases)

## Troubleshooting

### Build Issues

- Ensure all tests pass: `make test`
- Check Go version: `go version` (requires 1.20+)
- Clean and retry: `make clean && make release-build`

### GitHub CLI Issues

- Check authentication: `gh auth status`
- Re-authenticate: `gh auth login`
- Check repository access: `gh repo view johncoder/jot`

### Version Issues

- Ensure git tag exists: `git tag --list | grep v0.9.0`
- Check if tag is pushed: `git ls-remote --tags origin`
- Verify VERSION detection: `make info`

### Archive Issues

- Check all platforms built: `ls -la dist/`
- Verify checksums: `cd dist && sha256sum -c checksums.txt`
- Test archives: `tar -tzf dist/jot_v0.9.0_linux_amd64.tar.gz`

## File Structure

```
dist/
├── checksums.txt
├── jot                           # Current platform binary
├── jot_linux_amd64              # Platform binaries
├── jot_darwin_arm64
├── jot_windows_amd64.exe
├── jot_v0.9.0_linux_amd64.tar.gz    # Release archives
├── jot_v0.9.0_darwin_arm64.tar.gz
└── jot_v0.9.0_windows_amd64.zip
```

## Release Checklist

- [ ] All tests passing (`make test`)
- [ ] Version tag created and pushed
- [ ] Release notes prepared
- [ ] GitHub CLI authenticated
- [ ] Run `make release-build`
- [ ] Review artifacts in `dist/`
- [ ] Run `make release-publish`
- [ ] Customize release notes
- [ ] Test installation methods
- [ ] Update documentation if needed
- [ ] Announce release

## Automation Notes

The current process is manual to maintain control, but could be automated with GitHub Actions in the future. The two-step process (`release-build` + `release-publish`) allows for manual verification of artifacts before publishing.

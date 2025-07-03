# Development Guide

This document provides information for developers working on the jot project.

## Quick Start

For new contributors:

```bash
# 1. Clone the repository
git clone https://github.com/johncoder/jot.git
cd jot

# 2. Set up development environment (interactive setup)
./scripts/setup.sh

# OR: Non-interactive setup
./scripts/setup.sh --yes

# 3. Build and test (if setup didn't already do this)
make build
make test
```

**Need help?** The setup script provides guidance:

- `./scripts/setup.sh --help` - Requirements and common issues
- `./scripts/setup.sh --troubleshoot` - Detailed troubleshooting guide

## Development Scripts

The `scripts/` directory contains development automation tools:

### Setup

```bash
./scripts/setup.sh              # Interactive setup (recommended for first time)
./scripts/setup.sh --yes        # Non-interactive setup (for automation)
./scripts/setup.sh --help       # Detailed help and prerequisites
./scripts/setup.sh --troubleshoot  # Troubleshooting guide

# Makefile aliases
make setup                       # Same as ./scripts/setup.sh --yes
```

**What it does:**
Verifies Go is installed, downloads dependencies, installs staticcheck, runs code checks, builds, and tests.

**Features:**

- Interactive mode with confirmation (default)
- Non-interactive mode for automation (`--yes`)
- Concise output with specific error guidance
- Complete troubleshooting documentation

### Testing

```bash
make test               # Run all tests (linting + unit + integration + build)
make test-unit          # Run unit tests only with coverage
make test-integration   # Run integration tests only

./scripts/test.sh       # All tests (same as 'make test')
./scripts/test.sh --unit         # Unit tests only
./scripts/test.sh --integration  # Integration tests only
./scripts/test.sh --lint         # Linting only
```

**Test outputs:**

- `coverage.out` - Coverage data for CI/tooling
- `coverage.html` - HTML coverage report (open in browser)

### Code Quality

```bash
make lint               # Run all linting checks
make fmt                # Format code with gofmt
make vet                # Run go vet
make staticcheck        # Run staticcheck

./scripts/lint.sh       # All linting (same as 'make lint')
./scripts/lint.sh --fmt        # Format checking only
./scripts/lint.sh --vet        # go vet only
./scripts/lint.sh --staticcheck # staticcheck only
```

**Linting includes:**

- Code formatting (gofmt)
- Code correctness (go vet)
- Static analysis (staticcheck)
- Module integrity (go mod tidy)
- Optional: goimports, golint, govulncheck

### Building

```bash
make build              # Build for current platform
make build-all          # Build for all platforms (uses build.sh)
make install            # Build and install to $GOPATH/bin
```

### Version Information

jot uses Git tags for version management:

```bash
# Check current version detection
git describe --tags --always --dirty

# Development builds show commit hash
jot --version  # jot version <hash> (build: ..., commit: ...)

# Tagged builds show clean semver
git tag v0.9.0
jot --version  # jot version v0.9.0

# Manual version override
VERSION=v1.0.0-beta make build
```

Version is injected at build time via LDFLAGS:

- `version`: Git tag or commit hash
- `buildTime`: UTC timestamp
- `gitCommit`: Short commit hash

See `RELEASE.md` for full versioning strategy.

### Cleanup

```bash
make clean              # Clean build artifacts and temp files
./scripts/clean.sh      # Same as above
./scripts/clean.sh --all # Deep clean (includes Go module cache)
```

### Utilities

```bash
make deps               # Download and tidy dependencies
make info               # Show build information
make release            # Full release build (clean + lint + test + build-all)
```

## Development Tools

### Required Tools

These tools are automatically installed by `make setup`:

- **Go 1.24+** - Required Go version
- **staticcheck** - Static analysis tool

### Optional Tools

These provide additional code quality checks (install manually):

```bash
# Import formatting
go install golang.org/x/tools/cmd/goimports@latest

# Vulnerability scanning
go install golang.org/x/vuln/cmd/govulncheck@latest

# Additional linting
go install golang.org/x/lint/golint@latest
```

## Makefile Targets

Run `make help` to see all available targets:

```
Development:
  setup           Setup development environment
  test            Run all tests (unit + integration + linting)
  lint            Run all linting checks
  fmt             Format code
  vet             Run go vet
  staticcheck     Run staticcheck

Building:
  build           Build for current platform
  build-all       Build for all platforms
  clean           Clean build artifacts and temporary files
  install         Install to $GOPATH/bin

Utilities:
  deps            Download and tidy dependencies
  release         Prepare a release build
  info            Show build information
```

## Code Quality Standards

The project follows strict code quality standards:

### Formatting

- All code must pass `gofmt`
- Imports should be organized (use `goimports` if available)

### Linting

- Code must pass `go vet` without warnings
- Code must pass `staticcheck` without issues
- `go.mod` and `go.sum` must be tidy

### Testing

- Unit tests must pass with good coverage
- Integration tests must pass
- Build process must complete successfully

## Editor Configuration

The project includes `.editorconfig` for consistent formatting across editors:

- Go files: tabs, width 4
- Other files: spaces, width 2
- Line endings: LF
- Charset: UTF-8
- Trim trailing whitespace
- Insert final newline

Most modern editors support EditorConfig automatically.

## Continuous Integration

The scripts are designed to work well in CI environments:

```bash
# CI workflow example
./scripts/setup.sh      # Setup (fails fast on issues)
./scripts/test.sh       # Comprehensive testing
./scripts/build.sh --all # Multi-platform builds
```

All scripts:

- Use proper exit codes (0 = success, 1 = failure)
- Provide colored output for human readability
- Support --help for documentation
- Are designed to be fast and reliable

## Troubleshooting

### Common Issues

**Setup fails with "Go not found":**

- Install Go 1.24 or later
- Ensure Go is in your PATH

**Tests fail with "staticcheck not found":**

```bash
go install honnef.co/go/tools/cmd/staticcheck@latest
```

**Build fails with permission errors:**

```bash
chmod +x scripts/*.sh
```

**Coverage files not generated:**

- Ensure tests are actually running (check test output)
- Coverage files are created in project root

### Getting Help

1. Run any script with `--help` for usage information
2. Check the main project documentation in `docs/`
3. Look at existing code for patterns and examples
4. Open an issue for bugs or feature requests

## Project Structure

```
jot/
├── scripts/           # Development automation
│   ├── setup.sh      # Environment setup
│   ├── test.sh       # Test runner
│   ├── lint.sh       # Code quality
│   └── clean.sh      # Cleanup
├── cmd/              # CLI commands
├── internal/         # Internal packages
├── docs/             # Documentation
├── Makefile          # Build automation
├── .editorconfig     # Editor configuration
└── tools.go          # Development tool versions
```

This structure follows Go community best practices for project organization and tooling.

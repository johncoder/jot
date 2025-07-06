# Contributing to jot

Thank you for your interest in contributing to jot! This document provides guidelines for contributing to the project.

## Quick Start

For new contributors:

```bash
# 1. Clone the repository
git clone https://github.com/johncoder/jot.git
cd jot

# 2. Set up development environment
make setup

# 3. Build and test
make build
make test
```

## Development Workflow

### Before You Start

1. **Check existing issues** - Look for existing issues or discussions related to your contribution
2. **Open an issue** - For significant changes, open an issue to discuss the approach first
3. **Fork the repository** - Create your own fork to work on

### Making Changes

1. **Create a feature branch** from `main`
2. **Write tests** - Add tests for new functionality
3. **Follow coding standards** - Use `make lint` to check code quality
4. **Test your changes** - Run `make test` to ensure all tests pass
5. **Update documentation** - Update relevant documentation in `docs/`

### Coding Standards

- Follow Go best practices and idiomatic code
- Use `go fmt` for consistent formatting
- Write clear, descriptive commit messages
- Add tests for new functionality
- Update documentation for user-facing changes

### Code Quality

All contributions must pass:
- `make lint` - Code formatting and static analysis
- `make test` - Unit and integration tests
- `make build` - Cross-platform build verification

## Pull Request Process

1. **Create a pull request** with a clear description of changes
2. **Reference related issues** using keywords like "fixes #123"
3. **Provide context** - Explain the problem and your solution
4. **Include test steps** - Describe how to test your changes
5. **Be responsive** - Address feedback promptly

## Types of Contributions

### Bug Reports
- Use the bug report template
- Include reproduction steps
- Provide system information

### Feature Requests
- Use the feature request template
- Explain the use case and benefits
- Consider implementation complexity

### Documentation
- Fix typos, improve clarity
- Add examples and usage guides
- Update command documentation

### Code
- Bug fixes
- New features
- Performance improvements
- Test coverage improvements

## Development Environment

### Setup
```bash
./scripts/setup.sh --help    # See setup options
make setup                   # Automated setup
```

### Testing
```bash
make test                    # Run all tests
make test-unit              # Unit tests only
make test-integration       # Integration tests only
```

### Building
```bash
make build                  # Build for current platform
make build-all             # Build for all platforms
make install               # Install locally
```

## Getting Help

- **Documentation** - Check the [docs/](docs/) directory
- **Issues** - Search existing issues or create a new one
- **Development Guide** - See [DEVELOPMENT.md](DEVELOPMENT.md) for detailed setup

## Code of Conduct

Please note that this project has a [Code of Conduct](CODE_OF_CONDUCT.md). By participating in this project, you agree to abide by its terms.

## License

By contributing to jot, you agree that your contributions will be licensed under the MIT License.

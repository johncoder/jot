# Development Guide

This document provides information for developers working on the jot project.

For contributing guidelines, see [CONTRIBUTING.md](../CONTRIBUTING.md).

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

See [CONTRIBUTING.md](../CONTRIBUTING.md) for complete development workflow and guidelines.

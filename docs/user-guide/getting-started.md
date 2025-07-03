# Getting Started

Welcome to jot! This guide will get you up and running in just a few minutes.

## What is jot?

jot is a git-inspired CLI tool for capturing, organizing, and finding notes directly from your terminal. It's designed for developers and knowledge workers who want fast, frictionless note-taking without leaving their workflow.

Think of it as your personal knowledge management system that lives where you work - in the terminal.

## Installation

### Quick Install (Recommended)

**Linux/macOS:**

```bash
curl -sSL https://raw.githubusercontent.com/johncoder/jot/main/install.sh | sh
```

This installs to `~/.local/bin/jot` and provides helpful PATH setup instructions.

### Package Managers

**Homebrew:**

```bash
brew install johncoder/tap/jot
```

**Go Install:**

```bash
go install github.com/johncoder/jot@latest
```

### Manual Download

Download pre-built binaries from [GitHub Releases](https://github.com/johncoder/jot/releases).

## First Steps

### 1. Initialize Your First Workspace

```bash
# Create a directory for your notes
mkdir ~/notes
cd ~/notes

# Initialize jot workspace
jot init
```

This creates:
- `inbox.md` - Where new notes are captured
- `lib/` - Directory for organized notes
- `.jot/` - Internal jot data (logs, config, templates)

### 2. Capture Your First Note

```bash
# Quick note capture (opens your editor)
jot capture

# Or capture directly from command line
jot capture --content "Remember to update the documentation"

# Or pipe content
echo "Meeting notes from standup" | jot capture
```

### 3. Check Your Workspace

```bash
# See what's in your workspace
jot status

# View your inbox
cat inbox.md
```

### 4. Organize Your Notes

```bash
# Move notes from inbox to organized files
jot refile

# This opens an interactive interface to move notes to files in lib/
```

### 5. Find Your Notes

```bash
# Search through all notes
jot find "documentation"

# List all files in workspace
jot files
```

## Core Concepts

### Workspace Structure

```
your-notes/
├── inbox.md          # New notes go here
├── lib/              # Organized notes
│   ├── work.md
│   ├── personal.md
│   └── projects/
│       └── project-x.md
└── .jot/             # Internal jot data
    ├── config.json
    ├── logs/
    └── templates/
```

### Basic Workflow

1. **Capture** - Quickly save thoughts to `inbox.md`
2. **Refile** - Move notes from inbox to organized topics in `lib/`
3. **Find** - Search through all your notes
4. **Archive** - Move old notes to long-term storage

## Next Steps

- **[Basic Workflows](basic-workflows.md)** - Learn common patterns and use cases
- **[Command Reference](commands.md)** - Explore all available commands
- **[Templates](templates.md)** - Create structured notes with templates
- **[Configuration](configuration.md)** - Customize jot for your needs

## Quick Reference

```bash
# Essential commands
jot capture           # Capture a new note
jot refile            # Organize notes from inbox
jot find <query>      # Search through notes
jot status            # Check workspace status
jot doctor            # Diagnose issues

# Get help
jot --help            # General help
jot [command] --help  # Command-specific help
```

## Need Help?

- Run `jot doctor` to check for common issues
- Check the **[Troubleshooting Guide](troubleshooting.md)**
- See **[Examples](../examples/)** for real-world workflows

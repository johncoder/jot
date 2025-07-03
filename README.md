# jot

A git-inspired CLI tool for capturing, refiling, archiving, finding, and maintaining a hub of notes and information.

## Overview

**jot** is designed for tech industry knowledge workers who spend time in the terminal and need notes close to their work environment. It provides a pragmatic, fast, and convenient way to manage notes with workflows inspired by git and other productivity tools.

Think of it as your personal knowledge management system that lives where you work - in the terminal.

## Key Features

- **🚀 Quick Capture**: Instantly capture notes from the terminal with multiple input methods
- **📁 Smart Organization**: Refile notes from inbox to organized topics with flexible targeting
- **🔍 Powerful Search**: Find notes efficiently with full-text search across all files
- **📝 Template System**: Create structured notes with dynamic shell command integration
- **🏥 Health Monitoring**: Built-in diagnostics and repair tools
- **🔧 Code Integration**: Evaluate and extract code blocks from Markdown files
- **⚡ Extensible**: Support for external commands and git-like workflows

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

### From Source

```bash
git clone https://github.com/johncoder/jot.git
cd jot
make install
```

### Verify Installation

```bash
jot --version
jot --help
```

## Quick Start

### 1. Initialize Your First Workspace

```bash
mkdir my-notes && cd my-notes
jot init
```

This creates:

- `inbox.md` - Where all new notes are captured
- `lib/` - Directory for organized notes
- `.jot/` - Internal data, templates, and configuration

### 2. Capture Your First Note

```bash
# Quick note (opens your editor)
jot capture

# Direct input
jot capture --content "This is my first note"

# From a pipe
echo "Meeting notes from standup" | jot capture
```

### 3. Check Your Workspace

```bash
jot status
```

Shows your workspace health, note counts, and recent activity.

### 4. Search Your Notes

```bash
jot find "first note"
jot find "meeting" --limit 5
```

### 5. Organize Your Notes

```bash
# Interactive organization
jot refile

# Move specific notes by index
jot refile 1,3,5 --dest work.md

# Move all notes
jot refile --all --dest topics.md
```

## Core Commands

### `jot init [path]`

Initialize a new jot workspace with the proper directory structure.

```bash
jot init                    # Initialize in current directory
jot init ~/my-notes         # Initialize in specific directory
```

**What it creates:**

- `inbox.md` - Default capture location
- `lib/` - Directory for organized notes
- `.jot/` - Internal data and configuration

### `jot capture [template]`

Capture new notes with multiple input methods and optional templates.

```bash
# Basic capture (opens editor)
jot capture

# Quick content without editor
jot capture --content "Quick thought"

# Use a template
jot capture meeting
jot capture --template standup --content "Completed API design"

# Pipe content
echo "Notes from terminal" | jot capture
echo "Meeting notes" | jot capture meeting
```

**Input Methods:**

- **Editor**: Opens your `$EDITOR` for interactive note writing
- **Direct**: Use `--content "text"` for quick notes
- **Piped**: Pipe content from other commands
- **Template**: Use templates for structured notes

### `jot refile [indices]`

Move notes from inbox to organized files with flexible targeting.

```bash
# Interactive selection
jot refile

# Move specific notes by index
jot refile 1,3,5 --dest work.md
jot refile 1-3 --dest projects.md

# Move all notes
jot refile --all --dest topics.md

# Advanced targeting
jot refile --exact "2025-06-19 10:30" --dest meeting.md
jot refile --pattern "bug|fix" --dest development.md
jot refile --offset 150 --dest current.md  # For editor integration
```

**Targeting Options:**

- **Index**: `1,3,5` or `1-3` for specific notes
- **Exact Match**: `--exact "timestamp"` for precise targeting
- **Pattern**: `--pattern "regex"` for content-based selection
- **All**: `--all` to process entire inbox
- **Offset**: `--offset N` for editor cursor integration

### `jot find <query>`

Search through all notes with full-text search and context display.

```bash
jot find "project meeting"       # Search for phrase
jot find golang --limit 10       # Limit results
jot find todo --archive          # Include archived notes
```

**Features:**

- Full-text search across inbox, lib/, and optionally archive
- Context display with match highlighting
- Relevance-based result ranking
- Configurable result limits

### `jot status`

Show workspace health, statistics, and recent activity.

```bash
jot status                     # Basic status
jot status --verbose           # Detailed information
```

**Information Displayed:**

- Workspace location and structure validation
- Note counts by location (inbox, library, archive)
- Recent activity summary
- Health indicators and warnings

### `jot doctor`

Diagnose and fix common workspace issues.

```bash
jot doctor                     # Check for problems
jot doctor --fix               # Automatically fix issues
```

**Health Checks:**

- Workspace structure integrity
- File permissions and accessibility
- Configuration validation
- External tool availability

### `jot template`

Manage note templates for structured capture with shell command integration.

```bash
jot template list                    # Show all templates
jot template new meeting             # Create new template
jot template edit meeting            # Edit existing template
jot template view meeting            # View template content
jot template approve meeting         # Approve for shell execution
```

**Template Features:**

- Markdown format with embedded shell commands
- Dynamic content using `$(command)` syntax
- Security approval workflow for shell commands
- Integration with capture command

### `jot archive`

Initialize archive structure for long-term note storage.

```bash
jot archive                    # Create archive structure
```

**Current Functionality:**

- Creates `.jot/archive/` directory structure
- Sets up monthly archive files
- Foundation for future automated archiving

_Note: Interactive archiving features are planned for future releases._

### `jot eval`

Evaluate code blocks in Markdown files with approval-based security.

```bash
jot eval notes.md                           # List evaluable blocks
jot eval notes.md python_block              # Execute specific block
jot eval notes.md --all                     # Execute all approved blocks
jot eval notes.md block_name --approve      # Approve block for execution
jot eval --list-approved                    # Show all approved blocks
```

**Features:**

- Standards-compliant metadata parsing
- Session-based execution for persistent interpreters
- Security approval workflow
- Result integration into Markdown files

### `jot tangle`

Extract code blocks from Markdown files into standalone source files.

```bash
jot tangle notes.md              # Extract code blocks
jot tangle docs/tutorial.md      # Process tutorial file
```

**Features:**

- Extracts blocks with `:tangle` or `:file` headers
- Creates directories as needed
- Supports multiple programming languages

## Advanced Features

### Template System

Templates enable structured note capture with dynamic content generation.

**Creating Templates:**

```bash
jot template new meeting
```

**Example Template** (`.jot/templates/meeting.md`):

```markdown
# Meeting Notes - $(date '+%Y-%m-%d %H:%M')

**Date:** $(date '+%Y-%m-%d')
**Project:** $(git branch --show-current 2>/dev/null || echo "Unknown")

## Attendees

-

## Agenda

-

## Notes

## Action Items

- [ ]
```

**Using Templates:**

```bash
jot capture meeting                    # Use template in editor
jot capture meeting --content "Brief notes"
echo "Additional notes" | jot capture meeting
```

**Security Model:**

Templates with shell commands require explicit approval:

```bash
jot template approve meeting           # Approve template
```

This creates a SHA256 hash stored in `.jot/template_permissions`. If the template changes, it must be re-approved.

### External Commands

Like git, jot supports external commands for extensibility:

```bash
# If you have 'jot-sync' in your PATH:
jot sync                              # Runs jot-sync

# Any 'jot-*' executable becomes a subcommand
```

### Code Block Integration

**Eval Example:**

````markdown
```python :name hello :session main
x = 42
print(f"Hello, the answer is {x}")
```
````

```bash
jot eval notes.md hello --approve     # Approve and execute
```

**Tangle Example:**

````markdown
```bash :tangle scripts/setup.sh
#!/bin/bash
echo "Setting up project..."
npm install
```
````

```bash
jot tangle notes.md                   # Extracts to scripts/setup.sh
```

## Workspace Structure

A jot workspace maintains a clean, organized structure:

```
your-workspace/
├── inbox.md              # All new notes start here
├── lib/                  # Organized notes
│   ├── projects/         # Project-related notes
│   ├── meetings/         # Meeting notes
│   ├── ideas.md          # Ideas and thoughts
│   └── work.md           # Work-related notes
└── .jot/                 # Internal data (hidden)
    ├── templates/        # Note templates
    ├── archive/          # Archived notes
    ├── template_permissions  # Template approvals
    └── config/           # Configuration files
```

## Configuration

jot uses JSON5 configuration files for customization:

- **Global**: `~/.jotrc` (affects all workspaces)
- **Workspace**: `.jotrc` (workspace-specific settings)

**Example Configuration:**

```json5
{
  // Editor settings
  editor: {
    command: "code", // Override $EDITOR
    args: ["--wait"],
  },

  // Search settings
  search: {
    default_limit: 20,
    include_archive: false,
  },

  // Template settings
  templates: {
    default_location: ".jot/templates",
  },
}
```

**Environment Variables:**

- `JOT_EDITOR` - Override editor
- `JOT_CONFIG` - Custom config file location
- Standard: `$EDITOR`, `$VISUAL`, `$PAGER`

## Integration & Workflow

### Editor Integration

jot respects your editor preferences:

```bash
export EDITOR="code --wait"    # VS Code
export EDITOR="vim"            # Vim
export EDITOR="emacs"          # Emacs
```

### Git Integration

Since notes are just Markdown files, they work perfectly with git:

```bash
cd my-notes
git init
git add .
git commit -m "Initial notes"
```

### Terminal Workflow

Example daily workflow:

```bash
# Morning capture
echo "TODO: Review pull requests" | jot capture

# During meetings
jot capture meeting

# End of day organization
jot refile --all --dest today.md

# Search when needed
jot find "pull request"

# Check workspace health
jot status
```

## Requirements

- **Go** (version specified in go.mod - for building from source)
- **Text Editor** (configured via `$EDITOR` or `$VISUAL`)
- **Git** (recommended for version control)
- **Unix-like Shell** (bash, zsh, fish - Windows compatible)

## Philosophy

jot follows these guiding principles:

- **Pragmatism**: Focused on real-world workflows and daily use
- **Speed**: Fast capture and retrieval with minimal friction
- **Convenience**: Always accessible from your terminal
- **Interoperability**: Works with existing tools (git, editors, shell)
- **Simplicity**: Markdown files, simple structure, clear commands
- **Extensibility**: External commands, templates, and future plugins

## Roadmap

Current functionality is stable and production-ready. Planned features include:

- **Enhanced Archive**: Interactive archiving with date-based rules
- **Plugin System**: Extensible architecture for community plugins
- **Sync Integration**: Built-in synchronization options
- **Enhanced Search**: Tag-based search and advanced filtering
- **Mobile Companion**: Companion apps for mobile capture

## Contributing

Contributions are welcome! Please read the documentation in `docs/` for development guidelines and project structure.

Key areas for contribution:

- Additional template examples
- External command integrations
- Documentation improvements
- Bug reports and feature requests

## Documentation

For detailed documentation, see the `docs/` directory:

- [Product Requirements](docs/jot_prd.md)
- [Implementation Summary](docs/IMPLEMENTATION_SUMMARY.md)
- [Roadmap](docs/roadmap.md)
- [Project Details](docs/projects/)

## License

[License information would go here]

---

**jot** - Because your notes should be as close to your code as your terminal is to your workflow.

_Fast • Simple • Extensible • Terminal-Native_

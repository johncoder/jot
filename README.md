# jot

A git-inspired CLI tool for capturing, refiling, archiving, finding, and maintaining a hub of notes and information.

## Overview

**jot** is designed for tech industry knowledge workers who spend time in the terminal and need notes close to their work environment. It provides a pragmatic, fast, and convenient way to manage notes with workflows inspired by git and other productivity tools.

## Key Features

- **🚀 Quick Capture**: Instantly capture notes from the terminal
- **📁 Smart Organization**: Refile notes from inbox to organized topics
- **🗂️ Archive System**: Archive notes for long-term storage
- **🔍 Powerful Search**: Find notes efficiently with full-text search
- **🏥 Health Monitoring**: Built-in diagnostics and repair tools
- **📝 Template System**: Create structured notes with dynamic templates
- **🔧 Extensible**: Support for external commands and plugins

## Installation

### From Source

```bash
git clone https://github.com/johncoder/jot.git
cd jot
go build -o jot .
```

### Quick Start

1. **Initialize a workspace**:
   ```bash
   mkdir my-notes && cd my-notes
   jot init
   ```

2. **Capture your first note**:
   ```bash
   jot capture --note "This is my first note"
   ```

3. **Check your workspace status**:
   ```bash
   jot status
   ```

4. **Search your notes**:
   ```bash
   jot find "first note"
   ```

## Core Commands

### `jot init [path]`
Initialize a new jot workspace with the proper directory structure.

```bash
jot init                    # Initialize in current directory
jot init ~/my-notes         # Initialize in specific directory
```

Creates:
- `inbox.md` - Default capture location
- `lib/` - Directory for organized notes
- `.jot/` - Internal data and configuration

### `jot capture`
Capture new notes quickly with multiple input methods.

```bash
jot capture                           # Open editor for note
jot capture --note "Quick thought"    # Direct input
jot capture --stdin                   # Read from stdin
echo "Note content" | jot capture     # Pipe input
jot capture --template meeting        # Use a template
```

### `jot refile`
Move notes from inbox to organized files in the library.

```bash
jot refile                            # Interactive refile mode
jot refile --index 1,3,5              # Refile specific notes by index
jot refile --index 1-3                # Refile range of notes
jot refile --exact "2025-06-06"       # Refile by exact match
jot refile --pattern "Meeting|Task"   # Refile by regex pattern
```

### `jot find <query>`
Search through all notes with full-text search.

```bash
jot find "project meeting"            # Search for text
jot find --limit 5 "important"        # Limit results
```

### `jot archive`
Archive notes for long-term storage with automatic organization.

```bash
jot archive                           # Archive old notes
```

### `jot status`
Show workspace health and statistics.

```bash
jot status                            # Display workspace overview
```

### `jot doctor`
Diagnose and fix workspace issues.

```bash
jot doctor                            # Check for problems
jot doctor --fix                      # Automatically fix issues
```

### `jot template`
Manage note templates for structured capture.

```bash
jot template list                     # Show available templates
jot template new meeting              # Create new template
jot template edit meeting             # Edit existing template
jot template approve meeting          # Approve template for execution
```

## Workspace Structure

A jot workspace has a simple, organized structure:

```
your-workspace/
├── inbox.md              # Default capture location
├── lib/                  # Organized notes directory
│   ├── projects/
│   ├── meetings/
│   └── ...
└── .jot/                 # Internal data (hidden)
    ├── config/
    ├── templates/
    └── ...
```

## Configuration

jot uses JSON5 configuration files for customization:

- **Global**: `~/.jotrc`
- **Workspace**: `.jotrc` (in workspace root)

Environment variables can override configuration settings.

## Template System

Templates enable structured note capture with dynamic content:

```markdown
---
title: "Meeting Notes"
tags: ["meeting", "work"]
refile_target: "lib/meetings/"
---

# Meeting: {{.title}}

**Date**: {{shell "date +%Y-%m-%d"}}
**Attendees**: 

## Agenda

## Notes

## Action Items
```

Templates support:
- YAML frontmatter for metadata
- Shell command execution for dynamic content
- Security approval workflow for commands
- Integration with capture workflow

## External Commands

Like git, jot supports external commands. Any executable named `jot-<command>` in your PATH becomes available as `jot <command>`.

```bash
# If you have 'jot-sync' in PATH:
jot sync                              # Runs jot-sync
```

## Integration

jot integrates seamlessly with your existing workflow:

- **Editor Integration**: Uses `$EDITOR` or `$VISUAL` for note editing
- **Pager Support**: Respects `$PAGER` for output display
- **Markdown Output**: All notes are standard Markdown files
- **Cross-Platform**: Works on Linux, macOS, and Windows

## Philosophy

jot follows these guiding principles:

- **Pragmatism**: Focused on real-world workflows
- **Speed**: Fast capture and retrieval with minimal friction
- **Convenience**: Always accessible from your terminal
- **Purpose-driven**: Notes serve to complete tasks or recall information
- **Interoperability**: Works well with other tools and workflows

## Requirements

- Go 1.24 or later (for building from source)
- Git (recommended for workspace versioning)
- A text editor (configured via `$EDITOR` or `$VISUAL`)

## Contributing

Contributions are welcome! Please read the documentation in `docs/` for development guidelines and project structure.

## License

[License information would go here]

## Documentation

For detailed documentation, see the `docs/` directory:

- [Product Requirements](docs/jot_prd.md)
- [Implementation Summary](docs/IMPLEMENTATION_SUMMARY.md)
- [Roadmap](docs/roadmap.md)
- [Project Details](docs/projects/)

---

**jot** - Because your notes should be as close to your code as your terminal is to your workflow.

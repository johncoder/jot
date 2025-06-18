# jot: A Git-like CLI Note System

## One-Sentence Summary
A git-inspired CLI tool for capturing, refiling, archiving, finding, and maintaining a hub of notes and information, designed for pragmatic workflows.

## Target Users
Tech industry knowledge workers—people who spend time in a terminal and need notes close to their work environment.

## Core Use Cases & Workflows
- Quickly capturing ideas and notes from the terminal
- Refiling notes to different topics or categories
- Archiving notes for long-term storage and retrieval
- Finding and searching through notes efficiently
- Maintaining a central hub of information, close to the user's workflow
- Pragmatic features inspired by git and other productivity tools

## Key Features & Commands (Initial Version)
- `jot capture`: Quickly add a new note from the terminal
- `jot refile`: Move or re-categorize a note
- `jot archive`: Archive notes for long-term storage
- `jot find`: Search and retrieve notes efficiently
- `jot project`: Organize notes by project or context
- Additional maintenance and organizational features to be determined organically as the tool evolves

## Guiding Principles & Philosophy
- Pragmatism: Focused on real-world workflows and needs, not theoretical ideals
- Speed: Fast capture and retrieval, minimizing friction
- Convenience: Always accessible from the terminal, at your fingertips
- Purpose-driven note-taking: 
  1. To complete a task
  2. To recall information in the future
- Inspired by org-capture in Emacs, but leverages Markdown for its ubiquity and compatibility
- Designed for robust integration with other tools and workflows

## Interoperability & Integration
- Standalone executable with CLI flags to control behavior
- Editor integration: When text input is needed, `jot` launches the user's default editor (like `git commit`)
- Supports multiple input methods for note capture:
  - Editor prompt (default)
  - Piped input via `--stdin`
  - Direct argument via `--note <note text>`
- Output is Markdown files for compatibility and easy integration with other tools
- Designed for robust terminal-based workflows, with potential for further integrations as needed

## Note Organization & Storage
- All captured notes are appended to an `inbox.md` file by default
- A `.jot` directory (next to `inbox.md`) stores internal implementation details (e.g., SQLite databases, log files, etc.)
- A `lib` directory (next to `inbox.md`) contains additional markdown files for organizing and refiling notes
- This structure keeps user-facing notes simple and accessible, while internal data is separated and hidden

## Security, Privacy & Data Portability
- Local-only storage: All notes and data remain on the user's machine
- No encryption in the initial version; future support for GPG-based encryption is a goal
- Notes are stored in plain Markdown for easy export, import, and portability
- No cloud sync or remote storage by default

## Configuration
- Minimal configuration in the initial version to keep setup simple
- Default locations for `inbox.md`, `.jot`, and `lib` directories
- Support for environment variables or a config file (e.g., `.jotrc`) to override default paths and behaviors in the future
- Future consideration: extensibility, plugin support, and advanced customization options

### Configuration Details
- Jot operates in the context of a folder containing a `.jot` directory (or any parent directory containing one)
- Supports named locations (e.g., `documents`) configured in a global config file at `~/.jotrc`
- `~/.jotrc` allows users to define and manage multiple note locations for different contexts or projects
- Jot automatically detects the working context based on the current directory and `.jot` presence

## Error Handling & Documentation
- Commands exit with a non-zero status code on errors, following standard CLI conventions
- Usage information is available via `--help`, and should be compatible with `man` and `info` documentation systems
- Clear, concise error messages and help output to guide users

## Roadmap & Project Planning
- The product roadmap is maintained in a separate `roadmap.md` file
- Each major project or initiative is documented in its own markdown file under a `projects/` directory
- The roadmap and project files are updated iteratively as the product evolves

### Current Projects
- [CLI & Config Setup](projects/cli_and_config_setup.md)
- [jot init (Folder Structure)](projects/jot_init.md)
- [jot capture](projects/jot_capture.md)
- [jot refile](projects/jot_refile.md)
- [jot archive](projects/jot_archive.md)
- [jot find](projects/jot_find.md)
- [jot Maintenance Commands](projects/jot_maintenance.md)
- [Editor Integration](projects/editor_integration.md)

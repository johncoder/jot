# Project: CLI & Config Setup

## Objective
Establish the basic CLI structure for jot and implement configuration file support.

## Scope
- CLI entry point (`jot` command)
- Argument parsing and subcommand structure
- Support for JSON5 config files (with comments) for both global (`~/.jotrc`) and project-specific (`.jotrc` in project root) configuration
- Environment variable overrides
- Help output and error handling

## Deliverables
- CLI skeleton with subcommand placeholders
- Config file loading logic (JSON5)
- Example global and project `.jotrc` files
- Basic documentation for setup and usage

## Next Steps
- Implement CLI entry point and argument parsing
- Implement config file loading (JSON5) with support for both global and project configs
- Add environment variable support
- Add help and error output
- Provide example `.jotrc` files

## CLI Entry Point & Argument Parsing (Proposal)
- The `jot` command acts as the main entry point, inspired by git's architecture
- Subcommands are invoked as `jot <subcommand>`, e.g., `jot capture`, `jot refile`, etc.
- If a subcommand is not built-in, `jot` will search for an executable named `jot-<subcommand>` in the user's PATH and execute it with the remaining arguments (mirroring git's extensibility)
- Arguments follow traditional Unix conventions (e.g., `--help`, `-h`, `--version`, etc.)
- Helpful usage message is displayed if no subcommand is given or if `--help` is passed
- This structure allows for future extensibility and a familiar CLI UX for Unix users

## Editor & Pager Integration (Considerations)
- For commands requiring text input (e.g., `jot capture`), drop the user into their configured editor (defaulting to `$EDITOR` or `$VISUAL`), mirroring `git commit` behavior
- Implement a reusable pattern for editor invocation, so it can be leveraged by multiple commands
- Support a `--pager` flag for commands that produce lengthy output, piping results into `less` (or the user's configured pager, defaulting to `$PAGER`)
- Ensure both editor and pager integration follow Unix conventions and respect user environment variables

## Installation (Requirements & Considerations)
- Support installation on Linux, macOS, and Windows
- Distribute pre-built binaries for all supported platforms
- Provide a simple install script (e.g., via curl) for easy setup
- Package manager distribution (e.g., Homebrew, apt, etc.) is deferred for later
- The `jot` binary should support a flag (e.g., `--print-path-modification`) to output the necessary shell command to add it to the user's PATH; installation scripts will leverage this
- Shell completions are a future enhancement
- Build in development tooling to manage dependencies and keep them up to date
- Use SQLite and standard CLI tools, but be mindful of Windows compatibility and equivalents
- Where permitted by OSS licenses, attempt to check for and install dependencies automatically, or provide clear instructions if not possible

## Go Implementation Proposal
- Implement the CLI and config system in Go for cross-platform support and easy binary distribution
- Anticipated Go packages:
  - `spf13/cobra`: For CLI argument parsing and subcommand structure (widely used, git-like UX)
  - `spf13/viper`: For configuration file loading (supports JSON5 via plugins, environment variables, and project/global config)
  - `go-sqlite/sqlite3` or `mattn/go-sqlite3`: For SQLite integration (future-proofing for note indexing/search)
  - `os/exec`, `os/user`, `os`, `io/ioutil`, `path/filepath`: For process management, file system, and editor/pager invocation
  - `github.com/json5/json5`: For direct JSON5 parsing if needed (Viper may require a plugin or custom loader)
  - `github.com/mitchellh/go-homedir`: For cross-platform home directory resolution
  - `github.com/olekukonko/tablewriter` or similar: For pretty terminal output (optional)
- Use Go's standard library for shell command execution, environment variable handling, and cross-platform compatibility
- Build scripts for cross-compiling binaries for Linux, macOS, and Windows
- Provide a simple install script (e.g., via curl) that downloads the correct binary and prints a shell command for PATH modification
- Future: Explore Go packages for interactive menus (e.g., `charmbracelet/bubbletea`, `manifoldco/promptui`, or `AlecAivazis/survey`)

## Implementation Tasks
1. Scaffold Go project structure and initialize Go module
2. Add and configure `spf13/cobra` for CLI entry point and subcommand parsing
3. Implement subcommand dispatch logic, including external `jot-<subcommand>` executable support
4. Add `--help`, `-h`, and `--version` flags with usage output
5. Integrate `spf13/viper` for config file loading (supporting JSON5 for both global and project configs)
6. Add environment variable overrides for config values
7. Implement reusable editor invocation pattern (using `$EDITOR`/`$VISUAL`)
8. Implement reusable pager invocation pattern (using `$PAGER`/`less` and `--pager` flag)
9. Add a CLI flag (e.g., `--print-path-modification`) to output shell command for PATH modification
10. Write a simple install script (e.g., via curl) for binary download and PATH setup
11. Document installation, configuration, and usage basics
12. Add development tooling for dependency management and cross-compilation
13. Research and document Windows compatibility and dependency handling
14. Provide example `.jotrc` files (global and project)

### Related Projects
- [Interactive Menu System](projects/interactive_menu_system.md)

---
applyTo: '**'
---
# jot Coding Standards, Domain Knowledge, and Preferences

## Domain Knowledge
- jot is a git-inspired CLI tool for capturing, refiling, archiving, finding, and maintaining a hub of notes and information.
- Target users: tech industry knowledge workers who use the terminal and need notes close to their workflow.
- Core use cases: quick note capture, refiling to topics, archiving, searching, and maintaining a central information hub.
- Notes are stored in Markdown: `inbox.md` for capture, `lib/` for organized notes, `.jot/` for internal data (e.g., SQLite, logs).
- CLI is modeled after git: subcommands, external command support, Unix-style arguments, extensibility, and integration with editors/pagers.
- Configuration uses JSON5 files (`~/.jotrc` for global, `.jotrc` for project), with environment variable overrides.
- Documentation and roadmap are maintained in `docs/` and `docs/projects/`.

## Coding Standards
- Use Go as the primary implementation language.
- Follow idiomatic Go practices: clear naming, error handling, and modular code.
- Always use `go fmt` to ensure consistent code formatting.
- Always use parameterized SQL queries to prevent SQL injection.
- Use raw SQL files for all database queries and embed them in the executable (preferably using `go:embed`).
- Use `spf13/cobra` for CLI structure and `spf13/viper` for configuration.
- Support cross-platform compatibility (Linux, macOS, Windows).
- Ensure commands exit with appropriate status codes on error.
- Respect user environment variables for editor (`$EDITOR`/`$VISUAL`) and pager (`$PAGER`), and provide sensible fallbacks.
- Write clear, concise, and user-friendly CLI help output.
- Provide example config files and usage documentation.
- Favor pragmatism, speed, and convenience in all workflows and code.

## Preferences
- Prioritize minimal setup and fast workflows.
- Design for extensibility (external subcommands, future plugins).
- Use Markdown for all user-facing note content.
- Keep internal data and user-facing notes clearly separated.
- Be mindful of Windows compatibility and equivalents for CLI tools.
- Document all major features, commands, and configuration options in `docs/`.
- Always use context from `docs/` directory when working on any tasks to understand requirements and design decisions.
- Make intelligent guesses about which project documents in `docs/` are most relevant to the task at hand, and ask for confirmation if there is any uncertainty about which documents to reference.

# End of instructions
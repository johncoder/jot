# Project: jot CLI MVP

## Objective
Deliver the first usable version of jot, providing the core CLI, configuration, and essential note management commands.

## Scope
This MVP will include:
- Installation instructions and CLI setup
- `.jotrc` configuration file support
- `jot init`: Initialize a jot folder (creates `inbox.md`, `.jot/`, and `lib/`)
- `jot capture`: Add a new note to `inbox.md`
- `jot refile`: Move notes from `inbox.md` to files in `lib/`
- `jot archive`: Move notes to an archive location
- `jot find`: Search notes
- Maintenance commands as necessary (e.g., `jot status`, `jot doctor`)

## Deliverables
- Working CLI tool with the above commands
- Documentation for installation, configuration, and usage
- Example `.jotrc` file

## Out of Scope
- Encryption and cloud sync
- Advanced configuration and plugin system

## Next Steps
1. Project: CLI & Config Setup
2. Project: jot init (folder structure)
3. Project: jot capture
4. Project: jot refile
5. Project: jot archive
6. Project: jot find
7. Project: Maintenance commands

Each step will be tracked as a separate project file under `projects/`.

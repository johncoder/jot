# jot

A terminal-first note management tool that captures, organizes, and retrieves your thoughts with ease. Inspired by org-mode, it brings its structured power to markdown right where you work.

## Special Note

For a long time, I've wanted a tool that does what org-mode does, but increasingly the industry pulls me away from its natural habitat. The jot CLI is my attempt to bring the best of org-mode to markdown, in a way that fits naturally into my terminal-based workflow, along with abundant opportunities to leverage these features in editors, tools, and scripts.

Vast majorities of this code base were generated with the help of Copilot. Keep this in mind as you read the code and documentation, as there maybe inaccuracies or inconsistencies. Features and functionality of jot were well planned out ahead of unleashing any coding agents, and tested throughout the development process.

The use of AI tools in the development of this project does not imply any special endorsement or affiliation with them, nor does it necessitate continued use of such tools in future development. Simply put, these tools helped make this possible sooner than I could have done it in my spare time.

## Installation

### Quick Install

**Linux/macOS:**

```bash
curl -sSL https://raw.githubusercontent.com/johncoder/jot/main/install.sh | sh
```

This installs to `~/.local/bin/jot` and provides helpful PATH setup instructions.

### Package Managers

**Homebrew:** (Coming Soon)

```bash
brew install johncoder/tap/jot
```

**Go Install:**

```bash
go install github.com/johncoder/jot@latest
```

### Manual Download

Download pre-built binaries from [GitHub Releases](https://github.com/johncoder/jot/releases).

## Quick Start

### Create your workspace

```bash
mkdir my-notes && cd my-notes
# create a jot workspace (.jot, inbox.md, lib/)
jot init
# add workspace to registry
jot workspace add my-notes . 
```

You can create multiple, self-contained workspaces. Each workspace has its own `.jot` directory, `inbox.md` for new notes, and a `lib/` directory for organized notes. Jot has additional commands to manage these workspaces.

```bash
jot status         # a bit of information about your workspace
jot workspace list # see all registered workspaces, default is noted here
jot workspace      # outputs the default workspace path
```

### Capturing

Capturing notes is fast and flexible. By default, jot captures notes to `inbox.md` in your workspace. There are several ways to capture notes:

| Method | Command |
|--------|---------|
| Quick thought | `jot capture --content "Remember to update the API docs"` |
| Longer note | `jot capture` (opens your editor via `EDITOR` env var) |
| Standard input | `echo "My note" \| jot capture` |
| Rendering templates | `jot capture meeting` (uses a template named `meeting`) |

Templates are the preferred way, as they provide structure and consistency. They include front-matter for configuring how notes are stored, and can execute shell commands for dynamic content.

```bash
# add a new template called "note"
# uses EDITOR to open the template file; modify it as needed, save & exit
jot template new note

# verify template creation
# indicates that the new template needs approval
jot template list

# always review your templates, whether you write them yourself or get them
# from others; jot will only allow approved templates to be used
jot template approve note

# Capture using the new template
# uses EDITOR to allow you to write your note; contents is the rendered template
jot capture note
```

#### Sample Template

This is a sample template that just incorporates a level two header:

```raw
---
destination: inbox.md
refile_mode: append
---
##

```

## Documentation

View the [documentation](docs/README.md) for detailed guides on:

- **[Command Reference](docs/commands/README.md)** - Complete command documentation


## Contributing

jot is built with Go and follows standard development practices. See [Contributing Guide](docs/contributing/development.md) for:

**Quick development setup:**
```bash
git clone https://github.com/johncoder/jot.git
cd jot
make setup    # run the first time
make build    # build the binary
make test     # run tests
make install  # Build and install locally
```

## License

MIT License - see [LICENSE](LICENSE) file for details.

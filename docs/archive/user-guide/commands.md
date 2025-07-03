# Command Reference

Complete reference for all jot commands and options.

## Global Options

These options work with all commands:

| Option        | Type   | Description                                 |
| ------------- | ------ | ------------------------------------------- |
| `--config`    | string | Config file (default: `$HOME/.jotrc`)       |
| `--json`      | flag   | Output in JSON format                       |
| `--workspace` | string | Use specific workspace (bypasses discovery) |
| `--help`      | flag   | Show help for any command                   |
| `--version`   | flag   | Show version information                    |

## Core Commands

### `jot init`

Initialize a new jot workspace in the current directory.

```bash
jot init [flags]
```

**What it creates:**

- `inbox.md` - Capture file for new notes
- `lib/` - Directory for organized notes
- `.jot/` - Internal data directory
- `.jot/config.json` - Workspace configuration

**Examples:**

```bash
jot init                    # Initialize in current directory
jot init --workspace ~/work # Initialize specific location
```

### `jot capture`

Capture a new note and add it to inbox.md.

```bash
jot capture [template] [flags]
```

**Flags:**

| Flag          | Type   | Description                                         |
| ------------- | ------ | --------------------------------------------------- |
| `--content`   | string | Note content to append (skips editor)               |
| `--template`  | string | Use a named template for structured capture         |
| `--no-verify` | flag   | Skip hooks verification                             |
| `--note`      | string | Note content to append (legacy alias for --content) |

**Input Methods:**

| Method             | Usage                        | Description                        |
| ------------------ | ---------------------------- | ---------------------------------- |
| **Interactive**    | `jot capture`                | Opens your editor (default)        |
| **Direct content** | `--content "text"`           | Quick notes without editor         |
| **Piped input**    | `echo "text" \| jot capture` | Pipe content from other commands   |
| **Template-based** | `jot capture template`       | Use templates for structured notes |

**Examples:**

```bash
jot capture                               # Open editor
jot capture --content "Quick note"        # Direct append
echo "Notes" | jot capture               # Piped input
jot capture meeting                      # Use meeting template
jot capture --template standup           # Use standup template
jot capture standup --content "Done: API design"  # Template + content
```

### `jot refile`

Move markdown subtrees between files using interactive interface.

```bash
jot refile [flags]
```

**Features:**

| Feature             | Description                                    |
| ------------------- | ---------------------------------------------- |
| Interactive menu    | Visual interface for organizing notes          |
| Move entire notes   | Relocate complete note entries or sections     |
| Create new files    | Generate new destination files during refiling |
| Preview changes     | Review modifications before applying           |
| File path targeting | Specify exact destination paths                |

**Examples:**

```bash
jot refile                             # Interactive refiling
jot refile --json                      # JSON output for scripting
```

### `jot find`

Search through notes with full-text search.

```bash
jot find <query> [files...] [flags]
```

**Features:**

| Feature             | Description                               |
| ------------------- | ----------------------------------------- |
| Full-text search    | Search across all files in workspace      |
| Context lines       | Show lines around matches for context     |
| Regular expressions | Support for regex patterns (if supported) |
| File targeting      | Search specific files or entire workspace |

**Examples:**

```bash
jot find "authentication"              # Basic search
jot find "API.*error" --regex          # Regex search (if supported)
jot find "meeting" lib/work.md         # Search specific file
jot find "TODO" --context 2           # Show context lines
```

### `jot status`

Show workspace status and statistics.

```bash
jot status [flags]
```

**Shows:**

| Information        | Description                            |
| ------------------ | -------------------------------------- |
| Workspace location | Current workspace directory path       |
| Inbox note count   | Number of notes in inbox.md            |
| Library file count | Number of files in lib/ directory      |
| Recent activity    | Summary of recent captures and changes |
| Health status      | Overall workspace health indicators    |

**Examples:**

```bash
jot status                            # Basic status
jot status --json                     # JSON format
```

## File and Content Management

### `jot files`

List and browse workspace files.

```bash
jot files [flags]
```

**Features:**

| Feature           | Description                            |
| ----------------- | -------------------------------------- |
| File listing      | List all markdown files in workspace   |
| File metadata     | Show file sizes and modification times |
| Pattern filtering | Filter by file name patterns           |

**Examples:**

```bash
jot files                            # List all files
jot files --json                     # JSON output
```

### `jot peek`

View specific markdown content without opening files.

```bash
jot peek <file> [flags]
```

**Features:**

| Feature            | Description                            |
| ------------------ | -------------------------------------- |
| Content viewing    | View entire files or specific sections |
| No editor required | Display content without opening editor |
| Pager integration  | Respects `$PAGER` environment variable |

**Examples:**

```bash
jot peek lib/work.md                 # View entire file
jot peek lib/work.md --section "API" # View specific section
```

### `jot archive`

Archive notes to configured archive location.

```bash
jot archive [flags]
```

**Features:**

| Feature                | Description                          |
| ---------------------- | ------------------------------------ |
| Archive old notes      | Move old notes to archive location   |
| Configurable location  | Set custom archive directory         |
| Structure preservation | Maintains original file organization |

**Examples:**

```bash
jot archive                          # Archive old notes
jot archive --dry-run                # Preview what would be archived
```

## Templates

### `jot template`

Manage note templates for structured capture.

```bash
jot template [command]
```

**Subcommands:**

| Command          | Description                         |
| ---------------- | ----------------------------------- |
| `list`           | List available templates            |
| `new <name>`     | Create new template                 |
| `edit <name>`    | Edit existing template              |
| `view <name>`    | View template content               |
| `approve <name>` | Approve template for execution      |
| `render <name>`  | Render template with shell commands |

**Examples:**

```bash
jot template list                    # List templates
jot template new meeting             # Create meeting template
jot template edit standup            # Edit standup template
jot template approve meeting         # Approve for shell execution
```

**Template Features:**

| Feature                    | Description                                 |
| -------------------------- | ------------------------------------------- |
| Shell command execution    | Dynamic content with `$(command)` syntax    |
| Security approval system   | Explicit approval required before execution |
| Metadata and frontmatter   | YAML frontmatter for template configuration |
| Dynamic content generation | Date, git info, system data integration     |

## Code Integration

### `jot eval`

Evaluate code blocks in markdown files with approval-based security.

```bash
jot eval <file> [block_name] [flags]
```

**Features:**

| Feature                | Description                                |
| ---------------------- | ------------------------------------------ |
| Code block execution   | Execute code blocks from markdown files    |
| Multi-language support | Support for multiple programming languages |
| Result integration     | Append execution results back to notes     |
| Security controls      | Safe execution with user approval          |

**Examples:**

```bash
jot eval lib/scripts.md              # List blocks with approval status
jot eval lib/scripts.md python_block # Execute specific approved block
jot eval lib/scripts.md --all        # Execute all approved blocks
jot eval lib/scripts.md block --approve --mode hash  # Approve block
jot eval --list-approved             # Show all approved blocks
```

**For detailed documentation on code evaluation, approval workflows, security model, and advanced features, see the [Code Evaluation Guide](eval.md).**

### `jot tangle`

Extract code blocks into standalone source files.

```bash
jot tangle <file> [flags]
```

**Features:**

| Feature                | Description                                |
| ---------------------- | ------------------------------------------ |
| Code extraction        | Extract code blocks to separate files      |
| File organization      | Maintain organized file structure          |
| Multi-language support | Support for multiple programming languages |

**Examples:**

```bash
jot tangle lib/setup.md              # Extract all code blocks
jot tangle lib/setup.md --output src/ # Extract to specific directory
```

## Workspace Management

### `jot workspace`

Manage multiple jot workspaces.

```bash
jot workspace [command]
```

**Subcommands:**

| Command             | Description                |
| ------------------- | -------------------------- |
| `list`              | List configured workspaces |
| `add <name> <path>` | Add named workspace        |
| `remove <name>`     | Remove workspace           |
| `switch <name>`     | Switch to workspace        |

**Examples:**

```bash
jot workspace list                   # List workspaces
jot workspace add work ~/work-notes  # Add work workspace
jot workspace switch work           # Switch to work workspace
```

## Maintenance and Diagnostics

### `jot doctor`

Diagnose and fix common issues.

```bash
jot doctor [flags]
```

**Checks:**

| Check                  | Description                           |
| ---------------------- | ------------------------------------- |
| Workspace integrity    | Verify workspace structure and files  |
| File permissions       | Check file and directory permissions  |
| Configuration validity | Validate configuration file syntax    |
| Template security      | Review template approvals and content |

**Examples:**

```bash
jot doctor                          # Run all diagnostics
jot doctor --fix                    # Auto-fix issues (if supported)
```

### `jot hooks`

Manage workspace hooks for automation.

```bash
jot hooks [command]
```

**Subcommands:**

| Command          | Description          |
| ---------------- | -------------------- |
| `list`           | List available hooks |
| `enable <hook>`  | Enable hook          |
| `disable <hook>` | Disable hook         |

**Hook Types:**

| Hook           | Trigger                 |
| -------------- | ----------------------- |
| `post-capture` | Runs after note capture |
| `post-refile`  | Runs after refiling     |
| `pre-archive`  | Runs before archiving   |

**Examples:**

```bash
jot hooks list                      # List hooks
jot hooks enable post-capture       # Enable post-capture hook
```

## Output Formats

### JSON Output

Most commands support `--json` for machine-readable output:

```bash
jot status --json                   # Status in JSON
jot find "query" --json             # Search results in JSON
jot files --json                    # File list in JSON
```

JSON output is useful for:

| Use Case                     | Description                                  |
| ---------------------------- | -------------------------------------------- |
| Scripting and automation     | Process command output programmatically      |
| Integration with other tools | Export data to external systems              |
| Processing with `jq`         | Use `jq` for JSON manipulation and filtering |

**Example with jq:**

```bash
jot find "TODO" --json | jq '.[] | .file' | sort | uniq
```

## Configuration Files

### Global Config: `~/.jotrc`

```json5
{
  // Named workspaces
  workspaces: {
    work: "/home/user/work-notes",
    personal: "/home/user/personal-notes",
  },

  // Default settings
  editor: "vim",
  pager: "less",

  // Archive settings
  archive: {
    location: "archive/",
    format: "YYYY/MM",
  },
}
```

### Workspace Config: `.jot/config.json`

```json5
{
  // Workspace-specific settings
  templates: {
    default: "note",
    capture_template: "daily",
  },

  // Hook configuration
  hooks: {
    post_capture: true,
    post_refile: true,
  },
}
```

## Environment Variables

| Variable            | Description                            |
| ------------------- | -------------------------------------- |
| `EDITOR` / `VISUAL` | Default editor for interactive editing |
| `PAGER`             | Default pager for viewing content      |
| `JOT_WORKSPACE`     | Override default workspace discovery   |
| `JOT_CONFIG`        | Override config file location          |

## Exit Codes

jot follows standard CLI conventions:

| Code  | Meaning                               |
| ----- | ------------------------------------- |
| `0`   | Success                               |
| `1`   | General error                         |
| `2`   | Misuse of command (invalid arguments) |
| `126` | Command found but not executable      |
| `127` | Command not found                     |

## See Also

- **[Getting Started](getting-started.md)** - First steps with jot
- **[Basic Workflows](basic-workflows.md)** - Common patterns and use cases
- **[Templates](templates.md)** - Template system guide
- **[Configuration](configuration.md)** - Detailed configuration options

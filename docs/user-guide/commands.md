# Command Reference

Complete reference for all jot commands and options.

## Global Options

These options work with all commands:

- `--config string` - Config file (default: `$HOME/.jotrc`)
- `--json` - Output in JSON format
- `--workspace string` - Use specific workspace (bypasses discovery)
- `--help` - Show help for any command
- `--version` - Show version information

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
- `--content string` - Note content to append (skips editor)
- `--template string` - Use a named template for structured capture
- `--no-verify` - Skip hooks verification
- `--note string` - Note content to append (legacy alias for --content)

**Input Methods:**
1. **Interactive (default)** - Opens your editor
2. **Direct content** - Use `--content` for quick notes
3. **Piped input** - Pipe content from other commands
4. **Template-based** - Use templates for structured notes

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
- Interactive menu for organizing notes
- Move entire notes or sections
- Create new destination files
- Preview changes before applying
- Supports file path targeting

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
- Full-text search across all files
- Context lines around matches
- Support for regular expressions
- Search specific files or entire workspace

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
- Workspace location
- Number of notes in inbox
- File count in lib/
- Recent activity
- Health status

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
- List all markdown files in workspace
- Show file sizes and modification times
- Filter by patterns

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
- View entire files or specific sections
- No editor required
- Respects pager settings

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
- Move old notes to archive
- Configurable archive location
- Maintains file structure

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
- `list` - List available templates
- `new <name>` - Create new template
- `edit <name>` - Edit existing template
- `view <name>` - View template content
- `approve <name>` - Approve template for execution
- `render <name>` - Render template with shell commands

**Examples:**
```bash
jot template list                    # List templates
jot template new meeting             # Create meeting template
jot template edit standup            # Edit standup template
jot template approve meeting         # Approve for shell execution
```

**Template Features:**
- Support for shell command execution
- Security approval system
- Metadata and frontmatter
- Dynamic content generation

## Code Integration

### `jot eval`

Evaluate code blocks in markdown files.

```bash
jot eval <file> [flags]
```

**Features:**
- Execute code blocks from markdown
- Support for multiple languages
- Append results back to notes
- Security controls

**Examples:**
```bash
jot eval lib/scripts.md              # Evaluate code blocks
jot eval lib/scripts.md --lang bash  # Specific language only
```

### `jot tangle`

Extract code blocks into standalone source files.

```bash
jot tangle <file> [flags]
```

**Features:**
- Extract code to separate files
- Maintain file organization
- Support for multiple languages

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
- `list` - List configured workspaces
- `add <name> <path>` - Add named workspace
- `remove <name>` - Remove workspace
- `switch <name>` - Switch to workspace

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
- Workspace integrity
- File permissions
- Configuration validity
- Template security

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
- `list` - List available hooks
- `enable <hook>` - Enable hook
- `disable <hook>` - Disable hook

**Hook Types:**
- `post-capture` - Runs after note capture
- `post-refile` - Runs after refiling
- `pre-archive` - Runs before archiving

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
- Scripting and automation
- Integration with other tools
- Processing with `jq`

**Example with jq:**
```bash
jot find "TODO" --json | jq '.[] | .file' | sort | uniq
```

## Configuration Files

### Global Config: `~/.jotrc`

```json5
{
  // Named workspaces
  "workspaces": {
    "work": "/home/user/work-notes",
    "personal": "/home/user/personal-notes"
  },
  
  // Default settings
  "editor": "vim",
  "pager": "less",
  
  // Archive settings
  "archive": {
    "location": "archive/",
    "format": "YYYY/MM"
  }
}
```

### Workspace Config: `.jot/config.json`

```json5
{
  // Workspace-specific settings
  "templates": {
    "default": "note",
    "capture_template": "daily"
  },
  
  // Hook configuration
  "hooks": {
    "post_capture": true,
    "post_refile": true
  }
}
```

## Environment Variables

- `EDITOR` / `VISUAL` - Default editor for interactive editing
- `PAGER` - Default pager for viewing content
- `JOT_WORKSPACE` - Override default workspace discovery
- `JOT_CONFIG` - Override config file location

## Exit Codes

jot follows standard CLI conventions:

- `0` - Success
- `1` - General error
- `2` - Misuse of command (invalid arguments)
- `126` - Command found but not executable
- `127` - Command not found

## See Also

- **[Getting Started](getting-started.md)** - First steps with jot
- **[Basic Workflows](basic-workflows.md)** - Common patterns and use cases
- **[Templates](templates.md)** - Template system guide
- **[Configuration](configuration.md)** - Detailed configuration options

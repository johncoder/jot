# Configuration

jot uses JSON5 configuration files to customize behavior for your workflow. This guide covers all configuration options and common setups.

## Configuration Files

### Global Config: `~/.jotrc`

The global configuration applies to all jot workspaces and sets defaults.

**Location:** `~/.jotrc` (or `$HOME/.jotrc`)

**Example:**

```json5
{
  // Named workspaces for quick access
  workspaces: {
    work: "/home/user/work-notes",
    personal: "/home/user/personal-notes",
    project: "/home/user/projects/myapp/docs",
  },

  // Default editor and pager
  editor: "vim",
  pager: "less -R",

  // Archive settings
  archive: {
    location: "archive/",
    format: "YYYY/MM",
    auto_archive_days: 90,
  },

  // Default templates
  templates: {
    default_capture: "note",
    daily: "daily-log",
  },

  // Git integration
  git: {
    auto_commit: true,
    commit_message: "jot: automated update",
  },
}
```

### Workspace Config: `.jot/config.json`

Workspace-specific configuration overrides global settings.

**Location:** `.jot/config.json` (in each workspace)

**Example:**

```json5
{
  // Workspace-specific templates
  templates: {
    default_capture: "meeting",
    directories: {
      daily: "lib/daily/",
      meetings: "lib/meetings/",
      projects: "lib/projects/",
    },
  },

  // Hook configuration
  hooks: {
    post_capture: true,
    post_refile: true,
    pre_archive: false,
  },

  // File organization
  organization: {
    inbox_size_warning: 50,
    auto_refile_patterns: [
      { pattern: "standup", destination: "lib/standups.md" },
      { pattern: "meeting", destination: "lib/meetings.md" },
    ],
  },
}
```

## Configuration Options

### Editor and Pager Settings

```json5
{
  // Editor for interactive editing (falls back to $EDITOR, then $VISUAL)
  editor: "code --wait",

  // Pager for viewing content (falls back to $PAGER)
  pager: "bat --style=plain --paging=always",

  // Editor-specific settings
  editor_options: {
    code: ["--wait", "--new-window"],
    vim: ["-c", "set ft=markdown"],
    emacs: ["-nw"],
  },
}
```

### Workspace Management

```json5
{
  workspaces: {
    // Named workspaces
    work: "/path/to/work-notes",
    personal: "~/personal-notes",
    project: "./docs",

    // Default workspace when none specified
    default: "personal",
  },

  // Workspace discovery
  workspace_discovery: {
    search_parent_dirs: true,
    max_depth: 5,
  },
}
```

### Template Configuration

```json5
{
  templates: {
    // Default template for capture
    default_capture: "note",

    // Template directories
    directories: {
      user: "~/.jot/templates",
      global: "/usr/local/share/jot/templates",
    },

    // Template security
    security: {
      require_approval: true,
      auto_approve_safe: false,
      allowed_commands: ["date", "pwd", "whoami"],
    },

    // Template variables
    variables: {
      author: "John Doe",
      email: "john@example.com",
      timezone: "UTC",
    },
  },
}
```

### Archive Settings

```json5
{
  archive: {
    // Archive location relative to workspace
    location: "archive/",

    // Directory structure for archived files
    format: "YYYY/MM", // Options: YYYY, YYYY/MM, YYYY/MM/DD

    // Auto-archive old files
    auto_archive_days: 90,

    // Compression
    compress: true,
    compression_format: "gzip", // Options: gzip, zip, none
  },
}
```

### Search and Finding

```json5
{
  search: {
    // Default search behavior
    case_sensitive: false,
    include_archived: false,
    context_lines: 2,

    // Search index
    enable_index: true,
    index_update_interval: 300, // seconds

    // File patterns to exclude from search
    exclude_patterns: ["*.tmp", ".jot/logs/*", "node_modules/*"],
  },
}
```

### Git Integration

```json5
{
  git: {
    // Automatically commit changes
    auto_commit: true,

    // Commit message templates
    commit_messages: {
      capture: "Add new note: {title}",
      refile: "Organize notes: {count} items moved",
      archive: "Archive old notes: {date}",
    },

    // Auto-push settings
    auto_push: false,
    remote: "origin",
    branch: "main",
  },
}
```

### Output and Formatting

```json5
{
  output: {
    // Default output format
    format: "text", // Options: text, json, yaml

    // Color output
    color: "auto", // Options: auto, always, never

    // Date/time formatting
    date_format: "2006-01-02",
    time_format: "15:04",
    datetime_format: "2006-01-02 15:04",
  },
}
```

## Environment Variables

Environment variables override configuration file settings:

### Core Settings

- `JOT_CONFIG` - Path to global config file
- `JOT_WORKSPACE` - Override workspace discovery
- `JOT_EDITOR` - Override editor setting
- `JOT_PAGER` - Override pager setting

### Standard Environment Variables

- `EDITOR` / `VISUAL` - Default editor (if not in config)
- `PAGER` - Default pager (if not in config)
- `HOME` - Used for config file location

### Example Usage

```bash
# Use specific config
JOT_CONFIG=~/.config/jot/work.json jot capture

# Override workspace
JOT_WORKSPACE=~/project-notes jot status

# Use different editor
JOT_EDITOR="code --wait" jot capture
```

## Common Configuration Examples

### Minimal Setup

```json5
// ~/.jotrc
{
  workspaces: {
    default: "~/notes",
  },
}
```

### Developer Workflow

```json5
// ~/.jotrc
{
  workspaces: {
    work: "~/work-notes",
    personal: "~/personal-notes",
    project: "./docs",
  },

  editor: "code --wait",
  pager: "bat --style=plain",

  templates: {
    default_capture: "dev-note",
  },

  git: {
    auto_commit: true,
    commit_message: "docs: update notes",
  },
}
```

### Team Sharing Setup

```json5
// .jot/config.json (in shared workspace)
{
  templates: {
    directories: {
      shared: "../.jot-templates",
    },
  },

  git: {
    auto_commit: true,
    auto_push: true,
    commit_messages: {
      capture: "Add note by {author}",
      refile: "Organize by {author}",
    },
  },

  hooks: {
    post_capture: true, // Run team notification
    post_refile: true,
  },
}
```

### Multi-Project Setup

```json5
// ~/.jotrc
{
  workspaces: {
    "client-a": "~/clients/client-a/notes",
    "client-b": "~/clients/client-b/notes",
    internal: "~/company/internal-notes",
    personal: "~/personal/notes",
  },

  templates: {
    variables: {
      company: "ACME Corp",
      author: "John Doe",
    },
  },

  archive: {
    format: "YYYY/MM",
    auto_archive_days: 60,
  },
}
```

## Configuration Validation

### Check Configuration

```bash
# Validate current config
jot doctor

# Show effective configuration
jot doctor --show-config

# Test workspace discovery
jot status --verbose
```

### Debug Configuration Issues

```bash
# Show config file locations
jot doctor --config-paths

# Test specific config file
jot --config /path/to/config.json status

# Validate JSON5 syntax
json5 validate ~/.jotrc
```

## Configuration Migration

### Upgrading Configuration

When jot updates configuration format:

1. **Backup current config**

   ```bash
   cp ~/.jotrc ~/.jotrc.backup
   ```

2. **Run migration**

   ```bash
   jot doctor --migrate-config
   ```

3. **Verify new config**
   ```bash
   jot doctor --show-config
   ```

### Manual Migration

Convert old format to new:

```bash
# Old format (v1.0)
{
  "default_workspace": "~/notes",
  "editor": "vim"
}

# New format (v2.0)
{
  "workspaces": {
    "default": "~/notes"
  },
  "editor": "vim"
}
```

## Security Considerations

### Template Security

- Always review templates before approval
- Limit shell command usage
- Use `allowed_commands` whitelist
- Regularly audit approved templates

### File Permissions

```bash
# Secure config files
chmod 600 ~/.jotrc
chmod 600 .jot/config.json

# Secure template permissions
chmod 644 .jot/template_permissions
```

### Multi-User Workspaces

```json5
{
  security: {
    require_template_approval: true,
    shared_templates: false,
    audit_log: true,
  },
}
```

## Troubleshooting Configuration

### Common Issues

1. **Config not loading**

   ```bash
   # Check file exists and is readable
   ls -la ~/.jotrc

   # Validate JSON5 syntax
   json5 validate ~/.jotrc
   ```

2. **Workspace not found**

   ```bash
   # Check workspace paths
   jot workspace list

   # Test workspace discovery
   jot status --verbose
   ```

3. **Template issues**

   ```bash
   # Check template directory
   ls -la .jot/templates/

   # Verify permissions
   jot template list
   ```

### Configuration Precedence

1. Command-line flags (highest priority)
2. Environment variables
3. Workspace config (`.jot/config.json`)
4. Global config (`~/.jotrc`)
5. Built-in defaults (lowest priority)

## See Also

- **[Getting Started](getting-started.md)** - Initial setup
- **[Templates](templates.md)** - Template configuration
- **[Command Reference](commands.md)** - Command-line options
- **[Basic Workflows](basic-workflows.md)** - Configuration in practice

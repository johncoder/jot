[Documentation](../README.md) > Configuration Guide

# Configuration Guide

jot uses a hierarchical configuration system that supports both global and project-specific settings. Configuration is stored in JSON5 format with support for environment variable expansion, enabling flexible and maintainable configuration management.

This guide covers:
- Configuration file structure and format
- Environment variable integration
- Editor and pager configuration
- Workspace management and discovery
- Configuration options and examples

## Configuration Structure

### File Locations

jot looks for configuration files in the following order:

1. **Command-line specified**: `--config /path/to/config.json`
2. **Current directory**: `.jotrc` (project-specific)
3. **Home directory**: `~/.jotrc` (global)

The first configuration file found takes precedence. Project-specific configuration in `.jotrc` overrides global settings for that workspace.

### Configuration Format

Configuration files use JSON5 format, which supports comments, trailing commas, unquoted property names, and single quotes for strings.

## Configuration Options

### Workspace Configuration

#### workspaces

Define named workspaces for quick access:

```json5
{
  "workspaces": {
    "personal": "/home/user/notes",
    "work": "/home/user/work-notes",
    "research": "/home/user/research"
  }
}
```

**Usage:**
```bash
jot workspace use personal
jot capture --workspace work "Meeting notes"
```

#### defaults

Set default values for workspace structure:

```json5
{
  "defaults": {
    "inbox": "inbox.md",      // Default inbox filename
    "lib": "lib",             // Default library directory
    "templates": "templates", // Default templates directory
    "archive": "archive"      // Default archive directory
  }
}
```

### Hook Configuration

| Option | Type | Description | Default |
|--------|------|-------------|---------|
| `enabled` | boolean | Enable or disable the hook system globally | `true` |
| `timeout` | number | Timeout for hook execution (in seconds) | `30` |
| `env` | object | Environment variables passed to hook execution | `{}` |

```json5
{
  "hooks": {
    "enabled": true,
    "timeout": 30,
    "env": {
      "NOTIFICATION_URL": "https://webhooks.example.com/notify",
      "LOG_LEVEL": "info"
    }
  }
}
```

### External Command Configuration

| Option | Type | Description | Default |
|--------|------|-------------|---------|
| `timeout` | number | Timeout for external command execution (in seconds) | `60` |
| `env` | object | Environment variables passed to external commands | `{}` |

```json5
{
  "external": {
    "timeout": 60,
    "env": {
      "API_KEY": "${API_KEY}",
      "BASE_URL": "https://api.example.com",
      "TIMEOUT": "30"
    }
  }
}
```

## Environment Variables

### Editor Integration

jot uses standard environment variables for editor integration:

| Variable | Description | Default |
|----------|-------------|---------|
| `EDITOR` | Primary editor command | `vi` |
| `VISUAL` | Visual editor command | `$EDITOR` |
| `JOT_EDITOR` | jot-specific editor override | `$VISUAL` or `$EDITOR` |

**Example:**
```bash
export EDITOR="code --wait"
export JOT_EDITOR="vim"  # Use vim specifically for jot
```

### Pager Integration

Configure pager for viewing output:

| Variable | Description | Default |
|----------|-------------|---------|
| `PAGER` | Default pager command | `less` |
| `JOT_PAGER` | jot-specific pager override | `$PAGER` |

**Example:**
```bash
export PAGER="less -R"
export JOT_PAGER="cat"  # Disable paging for jot
```

### Runtime Configuration

Override configuration at runtime:

| Variable | Description | Example |
|----------|-------------|---------|
| `JOT_CONFIG` | Custom config file path | `/path/to/config.json` |
| `JOT_WORKSPACE` | Override workspace discovery | `/path/to/workspace` |
| `JOT_HOOKS_ENABLED` | Enable/disable hooks | `true` or `false` |
| `JOT_DEBUG` | Enable debug output | `true` or `false` |

## Example Configuration

```json5
// ~/.jotrc
{
  "workspaces": {
    "personal": "~/notes",
    "work": "~/work-notes",
    "research": "~/research"
  },
  
  "defaults": {
    "inbox": "inbox.md",
    "lib": "reference",
    "templates": "templates",
    "archive": "archive"
  },
  
  "hooks": {
    "enabled": true,
    "timeout": 45,
    "env": {
      "BACKUP_DIR": "~/backups/notes",
      "NOTIFICATION_URL": "https://webhooks.example.com/notify"
    }
  },
  
  "external": {
    "timeout": 120,
    "env": {
      "API_KEY": "${API_KEY}",
      "BASE_URL": "https://api.example.com",
      "JIRA_TOKEN": "${JIRA_TOKEN}"
    }
  }
}
```

## Environment Variable Expansion

Configuration supports environment variable expansion with these patterns:

- `${VAR}` - Required variable (fails if not set)
- `${VAR:-default}` - Optional variable with default value
- `${VAR:+value}` - Value only if variable is set
- `~` - Home directory expansion

```json5
{
  "workspaces": {
    "user": "${HOME}/notes",
    "project": "${PROJECT_ROOT}/docs"
  },
  
  "hooks": {
    "enabled": "${JOT_HOOKS_ENABLED:-true}",
    "env": {
      "API_KEY": "${API_KEY}",
      "LOG_FILE": "${HOME}/.jot/logs/hooks.log"
    }
  }
}
```

## See Also

- [Commands Reference](../commands/README.md) - All jot commands
- [Hook System](../commands/jot-hooks.md) - Hook configuration and usage
- [External Commands](../commands/jot-external.md) - External command integration
- [Workspace Management](../commands/jot-workspace.md) - Workspace configuration
- [Template System](../commands/jot-template.md) - Template configuration
- [JSON Output Reference](../reference/json-output.md) - JSON output configuration

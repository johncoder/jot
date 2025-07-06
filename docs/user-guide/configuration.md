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

### Code Execution Environment

The evaluator system allows code execution within jot documents using external evaluators. jot discovers evaluators through a convention-based approach.

#### Evaluator Discovery

When `jot eval` encounters a code block, it uses this discovery order:

1. **PATH evaluator**: Look for `jot-eval-<lang>` in PATH (e.g., `jot-eval-haskell`)
2. **Built-in evaluator**: Use `jot evaluator <lang>` for supported languages
3. **Error**: No evaluator found

#### Environment Variables

Evaluators receive standardized environment variables:

**Standard Context:**
- `JOT_WORKSPACE_PATH` - Path to current workspace
- `JOT_WORKSPACE_NAME` - Name of current workspace  
- `JOT_CONFIG_FILE` - Path to configuration file

**Eval-Specific Context:**
- `JOT_EVAL_CODE` - The actual code to execute
- `JOT_EVAL_LANG` - Language from code fence or shell parameter
- `JOT_EVAL_FILE` - Source markdown file path
- `JOT_EVAL_BLOCK_NAME` - Block name from eval element
- `JOT_EVAL_CWD` - Working directory for execution
- `JOT_EVAL_TIMEOUT` - Timeout duration (e.g., "30s")
- `JOT_EVAL_ARGS` - Additional arguments from eval element

**Custom Environment Variables:**
- `JOT_EVAL_ENV_*` - Any custom env vars from eval element's `env` parameter

#### Built-in Evaluators

jot provides built-in evaluators for core languages:

- **python3**: Python code execution
- **javascript**: Node.js JavaScript execution  
- **bash**: Bash shell command execution
- **sh**: POSIX shell command execution
- **go**: Go code execution

**Example:**
```bash
# List available evaluators
jot evaluator

# Use built-in Python evaluator
jot evaluator python3

# Create custom evaluator
#!/usr/bin/env bash
# File: /usr/local/bin/jot-eval-ruby
ruby -e "$JOT_EVAL_CODE"
```

#### Custom Evaluators

Create custom evaluators by placing executables named `jot-eval-<lang>` in your `$PATH`:

```bash
#!/usr/bin/env bash
# File: ~/.local/bin/jot-eval-ruby
# Custom Ruby evaluator

# Use environment variables provided by jot
cd "$JOT_EVAL_CWD" || exit 1
echo "Executing Ruby code in $JOT_WORKSPACE_PATH"
ruby -e "$JOT_EVAL_CODE"
```

Make the evaluator executable:
```bash
chmod +x ~/.local/bin/jot-eval-ruby
```
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
- [Evaluator System](../commands/jot-evaluator.md) - Evaluator configuration and usage
- [Code Evaluation](../commands/jot-eval.md) - Code execution commands
- [Hook System](../commands/jot-hooks.md) - Hook configuration and usage
- [External Commands](../commands/jot-external.md) - External command integration
- [Workspace Management](../commands/jot-workspace.md) - Workspace configuration
- [Template System](../commands/jot-template.md) - Template configuration
- [JSON Output Reference](../reference/json-output.md) - JSON output configuration

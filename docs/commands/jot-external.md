[Documentation](../README.md) > [Commands](README.md) > external

# External Commands

## Description

jot supports external commands through a git-style extension system. External commands are executable programs named `jot-<subcommand>` that can be installed anywhere in your PATH. This allows third-party extensions, custom workflows, and specialized tools to integrate seamlessly with jot.

External commands receive full context about the current workspace, configuration, and global flags through environment variables, enabling them to work consistently with jot's ecosystem.

## How External Commands Work

### Command Discovery

When jot receives a subcommand it doesn't recognize, it looks for an executable named `jot-<subcommand>` in your PATH:

```bash
# jot tries to find 'jot-sync' in PATH
jot sync --all

# jot tries to find 'jot-backup' in PATH  
jot backup --destination /backup/path
```

### Environment Variables

External commands receive comprehensive context through environment variables:

#### Core Context

| Variable | Description | Example |
|----------|-------------|---------|
| `JOT_SUBCOMMAND` | The subcommand being executed | `sync` |
| `JOT_JSON_OUTPUT` | Whether JSON output is requested | `true` or `false` |
| `JOT_CONFIG_FILE` | Custom config file path (if specified) | `/path/to/config.json` |
| `JOT_DISCOVERY_METHOD` | How workspace was discovered | `workspace flag`, `current dir`, `none` |

#### Workspace Context

| Variable | Description | Example |
|----------|-------------|---------|
| `JOT_WORKSPACE_ROOT` | Workspace root directory | `/home/user/notes` |
| `JOT_WORKSPACE_NAME` | Workspace name | `personal` |
| `JOT_WORKSPACE_INBOX` | Inbox file path | `/home/user/notes/inbox.md` |
| `JOT_WORKSPACE_LIB` | Library directory path | `/home/user/notes/lib` |
| `JOT_WORKSPACE_JOTDIR` | Internal jot directory | `/home/user/notes/.jot` |

#### Configuration Context

| Variable | Description | Example |
|----------|-------------|---------|
| `JOT_EDITOR` | Configured editor | `vim` |
| `JOT_PAGER` | Configured pager | `less` |

### Global Flag Support

External commands automatically receive parsed global flags:

```bash
# Global flags are parsed and passed to external command
jot sync --workspace personal --json --config /custom/config
```

The external command receives:
- `JOT_WORKSPACE_ROOT` pointing to the personal workspace
- `JOT_JSON_OUTPUT=true`
- `JOT_CONFIG_FILE=/custom/config`
- All other workspace context variables

## Creating External Commands

### Basic External Command

```bash
#!/bin/bash
# File: jot-hello (make executable with chmod +x)

echo "Hello from external command!"
echo "Workspace: $JOT_WORKSPACE_NAME"
echo "Root: $JOT_WORKSPACE_ROOT"
echo "Subcommand: $JOT_SUBCOMMAND"

# Check if JSON output is requested
if [ "$JOT_JSON_OUTPUT" = "true" ]; then
    cat <<EOF
{
  "message": "Hello from external command",
  "workspace": "$JOT_WORKSPACE_NAME",
  "root": "$JOT_WORKSPACE_ROOT",
  "metadata": {
    "success": true,
    "command": "jot hello",
    "execution_time_ms": 5,
    "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  }
}
EOF
else
    echo "Use --json for structured output"
fi
```

### Advanced External Command

```bash
#!/bin/bash
# File: jot-sync (synchronization tool)

set -e

# Check if workspace is available
if [ -z "$JOT_WORKSPACE_ROOT" ]; then
    echo "Error: No workspace found" >&2
    exit 1
fi

# Default options
REMOTE=""
DRY_RUN=false
VERBOSE=false

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --remote)
            REMOTE="$2"
            shift 2
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --verbose|-v)
            VERBOSE=true
            shift
            ;;
        *)
            echo "Unknown option: $1" >&2
            exit 1
            ;;
    esac
done

# Validate required options
if [ -z "$REMOTE" ]; then
    echo "Error: --remote is required" >&2
    exit 1
fi

# Perform synchronization
cd "$JOT_WORKSPACE_ROOT"

if [ "$DRY_RUN" = "true" ]; then
    if [ "$VERBOSE" = "true" ]; then
        echo "Dry run: would sync $JOT_WORKSPACE_ROOT to $REMOTE"
    fi
    rsync -avzn --delete . "$REMOTE/"
else
    if [ "$VERBOSE" = "true" ]; then
        echo "Syncing $JOT_WORKSPACE_ROOT to $REMOTE"
    fi
    rsync -avz --delete . "$REMOTE/"
fi

# JSON output if requested
if [ "$JOT_JSON_OUTPUT" = "true" ]; then
    cat <<EOF
{
  "operation": "sync",
  "workspace": "$JOT_WORKSPACE_NAME",
  "remote": "$REMOTE",
  "dry_run": $DRY_RUN,
  "metadata": {
    "success": true,
    "command": "jot sync",
    "execution_time_ms": 250,
    "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  }
}
EOF
fi
```

### Python External Command

```python
#!/usr/bin/env python3
# File: jot-stats (statistics generator)

import os
import json
import sys
from datetime import datetime
from pathlib import Path

def main():
    # Get workspace context
    workspace_root = os.environ.get('JOT_WORKSPACE_ROOT')
    workspace_name = os.environ.get('JOT_WORKSPACE_NAME')
    json_output = os.environ.get('JOT_JSON_OUTPUT', 'false').lower() == 'true'
    
    if not workspace_root:
        print("Error: No workspace found", file=sys.stderr)
        sys.exit(1)
    
    # Collect statistics
    workspace_path = Path(workspace_root)
    stats = {
        'total_files': 0,
        'total_lines': 0,
        'total_size': 0,
        'markdown_files': 0,
    }
    
    # Count files and lines
    for file_path in workspace_path.rglob('*.md'):
        if file_path.is_file() and not file_path.name.startswith('.'):
            stats['total_files'] += 1
            stats['markdown_files'] += 1
            stats['total_size'] += file_path.stat().st_size
            
            try:
                with file_path.open('r', encoding='utf-8') as f:
                    stats['total_lines'] += sum(1 for _ in f)
            except UnicodeDecodeError:
                # Skip files with encoding issues
                pass
    
    # Output results
    if json_output:
        response = {
            'operation': 'stats',
            'workspace': workspace_name,
            'statistics': stats,
            'metadata': {
                'success': True,
                'command': 'jot stats',
                'execution_time_ms': 100,
                'timestamp': datetime.utcnow().isoformat() + 'Z'
            }
        }
        print(json.dumps(response, indent=2))
    else:
        print(f"Workspace Statistics for '{workspace_name}':")
        print(f"  Total files: {stats['total_files']}")
        print(f"  Markdown files: {stats['markdown_files']}")
        print(f"  Total lines: {stats['total_lines']}")
        print(f"  Total size: {stats['total_size']} bytes")

if __name__ == '__main__':
    main()
```

## Installation and Distribution

### Local Installation

1. **Create the executable**:
   ```bash
   # Create script
   cat > jot-hello << 'EOF'
   #!/bin/bash
   echo "Hello from jot-hello!"
   echo "Workspace: $JOT_WORKSPACE_NAME"
   EOF
   
   # Make executable
   chmod +x jot-hello
   
   # Move to PATH
   sudo mv jot-hello /usr/local/bin/
   ```

2. **Test the command**:
   ```bash
   jot hello
   ```

### Package Distribution

External commands can be distributed as packages:

```bash
# Package structure
jot-extension-pack/
├── bin/
│   ├── jot-sync
│   ├── jot-backup
│   └── jot-stats
├── README.md
└── install.sh
```

Example install script:
```bash
#!/bin/bash
# install.sh

INSTALL_DIR="/usr/local/bin"

for cmd in bin/jot-*; do
    if [ -f "$cmd" ]; then
        echo "Installing $(basename "$cmd")..."
        sudo cp "$cmd" "$INSTALL_DIR/"
        sudo chmod +x "$INSTALL_DIR/$(basename "$cmd")"
    fi
done

echo "Installation complete!"
echo "Available commands:"
ls -1 "$INSTALL_DIR"/jot-* | sed 's/.*jot-/  jot /'
```

## Best Practices

### Command Design

1. **Follow jot conventions**:
   - Use standard exit codes (0 for success, non-zero for errors)
   - Support `--json` flag through `JOT_JSON_OUTPUT` environment variable
   - Provide helpful error messages
   - Document all command options

2. **Respect workspace context**:
   - Always check for `JOT_WORKSPACE_ROOT` before operating on files
   - Use workspace-relative paths when possible
   - Respect the workspace structure (inbox, lib, .jot)

3. **Handle missing context gracefully**:
   - Some commands might work without workspace context
   - Provide clear error messages when workspace is required
   - Consider fallback behaviors for missing context

### Error Handling

```bash
#!/bin/bash
# Good error handling example

# Check for required environment
if [ -z "$JOT_WORKSPACE_ROOT" ]; then
    if [ "$JOT_JSON_OUTPUT" = "true" ]; then
        cat <<EOF
{
  "error": {
    "message": "No workspace found",
    "code": "workspace_not_found"
  },
  "metadata": {
    "success": false,
    "command": "jot $JOT_SUBCOMMAND",
    "execution_time_ms": 1,
    "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  }
}
EOF
    else
        echo "Error: No workspace found" >&2
        echo "Run this command from within a jot workspace" >&2
    fi
    exit 1
fi

# Rest of command logic...
```

### JSON Output

Follow jot's JSON format conventions:

```json
{
  "operation": "command_name",
  "data": {
    // Command-specific data
  },
  "metadata": {
    "success": true,
    "command": "jot command_name",
    "execution_time_ms": 125,
    "timestamp": "2025-01-01T12:00:00Z"
  }
}
```

## Examples

### Usage Examples

```bash
# Basic external command
jot hello

# External command with arguments
jot sync --remote user@server:/backup/notes

# External command with JSON output
jot stats --json

# External command with workspace override
jot backup --workspace personal --destination /backup/

# External command with custom config
jot export --config /custom/config.json --format pdf
```

### Common Use Cases

1. **Synchronization Tools**:
   ```bash
   jot sync --remote backup-server:/notes
   jot push --git-remote origin
   ```

2. **Export and Conversion**:
   ```bash
   jot export --format pdf --output report.pdf
   jot convert --to html --theme dark
   ```

3. **Analysis and Statistics**:
   ```bash
   jot stats --detailed
   jot analyze --report weekly
   ```

4. **Integration Tools**:
   ```bash
   jot slack --channel notes --latest
   jot email --digest weekly
   ```

5. **Backup and Restore**:
   ```bash
   jot backup --destination /backup/notes
   jot restore --from /backup/notes --date 2025-01-01
   ```

## Error Conditions

| Error | Cause | Solution |
|-------|-------|----------|
| `external command not found` | Command not in PATH | Install the external command |
| `permission denied` | Command not executable | Run `chmod +x jot-command` |
| `workspace not found` | No workspace context | Run from within a workspace |
| `command failed` | External command error | Check external command logs |

## Cross-references

- [jot init](jot-init.md) - Creating workspaces for external commands
- [jot workspace](jot-workspace.md) - Managing workspace context
- [JSON Output Reference](../reference/json-output.md) - JSON output format reference
- [Global Options](README.md#global-options) - Global flags passed to external commands

## See Also

- [Global Options](README.md#global-options)
- [Configuration Guide](../user-guide/configuration.md) - External command configuration and environment variables
- [Hook System](jot-hooks.md) - Hook integration with external commands

[Documentation](../README.md) > [Commands](README.md) > hooks

# jot hooks

## Description

The `jot hooks` command manages git-style hooks for workspace automation. Hooks are executable scripts stored in `.jot/hooks/` that run at specific points in jot operations, enabling content modification, notifications, logging, and integration with external tools.

This command supports:
- Listing all hooks in the current workspace
- Installing sample hook scripts for common use cases
- Testing hooks with sample data to verify functionality
- Managing hook lifecycle and activation

## Subcommands

### list

List all hooks in the workspace, showing their status and basic information.

```bash
jot hooks list
```

### install-samples

Install sample hook scripts as templates for common automation tasks.

```bash
jot hooks install-samples
```

### test

Test a specific hook type with sample data to verify functionality.

```bash
jot hooks test <hook-type>
```

## Usage

```bash
jot hooks <subcommand> [options]
```

## Options

### Global Options

See [Global Options](README.md#global-options) for options available to all commands.

### Subcommand-Specific Options

All subcommands support the standard global options. No additional options are specific to individual subcommands.

## Hook Types

jot supports the following hook types:

| Hook Type | Trigger | Purpose | Can Modify Content |
|-----------|---------|---------|-------------------|
| `pre-capture` | Before content capture | Modify content, validate templates | ✅ |
| `post-capture` | After content capture | Notifications, logging, sync | ❌ |
| `pre-refile` | Before content refile | Modify content, validate destinations | ✅ |
| `post-refile` | After content refile | Cleanup, notifications, indexing | ❌ |
| `pre-archive` | Before content archive | Modify content, validate archive | ✅ |
| `post-archive` | After content archive | Cleanup, notifications, sync | ❌ |
| `workspace-change` | When switching workspaces | Environment setup, notifications | ❌ |

## Hook Environment Variables

Each hook type receives specific environment variables:

### Common Variables
- `JOT_WORKSPACE_ROOT` - Workspace root directory
- `JOT_WORKSPACE_NAME` - Workspace name
- `JOT_HOOK_TYPE` - Type of hook being executed

### Hook-Specific Variables

**pre-capture / post-capture:**
- `JOT_TEMPLATE_NAME` - Template used for capture
- `JOT_CAPTURE_FILE` - Target file for capture

**pre-refile / post-refile:**
- `JOT_REFILE_SOURCE` - Source path selector
- `JOT_REFILE_DEST` - Destination path selector

**pre-archive / post-archive:**
- `JOT_ARCHIVE_SOURCE` - Source path selector
- `JOT_ARCHIVE_LOCATION` - Archive destination path

**workspace-change:**
- `JOT_OLD_WORKSPACE` - Previous workspace path
- `JOT_NEW_WORKSPACE` - New workspace path

## Hook Activation

Hooks must be executable to be active:

1. **Sample hooks**: Created with `.sample` extension (inactive)
2. **Active hooks**: Named exactly as hook type (e.g., `pre-capture`)
3. **Permissions**: Must have execute permissions (`chmod +x`)

## Examples

### Basic Hook Management

```bash
# List all hooks in current workspace
jot hooks list

# Install sample hook scripts
jot hooks install-samples

# Test a specific hook type
jot hooks test pre-capture
```

### Hook Activation Workflow

```bash
# Install samples first
jot hooks install-samples

# Copy sample to active hook
cp .jot/hooks/pre-capture.sample .jot/hooks/pre-capture

# Make it executable
chmod +x .jot/hooks/pre-capture

# Test the hook
jot hooks test pre-capture

# Verify it's listed as active
jot hooks list
```

### Hook Development and Testing

```bash
# Test each hook type individually
jot hooks test pre-capture
jot hooks test post-capture
jot hooks test pre-refile
jot hooks test post-refile

# Test with JSON output for debugging
jot hooks test pre-capture --json
```

## Hook Script Examples

### Content Modification Hook (pre-capture)

```bash
#!/bin/bash
# .jot/hooks/pre-capture

# Add timestamp to captured content
TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')
echo "<!-- Captured on $TIMESTAMP -->" >> /tmp/jot_content
cat >> /tmp/jot_content

# Replace stdin with modified content
cat /tmp/jot_content
rm /tmp/jot_content
```

### Notification Hook (post-capture)

```bash
#!/bin/bash
# .jot/hooks/post-capture

# Send notification when content is captured
if command -v notify-send &> /dev/null; then
    notify-send "jot" "Content captured successfully"
fi

# Log to file
echo "$(date): Content captured to $JOT_CAPTURE_FILE" >> "$JOT_WORKSPACE_ROOT/.jot/capture.log"
```

### Integration Hook (post-refile)

```bash
#!/bin/bash
# .jot/hooks/post-refile

# Commit changes to git if workspace is a git repository
if [ -d "$JOT_WORKSPACE_ROOT/.git" ]; then
    cd "$JOT_WORKSPACE_ROOT"
    git add .
    git commit -m "jot: refiled content from $JOT_REFILE_SOURCE to $JOT_REFILE_DEST"
fi
```

## JSON Output

All hook commands support JSON output with `--json`:

### List Hooks JSON

```json
{
  "operation": "list",
  "hooks_dir": "/path/to/workspace/.jot/hooks",
  "hooks": [
    {
      "name": "pre-capture",
      "path": "/path/to/workspace/.jot/hooks/pre-capture",
      "executable": true,
      "sample": false,
      "size": 256,
      "mod_time": "2025-01-01T12:00:00Z"
    }
  ],
  "summary": {
    "total_hooks": 1,
    "executable_hooks": 1,
    "sample_hooks": 0
  },
  "metadata": {
    "success": true,
    "command": "jot hooks list",
    "execution_time_ms": 15,
    "timestamp": "2025-01-01T12:00:00Z"
  }
}
```

### Test Hook JSON

```json
{
  "operation": "test",
  "hook_type": "pre-capture",
  "hook_path": "/path/to/workspace/.jot/hooks/pre-capture",
  "success": true,
  "test_data": {
    "content": "# Test Capture\n\nThis is test content for the capture hook.",
    "extra_env": {
      "JOT_TEMPLATE_NAME": "test-template"
    }
  },
  "result": {
    "exit_code": 0,
    "output": "Hook executed successfully",
    "content": "# Test Capture\n\nModified test content",
    "aborted": false
  },
  "metadata": {
    "success": true,
    "command": "jot hooks test pre-capture",
    "execution_time_ms": 125,
    "timestamp": "2025-01-01T12:00:00Z"
  }
}
```

## Error Conditions

| Error | Cause | Solution |
|-------|-------|----------|
| `no hooks directory found` | No `.jot/hooks/` directory | Run `jot hooks install-samples` |
| `hook not found` | Hook file doesn't exist | Check hook name and path |
| `hook not executable` | Hook lacks execute permissions | Run `chmod +x hookname` |
| `invalid hook type` | Unknown hook type in test | Use valid hook type name |
| `hook failed to execute` | Hook script error | Check hook script for errors |
| `hook execution timeout` | Hook took too long | Optimize hook script performance |

## Hook Best Practices

### Security
- **Validate inputs**: Always validate environment variables and stdin
- **Limit privileges**: Run hooks with minimal required permissions
- **Sanitize content**: Be careful with user-provided content

### Performance
- **Keep hooks fast**: Hooks run synchronously and can slow operations
- **Use background tasks**: For long-running operations, spawn background processes
- **Cache results**: Cache expensive operations where possible

### Error Handling
- **Exit codes**: Use appropriate exit codes (0 for success, non-zero for failure)
- **Logging**: Log errors and important events
- **Graceful degradation**: Don't break user workflows with hook failures

### Testing
- **Test regularly**: Use `jot hooks test` to verify functionality
- **Test edge cases**: Test with various content types and scenarios
- **Version control**: Keep hook scripts in version control

## Integration with Other Commands

Hooks integrate seamlessly with jot operations:

- **capture**: Calls pre-capture and post-capture hooks
- **refile**: Calls pre-refile and post-refile hooks  
- **archive**: Calls pre-archive and post-archive hooks
- **workspace**: Calls workspace-change hooks when switching

### Hook Bypass

Use `--no-verify` flag to skip hook execution:

```bash
jot capture --content "Quick note" --no-verify
jot refile "inbox.md#notes" --to "work.md#tasks" --no-verify
```

## Cross-references

- [jot capture](jot-capture.md) - Content capture with hooks
- [jot refile](jot-refile.md) - Content refile with hooks
- [jot archive](jot-archive.md) - Content archive with hooks
- [jot workspace](jot-workspace.md) - Workspace switching with hooks
- [jot init](jot-init.md) - Workspace initialization
- [jot external](jot-external.md) - External command integration with hooks

## See Also

- [Global Options](README.md#global-options)
- [Configuration Guide](../user-guide/configuration.md) - Hook configuration and environment variables
- [JSON Output Reference](../reference/json-output.md) - JSON output formats

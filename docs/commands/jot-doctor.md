[Documentation](../README.md) > [Commands](README.md) > doctor

# jot doctor

## Description

The `jot doctor` command performs comprehensive health checks on your jot workspace, identifying and optionally fixing common issues. It's the first tool to run when experiencing problems with your workspace.

This command is useful for:
- Diagnosing workspace problems
- Automatically fixing common issues
- Validating workspace integrity
- Checking external tool dependencies

## Usage

## Options

| Option | Description |
|--------|-------------|
| `--fix` | Automatically fix detected issues |

## What It Checks

The doctor command performs the following diagnostic checks:

### Workspace Structure
- **Workspace detection**: Confirms you're in a valid jot workspace
- **inbox.md file**: Verifies the main inbox file exists
- **lib/ directory**: Checks for the organized notes directory
- **.jot/ directory**: Validates the internal data directory

### File Permissions
- **inbox.md writability**: Ensures the inbox file can be written to
- **Directory permissions**: Checks directory access permissions

### External Tools
- **Editor availability**: Looks for common editors (vim, nvim, nano, emacs)
- **Pager availability**: Checks for pagers (less, more)

## Examples

### Basic Diagnostics

```bash
# Run basic health check
jot doctor
```

Output:
```
Running jot workspace diagnostics...

✓ Found workspace at: /home/user/notes
✓ inbox.md exists
✓ lib/ directory exists
✓ .jot/ directory exists
✓ inbox.md is writable
✓ Editor 'vim' is available
✓ Pager 'less' is available

Workspace health: ✓ Excellent
```

### With Issues Found

```bash
jot doctor
```

Output:
```
Running jot workspace diagnostics...

✓ Found workspace at: /home/user/notes
✗ inbox.md is missing
✓ lib/ directory exists
✗ .jot/ directory is missing
✓ Editor 'vim' is available
! No pager found in PATH

Workspace health: ✗ Issues found (2 issues, 1 warning)
Run 'jot doctor --fix' to apply automatic fixes
```

### Auto-Fix Issues

```bash
jot doctor --fix
```

Output:
```
Running jot workspace diagnostics...

✓ Found workspace at: /home/user/notes
✗ inbox.md is missing
✓ lib/ directory exists
✗ .jot/ directory is missing
✓ Editor 'vim' is available
! No pager found in PATH

Applying fixes...
✓ Created inbox.md
✓ Created .jot/ directory

Workspace health: ✓ Good (1 warning)
```

### Outside Workspace

```bash
jot doctor
```

Output:
```
Running jot workspace diagnostics...

✗ Not in a jot workspace
  Run 'jot init' to initialize a workspace

Workspace health: ✗ Critical (1 issues)
```

## JSON Output

```bash
jot doctor --json
```

```json
{
  "operation": "doctor",
  "workspace_found": true,
  "workspace_root": "/home/user/notes",
  "health_status": "excellent",
  "checks": [
    {
      "name": "workspace_detection",
      "status": "passed",
      "message": "Found workspace at: /home/user/notes"
    },
    {
      "name": "inbox_exists",
      "status": "passed",
      "message": "inbox.md exists"
    },
    {
      "name": "lib_exists",
      "status": "passed",
      "message": "lib/ directory exists"
    },
    {
      "name": "jot_dir_exists",
      "status": "passed",
      "message": ".jot/ directory exists"
    },
    {
      "name": "inbox_writable",
      "status": "passed",
      "message": "inbox.md is writable"
    },
    {
      "name": "editor_available",
      "status": "passed",
      "message": "Editor 'vim' is available"
    },
    {
      "name": "pager_available",
      "status": "passed",
      "message": "Pager 'less' is available"
    }
  ],
  "issues": [],
  "warnings": [],
  "fixes_applied": [],
  "summary": {
    "total_checks": 7,
    "passed_checks": 7,
    "failed_checks": 0,
    "issues_found": 0,
    "warnings_found": 0,
    "fixes_applied": 0,
    "overall_health": "excellent"
  },
  "metadata": {
    "command": "doctor",
    "success": true,
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

## Health Status Levels

| Status | Description |
|--------|-------------|
| **Excellent** | No issues or warnings found |
| **Good** | No issues, but warnings present |
| **Issues** | Non-critical issues found |
| **Critical** | Critical issues found (e.g., no workspace) |

## Issue Types and Fixes

### Workspace Structure Issues

| Issue | Severity | Auto-fixable | Fix Action |
|-------|----------|--------------|------------|
| Not in workspace | Critical | No | Run `jot init` |
| Missing inbox.md | High | Yes | Creates inbox with template |
| Missing lib/ directory | High | Yes | Creates lib/ with README |
| Missing .jot/ directory | High | Yes | Creates internal data directory |

### Permission Issues

| Issue | Severity | Auto-fixable | Fix Action |
|-------|----------|--------------|------------|
| inbox.md not writable | Medium | No | Check file permissions |
| Directory not accessible | Medium | No | Check directory permissions |

### External Tool Warnings

| Issue | Severity | Auto-fixable | Fix Action |
|-------|----------|--------------|------------|
| No editor found | Low | No | Install vim, nvim, nano, or emacs |
| No pager found | Low | No | Install less or ensure more is available |

## Automatic Fixes

When using `--fix`, the doctor command automatically applies fixes for:

### Missing inbox.md
Creates a new inbox file with template content:
```markdown
# Inbox

This is your inbox for capturing new notes quickly. Use 'jot capture' to add new notes here.

---
```

### Missing lib/ directory  
Creates the library directory with a README.md explaining organization strategies.

### Missing .jot/ directory
Creates the internal data directory for jot's workspace metadata.

## When to Run Doctor

### Regular Maintenance
- After moving or copying workspace directories
- When experiencing unexpected behavior
- Before important operations

### Error Troubleshooting
- When commands fail unexpectedly
- After manual file system changes
- When setting up new workspaces

### Before Sharing Workspaces
- Ensure workspace integrity
- Fix common issues before collaboration
- Validate external tool dependencies

## Troubleshooting Common Issues

### Workspace Not Found
```bash
jot doctor
# Output: ✗ Not in a jot workspace
```
**Solution**: Run `jot init` to initialize a workspace or navigate to an existing workspace.

### Permission Denied
```bash
jot doctor
# Output: ✗ inbox.md is not writable
```
**Solution**: Check file permissions with `ls -la inbox.md` and fix with `chmod 644 inbox.md`.

### Missing External Tools
```bash
jot doctor
# Output: ! No common editor found in PATH
```
**Solution**: Install a text editor like vim, nvim, nano, or emacs.

## Integration with Other Commands

The doctor command works well with:

- **jot status**: Compare workspace health with status information
- **jot init**: Initialize workspaces after doctor identifies issues
- **jot workspace**: Validate workspace registry integrity

## Automation and Scripting

```bash
# Check workspace health in scripts
if jot doctor --json | jq -e '.health_status == "excellent"' > /dev/null; then
    echo "Workspace is healthy"
else
    echo "Workspace needs attention"
    jot doctor --fix
fi

# Extract specific check results
jot doctor --json | jq '.checks[] | select(.name == "inbox_exists")'
```

## Cross-references

- [jot init](jot-init.md) - Initialize workspace when doctor finds critical issues
- [jot status](jot-status.md) - Check workspace status and statistics
- [jot workspace](jot-workspace.md) - Manage workspace registry

## See Also

- [Global Options](README.md#global-options)
- [Workspace Structure](../topics/workspace.md)
- [Troubleshooting](../user-guide/troubleshooting.md)
- [Configuration](../topics/configuration.md)

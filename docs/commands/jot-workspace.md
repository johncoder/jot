[Documentation](../README.md) > [Commands](README.md) > workspace

# jot workspace

## Description

The workspace command allows you to manage your registered workspaces from anywhere on your filesystem. Each workspace represents a collection of notes and can be accessed globally using the `--workspace` flag.

This command is useful for:
- Managing multiple note collections
- Switching between different projects or contexts
- Setting up workspace discovery and defaults
- Registering workspaces for global access

When run without subcommands, shows the current workspace path for piping to other commands.

## Usage

## Subcommands

### `jot workspace` (default)

Show the current workspace path:

```bash
jot workspace
```

Output:
```
/home/user/notes
```

### `jot workspace list`

List all registered workspaces:

```bash
jot workspace list
```

Output:
```
Registered Workspaces:

* notes           ~/notes                         (default, active) (valid)
  work            ~/work-notes                    (valid)
  archive         ~/old-notes                     (invalid - path missing or not initialized)

* = currently active workspace
Use 'jot workspace default <name>' to change default
```

### `jot workspace add`

Add a workspace to the registry:

```bash
jot workspace add <name> <path>
```

**Arguments:**
- `<name>` - Workspace name for reference
- `<path>` - Path to the workspace directory

**Requirements:**
- Path must exist and be initialized (contain `.jot` directory)
- Name must not already exist in registry
- If this is the first workspace, it becomes the default

### `jot workspace remove`

Remove a workspace from the registry:

```bash
jot workspace remove <name>
```

**Arguments:**
- `<name>` - Name of workspace to remove

**Effects:**
- Removes workspace from global registry
- Does not delete workspace files
- If removing default workspace, automatically selects new default

### `jot workspace default`

Set the default workspace:

```bash
jot workspace default <name>
```

**Arguments:**
- `<name>` - Name of workspace to set as default

The default workspace is used when no local workspace is found during workspace discovery.

## Examples

### Setting Up Workspaces

```bash
# Initialize a new workspace
jot init ~/work-notes

# Add it to the registry
jot workspace add work ~/work-notes

# Set it as default
jot workspace default work

# List all workspaces
jot workspace list
```

### Managing Multiple Workspaces

```bash
# Add personal notes workspace
jot workspace add personal ~/personal-notes

# Add project workspace
jot workspace add project ~/project-docs

# Switch between workspaces using --workspace flag
jot status --workspace personal
jot capture --workspace project --content "Meeting notes"

# Remove unused workspace
jot workspace remove old-project
```

### Using Current Workspace Path

```bash
# Show current workspace path
jot workspace

# Use in scripts
WORKSPACE_PATH=$(jot workspace)
cd "$WORKSPACE_PATH"
```

## JSON Output

All subcommands support JSON output for automation:

### List Workspaces JSON

```bash
jot workspace list --json
```

```json
{
  "workspaces": [
    {
      "name": "notes",
      "path": "/home/user/notes",
      "status": "valid",
      "is_default": true,
      "is_active": true
    },
    {
      "name": "work",
      "path": "/home/user/work-notes",
      "status": "valid",
      "is_default": false,
      "is_active": false
    }
  ],
  "summary": {
    "total_workspaces": 2,
    "valid_workspaces": 2,
    "default_workspace": "notes",
    "active_workspace": "notes"
  },
  "metadata": {
    "command": "workspace list",
    "success": true,
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

### Add Workspace JSON

```bash
jot workspace add project ~/project-docs --json
```

```json
{
  "operations": [
    {
      "operation": "add_workspace",
      "result": "success",
      "details": {
        "workspace_name": "project",
        "workspace_path": "/home/user/project-docs",
        "set_as_default": false,
        "validation_passed": true
      }
    }
  ],
  "metadata": {
    "command": "workspace add",
    "success": true,
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

## Workspace Status Indicators

| Status | Description |
|--------|-------------|
| `valid` | Workspace exists and is properly initialized |
| `invalid` | Path missing or not initialized |
| `active` | Currently active workspace |
| `default` | Default workspace for workspace discovery |

## Error Conditions

| Error | Cause | Solution |
|-------|-------|----------|
| `workspace already exists` | Name already registered | Use different name or remove existing |
| `path does not exist or is not initialized` | Invalid workspace path | Run `jot init` on the path first |
| `workspace not found in registry` | Trying to modify non-existent workspace | Check `jot workspace list` for available names |
| `no workspace found` | No workspace available | Initialize or register a workspace |

## Workspace Discovery

jot uses the following priority order to find workspaces:

1. **--workspace flag** - Explicit workspace override
2. **Current directory** - Look for `.jot` directory in current path
3. **Parent directories** - Walk up directory tree looking for `.jot`
4. **Registry lookup** - Use registered workspaces
5. **Default workspace** - Fall back to configured default

## Cross-references

- [jot init](jot-init.md) - Initialize new workspaces
- [jot status](jot-status.md) - Check workspace status
- [jot doctor](jot-doctor.md) - Diagnose workspace issues

## See Also

- [Global Options](README.md#global-options)
- [Workspace Structure](../topics/workspace.md)
- [Configuration](../topics/configuration.md)
- [Multiple Workspaces](../topics/multiple-workspaces.md)

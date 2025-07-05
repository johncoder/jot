[Documentation](../README.md) > [Commands](README.md) > archive

# jot archive

## Description

The `jot archive` command is a smart alias for `jot refile` that automatically uses the workspace's configured archive location as the destination. It provides a convenient way to move older or completed notes to a dedicated archive location.

This command is useful for:
- Moving completed projects to archives
- Cleaning up old meeting notes
- Organizing historical content
- Maintaining workspace hygiene

## Usage

## Arguments

| Argument | Description |
|----------|-------------|
| `[SOURCE]` | Optional source selector to archive (if omitted, sets up archive structure) |

## Options

| Option | Description |
|--------|-------------|
| `--config` | Show current archive configuration |
| `--set-location` | Set archive location path |
| `--no-verify` | Skip hooks verification |

## Operation Modes

### 1. Initialize Archive Structure

When run without arguments, creates the archive directory and file structure:

```bash
jot archive
```

### 2. Archive Content

Archive specific content using selectors:

```bash
jot archive "inbox.md#old-project"
```

### 3. Configuration Management

View or update archive configuration:

```bash
jot archive --config
jot archive --set-location "archive/2025.md#Archived"
```

## Examples

### Setting Up Archives

```bash
# Initialize archive structure
jot archive

# Check current configuration
jot archive --config

# Set custom archive location
jot archive --set-location "archive/completed.md#Archive"
```

### Archiving Content

```bash
# Archive specific project notes
jot archive "inbox.md#old-project"

# Archive meeting notes
jot archive "work.md#meetings/Q1-2024"

# Archive completed tasks
jot archive "tasks.md#completed"
```

### Archive Configuration

```bash
# Show current archive settings
jot archive --config

# Set archive to specific file and section
jot archive --set-location "archive/2025.md#Archived Items"

# Set archive to simple file
jot archive --set-location "archive.md"
```

## Default Archive Location

The default archive location is `archive/archive.md#Archive`. This can be configured per workspace using the `--set-location` option.

## Archive Structure

When initializing archives, the command creates:

1. **Archive directory**: Creates `archive/` directory if it doesn't exist
2. **Archive file**: Creates the target file with proper heading structure
3. **Configuration**: Updates workspace configuration with archive location

## Hook Integration

The archive command integrates with the hook system:

- **Pre-archive hooks**: Execute before archiving (can abort operation)
- **Post-archive hooks**: Execute after successful archiving (informational)
- **Skip hooks**: Use `--no-verify` to bypass hook execution

## Output Examples

### Initialize Archive Structure

```bash
jot archive
```

Output:
```
Archive structure ready!
Archive location: archive/archive.md#Archive
Full path: /workspace/archive/archive.md

Use 'jot archive "source.md#section"' to archive specific content.
```

### Archive Content

```bash
jot archive "inbox.md#old-project"
```

Output:
```
Archiving 'inbox.md#old-project' to 'archive/archive.md#Archive'...
✓ Successfully archived content
```

### Show Configuration

```bash
jot archive --config
```

Output:
```
Archive Configuration:
  Location: archive/archive.md#Archive
  Resolved: archive/archive.md#Archive
  Full path: /workspace/archive/archive.md
```

## JSON Output

### Configuration JSON

```bash
jot archive --config --json
```

```json
{
  "operation": "show_config",
  "archive_location": "archive/archive.md#Archive",
  "resolved_path": "archive/archive.md#Archive",
  "metadata": {
    "command": "archive",
    "success": true,
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

### Initialize Structure JSON

```bash
jot archive --json
```

```json
{
  "operation": "initialize",
  "archive_dir": "/workspace/archive",
  "created_items": [
    {
      "path": "archive/",
      "type": "directory",
      "description": "Archive directory",
      "created": true
    },
    {
      "path": "archive/archive.md",
      "type": "file",
      "description": "Archive file",
      "created": true,
      "size": 28
    }
  ],
  "operations": [
    "Created archive directory: archive",
    "Created archive file: archive/archive.md"
  ],
  "summary": {
    "total_items": 2,
    "items_created": 2,
    "items_existing": 0,
    "directory_ready": true
  },
  "metadata": {
    "command": "archive",
    "success": true,
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

## Archive Location Format

Archive locations use the same format as refile destinations:

```
file.md#section/subsection
```

Examples:
- `archive.md` - Root of archive file
- `archive/2025.md#Archived` - Specific section in yearly archive
- `completed/tasks.md#Done` - Organized by category

## Workflow Integration

### Typical Archive Workflow

1. **Setup**: `jot archive` to create structure
2. **Review**: Use `jot find` to identify content to archive
3. **Archive**: `jot archive "source#selector"` to move content
4. **Verify**: Use `jot peek` to check archived content

### Automated Archiving

```bash
# Archive all completed tasks
jot find "completed" --json | jq -r '.results[].relative_path' | \
  xargs -I {} jot archive "{}"

# Archive old meeting notes
jot archive "meetings.md#2024" --no-verify
```

## Error Conditions

| Error | Cause | Solution |
|-------|-------|----------|
| `archive location not configured` | No archive location set | Run `jot archive --set-location` |
| `archive file not found` | Archive file doesn't exist | Run `jot archive` to initialize |
| `pre-archive hook aborted` | Hook script prevented archiving | Check hook output and fix issues |
| `source not found` | Source selector doesn't exist | Check source selector syntax |
| `permission denied` | Cannot write to archive location | Check file permissions |

## Archive Maintenance

### Organizing Archives

```bash
# Set yearly archive
jot archive --set-location "archive/2025.md#Archive"

# Set topic-based archive
jot archive --set-location "archive/projects.md#Completed"

# Set simple flat archive
jot archive --set-location "archive.md"
```

### Archive Cleanup

```bash
# Check archive contents
jot peek "archive/archive.md"

# Find archived items
jot find "archived" --archive
```

## Cross-references

- [jot refile](jot-refile.md) - Understanding refile functionality
- [jot find](jot-find.md) - Finding content to archive
- [jot peek](jot-peek.md) - Reviewing archived content
- [jot hooks](jot-hooks.md) - Configuring archive hooks

## See Also

- [Global Options](README.md#global-options)
- [Configuration Guide](../user-guide/configuration.md) - Archive directory defaults and hook configuration

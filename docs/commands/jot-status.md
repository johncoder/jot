[Documentation](../README.md) > [Commands](README.md) > status

# jot status

## Description

The `jot status` command provides an overview of your current jot workspace. It shows workspace details, file statistics, and health indicators to help you understand the current state of your notes.

This command is useful for:
- Getting a quick overview of your workspace
- Checking workspace health and configuration
- Understanding file distribution across your workspace
- Monitoring recent activity

## Usage

```bash
jot status [options]
```

## Options

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--json` | | Output in JSON format for automation | false |

*See [Global Options](README.md#global-options) for `--config`, `--workspace`, and `--help` flags.*

## Examples

### Basic workspace status

```bash
jot status
```

Output:

```
Workspace: my-notes (/home/user/notes)
Discovery: workspace registry

Files:
  Inbox: 3 items (245 words)
  Library: 12 files (1,847 words)
  Templates: 2 approved, 0 pending
  Archive: 8 files

Recent Activity:
  Last capture: 2 hours ago
  Last refile: 1 day ago
```

### JSON output for automation

```bash
jot status --json
```

Output:

```json
{
  "operation": "status",
  "workspace": {
    "name": "my-notes",
    "path": "/home/user/notes",
    "discovery_method": "workspace registry"
  },
  "files": {
    "inbox_items": 3,
    "inbox_words": 245,
    "library_files": 12,
    "library_words": 1847,
    "templates_approved": 2,
    "templates_pending": 0,
    "archive_files": 8
  },
  "recent_activity": {
    "last_capture": "2025-07-04T14:30:00Z",
    "last_refile": "2025-07-03T09:15:00Z"
  },
  "metadata": {
    "timestamp": "2025-07-04T16:30:00Z",
    "success": true
  }
}
```

### Using specific workspace

```bash
jot status --workspace work-notes
```

Shows status for the "work-notes" workspace instead of auto-detecting the current workspace.

## What Happens

When you run `jot status`, the command:

1. **Discovers the workspace** using the current directory, workspace registry, or the `--workspace` override
2. **Counts files and content** in `inbox.md`, `lib/` directory, templates, and archive
3. **Checks recent activity** by examining file modification times
4. **Validates workspace health** including configuration and file structure
5. **Displays the summary** in human-readable or JSON format

## Workspace Information

The status output includes:

**Workspace Details**
- Name and full path of the current workspace
- Discovery method (current directory, workspace registry, or flag override)

**File Statistics**
- Inbox items and word count from parsing `inbox.md`
- Library file count and total word count from `lib/` directory
- Template counts with approval status
- Archive file count

**Recent Activity**
- Timestamp of most recent capture operation
- Timestamp of most recent refile operation
- Other recent workspace modifications

## Error Conditions

| Error | Cause | Solution |
|-------|-------|----------|
| "No workspace found" | Not in a jot workspace | Run [jot init](jot-init.md) or use `--workspace` |
| "Permission denied" | Cannot read workspace files | Check file permissions |
| "Workspace corrupted" | Missing required directories | Run [jot doctor](jot-doctor.md) |

## See Also

- [jot init](jot-init.md) - Initialize a new workspace
- [jot doctor](jot-doctor.md) - Diagnose and repair workspace issues
- [jot workspace](jot-workspace.md) - Manage workspace registry
- [jot find](jot-find.md) - Search workspace content

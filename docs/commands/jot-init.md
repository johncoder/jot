[Documentation](../README.md) > [Commands](README.md) > init

# jot init

## Description

The `jot init` command creates a new jot workspace in the specified directory (or current directory if none provided). It sets up the essential structure including `inbox.md` for note capture, `lib/` for organized notes, and `.jot/` for internal data.

This command is useful for:
- Setting up your first jot workspace
- Creating project-specific note repositories
- Establishing organized note-taking structure
- Preparing directories for team collaboration

## Usage

```bash
jot init [path] [options]
```

## Arguments

| Argument | Description | Default |
|----------|-------------|---------|
| `path` | Directory to initialize as workspace | current directory |

## Options

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--json` | | Output in JSON format | false |

*See [Global Options](README.md#global-options) for additional flags.*

## Examples

### Initialize current directory

```bash
jot init
```

Output:

```
Initialized jot workspace in /home/user/notes
Created:
  inbox.md
  lib/
  .jot/
  .jot/config/
  .jot/templates/
  .jot/archive/
```

### Initialize specific directory

```bash
jot init my-project-notes
```

Creates and initializes the `my-project-notes` directory.

### JSON output for automation

```bash
jot init --json
```

Output:

```json
{
  "operation": "init",
  "workspace_path": "/home/user/notes",
  "created_files": [
    "inbox.md",
    "lib/",
    ".jot/",
    ".jot/config/",
    ".jot/templates/",
    ".jot/archive/"
  ],
  "metadata": {
    "timestamp": "2025-07-04T16:30:00Z",
    "success": true
  }
}
```

## What Happens

When you run `jot init`, the command:

1. **Creates the target directory** if it doesn't exist
2. **Sets up workspace structure** with required directories
3. **Creates initial files** including empty `inbox.md`
4. **Initializes configuration** in `.jot/config/`
5. **Prepares template storage** in `.jot/templates/`
6. **Sets up archive directory** in `.jot/archive/`

## Workspace Structure

The initialized workspace contains:

```
workspace/
├── inbox.md              # Capture destination for new notes
├── lib/                  # Organized notes directory
└── .jot/                 # Internal jot data
    ├── config/           # Workspace-specific configuration  
    ├── templates/        # Note templates
    └── archive/          # Archived notes
```

**Key Files and Directories:**

- **`inbox.md`**: Default capture location for new notes
- **`lib/`**: Directory for organized, refiled notes
- **`.jot/config/`**: Workspace configuration files
- **`.jot/templates/`**: Custom templates for structured note capture
- **`.jot/archive/`**: Storage for archived notes

## Next Steps

After initialization, you typically want to:

1. **Register the workspace** with [jot workspace add](jot-workspace.md#add)
2. **Create your first note** with [jot capture](jot-capture.md)
3. **Set up templates** with [jot template new](jot-template.md#new)
4. **Check workspace status** with [jot status](jot-status.md)

## Error Conditions

| Error | Cause | Solution |
|-------|-------|----------|
| "Directory already initialized" | `.jot/` directory exists | Use existing workspace or choose different path |
| "Permission denied" | Cannot create directories | Check permissions for target directory |
| "Directory not empty" | Target contains files | Use `--force` or choose empty directory |
| "Invalid path" | Path contains invalid characters | Use valid directory path |

## See Also

- [jot workspace add](jot-workspace.md#add) - Register workspace in global registry
- [jot status](jot-status.md) - Check workspace information
- [jot capture](jot-capture.md) - Start capturing notes
- [Configuration Guide](../user-guide/configuration.md) - Workspace defaults and directory structure

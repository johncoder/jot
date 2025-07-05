[Documentation](../README.md) > [Commands](README.md) > files

# jot files

## Description

The `jot files` command lists and browses files in the current jot workspace. It provides both simple listing and interactive browsing modes with FZF integration for enhanced file navigation.

This command is useful for:
- Getting an overview of all markdown files in your workspace
- Interactive file browsing and selection
- Opening files in your editor from the command line
- Integrating file selection with other command-line tools

## Usage

```bash
jot files [options]
```

## Options

| Option | Short | Description |
|--------|-------|-------------|
| `--interactive` | `-i` | Interactive file browser (requires `JOT_FZF=1`) |
| `--edit` | | Open selected file in editor (use with `--interactive`) |
| `--select` | `-s` | Output selected file path for composition (use with `--interactive`) |

## Operation Modes

### 1. Simple Listing

List all markdown files in the workspace:

```bash
jot files
```

### 2. Interactive Browser

Browse files interactively with FZF (requires `JOT_FZF=1`):

```bash
export JOT_FZF=1
jot files --interactive
```

### 3. Interactive Editor

Browse and open files in your editor:

```bash
jot files --interactive --edit
```

### 4. Selection Mode

Select a file path for use with other tools:

```bash
jot files --interactive --select
```

## Examples

### Basic File Listing

```bash
jot files
```

Output:
```
inbox.md
lib/projects.md
lib/meetings.md
lib/research/databases.md
lib/research/algorithms.md
archive/2024-q1.md
archive/2024-q2.md
```

### Interactive File Browser

```bash
export JOT_FZF=1
jot files --interactive
```

This opens an FZF interface showing:
```
> inbox.md                          📄 inbox.md
  lib/projects.md                   📄 projects.md
  lib/meetings.md                   📄 meetings.md
  lib/research/databases.md         📄 databases.md
  lib/research/algorithms.md        📄 algorithms.md
  archive/2024-q1.md               📄 2024-q1.md
  archive/2024-q2.md               📄 2024-q2.md
7/7
> 
```

### Interactive Editor Mode

```bash
jot files --interactive --edit
```

Provides the same interface, but pressing ENTER opens the selected file in your configured editor (`$EDITOR` or fallback).

### Selection for Command Composition

```bash
# View content of selected file
cat $(jot files --interactive --select)

# Edit selected file with specific editor
vim $(jot files --interactive --select)

# Peek at selected file
jot peek $(jot files --interactive --select)

# Search in selected file
jot find "keyword" $(jot files --interactive --select)
```

### Integration with Other Commands

```bash
# Browse files and then peek at selection
FILE=$(jot files -i -s) && jot peek "$FILE"

# Browse files and refile content from selection
FILE=$(jot files -i -s) && jot refile "$FILE#section" --to "archive.md#old"

# Browse files and capture template to selection
FILE=$(jot files -i -s) && jot capture --template meeting | jot refile - --to "$FILE#meetings"
```

## File Discovery

The command recursively finds all `.md` files in the workspace, excluding:
- Files in the `.jot/` directory (internal data)
- Hidden files and directories (starting with `.`)

### Included Locations
- `inbox.md` (workspace root)
- `lib/` directory and subdirectories
- `archive/` directory and subdirectories
- Any other `.md` files in the workspace

## JSON Output

```bash
jot files --json
```

```json
{
  "total_files": 7,
  "files": [
    {
      "path": "/workspace/inbox.md",
      "relative_path": "inbox.md",
      "name": "inbox.md",
      "size": 1024,
      "modified": "2024-01-15T14:30:00Z"
    },
    {
      "path": "/workspace/lib/projects.md",
      "relative_path": "lib/projects.md", 
      "name": "projects.md",
      "size": 2048,
      "modified": "2024-01-15T10:15:00Z"
    },
    {
      "path": "/workspace/lib/meetings.md",
      "relative_path": "lib/meetings.md",
      "name": "meetings.md",
      "size": 1536,
      "modified": "2024-01-14T16:45:00Z"
    }
  ],
  "workspace": {
    "root": "/workspace",
    "name": "workspace"
  },
  "metadata": {
    "command": "files",
    "success": true,
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

## Interactive Mode Features

### FZF Integration

When `JOT_FZF=1` is set and `--interactive` is used:

- **Live search**: Type to filter files in real-time
- **File preview**: Preview file contents (when supported by FZF configuration)
- **Keyboard navigation**: Use arrow keys or vim-style navigation
- **Multiple selection modes**: Different actions based on flags

### Key Bindings

| Key | Action |
|-----|--------|
| `Enter` | Select file (behavior depends on mode) |
| `Escape` / `Ctrl-C` | Cancel selection |
| `↑/↓` or `k/j` | Navigate file list |
| `Tab` | Toggle selection (if applicable) |

### Preview Integration

FZF can show file previews when configured. Add to your shell profile:

```bash
export FZF_DEFAULT_OPTS="--preview 'cat {}' --preview-window=right:50%"
```

## Editor Integration

### Editor Selection

The command respects standard editor environment variables:

1. `$EDITOR` - Primary editor preference
2. `$VISUAL` - Visual editor preference  
3. Fallback to common editors: `vim`, `nvim`, `nano`, `emacs`

### Example Editor Configurations

```bash
# Use VS Code
export EDITOR="code"

# Use Neovim
export EDITOR="nvim"

# Use nano for simplicity
export EDITOR="nano"
```

## Error Conditions

| Error | Cause | Solution |
|-------|-------|----------|
| `no markdown files found` | Workspace has no `.md` files | Create some markdown files or check workspace |
| `interactive mode not available` | FZF not configured with JSON output | Remove `--json` or use standard mode |
| `--select requires --interactive` | Selection mode without interactive | Add `--interactive` flag |
| `FZF not found` | Interactive mode requires FZF | Install `fzf` or use standard mode |
| `no editor configured` | Editor mode but no editor available | Set `$EDITOR` or install an editor |

## Performance Considerations

- **Large workspaces**: File discovery may be slow with many files
- **Deep directory structures**: Recursive search can take time
- **FZF startup**: Interactive mode has initial overhead

## Use Cases

### Daily Workflow

```bash
# Quick overview of workspace files
jot files

# Browse and open file for editing
jot files -i --edit

# Find specific file interactively
jot files -i | grep "meeting"
```

### Automation and Scripting

```bash
#!/bin/bash
# Script to process all markdown files
jot files --json | jq -r '.files[].path' | while read file; do
    echo "Processing: $file"
    # Process each file
done

# Interactive file selection in scripts
selected_file=$(jot files --interactive --select)
if [ -n "$selected_file" ]; then
    echo "Selected: $selected_file"
    # Do something with selected file
fi
```

### Integration with Other Tools

```bash
# Search in specific file
FILE=$(jot files -i -s) && grep -n "TODO" "$FILE"

# Copy file path to clipboard
jot files -i -s | pbcopy  # macOS
jot files -i -s | xclip   # Linux

# Open file in external application
jot files -i -s | xargs typora  # Open in Typora editor
```

## Cross-references

- [jot peek](jot-peek.md) - Preview file contents
- [jot find](jot-find.md) - Search within files
- [jot refile](jot-refile.md) - Move content between files
- [jot status](jot-status.md) - Workspace file statistics

## See Also

- [Global Options](README.md#global-options)
- [Configuration Guide](../user-guide/configuration.md) - Editor integration and pager configuration

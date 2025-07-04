[Documentation](../README.md) > [Commands](README.md) > refile

# jot refile

## Description

The `jot refile` command moves entire markdown subtrees (headings with all nested content) between files. It supports path-based selectors with case-insensitive contains matching and provides multiple operational modes for different workflows.

This command is useful for:
- Moving content from inbox to organized locations
- Reorganizing existing notes between files
- Interactive content management with FZF
- Batch operations with automation

## Usage

## Options

| Option | Short | Description |
|--------|-------|-------------|
| `--to` | | Destination path (e.g., `work.md#projects/frontend`) |
| `--prepend` | | Insert content at the beginning under target heading |
| `--verbose` | `-v` | Show detailed information about the refile operation |
| `--interactive` | `-i` | Interactive mode using FZF (requires `JOT_FZF=1`) |
| `--no-verify` | | Skip hooks verification |

## Path-based Selector Syntax

- **File specification**: `filename.md` or absolute path
- **Heading specification**: `#heading/subheading/deep`
- **Case-insensitive matching**: Each segment uses contains matching
- **Exact matching**: Must match exactly one subtree
- **Leading slashes**: Handle unusual document structures (`#/foo/bar` skips level 1)

## Operation Modes

### 1. Standard Refile

Move a subtree from source to destination:

```bash
jot refile "inbox.md#meeting notes" --to "work.md#projects"
```

### 2. Interactive Mode

Use FZF for interactive selection (requires `JOT_FZF=1`):

```bash
jot refile --interactive
jot refile "inbox.md" --interactive  # Pre-select source file
```

### 3. Destination Inspection

Inspect destination structure without source:

```bash
jot refile --to "work.md#projects/frontend"
```

### 4. File Structure Display

Show available selectors for a file:

```bash
jot refile "work.md"
```

## Examples

### Basic Refile Operations

```bash
# Move meeting notes from inbox to work section
jot refile "inbox.md#meeting" --to "work.md#projects"

# Move research notes to archive with deep path
jot refile "notes.md#research/database" --to "archive.md#technical"

# Handle unusual document structures
jot refile "inbox.md#/foo/bar" --to "work.md#tasks"
```

### Content Positioning

```bash
# Insert content at beginning of target section
jot refile "inbox.md#urgent" --to "work.md#tasks" --prepend

# Default behavior appends to end of target section
jot refile "inbox.md#notes" --to "work.md#projects"
```

### Interactive Workflow

```bash
# Full interactive mode - select both source and destination
jot refile --interactive

# Interactive with pre-selected source file
jot refile "inbox.md" --interactive

# Interactive with pre-selected source subtree
jot refile "inbox.md#meeting" --interactive
```

### Inspection and Planning

```bash
# Show all available selectors in a file
jot refile "work.md"

# Inspect destination structure
jot refile --to "work.md#projects/frontend"

# Verbose output for detailed operation info
jot refile "inbox.md#notes" --to "work.md#projects" --verbose
```

## Same-File Operations

When source and destination are in the same file, `jot refile` uses safe text manipulation to avoid conflicts:

```bash
# Move subtree within same file
jot refile "work.md#old-section" --to "work.md#new-section"
```

## Hook Integration

The refile command integrates with the hook system for automation:

- **Pre-refile hooks**: Execute before the operation (can abort)
- **Post-refile hooks**: Execute after successful operation (informational)
- **Skip hooks**: Use `--no-verify` to bypass hook execution

## Error Conditions

| Error | Cause | Solution |
|-------|-------|----------|
| `source file not found` | Source file doesn't exist | Check file path and spelling |
| `destination file not found` | Destination file doesn't exist | Create destination file first |
| `subtree not found` | Selector doesn't match any heading | Check selector syntax and case |
| `multiple subtrees match` | Selector matches multiple headings | Use more specific selector |
| `pre-refile hook aborted` | Hook script prevented operation | Check hook output, fix issues |
| `permission denied` | File access restrictions | Check file permissions |

## Interactive Mode Requirements

Interactive mode requires:
- `JOT_FZF=1` environment variable
- `fzf` command available in PATH
- Terminal with interactive capabilities

## Level Transformation

When moving subtrees between different heading levels, content is automatically transformed:

- **Level adjustment**: Headings are adjusted to match destination level
- **Consistent formatting**: Maintains proper markdown structure
- **Content preservation**: All nested content is preserved

## Cross-references

- [jot capture](jot-capture.md) - Capturing new content
- [jot find](jot-find.md) - Finding content to refile
- [jot peek](jot-peek.md) - Previewing content before refile
- [jot status](jot-status.md) - Checking workspace state
- [Hooks](jot-hooks.md) - Configuring refile hooks

## See Also

- [Global Options](README.md#global-options)
- [Path-based Selectors](../topics/selectors.md)
- [Interactive Workflows](../topics/interactive.md)
- [Hook System](../topics/hooks.md)

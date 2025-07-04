[Documentation](../README.md) > [Commands](README.md) > peek

# jot peek

## Description

The `jot peek` command allows you to view markdown content without opening it in an editor. It supports two main modes:

1. **Whole file**: Display entire file content
2. **Subtree**: Display specific section using path-based selectors

This command is useful for:
- Quickly reviewing files or sections without opening an editor
- Previewing content before refile operations
- Generating table of contents for navigation
- Inspecting file structure and selectors

## Usage

1. **Whole file**: Display entire file content
2. **Subtree**: Display specific section using path-based selectors

This is useful for quickly reviewing files or specific sections without opening them in an editor.

## Arguments

| Argument | Description |
|----------|-------------|
| `SELECTOR` | File path or path-based selector (e.g., `inbox.md` or `work.md#projects`) |

## Options

| Option | Short | Description |
|--------|-------|-------------|
| `--raw` | `-r` | Output raw content without formatting |
| `--info` | `-i` | Show subtree metadata information |
| `--toc` | `-t` | Show table of contents for file or subtree |
| `--short` | `-s` | Generate shortest possible selectors (use with `--toc`) |
| `--no-workspace` | | Resolve file paths relative to current directory instead of workspace |

## Selector Syntax

### Whole File
```bash
jot peek "filename.md"
```

### Subtree Selection
```bash
jot peek "file.md#path/to/heading"
```

**Path-based selector features:**
- **Case-insensitive matching**: Each segment uses contains matching
- **Exact matching**: Must match exactly one subtree
- **Leading slashes**: Handle unusual document structures (`#/foo/bar` skips level 1)

### Enhanced Selectors
```bash
jot peek "file:42"           # Line number to heading conversion
jot peek "file:42#heading"   # Line number with heading context
```

## Examples

### Basic File Viewing

```bash
# View entire inbox file
jot peek "inbox.md"

# View specific file with info
jot peek "work.md" --info

# View file in raw format
jot peek "notes.md" --raw
```

### Subtree Viewing

```bash
# View meeting notes subtree
jot peek "inbox.md#meeting"

# View frontend project section
jot peek "work.md#projects/frontend"

# View database research notes
jot peek "notes.md#research/database"

# Handle unusual document structures
jot peek "inbox.md#/foo/bar"
```

### Table of Contents

```bash
# Show TOC for entire file
jot peek "inbox.md" --toc

# Show TOC for specific subtree
jot peek "work.md#projects" --toc

# Show TOC with shortest selectors
jot peek "work.md" --toc --short
```

### Enhanced Selectors

```bash
# Convert line number to heading
jot peek "work.md:42"

# Use enhanced selector with context
jot peek "work.md:42#projects"
```

## Output Modes

### Standard Output

```bash
jot peek "inbox.md#meeting"
```

Output:
```
## Meeting Notes

### Sprint Planning - July 4, 2024

- Review current sprint progress
- Plan next sprint items
- Discuss blockers and dependencies

#### Action Items
- [ ] Update documentation
- [ ] Code review for PR #123
```

### With Information

```bash
jot peek "inbox.md#meeting" --info
```

Output:
```
Subtree Information:
  File: inbox.md
  Heading: meeting
  Level: 2
  Content size: 245 bytes
  Child headings: 2

## Meeting Notes

### Sprint Planning - July 4, 2024

- Review current sprint progress
- Plan next sprint items
- Discuss blockers and dependencies

#### Action Items
- [ ] Update documentation
- [ ] Code review for PR #123
```

### Table of Contents

```bash
jot peek "work.md" --toc
```

Output:
```
Table of Contents for work.md:

work.md#projects
work.md#projects/frontend
work.md#projects/frontend/components
work.md#projects/backend
work.md#projects/backend/api
work.md#meetings
work.md#meetings/standup
work.md#meetings/planning
work.md#tasks
```

### Raw Output

```bash
jot peek "inbox.md#meeting" --raw
```

Output:
```
## Meeting Notes

### Sprint Planning - July 4, 2024

- Review current sprint progress
- Plan next sprint items
- Discuss blockers and dependencies

#### Action Items
- [ ] Update documentation
- [ ] Code review for PR #123
```

## JSON Output

```bash
jot peek "inbox.md#meeting" --json
```

```json
{
  "selector": "inbox.md#meeting",
  "file_path": "/workspace/inbox.md",
  "subtree": {
    "heading": "Meeting Notes",
    "level": 2,
    "content_size": 245,
    "child_headings": 2,
    "content": "## Meeting Notes\n\n### Sprint Planning..."
  },
  "metadata": {
    "command": "peek",
    "success": true,
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

## Table of Contents JSON

```bash
jot peek "work.md" --toc --json
```

```json
{
  "file_path": "/workspace/work.md",
  "table_of_contents": [
    {
      "selector": "work.md#projects",
      "heading": "Projects",
      "level": 1,
      "short_selector": "work.md#proj"
    },
    {
      "selector": "work.md#projects/frontend",
      "heading": "Frontend Development",
      "level": 2,
      "short_selector": "work.md#proj/front"
    }
  ],
  "metadata": {
    "command": "peek",
    "success": true,
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

## File Path Resolution

By default, jot peek resolves files relative to the workspace:

- `inbox.md` → workspace root
- `work.md` → workspace root  
- `lib/notes.md` → workspace lib directory

Use `--no-workspace` to resolve files relative to current directory instead.

## Enhanced Selector Features

### Line Number Conversion

When using line numbers, peek automatically converts them to the nearest heading:

```bash
# Line 42 becomes nearest heading path
jot peek "work.md:42"
# Might become: jot peek "work.md#projects/frontend"
```

### Smart Matching

The selector matching algorithm:
1. Splits selector path by `/`
2. Matches each segment case-insensitively
3. Uses contains matching for flexibility
4. Ensures exactly one match

## Error Conditions

| Error | Cause | Solution |
|-------|-------|----------|
| `file not found` | File doesn't exist | Check file path and spelling |
| `invalid selector` | Malformed selector syntax | Check selector format |
| `subtree not found` | Selector doesn't match any heading | Use more specific or different selector |
| `multiple subtrees match` | Selector matches multiple headings | Use more specific selector |
| `no workspace found` | Not in workspace and no --no-workspace | Use --no-workspace or run from workspace |

## Use Cases

### Quick Review

```bash
# Quick review of inbox content
jot peek "inbox.md"

# Check specific meeting notes
jot peek "inbox.md#meeting"
```

### Content Planning

```bash
# See structure before refiling
jot peek "work.md" --toc

# Check target location
jot peek "work.md#projects" --toc
```

### Debugging Selectors

```bash
# Test selector matches
jot peek "work.md#proj" --info

# Find shortest selectors
jot peek "work.md" --toc --short
```

## Cross-references

- [jot refile](jot-refile.md) - Moving content after peeking
- [jot find](jot-find.md) - Finding content to peek
- [jot capture](jot-capture.md) - Adding content to viewed locations

## See Also

- [Global Options](README.md#global-options)
- [Path-based Selectors](../topics/selectors.md)
- [Workspace Structure](../topics/workspace.md)
- [Content Navigation](../topics/navigation.md)

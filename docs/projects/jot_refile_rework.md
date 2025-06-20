# Jot Refile Rework: Org-mode Inspired Subtree Management

## Executive Summary

This proposal outlines a complete rework of the `jot refile` command to support flexible subtree-based refiling inspired by Org-mode. The new system will enable users to move entire markdown subtrees (headings with all nested content) between files using an intuitive path-based selector syntax.

**Key Changes:**

- **No backward compatibility** - complete redesign
- **Subtree-based refiling** - move headings with all nested content
- **Path-based selectors** - intuitive `#path/to/heading` syntax with contains matching
- **Auto-creation of destination paths** - missing heading hierarchies created automatically
- **Level transformation** - source headings adjusted to fit destination hierarchy
- **Flexible positioning** - append (default) or prepend content under target headings

## Design Philosophy

Following jot's git-inspired CLI approach and Org-mode's powerful refile capabilities:

1. **Predictable**: Clear syntax with obvious behavior
2. **Flexible**: Support various workflows without complexity
3. **Safe**: Exact matching prevents accidental operations
4. **Extensible**: Foundation for advanced features (interactive mode, bulk operations)
5. **Fast**: Optimized for tech workers' rapid note management workflows

## Core Functionality

### Subtree Definition

A **subtree** consists of:

- A heading at any level (`#`, `##`, `###`, etc.)
- All content until the next heading at the same or higher level
- All nested subheadings and their content

Example subtree:

```markdown
## Meeting Notes ← Target heading

Project kickoff discussion ← Content

### Attendees ← Nested heading

- John Smith ← Nested content
- Jane Doe

### Action Items ← Another nested heading

- [ ] Review requirements ← More nested content
- [ ] Schedule followup
```

### Selector Syntax

**Source Selection**: `file.md#path/to/heading`

- Each segment uses "contains" matching (case-insensitive)
- Must match exactly one subtree
- Path navigates down the heading hierarchy

**Examples:**

```bash
# Exact path matching
jot refile "inbox.md#Meeting Notes" --to "work.md#Projects/Frontend"

# Contains matching - finds "Daily Standup Meeting" under "Work Notes"
jot refile "inbox.md#work/daily" --to "work.md#Meetings"

# Cross-file refiling
jot refile "notes.md#research/database" --to "projects.md#Archive/Technical"
```

**Matching Rules:**

- `#meet` matches "Meeting Notes", "Team Meeting", "Standup Meeting"
- `#meet/attend` matches "Attendees" under any heading containing "meet"
- Must result in exactly one match (error if 0 or multiple matches)

**Handling Unusual Document Structures:**

Documents with missing top-level headings can be navigated using leading slashes to indicate the expected missing levels:

- `inbox.md#/foo/bar` - Skip level 1 (`#`), match level 2 heading containing "foo", then level 3 containing "bar"

  ```markdown
  ## foo ← Matches this (level 2, expecting level 1 to be missing)

  ### bar ← Then matches this
  ```

- `inbox.md#//foo/bar` - Skip levels 1-2 (`#`, `##`), match level 3 heading containing "foo", then level 4 containing "bar"

  ```markdown
  ### foo ← Matches this (level 3, expecting levels 1-2 to be missing)

  #### bar ← Then matches this
  ```

- Leading slashes allow navigation in documents that don't start at level 1 or have gaps in heading hierarchy

### Destination Handling

**Auto-creation**: Missing heading paths are automatically created

- `work.md#Projects/Frontend/Tasks` creates missing "Projects", "Frontend", "Tasks" headings
- Proper nesting levels maintained
- Empty headings created as placeholders

**Level Transformation**: Source headings adjusted to fit destination hierarchy

- Source heading level becomes destination level + 1
- All nested headings shift accordingly
- Preserves relative hierarchy within the subtree

**Example Transformation:**

```bash
jot refile "inbox.md#meeting" --to "work.md#events"
```

Source (`inbox.md`):

```markdown
# Meeting Notes ← Level 1

Project discussion

## Attendees ← Level 2

- John, Jane

## Action Items ← Level 2

- Review specs
```

Destination (`work.md`) before:

```markdown
# Events ← Level 1
```

Destination (`work.md`) after:

```markdown
# Events ← Level 1

## Meeting Notes ← Becomes Level 2 (dest level + 1)

Project discussion

### Attendees ← Becomes Level 3

- John, Jane

### Action Items ← Becomes Level 3

- Review specs
```

### Positioning Options

**Default (Append)**: Add content at the end under target heading

```markdown
# Target Heading

Existing content

## New Subtree ← Appended here
```

**Prepend**: Add content at the beginning under target heading

```bash
jot refile "inbox.md#urgent" --to "work.md#tasks" --prepend
```

```markdown
# Target Heading

## New Subtree ← Prepended here

Urgent content

Existing content
```

## Command Interface

### Basic Syntax

```bash
jot refile SOURCE --to DESTINATION [--prepend]
jot refile --to DESTINATION [--prepend]  # Source-less mode for destination inspection
```

### Source-less Refile Behavior

When no source is specified, `jot refile --to DESTINATION` provides information about the destination path:

```bash
# Check if destination path exists
$ jot refile --to "work.md#projects/frontend"
Destination analysis for "work.md#projects/frontend":
✓ File exists: work.md
✓ Path exists: Projects > Frontend
Ready to receive content at level 3

# Check if destination needs creation
$ jot refile --to "work.md#projects/backend/api"
Destination analysis for "work.md#projects/backend/api":
✓ File exists: work.md
✓ Partial path exists: Projects
✗ Missing path: Backend > API
Would create: ## Backend (level 2), ### API (level 3)
Ready to receive content at level 4

# Ambiguous destination path
$ jot refile --to "work.md#proj"
Destination analysis for "work.md#proj":
✓ File exists: work.md
✗ Ambiguous path: "proj" matches multiple headings:
  - "Projects" at line 15
  - "Project Alpha" at line 45
  - "Project Beta" at line 67
Use more specific path like "#projects", "#project alpha", or "#project beta"

# File doesn't exist
$ jot refile --to "nonexistent.md#tasks"
Destination analysis for "nonexistent.md#tasks":
✗ File not found: nonexistent.md
```

This mode helps users understand destination paths before attempting refile operations.

### Examples

```bash
# Basic refiling
jot refile "inbox.md#meeting" --to "work.md#projects"

# Cross-file operation
jot refile "notes.md#research" --to "archive.md#old-projects"

# Prepend positioning
jot refile "inbox.md#urgent" --to "work.md#tasks" --prepend

# Complex path matching
jot refile "daily.md#standup/blockers" --to "work.md#issues/current"
```

### Error Handling

```bash
# No matches found
$ jot refile "inbox.md#nonexistent" --to "work.md#tasks"
Error: No headings found matching path "nonexistent" in inbox.md

# Multiple matches
$ jot refile "inbox.md#meeting" --to "work.md#tasks"
Error: Multiple headings match "meeting" in inbox.md:
  - "Team Meeting" at line 15
  - "Client Meeting" at line 45
  Use a more specific path like "#team/meeting" or "#client/meeting"

# Invalid destination
$ jot refile "inbox.md#valid" --to "nonexistent.md#tasks"
Error: Destination file "nonexistent.md" not found
```

## Technical Implementation

### Core Components

**1. Path Parser**

```go
type HeadingPath struct {
    File     string   // "inbox.md"
    Segments []string // ["meeting", "attendees"]
}

func ParsePath(path string) (*HeadingPath, error)
```

**2. Subtree Extractor**

```go
type Subtree struct {
    Heading     string    // Original heading text
    Level       int       // Original heading level (1-6)
    Content     []byte    // Full subtree content (markdown)
    StartOffset int       // Byte position in source
    EndOffset   int       // Byte position in source
}

func ExtractSubtree(filePath string, path *HeadingPath) (*Subtree, error)
```

**3. Destination Resolver**

```go
type DestinationTarget struct {
    File         string // Target file path
    TargetLevel  int    // Level where content should be inserted
    InsertOffset int    // Byte position for insertion
    CreatePath   []string // Missing headings to create
}

func ResolveDestination(destPath string, prepend bool) (*DestinationTarget, error)
```

**4. Level Transformer**

```go
func TransformSubtreeLevel(subtree *Subtree, newBaseLevel int) []byte
```

### Implementation Phases

**Phase 1: Core Infrastructure**

- Path parsing and validation
- AST-based subtree extraction
- Basic file operations
- Level transformation logic

**Phase 2: Destination Management**

- Auto-creation of missing headings
- Insertion position calculation
- Prepend/append positioning
- Conflict detection and resolution

**Phase 3: Error Handling & UX**

- Comprehensive error messages
- Match disambiguation
- Dry-run mode (`--dry-run`)
- Verbose output (`--verbose`)

**Phase 4: Advanced Features**

- Interactive mode (`--interactive`)
- Bulk operations
- Undo functionality
- Configuration options

## Benefits Over Current System

1. **Flexibility**: Work with any heading structure, not just timestamped notes
2. **Intuitive**: Path-based selectors mirror directory navigation
3. **Powerful**: Move complex nested content as atomic units
4. **Safe**: Exact matching prevents accidental operations
5. **Extensible**: Foundation for advanced workflows
6. **Editor-Friendly**: Works well with editor integration

## Migration Strategy

Since this breaks backward compatibility:

1. **Deprecation Notice**: Add warning to current `jot refile` command
2. **New Command**: Implement as `jot refile-v2` initially
3. **Documentation**: Comprehensive migration guide with examples
4. **Cutover**: Replace `jot refile` with new implementation in next major version
5. **Legacy Support**: Provide conversion tool for existing workflows if needed

## Configuration Options

Future configuration support via `.jotrc`:

```json5
{
  refile: {
    defaultDestination: "inbox.md#Archive",
    matchStrategy: "contains", // "contains" | "exact" | "regex"
    autoCreate: true,
    confirmMatches: false,
    defaultPosition: "append", // "append" | "prepend"
  },
}
```

## Success Criteria

1. **Functional**: All specified operations work correctly
2. **Reliable**: Robust error handling and edge case management
3. **Intuitive**: Tech workers can use effectively without extensive documentation
4. **Performant**: Fast operations on typical note collections
5. **Extensible**: Foundation supports planned advanced features

## Conclusion

This rework transforms `jot refile` from a simple note mover into a powerful subtree management system. By drawing inspiration from Org-mode while maintaining jot's pragmatic approach, we create a tool that scales from simple workflows to complex knowledge management systems.

The path-based selector syntax provides an intuitive yet flexible interface, while the auto-creation and level transformation features ensure seamless integration with existing note hierarchies. This foundation enables future enhancements like interactive selection, bulk operations, and advanced workflow automation.

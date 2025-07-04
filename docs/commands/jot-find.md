[Documentation](../README.md) > [Commands](README.md) > find

# jot find

## Description

The `jot find` command searches through notes in `inbox.md`, `lib/` directory, and optionally archived notes. It supports keyword search and full-text search with context display. Results are ranked by relevance and recency.

This command is useful for:
- Finding specific content across your workspace
- Locating notes before refiling or archiving
- Interactive search with FZF integration
- Automated content discovery with JSON output

## Usage

## Arguments

| Argument | Description |
|----------|-------------|
| `<query>` | Search query (supports multiple words) |

## Options

| Option | Description |
|--------|-------------|
| `--archive` | Include archived notes in search |
| `--limit` | Limit number of results (default: 20) |
| `--interactive` | Use FZF for interactive search (requires `JOT_FZF=1`) |

## Search Scope

The command searches through:
- **inbox.md** - Unprocessed notes
- **lib/** - All markdown files in the library
- **archive/** - Archived notes (when `--archive` is specified)

## Examples

### Basic Search

```bash
# Search for a phrase
jot find "meeting notes"

# Search for single terms
jot find golang
jot find todo

# Search with multiple keywords
jot find project management api
```

### Limiting Results

```bash
# Limit to 10 results
jot find golang --limit 10

# Show more results
jot find todo --limit 50
```

### Including Archive

```bash
# Include archived notes in search
jot find todo --archive

# Search old meeting notes
jot find "quarterly review" --archive
```

### Interactive Search

```bash
# Interactive mode with FZF (requires JOT_FZF=1)
export JOT_FZF=1
jot find todo --interactive

# Interactive search with archive
jot find "project" --interactive --archive
```

## Output Format

### Standard Output

```
Searching for: meeting notes
Found 3 matches for 'meeting notes':

lib/work.md:15 | ## Meeting Notes from Sprint Planning
lib/projects.md:42 | - Review meeting notes from last week
inbox.md:8 | Daily standup meeting notes and action items
```

### JSON Output

```bash
jot find "meeting" --json
```

```json
{
  "query": "meeting",
  "total_found": 3,
  "results": [
    {
      "file_path": "/path/to/workspace/lib/work.md",
      "relative_path": "lib/work.md",
      "line_number": 15,
      "context": "## Meeting Notes from Sprint Planning",
      "score": 1
    }
  ],
  "search_info": {
    "include_archive": false,
    "limit": 20,
    "limited": false
  },
  "metadata": {
    "command": "find",
    "success": true,
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

## Interactive Mode Features

When using `--interactive` with `JOT_FZF=1`:

- **Live filtering**: Type to filter results in real-time
- **Enhanced selectors**: Shows file:line#heading paths
- **Preview**: View context around matches
- **Direct access**: Select result to view or edit

### Interactive Mode Requirements

- `JOT_FZF=1` environment variable
- `fzf` command available in PATH
- Terminal with interactive capabilities

## Search Algorithm

### Relevance Scoring

Results are scored based on:
- **Keyword frequency**: Number of query matches in line
- **Context relevance**: Proximity to headings and structure
- **File recency**: Recent modifications scored higher

### Context Display

- **Line trimming**: Long lines are truncated to 80 characters
- **Match centering**: Query matches are centered in context
- **Ellipsis handling**: Truncated content shows `...`

## Result Enhancement

In interactive mode, results are enhanced with:
- **Heading context**: Shows nearest markdown heading
- **Path selectors**: Creates `file:line#heading` format
- **Structured navigation**: Easier to locate specific sections

## Error Conditions

| Error | Cause | Solution |
|-------|-------|----------|
| `No matches found` | Query doesn't match any content | Try different keywords or check spelling |
| `Interactive mode not available` | FZF not configured with JSON output | Remove `--json` flag or use standard mode |
| `Permission denied` | Cannot read search files | Check file permissions |
| `FZF not found` | Interactive mode requires FZF | Install `fzf` or use standard mode |

## Performance Considerations

- **Large workspaces**: Use `--limit` to reduce result count
- **Archive inclusion**: May slow search significantly
- **Interactive mode**: Requires additional processing for enhanced results

## Cross-references

- [jot refile](jot-refile.md) - Moving found content
- [jot peek](jot-peek.md) - Previewing found content
- [jot capture](jot-capture.md) - Adding new content
- [jot archive](jot-archive.md) - Working with archived notes

## See Also

- [Global Options](README.md#global-options)
- [Interactive Workflows](../topics/interactive.md)
- [Search Strategies](../topics/search.md)
- [Workspace Structure](../topics/workspace.md)

[Documentation](../README.md) > [Commands](README.md) > capture

# jot capture

## Description

The `jot capture` command is the primary way to add new content to your jot workspace. It supports multiple input methods and integrates with the template system for structured note-taking. By default, notes are captured to `inbox.md` for later organization.

This command is useful for:
- Quick thought capture during workflow
- Structured note-taking with templates  
- Batch content processing from other tools
- Interactive note writing in your editor

## Usage

```bash
jot capture [template] [options]
```

## Arguments

| Argument | Description | Default |
|----------|-------------|---------|
| `template` | Name of template to use for structured capture | none |

## Options

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--content TEXT` | | Direct content to capture | none |
| `--template NAME` | | Explicit template selection | none |
| `--no-verify` | | Skip pre-capture hooks | false |
| `--json` | | Output in JSON format | false |

*See [Global Options](README.md#global-options) for additional flags.*

## Examples

### Quick content capture

```bash
jot capture --content "Remember to update the API documentation"
```

Appends the content directly to `inbox.md`.

### Interactive capture with editor

```bash
jot capture
```

Opens your default editor (`$EDITOR`) to write a note, then appends to `inbox.md`.

### Template-based capture

```bash
jot capture meeting
```

Uses the "meeting" template (must be approved) and opens editor with rendered template content.

### Explicit template selection

```bash
jot capture --template daily
```

Equivalent to `jot capture daily` - uses the "daily" template.

### Piped content capture

```bash
echo "Meeting notes from standup" | jot capture
```

```bash
git log --oneline -5 | jot capture --template code-review
```

### Capture with template and additional content

```bash
jot capture standup --content "Completed API design review"
```

Renders the "standup" template and appends the additional content.

### JSON output for automation

```bash
jot capture --content "Automated note" --json
```

Output:

```json
{
  "operation": "capture",
  "content_added": "Automated note",
  "destination": "inbox.md",
  "template_used": null,
  "word_count": 2,
  "metadata": {
    "timestamp": "2025-07-04T16:30:00Z",
    "success": true
  }
}
```

## What Happens

When you run `jot capture`, the command:

1. **Runs pre-capture hooks** (unless `--no-verify` is used)
2. **Determines input method** based on arguments and options
3. **Processes template** if specified (requires approval)
4. **Collects content** from direct input, stdin, or editor
5. **Writes to destination** (default: `inbox.md`, or template-specified)
6. **Runs post-capture hooks** after successful capture

## Input Methods

**Direct Content**
```bash
jot capture --content "Your note text"
```
Immediately captures the provided text.

**Standard Input (Piped)**
```bash
command | jot capture
echo "content" | jot capture [template]
```
Captures piped content, optionally with template structure.

**Editor-Based**
```bash
jot capture
jot capture [template]
```
Opens your editor for interactive writing. Template provides starting content.

**Template Selection Priority**
1. Positional argument: `jot capture template-name`
2. Explicit flag: `--template template-name`  
3. No template: plain capture

## Template Integration

Templates provide structure and consistency:

- **Require approval** before use (see [jot template approve](jot-template.md#approve))
- **Include frontmatter** for destination and refile mode configuration
- **Support dynamic content** through shell command execution
- **Open in editor** for customization during capture

## Destination Handling

Content destination is determined by:

1. **Template frontmatter** `destination` field
2. **Default workspace inbox** (`inbox.md`)
3. **Refile mode** from template (`append`, `prepend`)

## Hook Integration

Capture integrates with the hooks system:

- **Pre-capture hooks** can modify content before writing
- **Post-capture hooks** can trigger actions after successful capture
- **Use `--no-verify`** to skip hooks for automation

## Error Conditions

| Error | Cause | Solution |
|-------|-------|----------|
| "Template not found" | Specified template doesn't exist | Check available templates with [jot template list](jot-template.md#list) |
| "Template not approved" | Template contains unapproved shell commands | Approve template with [jot template approve](jot-template.md#approve) |
| "Editor failed" | Editor exited with error or empty content | Check `$EDITOR` setting and try again |
| "No workspace found" | Not in a jot workspace | Run [jot init](jot-init.md) or use `--workspace` |
| "Permission denied" | Cannot write to destination file | Check file permissions |

## See Also

- [jot template](jot-template.md) - Manage templates for structured capture
- [jot refile](jot-refile.md) - Organize captured notes
- [jot status](jot-status.md) - Check workspace and recent activity

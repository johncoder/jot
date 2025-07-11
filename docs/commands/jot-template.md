[Documentation](../README.md) > [Commands](README.md) > template

# jot template

## Description

Templates are stored in `.jot/templates/` and require explicit approval before they can execute shell commands for security. Templates contain frontmatter for configuration and can generate dynamic content using shell commands.

This command group is useful for:
- Creating structured, consistent note formats
- Building templates with dynamic content (dates, git info, etc.)
- Managing template approval for security
- Previewing template output before use

## list

List all available templates and their approval status.

### Usage

```bash
jot template list [options]
```

### Options

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--json` | | Output in JSON format | false |

*See [Global Options](README.md#global-options) for additional flags.*

### Examples

#### Basic template listing

```bash
jot template list
```

Output:

```
Templates:
  meeting     ✓ approved   (modified 2 days ago)
  daily       ✗ pending    (created 1 hour ago)
  standup     ✓ approved   (modified 1 week ago)
```

#### JSON output

```bash
jot template list --json
```

Output:

```json
{
  "operation": "template_list",
  "templates": [
    {
      "name": "meeting",
      "approved": true,
      "hash": "a1b2c3d4",
      "modified": "2025-07-02T10:30:00Z"
    },
    {
      "name": "daily", 
      "approved": false,
      "hash": "",
      "modified": "2025-07-04T15:30:00Z"
    }
  ],
  "metadata": {
    "timestamp": "2025-07-04T16:30:00Z",
    "success": true
  }
}
```

### What Happens

The `list` subcommand:
1. **Scans the templates directory** (`.jot/templates/`)
2. **Checks approval status** by reading approval metadata
3. **Displays template information** with approval indicators
4. **Shows modification times** for template maintenance

## new

Create a new template with default structure and open it in your editor.

### Usage

```bash
jot template new <name> [options]
```

### Arguments

| Argument | Description |
|----------|-------------|
| `name` | Name of the template to create |

### Examples

#### Create a new template

```bash
jot template new meeting
```

This creates `.jot/templates/meeting.md` and opens it in your editor with default content:

```
---
destination: inbox.md
refile_mode: append
---
## Meeting - $(date '+%Y-%m-%d %H:%M')


```

#### After creation

```
Created template 'meeting'

To use this template, first approve it:
  jot template approve meeting
```

### What Happens

The `new` subcommand:
1. **Creates template file** in `.jot/templates/<name>.md`
2. **Adds default frontmatter** with destination and refile mode
3. **Opens in editor** using `$EDITOR` environment variable
4. **Prompts for approval** after creation

## edit

Modify an existing template in your editor, or overwrite it with content from stdin.

### Usage

```bash
# Edit in your editor
jot template edit <name>

# Overwrite with piped content (no editor)
echo "new template content" | jot template edit <name>
cat template.md | jot template edit <name>
```

### Arguments

| Argument | Description |
|----------|-------------|
| `name` | Name of the template to edit |

### Examples

#### Edit existing template in your editor

```bash
jot template edit meeting
```

#### Overwrite a template with piped content

```bash
echo "---\ndestination: inbox.md\n---\n# New Template" | jot template edit meeting
cat my-template.md | jot template edit meeting
```

If you pipe content to `jot template edit`, the template will be overwritten with the piped content and the editor will not be launched.

### What Happens

The `edit` subcommand:
1. **If stdin is a pipe**, reads content from stdin and overwrites the template file (no editor is launched)
2. **Otherwise**, opens the template file in your editor
3. **Checks for modifications** after editing
4. **Invalidates approval** if shell commands changed
5. **Prompts for re-approval** if needed

## approve

Approve a template for shell command execution.

### Usage

```bash
jot template approve <name> [options]
```

### Arguments

| Argument | Description |
|----------|-------------|
| `name` | Name of the template to approve |

### Examples

#### Approve a template

```bash
jot template approve meeting
```

Output:

```
Template 'meeting' approved for execution.
Hash: a1b2c3d4e5f6
```

### What Happens

The `approve` subcommand:
1. **Reviews template content** for shell commands
2. **Calculates content hash** for security validation
3. **Stores approval metadata** with hash
4. **Enables template execution** for [jot capture](jot-capture.md)

## view

Display template content without executing shell commands.

### Usage

```bash
jot template view <name> [options]
```

### Arguments

| Argument | Description |
|----------|-------------|
| `name` | Name of the template to view |

### Examples

#### View template content

```bash
jot template view meeting
```

Output:

```
---
destination: inbox.md
refile_mode: append
---
## Meeting - $(date '+%Y-%m-%d %H:%M')

**Date:** $(date '+%B %d, %Y')
**Attendees:** 

**Agenda:**
- 

**Notes:**


**Action Items:**
- 
```

### What Happens

The `view` subcommand:
1. **Reads template file** from `.jot/templates/`
2. **Displays raw content** without shell command execution
3. **Shows approval status** and metadata

## render

Preview template output with shell command execution.

### Usage

```bash
jot template render <name> [options]
```

### Arguments

| Argument | Description |
|----------|-------------|
| `name` | Name of the template to render |

### Examples

#### Render template with dynamic content

```bash
jot template render meeting
```

Output:

```
---
destination: inbox.md
refile_mode: append
---
## Meeting - 2025-07-04 16:30

**Date:** July 04, 2025
**Attendees:** 

**Agenda:**
- 

**Notes:**


**Action Items:**
- 
```

### What Happens

The `render` subcommand:
1. **Checks template approval** status
2. **Executes shell commands** in template content
3. **Displays rendered output** with dynamic content
4. **Respects security settings** and approval requirements

## remove

Remove (delete) a template file from the templates directory.

### Usage

```bash
jot template remove <name> [options]
```

### Arguments

| Argument | Description |
|----------|-------------|
| `name`   | Name of the template to remove |

### Examples

#### Remove a template

```bash
jot template remove meeting
```

#### Remove a template with JSON output

```bash
jot template remove meeting --json
```

### What Happens

The `remove` subcommand:
1. **Checks if the template file exists** in `.jot/templates/`
2. **Deletes the template file** if found
3. **Prints a confirmation message** (or outputs JSON if `--json` is used)
4. **Returns an error** if the template does not exist or cannot be deleted

## Template Structure

Templates consist of:

**Frontmatter**
- `destination`: Target file (default: `inbox.md`)
- `refile_mode`: How to add content (`append`, `prepend`)

**Content**
- Markdown content with optional shell commands
- Shell commands use `$(command)` syntax
- Dynamic content is executed during rendering

## Error Conditions

| Error | Cause | Solution |
|-------|-------|----------|
| "Template not found" | Template doesn't exist | Check name with `jot template list` |
| "Template not approved" | Contains unapproved shell commands | Run `jot template approve <name>` |
| "Editor not found" | `$EDITOR` not set or invalid | Set `EDITOR` environment variable |
| "Permission denied" | Cannot write to templates directory | Check `.jot/templates/` permissions |

## See Also

- [jot capture](jot-capture.md) - Use templates for note capture
- [jot status](jot-status.md) - Check template counts and status
- [Configuration Guide](../user-guide/configuration.md) - Template directory configuration and editor integration

# jot Templates

## Overview
Template system for structured note capture in jot, allowing users to define reusable note formats with shell command integration for dynamic content generation.

## Problem Statement
Currently, `jot capture` creates simple timestamped entries, but users need:
- Structured note formats (meeting notes, daily logs, project updates)
- Dynamic content generation using shell commands (dates, git info, etc.)
- Consistent metadata fields and sections
- Security controls for shell command execution

## Requirements (Defined Through Q&A)

### Template Usage Patterns
- Templates for specific note types (meetings, standups, project updates, journals)
- Workspace-specific templates stored in `.jot/templates/`
- Named templates accessible via `--template <name>`
- Default template support for quick capture

### Template Content & Syntax
- **Bash-friendly templates** supporting shell command execution
- **Shell command substitution** using `$(command)` syntax for dynamic content
- **Template preprocessing** where jot executes shell commands before presenting content
- **Metadata integration** for refile/archive automation using frontmatter

### Template Workflows
1. **Interactive Template Capture (Default)**: `jot capture --template meeting`
   - Renders template with shell commands executed
   - Opens in editor for user completion
   - Appends final content to inbox.md

2. **Quick Content with Template**: `jot capture --template standup --content "Completed API design"`
   - Applies template structure
   - Appends provided content to template body
   - No editor interaction

3. **Piped Input with Template**: `echo "Notes here" | jot capture --template meeting`
   - Renders template structure
   - Appends piped content to template body
   - Direct append to inbox.md

### Security Model
- **Explicit approval required** before template execution
- **Permission tokens** based on template content hash stored in `.jot/template_permissions`
- **Template invalidation** when content changes (hash mismatch)
- **Error on unauthorized use** with clear approval workflow
- Approval workflow: `jot template approve <name>`

## Implementation Design

### Updated Template Structure
Templates now support specifying a destination file for captured notes using the `destination_file` field in the frontmatter:

```markdown
---
destination_file: lib/meeting_notes.md
refile_to: meetings/$(date '+%Y')/
archive_after: 30days
tags: [meeting, work]
---
# Meeting Notes - $(date '+%A, %B %d, %Y')

**Date:** $(date '+%Y-%m-%d')
**Time:** $(date '+%H:%M %Z')
**Branch:** $(git branch --show-current 2>/dev/null || echo "N/A")

## Attendees


## Agenda


## Notes


## Action Items

```

### Command Structure
- `jot template list` - Show available templates and approval status
- `jot template new <name>` - Create new template (opens in editor)
- `jot template edit <name>` - Edit existing template (requires re-approval)
- `jot template approve <name>` - Approve template for shell execution
- `jot template remove <name>` - Delete template and permissions
- `jot template show <name>` - Display template content

### Capture Integration
- `jot capture --template <name>`: Captures notes using the specified template and appends them to the `destination_file` if provided. Defaults to `inbox.md` if not specified.
- `jot capture --template <name> --content "text"`: Appends content to the template and stores it in the `destination_file`.
- `echo "text" | jot capture --template <name>`: Pipes content to the template and stores it in the `destination_file`.

### Storage & Security
- **Templates**: `.jot/templates/<name>.md`
- **Permissions**: `.jot/template_permissions` (contains approved template hashes)
- **Permission format**: `<template_name>:<sha256_hash>`
- **Hash calculation**: SHA256 of template file content (excluding frontmatter metadata)

## Technical Implementation

### Core Components
1. **Template Manager** (`internal/template/template.go`)
   - Template CRUD operations
   - Shell command execution
   - Permission validation
   - Frontmatter parsing

2. **Template Commands** (`cmd/template.go`)
   - CLI interface for template management
   - Approval workflow
   - Editor integration

3. **Capture Enhancement** (`cmd/capture.go`)
   - Template integration
   - Content merging logic
   - Editor workflow

### Security Considerations
- Templates execute arbitrary shell commands (similar to Makefiles, git hooks)
- Hash-based approval prevents tampering
- Clear error messages guide users through approval process
- Templates run in workspace context with user permissions
- Shell command failures don't break note capture

### Cross-Platform Compatibility
- Shell command execution varies by OS (bash/zsh on Unix, cmd/powershell on Windows)
- Template examples should include cross-platform alternatives
- Graceful degradation when shell commands fail

## User Stories (Implemented)

✅ **As a user, I want to create templates with dynamic content**
- Shell command integration with `$(command)` syntax
- Automatic execution during template rendering

✅ **As a user, I want secure template execution**  
- Explicit approval workflow prevents unauthorized command execution
- Template changes require re-approval

✅ **As a user, I want flexible capture workflows**
- Editor-first approach with `--content` for quick append
- Piped input support for programmatic note creation

✅ **As a user, I want templates to facilitate organization**
- Frontmatter metadata support for automated refile/archive operations
- Template-based structure improves note consistency

## Example Templates

### Daily Standup Template
```markdown
---
refile_to: standups/$(date '+%Y-%m')/
tags: [standup, daily]
---
# Daily Standup - $(date '+%Y-%m-%d')

**Sprint:** 
**Weather:** $(curl -s "wttr.in/?format=3" 2>/dev/null || echo "☀️")

## Yesterday
- 

## Today  
- 

## Blockers
- 

## Notes

```

### Meeting Notes Template  
```markdown
---
refile_to: meetings/$(date '+%Y')/
archive_after: 90days
tags: [meeting]
---
# $(echo -n "Meeting: "; read -r topic; echo "$topic") - $(date '+%Y-%m-%d')

**Date:** $(date '+%A, %B %d, %Y')
**Time:** $(date '+%H:%M %Z')  
**Project:** $(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "N/A")
**Attendees:** 

## Agenda


## Discussion Notes


## Decisions Made
- 

## Action Items
- [ ] 

## Follow-up
**Next meeting:** 
**Owner:** 
```

### Bug Report Template
```markdown
---
refile_to: bugs/$(date '+%Y-%m')/
tags: [bug, issue]
---
# Bug Report - $(date '+%Y-%m-%d %H:%M')

**Severity:** 
**Component:** 
**Environment:** $(uname -s) $(uname -r)
**Git Hash:** $(git rev-parse --short HEAD 2>/dev/null || echo "N/A")

## Description


## Steps to Reproduce
1. 
2. 
3. 

## Expected Behavior


## Actual Behavior


## Additional Context

```

## CLI Usage Examples

```bash
# Create and manage templates
jot template new meeting
jot template list
jot template approve meeting
jot template edit meeting

# Use templates for capture
jot capture --template meeting
jot capture --template standup --content "Finished user auth"
echo "Found critical bug in payment flow" | jot capture --template bug

# Quick workflows
alias daily="jot capture --template standup"
alias meeting="jot capture --template meeting"
alias bug="jot capture --template bug --content"
```

---

## Status
- **Phase**: Implementation Ready
- **Created**: June 6, 2025
- **Last Updated**: June 6, 2025
- **Next Steps**: Complete implementation, testing, and documentation

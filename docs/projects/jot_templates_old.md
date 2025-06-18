# jot Templates

## Overview
Add template system for note entries to standardize note capture and improve organization.

## Problem Statement
Currently, `jot capture` creates simple timestamped entries, but users may want:
- Structured note formats (meeting notes, daily logs, project updates)
- Consistent metadata fields
- Predefined sections or prompts
- Context-specific templates based on location or tags

## User Stories
- As a user, I want to create notes with consistent structure
- As a user, I want different templates for different types of notes
- As a user, I want to quickly apply a template during capture

## Requirements (Confirmed)

### Core Design Principles
1. **Editor-first workflow**: Templates open in `$EDITOR` by default for interactive editing
2. **Content appending**: Piped/argument content appends to template body (under headers)
3. **Security model**: Templates require explicit user permission via hash-based validation
4. **Shell integration**: Templates support bash-friendly shell command execution
5. **Git-like CLI**: Integrate seamlessly with existing `jot capture` command

### Template Workflow
- **Default**: `jot capture --template meeting` → opens template in editor
- **Quick content**: `jot capture --template meeting --content "Discussed API design"`
- **Piped content**: `echo "Notes here" | jot capture --template meeting`
- **Permission required**: Templates must be explicitly approved before first use

### Template Format
```markdown
# Meeting Notes - $(date '+%Y-%m-%d %H:%M')

**Project:** $(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "N/A")
**Attendees:** 

## Agenda


## Discussion


## Action Items

<!-- Content from --content or stdin appends here -->
```

### Security Model
- Templates stored in `.jot/templates/`
- Permission tracking in `.jot/template_permissions`
- Hash-based validation prevents unauthorized template execution
- Template changes invalidate permissions (require re-approval)

### CLI Integration
```bash
# Template management
jot template list                    # List available templates
jot template new meeting             # Create new template
jot template edit meeting            # Edit existing template
jot template approve meeting         # Grant permission to execute

# Capture with templates
jot capture --template meeting       # Open in editor (default)
jot capture --template meeting --content "Quick note"
echo "Details" | jot capture --template standup
```

## Technical Considerations
- Store templates in `.jot/templates/` as markdown files with shell command support
- Permission file `.jot/template_permissions` contains SHA256 hashes of approved templates
- Use Go's `os/exec` for shell command execution in templates
- Integrate with existing editor package for interactive editing
- Cross-platform shell compatibility (bash/zsh/cmd/powershell)

---
*Created: 2025-06-06*
*Status: Planning*

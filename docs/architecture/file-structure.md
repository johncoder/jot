# File Structure

Understanding how jot organizes files and directories in a workspace.

## Workspace Overview

A jot workspace is a directory containing a `.jot` subdirectory. This structure keeps user-facing content separate from internal implementation details.

```
my-notes/                    # Workspace root
├── inbox.md                 # Capture destination
├── lib/                     # Organized notes
│   ├── work.md
│   ├── personal.md
│   └── projects/
│       └── project-x.md
└── .jot/                    # Internal jot data
    ├── config.json          # Workspace configuration
    ├── logs/                # Operation logs
    ├── templates/           # Note templates
    ├── template_permissions # Template security approvals
    └── hooks/               # Automation scripts
```

## User-Facing Files

### `inbox.md` - Capture Destination

The primary destination for new notes. All `jot capture` operations append to this file unless redirected by templates.

**Characteristics:**

- **Standard Markdown** - Compatible with any Markdown editor or viewer
- **Timestamped entries** - Each capture includes timestamp
- **Append-only** - New content is added to the end
- **Temporary staging** - Content is eventually refiled to organized locations

**Example content:**

```markdown
# Inbox

## 2025-01-15 09:23

Quick thought about improving the API rate limiting algorithm.

## 2025-01-15 14:45

Meeting notes from architecture review:

- Discussed microservices approach
- Decided on API gateway pattern
- Action: Draft service boundaries document

## 2025-01-15 16:30

Learned about React.memo performance optimization.
Source: https://react.dev/reference/react/memo
```

### `lib/` - Organized Notes

The directory for organized, categorized notes. Files here are created during the refiling process.

**Organization patterns:**

- **Topic-based:** `work.md`, `personal.md`, `learning.md`
- **Project-based:** `projects/project-alpha.md`, `projects/project-beta.md`
- **Time-based:** `daily/2025-01.md`, `weekly/week-03.md`
- **Type-based:** `meetings.md`, `ideas.md`, `references.md`

**File naming conventions:**

- Use kebab-case for multi-word files: `project-alpha.md`
- Use clear, descriptive names: `meeting-notes.md` not `notes.md`
- Group related files in subdirectories when helpful

### Optional Organization Directories

**`archive/`** - Long-term storage for old notes:

```
archive/
├── 2024/
│   ├── 01/
│   │   ├── old-project.md
│   │   └── completed-tasks.md
│   └── 02/
└── 2025/
```

**`templates/`** - User templates (alternative to `.jot/templates/`):

```
templates/
├── meeting.md
├── daily-log.md
└── project-update.md
```

## Internal Structure (`.jot/`)

### Configuration Files

**`config.json`** - Workspace-specific configuration:

```json
{
  "templates": {
    "default_capture": "note",
    "directories": {
      "daily": "lib/daily/",
      "meetings": "lib/meetings/"
    }
  },
  "hooks": {
    "post_capture": true,
    "post_refile": true
  },
  "organization": {
    "inbox_size_warning": 50
  }
}
```

**`template_permissions`** - Template security approvals:

```
meeting:sha256:a1b2c3d4e5f6...
daily:sha256:f6e5d4c3b2a1...
standup:sha256:9876543210ab...
```

### Templates Directory

**`.jot/templates/`** - Workspace-specific templates:

```
templates/
├── meeting.md               # Meeting notes template
├── daily.md                 # Daily log template
├── project-update.md        # Project status template
└── bug-report.md           # Bug investigation template
```

### Logs and History

**`.jot/logs/`** - Operation history and debugging:

```
logs/
├── captures.log             # Capture operations
├── refiling.log            # Refiling history
├── errors.log              # Error messages
└── hooks.log               # Hook execution logs
```

### Automation Scripts

**`.jot/hooks/`** - Automation and workflow scripts:

```
hooks/
├── post-capture.sh         # Run after each capture
├── post-refile.sh          # Run after refiling
├── pre-archive.sh          # Run before archiving
└── common.sh               # Shared utility functions
```

## File Format Standards

### Markdown Files

All user content is stored in **standard Markdown** with optional YAML frontmatter:

```markdown
---
tags: [meeting, work]
project: alpha
created: 2025-01-15T14:30:00Z
---

# Meeting: API Architecture Review

## Attendees

- John Doe (Tech Lead)
- Jane Smith (Backend Dev)
- Bob Johnson (Frontend Dev)

## Discussion

...
```

### Configuration Files

**JSON5 format** for human readability:

```json5
{
  // Comments are allowed
  workspaces: {
    work: "/home/user/work-notes",
    personal: "~/personal-notes",
  },

  // Trailing commas are fine
  editor: "vim",
}
```

### Template Files

Templates are Markdown with shell command substitution:

```markdown
---
destination: lib/daily/$(date '+%Y-%m').md
tags: [daily]
---

# Daily Log - $(date '+%Y-%m-%d')

**Weather:** $(curl -s wttr.in?format=3)
**Time:** $(date '+%H:%M')

## Notes
```

## Directory Discovery

jot finds workspaces by searching for `.jot` directories:

1. **Current directory** - Check if `.jot` exists here
2. **Parent directories** - Walk up directory tree looking for `.jot`
3. **Named workspaces** - Check configured workspace locations
4. **Default workspace** - Fall back to configured default

**Discovery examples:**

```bash
# Working in a project subdirectory
/home/user/notes/projects/alpha/src/
# Finds workspace at: /home/user/notes/.jot

# Working in any directory
/some/random/path/
# With ~/.jotrc containing: {"workspaces": {"default": "~/notes"}}
# Uses workspace at: /home/user/notes/.jot
```

## Workspace Patterns

### Single Workspace

Simple setup for personal notes:

```
~/notes/
├── inbox.md
├── lib/
│   ├── work.md
│   ├── personal.md
│   └── learning.md
└── .jot/
```

### Project-Specific Workspaces

Separate workspace for each project:

```
~/projects/
├── project-alpha/
│   ├── src/
│   ├── docs/
│   │   ├── inbox.md
│   │   ├── lib/
│   │   └── .jot/
│   └── README.md
└── project-beta/
    ├── src/
    ├── notes/
    │   ├── inbox.md
    │   ├── lib/
    │   └── .jot/
    └── README.md
```

### Hybrid Workspace

Shared workspace with project-specific organization:

```
~/all-notes/
├── inbox.md
├── lib/
│   ├── personal/
│   ├── work/
│   ├── projects/
│   │   ├── alpha.md
│   │   └── beta.md
│   └── learning/
└── .jot/
    ├── templates/
    │   ├── personal-note.md
    │   ├── work-update.md
    │   └── project-status.md
    └── config.json
```

## File Permissions and Security

### Recommended Permissions

```bash
# Workspace directories
chmod 755 .                  # Workspace root
chmod 755 lib/               # Notes directory
chmod 755 .jot/              # Internal directory

# User files
chmod 644 inbox.md           # Capture file
chmod 644 lib/*.md           # Note files

# Configuration files
chmod 600 ~/.jotrc           # Global config (private)
chmod 644 .jot/config.json   # Workspace config

# Templates and scripts
chmod 644 .jot/templates/*.md    # Templates
chmod 755 .jot/hooks/*.sh        # Executable scripts
chmod 644 .jot/template_permissions # Security data
```

### Security Considerations

- **Configuration files** may contain sensitive paths - protect appropriately
- **Hook scripts** execute with user permissions - review before enabling
- **Template approvals** prevent unauthorized shell execution
- **Log files** may contain sensitive information - consider log rotation

## Backup and Portability

### What to Backup

**Essential files:**

- `inbox.md` and all files in `lib/`
- `.jot/config.json` (workspace configuration)
- `.jot/templates/` (custom templates)
- `.jot/template_permissions` (security approvals)

**Optional files:**

- `.jot/hooks/` (automation scripts)
- `.jot/logs/` (operation history)

**Example backup script:**

```bash
#!/bin/bash
BACKUP_DIR="backup-$(date +%Y%m%d)"
mkdir "$BACKUP_DIR"

# Essential user content
cp inbox.md "$BACKUP_DIR/"
cp -r lib/ "$BACKUP_DIR/"

# Workspace configuration
cp .jot/config.json "$BACKUP_DIR/"
cp -r .jot/templates/ "$BACKUP_DIR/"
cp .jot/template_permissions "$BACKUP_DIR/"

tar -czf "${BACKUP_DIR}.tar.gz" "$BACKUP_DIR"
```

### Migration Between Systems

Notes are portable across systems:

1. **Copy files** - Standard Markdown files work anywhere
2. **Update paths** - Adjust workspace paths in configuration
3. **Re-approve templates** - Security approvals are system-specific
4. **Test hooks** - Verify automation scripts work on new system

## Integration with Version Control

### Git Integration

jot workspaces work well with git:

```bash
# Initialize git in workspace
git init
git add inbox.md lib/ .jot/config.json .jot/templates/
git commit -m "Initial jot workspace"

# Recommended .gitignore
echo ".jot/logs/" >> .gitignore
echo ".jot/template_permissions" >> .gitignore  # Optional
```

### Shared Workspaces

For team collaboration:

```bash
# Include essential files
git add inbox.md lib/ .jot/config.json .jot/templates/

# Exclude per-user data
echo ".jot/logs/" >> .gitignore
echo ".jot/template_permissions" >> .gitignore
echo ".jot/hooks/" >> .gitignore  # Or include if team-shared
```

## Troubleshooting File Issues

### Workspace Not Found

```bash
# Check for .jot directory
find . -name ".jot" -type d

# Verify workspace discovery
jot status --verbose

# Initialize if needed
jot init
```

### Permission Problems

```bash
# Fix common permission issues
chmod 755 .jot/
chmod 644 .jot/config.json
chmod 644 inbox.md lib/*.md

# Check file ownership
ls -la .jot/ lib/
```

### File Corruption

```bash
# Validate file formats
file inbox.md lib/*.md          # Should show "text"
json5 validate .jot/config.json # Check JSON5 syntax

# Recovery from backup
cp inbox.md.backup inbox.md
```

## See Also

- **[Design Principles](design-principles.md)** - Why files are organized this way
- **[Security Model](security-model.md)** - Template security and permissions
- **[Configuration Guide](../user-guide/configuration.md)** - Configuring file locations
- **[Getting Started](../user-guide/getting-started.md)** - Creating your first workspace

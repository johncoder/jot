# Templates

Templates in jot allow you to create structured, reusable note formats with dynamic content generation through shell commands.

## Overview

Templates help you:
- Create consistent note structures
- Generate dynamic content (dates, git info, system data)
- Streamline common note-taking patterns
- Maintain metadata for better organization

## Template Basics

### Creating Your First Template

```bash
# Create a new template
jot template new meeting

# This opens your editor with a basic template structure
```

### Template Structure

Templates are markdown files with optional frontmatter:

```markdown
---
destination: inbox.md
refile_mode: append
tags: [meeting, work]
---

# Meeting: $(date '+%Y-%m-%d')

**Date:** $(date)
**Attendees:** 

## Agenda

1. 

## Notes



## Action Items

- [ ] 

## Next Steps

```

### Using Templates

```bash
# Use template interactively (opens editor)
jot capture meeting

# Quick capture with template
jot capture meeting --content "Discussed API architecture"

# Pipe content to template
echo "Sprint planning discussion" | jot capture meeting
```

## Template Features

### Shell Command Execution

Templates support shell commands using `$(command)` syntax:

```markdown
# Daily Standup - $(date '+%A, %B %d, %Y')

**Weather:** $(curl -s wttr.in/YourCity?format=3)
**Git Status:** $(git status --porcelain | wc -l) files changed

## Yesterday
- 

## Today  
- 

## Blockers
- 
```

### Metadata and Frontmatter

Use YAML frontmatter to control template behavior:

```yaml
---
# Where the note should go
destination: lib/meetings.md

# How to add content (append, prepend, replace)
refile_mode: append

# Tags for organization
tags: [meeting, standup, team]

# Custom metadata
project: "web-app"
priority: "normal"
---
```

**Frontmatter Options:**
- `destination` - Target file for the note
- `refile_mode` - How to insert content (`append`, `prepend`, `replace`)
- `tags` - Array of tags for organization
- Custom fields for your workflow

### Dynamic Content Examples

```markdown
---
destination: lib/daily/$(date '+%Y-%m').md
tags: [daily, journal]
---

# Daily Notes - $(date '+%Y-%m-%d %A')

**Time:** $(date '+%H:%M')
**Location:** $(pwd | basename)
**Git Branch:** $(git branch --show-current 2>/dev/null || echo "not a git repo")

## Today's Focus


## Quick Notes


## System Info
- **Disk Usage:** $(df -h ~ | tail -1 | awk '{print $5}')
- **Load Average:** $(uptime | awk -F'load average:' '{print $2}')
```

## Security Model

Templates can execute shell commands, so jot implements a security approval system.

### Approval Process

1. **First Use**: Template requires approval
2. **Content Verification**: Changes to template invalidate approval
3. **Explicit Approval**: Use `jot template approve <name>`

```bash
# Check template status
jot template list

# Output shows approval status:
# meeting (✓ approved)
# standup (✗ needs approval)

# Approve a template
jot template approve standup
```

### How Approval Works

- Templates are hashed when approved
- Hash is stored in `.jot/template_permissions`
- Template changes require re-approval
- Prevents unauthorized code execution

### Approval Best Practices

1. **Review before approval** - Understand what commands will run
2. **Minimal commands** - Use simple, safe shell commands
3. **Regular review** - Periodically check your approved templates
4. **Team templates** - Share approved templates with your team

## Template Management

### Listing Templates

```bash
# List all templates with approval status
jot template list

# Output:
# Available templates:
#   meeting (✓ approved)
#   standup (✓ approved)  
#   daily (✗ needs approval)
#   project-update (✗ needs approval)
```

### Editing Templates

```bash
# Edit existing template
jot template edit meeting

# View template without editing
jot template view meeting

# Test template rendering
jot template render meeting
```

### Template Storage

Templates are stored in `.jot/templates/` as markdown files:

```
.jot/
├── templates/
│   ├── meeting.md
│   ├── standup.md
│   ├── daily.md
│   └── project-update.md
├── template_permissions
└── config.json
```

## Example Templates

### Meeting Template

```markdown
---
destination: lib/meetings.md
tags: [meeting]
---

# $(echo "$1" | tr '[:lower:]' '[:upper:]') Meeting - $(date '+%Y-%m-%d')

**Date:** $(date '+%A, %B %d, %Y')
**Time:** $(date '+%H:%M')
**Duration:** 

## Attendees
- 

## Agenda
1. 

## Discussion


## Decisions Made
- 

## Action Items
- [ ] 

## Next Meeting
- **Date:** 
- **Topics:** 
```

### Daily Standup Template

```markdown
---
destination: lib/standups/$(date '+%Y-%m').md
refile_mode: append
tags: [standup, daily]
---

## Standup - $(date '+%Y-%m-%d')

### Yesterday
- 

### Today
- 

### Blockers
- 

### Notes
- **Sprint:** $(git branch --show-current | sed 's/feature\///' | sed 's/-/ /g')
- **Commits:** $(git log --oneline --since="yesterday" | wc -l)

---
```

### Project Update Template

```markdown
---
destination: lib/projects/$(basename $(pwd)).md
tags: [project, update]
project: "$(basename $(pwd))"
---

# Project Update - $(date '+%Y-%m-%d')

**Project:** $(basename $(pwd))
**Branch:** $(git branch --show-current)
**Last Commit:** $(git log -1 --pretty=format:"%h - %s")

## Progress This Week
- 

## Completed
- 

## In Progress
- 

## Planned Next Week
- 

## Blockers
- 

## Metrics
- **Files Changed:** $(git diff --name-only HEAD~7 | wc -l)
- **Lines Added:** $(git diff --stat HEAD~7 | tail -1 | awk '{print $4}')
- **Tests:** $(find . -name "*test*" | wc -l)
```

### Learning Notes Template

```markdown
---
destination: lib/learning/$(date '+%Y-%m').md
tags: [learning, notes]
---

# Learning Notes - $(date '+%Y-%m-%d')

**Topic:** 
**Source:** 
**Time Spent:** 

## Key Concepts


## Code Examples

\`\`\`


\`\`\`

## Questions
- 

## Action Items
- [ ] 

## Related Resources
- 
```

## Advanced Template Patterns

### Conditional Content

```markdown
# Project Status - $(date '+%Y-%m-%d')

$(if git status --porcelain > /dev/null 2>&1; then echo "**Git Status:** $(git status --porcelain | wc -l) files changed"; fi)

$(if [ -f package.json ]; then echo "**Node Project:** $(node --version)"; fi)

$(if [ -f go.mod ]; then echo "**Go Project:** $(go version | awk '{print $3}')"; fi)
```

### Environment-Specific Content

```markdown
# Development Notes - $(date '+%Y-%m-%d')

**Environment:** $(echo $NODE_ENV || echo "development")
**User:** $(whoami)
**Host:** $(hostname)

$(if [ "$NODE_ENV" = "production" ]; then echo "⚠️  **PRODUCTION ENVIRONMENT**"; fi)
```

### Integration with External Tools

```markdown
# Weekly Review - $(date '+%Y-%m-%d')

## GitHub Activity
$(gh pr list --author @me --state merged --limit 5 --json title,url | jq -r '.[] | "- [\(.title)](\(.url))"')

## Jira Tickets
$(if command -v jira >/dev/null; then jira list --assignee $(whoami); fi)

## Calendar
$(if command -v calendar >/dev/null; then calendar -A 7; fi)
```

## Tips and Best Practices

### 1. Start Simple
Begin with basic templates and add complexity gradually:

```markdown
# Meeting - $(date)

## Notes


## Action Items
- 
```

### 2. Use Consistent Naming
Develop a naming convention:
- `meeting` - General meetings
- `standup` - Daily standups  
- `weekly-review` - Weekly reviews
- `project-update` - Project status

### 3. Test Before Approval
Always test templates before approving:

```bash
# Test template rendering
jot template render meeting

# Review the commands
jot template view meeting
```

### 4. Share Team Templates
Keep templates in version control:

```bash
# Export template
jot template view meeting > templates/meeting.md

# Import template
cp templates/meeting.md .jot/templates/
jot template approve meeting
```

### 5. Use Meaningful Metadata
Include useful frontmatter:

```yaml
---
destination: lib/$(date '+%Y')/meetings.md
tags: [meeting, $(git branch --show-current)]
project: "$(basename $(pwd))"
created: "$(date --iso-8601)"
---
```

## Troubleshooting Templates

### Permission Errors
```bash
# Check permissions
ls -la .jot/templates/

# Fix permissions
chmod 644 .jot/templates/*.md
```

### Command Failures
```bash
# Test commands individually
date '+%Y-%m-%d'
git branch --show-current

# Use fallbacks in templates
$(git branch --show-current 2>/dev/null || echo "no-branch")
```

### Approval Issues
```bash
# Check approval status
jot template list

# Re-approve after changes
jot template approve template-name

# Clear all approvals (careful!)
rm .jot/template_permissions
```

## See Also

- **[Basic Workflows](basic-workflows.md)** - Using templates in daily workflows
- **[Command Reference](commands.md)** - Template command details
- **[Examples](../examples/templates/)** - Ready-to-use template examples
- **[Configuration](configuration.md)** - Template configuration options

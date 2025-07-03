# Basic Workflows

This guide covers common jot workflows and patterns that will make your note-taking more effective.

## Daily Note Capture

### Quick Thoughts and Ideas

```bash
# Capture a quick thought (opens editor)
jot capture

# Direct capture without editor
jot capture --content "Fix the bug in user authentication"

# Capture from clipboard (if you have xclip/pbpaste)
pbpaste | jot capture    # macOS
xclip -o | jot capture   # Linux
```

### Meeting Notes

```bash
# Capture meeting notes with timestamp
jot capture --content "## Team Standup $(date)"

# Or use a template (see Templates guide)
jot capture --template meeting
```

### Code Snippets and Commands

```bash
# Capture a useful command
jot capture --content "\`\`\`bash
docker-compose up -d
\`\`\`"

# Capture from command history
history | tail -1 | jot capture
```

## Organizing Notes

### Weekly Refiling

```bash
# Review and organize your inbox
jot refile

# Check what needs organizing
jot status
```

The refile interface lets you:
- Move entire notes to different files
- Split notes into multiple parts
- Create new files for new topics
- Add metadata for better organization

### Creating Topic-Based Files

```bash
# Refile creates files automatically, but you can also:
touch lib/work.md
touch lib/personal.md
touch lib/projects/new-project.md

# Then refile to these destinations
jot refile
```

## Finding and Browsing

### Searching Notes

```bash
# Basic text search
jot find "authentication"

# Search with context
jot find "bug fix" --context 2

# Search specific files
jot find "meeting" lib/work.md

# Use with other tools
jot find "TODO" | grep -n "high priority"
```

### Browsing Files

```bash
# List all workspace files
jot files

# Browse file contents without opening
jot peek lib/work.md

# View specific sections
jot peek lib/work.md --section "Project Updates"
```

## Code Integration

### Working with Code Blocks

```bash
# Extract and run code from notes
jot eval lib/scripts.md

# Extract code to files for reuse
jot tangle lib/setup-notes.md

# Append command output to notes
git log --oneline -5 | jot capture --content "Recent commits:"
```

### Project Documentation

```bash
# Capture project status
jot capture --content "## Project Status $(date)
- Feature X: In progress
- Bug fixes: $(git log --oneline --since='1 week ago' | wc -l) commits this week"

# Document setup steps
jot capture --template project-setup
```

## Advanced Patterns

### Multi-Workspace Management

```bash
# Work in different note collections
jot --workspace ~/work-notes capture
jot --workspace ~/personal-notes capture

# Or set up named workspaces in ~/.jotrc
jot workspace add work ~/work-notes
jot workspace add personal ~/personal-notes
```

### Integration with Other Tools

```bash
# Pipe from other tools
git log --oneline -10 | jot capture --content "Recent git activity:"

# Use with fzf for interactive search
jot find "" | fzf --preview 'cat {}'

# Export for other tools
jot find "project-x" --json | jq '.[] | .file'
```

### Automation and Hooks

jot supports hooks for automation:

```bash
# .jot/hooks/post-capture.sh - runs after each capture
#!/bin/bash
echo "Captured note at $(date)" >> .jot/logs/activity.log

# .jot/hooks/post-refile.sh - runs after refiling
#!/bin/bash
git add -A && git commit -m "Updated notes: $(date)"
```

## Common Workflows

### Daily Standup Notes

```bash
# Start of day - capture agenda
jot capture --template standup

# During standup - quick captures
jot capture --content "Action: Review PR #123"
jot capture --content "Blocker: Waiting on API spec"

# End of day - organize
jot refile
```

### Project Research

```bash
# Capture research findings
jot capture --content "## React Performance Tips
- Use React.memo for expensive components
- Source: https://..."

# Organize by project
jot refile  # Move to lib/projects/react-optimization.md
```

### Learning and References

```bash
# Capture learning notes
jot capture --template learning

# Create reference files
jot refile  # Organize into lib/references/

# Find when you need it
jot find "react memo"
```

## Tips for Effective Use

### 1. Capture First, Organize Later
Don't worry about perfect organization during capture. Just get thoughts down quickly.

### 2. Use Descriptive Content
Include context in your notes:
```bash
jot capture --content "## API Rate Limiting Issue
Context: User login endpoint
Error: 429 Too Many Requests
Solution: Implement exponential backoff"
```

### 3. Regular Maintenance
```bash
# Weekly routine
jot status        # Check inbox size
jot refile        # Organize notes
jot doctor        # Check for issues
```

### 4. Consistent File Structure
Develop a consistent way to organize files:
```
lib/
├── daily/          # Daily notes
├── projects/       # Project-specific notes
├── references/     # Long-term reference material
├── meetings/       # Meeting notes
└── learning/       # Study notes and tutorials
```

## Next Steps

- **[Templates](templates.md)** - Create structured notes
- **[Configuration](configuration.md)** - Customize your workflow
- **[Command Reference](commands.md)** - Complete command documentation
- **[Examples](../examples/)** - See real workflow examples

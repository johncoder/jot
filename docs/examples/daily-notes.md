# Daily Notes Workflow

This example shows how to set up and maintain a daily notes system with jot, perfect for developers and knowledge workers who want to track daily progress, meetings, and thoughts.

## Overview

This workflow creates:
- **Daily capture** with timestamps and context
- **Organized monthly files** for easy browsing
- **Project tracking** across days
- **Meeting and standup notes**
- **Quick retrospectives**

## Setup

### 1. Initialize Workspace

```bash
# Create your daily notes workspace
mkdir ~/daily-notes
cd ~/daily-notes
jot init
```

### 2. Create Templates

Create these templates for different daily activities:

**Daily Log Template** (`.jot/templates/daily.md`):
```markdown
---
destination: lib/daily/$(date '+%Y-%m').md
refile_mode: append
tags: [daily, log]
---

## $(date '+%A, %B %d, %Y')

**Weather:** $(curl -s wttr.in?format="%C+%t" 2>/dev/null || echo "Unknown")
**Location:** $(pwd | basename)
**Time Started:** $(date '+%H:%M')

### Today's Focus
- 

### Accomplished
- 

### Challenges
- 

### Notes


### Tomorrow's Plan
- 

---
```

**Standup Template** (`.jot/templates/standup.md`):
```markdown
---
destination: lib/standups/$(date '+%Y-%m').md
refile_mode: append
tags: [standup, daily]
---

### Standup - $(date '+%Y-%m-%d')

**Yesterday:**
- 

**Today:**
- 

**Blockers:**
- 

**Sprint:** $(git branch --show-current 2>/dev/null | sed 's/feature\///' | tr '-' ' ' || echo "N/A")

---
```

**Quick Note Template** (`.jot/templates/note.md`):
```markdown
---
destination: inbox.md
tags: [note, $(date '+%Y-%m-%d')]
---

**$(date '+%H:%M')** - 
```

### 3. Approve Templates

```bash
jot template approve daily
jot template approve standup
jot template approve note
```

### 4. Create Directory Structure

```bash
mkdir -p lib/daily lib/standups lib/meetings lib/projects
```

## Daily Routine

### Morning Startup

```bash
# Start your daily log
jot capture daily

# Quick standup notes
jot capture standup
```

### Throughout the Day

```bash
# Quick thoughts and observations
jot capture note --content "Interesting pattern in user behavior data"

# Meeting notes
jot capture --content "## Team Sync - $(date '+%H:%M')
- Discussed API rate limiting
- Decided on Redis implementation
- Action: Draft proposal by Friday"

# Code discoveries
jot capture --content "## TIL: Docker layer caching
Using COPY package.json before COPY . improves build times
Source: https://docs.docker.com/develop/dev-best-practices/"
```

### End of Day

```bash
# Quick retrospective
jot capture --content "## End of Day - $(date '+%H:%M')
**Completed:**
- Fixed authentication bug
- Reviewed 3 PRs  
- Updated documentation

**Tomorrow:**
- Finish API rate limiting
- Team retrospective at 2pm"

# Organize notes
jot refile
```

## Weekly Organization

### Friday Review

```bash
# Check what needs organizing
jot status

# Organize inbox to appropriate files
jot refile

# Create weekly summary
jot capture --content "## Week of $(date -d 'monday' '+%B %d') - $(date '+%B %d')

**Major Accomplishments:**
- 

**Challenges This Week:**
- 

**Key Learnings:**
- 

**Next Week Focus:**
- "
```

### Monthly Archive

```bash
# Archive last month's daily notes
jot archive

# Create monthly review
jot capture --content "## $(date -d 'last month' '+%B %Y') Review

**Projects Worked On:**
- 

**Skills Developed:**
- 

**Notable Achievements:**
- 

**Areas for Improvement:**
- "
```

## Advanced Patterns

### Project Tracking

Track project progress across daily notes:

```bash
# Tag project work
jot capture --content "## Project Alpha Update
**Progress:** Completed user authentication
**Next:** Implement authorization middleware
**Blockers:** None
#project-alpha"

# Find all project-related notes
jot find "project-alpha"
```

### Integration with Development Workflow

```bash
# Capture git activity
jot capture --content "## Code Changes - $(date)
$(git log --oneline --since='1 day ago')

**Key Changes:**
- 

**Testing Status:**
- "

# Document bug fixes
jot capture --content "## Bug Fix: $(git log -1 --pretty=format:'%s')
**Issue:** 
**Root Cause:** 
**Solution:** 
**Commit:** $(git log -1 --pretty=format:'%h')"
```

### Meeting Integration

```bash
# Pre-meeting preparation
jot capture --content "## Meeting Prep: Team Sync
**Agenda Items:**
- API rate limiting discussion
- Sprint planning
- Code review process

**Questions to Ask:**
- 

**Updates to Share:**
- "

# Post-meeting summary
jot capture --content "## Meeting Summary: Team Sync
**Decisions Made:**
- Use Redis for rate limiting
- Sprint 2-week cycles
- Mandatory PR reviews

**Action Items:**
- [ ] @john: Draft rate limiting proposal (Due: Friday)
- [ ] @sarah: Set up review process (Due: Monday)

**Next Meeting:** $(date -d '+1 week' '+%Y-%m-%d')"
```

## Automation and Hooks

### Auto-commit Daily Notes

Create `.jot/hooks/post-capture.sh`:

```bash
#!/bin/bash
# Auto-commit daily notes to git
cd "$(dirname "$0")/../.."

if git rev-parse --git-dir > /dev/null 2>&1; then
    git add -A
    git commit -m "Daily notes: $(date '+%Y-%m-%d %H:%M')" || true
fi
```

Make it executable:
```bash
chmod +x .jot/hooks/post-capture.sh
```

### Daily Summary Automation

Create a script for automated daily summaries:

```bash
#!/bin/bash
# daily-summary.sh

# Count today's activities
TODAY=$(date '+%Y-%m-%d')
NOTES_COUNT=$(grep -c "$TODAY" inbox.md 2>/dev/null || echo "0")
COMMITS_COUNT=$(git log --oneline --since="midnight" | wc -l 2>/dev/null || echo "0")

jot capture --content "## Daily Summary - $(date)
**Notes Captured:** $NOTES_COUNT
**Commits Made:** $COMMITS_COUNT
**Top Keyword:** $(jot find "$TODAY" | head -5 | grep -oE '\b[A-Za-z]{4,}\b' | sort | uniq -c | sort -nr | head -1 | awk '{print $2}' || echo "N/A")

**Day Rating:** /10
**Key Insight:** 
"
```

## Tips for Success

### 1. Start Simple
Begin with basic daily logs and add complexity over time:

```bash
# Week 1: Just daily logs
jot capture daily

# Week 2: Add standups
jot capture standup

# Week 3: Add project tracking
# Week 4: Add meeting notes
```

### 2. Consistent Timing
Establish regular capture times:
- **9:00 AM**: Daily log and standup
- **Throughout day**: Quick notes as needed
- **5:00 PM**: End-of-day summary

### 3. Use Context
Always include context in your notes:

```bash
# Good: Specific and contextual
jot capture --content "## API Rate Limiting Bug
Context: User login endpoint returning 500
Error: Redis connection timeout after 30s
Solution: Increase connection pool size from 5 to 20"

# Avoid: Vague and context-free
jot capture --content "Fixed bug"
```

### 4. Review and Refine
Regular maintenance keeps the system useful:

```bash
# Weekly review process
jot status                    # Check inbox size
jot find "TODO"              # Review action items
jot find "$(date -d '-7 days' '+%Y-%m')" # Review last week
jot refile                   # Organize notes
```

### 5. Cross-Reference
Link related notes and projects:

```bash
jot capture --content "## User Feedback Analysis
Related to: Project Alpha authentication
Previous notes: $(jot find 'user auth' --json | jq -r '.[0].file')
Key insight: Users prefer social login over email"
```

## Troubleshooting Daily Workflow

### Template Issues
```bash
# Template not rendering
jot template list
jot template approve daily

# Date commands failing
date '+%Y-%m-%d'  # Test manually
```

### Organization Problems
```bash
# Too many notes in inbox
jot status
jot refile

# Can't find old notes
jot find "keyword" lib/daily/
ls lib/daily/
```

### Automation Issues
```bash
# Hooks not running
chmod +x .jot/hooks/*.sh
ls -la .jot/hooks/

# Git integration failing
git status
git config --list
```

## Example Daily Workflow

Here's what a typical day looks like:

```bash
# 9:00 AM - Start day
cd ~/daily-notes
jot capture daily

# 9:15 AM - Standup prep
jot capture standup

# 10:30 AM - Quick insight
jot capture note --content "Discovered React.memo significantly improves render performance"

# 2:00 PM - Meeting notes
jot capture --content "## Architecture Review
- Discussed microservices vs monolith
- Decision: Start with monolith, extract services later
- Next: Draft service boundaries document"

# 5:00 PM - End of day
jot capture --content "## Day Complete - $(date '+%H:%M')
Solid progress on user auth. Tomorrow: finish rate limiting implementation."

# 5:05 PM - Organize
jot refile
```

## File Structure After One Month

```
daily-notes/
├── inbox.md                    # Current capture file
├── lib/
│   ├── daily/
│   │   ├── 2025-01.md         # January daily logs
│   │   └── 2025-02.md         # February daily logs
│   ├── standups/
│   │   ├── 2025-01.md         # January standups
│   │   └── 2025-02.md         # February standups
│   ├── meetings/
│   │   ├── team-syncs.md      # Regular team meetings
│   │   └── architecture.md    # Architecture discussions
│   └── projects/
│       ├── project-alpha.md   # Project-specific notes
│       └── learning.md        # TIL and learning notes
└── .jot/
    ├── templates/
    │   ├── daily.md
    │   ├── standup.md
    │   └── note.md
    └── hooks/
        └── post-capture.sh
```

## See Also

- **[Project Notes Workflow](project-notes.md)** - Managing project-specific notes
- **[Templates Guide](../user-guide/templates.md)** - Creating custom templates
- **[Basic Workflows](../user-guide/basic-workflows.md)** - Additional workflow patterns

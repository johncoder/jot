# Daily Log Template

Use this template for daily work logging and progress tracking.

## Template File

Save as `.jot/templates/daily.md`:

```markdown
---
destination: lib/daily/$(date '+%Y-%m').md
refile_mode: append
tags: [daily, log]
---

## $(date '+%A, %B %d, %Y')

**Time Started:** $(date '+%H:%M')
**Location:** $(pwd | basename)
**Weather:** $(curl -s wttr.in?format="%C+%t" 2>/dev/null || echo "Unknown")

### Today's Focus
- 

### Accomplished
- 

### Challenges
- 

### Notes


### Tomorrow's Plan
- 

### Mood/Energy
**Start of day:** /10
**End of day:** /10

---
```

## Usage

```bash
# Approve template first
jot template approve daily

# Use the template
jot capture daily

# Quick add to daily log
jot capture daily --content "Completed API documentation review"
```

## Customization Options

### Simple Version
```markdown
---
destination: lib/daily.md
tags: [daily]
---

## $(date '+%Y-%m-%d')

### Today
- 

### Notes
- 

---
```

### Developer-Focused Version
```markdown
---
destination: lib/daily/$(date '+%Y-%m').md
refile_mode: append
tags: [daily, dev]
---

## $(date '+%A, %B %d, %Y')

**Branch:** $(git branch --show-current 2>/dev/null || echo "N/A")
**Commits Today:** $(git log --oneline --since="midnight" | wc -l)

### Code Progress
- 

### Bugs Fixed
- 

### Learning
- 

### Blockers
- 

### Tomorrow's Code Goals
- 

---
```

### Team-Focused Version
```markdown
---
destination: lib/team-daily/$(date '+%Y-%m').md
refile_mode: append
tags: [daily, team]
---

## Team Daily - $(date '+%Y-%m-%d')

**Attendees:** 
**Duration:** $(date '+%H:%M') - 

### Team Updates
**John:**
- Yesterday: 
- Today: 
- Blockers: 

**Sarah:**
- Yesterday: 
- Today: 
- Blockers: 

### Team Decisions
- 

### Action Items
- [ ] 

---
```

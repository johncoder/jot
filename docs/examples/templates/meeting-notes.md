# Meeting Notes Template

Professional meeting notes template with agenda, discussion, and action items.

## Template File

Save as `.jot/templates/meeting.md`:

```markdown
---
destination: lib/meetings/$(date '+%Y-%m').md
refile_mode: append
tags: [meeting, $(echo "${MEETING_TYPE:-general}")]
---

# Meeting: ${MEETING_TITLE:-[Meeting Title]} - $(date '+%Y-%m-%d')

**Date:** $(date '+%A, %B %d, %Y')
**Time:** $(date '+%H:%M') - 
**Duration:** 
**Location/Platform:** 

## Attendees
- **Organizer:** 
- **Required:** 
- **Optional:** 
- **Absent:** 

## Agenda
1. 
2. 
3. 

## Discussion

### Topic 1: 


### Topic 2: 


### Topic 3: 


## Decisions Made
- 

## Action Items
- [ ] **[Name]** - [Task] - Due: [Date]
- [ ] **[Name]** - [Task] - Due: [Date]

## Next Steps
- 

## Next Meeting
- **Date:** 
- **Agenda Items:** 
- **Preparation Needed:** 

## Notes & Parking Lot
<!-- Ideas, questions, or topics for future discussion -->


---
```

## Usage

```bash
# Approve template
jot template approve meeting

# Basic usage
jot capture meeting

# With environment variables for context
MEETING_TYPE="standup" MEETING_TITLE="Daily Standup" jot capture meeting

# Quick meeting note
jot capture meeting --content "Discussed API rate limiting approach. Decision: Use Redis with sliding window."
```

## Specialized Meeting Templates

### Standup Meeting
Save as `.jot/templates/standup.md`:

```markdown
---
destination: lib/standups/$(date '+%Y-%m').md
refile_mode: append
tags: [standup, daily]
---

## Daily Standup - $(date '+%Y-%m-%d')

### Team Updates

**John Doe:**
- **Yesterday:** 
- **Today:** 
- **Blockers:** 

**Jane Smith:**
- **Yesterday:** 
- **Today:** 
- **Blockers:** 

### Sprint Progress
- **Sprint Goal:** 
- **Completed Stories:** 
- **In Progress:** 
- **Blocked:** 

### Team Notes
- 

---
```

### Retrospective Meeting
Save as `.jot/templates/retrospective.md`:

```markdown
---
destination: lib/retrospectives.md
refile_mode: append
tags: [retrospective, team]
---

# Sprint Retrospective - $(date '+%Y-%m-%d')

**Sprint:** 
**Duration:** $(date -d '-2 weeks' '+%m/%d') - $(date '+%m/%d')
**Team:** 

## What Went Well ✅
- 

## What Could Be Improved 🔄
- 

## Action Items 📋
- [ ] **[Owner]** - [Action] - Due: [Date]
- [ ] **[Owner]** - [Action] - Due: [Date]

## Blockers Removed 🚧
- 

## Team Metrics
- **Velocity:** 
- **Story Points Completed:** 
- **Bugs Found:** 
- **Team Satisfaction:** /10

## Next Sprint Focus
- 

---
```

### One-on-One Meeting
Save as `.jot/templates/one-on-one.md`:

```markdown
---
destination: lib/one-on-ones/$(echo "${PERSON:-person}").md
refile_mode: append
tags: [one-on-one, management]
---

# One-on-One: ${PERSON:-[Name]} - $(date '+%Y-%m-%d')

**Frequency:** Weekly/Bi-weekly/Monthly
**Duration:** 30 minutes

## Check-in
**Mood/Energy:** /10
**Workload:** Light/Moderate/Heavy
**Satisfaction:** /10

## Current Work Discussion
**Current Projects:**
- 

**Progress & Challenges:**
- 

**Support Needed:**
- 

## Career Development
**Goals Discussion:**
- 

**Skill Development:**
- 

**Growth Opportunities:**
- 

## Feedback
**From Manager:**
- 

**From Team Member:**
- 

## Action Items
- [ ] **Manager** - 
- [ ] **Team Member** - 

## Next Meeting
**Date:** $(date -d '+1 week' '+%Y-%m-%d')
**Topics to Cover:**
- 

---
```

### Client Meeting
Save as `.jot/templates/client-meeting.md`:

```markdown
---
destination: lib/client-meetings/$(echo "${CLIENT:-client}").md
refile_mode: append
tags: [client, external]
---

# Client Meeting: ${CLIENT:-[Client Name]} - $(date '+%Y-%m-%d')

**Client:** ${CLIENT:-[Client Name]}
**Project:** ${PROJECT:-[Project Name]}
**Meeting Type:** Status Update/Requirements/Feedback/Planning

## Attendees
**Our Team:**
- 

**Client Team:**
- 

## Agenda
1. 
2. 
3. 

## Client Updates
- 

## Project Status Update
**Completed:**
- 

**In Progress:**
- 

**Upcoming:**
- 

## Client Feedback
- 

## Requirements Discussion
**New Requirements:**
- 

**Changed Requirements:**
- 

**Questions/Clarifications:**
- 

## Decisions & Agreements
- 

## Action Items
- [ ] **Our Team** - 
- [ ] **Client** - 

## Next Steps
- 

## Follow-up
**Next Meeting:** 
**Deliverables Due:** 
**Communication Plan:** 

---
```

## Advanced Meeting Patterns

### Meeting with Auto-Population
```markdown
---
destination: lib/meetings/$(date '+%Y-%m').md
refile_mode: append
tags: [meeting, $(git branch --show-current 2>/dev/null | cut -d'/' -f1)]
---

# Meeting: $(echo "${MEETING_TITLE:-Project Sync}") - $(date '+%Y-%m-%d')

**Project Context:** $(basename $(pwd))
**Current Sprint:** $(git branch --show-current 2>/dev/null | sed 's/.*sprint-//' | sed 's/-.*//')
**Recent Commits:** $(git log --oneline -3 | wc -l) commits since last meeting

## Auto-Generated Context
**Repository:** $(git remote get-url origin 2>/dev/null | sed 's/.*\///' | sed 's/\.git//')
**Branch:** $(git branch --show-current)
**Last Deployment:** $(git log --grep="deploy" -1 --pretty=format:"%ai" 2>/dev/null || echo "Unknown")

## Discussion Points
1. **Progress Review**
   Recent commits:
$(git log --oneline --since="1 week ago" | head -5 | sed 's/^/   - /')

2. **Current Blockers**
   

3. **Next Priorities**
   

## Technical Decisions
- 

## Action Items
- [ ] 

---
```

### Meeting Minutes with Time Tracking
```markdown
---
destination: lib/meetings/detailed/$(date '+%Y-%m-%d-%H%M').md
tags: [meeting, detailed]
---

# Meeting Minutes: ${MEETING_TITLE} - $(date '+%Y-%m-%d %H:%M')

**Start Time:** $(date '+%H:%M')
**End Time:** <!-- Fill at meeting end -->
**Duration:** <!-- Calculate -->

## Time Log
| Time | Topic | Speaker | Notes |
|------|-------|---------|-------|
| $(date '+%H:%M') | Opening |  |  |
|  |  |  |  |
|  |  |  |  |

## Agenda Items
- [ ] **$(date '+%H:%M')** - Item 1
- [ ] **$(date '+%H:%M')** - Item 2  
- [ ] **$(date '+%H:%M')** - Item 3

## Detailed Notes


## Action Items with Owners
| Action | Owner | Due Date | Status |
|--------|-------|----------|--------|
|  |  |  | Pending |
|  |  |  | Pending |

## Meeting Effectiveness
**Time Management:** Good/Fair/Poor
**Participation:** Good/Fair/Poor  
**Outcomes:** Clear/Unclear
**Follow-up Needed:** Yes/No

---
```

## Usage Tips

### 1. Environment Variables
Set meeting context with environment variables:

```bash
# Set meeting context
export MEETING_TITLE="Sprint Planning"
export MEETING_TYPE="planning"
export CLIENT="Acme Corp"

jot capture meeting
```

### 2. Meeting Series
For recurring meetings, use consistent tagging:

```bash
# Weekly team meetings
jot capture meeting --template standup

# Monthly client check-ins  
CLIENT="acme" jot capture client-meeting
```

### 3. Integration with Calendar
Create calendar integration script:

```bash
#!/bin/bash
# meeting-start.sh
TITLE=$(osascript -e 'tell application "Calendar" to get summary of event 1 of calendar 1')
MEETING_TITLE="$TITLE" jot capture meeting
```

### 4. Post-Meeting Processing
Automate post-meeting tasks:

```bash
#!/bin/bash
# post-meeting.sh
# Extract action items and create reminders
jot find "Action Items" --since "1 hour" | grep "\[ \]" > /tmp/action-items.txt
```

## See Also

- **[Daily Notes Workflow](../daily-notes.md)** - Daily meeting integration
- **[Project Notes](../project-notes.md)** - Project meeting tracking
- **[Templates Guide](../../user-guide/templates.md)** - Template customization

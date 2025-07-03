# Project Notes Workflow

This example demonstrates how to use jot for managing project-specific notes, documentation, and knowledge across different development projects.

## Overview

This workflow enables:
- **Project-specific note capture** and organization
- **Technical documentation** as you build
- **Decision tracking** and architecture notes
- **Cross-project knowledge sharing**
- **Integration with development tools**

## Setup

### 1. Initialize Project Workspace

```bash
# Option 1: Notes in project directory
cd ~/projects/my-app
mkdir docs/notes
cd docs/notes
jot init

# Option 2: Dedicated notes workspace
mkdir ~/project-notes
cd ~/project-notes
jot init
```

### 2. Create Project Templates

**Project Update Template** (`.jot/templates/project-update.md`):
```markdown
---
destination: lib/projects/$(basename $(pwd)).md
refile_mode: append
tags: [project, update, $(basename $(pwd))]
---

## Project Update - $(date '+%Y-%m-%d')

**Project:** $(basename $(pwd))
**Branch:** $(git branch --show-current 2>/dev/null || echo "main")
**Sprint/Milestone:** 

### Completed This Week
- 

### In Progress
- 

### Planned Next Week
- 

### Blockers
- 

### Technical Decisions
- 

### Metrics
- **Files Changed:** $(git diff --name-only HEAD~7 2>/dev/null | wc -l || echo "0")
- **Commits:** $(git rev-list --count HEAD ^HEAD~7 2>/dev/null || echo "0")
- **Tests:** $(find . -name "*test*" -type f 2>/dev/null | wc -l || echo "0")

---
```

**Architecture Decision Template** (`.jot/templates/adr.md`):
```markdown
---
destination: lib/architecture/decisions.md
refile_mode: append
tags: [architecture, decision, $(basename $(pwd))]
---

## ADR-$(date '+%Y%m%d'): [Title]

**Date:** $(date '+%Y-%m-%d')
**Status:** Proposed | Accepted | Deprecated | Superseded
**Context:** $(basename $(pwd))

### Context
<!-- What is the issue that we're seeing that is motivating this decision or change? -->

### Decision
<!-- What is the change that we're proposing and/or doing? -->

### Consequences
<!-- What becomes easier or more difficult to do because of this change? -->

### Alternatives Considered
<!-- What other options did we consider? -->

---
```

**Bug Investigation Template** (`.jot/templates/bug.md`):
```markdown
---
destination: lib/bugs/$(date '+%Y-%m').md
refile_mode: append
tags: [bug, investigation]
---

## Bug Investigation - $(date '+%Y-%m-%d %H:%M')

**Project:** $(basename $(pwd))
**Branch:** $(git branch --show-current 2>/dev/null || echo "main")
**Environment:** 

### Issue Description


### Steps to Reproduce
1. 

### Expected Behavior


### Actual Behavior


### Investigation Notes


### Root Cause


### Solution


### Prevention
<!-- How can we prevent this in the future? -->

---
```

**Learning Template** (`.jot/templates/til.md`):
```markdown
---
destination: lib/learning/$(date '+%Y-%m').md
refile_mode: append
tags: [learning, til, $(basename $(pwd))]
---

## TIL: [Title] - $(date '+%Y-%m-%d')

**Context:** $(basename $(pwd))
**Source:** 

### What I Learned


### Code Example
```
<!-- Code snippet if applicable -->
```

### Why It Matters


### Related Concepts
- 

### Action Items
- [ ] 

---
```

### 3. Approve Templates

```bash
jot template approve project-update
jot template approve adr
jot template approve bug
jot template approve til
```

### 4. Create Directory Structure

```bash
mkdir -p lib/projects lib/architecture lib/bugs lib/learning lib/meetings
```

## Project Development Workflow

### Project Kickoff

```bash
# Document project start
jot capture --content "## Project Kickoff: $(basename $(pwd))
**Start Date:** $(date)
**Team:** 
**Goals:** 
**Timeline:** 
**Tech Stack:** 

### Initial Setup
$(ls -la | head -10)

### Repository Info
**Remote:** $(git remote get-url origin 2>/dev/null || echo "No remote")
**Initial Commit:** $(git log --oneline | tail -1 2>/dev/null || echo "No commits yet")
"
```

### Daily Development

```bash
# Capture technical insights
jot capture til --content "React useCallback optimization reduces unnecessary re-renders"

# Document architectural decisions
jot capture adr

# Track project progress
jot capture project-update

# Bug investigation
jot capture bug
```

### Feature Development

```bash
# Start feature work
jot capture --content "## Feature: User Authentication
**Branch:** $(git branch --show-current)
**Requirements:** 
- JWT token-based auth
- Password reset flow
- Social login integration

**Architecture Notes:**
- Auth service pattern
- Middleware for route protection
- Token refresh mechanism"

# During development
jot capture --content "## Auth Implementation Progress
**Completed:**
- JWT service setup
- Login/register endpoints
- Password hashing with bcrypt

**Current:** Working on password reset flow
**Next:** Social login integration

**Code Review Notes:**
- Consider rate limiting for auth endpoints
- Add comprehensive error handling"
```

### Code Review Process

```bash
# Before code review
jot capture --content "## Pre-Review Checklist: $(git branch --show-current)
**Files Changed:** $(git diff --name-only main | tr '\n' ', ')
**Tests Added:** $(git diff --name-only main | grep test | wc -l)
**Documentation Updated:** 

**Key Changes:**
- 

**Testing Strategy:**
- 

**Questions for Reviewers:**
- "

# After code review
jot capture --content "## Code Review Feedback: $(git branch --show-current)
**Reviewer:** 
**Overall:** Approved | Needs Changes
**Action Items:**
- [ ] 

**Key Learnings:**
- 

**Follow-up Discussion:**
- "
```

## Multi-Project Management

### Cross-Project Setup

```bash
# Configure named workspaces in ~/.jotrc
echo '{
  "workspaces": {
    "project-a": "~/projects/project-a/docs/notes",
    "project-b": "~/projects/project-b/docs/notes", 
    "shared": "~/project-notes"
  }
}' > ~/.jotrc
```

### Project Switching

```bash
# Work in specific project
jot --workspace project-a capture project-update

# Share knowledge across projects
jot --workspace shared capture --content "## Pattern: API Error Handling
Discovered in Project A, applicable to all projects:

\`\`\`javascript
const handleApiError = (error) => {
  if (error.response?.status === 401) {
    // Redirect to login
  } else if (error.response?.status >= 500) {
    // Show user-friendly error
  }
}
\`\`\`

**Projects Using:** A, B
**Benefits:** Consistent error UX"
```

### Knowledge Sharing

```bash
# Document reusable patterns
jot capture --content "## Reusable Pattern: Database Migration Helper
**Origin:** Project Alpha
**Applicable To:** All Node.js projects

\`\`\`javascript
// migrations/helpers.js
const createTimestampFields = (table) => {
  table.timestamp('created_at').defaultTo(knex.fn.now())
  table.timestamp('updated_at').defaultTo(knex.fn.now())
}
\`\`\`

**Usage Pattern:**
\`\`\`javascript
exports.up = (knex) => {
  return knex.schema.createTable('users', (table) => {
    table.increments('id')
    table.string('email').unique()
    createTimestampFields(table)
  })
}
\`\`\`"
```

## Integration with Development Tools

### Git Integration

```bash
# Capture commit context
alias git-commit-note='jot capture --content "## Commit: $(git log -1 --pretty=format:"%h - %s")
**Files:** $(git show --name-only --pretty=format:"" | tr "\n" ", ")
**Context:** 
**Testing:** 
**Notes:** "'

# Document git workflows
jot capture --content "## Git Workflow Documentation
**Branch Strategy:** $(git config --get gitflow.branch.master || echo "main-based")
**Current Branches:** 
$(git branch -a | head -10)

**Release Process:**
1. Feature branch → develop
2. develop → release/x.x.x
3. release → main + tag"
```

### CI/CD Integration

```bash
# Document deployment process
jot capture --content "## Deployment Notes
**Environment:** $(echo ${NODE_ENV:-development})
**Build Command:** $(cat package.json | jq -r '.scripts.build // "npm run build"')
**Deploy Target:** $(cat .github/workflows/*.yml | grep -o 'deploy.*' | head -1 || echo "Manual")

**Last Deployment:**
- **Date:** $(git log --grep="deploy" -1 --pretty=format:"%ai" || echo "Unknown")
- **Commit:** $(git log --grep="deploy" -1 --pretty=format:"%h" || echo "Unknown")
- **Status:** 

**Next Deployment:**
- **Planned:** 
- **Changes:** 
- **Risk Level:** Low | Medium | High"
```

### Issue Tracking

```bash
# Link to external systems
jot capture --content "## Issue Investigation: PROJ-123
**JIRA:** https://company.atlassian.net/browse/PROJ-123
**GitHub Issue:** #456
**Priority:** High

**Problem Statement:**
$(curl -s 'https://api.github.com/repos/user/repo/issues/456' | jq -r '.title' 2>/dev/null || echo "Manual description needed")

**Analysis:**


**Solution Approach:**


**Testing Plan:**
"
```

## Advanced Project Patterns

### Technical Debt Tracking

```bash
# Document technical debt
jot capture --content "## Technical Debt: Legacy Authentication
**Location:** src/auth/legacy.js
**Impact:** High - affects all user operations
**Effort:** Medium - 2-3 days to refactor

**Issues:**
- No password complexity validation
- Hardcoded JWT secret
- No rate limiting
- Missing audit logs

**Proposed Solution:**
1. Extract to auth service
2. Implement proper validation
3. Add rate limiting middleware
4. Centralize configuration

**Business Case:**
- Security risk reduction
- Easier maintenance
- Better user experience

**Timeline:** Sprint 3"
```

### Performance Tracking

```bash
# Document performance investigations
jot capture --content "## Performance Analysis: Dashboard Load Time
**Date:** $(date)
**Environment:** Production
**Tool:** Chrome DevTools

**Current Performance:**
- **Load Time:** 3.2s
- **First Paint:** 1.8s
- **Largest Contentful Paint:** 2.9s

**Bottlenecks Identified:**
1. Large bundle size (2.3MB)
2. Inefficient API queries (N+1 problem)
3. No image optimization

**Optimization Plan:**
- [ ] Code splitting for routes
- [ ] Implement query batching
- [ ] Add image compression
- [ ] Enable gzip compression

**Target Metrics:**
- Load Time: <1.5s
- First Paint: <800ms"
```

### Architecture Evolution

```bash
# Track architectural changes
jot capture --content "## Architecture Evolution: Microservices Migration
**Phase:** 1 of 3
**Timeline:** Q2 2025

**Current State:**
- Monolithic Node.js application
- Single PostgreSQL database
- Deployed on single server

**Target State:**
- User service (authentication, profiles)
- Product service (catalog, inventory)
- Order service (purchases, payments)
- API Gateway (routing, rate limiting)

**Migration Strategy:**
1. **Phase 1:** Extract user service
2. **Phase 2:** Extract product service  
3. **Phase 3:** Extract order service

**Current Progress:**
- [x] User service API design
- [x] Database schema separation
- [ ] Service implementation
- [ ] Integration testing
- [ ] Production deployment

**Challenges:**
- Data consistency across services
- Inter-service communication
- Deployment complexity"
```

## Team Collaboration

### Shared Project Knowledge

```bash
# Create team-sharable notes
jot capture --content "## Team Onboarding: $(basename $(pwd))
**Last Updated:** $(date)

### Quick Start
1. Clone repository: \`git clone $(git remote get-url origin)\`
2. Install dependencies: \`npm install\`
3. Set up environment: \`cp .env.example .env\`
4. Run locally: \`npm run dev\`

### Architecture Overview
- **Frontend:** React + TypeScript
- **Backend:** Node.js + Express
- **Database:** PostgreSQL
- **Deployment:** Docker + Kubernetes

### Key Files
- \`src/app.js\` - Main application entry
- \`src/routes/\` - API route definitions
- \`src/models/\` - Database models
- \`tests/\` - Test suites

### Development Workflow
1. Create feature branch from \`develop\`
2. Make changes + add tests
3. Submit PR to \`develop\`
4. After review, merge to \`develop\`
5. Deploy to staging for testing

### Common Issues
- **Port 3000 in use:** Change PORT in .env
- **Database connection:** Ensure PostgreSQL is running
- **Tests failing:** Run \`npm run test:setup\` first

### Contacts
- **Tech Lead:** @alice
- **DevOps:** @bob  
- **Product:** @charlie"
```

### Code Review Templates

```bash
# Standardize code review process
jot capture --content "## Code Review Checklist
**PR:** #$(git branch --show-current | grep -o '[0-9]*' | head -1 || echo "XXX")
**Reviewer:** 
**Author:** 

### Functional Review
- [ ] Feature works as expected
- [ ] Edge cases handled
- [ ] Error handling appropriate
- [ ] User experience smooth

### Code Quality
- [ ] Code is readable and well-commented
- [ ] Functions are small and focused
- [ ] Variable names are descriptive
- [ ] No code duplication

### Testing
- [ ] Unit tests cover new functionality
- [ ] Integration tests pass
- [ ] Manual testing completed
- [ ] Performance impact assessed

### Security
- [ ] Input validation implemented
- [ ] Authentication/authorization correct
- [ ] No sensitive data exposed
- [ ] SQL injection prevention

### Documentation
- [ ] README updated if needed
- [ ] API documentation current
- [ ] Comments explain complex logic
- [ ] Deployment notes updated

### Feedback
**Strengths:**
- 

**Improvements:**
- 

**Questions:**
- "
```

## File Organization After 3 Months

```
project-notes/
├── inbox.md
├── lib/
│   ├── projects/
│   │   ├── project-alpha.md      # Main project updates
│   │   ├── project-beta.md       # Secondary project
│   │   └── shared-patterns.md    # Cross-project knowledge
│   ├── architecture/
│   │   ├── decisions.md          # ADRs and design choices
│   │   ├── migration-notes.md    # Migration documentation
│   │   └── performance.md        # Performance investigations
│   ├── learning/
│   │   ├── 2025-01.md           # January learnings
│   │   ├── 2025-02.md           # February learnings
│   │   └── patterns.md          # Reusable patterns
│   ├── bugs/
│   │   ├── 2025-01.md           # January bug investigations
│   │   └── resolved.md          # Solved bugs for reference
│   └── meetings/
│       ├── sprint-planning.md    # Sprint planning notes
│       ├── retrospectives.md     # Team retrospectives
│       └── architecture-reviews.md
└── .jot/
    ├── templates/
    │   ├── project-update.md
    │   ├── adr.md
    │   ├── bug.md
    │   └── til.md
    └── config.json
```

## Maintenance and Review

### Weekly Project Review

```bash
# Generate project status
jot capture --content "## Weekly Project Review - $(date '+Week %U, %Y')

### Projects Active
$(jot find "project" --since="1 week" | grep -o "Project: [^,]*" | sort | uniq)

### Key Decisions This Week
$(jot find "ADR-" --since="1 week" | head -5)

### Bugs Resolved
$(jot find "Bug Investigation" --since="1 week" | head -5)

### New Learnings
$(jot find "TIL:" --since="1 week" | head -5)

### Focus Next Week
- 
"
```

### Monthly Architecture Review

```bash
# Summarize architectural evolution
jot capture --content "## Monthly Architecture Review - $(date '+%B %Y')

### Major Decisions Made
$(jot find "ADR-" lib/architecture/ | head -10)

### Technical Debt Status
- **High Priority:** 
- **Medium Priority:** 
- **Low Priority:** 

### Performance Metrics
- **Average Load Time:** 
- **Error Rate:** 
- **User Satisfaction:** 

### Next Month Focus
- 
"
```

## See Also

- **[Daily Notes Workflow](daily-notes.md)** - Personal daily note-taking
- **[Templates Guide](../user-guide/templates.md)** - Creating project templates
- **[Configuration](../user-guide/configuration.md)** - Multi-workspace setup

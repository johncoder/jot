# Jot Refile Implementation Milestone Checkin - June 20, 2025

## Executive Summary

The `jot refile` command represents the most complex and critical workflow in jot, designed to enable sophisticated markdown subtree management inspired by Org-mode. While the foundation and architecture are excellent, **the core path resolution logic is incomplete**, limiting the command to basic end-of-file insertion rather than the sophisticated hierarchical refiling documented in the requirements.

**Current Status: 70% Complete** - Infrastructure solid, core functionality missing.

## ✅ Completed Infrastructure

### Excellent Foundation

- **AST-based parsing** with goldmark library
- **Path selector syntax** (`file.md#path/to/heading`) fully parsed
- **Subtree extraction** correctly identifies and extracts heading + nested content
- **Level transformation** adjusts heading levels for destination hierarchy
- **File operations** robust reading/writing with proper error handling
- **CLI interface** with flags, verbose mode, and help text

### Working Components

```go
// ✅ Path parsing works correctly
type HeadingPath struct {
    File       string   // "inbox.md"
    Segments   []string // ["meeting", "attendees"]
    SkipLevels int      // Number of leading slashes
}

// ✅ Subtree extraction works correctly
type Subtree struct {
    Heading     string // Original heading text
    Level       int    // Original heading level (1-6)
    Content     []byte // Full subtree content (markdown)
    StartOffset int    // Byte position in source
    EndOffset   int    // Byte position in source
}

// ✅ Level transformation works correctly
func TransformSubtreeLevel(subtree *Subtree, newBaseLevel int) []byte
```

## ❌ Critical Missing Implementation

### The Core Problem: Placeholder Path Resolution

**Current Implementation (cmd/refile.go:219-232):**

```go
func resolveDestinationPath(doc ast.Node, content []byte, destPath *markdown.HeadingPath, prepend bool) (*DestinationTarget, error) {
    // For now, implement simple case: create at end of file
    // This is a simplified implementation - full version would search for existing paths

    insertOffset := len(content)
    if insertOffset > 0 && content[insertOffset-1] != '\n' {
        insertOffset-- // Insert before final newline if exists
    }

    targetLevel := len(destPath.Segments) + destPath.SkipLevels

    return &DestinationTarget{
        File:         destPath.File,
        TargetLevel:  targetLevel,
        InsertOffset: insertOffset,
        CreatePath:   destPath.Segments,
        Exists:       false,
    }, nil
}
```

**Impact:** This placeholder completely ignores the destination path and always appends to end of file.

## 🔥 Broken Use Cases

### Use Case 1: Basic Hierarchical Refiling

```bash
jot refile "inbox.md#Team Meeting" --to "work.md#Projects/Frontend"
```

**Expected:** Find "Projects" heading, find "Frontend" under it, insert meeting note as subsection  
**Actual:** ❌ Appends to end of `work.md`, completely ignoring `#Projects/Frontend`

**Files involved:**

```markdown
# inbox.md

## Team Meeting

- Discussed new features
- Assigned tasks

# work.md

## Projects

### Frontend Development

- Current tasks

### Backend API

- Database work
```

### Use Case 2: Contains Matching

```bash
jot refile "inbox.md#meeting" --to "work.md#proj/front"
```

**Expected:** Find heading containing "meeting", move to heading path containing "proj" → "front"  
**Actual:** ❌ Fails to find either source or destination paths

### Use Case 3: Auto-creation of Missing Paths

```bash
jot refile "inbox.md#bug-report" --to "work.md#Issues/Critical/Database"
```

**Expected:** Find "Issues", create "Critical" under it, create "Database" under that  
**Actual:** ❌ Appends to end of file

**File Structure:**

```markdown
# work.md

## Projects

- Current work

## Issues

- Some existing issues
```

**Should create:**

```markdown
## Issues

- Some existing issues

### Critical

#### Database

##### Bug Report

- [refiled content here]
```

### Use Case 4: Destination Analysis (Source-less Mode)

```bash
jot refile --to "work.md#projects/backend/api"
```

**Expected:**

```
Destination analysis for "work.md#projects/backend/api":
✓ File exists: work.md
✓ Partial path exists: Projects
✗ Missing path: Backend > API
Would create: ### Backend (level 3), #### API (level 4)
Ready to receive content at level 5
```

**Actual:** ❌ Basic file check only, no path analysis

### Use Case 5: Ambiguity Resolution

```bash
jot refile "inbox.md#meeting" --to "work.md#tasks"
```

**File with multiple matches:**

```markdown
# inbox.md

## Team Meeting

## Client Meeting

## Meeting Room Booking
```

**Expected:** Clear error listing all matches with line numbers  
**Actual:** ❌ May extract wrong subtree or fail silently

### Use Case 6: Skip Levels (Unusual Document Structure)

```bash
jot refile "messy.md#/section/topic" --to "organized.md#Archive"
```

**Source Structure:**

```markdown
Some intro text without heading

## Section A

Content here

### Topic Details

The content we want to move
```

**Expected:** Skip level 1, match "Section A" (level 2), then "Topic Details" (level 3)  
**Actual:** ❌ Skip level parsing exists but path resolution doesn't use it

## 🛠️ Required Implementation Components

### 1. Smart Path Resolution Engine

```go
// Need to implement
func findMatchingHeading(doc ast.Node, content []byte, segment string, expectedLevel int) ([]*HeadingMatch, error)

type HeadingMatch struct {
    Node     *ast.Heading
    Text     string
    Level    int
    Offset   int
    LineNum  int
}
```

### 2. Hierarchical Path Navigation

```go
// Need to implement
func navigateHeadingPath(doc ast.Node, content []byte, segments []string, skipLevels int) (*PathResolution, error)

type PathResolution struct {
    FoundSegments    []string  // Successfully matched segments
    MissingSegments  []string  // Segments that need creation
    InsertionPoint   int       // Where to insert content
    TargetLevel      int       // Level for new content
}
```

### 3. Auto-creation of Missing Headings

```go
// Need to implement
func createMissingHeadings(segments []string, baseLevel int) []byte
```

### 4. Insertion Point Calculation

```go
// Need to implement
func calculateInsertionPoint(heading *ast.Heading, content []byte, prepend bool) int
```

### 5. Ambiguity Detection

```go
// Need to implement
func detectAmbiguousMatches(matches []*HeadingMatch, segment string) error
```

### 6. Comprehensive Destination Analysis

```go
// Need to implement - enhance existing inspectDestination function
func analyzeDestinationPath(ws *workspace.Workspace, destPath *markdown.HeadingPath) error
```

## 📋 Implementation Priority Matrix

| Component                 | Priority    | Complexity | User Impact                   | Implementation Time |
| ------------------------- | ----------- | ---------- | ----------------------------- | ------------------- |
| **Smart Path Resolution** | 🔥 Critical | High       | Blocks all advanced use cases | 1-2 days            |
| **Contains Matching**     | 🔥 Critical | Medium     | Core UX expectation           | 0.5 days            |
| **Auto-creation Logic**   | 🔥 Critical | Medium     | Essential for workflow        | 0.5 days            |
| **Insertion Point Calc**  | 🔥 Critical | Low        | Correct positioning           | 0.5 days            |
| **Ambiguity Resolution**  | ⚠️ High     | Low        | Error handling UX             | 0.5 days            |
| **Destination Analysis**  | ⚠️ High     | Medium     | Inspection workflow           | 0.5 days            |
| **Skip Levels Support**   | 📝 Medium   | Low        | Edge case handling            | 0.25 days           |

**Total Estimated Effort: 3-4 days for experienced Go developer**

## 🎯 Implementation Strategy

### Phase 1: Core Path Resolution (Day 1-2)

1. Implement `findMatchingHeading()` with contains matching
2. Build `navigateHeadingPath()` for hierarchical traversal
3. Add proper insertion point calculation
4. Replace placeholder `resolveDestinationPath()` function

### Phase 2: Auto-creation & UX (Day 3)

1. Implement missing heading creation logic
2. Add comprehensive destination analysis
3. Improve error messages and ambiguity resolution

### Phase 3: Edge Cases & Polish (Day 4)

1. Handle skip levels properly
2. Add comprehensive test coverage
3. Validate against all documented use cases

## 🧪 Test Cases Needed

### Critical Test Scenarios

1. **Basic path resolution** - `#projects/frontend` finds correct heading
2. **Contains matching** - `#proj/front` matches "Projects" → "Frontend"
3. **Missing path creation** - `#new/section` creates both headings
4. **Ambiguous resolution** - Multiple matches produce clear errors
5. **Level transformation** - Content adjusts to destination hierarchy
6. **Skip levels** - `#/section` handles unusual document structures
7. **Insertion positioning** - Content goes under correct heading, not EOF
8. **Cross-file operations** - Move between different markdown files

### Integration Tests

- Full workflow from command line to file modification
- Error scenarios and edge cases
- Performance with large documents
- Windows/macOS compatibility

## 🚀 Success Criteria

### Functional Requirements

- ✅ All documented use cases work correctly
- ✅ Contains matching finds headings intuitively
- ✅ Auto-creation builds missing hierarchy
- ✅ Clear error messages for ambiguous cases
- ✅ Proper content positioning under target headings

### Quality Requirements

- ✅ Comprehensive test coverage (>90% for refile module)
- ✅ Performance adequate for typical note collections (<1s for most operations)
- ✅ Robust error handling for all edge cases
- ✅ Cross-platform compatibility verified

## 📊 Current vs. Target State

| Feature                  | Current Status | Target Status          | Gap          |
| ------------------------ | -------------- | ---------------------- | ------------ |
| **Path Parsing**         | ✅ Complete    | ✅ Complete            | None         |
| **Subtree Extraction**   | ✅ Complete    | ✅ Complete            | None         |
| **Level Transformation** | ✅ Complete    | ✅ Complete            | None         |
| **Path Resolution**      | ❌ Placeholder | ✅ Smart matching      | **CRITICAL** |
| **Auto-creation**        | ❌ Missing     | ✅ Full hierarchy      | **HIGH**     |
| **Insertion Logic**      | ❌ EOF only    | ✅ Correct positioning | **CRITICAL** |
| **Error Handling**       | ⚠️ Basic       | ✅ Comprehensive       | **MEDIUM**   |
| **Destination Analysis** | ⚠️ Minimal     | ✅ Full inspection     | **MEDIUM**   |

## 🔮 Risk Assessment

### Technical Risks

- **AST traversal complexity** - Goldmark AST navigation can be tricky
- **Edge case explosion** - Many combinations of document structures
- **Performance concerns** - Complex path resolution on large documents

### Mitigation Strategies

- Leverage existing working subtree extraction patterns
- Comprehensive test suite with edge cases
- Profile with realistic document sizes

### Dependencies

- No external dependencies required
- Uses existing goldmark parsing infrastructure
- Builds on proven workspace and file handling code

## 🎉 Expected Outcome

Once implemented, jot refile will transform from a basic note mover into a sophisticated knowledge management tool:

1. **Intuitive workflows** - Tech workers can organize notes naturally
2. **Powerful automation** - Complex reorganization becomes simple
3. **Editor integration** - Seamless workflow with development tools
4. **Foundation for advanced features** - Interactive mode, bulk operations, etc.

The architecture and parsing foundation are excellent. The missing piece is the core path resolution logic that makes refile truly powerful and intuitive - approximately 3-4 days of focused development work to complete.

## Next Steps

1. **Immediate:** Implement core path resolution in `resolveDestinationPath()`
2. **Next:** Add comprehensive test coverage for new functionality
3. **Then:** Validate against all documented use cases
4. **Finally:** Update documentation and examples

This will complete the refile implementation and unlock jot's full potential as a sophisticated note management system.

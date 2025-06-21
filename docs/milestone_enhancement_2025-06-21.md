# Jot CLI Enhancement Milestone - June 21, 2025

## Executive Summary

Building on jot's production-ready foundation, this milestone outlines three strategic enhancements that significantly expand jot's integration capabilities and power user workflows:

1. **Enhanced Capture Templates**: Selector-based destination targeting for sophisticated note organization
2. **Improved Eval Element Ordering**: Repositioned eval directives for better code block readability
3. **Universal JSON Output**: Structured data output across all commands for toolchain integration

**Impact Assessment**: These enhancements transform jot from an excellent standalone tool into a powerful component for integrated knowledge management workflows, enabling seamless automation and external tool integration.

## 🎯 Enhancement Overview

### Enhancement 1: Capture Templates with Destination Selectors

**Current State**: Templates support `destination_file` field for simple file targeting
**Enhancement**: Extend destination mechanism to support full selector syntax

#### Technical Implementation

**Enhanced Template Structure**:

```markdown
---
destination_file: work.md#projects/frontend
refile_mode: append # append (default) | prepend
tags: [meeting, frontend, planning]
---

# Frontend Planning Meeting - $(date '+%Y-%m-%d')

**Attendees:**

-

**Agenda:**

-

## **Action Items:**
```

**Capture Integration Changes**:

- Update template parsing to recognize selector syntax in `destination_file`
- Integrate with existing refile path resolution engine
- Support file-only destinations with configurable append/prepend behavior
- Automatic refile operation during capture completion

**Workflow Example**:

```bash
# Template automatically refiles to specific subtree
jot capture --template meeting

# Content appears under work.md > Projects > Frontend instead of inbox.md
# Uses existing robust path resolution with auto-creation
```

#### Implementation Scope

**Core Changes**:

- `internal/template/template.go`: Enhance destination parsing
- `cmd/capture.go`: Integrate refile operation after template processing
- Template metadata parsing for `refile_mode` configuration
- Error handling for invalid destination selectors

**Integration Points**:

- Reuse existing `refile.ResolveDestination()` function
- Leverage `markdown.HeadingPath` parsing
- Maintain backward compatibility with simple file destinations

### Enhancement 2: Eval Element Ordering Improvement

**Current State**: Eval directives positioned after code blocks
**Enhancement**: Reposition eval elements to precede code blocks for improved readability

#### Current Structure vs Enhanced Structure

**Current Pattern**:

````markdown
```python
print("Hello, world!")
```
````

<eval name="hello" />
```
Hello, world!
```
```

**Enhanced Pattern**:

````markdown
<eval name="hello" />
```python
print("Hello, world!")
````

```
Hello, world!
```

````

#### Technical Implementation

**Parser Updates**:
- Modify `internal/eval/scan.go` to detect eval elements preceding code blocks
- Update AST traversal to associate eval directives with following code blocks
- Maintain backward compatibility with existing documents during transition

**Processing Flow**:
1. Scan document for `<eval .../>` elements
2. Check for immediately following fenced code blocks
3. Associate eval parameters with code block
4. Process execution using existing security and approval workflows

**Migration Strategy**:
- Support both patterns during transition period
- Add warning for deprecated pattern
- Provide migration tool to update existing documents

### Enhancement 3: Universal JSON Output

**Current State**: Human-readable output optimized for terminal usage
**Enhancement**: Add `--json` flag to all commands for structured data output

#### JSON Output Schema Design

**Core Principles**:
- Simple, practical schemas with consistent field types
- Complete information preservation from human-readable output
- No ambiguous field types (avoid union types like `string | object`)
- Machine-parseable success/error status

#### Command-Specific Schemas

**`jot status --json`**:
```json
{
  "workspace": {
    "root": "/path/to/workspace",
    "inbox_path": "/path/to/inbox.md",
    "lib_dir": "/path/to/lib"
  },
  "files": {
    "inbox_notes": 5,
    "lib_files": 12,
    "total_notes": 127
  },
  "health": {
    "status": "healthy",
    "issues": []
  }
}
````

**`jot find --json`**:

```json
{
  "query": "search term",
  "results": [
    {
      "file": "work.md",
      "line": 42,
      "context": "...surrounding text...",
      "match": "highlighted match",
      "relevance_score": 0.95
    }
  ],
  "total_results": 1,
  "execution_time_ms": 15
}
```

**`jot peek --json`**:

```json
{
  "selector": "work.md#projects/frontend",
  "subtree": {
    "heading": "Frontend Development",
    "level": 3,
    "content": "...full subtree content...",
    "nested_headings": 5,
    "line_count": 42
  },
  "metadata": {
    "file_path": "/path/to/work.md",
    "extraction_offset": [1024, 2048]
  }
}
```

**`jot refile --json`** (destination analysis):

```json
{
  "destination": "work.md#projects/backend",
  "analysis": {
    "file_exists": true,
    "path_exists": "partial",
    "existing_segments": ["projects"],
    "missing_segments": ["backend"],
    "target_level": 3,
    "ready_for_content": true
  },
  "operations": {
    "create_headings": ["## Backend"],
    "insert_position": "append"
  }
}
```

**`jot eval --json`**:

```json
{
  "file": "example.md",
  "execution": {
    "mode": "single_block",
    "block_name": "hello_python"
  },
  "results": [
    {
      "block": {
        "name": "hello_python",
        "language": "python",
        "start_line": 15,
        "end_line": 18
      },
      "execution": {
        "status": "success",
        "output": "Hello, world!\n",
        "exit_code": 0,
        "duration_ms": 245
      },
      "approval": {
        "required": true,
        "status": "approved",
        "mode": "hash"
      }
    }
  ]
}
```

#### Implementation Strategy

**Global Flag Implementation**:

- Add `--json` flag to root command definition
- Implement JSON marshaling for each command's data structures
- Replace human-readable output completely when `--json` is specified
- Ensure consistent error handling in JSON format

**Error Response Format**:

```json
{
  "error": {
    "message": "File not found: nonexistent.md",
    "code": "file_not_found",
    "details": {
      "file_path": "nonexistent.md"
    }
  }
}
```

**Success Response Wrapper**:

```json
{
  "success": true,
  "data": {
    // Command-specific response data
  },
  "metadata": {
    "command": "jot status",
    "execution_time_ms": 12,
    "timestamp": "2025-06-21T14:30:45Z"
  }
}
```

## 📋 Implementation Phases

### Phase 1: Enhanced Capture Templates (2-3 days)

**Day 1: Core Template Enhancement**

- [ ] Update `internal/template/template.go` to parse selector syntax
- [ ] Modify template metadata structure to support `refile_mode`
- [ ] Implement destination validation during template loading

**Day 2: Capture Integration**

- [ ] Update `cmd/capture.go` to perform refile after template processing
- [ ] Integrate with existing refile path resolution engine
- [ ] Implement error handling for invalid selectors

**Day 3: Testing & Documentation**

- [ ] Add unit tests for template selector parsing
- [ ] Test integration with complex path scenarios
- [ ] Update template documentation and examples

### Phase 2: Eval Element Ordering (1-2 days)

**Day 1: Parser Updates**

- [ ] Modify `internal/eval/scan.go` for new eval element positioning
- [ ] Update AST traversal logic to associate preceding eval elements
- [ ] Maintain backward compatibility during transition

**Day 2: Testing & Migration**

- [ ] Add tests for both eval element patterns
- [ ] Create migration utility for existing documents
- [ ] Update eval documentation and examples

### Phase 3: Universal JSON Output (3-4 days)

**Day 1: Schema Design & Core Implementation**

- [ ] Define JSON schemas for all commands
- [ ] Implement global `--json` flag handling
- [ ] Create base JSON response structures

**Day 2: Command-Specific Implementation**

- [ ] Implement JSON output for `status`, `find`, `peek`
- [ ] Add JSON support for `refile`, `eval`, `capture`
- [ ] Ensure complete information preservation

**Day 3: Advanced Commands & Error Handling**

- [ ] Add JSON support for `doctor`, `archive`, `template`
- [ ] Implement consistent error response format
- [ ] Add execution metadata to all responses

**Day 4: Testing & Documentation**

- [ ] Comprehensive JSON output testing
- [ ] Integration testing with external tools
- [ ] Documentation with schema examples

## 🚀 Expected Benefits

### For Individual Users

- **Enhanced Capture Workflows**: Direct-to-destination capture eliminates manual refiling steps
- **Improved Code Documentation**: Better eval element readability and organization
- **Power User Automation**: JSON output enables custom script integration

### for Tool Integration

- **Programmatic Access**: Complete jot functionality available to external tools
- **Workflow Automation**: JSON output enables sophisticated automation pipelines
- **Editor Integration**: Enhanced template targeting improves editor plugin capabilities

### For Future Development

- **API Foundation**: JSON schemas provide foundation for future web API
- **Tool Ecosystem**: Enables community development of complementary tools
- **Analytics Capability**: Structured data supports usage analytics and optimization

## 📊 Success Criteria

### Technical Validation

- [ ] All existing tests continue to pass
- [ ] New functionality covered by comprehensive test suite
- [ ] Backward compatibility maintained for existing workflows
- [ ] JSON schemas validate against real-world usage

### User Experience Validation

- [ ] Template-based capture reduces manual refiling operations
- [ ] Eval element ordering improves document readability
- [ ] JSON output enables external tool integration
- [ ] Performance impact minimal for interactive usage

### Integration Validation

- [ ] External tools can consume JSON output effectively
- [ ] Template selectors work with complex document structures
- [ ] Eval element repositioning doesn't break existing documents
- [ ] Error handling provides clear guidance in both modes

## 🔮 Future Integration Opportunities

### Immediate Extensions (Next Milestone)

- **Template Validation**: Validate destination selectors during template creation
- **Bulk Template Operations**: Apply templates to multiple captured items
- **Advanced JSON Filtering**: Support for partial response selection
- **Schema Versioning**: Prepare for future JSON schema evolution

### Long-term Possibilities

- **Web API**: RESTful API using JSON schemas as foundation
- **Real-time Integration**: WebSocket-based live workspace updates
- **Advanced Analytics**: Usage pattern analysis and optimization suggestions
- **Community Tools**: Plugin ecosystem leveraging JSON API

## 💡 Implementation Notes

### Technical Considerations

- **Template Parsing**: Reuse existing markdown frontmatter parsing
- **Refile Integration**: Leverage robust existing path resolution engine
- **JSON Performance**: Lazy marshaling for large datasets
- **Backward Compatibility**: Support transition periods for breaking changes

### Risk Mitigation

- **Incremental Implementation**: Each enhancement independent and testable
- **Comprehensive Testing**: Unit, integration, and user acceptance testing
- **Documentation**: Clear migration guides and examples
- **Community Feedback**: Early preview releases for power users

## 📈 Project Impact Assessment

These enhancements represent a strategic evolution of jot from an excellent standalone tool to a powerful platform component:

**Standalone Tool (Current)**: ⭐⭐⭐⭐⭐ (Excellent for individual note management)
**Platform Component (Enhanced)**: ⭐⭐⭐⭐⭐ (Excellent for integrated workflows and automation)

The universal JSON output, in particular, positions jot as a foundational component for knowledge management ecosystems, while the enhanced templates and improved eval ordering provide immediate user experience improvements.

This milestone maintains jot's core philosophy of pragmatic, fast, and convenient workflows while opening new possibilities for automation and integration.

---

**Timeline**: 6-9 days focused development
**Priority**: Medium-High (builds on solid foundation)
**Risk Level**: Low (incremental enhancements to proven architecture)
**Strategic Value**: High (enables ecosystem development)

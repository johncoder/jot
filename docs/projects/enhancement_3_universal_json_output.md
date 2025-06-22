# Enhancement 3: Universal JSON Output

## Overview

Add `--json` flag support to all jot commands to enable structured data output for programmatic access and tool integration. This enhancement transforms jot from a terminal-focused tool into a powerful component for integrated workflows and automation.

## Requirements Summary

Based on clarifying questions and requirements gathering:

1. **Response Format**: Top-level command data with consistent metadata field
2. **Error Handling**: Consistent wrapper format for both success and error cases
3. **Action Commands**: Detailed operation summaries with structured details
4. **Multi-item Operations**: Array format for all operations (single operations return array of 1)
5. **Output Behavior**: `--json` flag completely replaces human-readable output
6. **Command Coverage**: All commands except `find` (excluded from this enhancement)

## JSON Response Structure

### Success Response Format

```json
{
  "command_specific_data": {
    // Command-specific response fields
  },
  "metadata": {
    "success": true,
    "command": "jot status",
    "execution_time_ms": 12,
    "timestamp": "2025-06-21T14:30:45Z"
  }
}
```

### Error Response Format

```json
{
  "error": {
    "message": "File not found: nonexistent.md",
    "code": "file_not_found",
    "details": {
      "file_path": "nonexistent.md"
    }
  },
  "metadata": {
    "success": false,
    "command": "jot peek nonexistent.md",
    "execution_time_ms": 5,
    "timestamp": "2025-06-21T14:30:45Z"
  }
}
```

### Operations Array Format (for action commands)

```json
{
  "operations": [
    {
      "operation": "capture_note",
      "result": "success",
      "details": {
        "destination": "/path/to/inbox.md",
        "characters_captured": 150,
        "template_used": "default"
      }
    }
  ],
  "summary": {
    "total_operations": 1,
    "successful": 1,
    "failed": 0
  },
  "metadata": { ... }
}
```

## Command-Specific Schemas

### 1. `jot status --json`

**Information Command**: Returns workspace status and health information.

```json
{
  "workspace": {
    "root": "/path/to/workspace",
    "inbox_path": "/path/to/inbox.md",
    "lib_dir": "/path/to/lib",
    "jot_dir": "/path/to/.jot"
  },
  "files": {
    "inbox_notes": 5,
    "lib_files": 12,
    "total_notes": 127
  },
  "health": {
    "status": "healthy",
    "issues": []
  },
  "metadata": {
    "success": true,
    "command": "jot status",
    "execution_time_ms": 8,
    "timestamp": "2025-06-21T14:30:45Z"
  }
}
```

### 2. `jot peek --json`

**Information Command**: Returns content and metadata about files/sections.

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
  "file_info": {
    "file_path": "/path/to/work.md",
    "file_exists": true,
    "last_modified": "2025-06-21T14:25:30Z"
  },
  "extraction": {
    "start_line": 25,
    "end_line": 67,
    "content_offset": [1024, 2048]
  },
  "metadata": {
    "success": true,
    "command": "jot peek work.md#projects/frontend",
    "execution_time_ms": 15,
    "timestamp": "2025-06-21T14:30:45Z"
  }
}
```

### 3. `jot capture --json`

**Action Command**: Returns capture operation details.

```json
{
  "operations": [
    {
      "operation": "capture_note",
      "result": "success",
      "details": {
        "destination": "/path/to/inbox.md",
        "template_used": "meeting",
        "refile_destination": "work.md#meetings",
        "characters_captured": 245,
        "lines_added": 12,
        "capture_method": "template_interactive"
      }
    }
  ],
  "summary": {
    "total_operations": 1,
    "successful": 1,
    "failed": 0
  },
  "metadata": {
    "success": true,
    "command": "jot capture meeting",
    "execution_time_ms": 1250,
    "timestamp": "2025-06-21T14:30:45Z"
  }
}
```

### 4. `jot refile --json`

**Action Command**: Returns refile operation details.

```json
{
  "operations": [
    {
      "operation": "refile_note",
      "result": "success",
      "details": {
        "source_index": 1,
        "source_content": "Meeting notes from standup",
        "destination": "work.md#meetings/daily",
        "destination_created": true,
        "headings_created": ["Daily"],
        "lines_moved": 8,
        "refile_mode": "append"
      }
    }
  ],
  "summary": {
    "total_operations": 1,
    "successful": 1,
    "failed": 0
  },
  "metadata": {
    "success": true,
    "command": "jot refile 1 --dest work.md#meetings/daily",
    "execution_time_ms": 45,
    "timestamp": "2025-06-21T14:30:45Z"
  }
}
```

### 5. `jot eval --json`

**Action Command**: Returns code execution details.

```json
{
  "operations": [
    {
      "operation": "eval_block",
      "result": "success",
      "details": {
        "block_name": "hello_python",
        "language": "python",
        "file": "example.md",
        "start_line": 15,
        "end_line": 18,
        "exit_code": 0,
        "output": "Hello, world!\n",
        "duration_ms": 245,
        "approval_status": "approved",
        "approval_mode": "hash"
      }
    },
    {
      "operation": "eval_block",
      "result": "success",
      "details": {
        "block_name": "math_calc",
        "language": "python",
        "file": "example.md",
        "start_line": 22,
        "end_line": 25,
        "exit_code": 0,
        "output": "2 + 3 = 5\n",
        "duration_ms": 123,
        "approval_status": "approved",
        "approval_mode": "hash"
      }
    }
  ],
  "summary": {
    "total_operations": 2,
    "successful": 2,
    "failed": 0
  },
  "metadata": {
    "success": true,
    "command": "jot eval example.md --all",
    "execution_time_ms": 890,
    "timestamp": "2025-06-21T14:30:45Z"
  }
}
```

### 6. `jot archive --json`

**Action Command**: Returns archiving operation details.

```json
{
  "operations": [
    {
      "operation": "archive_notes",
      "result": "success",
      "details": {
        "source_file": "inbox.md",
        "archive_file": "archive/2025-06.md",
        "archive_created": true,
        "notes_archived": 5,
        "lines_moved": 87,
        "archive_date": "2025-06-21"
      }
    }
  ],
  "summary": {
    "total_operations": 1,
    "successful": 1,
    "failed": 0
  },
  "metadata": {
    "success": true,
    "command": "jot archive",
    "execution_time_ms": 32,
    "timestamp": "2025-06-21T14:30:45Z"
  }
}
```

### 7. `jot doctor --json`

**Information Command**: Returns diagnostic information and health checks.

```json
{
  "diagnosis": {
    "overall_status": "healthy",
    "issues_found": 0,
    "checks_performed": 8
  },
  "workspace_checks": {
    "structure_valid": true,
    "permissions_ok": true,
    "required_files_exist": true
  },
  "tool_checks": {
    "editor_available": true,
    "editor_path": "/usr/bin/vim",
    "shell_available": true,
    "shell_path": "/bin/zsh"
  },
  "file_checks": {
    "inbox_readable": true,
    "inbox_writable": true,
    "lib_dir_accessible": true,
    "jot_dir_accessible": true
  },
  "issues": [],
  "metadata": {
    "success": true,
    "command": "jot doctor",
    "execution_time_ms": 125,
    "timestamp": "2025-06-21T14:30:45Z"
  }
}
```

### 8. `jot template --json`

**Mixed Command**: Returns template information or operation results based on subcommand.

#### Template List (`jot template list --json`)

```json
{
  "templates": [
    {
      "name": "meeting",
      "path": "/path/to/.jot/templates/meeting.md",
      "approved": true,
      "last_modified": "2025-06-20T10:15:30Z",
      "destination": "work.md#meetings",
      "refile_mode": "append"
    },
    {
      "name": "idea",
      "path": "/path/to/.jot/templates/idea.md",
      "approved": false,
      "last_modified": "2025-06-21T09:30:15Z",
      "destination": "personal.md#ideas",
      "refile_mode": "prepend"
    }
  ],
  "summary": {
    "total_templates": 2,
    "approved": 1,
    "unapproved": 1
  },
  "metadata": {
    "success": true,
    "command": "jot template list",
    "execution_time_ms": 18,
    "timestamp": "2025-06-21T14:30:45Z"
  }
}
```

#### Template Operation (`jot template new meeting --json`)

```json
{
  "operations": [
    {
      "operation": "create_template",
      "result": "success",
      "details": {
        "template_name": "meeting",
        "template_path": "/path/to/.jot/templates/meeting.md",
        "content_length": 245,
        "frontmatter_included": true,
        "editor_used": "/usr/bin/vim"
      }
    }
  ],
  "summary": {
    "total_operations": 1,
    "successful": 1,
    "failed": 0
  },
  "metadata": {
    "success": true,
    "command": "jot template new meeting",
    "execution_time_ms": 2150,
    "timestamp": "2025-06-21T14:30:45Z"
  }
}
```

#### Template Render (`jot template render meeting --json`)

**Purpose**: Returns the fully rendered template content for external tools to use in their own editing workflows.

```json
{
  "template": {
    "name": "meeting",
    "path": "/path/to/.jot/templates/meeting.md",
    "approved": true,
    "destination": "work.md#meetings",
    "refile_mode": "append",
    "tags": ["meeting", "work"]
  },
  "rendered_content": "# Meeting Notes - Friday, June 21, 2025\n\n**Date:** 2025-06-21\n**Time:** 14:30 EDT\n**Branch:** feature/json-output\n\n## Attendees\n\n\n## Agenda\n\n\n## Notes\n\n\n## Action Items\n\n",
  "metadata": {
    "success": true,
    "command": "jot template render meeting",
    "execution_time_ms": 145,
    "timestamp": "2025-06-21T14:30:45Z"
  }
}
```

**Use Case**: External tools (editors, IDEs, web interfaces) can:

1. Call `jot template render <name> --json` to get the fully rendered template
2. Present the `rendered_content` to the user for editing in their preferred interface
3. Pass the final edited content to `jot capture --content "edited content"` or via stdin
4. The capture command will automatically handle refiling based on the template's destination configuration

### 9. `jot init --json`

**Action Command**: Returns workspace initialization details.

```json
{
  "operations": [
    {
      "operation": "init_workspace",
      "result": "success",
      "details": {
        "workspace_path": "/path/to/new/workspace",
        "files_created": [
          "inbox.md",
          ".jot/",
          ".jot/templates/",
          "lib/",
          "README.md"
        ],
        "directories_created": 3,
        "workspace_already_exists": false
      }
    }
  ],
  "summary": {
    "total_operations": 1,
    "successful": 1,
    "failed": 0
  },
  "metadata": {
    "success": true,
    "command": "jot init /path/to/new/workspace",
    "execution_time_ms": 85,
    "timestamp": "2025-06-21T14:30:45Z"
  }
}
```

## Implementation Plan

### Phase 1: Core Infrastructure (Day 1)

1. **Add Global `--json` Flag**

   - Update root command to include `--json` flag
   - Add flag inheritance to all subcommands
   - Create JSON output detection mechanism

2. **Base JSON Response Types**

   - Define `JSONResponse` struct with metadata
   - Define `JSONError` struct for error responses
   - Define `JSONOperation` and `JSONOperationSummary` structs
   - Create response builder utilities

3. **Metadata Generation**
   - Command name detection
   - Execution time tracking
   - Timestamp generation (ISO 8601 format)
   - Success/failure status tracking

### Phase 2: Information Commands (Day 2)

1. **`jot status --json`**

   - Workspace information extraction
   - File counting and health checks
   - Convert existing status logic to JSON format

2. **`jot peek --json`**

   - Content extraction and metadata
   - File information and line tracking
   - Subtree analysis for JSON output

3. **`jot doctor --json`**
   - Diagnostic result structure
   - Health check status aggregation
   - Issue reporting in structured format

### Phase 3: Action Commands (Day 3)

1. **`jot capture --json`**

   - Capture operation tracking
   - Template and refile integration details
   - Content metrics and destination info

2. **`jot refile --json`**

   - Multi-note refile operation tracking
   - Source and destination details
   - Path creation and content movement metrics

3. **`jot archive --json`**
   - Archive operation details
   - File creation and content movement tracking
   - Archive organization information

### Phase 4: Advanced Commands (Day 4)

1. **`jot eval --json`**

   - Multi-block execution tracking
   - Code execution results and metrics
   - Approval status and security information

2. **`jot template --json`**

   - Template listing and metadata
   - Template operation tracking (create, edit, approve)
   - Template configuration details
   - **Template rendering for external tools** (`jot template render <name> --json`)
     - Shell command execution and output tracking
     - Fully rendered content output
     - Integration workflow guidance

3. **`jot init --json`**
   - Workspace creation tracking
   - File and directory creation details
   - Initialization status and validation

### Phase 5: Testing & Validation (Day 5)

1. **Unit Tests**

   - JSON schema validation tests
   - Response format consistency tests
   - Error handling validation

2. **Integration Tests**

   - End-to-end JSON output testing
   - Multi-command workflow validation
   - External tool integration testing

3. **Documentation**
   - JSON schema documentation
   - Usage examples for each command
   - Migration guide for automation scripts

## Technical Implementation Details

### JSON Response Builder

```go
type JSONResponse struct {
    Data     interface{} `json:",inline"`
    Metadata JSONMetadata `json:"metadata"`
}

type JSONMetadata struct {
    Success       bool      `json:"success"`
    Command       string    `json:"command"`
    ExecutionTime int64     `json:"execution_time_ms"`
    Timestamp     time.Time `json:"timestamp"`
}

type JSONError struct {
    Message string                 `json:"message"`
    Code    string                 `json:"code"`
    Details map[string]interface{} `json:"details,omitempty"`
}
```

### Global Flag Integration

- Add `--json` to root command's persistent flags
- Check flag status in each command's execution
- Replace all output with JSON when flag is present
- Ensure proper error handling maintains JSON format

### Response Consistency

- All commands follow the same response structure
- Metadata is always present and consistent
- Error responses maintain the same wrapper format
- Operations array format used for all action commands

## Benefits

### For Automation

- **Programmatic Access**: Complete jot functionality available via structured data
- **Script Integration**: JSON output enables sophisticated automation workflows
- **CI/CD Integration**: Structured output perfect for build and deployment pipelines

### for Tool Development

- **Editor Plugins**: Structured data enables rich editor integrations
- **Dashboard Creation**: JSON output enables workspace analytics and visualization
- **API Foundation**: JSON schemas provide foundation for future web API development

### for Power Users

- **Custom Workflows**: JSON output enables complex custom automation
- **Data Analysis**: Structured output enables usage pattern analysis
- **External Tool Integration**: Connect jot with existing productivity tools

## Success Criteria

1. **Functional**: All commands (except find) support `--json` flag with complete functionality
2. **Consistent**: All JSON responses follow the same structural patterns
3. **Complete**: JSON output preserves all information from human-readable output
4. **Reliable**: JSON schemas are well-defined and validated
5. **Performant**: JSON generation adds minimal overhead to command execution

## Future Extensions

1. **Schema Versioning**: Prepare for future JSON schema evolution
2. **Partial Response Filtering**: Support for requesting specific fields only
3. **Streaming Output**: Large dataset streaming for better performance
4. **API Server**: REST API using the same JSON schemas

---

**Status**: Ready for Implementation  
**Estimated Effort**: 5 days focused development  
**Dependencies**: None (builds on existing command infrastructure)  
**Risk Level**: Low (additive feature with no breaking changes)

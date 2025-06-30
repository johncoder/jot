# Jot Tangle Implementation - Org-Mode Inspired Code Block Extraction

## Overview

This document outlines the implementation of enhanced tangle functionality for jot, inspired by Org Babel's code tangling features. The tangle command extracts code blocks from Markdown files into standalone source files, supporting literate programming workflows.

## Goals

- Enable literate programming workflows where documentation and code coexist
- Support modular code composition through noweb-style references
- Maintain consistency with jot's existing eval system syntax
- Provide a foundation for future advanced features
- Keep implementation pragmatic and focused on core use cases

## Design Decisions

### Unified Syntax with Eval System

We will use the existing `<eval>` element syntax for consistency and future integration:

  <eval tangle file="./script.sh" />
  ```bash
  #!/bin/bash
  echo "Hello from tangle!"
  ```

  <eval name="helper-function" />
  ```bash
  function setup_env() {
      export PATH="/usr/local/bin:$PATH"
  }
  ```

### Phase 1 Scope (MVP)

**Included Features:**
- Basic file extraction with `file` attribute
- Multi-file tangling from single markdown file
- Integration with existing eval element parsing
- **Workspace-relative file paths** - Tangle output files are resolved relative to the jot workspace root

**Excluded Features (Future Phases):**
- Noweb references using `<<block-name>>` syntax
- Prefix preservation for proper indentation
- Multiple blocks with same name (conflicts with eval's execution model)
- Function-style references with arguments: `<<block(arg=value)>>`
- Custom separators between concatenated blocks
- Conditional noweb references

**Rationale for Exclusions:**
- Multiple same-name blocks conflict with jot's eval system which expects unique block identification
- Noweb references add significant parsing and resolution complexity
- Advanced features can be added incrementally without breaking existing functionality
- Phase 1 focuses on core file extraction functionality that provides immediate value

## Technical Architecture

### 1. Extend Eval Parser

The existing eval parser in `internal/eval/parser.go` will be extended to recognize tangle-specific attributes:

- `tangle` - Boolean flag indicating this block should be tangled
- `file` - Target file path for tangled output
- `name` - Block identifier for future noweb references

### 2. File Writing System

- Create output directories as needed
- Support multiple files from single markdown source
- **Workspace-relative path resolution** - File paths are resolved relative to the jot workspace root
- Absolute paths are preserved as-is
- Preserve file permissions and shebangs (future enhancement)
- Handle file conflicts appropriately

### 3. Future: Noweb Reference Engine

A future noweb processor will handle:
- Pattern matching for `<<block-name>>` references
- Block name resolution to content
- Prefix preservation for indentation
- Recursive reference expansion

### 4. Future: Block Resolution System

- Build a map of named blocks during AST traversal
- Validate that all noweb references can be resolved
- Detect circular references
- Handle missing references gracefully

## User Interface

### Command Usage

```bash
jot tangle <file>                    # Extract all tangle blocks
jot tangle <file> --dry-run          # Preview what would be tangled
jot tangle <file> --verbose          # Show detailed output
```

### Markdown Syntax Examples

**Basic Tangling:**

  <eval tangle file="hello.py" />
  ```python
  #!/usr/bin/env python3
  print("Hello, World!")
  ```

**Workspace-relative Paths:**

  <eval tangle file="lib/utils.py" />
  ```python
  # This will be written to {workspace}/lib/utils.py
  def helper():
      return "workspace-relative"
  ```

**Absolute Paths:**

  <eval tangle file="/tmp/temp.sh" />
  ```bash
  # This will be written to /tmp/temp.sh (absolute path preserved)
  echo "Absolute path example"
  ```


**Multi-file Output:**

  <eval tangle file="client.py" />
  ```python
  import json
  import requests
  
  class Client:
      def __init__(self):
          pass
  ```

  <eval tangle file="server.py" />
  ```python
  import json
  import requests
  
  class Server:
      def __init__(self):
          pass
  ```


## Implementation Plan

### Phase 1: Core Implementation

1. **Extend Eval Parser** - Add tangle attribute support
2. **Create Tangle Engine** - New module for tangling logic
3. **Update Command Handler** - Replace stub with full implementation
4. **Add Tests** - Comprehensive test coverage
5. **Update Documentation** - User-facing documentation and examples

### Phase 2: Enhancements (Future)

1. **Noweb References** - Basic `<<name>>` references with prefix preservation
2. **Advanced Noweb** - Function-style references with arguments
3. **Conditional Tangling** - Based on file existence or other conditions
4. **Multiple Block Names** - If eval system constraints can be resolved
5. **Integration with Eval** - Combined eval+tangle operations

## File Structure

```
cmd/
  tangle.go              # Command implementation (existing stub)
internal/
  eval/
    parser.go            # Extend for tangle attributes
  tangle/
    engine.go            # Core tangling logic
    writer.go            # File writing utilities
    # Future: noweb.go   # Noweb reference processing
docs/
  projects/
    jot_tangle_rework.md # This document
examples/
  tangle/                # Example markdown files with tangle blocks
```

## Configuration

Future tangle-specific configuration options in `.jotrc`:

```json5
{
  "tangle": {
    "create_directories": true,
    "backup_existing": false,
    "default_permissions": "0644",
    "verbose_output": false
  }
}
```

## Error Handling

- **File Conflicts**: Options for handling existing files
- **Permission Errors**: Clear messaging for file system issues
- **Invalid Syntax**: Helpful error messages for malformed eval elements
- **Future: Missing References**: Clear error messages for unresolved `<<name>>` references
- **Future: Circular References**: Detection and prevention of infinite loops

## Testing Strategy

1. **Unit Tests**: Individual components (parser, noweb processor, writer)
2. **Integration Tests**: End-to-end tangling workflows
3. **Example Files**: Real-world examples in repository
4. **Error Cases**: Comprehensive error condition testing

## Success Criteria

- Users can extract code blocks to files using `<eval tangle file="..." />`
- Multiple output files from single markdown source
- Clear error messages for common issues
- Performance suitable for files with dozens of code blocks
- Foundation laid for future noweb reference features

## Future Considerations

- **Integration with eval command**: Combined evaluation and tangling
- **Template variables**: Dynamic file paths and content
- **Include directives**: Import code from other markdown files
- **Language-specific features**: Syntax highlighting preservation, etc.
- **Version control integration**: Tracking tangled file relationships

## Notes

This implementation prioritizes pragmatism and compatibility with jot's existing architecture. The design allows for future expansion while delivering immediate value for literate programming workflows.

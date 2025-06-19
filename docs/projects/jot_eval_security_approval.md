# Jot Eval: Security and Approval System

## Overview

This document outlines the security and approval system for jot's code evaluation feature. Building on the [standards-compliant proposal](jot_eval_standards_compliant_proposal.md), this system provides practical control over code execution while maintaining usability.

The security model focuses on change detection and user approval rather than complex code analysis or restrictions.

## Design Principles

1. **Security by Default**: Code blocks require explicit approval before execution
2. **Simple Approval**: Hash-based change detection with three clear modes
3. **Transparent Process**: Users always know what code will execute and when
4. **Standards Compliance**: Compatible with the HTML element metadata encoding

## Core Security Model

The security system supports three approval modes:

### 1. Hash-Based Approval (Expire on Change)
- Approved code blocks are tracked using content hashes
- Any change to the code requires re-approval
- Default mode for most use cases
- Best for code that doesn't change frequently

### 2. Prompt on Change
- System detects changes but prompts user each time
- Good for semi-trusted code that changes frequently
- User can approve on-the-fly without permanent storage
- Suitable for development/testing scenarios

### 3. File Path Approval (Always Allow)
- Entire files or paths can be marked as always trusted
- Useful for well-known, trusted documentation
- Bypasses individual block approval
- Use with caution - no change detection

## Permission Storage

Approvals are stored in `.jot/eval_permissions`:

```
# Eval block permissions - SHA256 hashes of approved code blocks
# Format: hash:mode:metadata
a1b2c3d4e5f6...	hash	example.md:hello_python:2025-06-18T10:30:45Z
f7g8h9i0j1k2...	prompt	notes.md:data_analysis:2025-06-18T11:00:00Z

# File path approvals
# Format: path:always:metadata  
/home/user/trusted-docs/*	always	approved_by:user:2025-06-18
```

## Interactive Approval Workflow

When executing unapproved blocks with the standards-compliant syntax:

    ```python
    import requests
    response = requests.get("https://api.github.com")
    print(response.status_code)
    ```
    <eval name="api_request" />

The approval workflow shows:

```bash
$ jot eval example.md --name api_request

Code block 'api_request' requires approval:
────────────────────────────────────────
File: example.md
Block: api_request (line 15)
Language: python
────────────────────────────────────────

Approve this block? [y/N/s/a] 
  y - Yes, approve and execute (expire on change)
  N - No, cancel execution  
  s - Show code content before deciding
  a - Always approve (no change detection)
```

## Configuration

Security settings in `.jotrc` using path-based configuration objects:

```json5
{
  "eval": {
    "security": [
      {
        // Global defaults
        "path": "*",
        "require_approval": true,
        "default_mode": "hash"
      },
      {
        // Development environment
        "path": "./dev-notes/*",
        "require_approval": true,
        "default_mode": "prompt"
      },
      {
        // Trusted documentation
        "path": "./docs/examples/*",
        "require_approval": false,
        "default_mode": "always"
      },
      {
        // High-security production files
        "path": "./production/*",
        "require_approval": true,
        "default_mode": "hash"
      }
    ]
  }
}
```

### Configuration Object Properties

Each security configuration object supports:

- **`path`**: Glob pattern for matching files (required)
- **`require_approval`**: Whether approval is needed (default: true)
- **`default_mode`**: Approval mode - "hash", "prompt", or "always" (default: "hash")

*Note: Additional properties like network and file system controls are deferred to future phases.*

### Path Matching Rules

- **Most specific path wins**: More specific patterns override general ones
- **Glob patterns supported**: `*`, `**`, `?`, `[...]` wildcards
- **Relative paths**: Resolved relative to the `.jotrc` file location
- **Absolute paths**: Used as-is for system-wide rules

Examples:
- `*` - Matches all files (global default)
- `*.md` - All markdown files in current directory
- `**/*.py` - All Python files recursively
- `./trusted/*` - All files in trusted subdirectory
- `/home/user/safe-docs/**` - All files under safe-docs directory
```

## Command Line Interface

### Basic Commands

```bash
# List blocks in a file with approval status
jot eval <file>

# Approve and execute a specific block
jot eval <file> <block_name> --approve --mode <mode>

# Execute a specific block (if approved)
jot eval <file> <block_name>

# Execute all approved blocks in a file
jot eval <file> --all

# List all approved blocks
jot eval --list-approved

# Remove approval for a block
jot eval --revoke <block_name> [file...]
```

### Examples

```bash
# List blocks and their approval status
$ jot eval example.md
Blocks in example.md:
✓ hello_python (line 5) - APPROVED (hash mode)
⚠ api_request (line 15) - NEEDS APPROVAL
⚠ file_ops (line 25) - NEEDS APPROVAL

# Approve and execute a block
$ jot eval example.md api_request --approve --mode hash
Approving and executing 'api_request'...
[execution output]

# Execute all approved blocks
$ jot eval example.md --all
Executing 2 approved blocks...
✓ hello_python completed
⚠ Skipping api_request (needs approval)
```

## Integration with Standards-Compliant Proposal

This security system works seamlessly with the HTML element metadata encoding:

```python
import requests
response = requests.get("https://api.github.com")
print(response.status_code)
```
<eval name="api_test" />

**Note**: The system does not store approval hashes directly in the document (as this would be insecure - users could modify the hash to match previously approved content). Instead, jot maintains its own internal tracking of approved content hashes and performs verification during execution.

## Example Workflows

### Development Workflow
For trusted development environments:

```json5
{
  "eval": {
    "security": [
      {
        "path": "*",
        "require_approval": true,
        "default_mode": "prompt"
      },
      {
        "path": "./dev-notes/*",
        "require_approval": false,
        "default_mode": "always"
      }
    ]
  }
}
```

### Production/Shared Workflow
For shared or production environments:

```json5
{
  "eval": {
    "security": [
      {
        "path": "*",
        "require_approval": true,
        "default_mode": "hash"
      }
    ]
  }
}
```

### Demo/Presentation Workflow
For presentations with known code:

```json5
{
  "eval": {
    "security": [
      {
        "path": "*",
        "require_approval": true,
        "default_mode": "hash"
      },
      {
        "path": "./presentation/*",
        "require_approval": false,
        "default_mode": "always"
      }
    ]
  }
}
```

## Implementation Approach

### Phase 1: Core Approval System
1. **Permission Storage**: Implement `.jot/eval_permissions` file
2. **Hash-Based Tracking**: Content hash generation and comparison
3. **Interactive Approval**: Command-line approval workflow
4. **Hierarchical Configuration**: 
   - Load all `.jotrc` files in the hierarchy (parent directories)
   - Apply rules with closer files taking precedence
   - Support path-based security configuration objects

### Phase 2: Enhanced Workflows
1. **File Path Approval**: Trusted path support
2. **Prompt Mode**: On-demand approval without storage
3. **Bulk Operations**: Approve multiple blocks at once
4. **Status Commands**: List and manage approved blocks

### Configuration Resolution
The system will load `.jotrc` files from the directory hierarchy:
1. Start from the directory containing the file being executed
2. Walk up the directory tree looking for `.jotrc` files
3. Merge configurations with closer files taking precedence
4. Apply the most specific path pattern that matches the target file

### Future Phases (Advanced Security)
The following features are planned for future development but are not part of this initial proposal:

- **Code Analysis**: Static analysis for risk assessment
- **Execution Restrictions**: Sandboxing, network controls, file system limits
- **Audit Logging**: Comprehensive execution logging
- **Command Filtering**: Block specific commands or operations

These advanced features would be designed and implemented as separate proposals once the core approval system is stable.

## Example: Complete Approval Flow

```bash
# List blocks in a file
$ jot eval analysis.md
Blocks in analysis.md:
⚠ data_load (line 5) - NEEDS APPROVAL

# Approve and execute the block
$ jot eval analysis.md data_load --approve --mode hash
Approving code block 'data_load':
────────────────────────────────────────
```python
import pandas as pd
data = pd.read_csv("data.csv")
print(f"Loaded {len(data)} rows")
```
────────────────────────────────────────
Approve with hash-based mode? [y/N]: y
Block 'data_load' approved and executed.
Loaded 1000 rows

# Now subsequent execution works without approval
$ jot eval analysis.md data_load
Loaded 1000 rows

# If code changes, re-approval needed
$ jot eval analysis.md data_load
Code block 'data_load' has changed and requires re-approval.
```

## Benefits

1. **Simple Security**: Three clear modes that cover most use cases
2. **Practical**: Focuses on change detection rather than complex analysis
3. **User-Friendly**: Clear approval workflows that don't impede productivity
4. **Flexible**: Configurable policies for different environments
5. **Transparent**: Users always know what code will execute
6. **Hierarchical Configuration**: Path-based rules with directory hierarchy support

## Conclusion

This simplified security and approval system provides essential protection for code execution while maintaining jot's focus on practical utility. By concentrating on change detection and user approval rather than complex code analysis, the system remains understandable and maintainable.

The three-mode approach (hash-based, prompt, always) combined with hierarchical path-based configuration covers the most common use cases while leaving room for advanced features in future phases.

---

**Status**: Design Complete  
**Created**: June 18, 2025  
**Dependencies**: [jot_eval_standards_compliant_proposal.md](jot_eval_standards_compliant_proposal.md)  
**Supersedes**: Previous complex security proposal  
**Next Steps**: Implementation of Phase 1 core approval system

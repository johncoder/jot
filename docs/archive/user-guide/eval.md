# Code Evaluation with jot eval

The `jot eval` command enables safe execution of code blocks embedded in Markdown files, with a comprehensive security approval system to prevent unauthorized code execution.

## Table of Contents

- [Overview](#overview)
- [Basic Usage](#basic-usage)
- [Code Block Format](#code-block-format)
- [Security Model](#security-model)
- [Approval Modes](#approval-modes)
- [Execution Examples](#execution-examples)
- [Result Integration](#result-integration)
- [Document-Level Approval](#document-level-approval)
- [Managing Approvals](#managing-approvals)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

## Overview

`jot eval` allows you to:

- Execute code blocks in Markdown files with safety controls
- Integrate execution results directly into your notes
- Maintain reproducible computational notebooks
- Support multiple programming languages and interpreters
- Ensure security through explicit approval workflows

**Key Security Features:**

- All code blocks require explicit approval before execution
- Content-based validation prevents unauthorized changes
- Hash-based tracking detects modifications
- Granular approval controls (block-level or document-level)

## Basic Usage

### List Code Blocks

View all evaluable code blocks in a file:

```bash
jot eval notes.md                    # Show blocks with approval status
jot eval --list-approved             # List all approved blocks globally
```

### Execute Code Blocks

```bash
jot eval notes.md block_name         # Execute specific approved block
jot eval notes.md --all              # Execute all approved blocks
```

### Approve Code Blocks

```bash
jot eval notes.md block_name --approve                    # Approve with hash mode
jot eval notes.md block_name --approve --mode always      # Approve with always mode
jot eval notes.md --approve-document --mode hash          # Approve entire document
```

### Revoke Approvals

```bash
jot eval notes.md block_name --revoke                     # Revoke block approval
jot eval notes.md --revoke-document                       # Revoke document approval
```

## Code Block Format

Code blocks use HTML-style `<eval>` elements with configuration parameters:

### Basic Structure

    <eval name="hello_python" />
    ```python
    print("Hello, world!")
    ````

    <eval name="current_date" shell="bash" />
    ```bash
    echo "Current date: $(date)"
    ```

### Required Parameters

| Parameter | Description                                     |
| --------- | ----------------------------------------------- |
| `name`    | Unique identifier for the code block (required) |

### Execution Parameters

| Parameter | Type     | Description                             | Default                |
| --------- | -------- | --------------------------------------- | ---------------------- |
| `shell`   | string   | Shell/interpreter to use                | Inferred from language |
| `timeout` | duration | Execution timeout                       | 30s                    |
| `cwd`     | path     | Working directory                       | Current directory      |
| `env`     | string   | Environment variables (comma-separated) | None                   |
| `args`    | string   | Additional arguments to interpreter     | None                   |

### Result Parameters

| Parameter           | Description                          |
| ------------------- | ------------------------------------ |
| `results="output"`  | Capture stdout/stderr (default)      |
| `results="value"`   | Return function/expression value     |
| `results="code"`    | Wrap results in code block (default) |
| `results="table"`   | Format results as markdown table     |
| `results="raw"`     | Insert results directly as markdown  |
| `results="silent"`  | Execute but don't show results       |
| `results="replace"` | Replace previous results (default)   |
| `results="append"`  | Add after previous results           |

### Complete Example

    <eval name="system_info" shell="bash" timeout="10s" cwd="/tmp" env="LANG=en_US" results="table" />
    ```bash
    echo "Hostname,$(hostname)"
    echo "User,$(whoami)"
    echo "Date,$(date '+%Y-%m-%d')"
    ````

## Security Model

### Approval Requirement

**All eval blocks require explicit approval before execution.** This prevents:

- Accidental execution of potentially dangerous code
- Unauthorized code changes from executing
- Security vulnerabilities from untrusted content

### Content Validation

When you approve a block, jot:

1. **Creates a SHA-256 hash** of the code content
2. **Stores the approval record** in `.jot/eval_permissions`
3. **Validates content on execution** - if code changes, re-approval is required

### Approval Process

```bash
# 1. First execution attempt fails
jot eval notes.md my_script
# Error: code block 'my_script' requires approval

# 2. Review and approve the block
jot eval notes.md my_script --approve
# Shows code content for review
# Prompts for confirmation
# ✓ Block 'my_script' approved with hash mode

# 3. Now execution succeeds
jot eval notes.md my_script
# ✓ Executed block 'my_script' in notes.md
```

## Approval Modes

### Hash Mode (Default)

```bash
jot eval notes.md block_name --approve --mode hash
```

- **Strictest security**: Re-approval required if code changes
- **Content validation**: Uses SHA-256 hash verification
- **Best for**: Production scripts, shared code blocks

### Prompt Mode

```bash
jot eval notes.md block_name --approve --mode prompt
```

- **Interactive approval**: Prompts before each execution
- **Flexible workflow**: No re-approval needed for code changes
- **Best for**: Development, experimental code

### Always Mode

```bash
jot eval notes.md block_name --approve --mode always
```

- **No prompts**: Executes immediately when approved
- **Persistent approval**: Survives code changes
- **Best for**: Trusted environments, personal scripts

**⚠️ Warning**: Always mode bypasses safety checks. Use only with trusted code.

## Execution Examples

### Python Data Analysis

    <eval name="data_analysis" results="table" />
    ```python
    import pandas as pd
    import numpy as np

    # Generate sample data

    data = {
    'metric': ['accuracy', 'precision', 'recall'],
    'value': [0.95, 0.92, 0.89]
    }
    df = pd.DataFrame(data)
    print(df.to_string(index=False))
    ```

### System Monitoring

    <eval name="disk_usage" shell="bash" results="code" />
    ```bash
    echo "Disk Usage Report - $(date)"
    echo "=========================="
    df -h / /home /tmp | grep -v "Filesystem"
    ```

### Git Information

    <eval name="git_status" shell="bash" env="GIT_DIR=.git" />
    ```bash
    echo "Branch: $(git branch --show-current)"
    echo "Commits: $(git rev-list --count HEAD)"
    echo "Status: $(git status --porcelain | wc -l) files changed"
    ````

### Environment Setup

    <eval name="env_check" timeout="5s" results="raw" />
    ```bash
    echo "**Environment Status**"
    echo ""
    echo "- Node.js: $(node --version 2>/dev/null || echo 'Not installed')"
    echo "- Python: $(python3 --version 2>/dev/null || echo 'Not installed')"
    echo "- Docker: $(docker --version 2>/dev/null || echo 'Not installed')"
    ```

## Result Integration

### Automatic Result Insertion

When you execute code blocks, results are automatically inserted after the code block:

**Before execution:**

    <eval name="hello" />
    ```python
    print("Hello, world!")
    ````

    ````

**After execution:**

    <eval name="hello" />
    ```python
    print("Hello, world!")
    ````

    ```
    Hello, world!
    ```

### Result Formatting

#### Code Block Format (Default)

    <eval name="example" results="code" />
    ```python
    print("Result")
    ````

    ```
    Result
    ```

#### Table Format

    <eval name="table_data" results="table" />
    ```python
    print("Name,Age,City")
    print("Alice,30,NYC")
    print("Bob,25,SF")
    ````

    | Name  | Age | City |
    | ----- | --- | ---- |
    | Alice | 30  | NYC  |
    | Bob   | 25  | SF   |

#### Raw Markdown

    <eval name="markdown_gen" results="raw" />
    ```python
    print("## Generated Section")
    print("")
    print("This is **bold** text.")
    print("")
    print("- Item 1")
    print("- Item 2")
    ```
    ```
    ## Generated Section

    This is **bold** text.

    - Item 1
    - Item 2
    ```

### Result Handling Modes

#### Replace Mode (Default)

Each execution replaces previous results:

```bash
jot eval notes.md script    # First run: shows result A
# Edit script to produce different output
jot eval notes.md script    # Second run: replaces A with result B
```

#### Append Mode

Each execution adds to previous results:

    <eval name="log_entry" results="append" />
    ```bash
    echo "$(date): Script executed"
    ````

Multiple executions create a log:

```

2025-01-15 10:00:00: Script executed
2025-01-15 10:05:00: Script executed
2025-01-15 10:10:00: Script executed

```

## Document-Level Approval

For files with many code blocks, you can approve the entire document:

### Approve Entire Document

```bash
jot eval notes.md --approve-document --mode hash
```

This shows all blocks in the document and approves them collectively:

```
Approving entire document 'notes.md' (5 blocks):
────────────────────────────────────────
Block: setup (lines 10-15) bash
Block: analysis (lines 25-35) python
Block: cleanup (lines 45-50) bash
Block: report (lines 60-70) python
Block: summary (lines 80-85) bash
────────────────────────────────────────
Approve entire document with hash mode? [y/N]: y
✓ Document 'notes.md' approved with hash mode (5 blocks).
```

### Document Approval Benefits

- **Bulk approval**: Approve many blocks at once
- **Workflow efficiency**: Faster for computational notebooks
- **Consistent security**: Same approval mode for all blocks

### Document vs. Block Approval

- **Document approval** overrides individual block approvals
- **Revoke document approval** to return to block-level control
- **Always mode at document level** bypasses all block checks

## Managing Approvals

### List All Approvals

```bash
jot eval --list-approved
```

Output:

```
Approved documents:
  ✓ /home/user/notes/analysis.md (hash mode)

Approved individual blocks:
  ✓ /home/user/notes/scripts.md:backup_db (hash mode)
  ✓ /home/user/notes/scripts.md:deploy_app (prompt mode)
  ✓ /home/user/notes/utils.md:system_info (always mode)
```

### Revoke Approvals

```bash
# Revoke individual block
jot eval notes.md block_name --revoke

# Revoke entire document
jot eval notes.md --revoke-document
```

### JSON Output

For scripting and automation:

```bash
jot eval notes.md --json                    # List blocks in JSON
jot eval --list-approved --json             # List approvals in JSON
jot eval notes.md --all --json              # Execute with JSON output
```

Example JSON output:

```json
{
  "operation": "list_blocks",
  "blocks": [
    {
      "name": "hello_python",
      "language": "python",
      "start_line": 5,
      "end_line": 8,
      "is_approved": true,
      "approval_mode": "hash"
    }
  ],
  "summary": {
    "total_blocks": 1,
    "approved_blocks": 1
  }
}
```

## Best Practices

### Security Best Practices

1. **Review before approval**: Always examine code before approving
2. **Use hash mode**: Default to hash mode for production code
3. **Regular audits**: Periodically review approved blocks
4. **Limit always mode**: Use always mode sparingly and only for trusted code
5. **Version control**: Track approval changes in git

### Code Organization

1. **Descriptive names**: Use clear, descriptive block names
2. **Single purpose**: Keep each block focused on one task
3. **Error handling**: Include error handling in your code blocks
4. **Documentation**: Comment complex code blocks
5. **Timeouts**: Set appropriate timeouts for long-running code

### Workflow Integration

1. **Computational notebooks**: Use for data analysis and reporting
2. **System monitoring**: Automate status checks and reports
3. **Environment setup**: Document and automate environment configuration
4. **Build automation**: Include build steps and deployment scripts
5. **Documentation**: Generate dynamic documentation with live examples

### Team Collaboration

1. **Shared approval**: Coordinate approval workflows in team settings
2. **Code review**: Review eval blocks during code review process
3. **Standards**: Establish team standards for block naming and structure
4. **Documentation**: Document approved workflows and patterns

## Troubleshooting

### Common Issues

#### "Code block requires approval"

**Problem**: Block hasn't been approved yet

**Solution**:

```bash
jot eval notes.md block_name --approve
```

#### "Content changed, re-approval needed"

**Problem**: Code content modified after approval (hash mode)

**Solution**:

1. Review the changes
2. Re-approve if changes are safe:

```bash
jot eval notes.md block_name --approve
```

#### "No block named 'xyz' found"

**Problem**: Block name doesn't exist or has typo

**Solution**:

1. Check available blocks:

```bash
jot eval notes.md
```

2. Verify block name spelling
3. Ensure `<eval name="xyz" />` element exists

#### "Timeout exceeded"

**Problem**: Code execution took longer than timeout

**Solution**:

1. Increase timeout:

```markdown
<eval name="slow_script" timeout="60s" />
```

2. Optimize the code
3. Check for infinite loops

#### "Command not found"

**Problem**: Interpreter or shell command not available

**Solution**:

1. Install required interpreter
2. Check PATH environment variable
3. Specify full path to interpreter:

```markdown
<eval name="python_script" shell="/usr/bin/python3" />
```

### Security Diagnostics

Check eval security status:

```bash
jot doctor                              # General health check
jot eval --list-approved                # Review all approvals
find .jot/ -name "eval_permissions" -exec cat {} \;  # Check approval files
```

### Performance Issues

1. **Large output**: Use `results="silent"` for blocks that produce large output
2. **Slow execution**: Set appropriate timeouts and optimize code
3. **Memory usage**: Monitor system resources during execution
4. **Concurrent execution**: Avoid running multiple eval commands simultaneously

### File Permissions

Ensure proper permissions:

```bash
chmod 644 .jot/eval_permissions          # Approval files
chmod 755 .jot/                          # jot directory
```

## See Also

- **[Command Reference](commands.md)** - Complete eval command syntax
- **[Security Model](../architecture/security-model.md)** - Detailed security information
- **[Basic Workflows](basic-workflows.md)** - Integration with other jot commands
- **[Configuration](configuration.md)** - Security configuration options

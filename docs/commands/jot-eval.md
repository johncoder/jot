[Documentation](../README.md) > [Commands](README.md) > eval

# jot eval

## Description

The `jot eval` command evaluates code blocks in markdown files using standards-compliant metadata. It provides a secure way to execute code directly from your notes with configurable parameters and approval workflows.

This command is useful for:
- Creating executable documentation and tutorials
- Running code examples in notes for validation
- Building interactive notebooks within markdown
- Automating tasks documented in your notes

## Usage

```bash
jot eval [file] [block_name] [options]
```

## Arguments

| Argument | Description |
|----------|-------------|
| `[file]` | Markdown file containing eval blocks |
| `[block_name]` | Specific block to execute (optional) |

## Options

| Option | Short | Description |
|--------|-------|-------------|
| `--all` | `-a` | Execute all approved evaluable code blocks |
| `--approve` | | Approve and execute the specified block |
| `--mode` | | Approval mode: `hash`, `prompt`, or `always` (default: `hash`) |
| `--revoke` | | Revoke approval for the specified block |
| `--list-approved` | | List all approved blocks |
| `--approve-document` | | Approve the entire document |
| `--revoke-document` | | Revoke document approval |
| `--no-workspace` | | Resolve file paths relative to current directory |
| `--no-verify` | | Skip hooks verification |

## Eval Element Syntax

Eval elements are HTML-style self-closing tags that precede code blocks:

    <eval name="hello" />
    ```python
    print("Hello, world!")
    ```

### Core Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `name="block_name"` | Unique identifier for the code block | Required |
| `shell="python3"` | Shell/interpreter to use | Inferred from language |
| `timeout="30s"` | Execution timeout | 30s |
| `cwd="/tmp"` | Working directory for execution | Current directory |
| `env="VAR=value"` | Environment variables (comma-separated) | None |
| `args="--verbose"` | Additional arguments to interpreter | None |

### Result Parameters

| Parameter | Description |
|-----------|-------------|
| `results="output"` | Capture stdout/stderr (default) |
| `results="value"` | Return function/expression value |
| `results="code"` | Wrap in code block (default) |
| `results="table"` | Format as markdown table |
| `results="raw"` | Insert directly as markdown |
| `results="replace"` | Replace previous results (default) |
| `results="append"` | Add after previous results |
| `results="silent"` | Execute but don't show results |

## Security Model

All eval blocks require explicit approval before execution. Approval is tied to the block's content hash - changes require re-approval.

### Approval Modes

- **hash**: Approve specific content hash (most secure)
- **prompt**: Prompt for approval on each execution
- **always**: Always execute without prompting (least secure)

## Examples

### Basic Code Execution

Given a markdown file `example.md`:

    <eval name="hello_python" />
    ```python
    print("Hello from Python!")
    print("Current working directory:", os.getcwd())
    ```

We can approve and execute it as follows:

```bash
# List blocks in file
jot eval example.md

# Approve and execute specific block
jot eval example.md hello_python --approve --mode hash

# Execute approved block
jot eval example.md hello_python
```

### Advanced Configuration

Given a markdown file with advanced parameters:

    <eval name="system_check" shell="bash" timeout="10s" env="PATH=/usr/bin" />
    ```bash
    echo "System check at $(date)"
    uname -a
    df -h
    ```

We can approve it with prompt mode:

```bash
# Approve with prompt mode
jot eval example.md system_check --approve --mode prompt
```

### Document-Level Approval

```bash
# Approve entire document
jot eval example.md --approve-document --mode always

# Execute all approved blocks
jot eval example.md --all
```

### Managing Approvals

```bash
# List all approved blocks
jot eval --list-approved

# Revoke specific block approval
jot eval example.md hello_python --revoke

# Revoke document approval
jot eval example.md --revoke-document
```

## Operation Modes

### 1. List Blocks

Show all eval blocks in a file with approval status:

```bash
jot eval example.md
```

Output:
```
Blocks in example.md:

✓ hello_python (approved)
  Shell: python3
  Timeout: 30s
  
⚠ system_check (not approved)
  Shell: bash
  Timeout: 10s
  Environment: PATH=/usr/bin
```

### 2. Execute Specific Block

Execute a named block (requires approval):

```bash
jot eval example.md hello_python
```

### 3. Execute All Blocks

Execute all approved blocks in a file:

```bash
jot eval example.md --all
```

### 4. Approval Management

Approve blocks for execution:

```bash
# Approve with hash verification
jot eval example.md hello_python --approve --mode hash

# Approve with interactive prompting
jot eval example.md system_check --approve --mode prompt
```

## Result Integration

Execution results are automatically inserted into the markdown file immediately following the source block being evaluated:


    <eval name="calculation" />
    ```python
    result = 2 + 2
    print(f"2 + 2 = {result}")
    ```
    
    ```
    2 + 2 = 4
    ```


## JSON Output

### List Blocks JSON

```bash
jot eval example.md --json
```

```json
{
  "operation": "list_blocks",
  "file": "example.md",
  "blocks": [
    {
      "name": "hello_python",
      "shell": "python3",
      "timeout": "30s",
      "approved": true,
      "approval_mode": "hash"
    },
    {
      "name": "system_check",
      "shell": "bash",
      "timeout": "10s",
      "approved": false,
      "environment": ["PATH=/usr/bin"]
    }
  ],
  "metadata": {
    "command": "eval",
    "success": true,
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

### Execution Results JSON

```bash
jot eval example.md hello_python --json
```

```json
{
  "operation": "execute_block",
  "file": "example.md",
  "block_name": "hello_python",
  "results": [
    {
      "block_name": "hello_python",
      "success": true,
      "output": "Hello from Python!",
      "exit_code": 0,
      "execution_time": "0.12s"
    }
  ],
  "metadata": {
    "command": "eval",
    "success": true,
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

## Hook Integration

The eval command integrates with the hook system:

- **Pre-eval hooks**: Execute before code evaluation (can abort)
- **Post-eval hooks**: Execute after successful evaluation (informational)
- **Skip hooks**: Use `--no-verify` to bypass hook execution

## Error Conditions

| Error | Cause | Solution |
|-------|-------|----------|
| `block requires approval` | Block not approved for execution | Use `--approve` with appropriate mode |
| `block not found` | Named block doesn't exist in file | Check block name and file content |
| `execution timeout` | Code execution exceeded timeout | Increase timeout or optimize code |
| `security violation` | Content hash mismatch | Re-approve block after changes |
| `interpreter not found` | Specified shell/interpreter unavailable | Install required interpreter |

## Security Considerations

### Content Verification

- **Hash-based approval**: Ensures exact content match
- **Change detection**: Modified blocks require re-approval
- **Isolation**: Each block runs in controlled environment

### Best Practices

- Use `hash` mode for production documentation
- Regularly audit approved blocks with `--list-approved`
- Use timeouts to prevent runaway processes
- Limit environment variables and working directories

## Use Cases

### Executable Documentation


    # API Testing Guide

    <eval name="test_api" shell="curl" timeout="5s" />
    ```bash
    curl -X GET https://api.example.com/health
    ```


### Code Validation


    # Algorithm Implementation

    <eval name="sorting_test" />
    ```python
    def bubble_sort(arr):
        n = len(arr)
        for i in range(n):
            for j in range(0, n-i-1):
                if arr[j] > arr[j+1]:
                    arr[j], arr[j+1] = arr[j+1], arr[j]
        return arr

    # Test the implementation
    test_array = [64, 34, 25, 12, 22, 11, 90]
    sorted_array = bubble_sort(test_array.copy())
    print(f"Original: {test_array}")
    print(f"Sorted: {sorted_array}")
    ```

### Environment Setup


    # Development Environment

    <eval name="setup_env" shell="bash" cwd="/tmp" />
    ```bash
    echo "Setting up development environment..."
    mkdir -p project-workspace
    cd project-workspace
    git init
    echo "# Project" > README.md
    echo "Environment ready!"
    ```

## Cross-references

- [jot tangle](jot-tangle.md) - Extracting code from markdown
- [jot peek](jot-peek.md) - Previewing code blocks
- [jot capture](jot-capture.md) - Adding eval blocks to notes

## See Also

- [Global Options](README.md#global-options)
- [Security Model](../topics/security.md)
- [Code Integration](../topics/code-integration.md)
- [Hook System](../topics/hooks.md)

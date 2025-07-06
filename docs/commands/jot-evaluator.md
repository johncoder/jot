[Documentation](../README.md) > [Commands](README.md) > evaluator

# jot evaluator

## Description

The `jot evaluator` command manages and runs code evaluators for the jot eval system.

This command serves two main purposes:
- **Run built-in evaluators**: Execute code using jot's built-in language support
- **List available evaluators**: Show which evaluators are available in the system

The evaluator system is the foundation of jot's code execution capabilities, providing an extensible and convention-based approach to running code in different languages.

## Usage

```bash
jot evaluator [language]
```

## Arguments

| Argument | Description |
|----------|-------------|
| `[language]` | Language to run built-in evaluator for (optional) |

If no language is specified, lists all available evaluators.

## Evaluator Discovery System

When `jot eval` encounters a code block, it uses this discovery order:

1. **PATH evaluator**: Look for `jot-eval-<lang>` in PATH (e.g., `jot-eval-haskell`)
2. **Built-in evaluator**: Use `jot evaluator <lang>` for supported languages
3. **Error**: No evaluator found

### Built-in Language Support

jot provides built-in evaluators for these languages:

| Language | Command | Interpreter Used |
|----------|---------|------------------|
| `python3` | `jot evaluator python3` | `python3` |
| `javascript` | `jot evaluator javascript` | `node` |
| `bash` | `jot evaluator bash` | `bash` |
| `go` | `jot evaluator go` | `go run` |

Note: `python` (without version) maps to `python3`, and `sh` maps to `bash`.

## Environment Variables

When running evaluators, jot sets these environment variables:

### Eval-Specific Context
- `JOT_EVAL_CODE` - The actual code to execute (required)
- `JOT_EVAL_LANG` - Language from code fence or shell parameter
- `JOT_EVAL_BLOCK_NAME` - Block name from eval element
- `JOT_EVAL_CWD` - Working directory for execution
- `JOT_EVAL_TIMEOUT` - Timeout duration (e.g., "30s")
- `JOT_EVAL_ARGS` - Additional arguments from eval element

### Custom Environment Variables
- `JOT_EVAL_ENV_*` - Any custom env vars from eval element's `env` parameter

## Examples

### List Available Evaluators

```bash
jot evaluator
```

Output:
```
Built-in evaluators:
  python3      (python3)
  javascript   (node)
  bash         (bash)
  go           (go run)

PATH evaluators:
  haskell      (jot-eval-haskell)
  r            (jot-eval-r)
```

### Run Built-in Evaluator

```bash
# Set environment variables and run python evaluator
export JOT_EVAL_CODE="print('Hello from Python')"
export JOT_EVAL_CWD="/tmp"
jot evaluator python3
```

### Test Evaluator Manually

```bash
# Test if a specific evaluator works
echo 'console.log("Hello World")' | JOT_EVAL_CODE='console.log("Hello World")' jot evaluator javascript
```

## JSON Output

### List Evaluators JSON

```bash
jot evaluator --json
```

```json
{
  "operation": "list_evaluators",
  "evaluators": {
    "built_in": [
      {
        "language": "python3",
        "command": "jot evaluator python3",
        "interpreter": "python3"
      },
      {
        "language": "javascript", 
        "command": "jot evaluator javascript",
        "interpreter": "node"
      }
    ],
    "path": [
      {
        "language": "haskell",
        "command": "jot-eval-haskell",
        "path": "/usr/local/bin/jot-eval-haskell"
      }
    ]
  },
  "metadata": {
    "success": true,
    "command": "jot evaluator",
    "execution_time_ms": 15,
    "timestamp": "2025-07-06T12:30:00Z"
  }
}
```

### Evaluator Execution JSON

```bash
JOT_EVAL_CODE="print('Hello')" jot evaluator python3 --json
```

```json
{
  "operation": "run_evaluator",
  "language": "python3",
  "success": true,
  "output": "Hello\n",
  "metadata": {
    "success": true,
    "command": "jot evaluator python3",
    "execution_time_ms": 120,
    "timestamp": "2025-07-06T12:30:00Z"
  }
}
```

## Creating Custom Evaluators

### PATH Evaluators

Create executable scripts named `jot-eval-<language>` and place them in your PATH:

#### Basic Haskell Evaluator

```bash
#!/bin/bash
# jot-eval-haskell

HASKELL_CMD="${HASKELL_CMD:-ghci}"
cd "$JOT_EVAL_CWD" || exit 1

if [ -n "$JOT_EVAL_TIMEOUT" ]; then
    timeout "$JOT_EVAL_TIMEOUT" echo "$JOT_EVAL_CODE" | "$HASKELL_CMD" $JOT_EVAL_ARGS
else
    echo "$JOT_EVAL_CODE" | "$HASKELL_CMD" $JOT_EVAL_ARGS
fi
```

#### Advanced R Evaluator

```bash
#!/bin/bash
# jot-eval-r

R_CMD="${R_CMD:-Rscript}"
TEMP_FILE=$(mktemp --suffix=.R)
echo "$JOT_EVAL_CODE" > "$TEMP_FILE"

cd "$JOT_EVAL_CWD" || exit 1

if [ -n "$JOT_EVAL_TIMEOUT" ]; then
    timeout "$JOT_EVAL_TIMEOUT" "$R_CMD" --vanilla "$TEMP_FILE" $JOT_EVAL_ARGS
else
    "$R_CMD" --vanilla "$TEMP_FILE" $JOT_EVAL_ARGS
fi

rm -f "$TEMP_FILE"
```

#### Windows Compatibility

```batch
@echo off
REM jot-eval-python.bat

set PYTHON_CMD=%PYTHON_CMD%
if "%PYTHON_CMD%"=="" set PYTHON_CMD=python

cd /d "%JOT_EVAL_CWD%"
echo %JOT_EVAL_CODE% | %PYTHON_CMD% %JOT_EVAL_ARGS%
```

## Integration with jot eval

The evaluator system is designed to work seamlessly with `jot eval`:

```markdown
<eval name="example" />
```python
import os
print(f"Running in: {os.getcwd()}")
print("Hello from Python!")
```

When you run `jot eval file.md`, it will:
1. Look for `jot-eval-python3` in PATH
2. Fall back to `jot evaluator python3`
3. Set all environment variables
4. Execute the code and capture output

## Error Conditions

| Error | Cause | Solution |
|-------|-------|----------|
| `no code provided` | `JOT_EVAL_CODE` environment variable not set | Set the required environment variable |
| `not a built-in evaluator` | Language not supported by built-in evaluators | Create PATH evaluator or use supported language |
| `evaluator failed` | Code execution failed | Check code syntax and interpreter availability |
| `command timed out` | Execution exceeded timeout | Increase timeout or optimize code |

## Error Messages

When no evaluator is found:

```
no evaluator found for 'haskell'

Tried:
  1. PATH evaluator: jot-eval-haskell (not found)
  2. Built-in evaluator: jot evaluator haskell (not available)

To add Haskell support:
  - Create jot-eval-haskell script in PATH
  - Check available evaluators: jot evaluator
```

## See Also

- [`jot eval`](jot-eval.md) - Execute code blocks in markdown files

## Cross-References

- [Code Execution Environment](../user-guide/configuration.md#code-execution-environment)
- [Environment Variables](../user-guide/configuration.md#environment-variable-overrides)

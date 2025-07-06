# jot evaluator

> [Documentation](../README.md) > [Commands](README.md) > jot evaluator

Manage and run code evaluators for the jot eval system.

## Synopsis

```bash
jot evaluator [language]              # Run evaluator for specific language
jot evaluator                         # List available evaluators
```

## Description

The `jot evaluator` command serves two main purposes:

1. **Run built-in evaluators**: Execute code using jot's built-in language support
2. **List available evaluators**: Show which evaluators are available in the system

This command is part of jot's extensible evaluator system, which allows code execution in markdown files through a convention-based approach.

## Evaluator Discovery System

When `jot eval` encounters a code block, it uses this discovery order:

1. **PATH evaluator**: Look for `jot-eval-<lang>` in PATH (e.g., `jot-eval-haskell`)
2. **Built-in evaluator**: Use `jot evaluator <lang>` for supported languages
3. **Error**: No evaluator found

### Built-in Language Support

jot provides built-in evaluators for these languages:

| Language | Command | Interpreter Used |
|----------|---------|------------------|
| `python` | `jot evaluator python` | `python3` |
| `python3` | `jot evaluator python3` | `python3` |
| `javascript` | `jot evaluator javascript` | `node` |
| `bash` | `jot evaluator bash` | `bash` |
| `sh` | `jot evaluator sh` | `bash` |
| `go` | `jot evaluator go` | `go run` |

## Environment Variables

When running evaluators, jot sets these environment variables:

### Standard Context (same as external commands)
- `JOT_WORKSPACE_PATH` - Path to current workspace
- `JOT_WORKSPACE_NAME` - Name of current workspace  
- `JOT_CONFIG_FILE` - Path to configuration file

### Eval-Specific Context
- `JOT_EVAL_CODE` - The actual code to execute
- `JOT_EVAL_LANG` - Language from code fence or shell parameter
- `JOT_EVAL_FILE` - Source markdown file path
- `JOT_EVAL_BLOCK_NAME` - Block name from eval element
- `JOT_EVAL_CWD` - Working directory for execution
- `JOT_EVAL_TIMEOUT` - Timeout duration (e.g., "30s")
- `JOT_EVAL_ARGS` - Additional arguments from eval element

### Custom Environment Variables
- `JOT_EVAL_ENV_*` - Any custom env vars from eval element's `env` parameter

## Usage Examples

### List Available Evaluators

```bash
jot evaluator
```

Output:
```
Built-in evaluators:
  python3     (python3)
  javascript  (node)
  bash        (bash) 
  go          (go run)

PATH evaluators:
  haskell     (jot-eval-haskell)
  r           (jot-eval-r)
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

## Creating Custom Evaluators

### PATH Evaluators

Create executable scripts named `jot-eval-<language>` and place them in your PATH:

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

### Advanced R Evaluator Example

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

### Windows Compatibility

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
1. Look for `jot-eval-python` in PATH
2. Fall back to `jot evaluator python`
3. Set all environment variables
4. Execute the code and capture output

## Error Handling

When no evaluator is found:

```
Error: No evaluator found for 'haskell'

Tried:
  1. PATH evaluator: jot-eval-haskell (not found)
  2. Built-in evaluator: jot evaluator haskell (not available)

To add Haskell support:
  - Create jot-eval-haskell script in PATH
  - Check available evaluators: jot evaluator
```

## Exit Codes

- **0**: Success
- **1**: General error (evaluator failed, code execution error)
- **2**: Usage error (invalid arguments)
- **3**: Language not supported
- **4**: Evaluator not found in PATH
- **124**: Timeout (when using timeout command)

## See Also

- [`jot eval`](jot-eval.md) - Execute code blocks in markdown files
- [Configuration Guide](../user-guide/configuration.md#code-execution-environment) - Code execution environment setup
- [External Commands](jot-external.md) - External command system (similar pattern)

## Cross-References

### Related Commands
- [`jot eval`](jot-eval.md) - Main eval command that uses evaluators
- [`jot external`](jot-external.md) - External command system

### Configuration Topics
- [Code Execution Environment](../user-guide/configuration.md#code-execution-environment)
- [Environment Variables](../user-guide/configuration.md#environment-variable-overrides)

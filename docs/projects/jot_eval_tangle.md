# `eval` and `tangle` Subcommands for jot

## Overview
The `eval` and `tangle` subcommands extend jot's functionality to support Org Babel–like workflows. These features enable users to evaluate code blocks embedded in Markdown files and extract ("tangle") them into standalone files.

### `eval` Subcommand
The `eval` subcommand locates code blocks in Markdown files, executes them, and appends the results inline. It supports both direct execution and session-based execution for persistent interpreters.

#### Usage
```bash
jot eval [flags] <file>
```

#### Supported Header Arguments
- `:name` - Assigns a name to the code block.
- `:session` - Specifies a session for persistent interpreters (e.g., Python, R).
- `:results` - Controls how results are appended (e.g., `replace`, `append`).
- `:exports` - Determines if the block is exported (e.g., `code`, `results`).

#### Example
```markdown
```python :session mysession :results replace
x = 42
x * 2
```
```
After running `jot eval`, the result is appended:
```markdown
```python :session mysession :results replace
x = 42
x * 2
84
```
```

### `tangle` Subcommand
The `tangle` subcommand extracts code blocks with `:tangle` or `:file` header arguments into standalone files. It creates directories as needed.

#### Usage
```bash
jot tangle [flags] <file>
```

#### Supported Header Arguments
- `:tangle` - Specifies the output file path.
- `:file` - Alias for `:tangle`.
- `:mkdirp` - Creates parent directories if they do not exist.

#### Example
```markdown
```bash :tangle scripts/setup.sh
#!/bin/bash
echo "Setting up..."
```
```
After running `jot tangle`, the file `scripts/setup.sh` is created with the code block content.

## Configuration
Both subcommands respect configuration options in `.jotrc` files. Example:
```json5
{
  "eval": {
    "default_session": "default",
    "results_behavior": "append"
  },
  "tangle": {
    "create_directories": true
  }
}
```

## Testing
Automated tests for these features are located in the `tests/` directory. Run them using:
```bash
bash tests/test_eval.sh
bash tests/test_tangle.sh
```

## Known Limitations
- Limited support for inline code execution.
- Advanced result handling (e.g., tables, images) is not yet implemented.

## Future Enhancements
- Inline code execution.
- Improved session management.
- Support for additional languages and result types.

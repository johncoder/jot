# Proposal: Enhanced Code Block Evaluation for jot

## Overview
This proposal outlines the implementation of enhanced code block evaluation features for `jot`, inspired by Org Babel. These features aim to make `jot` a powerful tool for literate programming, reproducible research, and interactive workflows.

---

## Proposed Features

### 1. Code Block Execution
- **Interactive and Programmatic Execution**:
  - Execute code blocks directly from Markdown files.
  - Support both functional mode (`:results value`) and scripting mode (`:results output`).

### 2. Header Arguments
- **Control Behavior**:
  - Use header arguments to control evaluation, tangling, and result handling.
  - Example:
    ```markdown
    ```python :name example :eval yes :results output :tangle yes :file scripts/example.py
    print("Hello, jot!")
    ```
    ```

### 3. Session-Based Evaluation
- **Persistent Sessions**:
  - Allow session-based evaluation for supported languages (e.g., Python, R).
  - Use `:session` to specify session names.

### 4. Arguments and Variables
- **Parameterization**:
  - Pass arguments to code blocks, enabling them to act as functions.
  - Example:
    ```markdown
    ```python :name square :var x=4
    return x * x
    ```
    ```

### 5. Tangling
- **Literate Programming**:
  - Extract code blocks into standalone source files.
  - Support Noweb-style references for combining blocks.

### 6. Results Handling
- **Flexible Output**:
  - Capture results as raw text, tables, or graphics.
  - Append results inline or export them to separate files.

### 7. Inline Code
- **Evaluate Inline**:
  - Support inline code evaluation with syntax like:
    ```markdown
    src_python[:session]{10 * x}
    ```

### 8. Library of Blocks
- **Reusable Code**:
  - Maintain a library of reusable code blocks across files.

### 9. Debugging and Expansion
- **Preview Expanded Code**:
  - Expand code blocks with variable assignments for debugging.

### 10. Reproducible Research
- **Executable Documents**:
  - Enable Markdown files to serve as executable research documents.

---

## Implementation Plan

### 1. Code Block Parsing
- Use `github.com/yuin/goldmark` to parse Markdown files into an AST.
- Traverse the AST to locate fenced code blocks and extract metadata.

### 2. Execution Engine
- Use `os/exec` to execute code blocks in subprocesses.
- Capture stdout, stderr, and exit codes.

### 3. Header Argument Parsing
- Parse header arguments to control evaluation, tangling, and result handling.

### 4. Configuration
- Extend `.jotrc` to support language-specific settings and global defaults.
- Example:
  ```json5
  {
    "languages": {
      "python": {
        "cmd": "/usr/bin/python3.9",
        "extra": "--some-flag"
      },
      "javascript": {
        "cmd": "deno",
        "extra": "--allow-all"
      }
    }
  }
  ```

### 5. Tangling
- Implement a `jot tangle` subcommand to extract code blocks into files.
- Support Noweb-style references for combining blocks.

### 6. Results Handling
- Append results inline or export them to separate files based on header arguments.

### 7. Debugging and Expansion
- Add a `jot expand` subcommand to preview expanded code blocks.

### 8. Documentation and Testing
- Update `docs/` with usage examples and configuration details.
- Write comprehensive tests for parsing, execution, and tangling.

---

## Example Workflow

### Markdown Note
```markdown
# Example Note with Code Blocks

## Python Example
This block demonstrates a Python script that prints a message.

```python :name greet :eval yes :results output :tangle yes :file scripts/greet.py
print("Hello, jot!")
```

## JavaScript Example
This block uses Deno to execute a JavaScript snippet.

```javascript :name js_example :eval yes :results output :tangle yes :file scripts/example.js
console.log("Hello from JavaScript!");
```

## Bash Example
This block runs a simple Bash command.

```bash :name bash_example :eval yes :results output
echo "Hello from Bash!"
```
```

### CLI Commands
1. **Evaluate a Specific Block**:
   ```bash
   jot eval inbox.md --name greet
   ```

2. **View a Code Block**:
   ```bash
   jot view inbox.md --name js_example
   ```

3. **Tangle a Code Block**:
   ```bash
   jot tangle inbox.md --name greet
   ```

4. **Run All Approved Blocks**:
   ```bash
   jot eval inbox.md --all
   ```

---

## Conclusion
This proposal brings `jot` closer to Org Babel's feature set while maintaining its simplicity and pragmatism. By implementing these features, `jot` can become a powerful tool for literate programming and reproducible workflows.

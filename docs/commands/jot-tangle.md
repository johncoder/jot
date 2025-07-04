[Documentation](../README.md) > [Commands](README.md) > tangle

# jot tangle

## Description

The `jot tangle` command extracts code blocks from Markdown files into standalone source files. This enables literate programming workflows where documentation and code coexist in the same files.

This command is useful for:
- Implementing literate programming practices
- Extracting code examples into runnable files
- Maintaining code and documentation in sync
- Building multi-file projects from notebook-style documentation

## Usage

```bash
jot tangle <file> [options]
```

## Arguments

| Argument | Description |
|----------|-------------|
| `<file>` | Markdown file containing tangle blocks |

## Options

| Option | Short | Description |
|--------|-------|-------------|
| `--dry-run` | | Show what would be tangled without writing files |
| `--verbose` | `-v` | Show detailed information about the tangle operation |
| `--no-workspace` | | Resolve file paths relative to current directory |

## Tangle Element Syntax

Tangle elements use the same syntax as eval elements but with `file` parameter:

```markdown
<eval tangle file="path/to/output.py" />
```python
def hello_world():
    print("Hello, World!")

if __name__ == "__main__":
    hello_world()
```
```

### Required Parameters

| Parameter | Description |
|-----------|-------------|
| `file="path"` | Target file path for the extracted code |

### Optional Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `name="block_name"` | Identifier for the code block | None |
| `tangle` | Flag to mark block for tangling | Implied by `file` |

## Examples

### Basic Code Extraction

Create a markdown file with tangle blocks:

```markdown
# Python Project

## Main Module

<eval tangle file="src/main.py" />
```python
#!/usr/bin/env python3
"""Main application module."""

import sys
from utils import helper_function

def main():
    print("Starting application...")
    result = helper_function("Hello, World!")
    print(f"Result: {result}")
    return 0

if __name__ == "__main__":
    sys.exit(main())
```

## Utility Functions

<eval tangle file="src/utils.py" />
```python
"""Utility functions for the application."""

def helper_function(message):
    """Process a message and return formatted output."""
    return f"Processed: {message.upper()}"

def calculate_sum(numbers):
    """Calculate sum of a list of numbers.""" 
    return sum(numbers)
```

## Configuration

<eval tangle file="config.yaml" />
```yaml
app:
  name: "My Application"
  version: "1.0.0"
  debug: false
  
database:
  host: "localhost"
  port: 5432
  name: "myapp_db"
```
```

### Extract All Code

```bash
# Extract all tangle blocks
jot tangle project.md
```

Output:
```
Tangling code blocks in file: project.md
Tangled src/main.py
Tangled src/utils.py  
Tangled config.yaml
```

### Dry Run

```bash
# Preview what would be extracted
jot tangle project.md --dry-run
```

Output:
```
Dry run - analyzing file: project.md
Dry run - would tangle the following files:

src/main.py (1 block, 18 lines)
src/utils.py (1 block, 8 lines)  
config.yaml (1 block, 10 lines)

Total: 3 files, 3 blocks, 36 lines
```

### Verbose Output

```bash
# Show detailed information
jot tangle project.md --verbose
```

Output:
```
Tangling code blocks in file: project.md

Processing tangle block: src/main.py
  Language: python
  Lines: 18
  Creating directory: src/
  Writing file: src/main.py

Processing tangle block: src/utils.py
  Language: python
  Lines: 8
  Directory exists: src/
  Writing file: src/utils.py

Processing tangle block: config.yaml
  Language: yaml
  Lines: 10
  Writing file: config.yaml

Summary:
  Files created: 3
  Directories created: 1
  Total lines: 36
```

## Multiple Blocks per File

Multiple code blocks can target the same file - they will be concatenated:

```markdown
# Configuration File

## Database Settings

<eval tangle file="config.py" />
```python
# Database configuration
DATABASE = {
    'host': 'localhost',
    'port': 5432,
    'name': 'myapp'
}
```

## Logging Settings

<eval tangle file="config.py" />
```python
# Logging configuration
LOGGING = {
    'level': 'INFO',
    'format': '%(asctime)s - %(name)s - %(levelname)s - %(message)s'
}
```
```

Result in `config.py`:
```python
# Database configuration
DATABASE = {
    'host': 'localhost',
    'port': 5432,
    'name': 'myapp'
}
# Logging configuration
LOGGING = {
    'level': 'INFO',
    'format': '%(asctime)s - %(name)s - %(levelname)s - %(message)s'
}
```

## Directory Creation

The tangle command automatically creates directories as needed:

```markdown
<eval tangle file="project/src/core/utils.py" />
```python
# Core utilities
def process_data(data):
    return data.strip().lower()
```
```

This creates the directory structure `project/src/core/` before writing `utils.py`.

## JSON Output

```bash
jot tangle project.md --json
```

```json
{
  "source_file": "project.md",
  "dry_run": false,
  "total_groups": 3,
  "target_files": [
    {
      "target_file": "src/main.py",
      "blocks": [
        {
          "content": "#!/usr/bin/env python3\n...",
          "language": "python"
        }
      ],
      "block_count": 1
    },
    {
      "target_file": "src/utils.py", 
      "blocks": [
        {
          "content": "\"\"\"Utility functions...",
          "language": "python"
        }
      ],
      "block_count": 1
    },
    {
      "target_file": "config.yaml",
      "blocks": [
        {
          "content": "app:\n  name: \"My Application\"...",
          "language": "yaml"
        }
      ],
      "block_count": 1
    }
  ],
  "metadata": {
    "command": "tangle",
    "success": true,
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

## File Path Resolution

### Workspace-relative Paths

By default, file paths are resolved relative to the workspace root:

```markdown
<eval tangle file="lib/myproject/core.py" />
```

Creates: `<workspace>/lib/myproject/core.py`

### Absolute Paths

```markdown
<eval tangle file="/tmp/test.py" />
```

Creates: `/tmp/test.py`

### Current Directory Relative

Use `--no-workspace` to resolve paths relative to current directory:

```bash
jot tangle notes.md --no-workspace
```

## Error Conditions

| Error | Cause | Solution |
|-------|-------|----------|
| `no tangle blocks found` | No blocks with `file` parameter | Add `file` parameter to code blocks |
| `permission denied` | Cannot write to target directory | Check directory permissions |
| `invalid file path` | Malformed file path in `file` parameter | Fix file path syntax |
| `directory creation failed` | Cannot create parent directories | Check parent directory permissions |

## Integration with Eval

Tangle and eval systems work together - blocks can be both evaluated and tangled:

```markdown
<eval name="main_function" tangle file="main.py" />
```python
def main():
    print("This code can be both executed and tangled!")
    return 0
```

```bash
# Execute the code
jot eval notes.md main_function --approve

# Extract to file  
jot tangle notes.md
```

## Use Cases

### Literate Programming

Combine documentation and implementation in single files:

```markdown
# Algorithm Implementation

This section implements the quicksort algorithm with detailed explanation.

<eval tangle file="algorithms/quicksort.py" />
```python
def quicksort(arr):
    """
    Quicksort implementation with in-place partitioning.
    Time complexity: O(n log n) average, O(n²) worst case.
    """
    if len(arr) <= 1:
        return arr
    
    pivot = partition(arr)
    return (quicksort(arr[:pivot]) + 
            [arr[pivot]] + 
            quicksort(arr[pivot + 1:]))
```
```

### Configuration Management

Document configuration with examples:

```markdown
# Application Configuration

The application uses YAML configuration files:

<eval tangle file="config/production.yaml" />
```yaml
environment: production
database:
  host: db.production.com
  ssl: true
logging:
  level: WARNING
```
```

### Tutorial Code

Create runnable examples from tutorials:

```markdown
# API Client Tutorial

Here's a complete API client implementation:

<eval tangle file="examples/api_client.py" />
```python
import requests

class APIClient:
    def __init__(self, base_url, api_key):
        self.base_url = base_url
        self.headers = {"Authorization": f"Bearer {api_key}"}
    
    def get(self, endpoint):
        response = requests.get(f"{self.base_url}{endpoint}", 
                              headers=self.headers)
        return response.json()
```
```

## Cross-references

- [jot eval](jot-eval.md) - Executing code blocks
- [jot capture](jot-capture.md) - Adding tangle blocks to notes
- [jot peek](jot-peek.md) - Previewing code blocks

## See Also

- [Global Options](README.md#global-options)
- [Code Integration](../topics/code-integration.md)
- [Literate Programming](../topics/literate-programming.md)
- [Workspace Structure](../topics/workspace.md)

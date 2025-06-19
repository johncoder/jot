# Jot Eval: Standards-Compliant Code Execution and Result Annotation

## Executive Summary

This proposal outlines a robust, standards-compliant approach for jot's code execution and result annotation feature (`jot eval`). The solution maintains full compatibility with base Markdown and GitHub Flavored Markdown (GFM) while providing extensible metadata encoding for code execution options and result storage.

## Background

The current jot eval implementation has several limitations:
- Results are incorrectly inserted inside code blocks due to offset calculation issues
- Uses non-standard four-backtick formatting for output blocks
- Limited metadata encoding options within standard Markdown constraints
- File truncation bugs during updates

## Core Requirements

1. **Markdown Compatibility**: Full compatibility with base Markdown and GFM parsers
2. **Standards Compliance**: No custom syntax that breaks standard Markdown rendering
3. **Extensibility**: Support for execution options, naming, and metadata
4. **Robustness**: Correct offset calculation and result placement
5. **Graceful Degradation**: Readable in any Markdown viewer, even without jot processing

## Proposed Solution: Link-Based Metadata Encoding

### Overview

Use HTML-style self-closing tags immediately following code blocks to encode execution metadata and options. This approach leverages valid HTML syntax that is compatible with Markdown parsers while providing structured metadata for code execution.

### Syntax

    ```javascript
    console.log("Hello, world!");
    ```
    <eval name="hello" timeout="5s" />
    
    ```python
    import datetime 
    print(f"Current time: {datetime.datetime.now()}")
    ```
    <eval name="timestamp" shell="python3" />
    
    ```bash
    ls -la | head -5
    ```
    <eval shell="bash" cwd="/tmp" />


### Element Format Specification

**Basic Format**: `<eval key1="value1" key2="value2" />`

#### Core Parameters

| Parameter | Type | Description | Example |
|-----------|------|-------------|---------|
| `name` | string | Unique identifier for the code block | `name="data_analysis"` |
| `shell` | string | Shell/interpreter to use (default: inferred from language) | `shell="python3"` |
| `timeout` | duration | Execution timeout | `timeout="30s"` |
| `cwd` | path | Working directory for execution | `cwd="/tmp"` |
| `env` | key=value pairs | Environment variables (comma-separated) | `env="PATH=/usr/bin,DEBUG=1"` |
| `args` | string | Additional arguments to pass to interpreter | `args="--verbose"` |

#### Results Parameters (inspired by org-mode)

**Collection** (how to get results):

| Parameter | Description | Use Case |
|-----------|-------------|----------|
| `results="output"` | Capture stdout/stderr (default, scripting mode) | Shell scripts, system commands |
| `results="value"` | Return function/expression value (functional mode) | Mathematical expressions, data analysis |

**Type** (what kind of result):

| Parameter | Description | Output Format | Example Use |
|-----------|-------------|---------------|-------------|
| `results="code"` | Wrap in code block (default) | ````text` | General purpose output |
| `results="table"` | Format as markdown table | `| col1 | col2 |` | Structured data, CSV output |
| `results="list"` | Format as markdown list | `- item1` | Lists, arrays |
| `results="raw"` | Insert directly as markdown | Raw markdown | Pre-formatted markdown |
| `results="file"` | Save to file and insert link | `![](path)` or `[link](path)` | Images, documents |
| `results="html"` | HTML export block | ```html` | HTML content |
| `results="verbatim"` | Plain text, no formatting | Plain text | Log output, simple text |

**Handling** (what to do with results):

| Parameter | Description | Behavior |
|-----------|-------------|----------|
| `results="replace"` | Replace previous results (default) | Overwrites existing output |
| `results="append"` | Add after previous results | Accumulates output |
| `results="prepend"` | Add before previous results | Adds to top |
| `results="silent"` | Execute but don't show results | Hidden execution |
| `results="none"` | Execute but don't store results | No output stored |

#### Advanced Parameters

| Parameter | Type | Description | Example |
|-----------|------|-------------|---------|
| `session` | string | Named persistent execution environment | `session="analysis"` |
| `var` | assignments | Pass variables from other blocks | `var="data=load_data"` |
| `file` | path | Output filename | `file="/tmp/result.png"` |
| `cache` | yes/no | Cache results to avoid re-execution | `cache="yes"` |
| `depends` | names | Dependencies on other named blocks | `depends="setup,load_data"` |
| `post` | block name | Post-processing block for result transformation | `post="format_output"` |

**Parameter Encoding**: Standard HTML attribute format with quoted values (e.g., `name="my script"`).

### Execution Rules

#### Code Block Execution Requirements

| Scenario | Behavior | Example |
|----------|----------|---------|
| **Fenced code block + eval element** | ✅ **Executable** - Code block will be executed with given parameters | ```` ```python`<br/>`code`<br/>```` `<br/>`<eval name="test" />` |
| **Fenced code block alone** | ❌ **Not executable** - Regular markdown code block, ignored by jot eval | ```` ```python`<br/>`code`<br/>```` ` |
| **Eval element alone** | ❌ **Ignored** - No preceding code block to execute | `<eval name="orphan" />` |
| **Eval element + following code block** | ❌ **Ignored** - Must have preceding code block to be valid | `<eval name="wrong" />`<br/>```` ```output`<br/>`result`<br/>```` ` |

#### Output Block Management

**Result Block Reuse**: If a fenced code block immediately follows the eval element, it will be **reused for output**:

    ```python
    print("Hello")
    ```
    <eval name="greeting" />
    ```  
    # This existing block will be replaced with new output
    Old output here
    ```

**Result Block Creation**: If no fenced code block follows the eval element, jot will **create a new one**:

    ```python
    print("Hello")
    ```
    <eval name="greeting" />
    # New code block will be created here after execution

#### Processing Flow

1. **Scan document** for fenced code blocks followed by `<eval .../>` elements
2. **Parse eval parameters** from the HTML attributes
3. **Execute code block** using specified parameters
4. **Update or create result block** immediately after the eval element
5. **Skip orphaned eval elements** that don't have preceding code blocks

### Result Storage: Adjacent Output Blocks

**Chosen Approach**: Adjacent output blocks for visible, standards-compliant result display.

**Convention**: 
1. Fenced code block (source code) - **required for execution**
2. HTML eval element (execution metadata) - **required for execution**
3. Optional fenced code block (execution output/results) - **reused or created**

**Key Behavior**: Only fenced code blocks with immediately following eval elements are executable. The eval element governs whether an existing result block is reused or a new one is created.

Example:

    ```javascript
    console.log("Hello, world!");
    ```
    <eval name="hello" />
    ```
    Hello, world!
    ```

    ```python
    import datetime
    print(f"Current time: {datetime.datetime.now()}")
    ```
    <eval name="timestamp" shell="python3" />
    ```
    Current time: 2025-06-17 10:30:45.123456
    ```

    ```bash
    ls -la | head -5
    ```
    <eval shell="bash" cwd="/tmp" />
    ```
    total 8
    drwxrwxrwt  2 root root 4096 Jun 17 10:30 .
    drwxr-xr-x 20 root root 4096 Jun 17 10:29 ..
    -rw-r--r--  1 user user   42 Jun 17 10:30 example.txt
    ```

**Benefits of This Approach**:
- **Visible Results**: Output is immediately visible in any Markdown viewer
- **Clear Association**: Results directly follow their source code blocks
- **Standards Compliant**: Uses valid HTML syntax within Markdown
- **No Special Formatting**: Output blocks are plain fenced code blocks
- **Flexible**: Results can be omitted if not yet executed
- **Clean**: HTML elements are rendered invisibly in most Markdown viewers

## Implementation Plan

| Phase | Focus | Key Components | Dependencies |
|-------|-------|----------------|--------------|
| **Phase 1: Core Infrastructure** | Basic parsing and results | Link parser, parameter decoder, results system, offset calculator | None |
| **Phase 2: Basic Execution** | Simple code execution | Shell detection, environment setup, process management, output capture | Phase 1 |
| **Phase 3: Advanced Results** | All result types and formatting | Result formatting, result handling, file output, error handling | Phase 2 |
| **Phase 4: Sessions & Dependencies** | Complex workflow features | Session management, variable passing, dependency tracking, caching, post-processing | Phase 3 |

### Phase 1: Core Infrastructure
1. **HTML Parser**: Parse `<eval .../>` elements following code blocks
2. **Attribute Decoder**: Parse HTML-style attributes with quoted values
3. **Results System**: Implement org-mode compatible results collection, type, and handling
4. **Offset Calculator**: Fix current offset calculation bugs using AST node end positions

### Phase 2: Basic Execution Engine
1. **Shell Detection**: Infer shell/interpreter from language and parameters
2. **Environment Setup**: Configure working directory, environment variables, timeouts
3. **Process Management**: Execute code with proper isolation and resource limits
4. **Output Capture**: Capture stdout, stderr, and exit codes with `results=output/value` support

### Phase 3: Advanced Results Processing
1. **Result Formatting**: Support all org-mode result types (table, list, raw, file, html, verbatim)
2. **Result Handling**: Implement replace, append, prepend, silent, none modes
3. **File Output**: Support `results=file` with automatic link generation
4. **Error Handling**: Clear error reporting and recovery

### Phase 4: Session and Dependency Management
1. **Session Management**: Persistent execution contexts with named sessions
2. **Variable Passing**: Support `var` parameter for inter-block data sharing
3. **Dependency Tracking**: Implement `depends` parameter for execution ordering
4. **Caching**: Cache results based on code content and parameters
5. **Post-processing**: Support `post` parameter for result transformation

## Command Line Interface

### List Evaluable Blocks
```bash
jot eval --list [file...]       # List all evaluable code blocks in documents
jot eval -l [file...]           # Short form
```

**Example Output:**
```
example.md:
  hello_python (line 15): python
  count_files (line 23): bash
  api_test (line 31): javascript

notes.md:
  database_query (line 8): sql
  data_analysis (line 45): python
```

## Examples

### Basic Code Execution

    ```python
    print("Hello from Python!")
    ```
    <eval name="hello_python" />
    ```
    Hello from Python!
    ```

### Advanced Configuration
    ```bash
    find . -name "*.go" | wc -l
    ```
    <eval name="count_go_files" shell="bash" cwd="/home/user/project" timeout="10s" results="verbatim" />
    ```
    42
    ```

### Session-Based Execution
    ```python
    import pandas as pd
    data = pd.DataFrame({'x': [1, 2, 3], 'y': [4, 5, 6]})
    print("Data loaded")
    ```
    <eval name="load_data" session="analysis" results="output" />
    ```
    Data loaded
    ```

    ```python
    print(data.describe())
    ```
    <eval name="describe_data" session="analysis" results="table" />
    |       |   x |   y |
    |-------|-----|-----|
    | count | 3.0 | 3.0 |
    | mean  | 2.0 | 5.0 |
    | std   | 1.0 | 1.0 |

### File Output
    ```python
    import matplotlib.pyplot as plt
    plt.plot([1, 2, 3], [4, 5, 6])
    plt.savefig('/tmp/plot.png')
    print("Plot saved")
    ```
    <eval name="create_plot" results="file" file="/tmp/plot.png" />
    ![Plot](file:/tmp/plot.png)


### Multiple Code Blocks with Dependencies

    ```python
    import json
    data = {"name": "jot", "version": "1.0.0"}
    with open("/tmp/data.json", "w") as f:
        json.dump(data, f)
    print("Data written to file")
    ```
    <eval name="write_data" shell="python3" />
    ```
    Data written to file
    ```

    ```bash
    cat /tmp/data.json | jq .
    ```
    <eval name="read_data" shell="bash" depends="write_data" />
    ```
    {
      "name": "jot",
      "version": "1.0.0"
    }
    ```


## Testing Strategy

| Test Category | Focus Area | Key Tests | Tools/Methods |
|---------------|------------|-----------|---------------|
| **Unit Tests** | Individual components | Link parsing, parameter extraction, offset calculation, result formatting | Go testing framework |
| **Integration Tests** | End-to-end workflows | File update robustness, error handling, cross-platform compatibility | Shell scripts, Go integration tests |
| **Compatibility Tests** | Markdown rendering | CommonMark/GFM compliance, viewer compatibility, export formats | Multiple markdown parsers |
| **Feature Tests** | Org-mode parity | All result types, session management, dependency resolution | Comparison testing |

### Detailed Test Areas

**Unit Tests**:
- HTML element parsing and attribute extraction
- Parameter validation and error handling
- Offset calculation accuracy
- Result formatting for all types (table, list, raw, etc.)
- Session state management
- Dependency graph resolution

**Integration Tests**:
- End-to-end execution workflows
- File update robustness without truncation
- Error handling and recovery scenarios
- Cross-platform compatibility (Linux, macOS, Windows)
- Performance with large documents
- Concurrent execution safety

**Compatibility Tests**:
- Markdown parser compatibility (CommonMark, GFM)
- Rendering in various viewers (GitHub, GitLab, VS Code, etc.)
- Export to other formats (HTML, PDF)
- Editor integration (syntax highlighting, folding)

## Benefits

1. **Standards Compliance**: Works with any Markdown parser or viewer
2. **Org-Mode Compatibility**: Supports the same comprehensive feature set as org-mode babel
3. **Extensibility**: Easy to add new execution parameters and result types
4. **Visible Results**: Execution output is immediately visible in any Markdown viewer
5. **Session Management**: Persistent execution environments for complex workflows
6. **Result Flexibility**: Multiple output formats (table, list, raw, file, HTML)
7. **Dependency Management**: Automatic execution ordering based on block dependencies
8. **Caching**: Intelligent result caching to avoid unnecessary re-execution
9. **Integration**: Compatible with existing Markdown workflows and tools
10. **Simplicity**: Uses only standard Markdown syntax with no custom extensions
11. **Clean Rendering**: Empty anchor links are invisible in rendered output

## Conclusion

This proposal provides a robust, standards-compliant foundation for jot's code execution feature. By leveraging HTML-style self-closing tags for metadata encoding and offering flexible result storage options, we maintain compatibility with the Markdown ecosystem while enabling powerful code execution and result annotation capabilities.

The HTML element approach is intuitive, extensible, and maintains jot's philosophy of being close to standard tools and formats while adding powerful functionality for knowledge workers.

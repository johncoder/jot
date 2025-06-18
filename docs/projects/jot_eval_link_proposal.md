# Jot Eval: Link-Based Result Annotation Proposal

## Overview

This proposal outlines a standards-compliant approach for associating code execution results with code blocks in Markdown files using special markdown links. This method maintains compatibility with standard Markdown/GitHub Flavored Markdown while providing the metadata needed for code execution and result annotation.

## Problem Statement

The current implementation of `jot eval` faces several challenges:
- Adding metadata to fenced code block opening lines violates Markdown standards (only language is allowed)
- Results inserted directly after code blocks can break Markdown parsing
- Need for a way to associate execution options and results with specific code blocks
- Must remain compatible with standard Markdown renderers and tools

## Proposed Solution: Special Markdown Links

Use markdown links with a special format immediately following code blocks to encode execution metadata and associate results. The link format is:

```
[eval](option1=value1&option2=value2)
```

### Key Benefits

1. **Standards Compliant**: Uses valid markdown link syntax
2. **Invisible Rendering**: Links can be styled to be invisible in rendered output
3. **Flexible Metadata**: Supports arbitrary key-value options
4. **Clear Association**: Link placement directly after code block creates clear relationship
5. **Backward Compatible**: Existing markdown tools ignore or render as regular links

## Detailed Specification

### Basic Syntax

```markdown
```language
code here
```
[eval](name=example&timeout=30s)
```

### Supported Options

- `name`: Identifier for the code block (for referencing, caching, dependencies)
- `timeout`: Maximum execution time (e.g., "30s", "2m")
- `env`: Environment variables to set (e.g., "PATH=/custom/path")
- `cwd`: Working directory for execution
- `deps`: Dependencies on other named blocks (e.g., "setup,config")
- `cache`: Enable/disable result caching ("true"/"false")
- `output`: Output handling ("replace", "append", "none")

### Result Annotation

After execution, results are inserted as a code block immediately following the eval link:

```markdown
```bash
echo "Hello, World!"
```
[eval](name=hello)

```output
Hello, World!
```
```

### Complete Example

```markdown
# Database Setup Example

First, let's set up the database schema:

```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT UNIQUE
);
```
[eval](name=schema&timeout=10s)

```output
Table 'users' created successfully.
```

Now let's insert some test data:

```sql
INSERT INTO users (name, email) VALUES 
    ('Alice', 'alice@example.com'),
    ('Bob', 'bob@example.com');
```
[eval](name=test_data&deps=schema)

```output
2 rows inserted successfully.
```

Finally, let's query the data:

```sql
SELECT * FROM users;
```
[eval](name=query&deps=schema,test_data)

```output
id | name  | email
---|-------|-------------------
1  | Alice | alice@example.com  
2  | Bob   | bob@example.com
```
```

## Implementation Details

### Parsing Algorithm

1. Parse markdown AST to find fenced code blocks
2. Check if the next sibling node is a link with href starting with evaluation options
3. Parse the link's href as URL-encoded parameters
4. Associate the options with the preceding code block
5. Execute code with the specified options
6. Insert or update result blocks after the eval link

### Link Detection Pattern

- Link text must be exactly `eval`
- Link href must contain option parameters in URL query format
- Link must immediately follow a fenced code block (allowing only whitespace between)

### Error Handling

- Invalid option syntax: Log warning, skip execution
- Missing dependencies: Report error with dependency chain
- Timeout exceeded: Kill process, report timeout error
- Execution failure: Capture stderr, report in result block

### Backward Compatibility

- Existing code blocks without eval links remain unchanged
- Eval links without preceding code blocks are ignored
- Standard markdown renderers show eval links as regular links (can be hidden with CSS)

## Migration Path

### Phase 1: Parser Implementation
- Implement link detection and option parsing
- Add support for basic options (name, timeout)
- Maintain existing behavior for code blocks without eval links

### Phase 2: Advanced Features
- Add dependency resolution between named blocks
- Implement caching based on code content and dependencies
- Add environment and working directory support

### Phase 3: Optimization
- Parallel execution of independent blocks
- Incremental updates (only re-execute changed blocks)
- Integration with external tools and languages

## Alternative Rendering

For users who want eval links to be invisible in rendered markdown:

### CSS Styling
```css
a[href^="name="], a[href^="timeout="], a[href*="&"] {
    display: none;
}
```

### HTML Comments (Alternative)
If links are deemed too visible, eval options could be encoded in HTML comments:

```markdown
```bash
echo "Hello"
```
<!-- eval: name=hello&timeout=30s -->

```output
Hello
```
```

However, the link approach is preferred for better markdown compatibility.

## Conclusion

This link-based approach provides a clean, standards-compliant method for associating execution metadata with code blocks in markdown. It maintains full compatibility with existing markdown tools while enabling powerful code execution and result annotation features in jot.

The proposal balances flexibility, readability, and standards compliance, making it suitable for integration into jot's eval command while ensuring the resulting markdown files remain portable and compatible with other tools.

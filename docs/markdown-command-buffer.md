# jot Markdown Command Buffer: Design Document

## Overview

The jot Markdown Command Buffer is a robust, composable, and extensible system for modifying Markdown files in jot. It enables batching of structural and content changes—such as append, prepend, insert, remove, replace, and metadata updates—using a unified API. The system is tightly coupled to a Markdown Abstract Syntax Tree (AST) (via goldmark), ensuring precise and reliable modifications, and is designed to support both current and future jot workflows (capture, refile, refactor, etc.).

## Core Components

### 1. Markdown AST

- **Dependency:** Uses the goldmark Markdown parser to build and manipulate the AST for each file.
- **Purpose:** All operations are performed on the AST, ensuring structural correctness and enabling advanced features like fuzzy heading matching and subtree manipulation.

### 2. Command Buffer

- **Definition:** A buffer that accumulates a sequence of operations (`MarkdownOp`) to be applied to a Markdown document.
- **Supported Operations:**
  - `AppendOp`: Append content to the end of a heading’s subtree.
  - `PrependOp`: Prepend content to the start of a heading’s subtree.
  - `InsertOp`: Insert content at a specific offset or AST node.
  - `RemoveOp`: Remove a subtree or content at a selector.
  - `ReplaceOp`: Replace a subtree or content at a selector.
  - `CreateHeadingOp`: Create missing headings along a selector path.
  - `UpdateMetadataOp`: Update frontmatter or heading attributes.
  - `CompositeOp`: Batch multiple operations as a single logical command.
- **Extensibility:** New operations or composites can be defined by implementing the `MarkdownOp` interface.

### 3. Execution Function

- **Signature:** `func (buf *CommandBuffer) Apply(doc *goldmarkAST.Document, w io.Writer) error`
- **Behavior:** Applies all operations in the buffer to the AST, then renders the modified AST to the provided writer (rewriting the whole file).
- **Diff Output:** If a debug flag is set, outputs a human-readable diff (unified diff format) before applying changes.

## Selector System

- **Selectors:** Used to identify headings or content blocks within the AST. Support for heading paths (e.g., `Topic/Subtopic`), fuzzy matching, and (optionally) attribute-based selection.
- **Resolution:** Selectors are resolved against the AST, and operations are applied to the matched nodes.

## Batching and Composite Operations

- **Batching:** Multiple operations can be batched in a single buffer and applied in one pass, ensuring atomic and consistent updates.
- **Composites:** Composite operations (e.g., move subtree, refactor topic) are built by combining basic operations.

## Integration with Hooks

- **Pre/Post Hooks:** The command buffer can be integrated with jot’s hook system. For example:
  - Run a pre-modification hook to validate or transform the AST before applying changes.
  - Run a post-modification hook to format the file, update metadata, or trigger external automation.
- **Example:** After a buffer is applied, a post-capture or post-refile hook can be used to normalize whitespace or enforce style rules.

## Debugging and Diff Output

- **Debug Mode:** When enabled, the system outputs a unified diff showing changes between the original and modified file.
- **Purpose:** Useful for development, debugging, and advanced user workflows.

## Error Handling and Validation

- **Selector Errors:** Invalid or ambiguous selectors result in clear, actionable errors.
- **Operation Errors:** All operations return errors if they cannot be applied (e.g., missing heading, invalid offset).
- **Consistency:** Error handling follows jot’s standards for CLI tools.

## Extensibility and Future Directions

- **New Operations:** The system is designed to support new operations (e.g., split/merge headings, advanced formatting, plugin-defined commands).
- **External Plugins:** Future support for external subcommands or plugins that can define and execute custom command buffers.
- **Advanced Refactoring:** Enables bulk refactoring, topic merging/splitting, and metadata management.

## Examples

### Example 1: Simple Append

```go
buf := &CommandBuffer{}
buf.Operations = append(buf.Operations,
    &CreateHeadingOp{Selector: SelectorPath{Segments: []string{"Meetings"}}, Level: 1},
    &AppendOp{Selector: SelectorPath{Segments: []string{"Meetings"}}, Content: "Discussed project X."},
)
buf.Apply(doc, fileWriter)
```

### Example 2: Move Subtree

```go
buf := &CommandBuffer{}
buf.Operations = append(buf.Operations,
    &RemoveOp{Selector: SelectorPath{Segments: []string{"OldTopic"}}},
    &CreateHeadingOp{Selector: SelectorPath{Segments: []string{"NewTopic"}}, Level: 1},
    &AppendOp{Selector: SelectorPath{Segments: []string{"NewTopic"}}, Content: subtreeContent},
)
buf.Apply(doc, fileWriter)
```

### Example 3: Integration with Hooks

```go
// Before applying buffer
runPreHook(doc)
// Apply buffer
buf.Apply(doc, fileWriter)
// After applying buffer
runPostHook(filePath)
```

## Limitations and Scope

- **Markdown-Only:** The system is tightly coupled to Markdown and goldmark AST; not intended for other formats.
- **Full File Rewrite:** Always rewrites the entire file for simplicity and reliability.
- **No User Preview/Revert:** Not designed for interactive preview or revert, but diff output is available for debugging.

## References

- [jot Coding Standards](../.github/instructions/coding.instructions.md)
- [goldmark Documentation](https://github.com/yuin/goldmark)
- jot internal documentation and roadmap in `docs/`

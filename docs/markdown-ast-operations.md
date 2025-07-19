# jot Markdown AST Operations: Design and Scenarios

## Summary of Supported Scenarios

### jot capture
- **Quick content capture**: Direct append to `inbox.md` via `--content` or stdin.
- **Interactive capture**: Opens `$EDITOR`, appends result to `inbox.md`.
- **Template-based capture**: Renders template, merges with user content, appends to destination file specified in frontmatter, defaults to `inbox.md`.
- **Piped content capture**: Content piped via stdin, optionally with template, appends to `inbox.md`.
- **Pre-capture hooks**: Run on raw text before capture.

### jot refile
- **Standard refile**: Move a heading and all nested content between files or within a file.
- **Interactive mode (FZF)**: User interactively selects source subtree and destination.
- **Destination inspection**: Show structure of destination file/heading.
- **File structure display**: Show outline of a file.
- **Line number selector**: Select subtree by line number (and optionally heading).
- **Batch/automation**: Not currently supported.

## Key Operations Needed
- Parse all incoming content as Markdown AST.
- Selector logic: By heading path, fuzzy match, or line number.
- Subtree extraction/insertion: Move/copy entire heading sections.
- Re-leveling: Adjust heading levels to fit destination context (always, clamp to H6).
- AST merge/append: Insert content at correct location.
- AST-to-Markdown rendering: Write back to file.
- Hooks: Operate on raw text only (for now).
- Ambiguity: Error out for ambiguous selectors.
- No frontmatter in refile; only in capture templates.

## Design Decisions and Clarifications
1. **Re-leveling**: Always re-level incoming subtree to fit under destination heading. Clamp to H6.
2. **Frontmatter**: Only relevant for templates/capture, not for refile. Never included in final captured content or refiles.
3. **Ambiguous destination**: Error out if ambiguous.
4. **Hooks**: Ignore for now.
5. **Batching**: Not supported.
6. **Selector ambiguity**: Return a simple error.
7. **Frontmatter in files**: Not present; only in templates/capture.

## Proposed New Utilities/Operations
- `ExtractSubtree(doc, heading)`: Extract heading and all nested content as a slice of AST nodes.
- `InsertSubtree(doc, index, subtree)`: Insert a subtree at a specific index in the destination AST.
- `RelevelSubtree(subtree, delta)`: Re-level all headings in a subtree by a given delta, clamping to H6.

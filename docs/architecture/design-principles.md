# Design Principles

The core design philosophy and principles that guide jot's development and user experience.

## Core Philosophy

### Pragmatism Over Perfection

jot prioritizes **real-world utility** over theoretical ideals:

- **Fast capture** - Getting thoughts down quickly matters more than perfect organization
- **Flexible organization** - Support multiple workflows rather than enforcing one "right" way
- **Progressive complexity** - Start simple, add features as needed
- **Integration-friendly** - Work well with existing tools and workflows

### Git-Inspired Design

Following git's successful CLI patterns:

- **Subcommand structure** - `jot [command] [options]` for clear, composable operations
- **Workspace context** - Operate on local workspace, like git repositories
- **Extensibility** - Support for external commands and scripts
- **Configuration layers** - Global and workspace-specific settings
- **Status and inspection** - Always know what's happening in your workspace

### Terminal-First Approach

Designed for people who live in the terminal:

- **No GUI dependencies** - Pure CLI interface
- **Respects environment** - Uses `$EDITOR`, `$PAGER`, standard conventions
- **Shell integration** - Works well with pipes, scripts, and automation
- **Fast interaction** - Optimized for keyboard-driven workflows

## User Experience Principles

### Minimize Friction

Every interaction should be as smooth as possible:

- **Smart defaults** - Work out of the box with minimal configuration
- **Context awareness** - Understand where you are and what you're likely trying to do
- **Progressive disclosure** - Basic commands are simple, advanced features are available when needed
- **Consistent patterns** - Similar operations work the same way across commands

### Respect User Agency

Users know their workflows best:

- **Multiple input methods** - Editor, command-line, piped input all supported
- **Flexible organization** - Support different note organization strategies
- **Optional features** - Templates, automation, and advanced features are opt-in
- **Escape hatches** - Always provide ways to work directly with files

### Fail Gracefully

When things go wrong, help users recover:

- **Clear error messages** - Explain what went wrong and how to fix it
- **Safe by default** - Destructive operations require confirmation
- **Recovery tools** - `jot doctor` helps diagnose and fix common issues
- **Backup-friendly** - Note files are standard Markdown for easy backup/recovery

## Technical Principles

### Local-First Data

User's notes belong to them:

- **No cloud dependencies** - Everything stored locally
- **Standard formats** - Markdown files, JSON configuration
- **Version control friendly** - Git-compatible file structures
- **Export-friendly** - Easy to migrate to other tools

### Security by Design

Especially important for template execution:

- **Explicit approval** - Templates require user approval before executing shell commands
- **Content verification** - Template changes invalidate previous approvals
- **Principle of least privilege** - Minimal permissions required
- **Audit trail** - Track what has been approved and when

### Maintainable Codebase

Built for long-term sustainability:

- **Clear interfaces** - Well-defined boundaries between components
- **Comprehensive testing** - Unit and integration tests for critical paths
- **Documentation** - Code is documented for contributors
- **Standard tools** - Uses established Go libraries and patterns

## Workflow Design Patterns

### Capture-First Mentality

Remove barriers to getting thoughts recorded:

```bash
# These should all be fast and easy:
jot capture                           # Opens editor immediately
jot capture --content "quick thought" # No editor needed  
echo "idea" | jot capture            # Works with pipes
```

### Organize Later Pattern

Don't force organization during capture:

1. **Capture everything** to `inbox.md` first
2. **Refile periodically** using interactive tools
3. **Archive occasionally** to manage workspace size
4. **Search anytime** across all notes

### Template-Driven Structure

For users who want more structure:

- **Optional templates** - Use when helpful, skip when not
- **Dynamic content** - Shell commands for dates, context, automation
- **Security model** - Explicit approval prevents accidental execution
- **Workspace-specific** - Different templates for different contexts

## Integration Philosophy

### Play Well with Others

jot should enhance, not replace, existing workflows:

- **Editor integration** - Use your preferred editor for note editing
- **Git compatibility** - Notes work well in git repositories
- **Search tool compatibility** - Files work with `grep`, `ag`, `rg`, etc.
- **Backup tool compatibility** - Standard files work with any backup system

### Extensibility Points

Designed for customization and extension:

- **Hooks system** - Run scripts after captures, refiling, etc.
- **External commands** - `jot-*` commands in PATH are automatically available
- **Template system** - Custom templates for any structured content
- **Configuration system** - JSON5 configuration for all preferences

### Unix Philosophy Alignment

Following established Unix principles:

- **Do one thing well** - Focus on note management, not everything
- **Compose with other tools** - Work well in pipelines and scripts
- **Text streams** - Input and output are text-based
- **Configuration files** - Human-readable configuration

## Performance Considerations

### Responsive by Default

Common operations should be near-instantaneous:

- **Fast startup** - No unnecessary initialization or loading
- **Efficient search** - Quick full-text search across reasonable-sized workspaces
- **Lazy loading** - Only read files when needed
- **Incremental operations** - Don't rebuild everything for small changes

### Scale Gracefully

Handle growing note collections:

- **Archive system** - Move old notes out of active workspace
- **Efficient file organization** - Month/year-based organization prevents huge directories
- **Search optimization** - Consider indexing for large workspaces
- **Workspace separation** - Multiple workspaces for different contexts

## Evolution Strategy

### Backward Compatibility

Respect user investment in notes and workflows:

- **File format stability** - Markdown files should always be readable
- **Configuration migration** - Automatic migration of configuration formats
- **Deprecation warnings** - Give users time to adapt to changes
- **Documentation** - Clear migration guides for major changes

### Feature Development Approach

- **Core features first** - Ensure basic capture/organize/find workflows are solid
- **User-driven features** - Add features based on actual user needs
- **Experimental features** - Use feature flags for testing new capabilities
- **Regular cleanup** - Remove features that don't add value

### Community Considerations

Built for open source sustainability:

- **Clear contribution guidelines** - Make it easy for others to contribute
- **Modular architecture** - New features don't require core changes
- **Comprehensive tests** - Prevent regressions from community contributions
- **Documentation** - Both user and developer documentation

## Decision Framework

When evaluating new features or changes, we ask:

1. **Does this align with pragmatic workflows?** - Will real users find this helpful in practice?
2. **Does this increase or decrease friction?** - Does it make common tasks easier or harder?
3. **Is this the right abstraction level?** - Too low-level or too high-level for a note-taking tool?
4. **Does this respect user agency?** - Does it give users choice and control?
5. **Is this maintainable long-term?** - Can we support this feature sustainably?

## See Also

- **[File Structure](file-structure.md)** - How these principles are reflected in file organization
- **[Security Model](security-model.md)** - Security implementation details
- **[Getting Started](../user-guide/getting-started.md)** - See principles in practice

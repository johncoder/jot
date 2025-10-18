# Copilot Instructions for jot

## Project Overview

jot is a git-inspired CLI tool for managing notes and information. It uses a workspace-centric model where notes are stored as Markdown files with a structured directory layout:
- `inbox.md` - Quick capture landing zone
- `lib/` - Organized topic-based notes  
- `.jot/` - Internal data (SQLite, logs, configuration)

## Architecture Patterns

### CLI Structure (Cobra-based)
- **Root command**: `cmd/root.go` with external command discovery
- **External commands**: Discovered via PATH as `jot-*` executables (see `cmd/external.go`)
- **Command framework**: All commands use `internal/cmdutil/framework.go` with standardized lifecycle:
  ```go
  type CommandRunner interface {
      ValidateArgs(args []string) error
      ResolveWorkspace() (*workspace.Workspace, error) 
      ExecuteWithHooks(ctx *ExecutionContext) error
  }
  ```

### Workspace Discovery Algorithm
The workspace system (`internal/workspace/workspace.go`) follows a specific discovery order:
1. Walk up directories looking for `.jot/` or `.jotrc` 
2. If `.jot/` found → use that workspace
3. If `.jotrc` found first → use default workspace from config
4. Check `~/.jotrc` for global default
5. Fail with guidance if no workspace found

**Key insight**: Commands work with workspace context, never assume current directory is the workspace root.

### Configuration System (JSON5 + Environment)
- **Global**: `~/.jotrc` (JSON5 format with environment variable expansion)
- **Project**: `.jotrc` in workspace directory
- **Workspace**: `.jot/config.json` for workspace-specific settings
- **Pattern**: Use `internal/config/config.go` for global config, workspace config is handled separately

Example config structure:
```json5
{
  "workspaces": {
    "main": "/path/to/workspace",
    "work": "/path/to/work/notes"
  },
  "default": "main"
}
```

### Error Handling Patterns
Use `internal/cmdutil/errors.go` for consistent error handling:
- `HandleError()` for unified JSON/text output
- `WrapError()` for error context
- `FileError` and `OperationError` for typed errors
- Always return errors to caller, don't exit directly

### Hooks System
Located in `internal/hooks/`, provides extensible lifecycle hooks:
- **Types**: pre-capture, post-capture, pre-refile, post-refile, pre-archive, post-archive
- **Discovery**: `.jot/hooks/` (workspace) and `~/.jot/hooks/` (global)
- **Context**: Rich context passed via environment variables
- **Bypass**: `--no-verify` flag disables hooks

## Development Workflow

### Build & Test Commands
```bash
make install          # Build and install locally
make test            # Run all tests (unit + integration + lint)
make lint            # Run linting only  
./scripts/setup.sh   # Interactive development setup
```

### Key Build Details
- Version injection via ldflags in Makefile
- Cross-platform builds via `build.sh --all`
- Coverage reports generated to `coverage.html`

### SQL Database Patterns
- Use `go:embed` for SQL file embedding (as per coding standards)
- Always use parameterized queries for SQL injection prevention
- Database files stored in `.jot/` directory

## Code Patterns Specific to jot

### Markdown Processing
- Uses `github.com/yuin/goldmark` for AST manipulation
- Custom subtree handling in `internal/markdown/markdown.go`
- **Pattern**: Parse → Transform → Write for content modifications

### Template System
- Templates stored in `.jot/templates/`
- Use `internal/template/template.go` for template rendering
- Support for content injection and editor integration

### Tangle System (Code Generation)
- Extract code blocks from Markdown to files
- Located in `internal/tangle/engine.go`
- **Usage**: `jot tangle` processes special code block metadata

### External Command Integration
- Automatic discovery of `jot-*` commands in PATH
- Global flag forwarding to external commands via environment variables
- See `cmd/external.go` for implementation details

## Key Files for Understanding

- `cmd/root.go` - CLI entry point and external command handling  
- `internal/cmdutil/framework.go` - Command execution framework
- `internal/workspace/workspace.go` - Workspace discovery and management
- `internal/hooks/common.go` - Hook system architecture
- `cmd/refile.go` - Complex markdown manipulation example (1800+ lines)
- `cmd/capture.go` - Template integration and content handling

## Testing Patterns

- Unit tests alongside source files (`*_test.go`)
- Integration tests via scripts in `scripts/test.sh`  
- Use `runTests` tool for targeted test execution
- Mock workspace creation for isolated testing

## Common Gotchas

1. **Workspace Context**: Always resolve workspace before file operations
2. **Markdown AST**: Use goldmark AST for reliable markdown manipulation, not regex
3. **Error Handling**: Use structured errors from `cmdutil/errors.go`, not fmt.Errorf directly  
4. **Hooks**: Remember hooks can modify content and abort operations
5. **External Commands**: Global flags must be parsed and forwarded correctly
6. **Cross-platform**: Use `filepath.Join()` and respect `$EDITOR`/`$VISUAL`/`$PAGER` environment variables
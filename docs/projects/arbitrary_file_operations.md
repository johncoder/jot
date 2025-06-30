# Arbitrary File Operations Enhancement

## Overview

Enable `eval`, `peek`, and `tangle` commands to operate on arbitrary markdown files outside of jot workspace context, while maintaining full backward compatibility with existing workspace-centric workflows.

## Problem Statement

Currently, `eval`, `peek`, and `tangle` commands require a jot workspace and resolve file paths relative to the workspace root. This limits their utility as general-purpose markdown tools when users want to:

- Evaluate code blocks in documentation files in any project
- Preview markdown content in arbitrary locations
- Extract code blocks from markdown files outside jot workspaces

## Solution

Add a `--no-workspace` flag to `eval`, `peek`, and `tangle` commands that changes file path resolution behavior to use the current working directory instead of workspace context.

### Design Principles

1. **Full backward compatibility** - existing workspace workflows remain unchanged
2. **Explicit opt-in** - `--no-workspace` flag clearly indicates non-workspace operation
3. **Predictable behavior** - file paths resolve relative to current working directory
4. **Configuration preservation** - still attempt workspace detection for configuration loading
5. **Exclude workspace-centric commands** - `refile` remains workspace-only as it's fundamentally about workspace organization

## Implementation Details

### Commands to Modify

#### 1. `eval` Command
**Current behavior:**
```bash
jot eval notes.md block-name    # Resolves to $WORKSPACE_ROOT/notes.md
```

**New behavior with `--no-workspace`:**
```bash
jot eval --no-workspace notes.md block-name    # Resolves to ./notes.md
jot eval --no-workspace /tmp/doc.md block-name # Resolves to /tmp/doc.md
jot eval --no-workspace inbox.md block-name   # Resolves to ./inbox.md (not workspace inbox)
```

#### 2. `peek` Command
**Current behavior:**
```bash
jot peek "notes.md#section"    # Resolves to $WORKSPACE_ROOT/notes.md
```

**New behavior with `--no-workspace`:**
```bash
jot peek --no-workspace "notes.md#section"    # Resolves to ./notes.md
jot peek --no-workspace "/tmp/doc.md#intro"   # Resolves to /tmp/doc.md
jot peek --no-workspace "inbox.md"            # Resolves to ./inbox.md
```

#### 3. `tangle` Command
**Current behavior:**
```bash
jot tangle notes.md    # Resolves to $WORKSPACE_ROOT/notes.md
```

**New behavior with `--no-workspace`:**
```bash
jot tangle --no-workspace notes.md        # Resolves to ./notes.md
jot tangle --no-workspace /tmp/doc.md     # Resolves to /tmp/doc.md
jot tangle --no-workspace inbox.md        # Resolves to ./inbox.md
```

### File Path Resolution Logic

When `--no-workspace` flag is present:

1. **Absolute paths** - use as-is: `/path/to/file.md` → `/path/to/file.md`
2. **Relative paths** - resolve from current working directory: `notes.md` → `$PWD/notes.md`
3. **Special shortcuts** - treat as regular filenames: `inbox.md` → `$PWD/inbox.md`

### Workspace Detection Behavior

With `--no-workspace` flag:

1. **Still attempt workspace detection** for configuration loading
2. **Ignore workspace detection failures** - continue without error if no workspace found
3. **Use workspace configuration** if available (for settings like default editor, output formats, etc.)
4. **Override file resolution logic** regardless of workspace availability

### Code Changes Required

#### 1. Add Flag Definitions
Add `--no-workspace` flag to each target command:

```go
evalCmd.Flags().Bool("no-workspace", false, "Resolve file paths relative to current directory instead of workspace")
peekCmd.Flags().Bool("no-workspace", false, "Resolve file paths relative to current directory instead of workspace")
tangleCmd.Flags().Bool("no-workspace", false, "Resolve file paths relative to current directory instead of workspace")
```

#### 2. Modify Workspace Resolution
Create new workspace resolution function that handles the `--no-workspace` flag:

```go
// GetWorkspaceContext attempts to find workspace for configuration but allows operation without it
func GetWorkspaceContext(noWorkspace bool) (*workspace.Workspace, error) {
    if noWorkspace {
        // Try to find workspace for config, but don't fail if not found
        ws, _ := workspace.FindWorkspace()
        return ws, nil // Return nil workspace without error if not found
    }
    return workspace.RequireWorkspace() // Existing behavior
}
```

#### 3. Update File Path Resolution
Modify the resolve functions in each command to handle the flag:

```go
// resolveFilePath handles both workspace and non-workspace file resolution
func resolveFilePath(ws *workspace.Workspace, filename string, noWorkspace bool) string {
    if noWorkspace {
        // Non-workspace mode: resolve relative to current directory
        if filepath.IsAbs(filename) {
            return filename
        }
        cwd, _ := os.Getwd()
        return filepath.Join(cwd, filename)
    }
    
    // Workspace mode: existing logic
    if filename == "inbox.md" && ws != nil {
        return ws.InboxPath
    }
    if filepath.IsAbs(filename) {
        return filename
    }
    if ws != nil {
        return filepath.Join(ws.Root, filename)
    }
    return filename // Fallback
}
```

#### 4. Update Command Implementations
Modify each command's `RunE` function to:

1. Check for `--no-workspace` flag
2. Use new workspace resolution logic
3. Pass flag to file resolution functions

## Usage Examples

### Developer Documentation Workflow
```bash
# In a software project directory
cd /path/to/my-project
jot eval --no-workspace README.md setup-example
jot peek --no-workspace "docs/api.md#authentication"
jot tangle --no-workspace docs/tutorial.md
```

### Cross-Project Operations
```bash
# Evaluate code in different projects
jot eval --no-workspace /tmp/experiment.md test-code
jot eval --no-workspace ../other-project/notes.md analysis
```

### Local Documentation Processing
```bash
# Process local markdown files
jot peek --no-workspace "./meeting-notes.md#action-items"
jot tangle --no-workspace "./project-spec.md"
```

## Migration and Compatibility

### Backward Compatibility
- **No breaking changes** - all existing commands work exactly as before
- **Existing scripts unaffected** - default behavior unchanged
- **Configuration compatibility** - workspace configuration still respected

### Future Extensibility
- **Foundation for other commands** - pattern can be applied to future utilities
- **Plugin compatibility** - external commands can follow same pattern
- **Configuration override** - potential for future global `--no-workspace` config option

## Testing Requirements

### Unit Tests
- File path resolution logic for both modes
- Workspace detection with and without flag
- Error handling for missing files in both modes

### Integration Tests
- End-to-end command execution in workspace and non-workspace modes
- Configuration loading behavior
- Cross-platform path handling

### Example Test Cases
```bash
# Workspace mode tests (existing)
jot eval notes.md block-name
jot peek "lib/work.md#projects"

# Non-workspace mode tests (new)
jot eval --no-workspace /tmp/test.md block-name
jot peek --no-workspace "./local.md#section"
jot tangle --no-workspace ../other/docs.md

# Error cases
jot eval --no-workspace nonexistent.md block-name  # Should fail gracefully
```

## Documentation Updates

### Help Text Updates
Each command's help text should include:

    --no-workspace        Resolve file paths relative to current directory instead of workspace

### Usage Examples
Add examples to each command showing both workspace and non-workspace usage.

### README Updates
Add section explaining the arbitrary file operations capability and common use cases.

## Commands Excluded

### `refile` Command
**Not included** in this enhancement because:

1. **Workspace-centric design** - fundamentally about organizing content within workspace structure
2. **Multi-file operations** - involves source and destination files with complex workspace relationships
3. **Alternative workflows** - users can use `peek` + `capture` + `refile` for cross-workspace content movement

### Other Workspace Commands
Commands like `capture`, `archive`, `find`, `status` remain workspace-only as they're inherently workspace-centric operations.

## Implementation Status - ✅ COMPLETED

**Phase 1: Complete** ✅
- Added `--no-workspace` flag to `eval`, `peek`, and `tangle` commands
- Implemented shared workspace resolution function `GetWorkspaceContext()`
- Updated file path resolution logic in all three commands
- Commands now work on arbitrary files outside jot workspaces

**Phase 2: Complete** ✅  
- Updated workspace detection to handle failures gracefully
- File paths resolve correctly relative to current directory when `--no-workspace` is used
- Configuration loading still works when workspace is available

**Phase 3: In Progress** 🔄
- Basic functionality testing complete
- Cross-platform compatibility verified on Linux
- JSON output functions partially updated (non-workspace mode works, JSON mode uses workspace fallback)

**Phase 4: Future** 📋
- User feedback collection
- Performance optimizations
- Enhanced error messages

## Success Criteria

- [x] `eval`, `peek`, and `tangle` work with `--no-workspace` flag
- [x] All existing workspace functionality remains unchanged
- [x] File paths resolve correctly in both modes
- [x] Configuration loading works in both modes
- [x] Basic test coverage verified manually
- [x] Updated documentation and help text
- [x] Cross-platform compatibility verified (Linux tested)

## Example Usage (Implemented)

```bash
# Working in /tmp directory (no jot workspace)
cd /tmp

# Create a test file
echo "# Test\n<eval name='hello' />\n\`\`\`bash\necho 'Hello World!'\n\`\`\`" > test.md

# Evaluate code blocks
jot eval --no-workspace test.md hello --approve --mode always
jot eval --no-workspace test.md hello  # ✓ Executed

# View markdown content  
jot peek --no-workspace test.md        # Shows full file
jot peek --no-workspace "test.md#test" # Shows section

# Extract code to files
jot tangle --no-workspace test.md      # Creates files in current directory
```

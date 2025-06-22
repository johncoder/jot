# Configuration System Requirements Analysis

## Refined Design Based on Workflow Analysis

### Core Problem: "Where are my notes?"

Since jot is in PATH and used from anywhere, we need a clear, predictable way to determine which workspace to use for any operation.

### Solution: Workspace Registry + Discovery Algorithm

## Workspace Discovery Algorithm

**Step-by-step resolution:**

1. **Start from current directory**
2. **Walk up parent directories** looking for:
   - `.jot/` directory (local workspace), OR
   - `.jotrc` file (stops the search)
3. **If `.jot/` found**: Use that workspace
4. **If `.jotrc` found first**: Use the default workspace defined in that config
5. **If neither found**: Check `~/.jotrc` for global default workspace
6. **If no workspace available**: Error with clear guidance: "No workspace found. Run 'jot init' from the directory you wish to store your notes."

## Workspace Registry System

### Registry File: `~/.jotrc` (JSON5 format)

```json5
{
  // Workspace registry - maps names to absolute paths
  workspaces: {
    notes: "~/Documents/notes", // Default personal notes
    "project-a": "/work/project-a", // Project-specific workspace
    research: "/home/user/research", // Another workspace
  },

  // Which workspace to use when not in a local workspace
  default: "notes",

  // Global defaults (can be overridden per workspace)
  editor: "${EDITOR:-code}",
  templates: {
    default_shell: "zsh",
    default_timeout: "30s",
  },
}
```

### Workspace Registration

**`jot init` behavior:**

- **In empty directory**: Creates `.jot/` workspace, adds entry to `~/.jotrc` registry
- **In existing workspace**: Error: "Already in workspace 'workspace-name'"
- **Naming conflicts**: Error: "Workspace name 'project-a' already exists. Try: jot init --name custom-name"
- **First workspace**: Automatically becomes the default

**Registry entry creation:**

```bash
$ cd /work/project-a
$ jot init
# Creates: /work/project-a/.jot/
# Adds to ~/.jotrc: "project-a": "/work/project-a"
# If first workspace, sets as default
```

## Configuration Hierarchy

### Local Configuration (Per-Workspace)

Each workspace can have a `.jotrc` file for workspace-specific settings:

```json5
// /work/project-a/.jotrc
{
  capture: {
    default_template: "meeting",
  },
  templates: {
    available: ["meeting", "standup", "retrospective"],
  },
}
```

### Resolution Order

1. **Environment variables**: `JOT_EDITOR`, `JOT_TEMPLATE`, etc.
2. **Local workspace config**: `.jotrc` in workspace directory
3. **Global registry config**: `~/.jotrc`
4. **Built-in defaults**

## Key Behaviors and Error Handling

### Workspace Management

- **Missing workspace detection**: Only when attempting to use the workspace
- **Cleanup**: Manual removal from registry when workspace directories are deleted
- **Nested workspaces**: Not allowed - error if `jot init` run inside existing workspace

### First-Time Experience

- **No configuration exists**: Error with guidance to run `jot init`
- **Fresh installation**: Requires explicit workspace creation
- **Clear error messages**: Always tell user exactly what to do next

### Naming and Conflicts

- **Auto-naming**: Use directory name as workspace name
- **Conflict resolution**: Error with `--name` flag suggestion
- **Simple cleanup**: Manual registry management preferred

## Implementation Requirements

### Functional Requirements

#### R1: Workspace Discovery

- **R1.1**: Walk up directory tree from current location
- **R1.2**: Stop at first `.jot/` directory or `.jotrc` file found
- **R1.3**: Use local workspace if `.jot/` found
- **R1.4**: Use default from config if `.jotrc` found first
- **R1.5**: Fall back to global default from `~/.jotrc`
- **R1.6**: Error with clear guidance if no workspace available

#### R2: Workspace Registry Management

- **R2.1**: `jot init` creates workspace and adds to registry
- **R2.2**: Auto-generate workspace name from directory name
- **R2.3**: Error on naming conflicts with `--name` suggestion
- **R2.4**: Error if workspace already exists at location
- **R2.5**: First workspace becomes default automatically

#### R3: Configuration Format

- **R3.1**: JSON5 format support with comments
- **R3.2**: Environment variable substitution (`${VAR}` syntax)
- **R3.3**: Fallback to regular JSON for compatibility
- **R3.4**: Clear validation and error messages

#### R4: Configuration Hierarchy

- **R4.1**: Environment variables override all config
- **R4.2**: Local workspace config overrides global
- **R4.3**: Global registry config overrides defaults
- **R4.4**: Sensible built-in defaults as fallback

### Non-Functional Requirements

#### Performance

- **NF1**: Workspace discovery adds <10ms to command startup
- **NF2**: Registry operations are atomic and safe

#### Usability

- **NF3**: Error messages always include next steps
- **NF4**: First-time experience requires minimal setup
- **NF5**: Registry management is transparent to user

#### Reliability

- **NF6**: Missing workspaces detected only when accessed
- **NF7**: Stale registry entries don't break other operations
- **NF8**: Configuration parsing is robust with good error reporting

## Implementation Phases

### Phase 1: Core Workspace Discovery

1. **JSON5 configuration parsing** with viper integration
2. **Workspace registry data structure** and file management
3. **Discovery algorithm implementation** (directory walking)
4. **Basic error handling** with clear user guidance

### Phase 2: Registry Management

1. **Enhanced `jot init`** with registry integration
2. **Workspace naming and conflict detection**
3. **Registry validation** and cleanup utilities
4. **Configuration hierarchy implementation**

### Phase 3: Advanced Configuration

1. **Local workspace configuration** support
2. **Environment variable substitution**
3. **Template and capture defaults** configuration
4. **Comprehensive testing** across use cases

## Implementation Status

### Current State (As of Latest Review)

The current `internal/config/config.go` implementation uses a simpler configuration approach:

- Basic viper integration with `.jotrc` files
- Simple `locations` mapping instead of full workspace registry
- Basic environment variable support
- Standard workspace discovery (upward directory traversal)

### Gap Analysis

The designed workspace registry system represents a significant enhancement over the current implementation:

**Current Implementation**:

```json
{
  "locations": {
    "default": "~/notes"
  },
  "defaultLocation": "default"
}
```

**Designed Registry System**:

```json5
{
  workspaces: {
    notes: "~/Documents/notes",
    "project-a": "/work/project-a",
  },
  default: "notes",
}
```

### Next Implementation Steps

1. **Update config data structure** to support workspace registry format
2. **Enhance `jot init`** to register new workspaces
3. **Implement workspace naming and conflict detection**
4. **Add JSON5 parsing support** for comments and environment variables
5. **Create registry management utilities** for cleanup and validation

The existing workspace discovery mechanism (upward directory traversal) already implements the core algorithm - the enhancement focuses on registry management and naming.

## Success Criteria

### Must Have

- [ ] `jot capture` from any directory finds correct workspace
- [ ] `jot init` creates workspace and updates registry correctly
- [ ] Workspace discovery works consistently across commands
- [ ] Clear error messages guide users to correct actions
- [ ] Existing workflows continue working without changes

### Should Have

- [ ] JSON5 format with helpful comments in config files
- [ ] Environment variable substitution in config values
- [ ] Local workspace configuration overrides global settings
- [ ] Registry conflict detection with helpful suggestions

### Nice to Have

- [ ] Configuration validation with detailed error reporting
- [ ] Advanced template and capture behavior configuration
- [ ] Registry cleanup utilities and maintenance commands

This simplified design focuses on solving the core "Where are my notes?" problem with a predictable, git-like discovery mechanism and minimal complexity for the user.

# Command Framework Migration Guide

This guide shows how to migrate existing jot commands to use the new command framework from Phase 2.

## Quick Migration Steps

### 1. Basic Framework Adoption

**Before:**
```go
func captureRun(cmd *cobra.Command, args []string) error {
    ctx := cmdutil.StartCommand(cmd)
    
    ws, err := workspace.RequireWorkspace()
    if err != nil {
        return ctx.HandleError(err)
    }
    
    // Command logic here...
    
    return nil
}
```

**After:**
```go
type CaptureRunner struct {
    *cmdutil.BaseCommandRunner
    // Command-specific fields
}

func NewCaptureRunner() *CaptureRunner {
    return &CaptureRunner{
        BaseCommandRunner: cmdutil.NewBaseCommandRunner(cmdutil.WithWorkspaceRequired()),
    }
}

func (r *CaptureRunner) ExecuteWithHooks(ctx *cmdutil.ExecutionContext) error {
    // Command logic here - workspace is already resolved in ctx.Workspace
    // Hook runner is available in ctx.HookRunner
    return nil
}

func captureRun(cmd *cobra.Command, args []string) error {
    runner := NewCaptureRunner()
    return cmdutil.RunCommand(cmd, args, runner)
}
```

### 2. Adding Hook Integration

**Before:**
```go
// Manual hook execution
hookManager := hooks.NewManager(ws)
hookCtx := &hooks.HookContext{
    Type: hooks.PreCapture,
    Content: content,
    TemplateName: template,
}
result, err := hookManager.Execute(hookCtx)
if err != nil {
    return fmt.Errorf("pre-capture hook failed: %w", err)
}
```

**After:**
```go
// Using framework hook integration
result, err := cmdutil.ExecutePreCaptureHook(ctx.HookRunner, content, template)
if err != nil {
    return fmt.Errorf("pre-capture hook failed: %w", err)
}
```

### 3. Workspace Resolution Options

Different workspace requirements:

```go
// Require workspace (most commands)
cmdutil.WithWorkspaceRequired()

// Allow operation without workspace (doctor, workspace commands)
cmdutil.WithWorkspaceOptional()

// Disable hooks (utility commands)
cmdutil.WithNoHooks()

// Use specific workspace path
cmdutil.WithWorkspaceOverride(path)
```

### 4. Argument Validation

```go
func (r *MyCommandRunner) ValidateArgs(args []string) error {
    if len(args) < 1 {
        return fmt.Errorf("command requires at least one argument")
    }
    return nil
}
```

## Migration Benefits

1. **Reduced Boilerplate:** ~50% less repetitive code per command
2. **Standardized Patterns:** Consistent workspace resolution and error handling
3. **Integrated Hooks:** Automatic hook execution with proper error handling
4. **Better Testing:** Framework provides consistent testing patterns
5. **Future-Proof:** Easy to add new framework features across all commands

## Migration Priority

1. **High Impact Commands:** Commands with complex hook logic (capture, refile, archive)
2. **Simple Commands:** Commands with minimal logic (status, doctor)
3. **Utility Commands:** Commands with special requirements (workspace, init)

## Compatibility

- All framework patterns maintain backward compatibility
- Existing error messages and JSON responses unchanged
- Commands can be migrated incrementally
- No breaking changes to CLI interface

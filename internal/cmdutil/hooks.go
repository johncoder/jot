package cmdutil

import (
	"fmt"

	"github.com/johncoder/jot/internal/hooks"
	"github.com/johncoder/jot/internal/workspace"
)

// HookRunner provides standardized hook execution for commands
type HookRunner struct {
	manager   *hooks.Manager
	workspace *workspace.Workspace
	noVerify  bool
}

// NewHookRunner creates a new hook runner for the given workspace
func NewHookRunner(ws *workspace.Workspace, noVerify bool) *HookRunner {
	if ws == nil {
		return &HookRunner{
			workspace: nil,
			noVerify:  noVerify,
		}
	}

	return &HookRunner{
		manager:   hooks.NewManager(ws),
		workspace: ws,
		noVerify:  noVerify,
	}
}

// ExecutePreHook executes a pre-operation hook
func (r *HookRunner) ExecutePreHook(hookType hooks.HookType, ctx *HookExecutionContext) (*hooks.HookResult, error) {
	if r.manager == nil || r.noVerify {
		return &hooks.HookResult{Content: ctx.Content}, nil
	}

	hookCtx := &hooks.HookContext{
		Type:         hookType,
		Workspace:    r.workspace,
		Content:      ctx.Content,
		SourceFile:   ctx.SourceFile,
		DestPath:     ctx.DestPath,
		TemplateName: ctx.TemplateName,
		ExtraEnv:     ctx.ExtraEnv,
		AllowBypass:  r.noVerify,
	}

	result, err := r.manager.Execute(hookCtx)
	if err != nil {
		return nil, fmt.Errorf("pre-hook execution failed: %w", err)
	}

	if result.Aborted {
		return result, fmt.Errorf("operation aborted by pre-hook")
	}

	return result, nil
}

// ExecutePostHook executes a post-operation hook (informational only)
func (r *HookRunner) ExecutePostHook(hookType hooks.HookType, ctx *HookExecutionContext) error {
	if r.manager == nil || r.noVerify {
		return nil
	}

	hookCtx := &hooks.HookContext{
		Type:         hookType,
		Workspace:    r.workspace,
		Content:      ctx.Content,
		SourceFile:   ctx.SourceFile,
		DestPath:     ctx.DestPath,
		TemplateName: ctx.TemplateName,
		ExtraEnv:     ctx.ExtraEnv,
		AllowBypass:  r.noVerify,
	}

	_, err := r.manager.Execute(hookCtx)
	// Post-hooks are informational only - don't fail the operation
	return err
}

// HookExecutionContext contains context information for hook execution
type HookExecutionContext struct {
	Content      string            // Content to be processed
	SourceFile   string            // Source file for operations
	DestPath     string            // Destination path for operations
	TemplateName string            // Template name for capture
	ExtraEnv     map[string]string // Additional environment variables
}

// NewHookExecutionContext creates a new hook execution context
func NewHookExecutionContext() *HookExecutionContext {
	return &HookExecutionContext{
		ExtraEnv: make(map[string]string),
	}
}

// WithContent sets the content for the hook context
func (ctx *HookExecutionContext) WithContent(content string) *HookExecutionContext {
	ctx.Content = content
	return ctx
}

// WithSourceFile sets the source file for the hook context
func (ctx *HookExecutionContext) WithSourceFile(file string) *HookExecutionContext {
	ctx.SourceFile = file
	return ctx
}

// WithDestPath sets the destination path for the hook context
func (ctx *HookExecutionContext) WithDestPath(path string) *HookExecutionContext {
	ctx.DestPath = path
	return ctx
}

// WithTemplateName sets the template name for the hook context
func (ctx *HookExecutionContext) WithTemplateName(name string) *HookExecutionContext {
	ctx.TemplateName = name
	return ctx
}

// WithEnv adds an environment variable for the hook context
func (ctx *HookExecutionContext) WithEnv(key, value string) *HookExecutionContext {
	ctx.ExtraEnv[key] = value
	return ctx
}

// ExecutePreCaptureHook is a convenience function for pre-capture hooks
func ExecutePreCaptureHook(runner *HookRunner, content, template string) (*hooks.HookResult, error) {
	if runner == nil {
		return &hooks.HookResult{Content: content}, nil
	}

	ctx := NewHookExecutionContext().
		WithContent(content).
		WithTemplateName(template)

	return runner.ExecutePreHook(hooks.PreCapture, ctx)
}

// ExecutePostCaptureHook is a convenience function for post-capture hooks
func ExecutePostCaptureHook(runner *HookRunner, content, destPath string) error {
	if runner == nil {
		return nil
	}

	ctx := NewHookExecutionContext().
		WithContent(content).
		WithDestPath(destPath)

	return runner.ExecutePostHook(hooks.PostCapture, ctx)
}

// ExecutePreRefileHook is a convenience function for pre-refile hooks
func ExecutePreRefileHook(runner *HookRunner, sourceFile, destPath string) (*hooks.HookResult, error) {
	if runner == nil {
		return &hooks.HookResult{}, nil
	}

	ctx := NewHookExecutionContext().
		WithSourceFile(sourceFile).
		WithDestPath(destPath)

	return runner.ExecutePreHook(hooks.PreRefile, ctx)
}

// ExecutePostRefileHook is a convenience function for post-refile hooks
func ExecutePostRefileHook(runner *HookRunner, sourceFile, destPath string) error {
	if runner == nil {
		return nil
	}

	ctx := NewHookExecutionContext().
		WithSourceFile(sourceFile).
		WithDestPath(destPath)

	return runner.ExecutePostHook(hooks.PostRefile, ctx)
}

// ExecutePreArchiveHook is a convenience function for pre-archive hooks
func ExecutePreArchiveHook(runner *HookRunner, sourceFile string) (*hooks.HookResult, error) {
	if runner == nil {
		return &hooks.HookResult{}, nil
	}

	ctx := NewHookExecutionContext().
		WithSourceFile(sourceFile)

	return runner.ExecutePreHook(hooks.PreArchive, ctx)
}

// ExecutePostArchiveHook is a convenience function for post-archive hooks
func ExecutePostArchiveHook(runner *HookRunner, sourceFile string) error {
	if runner == nil {
		return nil
	}

	ctx := NewHookExecutionContext().
		WithSourceFile(sourceFile)

	return runner.ExecutePostHook(hooks.PostArchive, ctx)
}

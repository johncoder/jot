package cmdutil

import (
	"fmt"
	"time"

	"github.com/johncoder/jot/internal/workspace"
	"github.com/spf13/cobra"
)

// CommandRunner defines the standard lifecycle for command execution
type CommandRunner interface {
	ValidateArgs(args []string) error
	ResolveWorkspace() (*workspace.Workspace, error)
	ExecuteWithHooks(ctx *ExecutionContext) error
}

// ExecutionContext extends CommandContext with additional execution data
type ExecutionContext struct {
	*CommandContext
	Args       []string
	Workspace  *workspace.Workspace
	Options    *CommandOptions
	HookRunner *HookRunner
}

// CommandOptions configures command execution behavior
type CommandOptions struct {
	RequireWorkspace    bool
	AllowNoWorkspace    bool
	WorkspaceOverride   string
	NoVerify           bool
	EnablePreHooks     bool
	EnablePostHooks    bool
}

// BaseCommandRunner provides a default implementation of CommandRunner
type BaseCommandRunner struct {
	Options *CommandOptions
}

// NewBaseCommandRunner creates a new BaseCommandRunner with default options
func NewBaseCommandRunner(opts *CommandOptions) *BaseCommandRunner {
	if opts == nil {
		opts = &CommandOptions{
			RequireWorkspace: true,
			EnablePreHooks:   true,
			EnablePostHooks:  true,
		}
	}
	return &BaseCommandRunner{Options: opts}
}

// ValidateArgs provides a default implementation that accepts any args
func (r *BaseCommandRunner) ValidateArgs(args []string) error {
	// Default implementation - override in specific commands for validation
	return nil
}

// ResolveWorkspace handles workspace resolution based on command options
func (r *BaseCommandRunner) ResolveWorkspace() (*workspace.Workspace, error) {
	return ResolveWorkspace(r.Options)
}

// ResolveWorkspace provides standardized workspace resolution based on options
func ResolveWorkspace(opts *CommandOptions) (*workspace.Workspace, error) {
	if opts == nil {
		opts = WithWorkspaceRequired()
	}

	if opts.AllowNoWorkspace {
		// Try to find workspace but don't fail if not found
		ws, err := workspace.FindWorkspace()
		if err != nil {
			return nil, nil // No workspace found, but that's okay
		}
		return ws, nil
	}

	if opts.WorkspaceOverride != "" {
		return workspace.RequireWorkspaceWithOverride(opts.WorkspaceOverride)
	}

	if opts.RequireWorkspace {
		return workspace.RequireWorkspace()
	}

	return workspace.FindWorkspace()
}

// ResolveWorkspaceFromCommand extracts workspace options from command flags and resolves
func ResolveWorkspaceFromCommand(cmd *cobra.Command) (*workspace.Workspace, error) {
	opts := extractCommandOptions(cmd)
	return ResolveWorkspace(opts)
}

// ExecuteWithHooks provides a base implementation that subcommands can override
func (r *BaseCommandRunner) ExecuteWithHooks(ctx *ExecutionContext) error {
	// This is meant to be overridden by specific command implementations
	return fmt.Errorf("ExecuteWithHooks must be implemented by specific command runner")
}

// RunCommand executes a command using the CommandRunner interface
func RunCommand(cmd *cobra.Command, args []string, runner CommandRunner) error {
	startTime := time.Now()
	ctx := NewExecutionContext(cmd, args, startTime)

	// Validate arguments
	if err := runner.ValidateArgs(args); err != nil {
		return ctx.HandleError(fmt.Errorf("invalid arguments: %w", err))
	}

	// Resolve workspace
	ws, err := runner.ResolveWorkspace()
	if err != nil {
		return ctx.HandleError(fmt.Errorf("workspace resolution failed: %w", err))
	}
	ctx.Workspace = ws
	
	// Initialize hook runner
	ctx.HookRunner = NewHookRunner(ws, ctx.Options.NoVerify)

	// Execute pre-hooks if enabled
	if ctx.Options != nil && ctx.Options.EnablePreHooks && !ctx.Options.NoVerify {
		if err := runPreHooks(ctx); err != nil {
			return ctx.HandleError(fmt.Errorf("pre-hook execution failed: %w", err))
		}
	}

	// Execute main command logic
	if err := runner.ExecuteWithHooks(ctx); err != nil {
		return ctx.HandleError(err)
	}

	// Execute post-hooks if enabled
	if ctx.Options != nil && ctx.Options.EnablePostHooks && !ctx.Options.NoVerify {
		if err := runPostHooks(ctx); err != nil {
			return ctx.HandleError(fmt.Errorf("post-hook execution failed: %w", err))
		}
	}

	return nil
}

// NewExecutionContext creates a new ExecutionContext
func NewExecutionContext(cmd *cobra.Command, args []string, startTime time.Time) *ExecutionContext {
	return &ExecutionContext{
		CommandContext: StartCommand(cmd),
		Args:           args,
		Options:        extractCommandOptions(cmd),
		HookRunner:     nil, // Will be set after workspace resolution
	}
}

// extractCommandOptions extracts command options from cobra flags
func extractCommandOptions(cmd *cobra.Command) *CommandOptions {
	opts := &CommandOptions{}

	// Extract common flags that might be present
	if cmd.Flags().Lookup("no-verify") != nil {
		opts.NoVerify, _ = cmd.Flags().GetBool("no-verify")
	}

	if cmd.Flags().Lookup("workspace") != nil {
		opts.WorkspaceOverride, _ = cmd.Flags().GetString("workspace")
	}

	if cmd.Flags().Lookup("no-workspace") != nil {
		noWorkspace, _ := cmd.Flags().GetBool("no-workspace")
		opts.AllowNoWorkspace = noWorkspace
		if noWorkspace {
			opts.RequireWorkspace = false
		}
	}

	return opts
}

// runPreHooks executes pre-command hooks - placeholder for future hook types
func runPreHooks(ctx *ExecutionContext) error {
	// This is a placeholder for command-specific pre-hooks
	// Individual commands should handle their specific hook types using ctx.HookRunner
	return nil
}

// runPostHooks executes post-command hooks - placeholder for future hook types
func runPostHooks(ctx *ExecutionContext) error {
	// This is a placeholder for command-specific post-hooks
	// Individual commands should handle their specific hook types using ctx.HookRunner
	return nil
}

// WithWorkspaceRequired creates command options that require a workspace
func WithWorkspaceRequired() *CommandOptions {
	return &CommandOptions{
		RequireWorkspace: true,
		EnablePreHooks:   true,
		EnablePostHooks:  true,
	}
}

// WithWorkspaceOptional creates command options that allow operation without a workspace
func WithWorkspaceOptional() *CommandOptions {
	return &CommandOptions{
		RequireWorkspace:  false,
		AllowNoWorkspace:  true,
		EnablePreHooks:    true,
		EnablePostHooks:   true,
	}
}

// WithNoHooks creates command options that disable hook execution
func WithNoHooks() *CommandOptions {
	return &CommandOptions{
		RequireWorkspace: true,
		EnablePreHooks:   false,
		EnablePostHooks:  false,
	}
}

// WithWorkspaceOverride creates command options with a specific workspace path
func WithWorkspaceOverride(path string) *CommandOptions {
	return &CommandOptions{
		RequireWorkspace:  true,
		WorkspaceOverride: path,
		EnablePreHooks:    true,
		EnablePostHooks:   true,
	}
}

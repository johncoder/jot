package hooks

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/johncoder/jot/internal/workspace"
)

// HookType represents the type of hook being executed
type HookType string

const (
	PreCapture      HookType = "pre-capture"
	PostCapture     HookType = "post-capture"
	PreRefile       HookType = "pre-refile"
	PostRefile      HookType = "post-refile"
	PreArchive      HookType = "pre-archive"
	PostArchive     HookType = "post-archive"
	PreEval         HookType = "pre-eval"
	PostEval        HookType = "post-eval"
	WorkspaceChange HookType = "workspace-change"
)

// HookContext contains the context information passed to hooks
type HookContext struct {
	Type         HookType
	Workspace    *workspace.Workspace
	Content      string            // Content to be processed (for content hooks)
	SourceFile   string            // Source file for operations
	DestPath     string            // Destination path for operations
	TemplateName string            // Template name for capture
	ExtraEnv     map[string]string // Additional environment variables
	Timeout      time.Duration
	AllowBypass  bool // Whether --no-verify flag was used
}

// HookResult contains the result of hook execution
type HookResult struct {
	Content  string // Modified content (for content hooks)
	ExitCode int    // Hook exit code
	Output   string // Hook stdout/stderr output
	Aborted  bool   // Whether the operation should be aborted
	Error    error  // Any execution error
}

// Manager handles hook discovery and execution
type Manager struct {
	workspace      *workspace.Workspace
	hooksDir       string
	globalHooksDir string
	enabled        bool
	timeout        time.Duration
}

// NewManager creates a new hook manager for the given workspace
func NewManager(ws *workspace.Workspace) *Manager {
	hooksDir := filepath.Join(ws.JotDir, "hooks")

	// Global hooks directory in user's home
	homeDir, _ := os.UserHomeDir()
	globalHooksDir := filepath.Join(homeDir, ".jot", "hooks")

	return &Manager{
		workspace:      ws,
		hooksDir:       hooksDir,
		globalHooksDir: globalHooksDir,
		enabled:        true,             // Default enabled, configurable in future releases
		timeout:        30 * time.Second, // Default 30s timeout, configurable in future releases
	}
}

// Execute runs hooks for the given context
func (m *Manager) Execute(ctx *HookContext) (*HookResult, error) {
	if !m.enabled || ctx.AllowBypass {
		return &HookResult{Content: ctx.Content}, nil
	}

	// Find all hooks for this type
	hooks, err := m.findHooks(ctx.Type)
	if err != nil {
		return nil, err
	}

	if len(hooks) == 0 {
		return &HookResult{Content: ctx.Content}, nil
	}

	result := &HookResult{Content: ctx.Content}

	// Execute hooks in order
	for _, hookPath := range hooks {
		hookResult, err := m.executeHook(hookPath, ctx, result.Content)
		if err != nil {
			return &HookResult{
				Content: ctx.Content,
				Error:   err,
				Aborted: true,
			}, err
		}

		if hookResult.ExitCode != 0 {
			return &HookResult{
				Content:  ctx.Content,
				ExitCode: hookResult.ExitCode,
				Output:   hookResult.Output,
				Aborted:  true,
			}, fmt.Errorf("hook %s failed with exit code %d", filepath.Base(hookPath), hookResult.ExitCode)
		}

		// Update content for next hook
		result.Content = hookResult.Content
		result.Output += hookResult.Output
	}

	return result, nil
}

// findHooks discovers all hooks for a given type, following git's ordering
func (m *Manager) findHooks(hookType HookType) ([]string, error) {
	var hooks []string

	// Check workspace hooks first (higher priority)
	workspaceHooks, err := m.findHooksInDir(m.hooksDir, hookType)
	if err != nil {
		return nil, err
	}
	hooks = append(hooks, workspaceHooks...)

	// Check global hooks if workspace hooks don't exist
	if len(workspaceHooks) == 0 {
		globalHooks, err := m.findHooksInDir(m.globalHooksDir, hookType)
		if err != nil {
			return nil, err
		}
		hooks = append(hooks, globalHooks...)
	}

	return hooks, nil
}

// findHooksInDir finds hooks in a specific directory
func (m *Manager) findHooksInDir(dir string, hookType HookType) ([]string, error) {
	var hooks []string

	// Check for exact match first
	exactHook := filepath.Join(dir, string(hookType))
	if m.isExecutableHook(exactHook) {
		hooks = append(hooks, exactHook)
	}

	// Check for numbered hooks (git style: pre-commit.01, pre-commit.02)
	pattern := string(hookType) + ".*"
	matches, err := filepath.Glob(filepath.Join(dir, pattern))
	if err != nil {
		return nil, err
	}

	for _, match := range matches {
		// Skip the exact match we already found
		if match == exactHook {
			continue
		}

		if m.isExecutableHook(match) {
			hooks = append(hooks, match)
		}
	}

	return hooks, nil
}

// isExecutableHook checks if a file is an executable hook
func (m *Manager) isExecutableHook(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}

	// Must be a regular file
	if !stat.Mode().IsRegular() {
		return false
	}

	// Must be executable
	return stat.Mode()&0111 != 0
}

// executeHook runs a single hook
func (m *Manager) executeHook(hookPath string, ctx *HookContext, content string) (*HookResult, error) {
	// Create context with timeout
	execCtx, cancel := context.WithTimeout(context.Background(), ctx.Timeout)
	defer cancel()

	// Create command
	cmd := exec.CommandContext(execCtx, hookPath)

	// Set up environment
	cmd.Env = m.buildEnvironment(ctx)

	// Set up stdin with content for content-processing hooks
	if m.isContentHook(ctx.Type) {
		cmd.Stdin = strings.NewReader(content)
	}

	// Capture output
	output, err := cmd.CombinedOutput()

	result := &HookResult{
		ExitCode: cmd.ProcessState.ExitCode(),
		Output:   string(output),
	}

	// For content hooks, use stdout as the new content
	if m.isContentHook(ctx.Type) && result.ExitCode == 0 {
		result.Content = string(output)
	} else {
		result.Content = content
	}

	return result, err
}

// buildEnvironment creates the environment variables for hook execution
func (m *Manager) buildEnvironment(ctx *HookContext) []string {
	env := os.Environ()

	// Standard hook environment
	env = append(env, "JOT_HOOK_TYPE="+string(ctx.Type))
	env = append(env, "JOT_WORKSPACE_ROOT="+ctx.Workspace.Root)
	env = append(env, "JOT_WORKSPACE_INBOX="+ctx.Workspace.InboxPath)
	env = append(env, "JOT_WORKSPACE_LIB="+ctx.Workspace.LibDir)
	env = append(env, "JOT_WORKSPACE_JOTDIR="+ctx.Workspace.JotDir)

	// Context-specific environment
	if ctx.SourceFile != "" {
		env = append(env, "JOT_SOURCE_FILE="+ctx.SourceFile)
	}
	if ctx.DestPath != "" {
		env = append(env, "JOT_DEST_PATH="+ctx.DestPath)
	}
	if ctx.TemplateName != "" {
		env = append(env, "JOT_TEMPLATE_NAME="+ctx.TemplateName)
	}

	// Extra environment variables
	for key, value := range ctx.ExtraEnv {
		env = append(env, key+"="+value)
	}

	return env
}

// isContentHook returns true if this hook type processes content via stdin/stdout
func (m *Manager) isContentHook(hookType HookType) bool {
	switch hookType {
	case PreCapture, PreRefile:
		return true
	default:
		return false
	}
}

// CreateSampleHooks creates sample hook files in the workspace
func (m *Manager) CreateSampleHooks() error {
	if err := os.MkdirAll(m.hooksDir, 0755); err != nil {
		return err
	}

	samples := map[string]string{
		"pre-capture.sample":  samplePreCaptureHook,
		"post-capture.sample": samplePostCaptureHook,
		"pre-refile.sample":   samplePreRefileHook,
		"post-refile.sample":  samplePostRefileHook,
		"pre-archive.sample":  samplePreArchiveHook,
		"post-archive.sample": samplePostArchiveHook,
	}

	for filename, content := range samples {
		path := filepath.Join(m.hooksDir, filename)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if err := os.WriteFile(path, []byte(content), 0644); err != nil {
				return err
			}
		}
	}

	return nil
}

// Sample hook templates
const samplePreCaptureHook = `#!/bin/bash
# Sample pre-capture hook
# This hook is called before content is captured to inbox.md
# Content is passed via stdin, modified content should be written to stdout
# Exit with non-zero to abort the capture

# Read content from stdin
content=$(cat)

# Example: Add timestamp to captures
timestamp=$(date '+%Y-%m-%d %H:%M:%S')
echo "<!-- Captured: $timestamp -->"
echo "$content"

# Exit 0 to continue with capture
exit 0
`

const samplePostCaptureHook = `#!/bin/bash
# Sample post-capture hook
# This hook is called after content has been captured
# Use this for notifications, logging, or triggering other workflows

# Available environment variables:
# JOT_HOOK_TYPE=post-capture
# JOT_WORKSPACE_ROOT=/path/to/workspace
# JOT_TEMPLATE_NAME=template_name (if used)

# Example: Notify when using specific templates
if [[ "$JOT_TEMPLATE_NAME" == "meeting" ]]; then
    echo "Meeting notes captured to $JOT_WORKSPACE_ROOT/inbox.md"
    # Could send notification, update calendar, etc.
fi

exit 0
`

const samplePreRefileHook = `#!/bin/bash
# Sample pre-refile hook
# This hook is called before content is refiled
# Content is passed via stdin, modified content should be written to stdout
# Exit with non-zero to abort the refile

# Read content from stdin
content=$(cat)

# Available environment variables:
# JOT_HOOK_TYPE=pre-refile
# JOT_SOURCE_FILE=source_file
# JOT_DEST_PATH=destination_path

# Example: Add metadata based on destination
if [[ "$JOT_DEST_PATH" =~ "work.md" ]]; then
    echo "<!-- Work-related content -->"
fi

echo "$content"

exit 0
`

const samplePostRefileHook = `#!/bin/bash
# Sample post-refile hook
# This hook is called after content has been refiled

# Available environment variables:
# JOT_HOOK_TYPE=post-refile
# JOT_SOURCE_FILE=source_file
# JOT_DEST_PATH=destination_path

# Example: Log refile operations
echo "$(date): Refiled from $JOT_SOURCE_FILE to $JOT_DEST_PATH" >> "$JOT_WORKSPACE_ROOT/.jot/refile.log"

exit 0
`

const samplePreArchiveHook = `#!/bin/bash
# Sample pre-archive hook
# This hook is called before content is archived
# Exit with non-zero to abort the archive operation

# Available environment variables:
# JOT_HOOK_TYPE=pre-archive
# JOT_WORKSPACE_ROOT=/path/to/workspace
# JOT_SOURCE_FILE=source_file
# JOT_DEST_PATH=destination_path

# Example: Confirm important content archival
if [[ "$JOT_SOURCE_FILE" =~ "inbox.md#important" ]]; then
    echo "Warning: Archiving important content from inbox"
    # Could add additional validation or prompts here
fi

exit 0
`

const samplePostArchiveHook = `#!/bin/bash
# Sample post-archive hook
# This hook is called after content has been archived

# Available environment variables:
# JOT_HOOK_TYPE=post-archive
# JOT_WORKSPACE_ROOT=/path/to/workspace  
# JOT_SOURCE_FILE=source_file
# JOT_DEST_PATH=destination_path

# Example: Log archive operations
echo "$(date): Archived from $JOT_SOURCE_FILE to $JOT_DEST_PATH" >> "$JOT_WORKSPACE_ROOT/.jot/archive.log"

# Example: Clean up or notify
echo "Content archived to: $JOT_DEST_PATH"

exit 0
`

package cmdutil

import (
	"os"
	"path/filepath"

	"github.com/johncoder/jot/internal/workspace"
)

// PathResolver provides standardized file path resolution with workspace context
type PathResolver struct {
	workspace   *workspace.Workspace
	noWorkspace bool
}

// NewPathResolver creates a new path resolver
func NewPathResolver(ws *workspace.Workspace, noWorkspace bool) *PathResolver {
	return &PathResolver{
		workspace:   ws,
		noWorkspace: noWorkspace,
	}
}

// Resolve resolves a file path using workspace context or current directory
func (r *PathResolver) Resolve(filename string) string {
	if r.noWorkspace {
		// Non-workspace mode: resolve relative to current directory
		if filepath.IsAbs(filename) {
			return filename
		}
		cwd, _ := os.Getwd()
		return filepath.Join(cwd, filename)
	}
	
	// Workspace mode: apply workspace-specific logic
	if filename == "inbox.md" && r.workspace != nil {
		return r.workspace.InboxPath
	}
	if filepath.IsAbs(filename) {
		return filename
	}
	if r.workspace != nil {
		return filepath.Join(r.workspace.Root, filename)
	}
	return filename // Fallback for no workspace
}

// ResolveMultiple resolves multiple file paths
func (r *PathResolver) ResolveMultiple(filenames []string) []string {
	resolved := make([]string, len(filenames))
	for i, filename := range filenames {
		resolved[i] = r.Resolve(filename)
	}
	return resolved
}

// ResolveWorkspacePath resolves a path relative to the workspace root (ignores noWorkspace)
func (r *PathResolver) ResolveWorkspacePath(filename string) string {
	if filename == "inbox.md" && r.workspace != nil {
		return r.workspace.InboxPath
	}
	if filepath.IsAbs(filename) {
		return filename
	}
	if r.workspace != nil {
		return filepath.Join(r.workspace.Root, filename)
	}
	return filename
}

// ResolvePath is a convenience function for single-file resolution
func ResolvePath(ws *workspace.Workspace, filename string, noWorkspace bool) string {
	resolver := NewPathResolver(ws, noWorkspace)
	return resolver.Resolve(filename)
}

// ResolveWorkspaceRelativePath resolves a path relative to workspace (ignores noWorkspace flag)
func ResolveWorkspaceRelativePath(ws *workspace.Workspace, filename string) string {
	resolver := NewPathResolver(ws, false)
	return resolver.ResolveWorkspacePath(filename)
}

// PathResolutionOptions configures path resolution behavior
type PathResolutionOptions struct {
	Workspace   *workspace.Workspace
	NoWorkspace bool
	BaseDir     string // Override base directory for non-workspace mode
}

// ResolveWithOptions resolves a path using the provided options
func ResolveWithOptions(filename string, opts *PathResolutionOptions) string {
	if opts == nil {
		opts = &PathResolutionOptions{}
	}
	
	if opts.NoWorkspace {
		// Non-workspace mode with optional base directory override
		if filepath.IsAbs(filename) {
			return filename
		}
		
		baseDir := opts.BaseDir
		if baseDir == "" {
			baseDir, _ = os.Getwd()
		}
		return filepath.Join(baseDir, filename)
	}
	
	// Workspace mode
	if filename == "inbox.md" && opts.Workspace != nil {
		return opts.Workspace.InboxPath
	}
	if filepath.IsAbs(filename) {
		return filename
	}
	if opts.Workspace != nil {
		return filepath.Join(opts.Workspace.Root, filename)
	}
	return filename
}

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

// PathUtil provides workspace-aware path utilities with common operations
type PathUtil struct {
	workspace *workspace.Workspace
}

// NewPathUtil creates a new path utility instance with workspace context
func NewPathUtil(ws *workspace.Workspace) *PathUtil {
	return &PathUtil{workspace: ws}
}

// WorkspaceJoin joins a path relative to the workspace root
func (p *PathUtil) WorkspaceJoin(relativePath string) string {
	if p.workspace == nil {
		return relativePath
	}
	return filepath.Join(p.workspace.Root, relativePath)
}

// LibJoin joins a path relative to the workspace lib directory
func (p *PathUtil) LibJoin(relativePath string) string {
	if p.workspace == nil {
		return relativePath
	}
	return filepath.Join(p.workspace.LibDir, relativePath)
}

// JotDirJoin joins a path relative to the workspace .jot directory
func (p *PathUtil) JotDirJoin(relativePath string) string {
	if p.workspace == nil {
		return relativePath
	}
	return filepath.Join(p.workspace.JotDir, relativePath)
}

// EnsureDir creates a directory if it doesn't exist with standard permissions (0755)
func (p *PathUtil) EnsureDir(dirPath string) error {
	return os.MkdirAll(dirPath, 0755)
}

// EnsureDirForFile creates the parent directory for a file path if it doesn't exist
func (p *PathUtil) EnsureDirForFile(filePath string) error {
	dir := filepath.Dir(filePath)
	return p.EnsureDir(dir)
}

// SafeWriteFile writes content to a file, creating parent directories as needed
func (p *PathUtil) SafeWriteFile(filePath string, content []byte) error {
	if err := p.EnsureDirForFile(filePath); err != nil {
		return err
	}
	return os.WriteFile(filePath, content, 0644)
}

// SafeAppendFile appends content to a file, creating parent directories as needed
func (p *PathUtil) SafeAppendFile(filePath string, content []byte) error {
	if err := p.EnsureDirForFile(filePath); err != nil {
		return err
	}

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(content)
	return err
}

// ToAbsolute converts a path to an absolute path
func (p *PathUtil) ToAbsolute(path string) (string, error) {
	return filepath.Abs(path)
}

// ToWorkspaceRelative converts an absolute path to workspace-relative if possible
func (p *PathUtil) ToWorkspaceRelative(absolutePath string) (string, error) {
	if p.workspace == nil {
		return absolutePath, nil
	}
	return filepath.Rel(p.workspace.Root, absolutePath)
}

// IsWithinWorkspace checks if a path is within the workspace
func (p *PathUtil) IsWithinWorkspace(path string) bool {
	if p.workspace == nil {
		return false
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	relPath, err := filepath.Rel(p.workspace.Root, absPath)
	if err != nil {
		return false
	}

	// If relative path starts with "..", it's outside the workspace
	return !filepath.IsAbs(relPath) && !filepath.HasPrefix(relPath, "..")
}

// SplitPath splits a path into directory, base name, and extension
func (p *PathUtil) SplitPath(path string) (dir, base, ext string) {
	dir = filepath.Dir(path)
	base = filepath.Base(path)
	ext = filepath.Ext(base)
	if ext != "" {
		base = base[:len(base)-len(ext)]
	}
	return dir, base, ext
}

// CreateBackupFile creates a backup of a file and returns the backup path
func (p *PathUtil) CreateBackupFile(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	backupPath := filePath + ".backup"
	err = p.SafeWriteFile(backupPath, content)
	if err != nil {
		return "", err
	}

	return backupPath, nil
}

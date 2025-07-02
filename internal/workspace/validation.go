package workspace

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/johncoder/jot/internal/config"
)

// IsValid checks if a path contains a valid jot workspace
func IsValid(path string) bool {
	// Check if path exists
	if _, err := os.Stat(path); err != nil {
		return false
	}

	// Check if .jot directory exists
	jotDir := filepath.Join(path, ".jot")
	if info, err := os.Stat(jotDir); err != nil || !info.IsDir() {
		return false
	}

	return true
}

// GetNameFromPath attempts to find the workspace name from its path
func GetNameFromPath(path string) string {
	// Try to find workspace name in registry
	workspaces := config.ListWorkspaces()
	for name, wsPath := range workspaces {
		if absPath, err := filepath.Abs(wsPath); err == nil {
			if wsAbsPath, err := filepath.Abs(path); err == nil {
				if absPath == wsAbsPath {
					return name
				}
			}
		}
	}

	// Fall back to directory name
	return filepath.Base(path)
}

// GetDiscoveryMethod determines how the workspace was discovered
func GetDiscoveryMethod(ws *Workspace) string {
	// Check if we're in a workspace directory
	currentDir, _ := os.Getwd()
	if currentDir == ws.Root {
		return "local directory"
	}

	// Check if we're in a subdirectory of the workspace
	if strings.HasPrefix(currentDir, ws.Root) {
		return "local directory"
	}

	// Check if there's a local .jotrc
	if _, err := os.Stat(".jotrc"); err == nil {
		return "local config"
	}

	return "global default"
}

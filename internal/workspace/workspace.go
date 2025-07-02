package workspace

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/johncoder/jot/internal/config"
)

// WorkspaceConfig represents workspace-specific configuration
type WorkspaceConfig struct {
	ArchiveLocation string `json:"archive_location,omitempty"`
}

// Workspace represents a jot workspace
type Workspace struct {
	Root      string
	JotDir    string
	InboxPath string
	LibDir    string
	Config    *WorkspaceConfig
}

// LoadWorkspaceConfig loads workspace-specific configuration from .jot/config.json
func LoadWorkspaceConfig(jotDir string) (*WorkspaceConfig, error) {
	configPath := filepath.Join(jotDir, "config.json")

	// Return default config if file doesn't exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &WorkspaceConfig{
			ArchiveLocation: "archive/archive.md#Archive", // Default simple path
		}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read workspace config: %w", err)
	}

	var cfg WorkspaceConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse workspace config: %w", err)
	}

	// Set default if not specified
	if cfg.ArchiveLocation == "" {
		cfg.ArchiveLocation = "archive/archive.md#Archive"
	}

	return &cfg, nil
}

// SaveWorkspaceConfig saves workspace configuration to .jot/config.json
func (ws *Workspace) SaveWorkspaceConfig() error {
	configPath := filepath.Join(ws.JotDir, "config.json")

	data, err := json.MarshalIndent(ws.Config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal workspace config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write workspace config: %w", err)
	}

	return nil
}

// GetArchiveLocation returns the configured archive location
func (ws *Workspace) GetArchiveLocation() string {
	if ws.Config == nil || ws.Config.ArchiveLocation == "" {
		return "archive/archive.md#Archive"
	}

	return ws.Config.ArchiveLocation
}

// FindWorkspace searches for a jot workspace using the enhanced discovery algorithm:
// 1. Walk up parent directories looking for .jot/ directory or .jotrc file
// 2. If .jot/ found: Use that workspace
// 3. If .jotrc found first: Use the default workspace defined in that config
// 4. If neither found: Check ~/.jotrc for global default workspace
// 5. If no workspace available: Error with clear guidance
func FindWorkspace() (*Workspace, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	// Step 1 & 2: Walk up directories looking for .jot/ or .jotrc
	for {
		// Check for .jot/ directory (local workspace)
		jotDir := filepath.Join(dir, ".jot")
		if info, err := os.Stat(jotDir); err == nil && info.IsDir() {
			// Load workspace configuration
			cfg, err := LoadWorkspaceConfig(jotDir)
			if err != nil {
				return nil, fmt.Errorf("failed to load workspace config: %w", err)
			}

			return &Workspace{
				Root:      dir,
				JotDir:    jotDir,
				InboxPath: filepath.Join(dir, "inbox.md"),
				LibDir:    filepath.Join(dir, "lib"),
				Config:    cfg,
			}, nil
		}

		// Check for .jotrc file (stops the search)
		jotrcPath := filepath.Join(dir, ".jotrc")
		if _, err := os.Stat(jotrcPath); err == nil {
			// Step 3: Use default workspace from this config
			return findWorkspaceFromConfig(jotrcPath)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached the root directory
			break
		}
		dir = parent
	}

	// Step 4: Check ~/.jotrc for global default workspace
	return findWorkspaceFromGlobalConfig()
}

// findWorkspaceFromConfig loads a config file and returns the default workspace
func findWorkspaceFromConfig(configPath string) (*Workspace, error) {
	// For now, load global config since we don't have local config support yet
	// This is a placeholder for future local config support
	return findWorkspaceFromGlobalConfig()
}

// findWorkspaceFromGlobalConfig finds the default workspace from global config
func findWorkspaceFromGlobalConfig() (*Workspace, error) {
	defaultName, defaultPath, err := config.GetDefaultWorkspace()
	if err != nil {
		return nil, fmt.Errorf("no workspace found. Run 'jot init' from the directory you wish to store your notes")
	}

	// Verify the workspace directory exists and has .jot/
	jotDir := filepath.Join(defaultPath, ".jot")
	if info, err := os.Stat(jotDir); err != nil || !info.IsDir() {
		return nil, fmt.Errorf("default workspace %q (%s) is not valid - missing .jot/ directory. Run 'jot init' in %s or set a different default workspace", defaultName, defaultPath, defaultPath)
	}

	// Load workspace configuration
	cfg, err := LoadWorkspaceConfig(jotDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load workspace config: %w", err)
	}

	return &Workspace{
		Root:      defaultPath,
		JotDir:    jotDir,
		InboxPath: filepath.Join(defaultPath, "inbox.md"),
		LibDir:    filepath.Join(defaultPath, "lib"),
		Config:    cfg,
	}, nil
}

// RequireWorkspace finds a workspace or returns an error
func RequireWorkspace() (*Workspace, error) {
	return RequireWorkspaceWithOverride("")
}

// RequireWorkspaceWithOverride finds a workspace, optionally using a specific workspace name override
func RequireWorkspaceWithOverride(workspaceName string) (*Workspace, error) {
	// If workspace name is provided, use specific workspace
	if workspaceName != "" {
		return RequireSpecificWorkspace(workspaceName)
	}

	// Otherwise use normal discovery
	ws, err := FindWorkspace()
	if err != nil {
		return nil, fmt.Errorf("%w\nRun 'jot init' to initialize a workspace", err)
	}
	return ws, nil
}

// RequireSpecificWorkspace finds a workspace by name from the config registry
func RequireSpecificWorkspace(name string) (*Workspace, error) {
	path, err := config.GetWorkspace(name)
	if err != nil {
		return nil, fmt.Errorf("workspace '%s' not found in registry: %w\nUse 'jot workspace list' to see available workspaces", name, err)
	}

	// Validate that the path exists and is initialized
	jotDir := filepath.Join(path, ".jot")
	if info, err := os.Stat(jotDir); err != nil || !info.IsDir() {
		return nil, fmt.Errorf("workspace '%s' is not initialized (missing .jot directory at %s)\nRun 'jot init' in %s to initialize it", name, jotDir, path)
	}

	// Load workspace configuration
	cfg, err := LoadWorkspaceConfig(jotDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load workspace config: %w", err)
	}

	return &Workspace{
		Root:      path,
		JotDir:    jotDir,
		InboxPath: filepath.Join(path, "inbox.md"),
		LibDir:    filepath.Join(path, "lib"),
		Config:    cfg,
	}, nil
}

// IsWorkspace checks if the current directory is a jot workspace
func IsWorkspace() bool {
	_, err := FindWorkspace()
	return err == nil
}

// RelativePath converts an absolute path to a path relative to the workspace root
// If the path is not within the workspace, returns the original path
func (w *Workspace) RelativePath(absolutePath string) string {
	if relPath, err := filepath.Rel(w.Root, absolutePath); err == nil && !strings.HasPrefix(relPath, "..") {
		return relPath
	}
	return absolutePath
}

// AppendToInbox adds content to the inbox with a timestamp
func (w *Workspace) AppendToInbox(content string) error {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	entry := fmt.Sprintf("\n## %s\n\n%s\n\n", timestamp, content)

	file, err := os.OpenFile(w.InboxPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open inbox: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(entry)
	if err != nil {
		return fmt.Errorf("failed to write to inbox: %w", err)
	}

	return nil
}

// InboxExists checks if the inbox file exists
func (w *Workspace) InboxExists() bool {
	_, err := os.Stat(w.InboxPath)
	return err == nil
}

// LibExists checks if the lib directory exists
func (w *Workspace) LibExists() bool {
	info, err := os.Stat(w.LibDir)
	return err == nil && info.IsDir()
}

// AppendToFile appends content to a specified file
func (w *Workspace) AppendToFile(filePath, content string) error {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(content + "\n\n")
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}

// GetWorkspaceContext attempts to find workspace for configuration but allows operation without it
// When noWorkspace is true, workspace detection failures are ignored and nil workspace is returned
func GetWorkspaceContext(noWorkspace bool) (*Workspace, error) {
	if noWorkspace {
		// Try to find workspace for config, but don't fail if not found
		ws, _ := FindWorkspace()
		return ws, nil // Return nil workspace without error if not found
	}
	return RequireWorkspace() // Existing behavior
}

package workspace

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/johncoder/jot/internal/config"
)

// WorkspaceConfig represents workspace-specific configuration
// Added AllowCreateMissingFiles to support creation of missing files
type WorkspaceConfig struct {
	ArchiveLocation         string `json:"archive_location,omitempty"`
	AllowCreateMissingFiles bool   `json:"allow_create_missing_files,omitempty"`
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
			ArchiveLocation:         "archive/archive.md#Archive", // Default simple path
			AllowCreateMissingFiles: false, // Default: do not allow creating missing files
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

	// Set defaults if not specified
	if cfg.ArchiveLocation == "" {
		cfg.ArchiveLocation = "archive/archive.md#Archive"
	}
	// Default for AllowCreateMissingFiles is false

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
	// Initialize config if not already initialized
	if err := config.Initialize(""); err != nil {
		return nil, fmt.Errorf("failed to initialize config: %w", err)
	}

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
	// Initialize config if not already initialized
	if err := config.Initialize(""); err != nil {
		return nil, fmt.Errorf("failed to initialize config: %w", err)
	}

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
	file, err := os.OpenFile(w.InboxPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open inbox: %w", err)
	}
	defer file.Close()

	// Write a separator to ensure new content appears on a new line
	_, err = file.WriteString("\n\n")
	if err != nil {
		return fmt.Errorf("failed to write separator to inbox: %w", err)
	}

	_, err = file.WriteString(content)
	if err != nil {
		return fmt.Errorf("failed to write to inbox: %w", err)
	}

	return nil
}

// PrependToInbox inserts content at the top of the inbox, after the main heading if present
func (w *Workspace) PrependToInbox(content string) error {
	inboxBytes, err := os.ReadFile(w.InboxPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read inbox: %w", err)
	}
	inbox := string(inboxBytes)

	lines := strings.Split(inbox, "\n")
	// Skip leading blank lines
	firstNonBlank := 0
	for ; firstNonBlank < len(lines); firstNonBlank++ {
		if strings.TrimSpace(lines[firstNonBlank]) != "" {
			break
		}
	}
	// Find first top-level heading
	headingIdx := -1
	for i := firstNonBlank; i < len(lines); i++ {
		if strings.HasPrefix(strings.TrimSpace(lines[i]), "# ") {
			headingIdx = i
			break
		}
	}
	insertIdx := 0
	if headingIdx != -1 {
		// Find first blank line after heading
		for i := headingIdx + 1; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) == "" {
				insertIdx = i + 1
				break
			}
		}
		if insertIdx == 0 {
			insertIdx = headingIdx + 1
		}
	} else {
		insertIdx = firstNonBlank
	}
	// Insert content
	newLines := append([]string{}, lines[:insertIdx]...)
	newLines = append(newLines, content, "")
	newLines = append(newLines, lines[insertIdx:]...)
	newContent := strings.Join(newLines, "\n")
	return os.WriteFile(w.InboxPath, []byte(newContent), 0644)
}

// LibExists checks if the lib directory exists
func (w *Workspace) LibExists() bool {
	info, err := os.Stat(w.LibDir)
	return err == nil && info.IsDir()
}

// EnsureFileExists checks if a file exists, and creates it if allowed by config
func (w *Workspace) EnsureFileExists(filePath string) error {
	absPath := filePath
	if !filepath.IsAbs(filePath) {
		absPath = filepath.Join(w.Root, filePath)
	}
	rel, err := filepath.Rel(w.Root, absPath)
	if err != nil || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		return fmt.Errorf("refusing to create file outside workspace: %s", absPath)
	}
	if _, err := os.Stat(absPath); err != nil {
		if os.IsNotExist(err) {
			if w.Config != nil && w.Config.AllowCreateMissingFiles {
				f, err := os.Create(absPath)
				if err != nil {
					return fmt.Errorf("failed to create file: %w", err)
				}
				f.Close()
			} else {
				return fmt.Errorf("file does not exist and creation is not allowed by config: %s", absPath)
			}
		} else {
			return fmt.Errorf("failed to stat file: %w", err)
		}
	}
	return nil
}

// AppendToFile appends content to a specified file
func (w *Workspace) AppendToFile(filePath, content string) error {
	if err := w.EnsureFileExists(filePath); err != nil {
		return err
	}
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

// PrependToFile inserts content at the top of the specified file, after the main heading if present
func (w *Workspace) PrependToFile(filePath, content string) error {
	if err := w.EnsureFileExists(filePath); err != nil {
		return err
	}
	fileBytes, err := os.ReadFile(filePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read file: %w", err)
	}
	fileStr := string(fileBytes)

	lines := strings.Split(fileStr, "\n")
	firstNonBlank := 0
	for ; firstNonBlank < len(lines); firstNonBlank++ {
		if strings.TrimSpace(lines[firstNonBlank]) != "" {
			break
		}
	}
	headingIdx := -1
	for i := firstNonBlank; i < len(lines); i++ {
		if strings.HasPrefix(strings.TrimSpace(lines[i]), "# ") {
			headingIdx = i
			break
		}
	}
	insertIdx := 0
	if headingIdx != -1 {
		for i := headingIdx + 1; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) == "" {
				insertIdx = i + 1
				break
			}
		}
		if insertIdx == 0 {
			insertIdx = headingIdx + 1
		}
	} else {
		insertIdx = firstNonBlank
	}
	newLines := append([]string{}, lines[:insertIdx]...)
	newLines = append(newLines, content, "")
	newLines = append(newLines, lines[insertIdx:]...)
	newContent := strings.Join(newLines, "\n")
	return os.WriteFile(filePath, []byte(newContent), 0644)
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

// InboxExists checks if the inbox file exists
func (w *Workspace) InboxExists() bool {
	_, err := os.Stat(w.InboxPath)
	return err == nil
}

// ResolveFilePath resolves a user-supplied file path to an absolute path within the workspace.
// If ensureExists is true, it will create the file if it does not exist.
func (ws *Workspace) ResolveFilePath(path string, ensureExists bool) (string, error) {
	if path == "" || path == "inbox.md" {
		if ensureExists {
			if err := ws.EnsureFileExists(ws.InboxPath); err != nil {
				return "", err
			}
		}
		return ws.InboxPath, nil
	}
	absPath := path
	if !filepath.IsAbs(path) {
		absPath = filepath.Join(ws.Root, path)
	}
	if ensureExists {
		if err := ws.EnsureFileExists(absPath); err != nil {
			return "", err
		}
	}
	return absPath, nil
}

package cmdutil

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/johncoder/jot/internal/config"
	"github.com/spf13/cobra"
)

// ConfigManager provides standardized configuration management for commands
type ConfigManager struct {
	configFile string
	initialized bool
}

// NewConfigManager creates a new config manager
func NewConfigManager() *ConfigManager {
	return &ConfigManager{}
}

// NewConfigManagerWithFile creates a config manager with a specific config file
func NewConfigManagerWithFile(configFile string) *ConfigManager {
	return &ConfigManager{
		configFile: configFile,
	}
}

// Initialize initializes the config system with the specified config file
func (cm *ConfigManager) Initialize(configFile string) error {
	if cm.initialized {
		return nil // Already initialized
	}
	
	cm.configFile = configFile
	if err := config.Initialize(configFile); err != nil {
		return fmt.Errorf("failed to initialize config: %w", err)
	}
	
	cm.initialized = true
	return nil
}

// InitializeFromCommand extracts config file from command flags and initializes
func (cm *ConfigManager) InitializeFromCommand(cmd *cobra.Command) error {
	configFile, _ := cmd.Flags().GetString("config")
	return cm.Initialize(configFile)
}

// EnsureInitialized ensures the config system is initialized
func (cm *ConfigManager) EnsureInitialized() error {
	if !cm.initialized {
		return cm.Initialize(cm.configFile)
	}
	return nil
}

// GetConfigFile returns the current config file path
func (cm *ConfigManager) GetConfigFile() string {
	return cm.configFile
}

// IsInitialized returns true if the config system has been initialized
func (cm *ConfigManager) IsInitialized() bool {
	return cm.initialized
}

// WorkspaceManager provides workspace-specific config operations
type WorkspaceManager struct {
	*ConfigManager
}

// NewWorkspaceManager creates a workspace manager with config support
func NewWorkspaceManager() *WorkspaceManager {
	return &WorkspaceManager{
		ConfigManager: NewConfigManager(),
	}
}

// ListWorkspaces returns all registered workspaces
func (wm *WorkspaceManager) ListWorkspaces() map[string]string {
	wm.EnsureInitialized()
	return config.ListWorkspaces()
}

// GetDefaultWorkspace returns the default workspace information
func (wm *WorkspaceManager) GetDefaultWorkspace() (string, string, error) {
	wm.EnsureInitialized()
	return config.GetDefaultWorkspace()
}

// AddWorkspace adds a new workspace to the configuration
func (wm *WorkspaceManager) AddWorkspace(name, path string) error {
	wm.EnsureInitialized()
	
	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path: %w", err)
	}
	
	// Check if workspace already exists
	if existingPath, err := config.GetWorkspace(name); err == nil {
		return fmt.Errorf("workspace '%s' already exists at '%s'", name, existingPath)
	}
	
	return config.AddWorkspace(name, absPath)
}

// RemoveWorkspace removes a workspace from the configuration
func (wm *WorkspaceManager) RemoveWorkspace(name string) error {
	wm.EnsureInitialized()
	return config.RemoveWorkspace(name)
}

// SetDefaultWorkspace sets the default workspace
func (wm *WorkspaceManager) SetDefaultWorkspace(name string) error {
	wm.EnsureInitialized()
	return config.SetDefaultWorkspace(name)
}

// GetWorkspace returns the path for a named workspace
func (wm *WorkspaceManager) GetWorkspace(name string) (string, error) {
	wm.EnsureInitialized()
	return config.GetWorkspace(name)
}

// Convenience functions for common config patterns

// InitializeConfig is a convenience function for simple config initialization
func InitializeConfig(cmd *cobra.Command) error {
	configFile, _ := cmd.Flags().GetString("config")
	return config.Initialize(configFile)
}

// InitializeConfigWithError wraps config initialization with cmdutil error handling
func InitializeConfigWithError(ctx *CommandContext) error {
	configFile, _ := ctx.Cmd.Flags().GetString("config")
	if err := config.Initialize(configFile); err != nil {
		return ctx.HandleError(fmt.Errorf("failed to initialize config: %w", err))
	}
	return nil
}

// CreateDefaultWorkspaceConfig creates a default workspace configuration file
func CreateDefaultWorkspaceConfig(jotDir string) (string, []byte, error) {
	configPath := filepath.Join(jotDir, "config.json")
	configContent := `{
  "archive_location": "archive/archive.md#Archive"
}
`
	
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return "", nil, fmt.Errorf("failed to create workspace config: %w", err)
	}
	
	return configPath, []byte(configContent), nil
}

// ConfiguredWorkspaceManager returns a pre-configured workspace manager
func ConfiguredWorkspaceManager(cmd *cobra.Command) (*WorkspaceManager, error) {
	wm := NewWorkspaceManager()
	if err := wm.InitializeFromCommand(cmd); err != nil {
		return nil, err
	}
	return wm, nil
}

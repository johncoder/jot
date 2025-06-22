package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/viper"
	"github.com/titanous/json5"
)

// Config holds the application configuration
type Config struct {
	// Workspaces maps workspace names to their absolute paths
	Workspaces map[string]string `json:"workspaces,omitempty"`

	// Default is the default workspace name to use when not in a local workspace
	Default string `json:"default,omitempty"`

	// Legacy support for old configuration format
	Locations       map[string]string `json:"locations,omitempty"`       // Deprecated
	DefaultLocation string            `json:"defaultLocation,omitempty"` // Deprecated
}

var globalConfig *Config
var configFilePath string

// Initialize loads configuration from files and environment variables
func Initialize(cfgFile string) error {
	v := viper.New()

	// Set config name and type
	v.SetConfigName(".jotrc")
	v.SetConfigType("json")

	// Add config paths
	if cfgFile != "" {
		// Use config file from the flag
		v.SetConfigFile(cfgFile)
		configFilePath = cfgFile
	} else {
		// Look for config in home directory and current directory
		home, err := os.UserHomeDir()
		if err == nil {
			v.AddConfigPath(home)
		}
		v.AddConfigPath(".")
	}

	// Environment variable overrides
	v.SetEnvPrefix("JOT")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Set defaults
	setDefaults(v)

	// Try to read config file with JSON5 support
	var cfg Config
	if configData, err := readConfigFile(v); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Config file was found but another error was produced
			return fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found; using defaults and environment variables
		if err := v.Unmarshal(&cfg); err != nil {
			return fmt.Errorf("error unmarshaling default config: %w", err)
		}
	} else {
		// Parse JSON5 and apply environment variable substitution
		expandedData := expandEnvironmentVariables(string(configData))
		if err := json5.Unmarshal([]byte(expandedData), &cfg); err != nil {
			// Fall back to regular JSON if JSON5 parsing fails
			if err := json.Unmarshal([]byte(expandedData), &cfg); err != nil {
				return fmt.Errorf("error parsing config file (tried JSON5 and JSON): %w", err)
			}
		}
	}

	// Store the config file path if found
	if configFilePath == "" && v.ConfigFileUsed() != "" {
		configFilePath = v.ConfigFileUsed()
	}

	// Migrate legacy configuration format
	migrateLegacyConfig(&cfg)

	globalConfig = &cfg
	return nil
}

// readConfigFile attempts to read the configuration file
func readConfigFile(v *viper.Viper) ([]byte, error) {
	if configFilePath != "" {
		// Use explicit config file path
		return ioutil.ReadFile(configFilePath)
	}

	// Try to find config file in search paths
	paths := v.AllSettings()
	_ = paths // Suppress unused variable for now

	// Check each potential config file location
	home, _ := os.UserHomeDir()
	searchPaths := []string{
		".jotrc",
		filepath.Join(home, ".jotrc"),
	}

	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			configFilePath = path
			return ioutil.ReadFile(path)
		}
	}

	return nil, viper.ConfigFileNotFoundError{}
}

// expandEnvironmentVariables replaces ${VAR} and ${VAR:-default} patterns with environment variable values
func expandEnvironmentVariables(data string) string {
	// Pattern to match ${VAR} or ${VAR:-default}
	re := regexp.MustCompile(`\$\{([^}]+)\}`)

	return re.ReplaceAllStringFunc(data, func(match string) string {
		// Remove ${ and }
		varExpr := match[2 : len(match)-1]

		// Check for default value syntax (VAR:-default)
		parts := strings.SplitN(varExpr, ":-", 2)
		varName := parts[0]

		value := os.Getenv(varName)
		if value == "" && len(parts) == 2 {
			// Use default value if variable is empty or unset
			value = parts[1]
		}

		return value
	})
}

// migrateLegacyConfig converts old configuration format to new format
func migrateLegacyConfig(cfg *Config) {
	// If new format is empty but legacy format exists, migrate
	if len(cfg.Workspaces) == 0 && len(cfg.Locations) > 0 {
		cfg.Workspaces = make(map[string]string)
		for name, path := range cfg.Locations {
			// Expand path for migration
			if strings.HasPrefix(path, "~/") {
				if home, err := os.UserHomeDir(); err == nil {
					path = filepath.Join(home, path[2:])
				}
			}
			cfg.Workspaces[name] = path
		}

		if cfg.Default == "" && cfg.DefaultLocation != "" {
			cfg.Default = cfg.DefaultLocation
		}
	}
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// Legacy defaults (for backward compatibility)
	v.SetDefault("defaultLocation", "default")
	home, _ := os.UserHomeDir()
	defaultPath := filepath.Join(home, "Documents", "notes")
	v.SetDefault("locations.default", defaultPath)
}

// Get returns the global configuration
func Get() *Config {
	if globalConfig == nil {
		// Return default config if not initialized
		home, _ := os.UserHomeDir()
		return &Config{
			Default: "notes",
			Workspaces: map[string]string{
				"notes": filepath.Join(home, "Documents", "notes"),
			},
			// Legacy support
			DefaultLocation: "default",
			Locations: map[string]string{
				"default": filepath.Join(home, "Documents", "notes"),
			},
		}
	}
	return globalConfig
}

// getDefaultNotesPath returns the default notes directory path
func getDefaultNotesPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return filepath.Join(home, "Documents", "notes")
}

// GetWorkspace returns the path for a named workspace
func GetWorkspace(name string) (string, error) {
	cfg := Get()

	if name == "" {
		name = cfg.Default
	}

	if path, exists := cfg.Workspaces[name]; exists {
		return expandPath(path), nil
	}

	return "", fmt.Errorf("workspace %q not found in configuration", name)
}

// GetDefaultWorkspace returns the name and path of the default workspace
func GetDefaultWorkspace() (string, string, error) {
	cfg := Get()

	if cfg.Default == "" {
		return "", "", fmt.Errorf("no default workspace configured")
	}

	path, err := GetWorkspace(cfg.Default)
	if err != nil {
		return "", "", err
	}

	return cfg.Default, path, nil
}

// AddWorkspace adds a new workspace to the configuration
func AddWorkspace(name, path string) error {
	cfg := Get()

	// Initialize workspaces map if nil
	if cfg.Workspaces == nil {
		cfg.Workspaces = make(map[string]string)
	}

	// Check if workspace name already exists
	if _, exists := cfg.Workspaces[name]; exists {
		return fmt.Errorf("workspace %q already exists", name)
	}

	// Add the workspace
	cfg.Workspaces[name] = path

	// Set as default if this is the first workspace
	if cfg.Default == "" {
		cfg.Default = name
	}

	// Save the configuration
	return SaveConfig()
}

// SetDefaultWorkspace sets the default workspace
func SetDefaultWorkspace(name string) error {
	cfg := Get()

	// Check if workspace exists
	if _, exists := cfg.Workspaces[name]; !exists {
		return fmt.Errorf("workspace %q does not exist", name)
	}

	cfg.Default = name
	return SaveConfig()
}

// RemoveWorkspace removes a workspace from the configuration
func RemoveWorkspace(name string) error {
	cfg := Get()

	// Check if workspace exists
	if _, exists := cfg.Workspaces[name]; !exists {
		return fmt.Errorf("workspace %q does not exist", name)
	}

	// Remove the workspace
	delete(cfg.Workspaces, name)

	// If this was the default, clear the default
	if cfg.Default == name {
		cfg.Default = ""
		// Set a new default if there are remaining workspaces
		for workspaceName := range cfg.Workspaces {
			cfg.Default = workspaceName
			break
		}
	}

	return SaveConfig()
}

// ListWorkspaces returns all configured workspaces
func ListWorkspaces() map[string]string {
	cfg := Get()
	if cfg.Workspaces == nil {
		return make(map[string]string)
	}

	// Return a copy to prevent external modification
	result := make(map[string]string)
	for name, path := range cfg.Workspaces {
		result[name] = expandPath(path)
	}
	return result
}

// SaveConfig saves the current configuration to file
func SaveConfig() error {
	if configFilePath == "" {
		// Determine config file path
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("unable to determine home directory: %w", err)
		}
		configFilePath = filepath.Join(home, ".jotrc")
	}

	cfg := Get()

	// Create a JSON5-formatted config with comments
	configJSON5 := formatConfigAsJSON5(cfg)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(configFilePath), 0755); err != nil {
		return fmt.Errorf("unable to create config directory: %w", err)
	}

	// Write the config file
	if err := ioutil.WriteFile(configFilePath, []byte(configJSON5), 0644); err != nil {
		return fmt.Errorf("unable to write config file: %w", err)
	}

	return nil
}

// formatConfigAsJSON5 formats the configuration as JSON5 with helpful comments
func formatConfigAsJSON5(cfg *Config) string {
	var builder strings.Builder

	builder.WriteString("{\n")
	builder.WriteString("  // Workspace registry - maps names to absolute paths\n")
	builder.WriteString("  \"workspaces\": {\n")

	first := true
	for name, path := range cfg.Workspaces {
		if !first {
			builder.WriteString(",\n")
		}
		builder.WriteString(fmt.Sprintf("    \"%s\": \"%s\"", name, path))
		first = false
	}

	builder.WriteString("\n  }")

	if cfg.Default != "" {
		builder.WriteString(",\n\n")
		builder.WriteString("  // Which workspace to use when not in a local workspace\n")
		builder.WriteString(fmt.Sprintf("  \"default\": \"%s\"", cfg.Default))
	}

	builder.WriteString("\n}\n")

	return builder.String()
}

// expandPath expands ~ and environment variables in paths
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			path = filepath.Join(home, path[2:])
		}
	}
	return os.ExpandEnv(path)
}

// GetLocation returns the path for a named location (legacy support)
func GetLocation(name string) (string, error) {
	cfg := Get()

	// Try new workspace format first
	if len(cfg.Workspaces) > 0 {
		return GetWorkspace(name)
	}

	// Fall back to legacy format
	if name == "" {
		name = cfg.DefaultLocation
	}

	if path, exists := cfg.Locations[name]; exists {
		return expandPath(path), nil
	}

	return "", fmt.Errorf("location %q not found in configuration", name)
}

// GetEditor returns the configured editor command
// Follows Git's priority: $EDITOR -> $VISUAL -> default
func GetEditor() string {
	// Check environment variables first (like Git does)
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}
	if visual := os.Getenv("VISUAL"); visual != "" {
		return visual
	}

	// Fall back to reasonable defaults
	return "vi"
}

// GetPager returns the configured pager command
// Follows Git's priority: $PAGER -> default
func GetPager() string {
	// Check environment variable first
	if pager := os.Getenv("PAGER"); pager != "" {
		return pager
	}

	// Fall back to default
	return "less"
}

// Workspace registry management functions

// RegisterWorkspace adds a new workspace to the global registry
func RegisterWorkspace(name, path string) error {
	if name == "" {
		return fmt.Errorf("workspace name cannot be empty")
	}
	if path == "" {
		return fmt.Errorf("workspace path cannot be empty")
	}

	// Load configuration from file to get the current state
	cfg, err := loadConfigFromFile()
	if err != nil {
		// If no config file exists, start with empty config
		cfg = &Config{
			Workspaces: make(map[string]string),
		}
	}

	// Initialize workspaces map if needed
	if cfg.Workspaces == nil {
		cfg.Workspaces = make(map[string]string)
	}

	// Check for existing workspace
	if existingPath, exists := cfg.Workspaces[name]; exists {
		if existingPath == path {
			// Same workspace at same path - check if it actually exists
			jotDir := filepath.Join(path, ".jot")
			if _, err := os.Stat(jotDir); err == nil {
				// Workspace already exists and is initialized
				return fmt.Errorf("workspace %q already exists at %s", name, existingPath)
			}
			// Workspace is registered but not initialized - allow re-initialization
			return nil
		}
		return fmt.Errorf("workspace %q already exists at %s", name, existingPath)
	}

	// Add new workspace
	cfg.Workspaces[name] = path

	// Set as default if this is the first workspace
	if cfg.Default == "" || len(cfg.Workspaces) == 1 {
		cfg.Default = name
	}

	// Save configuration
	return saveConfig(cfg)
}

// loadConfigFromFile loads configuration directly from the file system
func loadConfigFromFile() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	configPath := filepath.Join(home, ".jotrc")

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file does not exist")
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Try JSON5 first
	var cfg Config
	if err := json5.Unmarshal(data, &cfg); err != nil {
		// Fallback to regular JSON
		if jsonErr := json.Unmarshal(data, &cfg); jsonErr != nil {
			return nil, fmt.Errorf("failed to parse config file (tried JSON5 and JSON): %w", err)
		}
	}

	return &cfg, nil
}

// saveConfig writes the configuration back to the global config file
func saveConfig(cfg *Config) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	configPath := filepath.Join(home, ".jotrc")

	// Convert to JSON5 with comments
	configData := generateConfigJSON5(cfg)

	if err := os.WriteFile(configPath, []byte(configData), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// Update global config
	globalConfig = cfg
	return nil
}

// generateConfigJSON5 creates a nicely formatted JSON5 config with comments
func generateConfigJSON5(cfg *Config) string {
	var builder strings.Builder

	builder.WriteString("{\n")
	builder.WriteString("  // Workspace registry - maps names to absolute paths\n")
	builder.WriteString("  \"workspaces\": {\n")

	first := true
	for name, path := range cfg.Workspaces {
		if !first {
			builder.WriteString(",\n")
		}
		builder.WriteString(fmt.Sprintf("    \"%s\": \"%s\"", name, path))
		first = false
	}

	builder.WriteString("\n  },\n\n")

	if cfg.Default != "" {
		builder.WriteString(",\n\n")
		builder.WriteString(fmt.Sprintf("  // Default workspace to use when not in a local workspace\n"))
		builder.WriteString(fmt.Sprintf("  \"default\": \"%s\"", cfg.Default))
	}

	builder.WriteString("\n}\n")

	return builder.String()
}

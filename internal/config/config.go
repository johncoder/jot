package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	// Locations maps named locations to their paths
	Locations map[string]string `json:"locations,omitempty"`

	// DefaultLocation is the default location name to use
	DefaultLocation string `json:"defaultLocation,omitempty"`

	// Editor specifies the preferred editor command
	Editor string `json:"editor,omitempty"`

	// Pager specifies the preferred pager command
	Pager string `json:"pager,omitempty"`
}

var globalConfig *Config

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

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Config file was found but another error was produced
			return fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found; using defaults and environment variables
	}

	// Unmarshal config
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return fmt.Errorf("error unmarshaling config: %w", err)
	}

	globalConfig = &cfg
	return nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// Default editor from environment or fallback
	if editor := os.Getenv("EDITOR"); editor != "" {
		v.SetDefault("editor", editor)
	} else if visual := os.Getenv("VISUAL"); visual != "" {
		v.SetDefault("editor", visual)
	} else {
		// Platform-specific defaults
		v.SetDefault("editor", getDefaultEditor())
	}

	// Default pager from environment or fallback
	if pager := os.Getenv("PAGER"); pager != "" {
		v.SetDefault("pager", pager)
	} else {
		v.SetDefault("pager", "less")
	}

	// Default location
	v.SetDefault("defaultLocation", "default")

	// Default locations
	home, _ := os.UserHomeDir()
	defaultPath := filepath.Join(home, "notes")
	v.SetDefault("locations.default", defaultPath)
}

// getDefaultEditor returns platform-specific default editor
func getDefaultEditor() string {
	// Try common editors in order of preference
	editors := []string{"nano", "vim", "vi"}

	for _, editor := range editors {
		if _, err := os.Stat(editor); err == nil {
			return editor
		}
	}

	// Last resort - notepad on Windows, vi elsewhere
	if isWindows() {
		return "notepad"
	}
	return "vi"
}

// isWindows checks if running on Windows
func isWindows() bool {
	return os.PathSeparator == '\\'
}

// Get returns the global configuration
func Get() *Config {
	if globalConfig == nil {
		// Return default config if not initialized
		return &Config{
			Locations: map[string]string{
				"default": getDefaultNotesPath(),
			},
			DefaultLocation: "default",
			Editor:          getDefaultEditor(),
			Pager:           "less",
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
	return filepath.Join(home, "notes")
}

// GetLocation returns the path for a named location
func GetLocation(name string) (string, error) {
	cfg := Get()

	if name == "" {
		name = cfg.DefaultLocation
	}

	if path, exists := cfg.Locations[name]; exists {
		return path, nil
	}

	return "", fmt.Errorf("location %q not found in configuration", name)
}

// GetEditor returns the configured editor command
func GetEditor() string {
	cfg := Get()
	if cfg.Editor != "" {
		return cfg.Editor
	}
	return getDefaultEditor()
}

// GetPager returns the configured pager command
func GetPager() string {
	cfg := Get()
	if cfg.Pager != "" {
		return cfg.Pager
	}
	return "less"
}

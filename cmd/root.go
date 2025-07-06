package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/johncoder/jot/internal/workspace"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile       string
	workspaceName string
	version       = "dev"
	buildTime     = "unknown"
	gitCommit     = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "jot",
	Short: "A git-inspired CLI tool for managing notes",
	Long: `jot is a CLI tool for capturing, refiling, archiving, finding, and 
maintaining a hub of notes and information. Designed for tech industry 
knowledge workers who need notes close to their terminal workflow.

Examples:
  jot init              # Initialize a new jot workspace
  jot capture           # Capture a new note
  jot refile            # Move notes from inbox to organized files
  jot archive           # Initialize archive structure
  jot find <query>      # Search through your notes
  jot status            # Show workspace status
  jot doctor            # Diagnose and fix common issues`,
}

func Execute() error {
	// Custom execution to handle external commands
	args := os.Args[1:]

	// If no arguments, show help
	if len(args) == 0 {
		return rootCmd.Help()
	}

	// Check if this is a known command first
	if cmd, _, err := rootCmd.Find(args); err == nil && cmd != rootCmd {
		// It's a known command, let cobra handle it normally
		return rootCmd.Execute()
	}

	// Not a known command, try external command with proper global flag handling
	if err := TryExecuteExternalCommand(args); err != nil {
		// If it's marked as a built-in command, handle normally
		if err.Error() == "built-in command" {
			return rootCmd.Execute()
		}
		// For other errors (external command not found), let cobra handle to show error
		return rootCmd.Execute()
	}

	// External command executed successfully
	return nil
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.jotrc)")
	rootCmd.PersistentFlags().StringVarP(&workspaceName, "workspace", "w", "", "use specific workspace (bypasses discovery)")
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "output in JSON format")

	// Version handling - format output according to Linux CLI conventions
	if version == "dev" || version == "" || !strings.HasPrefix(version, "v") {
		// Development builds show more detail
		rootCmd.Version = fmt.Sprintf("%s (build: %s, commit: %s)", version, buildTime, gitCommit)
	} else {
		// Release builds show clean version, with build info available via verbose flag
		rootCmd.Version = version
	}

	// Add all commands
	addCommands()
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err == nil {
			viper.AddConfigPath(home)
		}
		viper.AddConfigPath(".")
		viper.SetConfigName(".jotrc")
		viper.SetConfigType("json")
	}

	viper.SetEnvPrefix("JOT")
	viper.AutomaticEnv()

	// Set defaults
	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		// Config file not found is okay for now
	}
}

func setDefaults() {
	// Default editor
	if editor := os.Getenv("EDITOR"); editor != "" {
		viper.SetDefault("editor", editor)
	} else if visual := os.Getenv("VISUAL"); visual != "" {
		viper.SetDefault("editor", visual)
	} else {
		viper.SetDefault("editor", "vim")
	}

	// Default pager
	if pager := os.Getenv("PAGER"); pager != "" {
		viper.SetDefault("pager", pager)
	} else {
		viper.SetDefault("pager", "less")
	}
}

func addCommands() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(captureCmd)
	rootCmd.AddCommand(refileCmd)
	rootCmd.AddCommand(archiveCmd)
	rootCmd.AddCommand(findCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(doctorCmd)
	rootCmd.AddCommand(templateCmd)
	rootCmd.AddCommand(evalCmd)
	rootCmd.AddCommand(evaluatorCmd)
	rootCmd.AddCommand(tangleCmd)
	rootCmd.AddCommand(workspaceCmd)
	rootCmd.AddCommand(hooksCmd)
}

// getWorkspace returns a workspace using the global workspace flag override if provided
func getWorkspace(cmd *cobra.Command) (*workspace.Workspace, error) {
	workspaceName, _ := cmd.Flags().GetString("workspace")
	return workspace.RequireWorkspaceWithOverride(workspaceName)
}

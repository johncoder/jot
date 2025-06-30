package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/johncoder/jot/internal/config"
	"github.com/johncoder/jot/internal/workspace"
	"github.com/spf13/cobra"
)

// ExternalCommandContext holds the parsed global flags and workspace information
// for passing to external commands via environment variables
type ExternalCommandContext struct {
	Workspace        *workspace.Workspace
	WorkspaceName    string
	ConfigFile       string
	JSONOutput       bool
	DiscoveryMethod  string
	Subcommand       string
	RemainingArgs    []string
}

// ParseGlobalFlagsForExternal parses only the global flags from the command line
// and returns the remaining arguments and extracted flag values
func ParseGlobalFlagsForExternal(args []string) (*ExternalCommandContext, error) {
	// Create a temporary command just for parsing global flags
	tempCmd := &cobra.Command{
		Use:                "jot",
		DisableFlagParsing: false,
		SilenceErrors:      true,
		SilenceUsage:       true,
	}

	// Add the same global flags as the root command
	var configFile string
	var workspaceName string

	tempCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default is $HOME/.jotrc)")
	tempCmd.PersistentFlags().StringVarP(&workspaceName, "workspace", "w", "", "use specific workspace (bypasses discovery)")
	tempCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "output in JSON format")

	// Parse the flags
	if err := tempCmd.ParseFlags(args); err != nil {
		// If flag parsing fails, treat as external command with no global flags
		if len(args) > 0 {
			return &ExternalCommandContext{
				Subcommand:    args[0],
				RemainingArgs: args[1:],
			}, nil
		}
		return nil, fmt.Errorf("no arguments provided")
	}

	// Get the remaining args after flags are parsed
	remainingArgs := tempCmd.Flags().Args()
	if len(remainingArgs) == 0 {
		return nil, fmt.Errorf("no subcommand provided")
	}

	ctx := &ExternalCommandContext{
		ConfigFile:    configFile,
		JSONOutput:    jsonOutput,
		Subcommand:    remainingArgs[0],
		RemainingArgs: remainingArgs[1:],
	}

	// Initialize config if a custom config file was specified
	if configFile != "" {
		if err := config.Initialize(configFile); err != nil {
			// Don't fail on config errors for external commands
			// They might have their own way of handling configuration
		}
	}

	// Resolve workspace context using the parsed flags
	if err := ctx.resolveWorkspaceContext(workspaceName); err != nil {
		// For external commands, workspace resolution failures are not fatal
		// Extensions might work without workspace context
		ctx.DiscoveryMethod = "none"
	}

	return ctx, nil
}

// resolveWorkspaceContext attempts to find the workspace using the provided override
func (ctx *ExternalCommandContext) resolveWorkspaceContext(workspaceOverride string) error {
	var ws *workspace.Workspace
	var err error

	if workspaceOverride != "" {
		// Use specific workspace from flag
		ws, err = workspace.RequireSpecificWorkspace(workspaceOverride)
		if err != nil {
			return fmt.Errorf("workspace '%s' not found: %w", workspaceOverride, err)
		}
		ctx.DiscoveryMethod = "workspace flag"
		ctx.WorkspaceName = workspaceOverride
	} else {
		// Use normal workspace discovery
		ws, err = workspace.FindWorkspace()
		if err != nil {
			return fmt.Errorf("no workspace found: %w", err)
		}
		ctx.DiscoveryMethod = determineDiscoveryMethod(ws)
		ctx.WorkspaceName = getWorkspaceNameFromPath(ws.Root)
	}

	ctx.Workspace = ws
	return nil
}

// BuildEnvironment creates the environment variables for the external command
func (ctx *ExternalCommandContext) BuildEnvironment() []string {
	env := os.Environ()

	// Always set command context
	env = append(env, "JOT_SUBCOMMAND="+ctx.Subcommand)
	env = append(env, "JOT_JSON_OUTPUT="+strconv.FormatBool(ctx.JSONOutput))

	// Set config file if specified
	if ctx.ConfigFile != "" {
		env = append(env, "JOT_CONFIG_FILE="+ctx.ConfigFile)
	}

	// Set workspace information if available
	if ctx.Workspace != nil {
		env = append(env,
			"JOT_WORKSPACE_ROOT="+ctx.Workspace.Root,
			"JOT_WORKSPACE_NAME="+ctx.WorkspaceName,
			"JOT_WORKSPACE_INBOX="+ctx.Workspace.InboxPath,
			"JOT_WORKSPACE_LIB="+ctx.Workspace.LibDir,
			"JOT_WORKSPACE_JOTDIR="+ctx.Workspace.JotDir,
		)
	}

	// Set discovery method for debugging
	env = append(env, "JOT_DISCOVERY_METHOD="+ctx.DiscoveryMethod)

	// Add resolved configuration values
	env = append(env,
		"JOT_EDITOR="+config.GetEditor(),
		"JOT_PAGER="+config.GetPager(),
	)

	return env
}

// ExecuteExternalCommand finds and executes an external jot command with proper context
func ExecuteExternalCommand(ctx *ExternalCommandContext) error {
	externalCmd := "jot-" + ctx.Subcommand

	// Look for the external command in PATH
	externalPath, err := exec.LookPath(externalCmd)
	if err != nil {
		return fmt.Errorf("external command '%s' not found in PATH", externalCmd)
	}

	// Create the command with remaining arguments
	cmd := exec.Command(externalPath, ctx.RemainingArgs...)

	// Set up I/O
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Set environment with jot context
	cmd.Env = ctx.BuildEnvironment()

	// Execute the command
	return cmd.Run()
}

// IsBuiltinCommand checks if a command is handled by jot's built-in commands
func IsBuiltinCommand(subcommand string) bool {
	// Check if the root command can find this subcommand
	if cmd, _, err := rootCmd.Find([]string{subcommand}); err == nil && cmd != rootCmd {
		return true
	}
	return false
}

// TryExecuteExternalCommand attempts to parse global flags and execute an external command
// Returns nil if this should be handled as a built-in command, or an error if execution failed
func TryExecuteExternalCommand(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no arguments provided")
	}

	// Quick check if this is obviously a built-in command
	if IsBuiltinCommand(args[0]) {
		return fmt.Errorf("built-in command") // Signal that this should be handled normally
	}

	// Parse global flags and extract context
	ctx, err := ParseGlobalFlagsForExternal(args)
	if err != nil {
		return err
	}

	// Double-check that this isn't a built-in command after flag parsing
	if IsBuiltinCommand(ctx.Subcommand) {
		return fmt.Errorf("built-in command") // Signal that this should be handled normally
	}

	// Try to execute as external command
	return ExecuteExternalCommand(ctx)
}

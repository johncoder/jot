package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/johncoder/jot/internal/cmdutil"
	"github.com/johncoder/jot/internal/config"
	"github.com/johncoder/jot/internal/workspace"
	"github.com/spf13/cobra"
)

var workspaceCmd = &cobra.Command{
	Use:   "workspace",
	Short: "Manage jot workspaces",
	Long: `Manage jot workspaces - list, add, remove, and set default workspaces.

The workspace command allows you to manage your registered workspaces from anywhere
on your filesystem. Each workspace represents a collection of notes and can be
accessed globally using the --workspace flag.

Examples:
  jot workspace                    # Show current workspace path
  jot workspace list              # List all registered workspaces
  jot workspace add notes ~/notes # Add a workspace named 'notes'
  jot workspace remove old-proj  # Remove a workspace
  jot workspace default notes    # Set default workspace`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Show current workspace path only (for piping to other commands)
		return workspaceShowPath(cmd)
	},
}

var workspaceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registered workspaces",
	Long: `List all workspaces registered in the configuration file.

Shows workspace name, path, status (valid/invalid), and which workspace is
currently active and set as default.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return workspaceList(cmd)
	},
}

var workspaceAddCmd = &cobra.Command{
	Use:   "add <name> <path>",
	Short: "Add a workspace to the registry",
	Long: `Add a workspace to the registry with the specified name and path.

The path must exist and contain a .jot directory (be initialized).
If this is the first workspace being added, it will be set as the default.`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return workspaceAdd(cmd, args[0], args[1])
	},
}

var workspaceRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a workspace from the registry",
	Long: `Remove a workspace from the registry.

This does not delete the workspace files, only removes the workspace from
the global registry. If removing the default workspace, a new default
will be automatically selected.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return workspaceRemove(cmd, args[0])
	},
}

var workspaceDefaultCmd = &cobra.Command{
	Use:   "default <name>",
	Short: "Set the default workspace",
	Long: `Set the specified workspace as the default workspace.

The default workspace is used when no local workspace is found during
workspace discovery.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return workspaceSetDefault(cmd, args[0])
	},
}

func init() {
	workspaceCmd.AddCommand(workspaceListCmd)
	workspaceCmd.AddCommand(workspaceAddCmd)
	workspaceCmd.AddCommand(workspaceRemoveCmd)
	workspaceCmd.AddCommand(workspaceDefaultCmd)
}

// Workspace management command implementations

func workspaceShowPath(cmd *cobra.Command) error {
	ctx := cmdutil.StartCommand(cmd)

	// Initialize config system
	if err := cmdutil.InitializeConfigWithError(ctx); err != nil {
		return err
	}

	ws, err := workspace.FindWorkspace()
	if err != nil {
		return ctx.HandleError(fmt.Errorf("no workspace found: %w\nRun 'jot init' to initialize a workspace or 'jot workspace list' to see registered workspaces", err))
	}

	// Output just the path for piping to other commands
	fmt.Println(ws.Root)
	return nil
}

func workspaceList(cmd *cobra.Command) error {
	ctx := cmdutil.StartCommand(cmd)

	// Initialize config system
	if err := config.Initialize(cfgFile); err != nil {
		return ctx.HandleError(fmt.Errorf("failed to initialize config: %w", err))
	}

	workspaces := config.ListWorkspaces()
	defaultWorkspace, _, _ := config.GetDefaultWorkspace()

	// Get current workspace for comparison
	currentWs, _ := workspace.FindWorkspace()
	var currentPath string
	if currentWs != nil {
		currentPath = currentWs.Root
	}

	if cmdutil.IsJSONOutput(cmd) {
		return outputWorkspaceListJSON(ctx, workspaces, defaultWorkspace, currentPath)
	}

	if len(workspaces) == 0 {
		fmt.Println("No workspaces registered.")
		fmt.Println("\nUse 'jot workspace add <name> <path>' to register a workspace")
		return nil
	}

	fmt.Println("Registered Workspaces:")
	fmt.Println()

	validCount := 0
	for name, path := range workspaces {
		status := "valid"
		isActive := false
		isDefault := name == defaultWorkspace

		// Check if path exists and is initialized
		if !workspace.IsValid(path) {
			status = "invalid - path missing or not initialized"
		} else {
			validCount++
		}

		// Check if currently active
		if currentPath != "" {
			absPath, _ := filepath.Abs(path)
			currentAbs, _ := filepath.Abs(currentPath)
			isActive = absPath == currentAbs
		}

		// Format output
		prefix := " "
		if isActive {
			prefix = "*"
		}

		defaultMarker := ""
		if isDefault {
			defaultMarker = " (default"
			if isActive {
				defaultMarker += ", active)"
			} else {
				defaultMarker += ")"
			}
		} else if isActive {
			defaultMarker = " (active)"
		}

		statusText := ""
		if status != "valid" {
			statusText = fmt.Sprintf(" (%s)", status)
		} else {
			statusText = fmt.Sprintf(" (%s)", status)
		}

		fmt.Printf("%s %-15s %-35s%s%s\n", prefix, name, path, defaultMarker, statusText)
	}

	fmt.Printf("\n* = currently active workspace\n")
	if defaultWorkspace != "" {
		fmt.Printf("Use 'jot workspace default <name>' to change default\n")
	}

	return nil
}

func workspaceAdd(cmd *cobra.Command, name, path string) error {
	ctx := cmdutil.StartCommand(cmd)

	// Initialize config system with the same config file as the main command
	if err := config.Initialize(cfgFile); err != nil {
		return ctx.HandleError(fmt.Errorf("failed to initialize config: %w", err))
	}

	// Resolve absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		err := fmt.Errorf("failed to resolve path: %w", err)
		if cmdutil.IsJSONOutput(cmd) {
			return ctx.HandleError(err)
		}
		return err
	}

	// Check if workspace already exists
	if existingPath, err := config.GetWorkspace(name); err == nil {
		err := fmt.Errorf("workspace '%s' already exists at %s\nUse 'jot workspace remove %s' first, or choose a different name", name, existingPath, name)
		if cmdutil.IsJSONOutput(cmd) {
			return ctx.HandleError(err)
		}
		return err
	}

	// Validate that path exists and is initialized
	if !workspace.IsValid(absPath) {
		err := fmt.Errorf("path %s does not exist or is not initialized\nRun 'jot init %s' to initialize it first", absPath, absPath)
		if cmdutil.IsJSONOutput(cmd) {
			return ctx.HandleError(err)
		}
		return err
	}

	// Add workspace
	err = config.AddWorkspace(name, absPath)
	if err != nil {
		if cmdutil.IsJSONOutput(cmd) {
			return ctx.HandleError(fmt.Errorf("failed to add workspace: %w", err))
		}
		return fmt.Errorf("failed to add workspace: %w", err)
	}

	// Check if this was set as default (happens automatically if it's the first workspace)
	defaultWorkspace, _, _ := config.GetDefaultWorkspace()
	setAsDefault := defaultWorkspace == name

	if cmdutil.IsJSONOutput(cmd) {
		response := map[string]interface{}{
			"operations": []map[string]interface{}{
				{
					"operation": "add_workspace",
					"result":    "success",
					"details": map[string]interface{}{
						"workspace_name":    name,
						"workspace_path":    absPath,
						"set_as_default":    setAsDefault,
						"validation_passed": true,
					},
				},
			},
			"metadata": cmdutil.CreateJSONMetadata(ctx.Cmd, true, ctx.StartTime),
		}
		return cmdutil.OutputJSON(response)
	}

	cmdutil.ShowSuccess("✓ Added workspace '%s' at %s", name, absPath)
	cmdutil.ShowSuccess("✓ Workspace validated and ready to use")
	if setAsDefault {
		cmdutil.ShowSuccess("✓ Set as default workspace (first workspace added)")
	}

	return nil
}

func workspaceRemove(cmd *cobra.Command, name string) error {
	ctx := cmdutil.StartCommand(cmd)

	// Initialize config system
	if err := config.Initialize(cfgFile); err != nil {
		return ctx.HandleError(fmt.Errorf("failed to initialize config: %w", err))
	}

	// Check if workspace exists
	workspacePath, err := config.GetWorkspace(name)
	if err != nil {
		err := fmt.Errorf("workspace '%s' not found in registry\nUse 'jot workspace list' to see available workspaces", name)
		if cmdutil.IsJSONOutput(cmd) {
			return ctx.HandleError(err)
		}
		return err
	}

	// Check if removing default workspace
	defaultWorkspace, _, _ := config.GetDefaultWorkspace()
	isDefault := name == defaultWorkspace

	// Check if removing currently active workspace
	currentWs, _ := workspace.FindWorkspace()
	isActive := false
	if currentWs != nil {
		absPath, _ := filepath.Abs(workspacePath)
		currentAbs, _ := filepath.Abs(currentWs.Root)
		isActive = absPath == currentAbs
	}

	// Remove workspace
	err = config.RemoveWorkspace(name)
	if err != nil {
		if cmdutil.IsJSONOutput(cmd) {
			return ctx.HandleError(fmt.Errorf("failed to remove workspace: %w", err))
		}
		return fmt.Errorf("failed to remove workspace: %w", err)
	}

	if cmdutil.IsJSONOutput(cmd) {
		response := map[string]interface{}{
			"operations": []map[string]interface{}{
				{
					"operation": "remove_workspace",
					"result":    "success",
					"details": map[string]interface{}{
						"workspace_name": name,
						"workspace_path": workspacePath,
						"was_default":    isDefault,
						"was_active":     isActive,
					},
				},
			},
			"metadata": cmdutil.CreateJSONMetadata(ctx.Cmd, true, ctx.StartTime),
		}
		return cmdutil.OutputJSON(response)
	}

	if isActive {
		cmdutil.ShowWarning("Warning: Removing currently active workspace '%s'", name)
	}
	if isDefault {
		cmdutil.ShowWarning("Warning: Removing default workspace '%s'", name)

		// Get new default if any workspaces remain
		newDefault, _, _ := config.GetDefaultWorkspace()
		if newDefault != "" {
			cmdutil.ShowSuccess("✓ Set '%s' as new default workspace", newDefault)
		}
	}

	cmdutil.ShowSuccess("✓ Removed workspace '%s' from registry", name)
	cmdutil.ShowInfo("Note: Workspace files at %s remain unchanged", workspacePath)

	return nil
}

func workspaceSetDefault(cmd *cobra.Command, name string) error {
	ctx := cmdutil.StartCommand(cmd)

	// Initialize config system
	if err := config.Initialize(cfgFile); err != nil {
		return ctx.HandleError(fmt.Errorf("failed to initialize config: %w", err))
	}

	// Check if workspace exists
	_, err := config.GetWorkspace(name)
	if err != nil {
		err := fmt.Errorf("workspace '%s' not found in registry\nUse 'jot workspace list' to see available workspaces", name)
		if cmdutil.IsJSONOutput(cmd) {
			return ctx.HandleError(err)
		}
		return err
	}

	// Set as default
	err = config.SetDefaultWorkspace(name)
	if err != nil {
		if cmdutil.IsJSONOutput(cmd) {
			return ctx.HandleError(fmt.Errorf("failed to set default workspace: %w", err))
		}
		return fmt.Errorf("failed to set default workspace: %w", err)
	}

	if cmdutil.IsJSONOutput(cmd) {
		response := map[string]interface{}{
			"operations": []map[string]interface{}{
				{
					"operation": "set_default_workspace",
					"result":    "success",
					"details": map[string]interface{}{
						"workspace_name": name,
					},
				},
			},
			"metadata": cmdutil.CreateJSONMetadata(ctx.Cmd, true, ctx.StartTime),
		}
		return cmdutil.OutputJSON(response)
	}

	cmdutil.ShowSuccess("✓ Set '%s' as default workspace", name)
	return nil
}

func outputWorkspaceListJSON(ctx *cmdutil.CommandContext, workspaces map[string]string, defaultWorkspace, currentPath string) error {
	var workspaceList []map[string]interface{}
	validCount := 0

	for name, path := range workspaces {
		isValid := workspace.IsValid(path)
		if isValid {
			validCount++
		}

		isDefault := name == defaultWorkspace
		isActive := false

		if currentPath != "" {
			absPath, _ := filepath.Abs(path)
			currentAbs, _ := filepath.Abs(currentPath)
			isActive = absPath == currentAbs
		}

		workspaceItem := map[string]interface{}{
			"name":       name,
			"path":       path,
			"status":     "valid",
			"is_default": isDefault,
			"is_active":  isActive,
		}

		if !isValid {
			workspaceItem["status"] = "invalid"
			workspaceItem["error"] = "path does not exist or not initialized"
		}

		workspaceList = append(workspaceList, workspaceItem)
	}

	activeWorkspace := ""
	for name, path := range workspaces {
		if currentPath != "" {
			absPath, _ := filepath.Abs(path)
			currentAbs, _ := filepath.Abs(currentPath)
			if absPath == currentAbs {
				activeWorkspace = name
				break
			}
		}
	}

	response := map[string]interface{}{
		"workspaces": workspaceList,
		"summary": map[string]interface{}{
			"total_workspaces":  len(workspaces),
			"valid_workspaces":  validCount,
			"default_workspace": defaultWorkspace,
			"active_workspace":  activeWorkspace,
		},
		"metadata": cmdutil.CreateJSONMetadata(ctx.Cmd, true, ctx.StartTime),
	}
	return cmdutil.OutputJSON(response)
}

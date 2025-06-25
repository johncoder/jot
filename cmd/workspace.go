package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

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
  jot workspace                    # Show current workspace
  jot workspace list              # List all registered workspaces
  jot workspace add notes ~/notes # Add a workspace named 'notes'
  jot workspace remove old-proj  # Remove a workspace
  jot workspace default notes    # Set default workspace`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Show current workspace
		return workspaceShowCurrent(cmd)
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

func workspaceShowCurrent(cmd *cobra.Command) error {
	startTime := time.Now()

	// Initialize config system
	if err := config.Initialize(cfgFile); err != nil {
		if isJSONOutput(cmd) {
			return outputJSONError(cmd, fmt.Errorf("failed to initialize config: %w", err), startTime)
		}
		return fmt.Errorf("failed to initialize config: %w", err)
	}

	ws, err := workspace.FindWorkspace()
	if err != nil {
		if isJSONOutput(cmd) {
			return outputJSONError(cmd, fmt.Errorf("no workspace found: %w", err), startTime)
		}
		return fmt.Errorf("no workspace found: %w\nRun 'jot init' to initialize a workspace or 'jot workspace list' to see registered workspaces", err)
	}

	// Get workspace stats
	inboxNotes, libNotes, lastActivity := getWorkspaceStats(ws)

	// Determine discovery method and workspace name
	discoveryMethod := determineDiscoveryMethod(ws)
	workspaceName := getWorkspaceNameFromPath(ws.Root)

	if isJSONOutput(cmd) {
		response := map[string]interface{}{
			"current_workspace": map[string]interface{}{
				"name":            workspaceName,
				"path":            ws.Root,
				"discovery_method": discoveryMethod,
				"status":          "active",
				"stats": map[string]interface{}{
					"inbox_notes":   inboxNotes,
					"lib_notes":     libNotes,
					"last_activity": lastActivity,
				},
			},
			"metadata": createJSONMetadata(cmd, true, startTime),
		}
		return outputJSON(response)
	}

	fmt.Printf("Current Workspace: %s\n", workspaceName)
	fmt.Printf("Path: %s\n", ws.Root)
	fmt.Printf("Status: Active (%s)\n", discoveryMethod)
	fmt.Printf("\nNotes: %d in inbox, %d in library\n", inboxNotes, libNotes)
	if !lastActivity.IsZero() {
		fmt.Printf("Last activity: %s\n", formatTimeAgo(lastActivity))
	}

	return nil
}

func workspaceList(cmd *cobra.Command) error {
	startTime := time.Now()

	// Initialize config system
	if err := config.Initialize(cfgFile); err != nil {
		if isJSONOutput(cmd) {
			return outputJSONError(cmd, fmt.Errorf("failed to initialize config: %w", err), startTime)
		}
		return fmt.Errorf("failed to initialize config: %w", err)
	}

	workspaces := config.ListWorkspaces()
	defaultWorkspace, _, _ := config.GetDefaultWorkspace()

	// Get current workspace for comparison
	currentWs, _ := workspace.FindWorkspace()
	var currentPath string
	if currentWs != nil {
		currentPath = currentWs.Root
	}

	if isJSONOutput(cmd) {
		return outputWorkspaceListJSON(cmd, workspaces, defaultWorkspace, currentPath, startTime)
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
		if !isWorkspaceValid(path) {
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
	startTime := time.Now()

	// Initialize config system with the same config file as the main command
	if err := config.Initialize(cfgFile); err != nil {
		if isJSONOutput(cmd) {
			return outputJSONError(cmd, fmt.Errorf("failed to initialize config: %w", err), startTime)
		}
		return fmt.Errorf("failed to initialize config: %w", err)
	}

	// Resolve absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		err := fmt.Errorf("failed to resolve path: %w", err)
		if isJSONOutput(cmd) {
			return outputJSONError(cmd, err, startTime)
		}
		return err
	}

	// Check if workspace already exists
	if existingPath, err := config.GetWorkspace(name); err == nil {
		err := fmt.Errorf("workspace '%s' already exists at %s\nUse 'jot workspace remove %s' first, or choose a different name", name, existingPath, name)
		if isJSONOutput(cmd) {
			return outputJSONError(cmd, err, startTime)
		}
		return err
	}

	// Validate that path exists and is initialized
	if !isWorkspaceValid(absPath) {
		err := fmt.Errorf("path %s does not exist or is not initialized\nRun 'jot init %s' to initialize it first", absPath, absPath)
		if isJSONOutput(cmd) {
			return outputJSONError(cmd, err, startTime)
		}
		return err
	}

	// Add workspace
	err = config.AddWorkspace(name, absPath)
	if err != nil {
		if isJSONOutput(cmd) {
			return outputJSONError(cmd, fmt.Errorf("failed to add workspace: %w", err), startTime)
		}
		return fmt.Errorf("failed to add workspace: %w", err)
	}

	// Check if this was set as default (happens automatically if it's the first workspace)
	defaultWorkspace, _, _ := config.GetDefaultWorkspace()
	setAsDefault := defaultWorkspace == name

	if isJSONOutput(cmd) {
		response := map[string]interface{}{
			"operations": []map[string]interface{}{
				{
					"operation": "add_workspace",
					"result":    "success",
					"details": map[string]interface{}{
						"workspace_name":     name,
						"workspace_path":     absPath,
						"set_as_default":     setAsDefault,
						"validation_passed":  true,
					},
				},
			},
			"metadata": createJSONMetadata(cmd, true, startTime),
		}
		return outputJSON(response)
	}

	fmt.Printf("✓ Added workspace '%s' at %s\n", name, absPath)
	fmt.Printf("✓ Workspace validated and ready to use\n")
	if setAsDefault {
		fmt.Printf("✓ Set as default workspace (first workspace added)\n")
	}

	return nil
}

func workspaceRemove(cmd *cobra.Command, name string) error {
	startTime := time.Now()

	// Initialize config system
	if err := config.Initialize(cfgFile); err != nil {
		if isJSONOutput(cmd) {
			return outputJSONError(cmd, fmt.Errorf("failed to initialize config: %w", err), startTime)
		}
		return fmt.Errorf("failed to initialize config: %w", err)
	}

	// Check if workspace exists
	workspacePath, err := config.GetWorkspace(name)
	if err != nil {
		err := fmt.Errorf("workspace '%s' not found in registry\nUse 'jot workspace list' to see available workspaces", name)
		if isJSONOutput(cmd) {
			return outputJSONError(cmd, err, startTime)
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
		if isJSONOutput(cmd) {
			return outputJSONError(cmd, fmt.Errorf("failed to remove workspace: %w", err), startTime)
		}
		return fmt.Errorf("failed to remove workspace: %w", err)
	}

	if isJSONOutput(cmd) {
		response := map[string]interface{}{
			"operations": []map[string]interface{}{
				{
					"operation": "remove_workspace",
					"result":    "success",
					"details": map[string]interface{}{
						"workspace_name":     name,
						"workspace_path":     workspacePath,
						"was_default":        isDefault,
						"was_active":         isActive,
					},
				},
			},
			"metadata": createJSONMetadata(cmd, true, startTime),
		}
		return outputJSON(response)
	}

	if isActive {
		fmt.Printf("Warning: Removing currently active workspace '%s'\n", name)
	}
	if isDefault {
		fmt.Printf("Warning: Removing default workspace '%s'\n", name)
		
		// Get new default if any workspaces remain
		newDefault, _, _ := config.GetDefaultWorkspace()
		if newDefault != "" {
			fmt.Printf("✓ Set '%s' as new default workspace\n", newDefault)
		}
	}

	fmt.Printf("✓ Removed workspace '%s' from registry\n", name)
	fmt.Printf("Note: Workspace files at %s remain unchanged\n", workspacePath)

	return nil
}

func workspaceSetDefault(cmd *cobra.Command, name string) error {
	startTime := time.Now()

	// Initialize config system
	if err := config.Initialize(cfgFile); err != nil {
		if isJSONOutput(cmd) {
			return outputJSONError(cmd, fmt.Errorf("failed to initialize config: %w", err), startTime)
		}
		return fmt.Errorf("failed to initialize config: %w", err)
	}

	// Check if workspace exists
	_, err := config.GetWorkspace(name)
	if err != nil {
		err := fmt.Errorf("workspace '%s' not found in registry\nUse 'jot workspace list' to see available workspaces", name)
		if isJSONOutput(cmd) {
			return outputJSONError(cmd, err, startTime)
		}
		return err
	}

	// Set as default
	err = config.SetDefaultWorkspace(name)
	if err != nil {
		if isJSONOutput(cmd) {
			return outputJSONError(cmd, fmt.Errorf("failed to set default workspace: %w", err), startTime)
		}
		return fmt.Errorf("failed to set default workspace: %w", err)
	}

	if isJSONOutput(cmd) {
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
			"metadata": createJSONMetadata(cmd, true, startTime),
		}
		return outputJSON(response)
	}

	fmt.Printf("✓ Set '%s' as default workspace\n", name)
	return nil
}

// Helper functions

func getWorkspaceStats(ws *workspace.Workspace) (inboxNotes, libNotes int, lastActivity time.Time) {
	// Count inbox notes (sections starting with ##)
	if content, err := os.ReadFile(ws.InboxPath); err == nil {
		inboxNotes = strings.Count(string(content), "\n## ")
	}

	// Count lib notes (markdown files in lib directory)
	if entries, err := os.ReadDir(ws.LibDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
				libNotes++
			}
		}
	}

	// Get last activity time (most recent modification in workspace)
	lastActivity = getLastModificationTime(ws.Root)

	return
}

func getLastModificationTime(root string) time.Time {
	var latest time.Time

	// Check inbox.md
	if info, err := os.Stat(filepath.Join(root, "inbox.md")); err == nil {
		if info.ModTime().After(latest) {
			latest = info.ModTime()
		}
	}

	// Check lib directory
	libDir := filepath.Join(root, "lib")
	if entries, err := os.ReadDir(libDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
				if info, err := entry.Info(); err == nil {
					if info.ModTime().After(latest) {
						latest = info.ModTime()
					}
				}
			}
		}
	}

	return latest
}

func determineDiscoveryMethod(ws *workspace.Workspace) string {
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

func getWorkspaceNameFromPath(path string) string {
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

func isWorkspaceValid(path string) bool {
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

func formatTimeAgo(t time.Time) string {
	duration := time.Since(t)

	if duration < time.Minute {
		return "just now"
	} else if duration < time.Hour {
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	} else {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
}

func outputWorkspaceListJSON(cmd *cobra.Command, workspaces map[string]string, defaultWorkspace, currentPath string, startTime time.Time) error {
	var workspaceList []map[string]interface{}
	validCount := 0

	for name, path := range workspaces {
		isValid := isWorkspaceValid(path)
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
			"total_workspaces": len(workspaces),
			"valid_workspaces": validCount,
			"default_workspace": defaultWorkspace,
			"active_workspace": activeWorkspace,
		},
		"metadata": createJSONMetadata(cmd, true, startTime),
	}
	return outputJSON(response)
}

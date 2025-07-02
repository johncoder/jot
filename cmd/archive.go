package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/johncoder/jot/internal/cmdutil"
	"github.com/johncoder/jot/internal/hooks"
	"github.com/johncoder/jot/internal/workspace"
	"github.com/spf13/cobra"
)

var archiveNoVerify bool

var archiveCmd = &cobra.Command{
	Use:   "archive [SOURCE]",
	Short: "Archive notes to configured archive location",
	Long: `Archive notes to the configured archive location using refile functionality.

This command is a smart alias for 'jot refile' that automatically uses the 
workspace's configured archive location as the destination.

Examples:
  jot archive                              # Set up archive structure
  jot archive "inbox.md#old-project"       # Archive specific subtree
  jot archive --config                     # Show current archive configuration
  jot archive --set-location "archive/2025.md#Archived"  # Set archive location`,

	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmdutil.StartCommand(cmd)

		ws, err := workspace.RequireWorkspace()
		if err != nil {
			return ctx.HandleError(err)
		}

		// Handle configuration flags
		showConfig, _ := cmd.Flags().GetBool("config")
		setLocation, _ := cmd.Flags().GetString("set-location")

		if showConfig {
			return showArchiveConfig(ctx, ws)
		}

		if setLocation != "" {
			return setArchiveLocation(ctx, ws, setLocation)
		}

		// If no source provided, initialize archive structure
		if len(args) == 0 {
			return initializeArchiveStructure(ctx, ws)
		}

		// Archive the specified source using refile
		return archiveWithRefile(ctx, ws, args[0])
	},
}

// showArchiveConfig displays the current archive configuration
func showArchiveConfig(ctx *cmdutil.CommandContext, ws *workspace.Workspace) error {
	archiveLocation := ws.GetArchiveLocation()
	
	if ctx.IsJSONOutput() {
		response := ArchiveConfigResponse{
			Operation:       "show_config",
			ArchiveLocation: archiveLocation,
			ResolvedPath:    archiveLocation,
			Metadata:        cmdutil.CreateJSONMetadata(ctx.Cmd, true, ctx.StartTime),
		}
		return cmdutil.OutputJSON(response)
	}

	pathUtil := cmdutil.NewPathUtil(ws)
	fmt.Printf("Archive Configuration:\n")
	fmt.Printf("  Location: %s\n", ws.Config.ArchiveLocation)
	fmt.Printf("  Resolved: %s\n", archiveLocation)
	fmt.Printf("  Full path: %s\n", pathUtil.WorkspaceJoin(archiveLocation))
	
	return nil
}

// setArchiveLocation updates the archive location configuration
func setArchiveLocation(ctx *cmdutil.CommandContext, ws *workspace.Workspace, location string) error {
	ws.Config.ArchiveLocation = location
	if err := ws.SaveWorkspaceConfig(); err != nil {
		return ctx.HandleOperationError("save configuration", err)
	}

	if ctx.IsJSONOutput() {
		response := ArchiveConfigResponse{
			Operation:       "set_location",
			ArchiveLocation: location,
			ResolvedPath:    location,
			Metadata:        cmdutil.CreateJSONMetadata(ctx.Cmd, true, ctx.StartTime),
		}
		return cmdutil.OutputJSON(response)
	}

	cmdutil.ShowSuccess("✓ Archive location updated to: %s", location)
	
	return nil
}

// initializeArchiveStructure creates the archive directory and file structure
func initializeArchiveStructure(ctx *cmdutil.CommandContext, ws *workspace.Workspace) error {
	archiveLocation := ws.GetArchiveLocation()
	pathUtil := cmdutil.NewPathUtil(ws)
	
	// Parse the archive location to extract file path and section
	// Format: "archive/archive.md#Archive"
	parts := strings.SplitN(archiveLocation, "#", 2)
	archiveFile := pathUtil.WorkspaceJoin(parts[0])
	archiveDir := filepath.Dir(archiveFile)
	
	var createdItems []ArchiveItem
	var operations []string

	// Create archive directory if it doesn't exist
	if _, err := os.Stat(archiveDir); os.IsNotExist(err) {
		if err := pathUtil.EnsureDir(archiveDir); err != nil {
			return ctx.HandleOperationError("create archive directory", err)
		}
		
		relativeDir, _ := filepath.Rel(ws.Root, archiveDir)
		createdItems = append(createdItems, ArchiveItem{
			Path:        relativeDir + "/",
			Type:        "directory",
			Description: "Archive directory",
			Created:     true,
		})
		operations = append(operations, fmt.Sprintf("Created archive directory: %s", relativeDir))
	}

	// Create archive file if it doesn't exist
	fileCreated := false
	if _, err := os.Stat(archiveFile); os.IsNotExist(err) {
		sectionName := "Archive"
		if len(parts) > 1 {
			sectionName = parts[1]
		}
		
		archiveContent := fmt.Sprintf("# %s\n\nArchived notes.\n\n", sectionName)
		
		if err := pathUtil.SafeWriteFile(archiveFile, []byte(archiveContent)); err != nil {
			return ctx.HandleOperationError("create archive file", err)
		}
		
		fileCreated = true
		relativeFile, _ := filepath.Rel(ws.Root, archiveFile)
		createdItems = append(createdItems, ArchiveItem{
			Path:        relativeFile,
			Type:        "file", 
			Description: "Archive file",
			Created:     true,
			Size:        int64(len(archiveContent)),
		})
		operations = append(operations, fmt.Sprintf("Created archive file: %s", relativeFile))
	}

	// Output results
	if ctx.IsJSONOutput() {
		var totalCreated, totalExisting int
		for _, item := range createdItems {
			if item.Created {
				totalCreated++
			} else {
				totalExisting++
			}
		}

		response := ArchiveResponse{
			Operation:    "initialize",
			ArchiveDir:   archiveDir,
			CreatedItems: createdItems,
			Operations:   operations,
			Summary: ArchiveSummary{
				TotalItems:     len(createdItems),
				ItemsCreated:   totalCreated,
				ItemsExisting:  totalExisting,
				DirectoryReady: true,
			},
			Metadata: cmdutil.CreateJSONMetadata(ctx.Cmd, true, ctx.StartTime),
		}
		return cmdutil.OutputJSON(response)
	}

	if !fileCreated && len(createdItems) == 0 {
		fmt.Println("Archive structure already exists!")
	} else {
		fmt.Println("Archive structure ready!")
	}
	
	fmt.Printf("Archive location: %s\n", archiveLocation)
	fmt.Printf("Full path: %s\n", archiveFile)
	fmt.Println()
	fmt.Println("Use 'jot archive \"source.md#section\"' to archive specific content.")
	
	return nil
}

// archiveWithRefile delegates to refile command with archive destination
func archiveWithRefile(ctx *cmdutil.CommandContext, ws *workspace.Workspace, source string) error {
	pathUtil := cmdutil.NewPathUtil(ws)
	archiveLocation := ws.GetArchiveLocation()
	
	// Parse the archive location to extract file path
	parts := strings.SplitN(archiveLocation, "#", 2)
	archiveFile := pathUtil.WorkspaceJoin(parts[0])
	
	// Ensure archive file exists first
	if _, err := os.Stat(archiveFile); os.IsNotExist(err) {
		if err := initializeArchiveStructure(ctx, ws); err != nil {
			return err
		}
	}

	// Initialize hook manager and run pre-archive hook
	hookManager := hooks.NewManager(ws)
	if !archiveNoVerify {
		hookCtx := &hooks.HookContext{
			Type:        hooks.PreArchive,
			Workspace:   ws,
			SourceFile:  source,
			DestPath:    archiveLocation,
			Timeout:     30 * time.Second,
			AllowBypass: archiveNoVerify,
		}
		
		result, err := hookManager.Execute(hookCtx)
		if err != nil {
			return ctx.HandleErrorf("pre-archive hook failed: %s", err.Error())
		}
		
		if result.Aborted {
			return ctx.HandleErrorf("pre-archive hook aborted operation")
		}
	}
	
	if !ctx.IsJSONOutput() {
		fmt.Printf("Archiving '%s' to '%s'...\n", source, archiveLocation)
	}
	
	// Call the internal refile function directly to avoid recursion
	err := executeRefile(source, archiveLocation, ctx, ws)
	
	// Run post-archive hook (informational only)
	if !archiveNoVerify && err == nil {
		hookCtx := &hooks.HookContext{
			Type:        hooks.PostArchive,
			Workspace:   ws,
			SourceFile:  source,
			DestPath:    archiveLocation,
			Timeout:     30 * time.Second,
			AllowBypass: archiveNoVerify,
		}
		
		_, hookErr := hookManager.Execute(hookCtx)
		if hookErr != nil && !ctx.IsJSONOutput() {
			cmdutil.ShowWarning("Warning: post-archive hook failed: %s", hookErr.Error())
		}
	}
	
	return err
}

// JSON response structures for archive command
type ArchiveResponse struct {
	Operation    string         `json:"operation"`
	ArchiveDir   string         `json:"archive_dir"`
	CreatedItems []ArchiveItem  `json:"created_items"`
	Operations   []string       `json:"operations"`
	Summary      ArchiveSummary `json:"summary"`
	Metadata     cmdutil.JSONMetadata   `json:"metadata"`
}

type ArchiveConfigResponse struct {
	Operation       string       `json:"operation"`
	ArchiveLocation string       `json:"archive_location"`
	ResolvedPath    string       `json:"resolved_path"`
	Metadata        cmdutil.JSONMetadata `json:"metadata"`
}

type ArchiveItem struct {
	Path        string `json:"path"`
	Type        string `json:"type"` // "file" or "directory"
	Description string `json:"description"`
	Created     bool   `json:"created"`        // Whether this item was created in this operation
	Size        int64  `json:"size,omitempty"` // For files only
}

type ArchiveSummary struct {
	TotalItems     int  `json:"total_items"`
	ItemsCreated   int  `json:"items_created"`
	ItemsExisting  int  `json:"items_existing"`
	DirectoryReady bool `json:"directory_ready"`
}

func init() {
	archiveCmd.Flags().Bool("config", false, "Show current archive configuration")
	archiveCmd.Flags().String("set-location", "", "Set archive location path")
	archiveCmd.Flags().BoolVar(&archiveNoVerify, "no-verify", false, "Skip hooks verification")
}

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/johncoder/jot/internal/tangle"
	"github.com/johncoder/jot/internal/workspace"
	"github.com/spf13/cobra"
)

var tangleCmd = &cobra.Command{
	Use:   "tangle <file>",
	Short: "Extract code blocks into standalone source files",
	Long: `Extract code blocks from Markdown files into standalone source files.

The tangle command looks for code blocks with <eval tangle file="..."/> elements 
and extracts them to the specified file paths. Directories are created as needed.

Examples:
  jot tangle notes.md              # Extract code blocks from notes.md
  jot tangle docs/tutorial.md      # Extract from tutorial file
  jot tangle --dry-run notes.md    # Show what would be tangled
  jot tangle --verbose notes.md    # Show detailed output`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		startTime := time.Now()

		// Get workspace for file path resolution
		noWorkspace, _ := cmd.Flags().GetBool("no-workspace")
		ws, err := workspace.GetWorkspaceContext(noWorkspace)
		if err != nil {
			if isJSONOutput(cmd) {
				return outputJSONError(cmd, err, startTime)
			}
			return err
		}

		filename := args[0]
		// Resolve file path relative to workspace or current directory
		resolvedFilename := resolveTangleFilePath(ws, filename, noWorkspace)
		
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		verbose, _ := cmd.Flags().GetBool("verbose")
		
		if !isJSONOutput(cmd) {
			if dryRun {
				fmt.Printf("Dry run - analyzing file: %s\n", resolvedFilename)
			} else {
				fmt.Printf("Tangling code blocks in file: %s\n", resolvedFilename)
			}
		}
		
		return tangleMarkdown(ws, resolvedFilename, dryRun, verbose, noWorkspace, cmd, startTime)
	},
}

func init() {
	tangleCmd.Flags().Bool("dry-run", false, "Show what would be tangled without actually writing files")
	tangleCmd.Flags().BoolP("verbose", "v", false, "Show detailed information about the tangle operation")
	tangleCmd.Flags().Bool("no-workspace", false, "Resolve file paths relative to current directory instead of workspace")
}

func tangleMarkdown(ws *workspace.Workspace, filePath string, dryRun, verbose bool, noWorkspace bool, cmd *cobra.Command, startTime time.Time) error {
	// Create tangle engine and find tangle blocks
	engine := tangle.NewEngine()
	if err := engine.FindTangleBlocks(ws, filePath, noWorkspace); err != nil {
		if isJSONOutput(cmd) {
			return outputJSONError(cmd, fmt.Errorf("failed to find tangle blocks: %w", err), startTime)
		}
		return fmt.Errorf("failed to find tangle blocks: %w", err)
	}

	// Group blocks by target file
	groups := engine.GroupBlocksByFile()
	
	if len(groups) == 0 {
		if isJSONOutput(cmd) {
			return outputTangleJSON(cmd, []map[string]interface{}{}, filePath, dryRun, startTime)
		}
		fmt.Println("No tangle blocks found")
		return nil
	}

	// Handle JSON output for found blocks
	if isJSONOutput(cmd) {
		// Convert groups to JSON format
		jsonGroups := make([]map[string]interface{}, 0, len(groups))
		for targetFile, blocks := range groups {
			blockInfo := make([]map[string]interface{}, len(blocks))
			for i, block := range blocks {
				blockInfo[i] = map[string]interface{}{
					"content":  block.Content,
					"language": block.Language,
				}
			}
			jsonGroups = append(jsonGroups, map[string]interface{}{
				"target_file": targetFile,
				"blocks":      blockInfo,
				"block_count": len(blocks),
			})
		}
		return outputTangleJSON(cmd, jsonGroups, filePath, dryRun, startTime)
	}

	// Create writer and configure it
	writer := tangle.NewWriter()
	writer.SetVerbose(verbose)

	if dryRun {
		writer.DryRun(groups)
		return nil
	}

	// Write the files
	if err := writer.WriteBlocks(groups); err != nil {
		return fmt.Errorf("failed to write tangle blocks: %w", err)
	}

	return nil
}

// outputTangleJSON outputs tangle results in JSON format
func outputTangleJSON(cmd *cobra.Command, groups []map[string]interface{}, sourceFile string, dryRun bool, startTime time.Time) error {
	response := map[string]interface{}{
		"source_file":   sourceFile,
		"dry_run":       dryRun,
		"total_groups":  len(groups),
		"target_files":  groups,
		"metadata":      createJSONMetadata(cmd, true, startTime),
	}

	return outputJSON(response)
}

// resolveTangleFilePath consolidates file path resolution logic for tangle operations
func resolveTangleFilePath(ws *workspace.Workspace, filename string, noWorkspace bool) string {
	if noWorkspace {
		// Non-workspace mode: resolve relative to current directory
		if filepath.IsAbs(filename) {
			return filename
		}
		cwd, _ := os.Getwd()
		return filepath.Join(cwd, filename)
	}
	
	// Workspace mode: existing logic
	if filename == "inbox.md" && ws != nil {
		return ws.InboxPath
	}
	if filepath.IsAbs(filename) {
		return filename
	}
	if ws != nil {
		return filepath.Join(ws.Root, filename)
	}
	return filename // Fallback
}

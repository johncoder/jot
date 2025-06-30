package cmd

import (
	"fmt"
	"path/filepath"

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
		// Get workspace for file path resolution
		ws, err := workspace.RequireWorkspace()
		if err != nil {
			return err
		}

		filename := args[0]
		// Resolve file path relative to workspace
		resolvedFilename := resolveTangleFilePath(ws, filename)
		
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		verbose, _ := cmd.Flags().GetBool("verbose")
		
		if dryRun {
			fmt.Printf("Dry run - analyzing file: %s\n", resolvedFilename)
		} else {
			fmt.Printf("Tangling code blocks in file: %s\n", resolvedFilename)
		}
		
		return tangleMarkdown(ws, resolvedFilename, dryRun, verbose)
	},
}

func init() {
	tangleCmd.Flags().Bool("dry-run", false, "Show what would be tangled without actually writing files")
	tangleCmd.Flags().BoolP("verbose", "v", false, "Show detailed information about the tangle operation")
}

func tangleMarkdown(ws *workspace.Workspace, filePath string, dryRun, verbose bool) error {
	// Create tangle engine and find tangle blocks
	engine := tangle.NewEngine()
	if err := engine.FindTangleBlocks(ws, filePath); err != nil {
		return fmt.Errorf("failed to find tangle blocks: %w", err)
	}

	// Group blocks by target file
	groups := engine.GroupBlocksByFile()
	
	if len(groups) == 0 {
		fmt.Println("No tangle blocks found")
		return nil
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

// resolveTangleFilePath consolidates file path resolution logic for tangle operations
func resolveTangleFilePath(ws *workspace.Workspace, filename string) string {
	if filename == "inbox.md" {
		return ws.InboxPath
	}
	if filepath.IsAbs(filename) {
		return filename
	}
	return filepath.Join(ws.Root, filename)
}

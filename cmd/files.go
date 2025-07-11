package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/johncoder/jot/internal/cmdutil"
	"github.com/johncoder/jot/internal/config"
	"github.com/johncoder/jot/internal/fzf"
	"github.com/johncoder/jot/internal/workspace"
	"github.com/spf13/cobra"
)

var filesCmd = &cobra.Command{
	Use:   "files",
	Short: "List and browse workspace files",
	Long: `List and browse files in the current jot workspace.

By default, lists all markdown files in the workspace. With the --interactive
flag and JOT_FZF=1, provides an interactive file browser with preview.

Examples:
  jot files                              # List all markdown files
  JOT_FZF=1 jot files --interactive      # Interactive file browser
  JOT_FZF=1 jot files --interactive --edit  # Interactive browser with editor
  JOT_FZF=1 jot files -i -s             # Interactive selection for composition
  cat $(jot files -i -s)                # Example: view selected file content`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmdutil.StartCommand(cmd)

		interactive, _ := cmd.Flags().GetBool("interactive")
		edit, _ := cmd.Flags().GetBool("edit")
		selectMode, _ := cmd.Flags().GetBool("select")

		// Validate flag combinations
		if selectMode && !interactive {
			err := fmt.Errorf("--select flag requires --interactive mode")
			return ctx.HandleError(err)
		}
		if selectMode && edit {
			err := fmt.Errorf("--select and --edit flags cannot be used together")
			return ctx.HandleError(err)
		}

		ws, err := getWorkspace(cmd)
		if err != nil {
			return ctx.HandleError(err)
		}

		// Get all markdown files in workspace
		files, err := findMarkdownFiles(ws.Root)
		if err != nil {
			err = fmt.Errorf("failed to find files: %w", err)
			return ctx.HandleError(err)
		}

		if len(files) == 0 {
			if cmdutil.IsJSONOutput(ctx.Cmd) {
				return outputFilesJSON(ctx, files, ws)
			}
			fmt.Println("No markdown files found in workspace")
			return nil
		}

		// Check if interactive mode is requested (not available in JSON mode)
		if fzf.ShouldUseFZF(interactive) {
			if cmdutil.IsJSONOutput(ctx.Cmd) {
				err := fmt.Errorf("interactive mode not available with JSON output")
				return ctx.HandleError(err)
			}
			return runInteractiveFilesBrowser(ws, files, edit, selectMode)
		}

		// Handle JSON output
		if cmdutil.IsJSONOutput(ctx.Cmd) {
			return outputFilesJSON(ctx, files, ws)
		}

		// Default: simple file listing with workspace-relative paths
		for _, file := range files {
			fmt.Println(ws.RelativePath(file))
		}

		return nil
	},
}

// findMarkdownFiles recursively finds all .md files in the workspace
func findMarkdownFiles(root string) ([]string, error) {
	var files []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip .jot directory
		if info.IsDir() && info.Name() == ".jot" {
			return filepath.SkipDir
		}

		// Include .md files
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".md") {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

// outputFilesJSON outputs file list in JSON format
func outputFilesJSON(ctx *cmdutil.CommandContext, files []string, ws *workspace.Workspace) error {
	// Convert file paths to JSON-friendly format with metadata
	jsonFiles := make([]map[string]interface{}, len(files))
	for i, file := range files {
		// Get file info
		info, err := os.Stat(file)
		var fileSize int64
		var modTime string
		if err == nil {
			fileSize = info.Size()
			modTime = info.ModTime().Format("2006-01-02T15:04:05Z")
		}

		jsonFiles[i] = map[string]interface{}{
			"path":          file,
			"relative_path": ws.RelativePath(file),
			"name":          filepath.Base(file),
			"size":          fileSize,
			"modified":      modTime,
		}
	}

	response := map[string]interface{}{
		"total_files": len(files),
		"files":       jsonFiles,
		"workspace": map[string]interface{}{
			"root": ws.Root,
			"name": filepath.Base(ws.Root),
		},
		"metadata": cmdutil.CreateJSONMetadata(ctx.Cmd, true, ctx.StartTime),
	}

	return cmdutil.OutputJSON(response)
}

// runInteractiveFilesBrowser runs FZF file browser with optional editor integration
func runInteractiveFilesBrowser(ws *workspace.Workspace, files []string, edit bool, selectMode bool) error {
	if len(files) == 0 {
		fmt.Println("No files to browse")
		return nil
	}

	// Create FZF search results for files
	results := make([]fzf.SearchResult, len(files))
	for i, file := range files {
		// Use workspace-relative path for display
		displayPath := ws.RelativePath(file)

		results[i] = fzf.SearchResult{
			DisplayLine: displayPath, // Show workspace-relative path in FZF
			FilePath:    file,        // Full path available for preview
			LineNumber:  1,
			Context:     fmt.Sprintf("📄 %s", filepath.Base(file)),
			Score:       100,
		}
	}

	// Handle different modes
	if selectMode {
		return runInteractiveFilesWithSelection(results)
	} else if edit {
		return runInteractiveFilesWithEditor(results)
	}

	// Otherwise, run standard interactive browser
	return fzf.RunInteractiveSearch(results, "files")
}

// runInteractiveFilesWithEditor runs FZF and opens selected file in editor
func runInteractiveFilesWithEditor(results []fzf.SearchResult) error {
	// Create a simpler FZF command for file selection
	selectedFile, err := runSimpleFZFFileSelection(results, "ENTER: open in editor, ESC: cancel")
	if err != nil {
		return err
	}

	if selectedFile == "" {
		return nil // User cancelled
	}

	// Open file in editor
	editor := config.GetEditor()
	if editor == "" {
		return fmt.Errorf("no editor configured")
	}

	// Parse editor command
	parts := strings.Fields(editor)
	if len(parts) == 0 {
		return fmt.Errorf("invalid editor configuration")
	}

	// Check if the file exists
	if _, err := os.Stat(selectedFile); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", selectedFile)
	}

	// Prepare editor command with file
	args := append(parts[1:], selectedFile)
	cmd := exec.Command(parts[0], args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Execute editor
	return cmd.Run()
}

// runInteractiveFilesWithSelection runs FZF and outputs the selected file path for composition with other tools
func runInteractiveFilesWithSelection(results []fzf.SearchResult) error {
	// Create a simpler FZF command for file selection
	selectedFile, err := runSimpleFZFFileSelection(results, "ENTER: select file path, ESC: cancel")
	if err != nil {
		return err
	}

	if selectedFile == "" {
		return nil // User cancelled
	}

	// Output the full path to stdout for composition with other CLI tools
	fmt.Println(selectedFile)
	return nil
}

// runSimpleFZFFileSelection runs a simple FZF selection and returns the chosen file path
func runSimpleFZFFileSelection(results []fzf.SearchResult, headerText string) (string, error) {
	// Create temporary file with file paths
	tempFile, err := os.CreateTemp("", "jot-files-*.txt")
	if err != nil {
		return "", err
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Write display paths to temp file, but we need to map back to full paths
	pathMap := make(map[string]string)
	for _, result := range results {
		line := result.DisplayLine + "\n"
		if _, err := tempFile.WriteString(line); err != nil {
			return "", err
		}
		pathMap[result.DisplayLine] = result.FilePath
	}

	tempFile.Close()

	// Run FZF with preview
	cmd := exec.Command("fzf",
		"--preview", "jot peek {}",
		"--preview-window", "right:50%",
		"--prompt", "Select file > ",
		"--header", headerText,
	)

	// Set up input from temp file
	file, err := os.Open(tempFile.Name())
	if err != nil {
		return "", err
	}
	defer file.Close()

	cmd.Stdin = file
	cmd.Stderr = os.Stderr

	// Run FZF and capture selection
	output, err := cmd.Output()
	if err != nil {
		// User cancelled
		return "", nil
	}

	selectedDisplay := strings.TrimSpace(string(output))
	if selectedDisplay == "" {
		return "", nil
	}

	// Map back to full path
	if fullPath, exists := pathMap[selectedDisplay]; exists {
		return fullPath, nil
	}

	return selectedDisplay, nil
}

func init() {
	filesCmd.Flags().BoolP("interactive", "i", false, "Interactive file browser (requires JOT_FZF=1)")
	filesCmd.Flags().Bool("edit", false, "Open selected file in editor (use with --interactive)")
	filesCmd.Flags().BoolP("select", "s", false, "Output selected file path for composition with other tools (use with --interactive)")

	rootCmd.AddCommand(filesCmd)
}

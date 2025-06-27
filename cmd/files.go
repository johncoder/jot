package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/johncoder/jot/internal/config"
	"github.com/johncoder/jot/internal/fzf"
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
  JOT_FZF=1 jot files --interactive --edit  # Interactive browser with editor`,
	RunE: func(cmd *cobra.Command, args []string) error {
		interactive, _ := cmd.Flags().GetBool("interactive")
		edit, _ := cmd.Flags().GetBool("edit")

		ws, err := getWorkspace(cmd)
		if err != nil {
			return err
		}

		// Get all markdown files in workspace
		files, err := findMarkdownFiles(ws.Root)
		if err != nil {
			return fmt.Errorf("failed to find files: %w", err)
		}

		if len(files) == 0 {
			fmt.Println("No markdown files found in workspace")
			return nil
		}

		// Check if interactive mode is requested
		if fzf.ShouldUseFZF(interactive) {
			return runInteractiveFilesBrowser(files, edit)
		}

		// Default: simple file listing
		for _, file := range files {
			fmt.Println(file)
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

// runInteractiveFilesBrowser runs FZF file browser with optional editor integration
func runInteractiveFilesBrowser(files []string, edit bool) error {
	if len(files) == 0 {
		fmt.Println("No files to browse")
		return nil
	}

	// Create FZF search results for files
	results := make([]fzf.SearchResult, len(files))
	for i, file := range files {
		// Use relative path for display if possible
		displayPath := file
		if cwd, err := os.Getwd(); err == nil {
			if relPath, err := filepath.Rel(cwd, file); err == nil && !strings.HasPrefix(relPath, "..") {
				displayPath = relPath
			}
		}

		results[i] = fzf.SearchResult{
			DisplayLine: displayPath, // Show relative path in FZF
			FilePath:    file,        // Full path available for preview
			LineNumber:  1,
			Context:     fmt.Sprintf("📄 %s", filepath.Base(file)),
			Score:       100,
		}
	}

	// If edit mode, set up for editor opening
	if edit {
		return runInteractiveFilesWithEditor(results)
	}

	// Otherwise, run standard interactive browser
	return fzf.RunInteractiveSearch(results, "files")
}

// runInteractiveFilesWithEditor runs FZF and opens selected file in editor
func runInteractiveFilesWithEditor(results []fzf.SearchResult) error {
	// Create a simpler FZF command for file selection
	selectedFile, err := runSimpleFZFFileSelection(results)
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

// runSimpleFZFFileSelection runs a simple FZF selection and returns the chosen file path
func runSimpleFZFFileSelection(results []fzf.SearchResult) (string, error) {
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
		"--header", "ENTER: open in editor, ESC: cancel",
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

	rootCmd.AddCommand(filesCmd)
}

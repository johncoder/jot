package fzf

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/johncoder/jot/internal/config"
)

// IsAvailable checks if FZF is available in the system PATH
func IsAvailable() bool {
	_, err := exec.LookPath("fzf")
	return err == nil
}

// ShouldUseFZF checks if both JOT_FZF=1 and interactive mode are enabled
func ShouldUseFZF(interactive bool) bool {
	return os.Getenv("JOT_FZF") == "1" && interactive && IsAvailable()
}

// SearchResult represents a result item for FZF display
type SearchResult struct {
	DisplayLine  string // What FZF shows to the user
	FilePath     string // Full path to the file
	LineNumber   int    // Line number in the file
	Context      string // The actual line content
	Score        int    // Relevance score
}

// RunInteractiveSearch runs FZF with search results and handles user interaction
func RunInteractiveSearch(results []SearchResult, query string) error {
	if len(results) == 0 {
		fmt.Printf("No matches found for '%s'\n", query)
		return nil
	}

	// Create temporary file with search results
	tempFile, err := createResultsFile(results)
	if err != nil {
		return fmt.Errorf("failed to create results file: %w", err)
	}
	defer os.Remove(tempFile)

	// Run FZF with custom configuration
	return runFZFLoop(tempFile, results, query)
}

// createResultsFile creates a temporary file with formatted search results for FZF
func createResultsFile(results []SearchResult) (string, error) {
	tempFile, err := os.CreateTemp("", "jot-search-*.txt")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	for i, result := range results {
		// Format: index|displayline|filepath|context
		// Example: 0|inbox.md:3|/absolute/path/to/inbox.md|...notes quickly. Use 'jot capture' to add new notes here.
		line := fmt.Sprintf("%d|%s|%s|%s\n", i, result.DisplayLine, result.FilePath, result.Context)
		if _, err := tempFile.WriteString(line); err != nil {
			return "", err
		}
	}

	return tempFile.Name(), nil
}

// runFZFLoop runs the main FZF interaction loop with preview and actions
func runFZFLoop(resultsFile string, results []SearchResult, query string) error {
	for {
		// Build FZF command with custom bindings
		cmd := buildFZFCommand(resultsFile, query)
		
		// Run FZF and capture selection
		output, err := cmd.Output()
		if err != nil {
			// User cancelled (Ctrl+C or Esc)
			return nil
		}

		selectedLine := strings.TrimSpace(string(output))
		if selectedLine == "" {
			return nil
		}

		// Parse the selected line to get the result index
		parts := strings.SplitN(selectedLine, "|", 4)
		if len(parts) < 4 {
			continue
		}

		// Find the selected result
		var selectedResult *SearchResult
		for i := range results {
			if fmt.Sprintf("%d", i) == parts[0] {
				selectedResult = &results[i]
				break
			}
		}

		if selectedResult == nil {
			continue
		}

		// Handle the selected result
		action, err := promptForAction(selectedResult)
		if err != nil {
			return err
		}

		switch action {
		case "view":
			if err := viewFile(selectedResult); err != nil {
				fmt.Printf("Error viewing file: %v\n", err)
			}
			// Continue the loop to return to FZF
		case "edit":
			if err := editFile(selectedResult); err != nil {
				fmt.Printf("Error editing file: %v\n", err)
			}
			// Continue the loop to return to FZF
		case "quit":
			return nil
		default:
			// Unknown action, continue loop
		}
	}
}

// buildFZFCommand creates the FZF command with appropriate options
func buildFZFCommand(resultsFile, query string) *exec.Cmd {
	args := []string{
		"--delimiter=|",
		"--with-nth=2,4",  // Show displayline and context, hide index and filepath
		"--preview=" + buildPreviewCommand(),
		"--preview-window=right:50%:wrap",
		"--bind=tab:toggle-preview",
		"--bind=enter:accept",
		"--bind=ctrl-o:accept",
		"--header=ENTER/v: view, CTRL-O: edit, TAB: toggle preview, ESC: quit",
		"--prompt=" + fmt.Sprintf("Search '%s' > ", query),
	}

	cmd := exec.Command("fzf", args...)
	
	// Set up input from results file
	file, _ := os.Open(resultsFile)
	cmd.Stdin = file
	cmd.Stderr = os.Stderr

	return cmd
}

// buildPreviewCommand creates the preview command for FZF
func buildPreviewCommand() string {
	// Extract the absolute file path from the FZF line
	// Format: index|displayline|filepath|context
	// We want field 3 (absolute file path)
	return `f=$(echo {} | cut -d'|' -f3); [ -f "$f" ] && head -50 "$f" || echo "File not found: $f"`
}

// promptForAction determines what action to take based on FZF key binding
// For now, we'll default to view action
func promptForAction(result *SearchResult) (string, error) {
	// In the future, we can detect which key was pressed in FZF
	// For now, default to view
	return "view", nil
}

// viewFile opens the selected file in the configured pager
func viewFile(result *SearchResult) error {
	// Read the file content
	content, err := os.ReadFile(result.FilePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Use the configured pager to display content
	pager := config.GetPager()
	if pager == "" {
		// No pager configured, print to stdout
		fmt.Print(string(content))
		return nil
	}

	// Parse pager command
	parts := strings.Fields(pager)
	if len(parts) == 0 {
		fmt.Print(string(content))
		return nil
	}

	// Prepare pager command
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Pipe content to pager
	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Print(string(content))
		return nil
	}

	// Start pager
	if err := cmd.Start(); err != nil {
		stdin.Close()
		fmt.Print(string(content))
		return nil
	}

	// Write content to pager
	go func() {
		defer stdin.Close()
		stdin.Write(content)
	}()

	// Wait for pager to finish
	return cmd.Wait()
}

// editFile opens the selected file in the configured editor
func editFile(result *SearchResult) error {
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
	if _, err := os.Stat(result.FilePath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", result.FilePath)
	}

	// Prepare editor command with file
	args := append(parts[1:], result.FilePath)
	cmd := exec.Command(parts[0], args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Execute editor
	return cmd.Run()
}

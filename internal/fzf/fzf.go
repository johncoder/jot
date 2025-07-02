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
	DisplayLine string // What FZF shows to the user
	FilePath    string // Full path to the file
	LineNumber  int    // Line number in the file
	Context     string // The actual line content
	Score       int    // Relevance score
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

		selectedOutput := strings.TrimSpace(string(output))
		if selectedOutput == "" {
			return nil
		}

		// When using --expect, FZF outputs the key on first line, selection on second line
		lines := strings.Split(selectedOutput, "\n")
		var keyPressed, selectedLine string

		if len(lines) == 2 {
			keyPressed = lines[0]
			selectedLine = lines[1]
		} else if len(lines) == 1 {
			keyPressed = "" // Default enter
			selectedLine = lines[0]
		} else {
			continue
		}

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

		// Determine action based on key pressed
		var action string
		if keyPressed == "alt-enter" {
			action = "edit"
		} else {
			action = "view"
		}

		// Handle the selected result
		err = handleAction(selectedResult, action)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
		// Continue the loop to return to FZF
	}
}

// handleAction performs the specified action on the selected result
func handleAction(result *SearchResult, action string) error {
	switch action {
	case "view":
		return viewFile(result)
	case "edit":
		return editFile(result)
	default:
		return nil
	}
}

// buildFZFCommand creates the FZF command with appropriate options
func buildFZFCommand(resultsFile, query string) *exec.Cmd {
	args := []string{
		"--delimiter=|",
		"--with-nth=2,4", // Show displayline and context, hide index and filepath
		"--preview=" + buildPreviewCommand(),
		"--preview-window=right:50%:wrap",
		"--bind=tab:toggle-preview",
		"--bind=enter:accept",
		"--bind=alt-enter:accept",
		"--expect=alt-enter", // Distinguish between enter and alt-enter
		"--header=ENTER: view, ALT-ENTER: edit, TAB: toggle preview, ESC: quit",
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
	// Extract the selector from the FZF line and use jot peek
	// For find results: index|enhanced_selector|filepath|context -> use field 2 (enhanced_selector)
	// For file results: index|displaypath|filepath|context -> use field 3 (filepath)
	// Try field 2 first (enhanced selector), fallback to field 3 (filepath)
	return `selector=$(echo {} | cut -d'|' -f2); filepath=$(echo {} | cut -d'|' -f3); jot peek "$selector" 2>/dev/null || jot peek "$filepath" 2>/dev/null || echo "Preview not available"`
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

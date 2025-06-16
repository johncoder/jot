package editor

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/johncoder/jot/internal/config"
)

// OpenEditor opens the configured editor with the given content
// Returns the edited content and any error
func OpenEditor(initialContent string) (string, error) {
	// Create temporary file
	tempFile, err := ioutil.TempFile("", "jot-*.md")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())

	// Write initial content
	if initialContent != "" {
		if _, err := tempFile.WriteString(initialContent); err != nil {
			tempFile.Close()
			return "", fmt.Errorf("failed to write to temp file: %w", err)
		}
	}
	tempFile.Close()

	// Get editor command
	editorCmd := config.GetEditor()

	// Parse editor command (handle editors with arguments)
	parts := strings.Fields(editorCmd)
	if len(parts) == 0 {
		return "", fmt.Errorf("no editor configured")
	}

	// Prepare command with temp file
	args := append(parts[1:], tempFile.Name())
	cmd := exec.Command(parts[0], args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Execute editor
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("editor command failed: %w", err)
	}

	// Read edited content
	content, err := ioutil.ReadFile(tempFile.Name())
	if err != nil {
		return "", fmt.Errorf("failed to read edited content: %w", err)
	}

	return string(content), nil
}

// OpenPager opens the configured pager with the given content
func OpenPager(content string) error {
	if content == "" {
		return nil
	}

	pagerCmd := config.GetPager()

	// Parse pager command
	parts := strings.Fields(pagerCmd)
	if len(parts) == 0 {
		// No pager configured, just print to stdout
		fmt.Print(content)
		return nil
	}

	// Prepare pager command
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Pipe content to pager
	stdin, err := cmd.StdinPipe()
	if err != nil {
		// Fall back to printing directly
		fmt.Print(content)
		return nil
	}

	// Start pager
	if err := cmd.Start(); err != nil {
		stdin.Close()
		// Fall back to printing directly
		fmt.Print(content)
		return nil
	}

	// Write content to pager
	_, writeErr := stdin.Write([]byte(content))
	stdin.Close()

	// Wait for pager to finish
	waitErr := cmd.Wait()

	if writeErr != nil {
		return fmt.Errorf("failed to write to pager: %w", writeErr)
	}
	if waitErr != nil {
		return fmt.Errorf("pager command failed: %w", waitErr)
	}

	return nil
}

// GetWorkspaceRoot finds the nearest .jot directory walking up the directory tree
func GetWorkspaceRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	dir := cwd
	for {
		jotDir := filepath.Join(dir, ".jot")
		if info, err := os.Stat(jotDir); err == nil && info.IsDir() {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root directory
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("not in a jot workspace (no .jot directory found)")
}

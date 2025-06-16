package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Workspace represents a jot workspace
type Workspace struct {
	Root     string
	JotDir   string
	InboxPath string
	LibDir    string
}

// FindWorkspace searches upward from the current directory for a jot workspace
func FindWorkspace() (*Workspace, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	for {
		jotDir := filepath.Join(dir, ".jot")
		if info, err := os.Stat(jotDir); err == nil && info.IsDir() {
			return &Workspace{
				Root:      dir,
				JotDir:    jotDir,
				InboxPath: filepath.Join(dir, "inbox.md"),
				LibDir:    filepath.Join(dir, "lib"),
			}, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached the root directory
			break
		}
		dir = parent
	}

	return nil, fmt.Errorf("not in a jot workspace (or any parent directory)")
}

// RequireWorkspace finds a workspace or returns an error
func RequireWorkspace() (*Workspace, error) {
	ws, err := FindWorkspace()
	if err != nil {
		return nil, fmt.Errorf("%w\nRun 'jot init' to initialize a workspace", err)
	}
	return ws, nil
}

// IsWorkspace checks if the current directory is a jot workspace
func IsWorkspace() bool {
	_, err := FindWorkspace()
	return err == nil
}

// AppendToInbox adds content to the inbox with a timestamp
func (w *Workspace) AppendToInbox(content string) error {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	entry := fmt.Sprintf("\n## %s\n\n%s\n", timestamp, content)

	file, err := os.OpenFile(w.InboxPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open inbox: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(entry)
	if err != nil {
		return fmt.Errorf("failed to write to inbox: %w", err)
	}

	return nil
}

// InboxExists checks if the inbox file exists
func (w *Workspace) InboxExists() bool {
	_, err := os.Stat(w.InboxPath)
	return err == nil
}

// LibExists checks if the lib directory exists
func (w *Workspace) LibExists() bool {
	info, err := os.Stat(w.LibDir)
	return err == nil && info.IsDir()
}

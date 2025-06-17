package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Initialize a new jot workspace",
	Long: `Initialize a new jot workspace by creating the required folder structure.

This command creates:
- inbox.md: For capturing new notes
- lib/: Directory for organized notes
- .jot/: Directory for internal data (SQLite, logs, etc.)

The workspace will be created in the current directory or the specified path.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Determine target directory
		targetDir := "."
		if len(args) > 0 {
			targetDir = args[0]
		}

		absPath, err := filepath.Abs(targetDir)
		if err != nil {
			return fmt.Errorf("failed to resolve path: %w", err)
		}

		fmt.Printf("Initializing jot workspace in: %s\n", absPath)

		// Check if workspace already exists
		jotDir := filepath.Join(absPath, ".jot")
		if _, err := os.Stat(jotDir); err == nil {
			return fmt.Errorf("jot workspace already exists in %s", absPath)
		}

		// Create inbox.md
		inboxPath := filepath.Join(absPath, "inbox.md")
		inboxContent := `# Inbox

This is your inbox for capturing new notes quickly. Use 'jot capture' to add new notes here.

---

`
		if err := os.WriteFile(inboxPath, []byte(inboxContent), 0644); err != nil {
			return fmt.Errorf("failed to create inbox.md: %w", err)
		}

		// Create lib/ directory
		libDir := filepath.Join(absPath, "lib")
		if err := os.MkdirAll(libDir, 0755); err != nil {
			return fmt.Errorf("failed to create lib directory: %w", err)
		}

		// Create .jot/ directory
		if err := os.MkdirAll(jotDir, 0755); err != nil {
			return fmt.Errorf("failed to create .jot directory: %w", err)
		}

		// Create a .gitignore for the .jot directory contents
		gitignorePath := filepath.Join(jotDir, ".gitignore")
		gitignoreContent := `# Jot internal files
*.db
*.log
tmp/
`
		if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644); err != nil {
			return fmt.Errorf("failed to create .gitignore: %w", err)
		}

		// Create a README in lib/ to explain the organization
		libReadmePath := filepath.Join(libDir, "README.md")
		libReadmeContent := `# Library

This directory contains your organized notes. You can structure them however you like:

- By topic (e.g., go/, kubernetes/, databases/)
- By project (e.g., project-alpha/, project-beta/)
- By date (e.g., 2024/, 2024/01/)
- Or any combination

Use 'jot refile' to move notes from your inbox to organized files here.
`
		if err := os.WriteFile(libReadmePath, []byte(libReadmeContent), 0644); err != nil {
			return fmt.Errorf("failed to create lib/README.md: %w", err)
		}

		fmt.Println("✓ Created inbox.md")
		fmt.Println("✓ Created lib/ directory")
		fmt.Println("✓ Created .jot/ directory")
		fmt.Println("✓ Initialized workspace structure")
		fmt.Println()
		fmt.Println("Workspace initialized! You can now:")
		fmt.Println("  jot capture    # Add your first note")
		fmt.Println("  jot status     # Check workspace status")

		return nil
	},
}

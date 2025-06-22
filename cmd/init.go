package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/johncoder/jot/internal/config"
	"github.com/spf13/cobra"
)

var (
	initWorkspaceName string
)

var initCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Initialize a new jot workspace",
	Long: `Initialize a new jot workspace by creating the required folder structure.

This command creates:
- inbox.md: For capturing new notes
- lib/: Directory for organized notes
- .jot/: Directory for internal data (SQLite, logs, etc.)

The workspace will be created in the current directory or the specified path.
The workspace will be registered in the global ~/.jotrc configuration file.

Examples:
  jot init                    # Initialize in current directory
  jot init ~/my-notes         # Initialize in specific directory
  jot init --name project-a   # Initialize with custom workspace name`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		startTime := time.Now()

		// Determine target directory
		targetDir := "."
		if len(args) > 0 {
			targetDir = args[0]
		}

		absPath, err := filepath.Abs(targetDir)
		if err != nil {
			err := fmt.Errorf("failed to resolve path: %w", err)
			if isJSONOutput(cmd) {
				return outputJSONError(cmd, err, startTime)
			}
			return err
		}

		if !isJSONOutput(cmd) {
			fmt.Printf("Initializing jot workspace in: %s\n", absPath)
		}

		// Check if workspace already exists
		jotDir := filepath.Join(absPath, ".jot")
		if _, err := os.Stat(jotDir); err == nil {
			err := fmt.Errorf("jot workspace already exists in %s", absPath)
			if isJSONOutput(cmd) {
				return outputJSONError(cmd, err, startTime)
			}
			return err
		}

		// Track created files for JSON output
		var createdFiles []InitFile

		// Create inbox.md
		inboxPath := filepath.Join(absPath, "inbox.md")
		inboxContent := `# Inbox

This is your inbox for capturing new notes quickly. Use 'jot capture' to add new notes here.

---

`
		if err := os.WriteFile(inboxPath, []byte(inboxContent), 0644); err != nil {
			err := fmt.Errorf("failed to create inbox.md: %w", err)
			if isJSONOutput(cmd) {
				return outputJSONError(cmd, err, startTime)
			}
			return err
		}
		createdFiles = append(createdFiles, InitFile{
			Path:        "inbox.md",
			Type:        "file",
			Description: "Main inbox for capturing notes",
			Size:        int64(len(inboxContent)),
		})

		// Create lib/ directory
		libDir := filepath.Join(absPath, "lib")
		if err := os.MkdirAll(libDir, 0755); err != nil {
			err := fmt.Errorf("failed to create lib directory: %w", err)
			if isJSONOutput(cmd) {
				return outputJSONError(cmd, err, startTime)
			}
			return err
		}
		createdFiles = append(createdFiles, InitFile{
			Path:        "lib/",
			Type:        "directory",
			Description: "Directory for organized notes",
		})

		// Create .jot/ directory
		if err := os.MkdirAll(jotDir, 0755); err != nil {
			err := fmt.Errorf("failed to create .jot directory: %w", err)
			if isJSONOutput(cmd) {
				return outputJSONError(cmd, err, startTime)
			}
			return err
		}
		createdFiles = append(createdFiles, InitFile{
			Path:        ".jot/",
			Type:        "directory",
			Description: "Internal data directory",
		})

		// Create a .gitignore for the .jot directory contents
		gitignorePath := filepath.Join(jotDir, ".gitignore")
		gitignoreContent := `# Jot internal files
*.db
*.log
tmp/
`
		if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644); err != nil {
			err := fmt.Errorf("failed to create .gitignore: %w", err)
			if isJSONOutput(cmd) {
				return outputJSONError(cmd, err, startTime)
			}
			return err
		}
		createdFiles = append(createdFiles, InitFile{
			Path:        ".jot/.gitignore",
			Type:        "file",
			Description: "Git ignore file for internal data",
			Size:        int64(len(gitignoreContent)),
		})

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
			err := fmt.Errorf("failed to create lib/README.md: %w", err)
			if isJSONOutput(cmd) {
				return outputJSONError(cmd, err, startTime)
			}
			return err
		}
		createdFiles = append(createdFiles, InitFile{
			Path:        "lib/README.md",
			Type:        "file",
			Description: "Documentation for library organization",
			Size:        int64(len(libReadmeContent)),
		})

		// Register workspace in global configuration
		workspaceName := initWorkspaceName
		if workspaceName == "" {
			// Auto-generate workspace name from directory name
			workspaceName = filepath.Base(absPath)
			// Clean up the name to be filesystem-safe
			workspaceName = strings.ReplaceAll(workspaceName, " ", "-")
			workspaceName = strings.ToLower(workspaceName)
		}

		// Add workspace to registry
		if err := config.RegisterWorkspace(workspaceName, absPath); err != nil {
			if strings.Contains(err.Error(), "already exists") {
				err := fmt.Errorf("workspace name %q already exists. Try: jot init --name %s-2", workspaceName, workspaceName)
				if isJSONOutput(cmd) {
					return outputJSONError(cmd, err, startTime)
				}
				return err
			}
			// For other errors, continue but warn the user
			if !isJSONOutput(cmd) {
				fmt.Printf("Warning: Failed to register workspace in global config: %v\n", err)
			}
		}

		// Output results
		if isJSONOutput(cmd) {
			// Calculate summary
			var totalFiles, totalDirectories int
			for _, file := range createdFiles {
				if file.Type == "file" {
					totalFiles++
				} else if file.Type == "directory" {
					totalDirectories++
				}
			}

			response := InitResponse{
				Operation:     "init",
				WorkspacePath: absPath,
				CreatedFiles:  createdFiles,
				Summary: InitSummary{
					TotalFiles:       totalFiles,
					TotalDirectories: totalDirectories,
				},
				Metadata: createJSONMetadata(cmd, true, startTime),
			}

			return outputJSON(response)
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

func init() {
	initCmd.Flags().StringVar(&initWorkspaceName, "name", "", "custom name for the workspace (default: directory name)")
}

// JSON response structures for init command
type InitResponse struct {
	Operation     string       `json:"operation"`
	WorkspacePath string       `json:"workspace_path"`
	CreatedFiles  []InitFile   `json:"created_files"`
	Summary       InitSummary  `json:"summary"`
	Metadata      JSONMetadata `json:"metadata"`
}

type InitFile struct {
	Path        string `json:"path"`
	Type        string `json:"type"` // "file" or "directory"
	Description string `json:"description"`
	Size        int64  `json:"size,omitempty"` // For files only
}

type InitSummary struct {
	TotalFiles       int `json:"total_files"`
	TotalDirectories int `json:"total_directories"`
}

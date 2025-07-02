package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/johncoder/jot/internal/cmdutil"
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

The workspace will be created in the current directory or the specified path.

To register this workspace for global access:
  jot workspace add <name> <path>

Examples:
  jot init                    # Initialize in current directory
  jot init ~/my-notes         # Initialize in specific directory`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmdutil.StartCommand(cmd)

		// Determine target directory
		targetDir := "."
		if len(args) > 0 {
			targetDir = args[0]
		}

		absPath, err := filepath.Abs(targetDir)
		if err != nil {
			return ctx.HandleOperationError("resolve path", err)
		}

		if !ctx.IsJSONOutput() {
			fmt.Printf("Initializing jot workspace in: %s\n", absPath)
		}

		// Check if workspace already exists
		jotDir := filepath.Join(absPath, ".jot")
		if _, err := os.Stat(jotDir); err == nil {
			return ctx.HandleError(fmt.Errorf("jot workspace already exists in %s", absPath))
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
			return ctx.HandleOperationError("create inbox.md", err)
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
			return ctx.HandleOperationError("create lib directory", err)
		}
		createdFiles = append(createdFiles, InitFile{
			Path:        "lib/",
			Type:        "directory",
			Description: "Directory for organized notes",
		})

		// Create .jot/ directory
		if err := os.MkdirAll(jotDir, 0755); err != nil {
			return ctx.HandleOperationError("create .jot directory", err)
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
			return ctx.HandleOperationError("create .gitignore", err)
		}
		createdFiles = append(createdFiles, InitFile{
			Path:        ".jot/.gitignore",
			Type:        "file",
			Description: "Git ignore file for internal data",
			Size:        int64(len(gitignoreContent)),
		})

		// Create default workspace configuration
		configPath := filepath.Join(jotDir, "config.json")
		configContent := `{
  "archive_location": "archive/archive.md#Archive"
}
`
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			return ctx.HandleOperationError("create workspace config", err)
		}
		createdFiles = append(createdFiles, InitFile{
			Path:        ".jot/config.json",
			Type:        "file",
			Description: "Workspace configuration",
			Size:        int64(len(configContent)),
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
			return ctx.HandleOperationError("create lib/README.md", err)
		}
		createdFiles = append(createdFiles, InitFile{
			Path:        "lib/README.md",
			Type:        "file",
			Description: "Documentation for library organization",
			Size:        int64(len(libReadmeContent)),
		})

		// Output results
		if ctx.IsJSONOutput() {
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
				Metadata: cmdutil.CreateJSONMetadata(ctx.Cmd, true, ctx.StartTime),
			}

			return cmdutil.OutputJSON(response)
		}

		fmt.Println("✓ Created inbox.md")
		fmt.Println("✓ Created lib/ directory")
		fmt.Println("✓ Created .jot/ directory")
		fmt.Println("✓ Initialized workspace structure")
		fmt.Println()
		fmt.Println("Workspace created successfully!")
		fmt.Println()
		fmt.Printf("To register this workspace for global access:\n")
		fmt.Printf("  jot workspace add <name> %s\n", absPath)
		fmt.Println()
		fmt.Println("Or work locally by running commands from within the workspace directory.")

		return nil
	},
}

func init() {
	// No flags needed for init command
}

// JSON response structures for init command
type InitResponse struct {
	Operation     string       `json:"operation"`
	WorkspacePath string       `json:"workspace_path"`
	CreatedFiles  []InitFile   `json:"created_files"`
	Summary       InitSummary  `json:"summary"`
	Metadata      cmdutil.JSONMetadata `json:"metadata"`
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

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/johncoder/jot/internal/workspace"
	"github.com/spf13/cobra"
)

var archiveCmd = &cobra.Command{
	Use:   "archive",
	Short: "Initialize archive structure for notes",
	Long: `Initialize archive structure for long-term storage of notes.

Currently creates the archive directory structure and monthly archive files.
Interactive archiving functionality is planned for future releases.

Examples:
  jot archive                    # Initialize archive structure`,
	RunE: func(cmd *cobra.Command, args []string) error {
		startTime := time.Now()

		ws, err := workspace.RequireWorkspace()
		if err != nil {
			if isJSONOutput(cmd) {
				return outputJSONError(cmd, err, startTime)
			}
			return err
		}

		if !isJSONOutput(cmd) {
			fmt.Println("Starting archive process...")
			fmt.Println()
		}

		var createdItems []ArchiveItem
		var operations []string

		// Create archive directory structure if it doesn't exist
		archiveDir := filepath.Join(ws.JotDir, "archive")
		dirCreated := false
		if _, err := os.Stat(archiveDir); os.IsNotExist(err) {
			if err := os.MkdirAll(archiveDir, 0755); err != nil {
				err := fmt.Errorf("failed to create archive directory: %w", err)
				if isJSONOutput(cmd) {
					return outputJSONError(cmd, err, startTime)
				}
				return err
			}
			dirCreated = true
			createdItems = append(createdItems, ArchiveItem{
				Path:        "archive/",
				Type:        "directory",
				Description: "Archive directory for organized note storage",
				Created:     true,
			})
			operations = append(operations, "Created archive directory")
		} else {
			createdItems = append(createdItems, ArchiveItem{
				Path:        "archive/",
				Type:        "directory",
				Description: "Archive directory for organized note storage",
				Created:     false,
			})
			operations = append(operations, "Archive directory already exists")
		}

		// Create current month's archive file
		now := time.Now()
		relativeArchiveFile := filepath.Join("archive", now.Format("2006-01.md"))
		archiveFile := filepath.Join(archiveDir, now.Format("2006-01.md"))
		fileCreated := false

		if _, err := os.Stat(archiveFile); os.IsNotExist(err) {
			archiveContent := fmt.Sprintf("# Archive %s\n\nArchived notes from %s\n\n---\n\n",
				now.Format("January 2006"), now.Format("January 2006"))
			if err := os.WriteFile(archiveFile, []byte(archiveContent), 0644); err != nil {
				err := fmt.Errorf("failed to create archive file: %w", err)
				if isJSONOutput(cmd) {
					return outputJSONError(cmd, err, startTime)
				}
				return err
			}
			fileCreated = true
			createdItems = append(createdItems, ArchiveItem{
				Path:        relativeArchiveFile,
				Type:        "file",
				Description: fmt.Sprintf("Monthly archive file for %s", now.Format("January 2006")),
				Created:     true,
				Size:        int64(len(archiveContent)),
			})
			operations = append(operations, fmt.Sprintf("Created archive file: %s", relativeArchiveFile))
			if !isJSONOutput(cmd) {
				fmt.Printf("✓ Created archive file: %s\n", archiveFile)
			}
		} else {
			// Get file size for existing file
			if info, err := os.Stat(archiveFile); err == nil {
				createdItems = append(createdItems, ArchiveItem{
					Path:        relativeArchiveFile,
					Type:        "file",
					Description: fmt.Sprintf("Monthly archive file for %s", now.Format("January 2006")),
					Created:     false,
					Size:        info.Size(),
				})
			}
			operations = append(operations, fmt.Sprintf("Archive file already exists: %s", relativeArchiveFile))
		}

		// Output results
		if isJSONOutput(cmd) {
			var totalCreated, totalExisting int
			for _, item := range createdItems {
				if item.Created {
					totalCreated++
				} else {
					totalExisting++
				}
			}

			response := ArchiveResponse{
				Operation:    "archive",
				ArchiveDir:   archiveDir,
				CreatedItems: createdItems,
				Operations:   operations,
				Summary: ArchiveSummary{
					TotalItems:     len(createdItems),
					ItemsCreated:   totalCreated,
					ItemsExisting:  totalExisting,
					DirectoryReady: true,
				},
				Metadata: createJSONMetadata(cmd, true, startTime),
			}

			return outputJSON(response)
		}

		if !dirCreated && !fileCreated {
			fmt.Println("Archive structure already exists!")
		} else {
			fmt.Println("Archive structure ready!")
		}
		fmt.Println()
		fmt.Printf("Archive location: %s\n", archiveDir)
		fmt.Printf("Current archive: %s\n", archiveFile)
		fmt.Println()
		fmt.Println("Note: Interactive archiving functionality coming soon!")
		fmt.Println("For now, you can manually move old notes to archive files.")

		return nil
	},
}

// JSON response structures for archive command
type ArchiveResponse struct {
	Operation    string         `json:"operation"`
	ArchiveDir   string         `json:"archive_dir"`
	CreatedItems []ArchiveItem  `json:"created_items"`
	Operations   []string       `json:"operations"`
	Summary      ArchiveSummary `json:"summary"`
	Metadata     JSONMetadata   `json:"metadata"`
}

type ArchiveItem struct {
	Path        string `json:"path"`
	Type        string `json:"type"` // "file" or "directory"
	Description string `json:"description"`
	Created     bool   `json:"created"`        // Whether this item was created in this operation
	Size        int64  `json:"size,omitempty"` // For files only
}

type ArchiveSummary struct {
	TotalItems     int  `json:"total_items"`
	ItemsCreated   int  `json:"items_created"`
	ItemsExisting  int  `json:"items_existing"`
	DirectoryReady bool `json:"directory_ready"`
}

func init() {
	// Archive functionality is basic for now - no flags needed
}

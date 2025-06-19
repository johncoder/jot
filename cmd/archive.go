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
		ws, err := workspace.RequireWorkspace()
		if err != nil {
			return err
		}

		fmt.Println("Starting archive process...")
		fmt.Println()

		// Create archive directory structure if it doesn't exist
		archiveDir := filepath.Join(ws.JotDir, "archive")
		if err := os.MkdirAll(archiveDir, 0755); err != nil {
			return fmt.Errorf("failed to create archive directory: %w", err)
		}

		// Create current month's archive file
		now := time.Now()
		archiveFile := filepath.Join(archiveDir, now.Format("2006-01.md"))

		if _, err := os.Stat(archiveFile); os.IsNotExist(err) {
			archiveContent := fmt.Sprintf("# Archive %s\n\nArchived notes from %s\n\n---\n\n",
				now.Format("January 2006"), now.Format("January 2006"))
			if err := os.WriteFile(archiveFile, []byte(archiveContent), 0644); err != nil {
				return fmt.Errorf("failed to create archive file: %w", err)
			}
			fmt.Printf("✓ Created archive file: %s\n", archiveFile)
		}

		fmt.Println("Archive structure ready!")
		fmt.Println()
		fmt.Printf("Archive location: %s\n", archiveDir)
		fmt.Printf("Current archive: %s\n", archiveFile)
		fmt.Println()
		fmt.Println("Note: Interactive archiving functionality coming soon!")
		fmt.Println("For now, you can manually move old notes to archive files.")

		return nil
	},
}

func init() {
	// Archive functionality is basic for now - no flags needed
}

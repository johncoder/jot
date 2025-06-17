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
	Short: "Archive notes for long-term storage",
	Long: `Archive notes from inbox.md or lib/ files for long-term storage.

Archived notes are moved to the archive location while maintaining
searchability and metadata for future retrieval.

Examples:
  jot archive                    # Interactive archive selection
  jot archive --older-than 30d   # Archive notes older than 30 days
  jot archive --from lib/old.md  # Archive from specific file`,
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
	archiveCmd.Flags().String("older-than", "", "Archive notes older than specified duration (e.g., 30d, 6m)")
	archiveCmd.Flags().String("from", "", "Archive from specific file")
}

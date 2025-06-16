package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/johncoder/jot/internal/workspace"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show workspace status",
	Long: `Show the health and summary of the current jot workspace.

Displays information about:
- Workspace location and structure
- Note counts by location (inbox, lib, archive)
- Recent activity summary
- Workspace health indicators

Examples:
  jot status                     # Show workspace status
  jot status --verbose           # Show detailed information`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := workspace.RequireWorkspace()
		if err != nil {
			return err
		}

		fmt.Println("Jot Workspace Status")
		fmt.Println("===================")
		fmt.Println()
		
		fmt.Printf("Location: %s\n", ws.Root)
		
		// Check workspace structure
		issues := []string{}
		if !ws.InboxExists() {
			issues = append(issues, "inbox.md is missing")
		}
		if !ws.LibExists() {
			issues = append(issues, "lib/ directory is missing")
		}
		
		// Count notes in inbox
		inboxNotes := countNotesInFile(ws.InboxPath)
		
		// Count notes in lib
		libNotes, libFiles := countNotesInDir(ws.LibDir)
		
		fmt.Println()
		fmt.Println("Notes Summary:")
		fmt.Printf("  Inbox:     %d notes\n", inboxNotes)
		fmt.Printf("  Library:   %d notes (%d files)\n", libNotes, libFiles)
		fmt.Printf("  Total:     %d notes\n", inboxNotes+libNotes)
		fmt.Println()
		
		// Check recent activity
		if ws.InboxExists() {
			if info, err := os.Stat(ws.InboxPath); err == nil {
				lastModified := info.ModTime()
				fmt.Printf("Last inbox activity: %s\n", formatRelativeTime(lastModified))
			}
		}
		
		fmt.Println()
		if len(issues) == 0 {
			fmt.Println("Workspace Health: ✓ Good")
		} else {
			fmt.Println("Workspace Health: ⚠ Issues found")
			for _, issue := range issues {
				fmt.Printf("  - %s\n", issue)
			}
			fmt.Println("\nRun 'jot doctor' for repair suggestions")
		}
		
		return nil
	},
}

// countNotesInFile counts ## headers in a markdown file
func countNotesInFile(path string) int {
	file, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer file.Close()

	count := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "## ") {
			count++
		}
	}
	return count
}

// countNotesInDir counts notes in all markdown files in a directory
func countNotesInDir(dir string) (int, int) {
	totalNotes := 0
	fileCount := 0

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't read
		}

		if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".md") {
			// Skip README.md files in counting
			if strings.ToLower(info.Name()) == "readme.md" {
				return nil
			}
			
			notes := countNotesInFile(path)
			fileCount++
			
			// If file has ## headers, count those as individual notes
			// Otherwise, count the file itself as one note
			if notes > 0 {
				totalNotes += notes
			} else {
				totalNotes += 1
			}
		}
		return nil
	})

	if err != nil {
		return 0, 0
	}

	return totalNotes, fileCount
}

// formatRelativeTime formats a time relative to now
func formatRelativeTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "just now"
	}
	if diff < time.Hour {
		minutes := int(diff.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	}
	if diff < 24*time.Hour {
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}
	days := int(diff.Hours() / 24)
	if days == 1 {
		return "1 day ago"
	}
	if days < 7 {
		return fmt.Sprintf("%d days ago", days)
	}
	if days < 30 {
		weeks := days / 7
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	}
	months := days / 30
	if months == 1 {
		return "1 month ago"
	}
	return fmt.Sprintf("%d months ago", months)
}

func init() {
	statusCmd.Flags().BoolP("verbose", "v", false, "Show detailed information")
}

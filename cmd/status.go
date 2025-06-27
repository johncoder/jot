package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

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
		startTime := time.Now()
		ws, err := getWorkspace(cmd)
		if err != nil {
			if isJSONOutput(cmd) {
				return outputJSONError(cmd, err, startTime)
			}
			return err
		}

		// Collect all status data
		issues := []string{}
		if !ws.InboxExists() {
			issues = append(issues, "inbox.md is missing")
		}
		if !ws.LibExists() {
			issues = append(issues, "lib/ directory is missing")
		}

		inboxNotes := countNotesInFile(ws.InboxPath)
		libNotes, libFiles := countNotesInDir(ws.LibDir)
		totalNotes := inboxNotes + libNotes

		healthStatus := "healthy"
		if len(issues) > 0 {
			healthStatus = "issues_found"
		}

		var lastActivity *time.Time
		var lastActivityText string
		if ws.InboxExists() {
			if info, err := os.Stat(ws.InboxPath); err == nil {
				modTime := info.ModTime()
				lastActivity = &modTime
				lastActivityText = formatRelativeTime(modTime)
			}
		}

		// Output JSON if requested
		if isJSONOutput(cmd) {
			response := StatusResponse{
				Workspace: StatusWorkspace{
					Root:      ws.Root,
					InboxPath: ws.InboxPath,
					LibDir:    ws.LibDir,
					JotDir:    ws.JotDir,
				},
				Files: StatusFiles{
					InboxNotes: inboxNotes,
					LibFiles:   libFiles,
					LibNotes:   libNotes,
					TotalNotes: totalNotes,
				},
				Health: StatusHealth{
					Status: healthStatus,
					Issues: issues,
				},
				Metadata: createJSONMetadata(cmd, true, startTime),
			}

			if lastActivity != nil {
				response.Activity = StatusActivity{
					LastInboxActivity:     lastActivity,
					LastInboxActivityText: lastActivityText,
				}
			}

			return outputJSON(response)
		}

		// Get workspace discovery information
		discoveryMethod := determineDiscoveryMethod(ws)
		workspaceName := getWorkspaceNameFromPath(ws.Root)

		// Human-readable output
		fmt.Println("Jot Workspace Status")
		fmt.Println("===================")
		fmt.Println()

		fmt.Printf("Workspace: %s\n", workspaceName)
		fmt.Printf("Location:  %s\n", ws.Root)
		fmt.Printf("Status:    Active (%s)\n", discoveryMethod)

		fmt.Println()
		fmt.Println("Notes Summary:")
		fmt.Printf("  Inbox:     %d notes\n", inboxNotes)
		fmt.Printf("  Library:   %d notes (%d files)\n", libNotes, libFiles)
		fmt.Printf("  Total:     %d notes\n", totalNotes)
		fmt.Println()

		if lastActivityText != "" {
			fmt.Printf("Last inbox activity: %s\n", lastActivityText)
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

// StatusResponse represents the JSON response for status command
type StatusResponse struct {
	Workspace StatusWorkspace `json:"workspace"`
	Files     StatusFiles     `json:"files"`
	Health    StatusHealth    `json:"health"`
	Activity  StatusActivity  `json:"activity,omitempty"`
	Metadata  JSONMetadata    `json:"metadata"`
}

type StatusWorkspace struct {
	Root      string `json:"root"`
	InboxPath string `json:"inbox_path"`
	LibDir    string `json:"lib_dir"`
	JotDir    string `json:"jot_dir"`
}

type StatusFiles struct {
	InboxNotes int `json:"inbox_notes"`
	LibFiles   int `json:"lib_files"`
	LibNotes   int `json:"lib_notes"`
	TotalNotes int `json:"total_notes"`
}

type StatusHealth struct {
	Status string   `json:"status"`
	Issues []string `json:"issues"`
}

type StatusActivity struct {
	LastInboxActivity     *time.Time `json:"last_inbox_activity,omitempty"`
	LastInboxActivityText string     `json:"last_inbox_activity_text,omitempty"`
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

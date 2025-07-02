package workspace

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Stats contains workspace statistics
type Stats struct {
	InboxNotes   int
	LibNotes     int
	LastActivity time.Time
}

// GetStats returns statistics for the given workspace
func GetStats(ws *Workspace) Stats {
	inboxNotes, libNotes, lastActivity := calculateStats(ws)
	return Stats{
		InboxNotes:   inboxNotes,
		LibNotes:     libNotes,
		LastActivity: lastActivity,
	}
}

// calculateStats computes workspace statistics
func calculateStats(ws *Workspace) (inboxNotes, libNotes int, lastActivity time.Time) {
	// Count inbox notes (sections starting with ##)
	if content, err := os.ReadFile(ws.InboxPath); err == nil {
		inboxNotes = strings.Count(string(content), "\n## ")
	}

	// Count lib notes (markdown files in lib directory)
	if entries, err := os.ReadDir(ws.LibDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
				libNotes++
			}
		}
	}

	// Get last activity time (most recent modification in workspace)
	lastActivity = GetLastModificationTime(ws.Root)

	return
}

// GetLastModificationTime finds the most recent modification time in the workspace
func GetLastModificationTime(root string) time.Time {
	var latest time.Time

	// Check inbox.md
	if info, err := os.Stat(filepath.Join(root, "inbox.md")); err == nil {
		if info.ModTime().After(latest) {
			latest = info.ModTime()
		}
	}

	// Check lib directory
	libDir := filepath.Join(root, "lib")
	if entries, err := os.ReadDir(libDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
				if info, err := entry.Info(); err == nil {
					if info.ModTime().After(latest) {
						latest = info.ModTime()
					}
				}
			}
		}
	}

	return latest
}

package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/johncoder/jot/internal/workspace"
	"github.com/spf13/cobra"
)

var (
	findInArchive bool
	findLimit     int
)

var findCmd = &cobra.Command{
	Use:   "find <query>",
	Short: "Search through notes",
	Long: `Search through notes in inbox.md, lib/, and optionally archive.

Supports keyword search and full-text search with context display.
Results are ranked by relevance and recency.

Examples:
  jot find "meeting notes"       # Search for phrase
  jot find golang --limit 10     # Limit results
  jot find todo --archive        # Include archived notes`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := workspace.RequireWorkspace()
		if err != nil {
			return err
		}

		query := strings.Join(args, " ")
		fmt.Printf("Searching for: %s\n", query)
		
		if findInArchive {
			fmt.Println("Including archived notes in search...")
		}
		
		// Collect all markdown files to search
		var filesToSearch []string
		
		// Add inbox if it exists
		if ws.InboxExists() {
			filesToSearch = append(filesToSearch, ws.InboxPath)
		}
		
		// Add lib files
		err = filepath.Walk(ws.LibDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Skip files we can't read
			}
			
			if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".md") {
				filesToSearch = append(filesToSearch, path)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to scan lib directory: %w", err)
		}
		
		// Search files and collect results
		var results []SearchResult
		for _, filePath := range filesToSearch {
			matches := searchInFile(filePath, query, ws.Root)
			results = append(results, matches...)
		}
		
		// Sort results by relevance (simple keyword count for now)
		sort.Slice(results, func(i, j int) bool {
			return results[i].Score > results[j].Score
		})
		
		// Apply limit
		if len(results) > findLimit {
			results = results[:findLimit]
		}
		
		// Display results
		if len(results) == 0 {
			fmt.Printf("No matches found for '%s'\n", query)
			return nil
		}
		
		fmt.Printf("Found %d matches for '%s':\n\n", len(results), query)
		for _, result := range results {
			fmt.Printf("%s:%d | %s\n", result.RelativePath, result.LineNumber, result.Context)
		}
		
		if len(results) >= findLimit {
			fmt.Printf("\nShowing first %d results (use --limit to adjust)\n", findLimit)
		}
		
		return nil
	},
}

// SearchResult represents a search match
type SearchResult struct {
	FilePath     string
	RelativePath string
	LineNumber   int
	Context      string
	Score        int
}

// searchInFile searches for query in a file and returns matches
func searchInFile(filePath, query, workspaceRoot string) []SearchResult {
	file, err := os.Open(filePath)
	if err != nil {
		return nil
	}
	defer file.Close()

	var results []SearchResult
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	queryLower := strings.ToLower(query)
	
	// Get relative path for display
	relativePath, _ := filepath.Rel(workspaceRoot, filePath)
	
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		lineLower := strings.ToLower(line)
		
		if strings.Contains(lineLower, queryLower) {
			// Calculate relevance score (simple keyword count)
			score := strings.Count(lineLower, queryLower)
			
			// Trim long lines for display
			context := strings.TrimSpace(line)
			if len(context) > 80 {
				// Try to center the match in the context
				if pos := strings.Index(lineLower, queryLower); pos >= 0 {
					start := pos - 20
					if start < 0 {
						start = 0
					}
					end := start + 80
					if end > len(context) {
						end = len(context)
					}
					context = context[start:end]
					if start > 0 {
						context = "..." + context
					}
					if end < len(line) {
						context = context + "..."
					}
				} else {
					context = context[:80] + "..."
				}
			}
			
			results = append(results, SearchResult{
				FilePath:     filePath,
				RelativePath: relativePath,
				LineNumber:   lineNumber,
				Context:      context,
				Score:        score,
			})
		}
	}
	
	return results
}

func init() {
	findCmd.Flags().BoolVar(&findInArchive, "archive", false, "Include archived notes in search")
	findCmd.Flags().IntVar(&findLimit, "limit", 20, "Limit number of results")
}

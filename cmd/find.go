package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/johncoder/jot/internal/cmdutil"
	"github.com/johncoder/jot/internal/fzf"
	"github.com/johncoder/jot/internal/markdown"
	"github.com/johncoder/jot/internal/workspace"
	"github.com/spf13/cobra"
)

var (
	findInArchive   bool
	findLimit       int
	findInteractive bool
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
  jot find todo --archive        # Include archived notes
  JOT_FZF=1 jot find todo --interactive  # Interactive search with FZF`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmdutil.StartCommand(cmd)

		ws, err := getWorkspace(cmd)
		if err != nil {
			return ctx.HandleError(err)
		}

		query := strings.Join(args, " ")

		// Check for interactive mode with FZF (not available in JSON mode)
		if fzf.ShouldUseFZF(findInteractive) {
			if cmdutil.IsJSONOutput(ctx.Cmd) {
				err := fmt.Errorf("interactive mode not available with JSON output")
				return ctx.HandleError(err)
			}
			return runInteractiveFind(ws, query)
		}

		if !cmdutil.IsJSONOutput(ctx.Cmd) {
			fmt.Printf("Searching for: %s\n", query)
			if findInArchive {
				fmt.Println("Including archived notes in search...")
			}
		}

		// Collect search results
		results := collectSearchResults(ws, query)

		// Handle JSON output
		if cmdutil.IsJSONOutput(ctx.Cmd) {
			return outputFindJSON(ctx, results, query)
		}

		// Display results in text format
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

// runInteractiveFind handles the interactive FZF-based search workflow
func runInteractiveFind(ws *workspace.Workspace, query string) error {
	if findInArchive {
		fmt.Println("Including archived notes in search...")
	}

	// Collect search results using the same logic as normal find
	results := collectSearchResults(ws, query)

	if len(results) == 0 {
		fmt.Printf("No matches found for '%s'\n", query)
		return nil
	}

	// Enhance results with heading information for better selectors
	enhancedResults := enhanceResultsWithHeadings(results)

	// Convert to FZF format
	fzfResults := make([]fzf.SearchResult, len(enhancedResults))
	for i, result := range enhancedResults {
		fzfResults[i] = fzf.SearchResult{
			DisplayLine: result.RelativePath, // Now contains enhanced selector like "file:line#heading"
			FilePath:    result.FilePath,
			LineNumber:  result.LineNumber,
			Context:     result.Context,
			Score:       result.Score,
		}
	}

	// Run interactive FZF search
	return fzf.RunInteractiveSearch(fzfResults, query)
}

// enhanceResultsWithHeadings adds heading information to search results
// by parsing markdown files efficiently and creating enhanced selectors
func enhanceResultsWithHeadings(results []SearchResult) []SearchResult {
	// Group results by file for efficient processing
	fileGroups := make(map[string][]int) // filePath -> []resultIndex
	for i, result := range results {
		fileGroups[result.FilePath] = append(fileGroups[result.FilePath], i)
	}

	// Process each file once
	for filePath, resultIndices := range fileGroups {
		// Read file content
		content, err := os.ReadFile(filePath)
		if err != nil {
			continue // Skip files we can't read
		}

		// Extract line numbers for this file
		var targetLines []int
		for _, idx := range resultIndices {
			targetLines = append(targetLines, results[idx].LineNumber)
		}

		// Find headings for all target lines in this file
		headingMap, err := markdown.FindNearestHeadingsForLines(content, targetLines)
		if err != nil {
			continue // Skip files we can't parse
		}

		// Update results with enhanced selectors
		for _, idx := range resultIndices {
			result := &results[idx]
			lineNum := result.LineNumber

			if headingPath, found := headingMap[lineNum]; found && headingPath != "" {
				// Create enhanced selector: file:line#heading/path
				result.RelativePath = fmt.Sprintf("%s:%d#%s", result.RelativePath, lineNum, headingPath)
			}
			// If no heading found, keep the original format (file:line)
		}
	}

	return results
}

// collectSearchResults performs the actual search and returns results
func collectSearchResults(ws *workspace.Workspace, query string) []SearchResult {
	// Collect all markdown files to search
	var filesToSearch []string

	// Add inbox if it exists
	if ws.InboxExists() {
		filesToSearch = append(filesToSearch, ws.InboxPath)
	}

	// Add lib files
	err := filepath.Walk(ws.LibDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't read
		}

		if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".md") {
			filesToSearch = append(filesToSearch, path)
		}
		return nil
	})
	if err != nil {
		return nil
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

	return results
}

// outputFindJSON outputs search results in JSON format
func outputFindJSON(ctx *cmdutil.CommandContext, results []SearchResult, query string) error {
	// Convert search results to JSON-friendly format
	jsonResults := make([]map[string]interface{}, len(results))
	for i, result := range results {
		jsonResults[i] = map[string]interface{}{
			"file_path":     result.FilePath,
			"relative_path": result.RelativePath,
			"line_number":   result.LineNumber,
			"context":       result.Context,
			"score":         result.Score,
		}
	}

	response := map[string]interface{}{
		"query":       query,
		"total_found": len(results),
		"results":     jsonResults,
		"search_info": map[string]interface{}{
			"include_archive": findInArchive,
			"limit":           findLimit,
			"limited":         len(results) >= findLimit,
		},
		"metadata": cmdutil.CreateJSONMetadata(ctx.Cmd, true, ctx.StartTime),
	}

	return cmdutil.OutputJSON(response)
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
	findCmd.Flags().BoolVar(&findInteractive, "interactive", false, "Use FZF for interactive search (requires JOT_FZF=1)")
}

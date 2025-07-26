// Package cmd: refile_selectors.go
// Selector listing and parsing utilities for refile.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/johncoder/jot/internal/cmdutil"
	"github.com/johncoder/jot/internal/markdown"
	"github.com/johncoder/jot/internal/workspace"
)

// ShowSelectorsForFile displays available selectors for a specific file
func ShowSelectorsForFile(ws *workspace.Workspace, filename string) error {
	pathUtil := cmdutil.NewPathUtil(ws)
	// Determine the full file path
	var filePath string
	if filename == "inbox.md" {
		filePath = ws.InboxPath
	} else if filepath.IsAbs(filename) {
		filePath = filename
	} else {
		// Handle files in workspace root, add .md extension if needed
		if !strings.HasSuffix(filename, ".md") {
			filename += ".md"
		}
		// Use workspace root for relative paths, not lib/ directory
		filePath = pathUtil.WorkspaceJoin(filename)
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filename)
	}

	// Read and parse the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return cmdutil.NewFileError("read", filename, err)
	}

	if len(content) == 0 {
		fmt.Printf("File %s is empty (no selectors available)\n", filename)
		return nil
	}

	// Parse document and get headings
	doc := markdown.ParseMarkdownAST(content)
	headings := markdown.FindAllHeadings(doc)

	if len(headings) == 0 {
		fmt.Printf("No headings found in %s\n", filename)
		return nil
	}

	// Display selectors
	fmt.Printf("Available selectors in %s:\n", filename)

	for _, heading := range headings {
		if strings.TrimSpace(heading.Text) == "" {
			continue // Skip empty headings
		}

		// Build selector based on heading level and path
		var selector string
		if heading.Level == 1 {
			selector = fmt.Sprintf("%s#%s", filename, strings.ToLower(heading.Text))
		} else {
			// For level 2+ headings, use skip syntax
			skipPrefix := strings.Repeat("/", heading.Level-1)
			selector = fmt.Sprintf("%s#%s%s", filename, skipPrefix, strings.ToLower(heading.Text))
		}

		fmt.Printf("  %s\n", selector)
	}

	fmt.Printf("\nUsage: jot refile \"<selector>\" --to \"<destination>\"\n")
	return nil
}

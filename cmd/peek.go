package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gomarkdown/markdown/ast"

	"github.com/johncoder/jot/internal/cmdutil"
	"github.com/johncoder/jot/internal/markdown"
	"github.com/johncoder/jot/internal/workspace"
	"github.com/spf13/cobra"
)

var peekCmd = &cobra.Command{
	Use:   "peek SELECTOR",
	Short: "View a specific markdown subtree or entire file without opening it",
	Long: `View a specific markdown subtree (heading with all nested content) or entire file without opening it.

The peek command supports two modes:
1. Whole file: "filename.md" - displays entire file content
2. Subtree: "file.md#path/to/heading" - displays specific subtree

The subtree selector uses path-based syntax:
- Each segment uses case-insensitive contains matching
- Must match exactly one subtree
- Leading slashes handle unusual document structures

Examples:
  jot peek "inbox.md"                            # View entire inbox file
  jot peek "work.md" --info                      # Show file info and content
  jot peek "inbox.md#meeting"                    # View meeting notes subtree
  jot peek "work.md#projects/frontend"          # View frontend project section
  jot peek "notes.md#research/database"         # View database research
  jot peek "inbox.md#/foo/bar"                  # Skip level 1, find foo/bar
  jot peek "inbox.md" --toc                     # Show table of contents for entire file
  jot peek "work.md#projects" --toc             # Show TOC for projects subtree
  jot peek "work.md" --toc --short              # Show TOC with shortest selectors

This is useful for quickly reviewing files or specific sections without opening them in an editor.`,

	Args: cobra.RangeArgs(0, 1), // Allow 0 or 1 arguments for --toc mode
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmdutil.StartCommand(cmd)
		noWorkspace, _ := cmd.Flags().GetBool("no-workspace")
		ws, err := workspace.GetWorkspaceContext(noWorkspace)
		if err != nil {
			return ctx.HandleError(err)
		}

		// Get flags
		raw, _ := cmd.Flags().GetBool("raw")
		info, _ := cmd.Flags().GetBool("info")
		toc, _ := cmd.Flags().GetBool("toc")
		short, _ := cmd.Flags().GetBool("short")

		// Handle TOC mode
		if toc {
			if len(args) == 0 {
				err := fmt.Errorf("table of contents requires a file or selector (e.g., 'inbox.md' or 'work.md#projects')")
				return ctx.HandleError(err)
			}

			if cmdutil.IsJSONOutput(ctx.Cmd) {
				return showTableOfContentsJSON(ctx, ws, args[0], short)
			}
			return showTableOfContents(ws, args[0], short, noWorkspace)
		}

		// Regular peek mode requires exactly one argument
		if len(args) != 1 {
			err := fmt.Errorf("peek requires a selector argument (e.g., 'inbox.md#meeting' or 'filename.md' for whole file)")
			return ctx.HandleError(err)
		}

		selector := args[0]

		// Handle enhanced selectors with line numbers (e.g., "file:42" or "file:42#heading")
		if enhancedSelector, err := parseEnhancedSelector(ws, selector); err == nil && enhancedSelector != selector {
			// Successfully converted line number to heading, use the enhanced selector
			selector = enhancedSelector
		}

		// Check if this is a whole file request (no # selector) or a subtree request
		if !strings.Contains(selector, "#") {
			// Handle whole file display
			if cmdutil.IsJSONOutput(ctx.Cmd) {
				return showWholeFileJSON(ctx, ws, selector, noWorkspace)
			}
			return showWholeFile(ws, selector, raw, info, noWorkspace)
		}

		// Parse the source path selector for subtree extraction
		sourcePath, err := markdown.ParsePath(selector)
		if err != nil {
			err := fmt.Errorf("invalid selector: %w", err)
			return ctx.HandleError(err)
		}

		// Extract the subtree
		subtree, err := ExtractSubtreeWithOptions(ws, sourcePath, noWorkspace)
		if err != nil {
			err := fmt.Errorf("failed to extract subtree: %w", err)
			return ctx.HandleError(err)
		}

		// Handle JSON output for regular peek
		if cmdutil.IsJSONOutput(ctx.Cmd) {
			return outputPeekJSON(ctx, args[0], sourcePath, subtree, ws)
		}

		// Display subtree information if requested
		if info {
			printSubtreeInfo(subtree, sourcePath.File)
			fmt.Println()
		}

		// Display the subtree content
		if raw {
			// Raw mode: output just the content without any formatting
			os.Stdout.Write(subtree.Content)
		} else {
			// Formatted mode: clean content output without header
			// Remove trailing newlines for cleaner output
			content := subtree.Content
			for len(content) > 0 && content[len(content)-1] == '\n' {
				content = content[:len(content)-1]
			}

			fmt.Println(string(content))
		}

		return nil
	},
}

// showWholeFile displays the entire content of a file
func showWholeFile(ws *workspace.Workspace, filename string, raw bool, info bool, noWorkspace bool) error {
	// Construct full file path using the new resolution function
	filePath := resolvePeekFilePath(ws, filename, noWorkspace)

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return cmdutil.NewFileError("read", filename, err)
	}

	// Display file information if requested
	if info {
		cmdutil.ShowInfo("File Information:")
		cmdutil.ShowInfo("  File: %s", filename)
		cmdutil.ShowInfo("  Path: %s", filePath)
		cmdutil.ShowInfo("  Content length: %d bytes", len(content))
		cmdutil.ShowInfo("  Lines: %d", strings.Count(string(content), "\n")+1)
		fmt.Println()
	}

	// Display the file content
	if raw {
		// Raw mode: output just the content without any formatting
		os.Stdout.Write(content)
	} else {
		// Formatted mode: show with nice header
		fmt.Printf("# File: %s\n\n", filename)

		// Remove trailing newlines for cleaner output
		for len(content) > 0 && content[len(content)-1] == '\n' {
			content = content[:len(content)-1]
		}

		fmt.Println(string(content))
	}

	return nil
}

// showWholeFileJSON outputs the whole file content in JSON format
func showWholeFileJSON(ctx *cmdutil.CommandContext, ws *workspace.Workspace, filename string, noWorkspace bool) error {
	// Use the same file resolution logic as the non-JSON path
	filePath := resolvePeekFilePath(ws, filename, noWorkspace)

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		err := cmdutil.NewFileError("read", filename, err)
		return ctx.HandleError(err)
	}

	response := map[string]interface{}{
		"operation": "peek_file",
		"selector":  filename,
		"file": map[string]interface{}{
			"name":           filename,
			"path":           filePath,
			"content":        string(content),
			"content_length": len(content),
			"line_count":     strings.Count(string(content), "\n") + 1,
		},
		"metadata": cmdutil.CreateJSONMetadata(ctx.Cmd, true, ctx.StartTime),
	}

	return cmdutil.OutputJSON(response)
}

// printSubtreeInfo displays metadata about the subtree
func printSubtreeInfo(subtree *markdown.Subtree, filename string) {
	cmdutil.ShowInfo("Subtree Information:")
	cmdutil.ShowInfo("  File: %s", filename)
	cmdutil.ShowInfo("  Heading: %q", subtree.Heading)
	cmdutil.ShowInfo("  Level: %d", subtree.Level)
	cmdutil.ShowInfo("  Content length: %d bytes", len(subtree.Content))
	cmdutil.ShowInfo("  Byte range: %d-%d", subtree.StartOffset, subtree.EndOffset)

	// Count nested headings
	nestedCount := countNestedHeadings(subtree.Content, subtree.Level)
	if nestedCount > 0 {
		cmdutil.ShowInfo("  Nested headings: %d", nestedCount)
	}
}

// countNestedHeadings counts how many headings are nested within this subtree
func countNestedHeadings(content []byte, baseLevel int) int {
	lines := splitLines(content)
	count := 0

	for _, line := range lines {
		if len(line) > 0 && line[0] == '#' {
			// Count the heading level
			level := 0
			for i := 0; i < len(line) && line[i] == '#'; i++ {
				level++
			}

			// Only count headings deeper than the base level
			if level > baseLevel {
				count++
			}
		}
	}

	return count
}

// splitLines splits content into lines
func splitLines(content []byte) []string {
	if len(content) == 0 {
		return []string{}
	}

	var lines []string
	start := 0

	for i := 0; i < len(content); i++ {
		if content[i] == '\n' {
			lines = append(lines, string(content[start:i]))
			start = i + 1
		}
	}

	// Add the last line - even if it's empty due to trailing newline
	if start <= len(content) {
		lines = append(lines, string(content[start:]))
	}

	return lines
}

// showTableOfContents displays a table of contents for a file or subtree
func showTableOfContents(ws *workspace.Workspace, selector string, useShortSelectors bool, noWorkspace bool) error {
	// Check if this is a simple file name or a path selector
	var filename string
	var baseFilename string
	var subtreePath string
	var err error

	if strings.Contains(selector, "#") {
		// This is a path selector - extract the subtree first
		sourcePath, parseErr := markdown.ParsePath(selector)
		if parseErr != nil {
			return fmt.Errorf("invalid selector: %w", parseErr)
		}

		baseFilename = sourcePath.File
		subtreePath = strings.Join(sourcePath.Segments, "/")
		filename = fmt.Sprintf("%s#%s", baseFilename, subtreePath)
	} else {
		// This is just a file name
		baseFilename = selector
		filename = selector
		var filePath string

		if selector == "inbox.md" {
			filePath = ws.InboxPath
		} else if filepath.IsAbs(selector) {
			filePath = selector
		} else {
			// Use workspace root for relative paths, not lib/ directory
			if !strings.HasSuffix(selector, ".md") {
				selector += ".md"
				baseFilename = selector
				filename = selector
			}
			filePath = resolvePeekFilePath(ws, selector, noWorkspace)
		}

		// Check if file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", selector)
		}

		// Read file content
		_, err = os.ReadFile(filePath)
		if err != nil {
			return cmdutil.NewFileError("read", selector, err)
		}
	}

	// Parse document and extract headings
	doc := markdown.ParseMarkdownAST([]byte{})
	headings := extractHeadingsFromContent(doc, []byte{})

	if len(headings) == 0 {
		cmdutil.ShowInfo("No headings found in %s", filename)
		return nil
	}

	// Detect unselectable headings
	unselectableHeadings := detectUnselectableHeadings(headings)

	// Display table of contents
	fmt.Printf("Table of Contents: %s\n", filename)
	fmt.Printf("%s\n\n", strings.Repeat("=", len("Table of Contents: ")+len(filename)))

	for i, heading := range headings {
		// Create indentation based on heading level
		indent := strings.Repeat("  ", heading.Level-1)

		// Format the heading line
		fmt.Printf("%s%s %s", indent, strings.Repeat("#", heading.Level), heading.Text)

		// Check if this heading is unselectable
		// Mark unselectable headings with a warning indicator
		if unselectableHeadings[i] {
			fmt.Printf(" ⚠️")
		}
		fmt.Println()

		// Create accurate selector hint for navigation
		if subtreePath == "" { // Full file TOC
			var selectorHint string
			if useShortSelectors {
				selectorHint = generateShortSelector(baseFilename, heading, headings)
			} else {
				selectorHint = generateOptimalSelector(baseFilename, heading, headings)
			}
			fmt.Printf("%s%s\n", indent, fmt.Sprintf("  → %s", selectorHint))
		}

		// Add spacing between entries for readability
		if i < len(headings)-1 {
			fmt.Println()
		}
	}

	// Add helpful usage notes at the end
	fmt.Println()
	if subtreePath == "" {
		fmt.Printf("Use 'jot peek \"<selector>\"' to view specific sections.\n")
		fmt.Printf("Tip: Heading names are matched case-insensitively using 'contains' logic.\n")

		// Check if there are any unselectable headings
		hasUnselectable := false
		for _, unsel := range unselectableHeadings {
			if unsel {
				hasUnselectable = true
				break
			}
		}
		if hasUnselectable {
			cmdutil.ShowWarning("Warning: Headings marked with ⚠️ may have ambiguous selectors.")
		}
	} else {
		fmt.Printf("This is a table of contents for the subtree '%s'.\n", subtreePath)
		fmt.Printf("Use 'jot peek \"%s#%s/<subheading>\"' to view nested sections.\n", baseFilename, subtreePath)
	}

	return nil
}

// HeadingInfo represents a heading with its metadata
// (now matches markdown.HeadingInfo)
type HeadingInfo = markdown.HeadingInfo

// extractHeadingsFromContent extracts all headings from markdown content
func extractHeadingsFromContent(doc ast.Node, _ []byte) []HeadingInfo {
	return markdown.FindAllHeadings(doc)
}

// HeadingResponse represents the JSON response for peek command
type HeadingResponse struct {
	Selector        string               `json:"selector"`
	Subtree         *HeadingSubtree         `json:"subtree,omitempty"`
	FileInfo        HeadingFileInfo         `json:"file_info"`
	Extraction      *HeadingExtraction      `json:"extraction,omitempty"`
	TableOfContents *HeadingTOC             `json:"table_of_contents,omitempty"`
	Metadata        cmdutil.JSONMetadata `json:"metadata"`
}

type HeadingSubtree struct {
	Heading        string `json:"heading"`
	Level          int    `json:"level"`
	Content        string `json:"content"`
	NestedHeadings int    `json:"nested_headings"`
	LineCount      int    `json:"line_count"`
}

type HeadingFileInfo struct {
	FilePath     string  `json:"file_path"`
	FileExists   bool    `json:"file_exists"`
	LastModified *string `json:"last_modified,omitempty"`
}

type HeadingExtraction struct {
	StartLine     int    `json:"start_line"`
	EndLine       int    `json:"end_line"`
	ContentOffset [2]int `json:"content_offset"`
}

type HeadingTOC struct {
	IsFullFile   bool             `json:"is_full_file"`
	RootSelector string           `json:"root_selector,omitempty"`
	Headings     []HeadingTOCHeading `json:"headings"`
}

type HeadingTOCHeading struct {
	Text     string `json:"text"`
	Level    int    `json:"level"`
	Selector string `json:"selector"`
}

// outputPeekJSON outputs JSON response for regular peek mode
func outputPeekJSON(ctx *cmdutil.CommandContext, selector string, sourcePath *markdown.HeadingPath, subtree *markdown.Subtree, ws *workspace.Workspace) error {
	pathUtil := cmdutil.NewPathUtil(ws)
	// Build file info
	filePath := pathUtil.WorkspaceJoin(sourcePath.File)
	if sourcePath.File == "inbox.md" {
		filePath = ws.InboxPath
	}

	fileExists := true
	var lastModified *string
	if info, err := os.Stat(filePath); err == nil {
		modTime := info.ModTime().Format(time.RFC3339)
		lastModified = &modTime
	} else {
		fileExists = false
	}

	// Count nested headings and lines
	content := string(subtree.Content)
	lineCount := strings.Count(content, "\n") + 1
	if len(content) == 0 {
		lineCount = 0
	}

	// Count nested headings by parsing content
	doc := markdown.ParseMarkdownAST(subtree.Content)
	headings := extractHeadingsFromContent(doc, subtree.Content)
	nestedCount := len(headings)
	if nestedCount > 0 {
		nestedCount-- // Don't count the root heading itself
	}

	response := HeadingResponse{
		Selector: selector,
		Subtree: &HeadingSubtree{
			Heading:        subtree.Heading,
			Level:          subtree.Level,
			Content:        content,
			NestedHeadings: nestedCount,
			LineCount:      lineCount,
		},
		FileInfo: HeadingFileInfo{
			FilePath:     filePath,
			FileExists:   fileExists,
			LastModified: lastModified,
		},
		Extraction: &HeadingExtraction{
			StartLine:     0, // We don't have line info from markdown.Subtree
			EndLine:       0, // We don't have line info from markdown.Subtree
			ContentOffset: [2]int{subtree.StartOffset, subtree.EndOffset},
		},
		Metadata: cmdutil.CreateJSONMetadata(ctx.Cmd, true, ctx.StartTime),
	}

	return cmdutil.OutputJSON(response)
}

// showTableOfContentsJSON outputs JSON response for TOC mode
func showTableOfContentsJSON(ctx *cmdutil.CommandContext, ws *workspace.Workspace, selector string, useShortSelectors bool) error {
	pathUtil := cmdutil.NewPathUtil(ws)
	// Parse selector to determine if it's file-only or includes path
	var baseFilename string
	var subtreePath string
	var filePath string
	var err error
	isFullFile := true

	if strings.Contains(selector, "#") {
		// This is a path selector - extract the subtree first
		sourcePath, parseErr := markdown.ParsePath(selector)
		if parseErr != nil {
			return ctx.HandleError(fmt.Errorf("invalid selector: %w", parseErr))
		}

		baseFilename = sourcePath.File
		subtreePath = strings.Join(sourcePath.Segments, "/")
		isFullFile = false

		filePath = pathUtil.WorkspaceJoin(baseFilename)
		if baseFilename == "inbox.md" {
			filePath = ws.InboxPath
		}
	} else {
		// This is just a file name
		baseFilename = selector

		if selector == "inbox.md" {
			filePath = ws.InboxPath
		} else if filepath.IsAbs(selector) {
			filePath = selector
		} else {
			if !strings.HasSuffix(selector, ".md") {
				selector += ".md"
				baseFilename = selector
			}
			filePath = pathUtil.WorkspaceJoin(selector)
		}

		// Check if file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return ctx.HandleError(fmt.Errorf("file not found: %s", selector))
		}

		// Read file content
		_, err = os.ReadFile(filePath)
		if err != nil {
			return ctx.HandleError(err)
		}
	}

	// Parse document and extract headings
	doc := markdown.ParseMarkdownAST([]byte{})
	headings := extractHeadingsFromContent(doc, []byte{})

	if len(headings) == 0 {
		// No headings case
		response := HeadingResponse{
			Selector: selector,
			FileInfo: HeadingFileInfo{
				FilePath:   filePath,
				FileExists: true,
			},
			TableOfContents: &HeadingTOC{
				IsFullFile:   isFullFile,
				RootSelector: subtreePath,
				Headings:     []HeadingTOCHeading{},
			},
			Metadata: cmdutil.CreateJSONMetadata(ctx.Cmd, true, ctx.StartTime),
		}
		return cmdutil.OutputJSON(response)
	}

	// Build TOC headings
	tocHeadings := []HeadingTOCHeading{}
	for _, heading := range headings {
		var selectorText string
		if useShortSelectors {
			selectorText = generateShortSelector(baseFilename, heading, headings)
		} else {
			// Build full path
			pathSegments := buildPathToHeading(heading, headings)
			if len(pathSegments) > 0 {
				selectorText = fmt.Sprintf("%s#%s", baseFilename, strings.Join(pathSegments, "/"))
			} else {
				selectorText = fmt.Sprintf("%s#%s", baseFilename, strings.ToLower(heading.Text))
			}
		}

		tocHeadings = append(tocHeadings, HeadingTOCHeading{
			Text:     heading.Text,
			Level:    heading.Level,
			Selector: selectorText,
		})
	}

	response := HeadingResponse{
		Selector: selector,
		FileInfo: HeadingFileInfo{
			FilePath:   filePath,
			FileExists: true,
		},
		TableOfContents: &HeadingTOC{
			IsFullFile:   isFullFile,
			RootSelector: subtreePath,
			Headings:     tocHeadings,
		},
		Metadata: cmdutil.CreateJSONMetadata(ctx.Cmd, true, ctx.StartTime),
	}

	return cmdutil.OutputJSON(response)
}

// buildPathToHeading builds a hierarchical path array for a heading based on the document structure
func buildPathToHeading(target HeadingInfo, allHeadings []HeadingInfo) []string {
	// Find target index
	var targetIndex int = -1
	for i, h := range allHeadings {
		if h.Text == target.Text && h.Level == target.Level {
			targetIndex = i
			break
		}
	}

	if targetIndex == -1 {
		return []string{target.Text}
	}

	var path []string

	// Build path by walking backward to find parent headings
	currentLevel := target.Level
	for i := targetIndex; i >= 0; i-- {
		heading := allHeadings[i]
		if heading.Level < currentLevel {
			// This is a parent heading
			path = append([]string{strings.ToLower(heading.Text)}, path...)
			currentLevel = heading.Level
		} else if i == targetIndex {
			// Add the target heading itself
			path = append(path, strings.ToLower(heading.Text))
		}
	}

	return path
}

// parseEnhancedSelector handles enhanced selectors with line numbers
// Converts "file:42" to "file:42#heading/path" or "file:42#heading" to "file#heading"
func parseEnhancedSelector(ws *workspace.Workspace, selector string) (string, error) {
	// Check if selector contains a line number (has ":" but not necessarily "#")
	colonIndex := strings.Index(selector, ":")
	if colonIndex == -1 {
		// No line number, return as-is
		return selector, nil
	}

	hashIndex := strings.Index(selector, "#")

	var filename string
	var lineNumStr string
	var headingPart string

	if hashIndex == -1 {
		// Format: "file:42" (no heading part)
		filename = selector[:colonIndex]
		lineNumStr = selector[colonIndex+1:]
	} else if hashIndex > colonIndex {
		// Format: "file:42#heading/path"
		filename = selector[:colonIndex]
		lineNumStr = selector[colonIndex+1 : hashIndex]
		headingPart = selector[hashIndex:] // includes the #
	} else {
		// Hash comes before colon, this is invalid for our enhanced format
		return selector, nil
	}

	// Parse line number
	var lineNum int
	if _, err := fmt.Sscanf(lineNumStr, "%d", &lineNum); err != nil {
		// Not a valid line number, return as-is
		return selector, nil
	}

	// If we already have a heading part, just remove the line number
	if headingPart != "" {
		return filename + headingPart, nil
	}

	// Need to resolve line number to heading path
	filePath := filepath.Join(ws.Root, filename)
	if !filepath.IsAbs(filename) {
		// Try inbox
		if filename == "inbox.md" {
			filePath = ws.InboxPath
		} else {
			// Try lib directory
			filePath = filepath.Join(ws.LibDir, filename)
		}
	}

	// Read file content
	_, err := os.ReadFile(filePath)
	if err != nil {
		// Can't read file, return original selector
		return selector, nil
	}

	// Find heading for this line
	// (Feature not supported in new AST; stub or remove as needed)
	headingMap := make(map[int]string)
	if len(headingMap) == 0 {
		return filename, nil
	}

	if headingPath, found := headingMap[lineNum]; found && headingPath != "" {
		// Found heading, create enhanced selector
		return fmt.Sprintf("%s#%s", filename, headingPath), nil
	}

	// No heading found, return whole file selector
	return filename, nil
}

// detectUnselectableHeadings identifies headings that cannot be uniquely selected by path
func detectUnselectableHeadings(headings []HeadingInfo) map[int]bool {
	unselectable := make(map[int]bool)
	pathGroups := make(map[string][]int)
	for i, heading := range headings {
		pathKey := strings.ToLower(strings.Join(heading.Path, "/"))
		pathGroups[pathKey] = append(pathGroups[pathKey], i)
	}
	for _, indices := range pathGroups {
		if len(indices) > 1 {
			for _, idx := range indices {
				unselectable[idx] = true
			}
		}
	}
	return unselectable
}

// generateOptimalSelector creates a selector for a heading using its path
func generateOptimalSelector(filename string, target HeadingInfo, allHeadings []HeadingInfo) string {
	if len(target.Path) > 0 {
		return fmt.Sprintf("jot peek \"%s#%s\"", filename, strings.ToLower(strings.Join(target.Path, "/")))
	}
	return fmt.Sprintf("jot peek \"%s#%s\"", filename, strings.ToLower(target.Text))
}

// generateShortSelector creates a short selector for a heading (aggressive, but unique)
func generateShortSelector(filename string, target HeadingInfo, allHeadings []HeadingInfo) string {
	// Try shortest unique segment from path
	for i := len(target.Path) - 1; i >= 0; i-- {
		seg := target.Path[i]
		selector := fmt.Sprintf("jot peek \"%s#%s\"", filename, strings.ToLower(seg))
		count := 0
		for _, h := range allHeadings {
			if strings.Contains(strings.ToLower(h.Text), strings.ToLower(seg)) {
				count++
			}
		}
		if count == 1 {
			return selector
		}
	}
	// Fallback to full path
	return generateOptimalSelector(filename, target, allHeadings)
}

// resolvePeekFilePath consolidates file path resolution logic for peek operations
func resolvePeekFilePath(ws *workspace.Workspace, filename string, noWorkspace bool) string {
	if noWorkspace {
		// Non-workspace mode: resolve relative to current directory
		if filepath.IsAbs(filename) {
			return filename
		}
		cwd, _ := os.Getwd()
		return filepath.Join(cwd, filename)
	}

	// Workspace mode: existing logic
	if filename == "inbox.md" && ws != nil {
		return ws.InboxPath
	}
	if filepath.IsAbs(filename) {
		return filename
	}
	if ws != nil {
		return filepath.Join(ws.Root, filename)
	}
	return filename // Fallback
}

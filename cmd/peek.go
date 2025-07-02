package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/johncoder/jot/internal/cmdutil"
	"github.com/johncoder/jot/internal/markdown"
	"github.com/johncoder/jot/internal/workspace"
	"github.com/spf13/cobra"
	"github.com/yuin/goldmark/ast"
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
	var content []byte
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

		subtree, extractErr := ExtractSubtreeWithOptions(ws, sourcePath, noWorkspace)
		if extractErr != nil {
			return fmt.Errorf("failed to extract subtree: %w", extractErr)
		}

		content = subtree.Content
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
		content, err = os.ReadFile(filePath)
		if err != nil {
			return cmdutil.NewFileError("read", selector, err)
		}
	}

	if len(content) == 0 {
		cmdutil.ShowInfo("File %s is empty (no table of contents available)", filename)
		return nil
	}

	// Parse document and extract headings
	doc := markdown.ParseDocument(content)
	headings := extractHeadingsFromContent(doc, content)

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
type HeadingInfo struct {
	Text  string
	Level int
	Line  int
}

// extractHeadingsFromContent extracts all headings from markdown content
func extractHeadingsFromContent(doc ast.Node, content []byte) []HeadingInfo {
	var headings []HeadingInfo

	// Walk the AST to find all headings
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if heading, ok := n.(*ast.Heading); ok {
			headingText := markdown.ExtractHeadingText(heading, content)
			if strings.TrimSpace(headingText) != "" {
				offset := markdown.GetNodeOffset(heading, content)
				lineNum := markdown.CalculateLineNumber(content, offset)

				headings = append(headings, HeadingInfo{
					Text:  headingText,
					Level: heading.Level,
					Line:  lineNum,
				})
			}
		}

		return ast.WalkContinue, nil
	})

	return headings
}

// HeadingTrie represents a trie structure for heading hierarchies
type HeadingTrie struct {
	root *TrieNode
}

// TrieNode represents a node in the heading trie
type TrieNode struct {
	text     string               // The heading text
	level    int                  // The heading level (1-6)
	children map[string]*TrieNode // Child nodes (key is normalized text for matching)
	heading  *ast.Heading         // Reference to the actual AST node
	offset   int                  // Byte offset in the document
	isLeaf   bool                 // Whether this is a selectable leaf (has no children)
	fullPath []string             // Full path from root to this node
}

// NewHeadingTrie creates a new trie from document content
func NewHeadingTrie(doc ast.Node, content []byte) *HeadingTrie {
	trie := &HeadingTrie{
		root: &TrieNode{
			children: make(map[string]*TrieNode),
			level:    0,
		},
	}

	// Build the trie from the document
	trie.buildFromDocument(doc, content)

	// Mark leaf nodes and build full paths
	trie.markLeavesAndPaths()

	return trie
}

// buildFromDocument constructs the trie from the markdown document
func (t *HeadingTrie) buildFromDocument(doc ast.Node, content []byte) {
	headings := markdown.FindAllHeadings(doc, content)

	// Track current path at each level
	pathStack := make([]*TrieNode, 7) // Support levels 1-6
	pathStack[0] = t.root

	for _, headingInfo := range headings {
		level := headingInfo.Level
		text := headingInfo.Text
		normalizedText := normalizeForMatching(text)

		// Find the parent node (closest ancestor at a lower level)
		var parent *TrieNode
		for i := level - 1; i >= 0; i-- {
			if pathStack[i] != nil {
				parent = pathStack[i]
				break
			}
		}

		if parent == nil {
			parent = t.root
		}

		// Create or get the node for this heading
		var node *TrieNode
		if existing, exists := parent.children[normalizedText]; exists {
			node = existing
		} else {
			node = &TrieNode{
				text:     text,
				level:    level,
				children: make(map[string]*TrieNode),
				offset:   headingInfo.Offset,
			}
			parent.children[normalizedText] = node
		}

		// Update the heading reference (in case of duplicates, use the first one)
		if node.heading == nil {
			node.heading = findHeadingByOffset(doc, headingInfo.Offset)
		}

		// Update the path stack
		pathStack[level] = node
		for i := level + 1; i < 7; i++ {
			pathStack[i] = nil
		}
	}
}

// markLeavesAndPaths marks leaf nodes and builds full paths
func (t *HeadingTrie) markLeavesAndPaths() {
	t.markLeavesAndPathsRecursive(t.root, []string{})
}

func (t *HeadingTrie) markLeavesAndPathsRecursive(node *TrieNode, path []string) {
	// Build full path for this node
	if node.text != "" {
		node.fullPath = append(path, node.text)
	} else {
		node.fullPath = path
	}

	// Mark as leaf if no children
	node.isLeaf = len(node.children) == 0 && node.text != ""

	// Recursively process children
	for _, child := range node.children {
		childPath := node.fullPath
		t.markLeavesAndPathsRecursive(child, childPath)
	}
}

// normalizeForMatching normalizes text for case-insensitive contains matching
func normalizeForMatching(text string) string {
	return strings.ToLower(strings.TrimSpace(text))
}

// GenerateSelector creates an accurate selector path for a heading
func (t *HeadingTrie) GenerateSelector(filename string, targetHeading *ast.Heading, content []byte) (string, error) {
	// Find the node for this heading
	targetOffset := markdown.GetNodeOffset(targetHeading, content)

	node := t.findNodeByOffset(targetOffset)
	if node == nil {
		return "", fmt.Errorf("heading not found in trie")
	}

	// Try different selector strategies based on the position in the hierarchy
	selectors := t.generateSelectorCandidates(filename, node)

	// Return the most concise selector that would uniquely identify this heading
	for _, selector := range selectors {
		if t.isUniqueSelector(selector, node) {
			return selector, nil
		}
	}

	// Fallback to full path
	if len(node.fullPath) > 0 {
		return fmt.Sprintf("%s#%s", filename, strings.ToLower(strings.Join(node.fullPath, "/"))), nil
	}

	return "", fmt.Errorf("unable to generate selector for heading")
}

// generateSelectorCandidates creates multiple selector candidates in order of preference
func (t *HeadingTrie) generateSelectorCandidates(filename string, node *TrieNode) []string {
	var candidates []string

	if len(node.fullPath) == 0 {
		return candidates
	}

	// Strategy 1: Direct heading match (shortest)
	lastSegment := node.fullPath[len(node.fullPath)-1]
	candidates = append(candidates, fmt.Sprintf("%s#%s", filename, strings.ToLower(lastSegment)))

	// Strategy 2: Parent/child pattern
	if len(node.fullPath) >= 2 {
		parentSegment := node.fullPath[len(node.fullPath)-2]
		candidates = append(candidates, fmt.Sprintf("%s#%s/%s", filename,
			strings.ToLower(parentSegment), strings.ToLower(lastSegment)))
	}

	// Strategy 3: Skip-level syntax if not at level 1
	if node.level > 1 {
		skipPrefix := strings.Repeat("/", node.level-1)
		candidates = append(candidates, fmt.Sprintf("%s#%s%s", filename,
			skipPrefix, strings.ToLower(lastSegment)))
	}

	// Strategy 4: Full path
	candidates = append(candidates, fmt.Sprintf("%s#%s", filename,
		strings.ToLower(strings.Join(node.fullPath, "/"))))

	return candidates
}

// isUniqueSelector checks if a selector would uniquely identify the target node
func (t *HeadingTrie) isUniqueSelector(selector string, targetNode *TrieNode) bool {
	// Parse the selector to get the path segments
	parts := strings.Split(selector, "#")
	if len(parts) != 2 {
		return false
	}

	pathPart := parts[1]

	// Count how many nodes would match this selector
	matches := t.findMatchingNodes(pathPart)

	// Check if exactly one match and it's our target
	if len(matches) == 1 && matches[0] == targetNode {
		return true
	}

	return false
}

// findMatchingNodes finds all nodes that would match a given path selector
func (t *HeadingTrie) findMatchingNodes(pathStr string) []*TrieNode {
	var matches []*TrieNode

	// Parse path segments and skip levels
	segments := strings.Split(pathStr, "/")
	skipLevels := 0

	// Count leading empty segments (skip levels)
	for i, segment := range segments {
		if segment == "" {
			skipLevels++
		} else {
			segments = segments[i:]
			break
		}
	}

	if len(segments) == 0 {
		return matches
	}

	// Find all matching nodes using the same logic as the path navigation
	t.findMatchingNodesRecursive(t.root, segments, skipLevels, 0, &matches)

	return matches
}

// findMatchingNodesRecursive recursively searches for matching nodes
func (t *HeadingTrie) findMatchingNodesRecursive(node *TrieNode, segments []string, skipLevels, currentLevel int, matches *[]*TrieNode) {
	// If we have segments to match
	if len(segments) > 0 {
		targetSegment := normalizeForMatching(segments[0])

		// Check all children
		for normalizedText, child := range node.children {
			// Check if this child matches the current segment (contains matching)
			if strings.Contains(normalizedText, targetSegment) {
				if len(segments) == 1 {
					// This is the final segment, add to matches
					*matches = append(*matches, child)
				} else {
					// Continue with remaining segments
					t.findMatchingNodesRecursive(child, segments[1:], skipLevels, currentLevel+1, matches)
				}
			}

			// Also continue searching deeper if we haven't used up our skip levels
			if skipLevels > 0 {
				t.findMatchingNodesRecursive(child, segments, skipLevels-1, currentLevel+1, matches)
			}
		}
	}
}

// findNodeByOffset finds a node in the trie by its byte offset
func (t *HeadingTrie) findNodeByOffset(offset int) *TrieNode {
	return t.findNodeByOffsetRecursive(t.root, offset)
}

func (t *HeadingTrie) findNodeByOffsetRecursive(node *TrieNode, offset int) *TrieNode {
	if node.offset == offset && node.text != "" {
		return node
	}

	for _, child := range node.children {
		if result := t.findNodeByOffsetRecursive(child, offset); result != nil {
			return result
		}
	}

	return nil
}

// GetUnselectableHeadings returns headings that cannot be uniquely selected
func (t *HeadingTrie) GetUnselectableHeadings() []*TrieNode {
	var unselectable []*TrieNode
	t.findUnselectableRecursive(t.root, &unselectable)
	return unselectable
}

func (t *HeadingTrie) findUnselectableRecursive(node *TrieNode, unselectable *[]*TrieNode) {
	if node.text != "" {
		// Check if this node has duplicate text at the same level or ambiguous paths
		if t.hasAmbiguousSelector(node) {
			*unselectable = append(*unselectable, node)
		}
	}

	for _, child := range node.children {
		t.findUnselectableRecursive(child, unselectable)
	}
}

// hasAmbiguousSelector checks if a node's selector would be ambiguous
func (t *HeadingTrie) hasAmbiguousSelector(node *TrieNode) bool {
	// For now, check if there are other nodes with the same normalized text
	normalizedText := normalizeForMatching(node.text)
	count := 0

	t.countNodesWithText(t.root, normalizedText, &count)

	return count > 1
}

func (t *HeadingTrie) countNodesWithText(node *TrieNode, targetText string, count *int) {
	if node.text != "" && normalizeForMatching(node.text) == targetText {
		*count++
	}

	for _, child := range node.children {
		t.countNodesWithText(child, targetText, count)
	}
}

func init() {
	// Add flags
	peekCmd.Flags().BoolP("raw", "r", false, "Output raw content without formatting")
	peekCmd.Flags().BoolP("info", "i", false, "Show subtree metadata information")
	peekCmd.Flags().BoolP("toc", "t", false, "Show table of contents for file or subtree")
	peekCmd.Flags().BoolP("short", "s", false, "Generate shortest possible selectors (use with --toc)")
	peekCmd.Flags().Bool("no-workspace", false, "Resolve file paths relative to current directory instead of workspace")

	// Add to root command
	rootCmd.AddCommand(peekCmd)
}

// generateOptimalSelector creates the best selector for a heading using simple heuristics
func generateOptimalSelector(filename string, target HeadingInfo, allHeadings []HeadingInfo) string {
	targetText := normalizeForMatching(target.Text)

	// Build hierarchical path first - this ensures compatibility with path resolution
	path := buildHierarchicalPath(target, allHeadings)

	// Strategy 1: For level 1 headings, use simple selector if unique
	if target.Level == 1 {
		matchCount := 0
		for _, h := range allHeadings {
			if strings.Contains(normalizeForMatching(h.Text), targetText) {
				matchCount++
			}
		}

		if matchCount == 1 {
			return fmt.Sprintf("jot peek \"%s#%s\"", filename, strings.ToLower(target.Text))
		}
	}

	// Strategy 2: Use hierarchical path for deeper headings or non-unique level 1 headings
	if len(path) > 1 {
		pathStr := strings.Join(path, "/")
		return fmt.Sprintf("jot peek \"%s#%s\"", filename, strings.ToLower(pathStr))
	}

	// Strategy 3: Fall back to skip-level syntax for deeper headings
	if target.Level > 1 {
		skipPrefix := strings.Repeat("/", target.Level-1)
		return fmt.Sprintf("jot peek \"%s#%s%s\"", filename, skipPrefix, strings.ToLower(target.Text))
	}

	// Strategy 4: Final fallback
	return fmt.Sprintf("jot peek \"%s#%s\"", filename, strings.ToLower(target.Text))
}

// buildHierarchicalPath builds the path from root to target heading
func buildHierarchicalPath(target HeadingInfo, allHeadings []HeadingInfo) []string {
	var path []string

	// Find the target heading in the list
	targetIndex := -1
	for i, h := range allHeadings {
		if h.Line == target.Line && h.Text == target.Text && h.Level == target.Level {
			targetIndex = i
			break
		}
	}

	if targetIndex == -1 {
		return []string{target.Text}
	}

	// Build path by walking backwards to find parent headings
	var parents []HeadingInfo
	currentLevel := target.Level

	for i := targetIndex - 1; i >= 0; i-- {
		h := allHeadings[i]
		if h.Level < currentLevel {
			parents = append([]HeadingInfo{h}, parents...)
			currentLevel = h.Level
			if h.Level == 1 {
				break // Stop at level 1
			}
		}
	}

	// Build the path segments
	for _, parent := range parents {
		path = append(path, parent.Text)
	}
	path = append(path, target.Text)

	return path
}

// detectUnselectableHeadings identifies headings that cannot be uniquely selected
func detectUnselectableHeadings(headings []HeadingInfo) map[int]bool {
	unselectable := make(map[int]bool)

	// Group headings by their hierarchical paths
	pathGroups := make(map[string][]int)

	for i, heading := range headings {
		path := buildHierarchicalPath(heading, headings)
		pathKey := strings.ToLower(strings.Join(path, "/"))
		pathGroups[pathKey] = append(pathGroups[pathKey], i)
	}

	// Mark headings as unselectable if they share the same path
	for _, indices := range pathGroups {
		if len(indices) > 1 {
			// Multiple headings with the same path - they're unselectable
			for _, idx := range indices {
				unselectable[idx] = true
			}
		}
	}

	return unselectable
}

// generateShortSelector creates the most aggressively short selector possible
func generateShortSelector(filename string, target HeadingInfo, allHeadings []HeadingInfo) string {
	targetText := normalizeForMatching(target.Text)

	// Strategy 1: Single letter shortcuts for very common terms
	singleLetterShortcuts := map[string]string{
		"go":         "g",
		"javascript": "j",
		"python":     "p",
		"docker":     "d",
		"kubernetes": "k",
		"tools":      "t",
		"views":      "v",
		"models":     "m",
		"functions":  "f",
		"classes":    "c",
		"variables":  "v",
		"routing":    "r",
		"templates":  "t",
		"plugins":    "p",
		"jobs":       "j",
		"services":   "s",
		"arrays":     "a",
		"loops":      "l",
	}

	lowerTarget := strings.ToLower(target.Text)
	if shortcut, exists := singleLetterShortcuts[lowerTarget]; exists {
		// Check if this single letter is unique using jot's actual contains matching
		matchCount := 0
		for _, h := range allHeadings {
			if target.Level > 1 {
				if h.Level == target.Level && strings.Contains(normalizeForMatching(h.Text), shortcut) {
					matchCount++
				}
			} else {
				if strings.Contains(normalizeForMatching(h.Text), shortcut) {
					matchCount++
				}
			}
		}
		if matchCount == 1 {
			if target.Level > 1 {
				return fmt.Sprintf("jot peek \"%s#%s%s\"", filename, strings.Repeat("/", target.Level-1), shortcut)
			}
			return fmt.Sprintf("jot peek \"%s#%s\"", filename, shortcut)
		}
	}

	// Strategy 2: Ultra-short unique character sequences (1-3 chars) using contains matching
	for length := 1; length <= 4; length++ {
		if length > len(targetText) {
			break
		}

		prefix := targetText[:length]
		prefixMatches := 0

		// Use jot's actual contains matching logic
		for _, h := range allHeadings {
			if target.Level > 1 {
				// For deeper levels, only check same level using contains
				if h.Level == target.Level && strings.Contains(normalizeForMatching(h.Text), prefix) {
					prefixMatches++
				}
			} else {
				// For top level, check globally using contains
				if strings.Contains(normalizeForMatching(h.Text), prefix) {
					prefixMatches++
				}
			}
		}

		if prefixMatches == 1 {
			if target.Level > 1 {
				return fmt.Sprintf("jot peek \"%s#%s%s\"", filename, strings.Repeat("/", target.Level-1), prefix)
			}
			return fmt.Sprintf("jot peek \"%s#%s\"", filename, prefix)
		}
	}

	// Strategy 3: Unique word initials (first letter of each word)
	words := strings.Fields(strings.ToLower(target.Text))
	if len(words) > 1 {
		initials := ""
		for _, word := range words {
			if len(word) > 0 {
				initials += string(word[0])
			}
		}

		if len(initials) >= 2 && len(initials) <= 4 {
			matchCount := 0
			for _, h := range allHeadings {
				// Test if the heading contains this initials sequence
				if target.Level > 1 {
					if h.Level == target.Level && strings.Contains(normalizeForMatching(h.Text), initials) {
						matchCount++
					}
				} else {
					if strings.Contains(normalizeForMatching(h.Text), initials) {
						matchCount++
					}
				}
			}

			if matchCount == 1 {
				if target.Level > 1 {
					return fmt.Sprintf("jot peek \"%s#%s%s\"", filename, strings.Repeat("/", target.Level-1), initials)
				}
				return fmt.Sprintf("jot peek \"%s#%s\"", filename, initials)
			}
		}
	}

	// Strategy 4: Consonants only (aggressive compression) using contains matching
	consonants := extractConsonants(strings.ToLower(target.Text))
	if len(consonants) >= 2 && len(consonants) <= 6 {
		matchCount := 0
		for _, h := range allHeadings {
			hConsonants := extractConsonants(normalizeForMatching(h.Text))
			if target.Level > 1 {
				if h.Level == target.Level && strings.Contains(hConsonants, consonants) {
					matchCount++
				}
			} else {
				if strings.Contains(hConsonants, consonants) {
					matchCount++
				}
			}
		}

		if matchCount == 1 {
			if target.Level > 1 {
				return fmt.Sprintf("jot peek \"%s#%s%s\"", filename, strings.Repeat("/", target.Level-1), consonants)
			}
			return fmt.Sprintf("jot peek \"%s#%s\"", filename, consonants)
		}
	}

	// Strategy 5: Smart skip-level optimization - use minimum required skips
	if target.Level > 1 {
		// Try with minimal skip levels that still work
		for skipCount := 1; skipCount < target.Level; skipCount++ {
			skipPrefix := strings.Repeat("/", skipCount)

			// Try ultra-short prefixes with minimal skips using contains matching
			for length := 1; length <= 5; length++ {
				if length > len(targetText) {
					break
				}

				prefix := targetText[:length]
				prefixMatches := 0

				for _, h := range allHeadings {
					if h.Level >= target.Level-skipCount && strings.Contains(normalizeForMatching(h.Text), prefix) {
						prefixMatches++
					}
				}

				if prefixMatches == 1 {
					return fmt.Sprintf("jot peek \"%s#%s%s\"", filename, skipPrefix, prefix)
				}
			}
		}

		// Fall back to standard skip-level with first word or short prefix
		skipPrefix := strings.Repeat("/", target.Level-1)
		words := strings.Fields(strings.ToLower(target.Text))

		if len(words) > 0 {
			firstWord := words[0]

			// Try just first 2-4 letters of first word using contains matching
			for length := 2; length <= min(5, len(firstWord)); length++ {
				prefix := firstWord[:length]
				prefixMatches := 0

				for _, h := range allHeadings {
					if h.Level == target.Level && strings.Contains(normalizeForMatching(h.Text), prefix) {
						prefixMatches++
					}
				}

				if prefixMatches == 1 {
					return fmt.Sprintf("jot peek \"%s#%s%s\"", filename, skipPrefix, prefix)
				}
			}

			// Use full first word if needed
			return fmt.Sprintf("jot peek \"%s#%s%s\"", filename, skipPrefix, firstWord)
		}

		// Ultra-fallback with shortest possible representation
		return fmt.Sprintf("jot peek \"%s#%s%s\"", filename, skipPrefix, targetText[:min(8, len(targetText))])
	}

	// Strategy 6: For level 1 headings, try first word or aggressive abbreviation
	headingWords := strings.Fields(strings.ToLower(target.Text))
	if len(headingWords) > 0 {
		firstWord := headingWords[0]

		// Try progressively shorter versions of first word using contains matching
		for length := 2; length <= len(firstWord); length++ {
			prefix := firstWord[:length]
			matchCount := 0

			for _, h := range allHeadings {
				if strings.Contains(normalizeForMatching(h.Text), prefix) {
					matchCount++
				}
			}

			if matchCount == 1 {
				return fmt.Sprintf("jot peek \"%s#%s\"", filename, prefix)
			}
		}

		return fmt.Sprintf("jot peek \"%s#%s\"", filename, firstWord)
	}

	// Final ultra-fallback: shortest possible representation
	return fmt.Sprintf("jot peek \"%s#%s\"", filename, targetText[:min(6, len(targetText))])
}

// extractConsonants removes vowels for ultra-compressed representation
func extractConsonants(text string) string {
	vowels := "aeiou"
	result := ""
	for _, char := range strings.ToLower(text) {
		if char >= 'a' && char <= 'z' && !strings.ContainsRune(vowels, char) {
			result += string(char)
		}
	}
	return result
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// PeekResponse represents the JSON response for peek command
type PeekResponse struct {
	Selector        string          `json:"selector"`
	Subtree         *PeekSubtree    `json:"subtree,omitempty"`
	FileInfo        PeekFileInfo    `json:"file_info"`
	Extraction      *PeekExtraction `json:"extraction,omitempty"`
	TableOfContents *PeekTOC        `json:"table_of_contents,omitempty"`
	Metadata        cmdutil.JSONMetadata    `json:"metadata"`
}

type PeekSubtree struct {
	Heading        string `json:"heading"`
	Level          int    `json:"level"`
	Content        string `json:"content"`
	NestedHeadings int    `json:"nested_headings"`
	LineCount      int    `json:"line_count"`
}

type PeekFileInfo struct {
	FilePath     string  `json:"file_path"`
	FileExists   bool    `json:"file_exists"`
	LastModified *string `json:"last_modified,omitempty"`
}

type PeekExtraction struct {
	StartLine     int    `json:"start_line"`
	EndLine       int    `json:"end_line"`
	ContentOffset [2]int `json:"content_offset"`
}

type PeekTOC struct {
	IsFullFile   bool             `json:"is_full_file"`
	RootSelector string           `json:"root_selector,omitempty"`
	Headings     []PeekTOCHeading `json:"headings"`
}

type PeekTOCHeading struct {
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
	doc := markdown.ParseDocument(subtree.Content)
	headings := extractHeadingsFromContent(doc, subtree.Content)
	nestedCount := len(headings)
	if nestedCount > 0 {
		nestedCount-- // Don't count the root heading itself
	}

	response := PeekResponse{
		Selector: selector,
		Subtree: &PeekSubtree{
			Heading:        subtree.Heading,
			Level:          subtree.Level,
			Content:        content,
			NestedHeadings: nestedCount,
			LineCount:      lineCount,
		},
		FileInfo: PeekFileInfo{
			FilePath:     filePath,
			FileExists:   fileExists,
			LastModified: lastModified,
		},
		Extraction: &PeekExtraction{
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
	var content []byte
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

		subtree, extractErr := ExtractSubtreeWithOptions(ws, sourcePath, false) // TODO: Add noWorkspace support to JSON functions
		if extractErr != nil {
			return ctx.HandleError(fmt.Errorf("failed to extract subtree: %w", extractErr))
		}

		content = subtree.Content
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
		content, err = os.ReadFile(filePath)
		if err != nil {
			return ctx.HandleError(err)
		}
	}

	if len(content) == 0 {
		// Empty file case
		response := PeekResponse{
			Selector: selector,
			FileInfo: PeekFileInfo{
				FilePath:   filePath,
				FileExists: true,
			},
			TableOfContents: &PeekTOC{
				IsFullFile:   isFullFile,
				RootSelector: subtreePath,
				Headings:     []PeekTOCHeading{},
			},
			Metadata: cmdutil.CreateJSONMetadata(ctx.Cmd, true, ctx.StartTime),
		}
		return cmdutil.OutputJSON(response)
	}

	// Parse document and extract headings
	doc := markdown.ParseDocument(content)
	headings := extractHeadingsFromContent(doc, content)

	if len(headings) == 0 {
		// No headings case
		response := PeekResponse{
			Selector: selector,
			FileInfo: PeekFileInfo{
				FilePath:   filePath,
				FileExists: true,
			},
			TableOfContents: &PeekTOC{
				IsFullFile:   isFullFile,
				RootSelector: subtreePath,
				Headings:     []PeekTOCHeading{},
			},
			Metadata: cmdutil.CreateJSONMetadata(ctx.Cmd, true, ctx.StartTime),
		}
		return cmdutil.OutputJSON(response)
	}

	// Build TOC headings
	tocHeadings := []PeekTOCHeading{}
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

		tocHeadings = append(tocHeadings, PeekTOCHeading{
			Text:     heading.Text,
			Level:    heading.Level,
			Selector: selectorText,
		})
	}

	response := PeekResponse{
		Selector: selector,
		FileInfo: PeekFileInfo{
			FilePath:   filePath,
			FileExists: true,
		},
		TableOfContents: &PeekTOC{
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
		if h.Text == target.Text && h.Level == target.Level && h.Line == target.Line {
			targetIndex = i
			break
		}
	}

	if targetIndex == -1 {
		return []string{strings.ToLower(target.Text)}
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
		lineNumStr = selector[colonIndex+1:hashIndex]
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
	content, err := os.ReadFile(filePath)
	if err != nil {
		// Can't read file, return original selector
		return selector, nil
	}

	// Find heading for this line
	headingMap, err := markdown.FindNearestHeadingsForLines(content, []int{lineNum})
	if err != nil {
		// Can't parse markdown, return original selector  
		return selector, nil
	}

	if headingPath, found := headingMap[lineNum]; found && headingPath != "" {
		// Found heading, create enhanced selector
		return fmt.Sprintf("%s#%s", filename, headingPath), nil
	}

	// No heading found, return whole file selector
	return filename, nil
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

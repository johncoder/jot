package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/johncoder/jot/internal/markdown"
	"github.com/johncoder/jot/internal/workspace"
	"github.com/spf13/cobra"
	"github.com/yuin/goldmark/ast"
)

var peekCmd = &cobra.Command{
	Use:   "peek SELECTOR",
	Short: "View a specific markdown subtree without opening the entire file",
	Long: `View a specific markdown subtree (heading with all nested content) without opening the entire file.

The peek command uses the same path-based selector syntax as refile:
- Each segment uses case-insensitive contains matching
- Must match exactly one subtree
- Leading slashes handle unusual document structures

Examples:
  jot peek "inbox.md#meeting"                    # View meeting notes subtree
  jot peek "work.md#projects/frontend"          # View frontend project section
  jot peek "notes.md#research/database"         # View database research
  jot peek "inbox.md#/foo/bar"                  # Skip level 1, find foo/bar
  jot peek "inbox.md" --toc                     # Show table of contents for entire file
  jot peek "work.md#projects" --toc             # Show TOC for projects subtree

This is useful for quickly reviewing specific sections without opening full files in an editor.`,

	Args: cobra.RangeArgs(0, 1), // Allow 0 or 1 arguments for --toc mode
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := workspace.RequireWorkspace()
		if err != nil {
			return err
		}

		// Get flags
		raw, _ := cmd.Flags().GetBool("raw")
		info, _ := cmd.Flags().GetBool("info")
		toc, _ := cmd.Flags().GetBool("toc")

		// Handle TOC mode
		if toc {
			if len(args) == 0 {
				return fmt.Errorf("table of contents requires a file or selector (e.g., 'inbox.md' or 'work.md#projects')")
			}
			return showTableOfContents(ws, args[0])
		}

		// Regular peek mode requires exactly one argument
		if len(args) != 1 {
			return fmt.Errorf("peek requires a selector argument (e.g., 'inbox.md#meeting')")
		}

		// Parse the source path selector
		sourcePath, err := markdown.ParsePath(args[0])
		if err != nil {
			return fmt.Errorf("invalid selector: %w", err)
		}

		// Extract the subtree
		subtree, err := ExtractSubtree(ws, sourcePath)
		if err != nil {
			return fmt.Errorf("failed to extract subtree: %w", err)
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
			// Formatted mode: show with nice header
			fmt.Printf("# Subtree: %s#%s\n\n", sourcePath.File, 
				strings.Join(sourcePath.Segments, "/"))
			
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

// printSubtreeInfo displays metadata about the subtree
func printSubtreeInfo(subtree *markdown.Subtree, filename string) {
	fmt.Printf("Subtree Information:\n")
	fmt.Printf("  File: %s\n", filename)
	fmt.Printf("  Heading: %q\n", subtree.Heading)
	fmt.Printf("  Level: %d\n", subtree.Level)
	fmt.Printf("  Content length: %d bytes\n", len(subtree.Content))
	fmt.Printf("  Byte range: %d-%d\n", subtree.StartOffset, subtree.EndOffset)
	
	// Count nested headings
	nestedCount := countNestedHeadings(subtree.Content, subtree.Level)
	if nestedCount > 0 {
		fmt.Printf("  Nested headings: %d\n", nestedCount)
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
func showTableOfContents(ws *workspace.Workspace, selector string) error {
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
		
		subtree, extractErr := ExtractSubtree(ws, sourcePath)
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
		} else {
			// Handle files in lib/ directory
			if !strings.HasSuffix(selector, ".md") {
				selector += ".md"
				baseFilename = selector
				filename = selector
			}
			filePath = filepath.Join(ws.LibDir, selector)
		}
		
		// Check if file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", selector)
		}
		
		// Read file content
		content, err = os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", selector, err)
		}
	}
	
	if len(content) == 0 {
		fmt.Printf("File %s is empty (no table of contents available)\n", filename)
		return nil
	}
	
	// Parse document and extract headings
	doc := markdown.ParseDocument(content)
	headings := extractHeadingsFromContent(doc, content)
	
	if len(headings) == 0 {
		fmt.Printf("No headings found in %s\n", filename)
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
			selectorHint := generateOptimalSelector(baseFilename, heading, headings)
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
			fmt.Printf("Warning: Headings marked with ⚠️ may have ambiguous selectors.\n")
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
	text     string                // The heading text
	level    int                   // The heading level (1-6)
	children map[string]*TrieNode  // Child nodes (key is normalized text for matching)
	heading  *ast.Heading          // Reference to the actual AST node
	offset   int                   // Byte offset in the document
	isLeaf   bool                  // Whether this is a selectable leaf (has no children)
	fullPath []string              // Full path from root to this node
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

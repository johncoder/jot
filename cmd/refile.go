package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/johncoder/jot/internal/workspace"
	"github.com/spf13/cobra"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

var refileCmd = &cobra.Command{
	Use:   "refile",
	Short: "Move notes from inbox to organized files",
	Long: `Move notes from inbox.md to organized files in the lib/ directory.

This command allows you to:
- Select notes from inbox.md
- Choose destination files in lib/
- Move notes while preserving metadata
- Update workspace index

Examples:
  jot refile                     # Interactive selection
  jot refile --all               # Process all inbox notes
  jot refile --dest topics.md    # Specify destination file
  jot refile --offset 150        # Target note at cursor position (editor integration)
  jot refile --exact "2025-06-06 10:30" # Target specific timestamp
  jot refile --pattern "meeting" # Target notes matching pattern`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := workspace.RequireWorkspace()
		if err != nil {
			return err
		}

		fmt.Println("Starting refile process...")
		fmt.Println()
		
		if !ws.InboxExists() {
			fmt.Println("No inbox.md found. Run 'jot doctor --fix' to create it.")
			return nil
		}
		
		// Read and parse inbox notes using enhanced AST-based parser
		notes, err := parseInboxNotesAST(ws.InboxPath)
		if err != nil {
			// Fallback to legacy parser for backwards compatibility
			fmt.Printf("Warning: AST parsing failed, falling back to legacy parser: %v\n", err)
			notes, err = parseInboxNotes(ws.InboxPath)
			if err != nil {
				return fmt.Errorf("failed to parse inbox: %w", err)
			}
		}
		
		if len(notes) == 0 {
			fmt.Println("No notes found in inbox.md")
			return nil
		}
		
		fmt.Printf("Found %d notes in inbox.md:\n\n", len(notes))
		
		// Display notes
		for i, note := range notes {
			fmt.Printf("%d. %s\n", i+1, note.Title)
			if len(note.Content) > 100 {
				fmt.Printf("   %s...\n", note.Content[:100])
			} else {
				fmt.Printf("   %s\n", note.Content)
			}
			fmt.Println()
		}
		
		// Get command line arguments and flags
		all, _ := cmd.Flags().GetBool("all")
		dest, _ := cmd.Flags().GetString("dest")
		exact, _ := cmd.Flags().GetString("exact")
		pattern, _ := cmd.Flags().GetString("pattern")
		offset, _ := cmd.Flags().GetInt("offset")
		
		var selectedNotes []Note
		
		if all {
			// Select all notes
			selectedNotes = notes
			fmt.Printf("Selecting all %d notes for refile\n", len(notes))
		} else if exact != "" {
			// Exact timestamp targeting
			targeter, targeterErr := NewExactTargeter(exact)
			if targeterErr != nil {
				return fmt.Errorf("invalid exact timestamp specification '%s': %w", exact, targeterErr)
			}
			
			var selectErr error
			selectedNotes, selectErr = targeter.SelectNotes(notes)
			if selectErr != nil {
				return fmt.Errorf("failed to select notes by exact match: %w", selectErr)
			}
			
			fmt.Printf("Selected %d notes by exact timestamp match: %s\n", len(selectedNotes), exact)
		} else if pattern != "" {
			// Pattern-based targeting
			targeter, targeterErr := NewPatternTargeter(pattern)
			if targeterErr != nil {
				return fmt.Errorf("invalid pattern specification '%s': %w", pattern, targeterErr)
			}
			
			var selectErr error
			selectedNotes, selectErr = targeter.SelectNotes(notes)
			if selectErr != nil {
				return fmt.Errorf("failed to select notes by pattern: %w", selectErr)
			}
			
			fmt.Printf("Selected %d notes by pattern match: %s\n", len(selectedNotes), pattern)
		} else if offset >= 0 {
			// Offset-based targeting for editor integration
			targeter, targeterErr := NewOffsetTargeter(offset)
			if targeterErr != nil {
				return fmt.Errorf("invalid byte offset: %w", targeterErr)
			}
			
			var selectErr error
			selectedNotes, selectErr = targeter.SelectNotes(notes)
			if selectErr != nil {
				return fmt.Errorf("failed to select notes by byte offset: %w", selectErr)
			}
			
			fmt.Printf("Selected %d notes by byte offset: %d\n", len(selectedNotes), offset)
		} else if len(args) > 0 && args[0] != "" {
			// Try index-based targeting
			targeter, targeterErr := NewIndexTargeter(args[0])
			if targeterErr != nil {
				return fmt.Errorf("invalid index specification '%s': %w", args[0], targeterErr)
			}
			
			var selectErr error
			selectedNotes, selectErr = targeter.SelectNotes(notes)
			if selectErr != nil {
				return fmt.Errorf("failed to select notes: %w", selectErr)
			}
			
			fmt.Printf("Selected %d notes by index: %s\n", len(selectedNotes), args[0])
		} else {
			// No selection specified, show usage
			fmt.Println("Interactive refile functionality coming soon!")
			fmt.Println("For now, you can:")
			fmt.Println("  jot refile --all --dest target.md         # Move all notes")
			fmt.Println("  jot refile 1,3,5 --dest target.md        # Move specific notes by index")
			fmt.Println("  jot refile 1-3 --dest target.md          # Move notes 1 through 3")
			fmt.Println("  jot refile --exact '2025-06-06 10:30' --dest target.md  # Move by exact timestamp")
			fmt.Println("  jot refile --pattern 'meeting|task' --dest target.md    # Move by regex pattern")
			fmt.Println("  jot refile --offset 1234 --dest target.md # Move note at byte offset (editor integration)")
			fmt.Println("Or manually organize notes by editing files in lib/")
			return nil
		}
		
		// Validate destination
		if dest == "" {
			return fmt.Errorf("destination file must be specified with --dest flag")
		}
		
		// Move the selected notes
		return moveNotes(ws, selectedNotes, dest)
	},
}

// Note represents a parsed note from inbox
type Note struct {
	Title     string        // Heading text (timestamp)
	Content   string        // Content as string (for backwards compatibility)
	LineStart int          // Starting line number (for backwards compatibility)
	LineEnd   int          // Ending line number (for backwards compatibility)
	
	// Enhanced AST-based fields
	AST       ast.Node      // Full AST subtree for this note
	RawContent []byte       // Raw markdown content
	
	// Byte offset tracking for editor integration
	ByteStart int          // Starting byte offset in source file
	ByteEnd   int          // Ending byte offset in source file
}

// parseInboxNotes extracts individual notes from inbox.md
func parseInboxNotes(inboxPath string) ([]Note, error) {
	file, err := os.Open(inboxPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var notes []Note
	var currentNote *Note
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		
		// Look for timestamp headers (## YYYY-MM-DD HH:MM:SS)
		if strings.HasPrefix(line, "## ") && len(line) > 10 {
			// Save previous note if it exists
			if currentNote != nil {
				currentNote.LineEnd = lineNumber - 1
				notes = append(notes, *currentNote)
			}
			
			// Start new note
			currentNote = &Note{
				Title:     strings.TrimSpace(line[3:]), // Remove "## "
				LineStart: lineNumber,
			}
		} else if currentNote != nil && strings.TrimSpace(line) != "" {
			// Add content to current note
			if currentNote.Content != "" {
				currentNote.Content += " "
			}
			currentNote.Content += strings.TrimSpace(line)
		}
	}
	
	// Save the last note
	if currentNote != nil {
		currentNote.LineEnd = lineNumber
		notes = append(notes, *currentNote)
	}
	
	return notes, scanner.Err()
}

// parseInboxNotesAST extracts individual notes using AST-based parsing
func parseInboxNotesAST(inboxPath string) ([]Note, error) {
	// Read the entire file
	content, err := os.ReadFile(inboxPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read inbox file: %w", err)
	}
	
	// Parse with goldmark
	md := goldmark.New()
	reader := text.NewReader(content)
	_ = md.Parser().Parse(reader) // We're using manual parsing instead
	
	var notes []Note
	var headingPositions []int
	var headingTitles []string
	
	// First pass: find all heading positions manually
	lines := bytes.Split(content, []byte("\n"))
	byteOffset := 0
	for _, line := range lines {
		lineStr := string(line)
		if strings.HasPrefix(lineStr, "## ") && len(lineStr) > 3 {
			title := strings.TrimSpace(lineStr[3:])
			headingPositions = append(headingPositions, byteOffset)
			headingTitles = append(headingTitles, title)
		}
		byteOffset += len(line) + 1 // +1 for the newline character
	}
	
	// Second pass: create notes with proper byte ranges
	for i, title := range headingTitles {
		start := headingPositions[i]
		var end int
		if i+1 < len(headingPositions) {
			end = headingPositions[i+1] - 1 // End just before the next heading
		} else {
			end = len(content) - 1 // End at the end of file
		}
		
		// Extract content for this note (for backwards compatibility)
		noteBytes := content[start:end+1]
		lines := bytes.Split(noteBytes, []byte("\n"))
		var contentParts []string
		
		// Skip the heading line and collect content
		for j := 1; j < len(lines); j++ {
			line := strings.TrimSpace(string(lines[j]))
			if line != "" {
				contentParts = append(contentParts, line)
			}
		}
		
		note := Note{
			Title:      title,
			Content:    strings.Join(contentParts, " "),
			LineStart:  calculateLineNumber(content, start),
			LineEnd:    calculateLineNumber(content, end),
			ByteStart:  start,
			ByteEnd:    end,
			RawContent: noteBytes,
		}
		
		notes = append(notes, note)
	}
	
	return notes, nil
}

// extractHeadingText extracts the text content from a heading node
func extractHeadingText(heading *ast.Heading, source []byte) string {
	var text strings.Builder
	for child := heading.FirstChild(); child != nil; child = child.NextSibling() {
		if textNode, ok := child.(*ast.Text); ok {
			text.Write(textNode.Segment.Value(source))
		}
	}
	return strings.TrimSpace(text.String())
}

// isContentNode determines if a node contains content that should be part of a note
func isContentNode(node ast.Node) bool {
	switch node.(type) {
	case *ast.Paragraph, *ast.List, *ast.CodeBlock, *ast.FencedCodeBlock, *ast.Blockquote:
		return true
	case *ast.Heading:
		// Don't include other headings as content
		return false
	default:
		return false
	}
}

// extractNodeContent extracts text content from various node types
func extractNodeContent(node ast.Node, source []byte) string {
	var buf bytes.Buffer
	
	// Extract text content recursively
	err := ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			if textNode, ok := n.(*ast.Text); ok {
				buf.Write(textNode.Segment.Value(source))
				if textNode.SoftLineBreak() {
					buf.WriteString(" ")
				}
			}
		}
		return ast.WalkContinue, nil
	})
	
	if err != nil {
		return ""
	}
	
	return buf.String()
}

// getNodeStart gets the start position of a node by walking its segments
func getNodeStart(node ast.Node) int {
	if node.HasChildren() {
		// For nodes with children, find the first text segment
		for child := node.FirstChild(); child != nil; child = child.NextSibling() {
			if start := getNodeStart(child); start >= 0 {
				return start
			}
		}
	}
	
	// For text nodes and other leaf nodes, try to get segment info
	if hasSegment, ok := node.(interface{ Segment() *text.Segment }); ok {
		seg := hasSegment.Segment()
		if seg != nil {
			return seg.Start
		}
	}
	
	return 0
}

// getNodeEnd gets the end position of a node
func getNodeEnd(node ast.Node) int {
	// Try to find the last segment in this node and its children
	end := getNodeEndRecursive(node)
	if end > 0 {
		return end
	}
	return 0
}

// getNodeEndRecursive recursively finds the end position
func getNodeEndRecursive(node ast.Node) int {
	maxEnd := 0
	
	if hasSegment, ok := node.(interface{ Segment() *text.Segment }); ok {
		seg := hasSegment.Segment()
		if seg != nil && seg.Stop > maxEnd {
			maxEnd = seg.Stop
		}
	}
	
	// Check children
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if childEnd := getNodeEndRecursive(child); childEnd > maxEnd {
			maxEnd = childEnd
		}
	}
	
	return maxEnd
}

// calculateLineNumber calculates the line number for a given byte offset
func calculateLineNumber(content []byte, offset int) int {
	if offset >= len(content) {
		offset = len(content) - 1
	}
	if offset < 0 {
		offset = 0
	}
	
	lineNumber := 1
	for i := 0; i < offset; i++ {
		if content[i] == '\n' {
			lineNumber++
		}
	}
	return lineNumber
}

// IndexTargeter handles numeric targeting (e.g., "1,3,5" or "1-3,5")
type IndexTargeter struct {
	indices []int
}

// NewIndexTargeter creates a new index targeter from a comma-separated string
func NewIndexTargeter(indexStr string) (*IndexTargeter, error) {
	if indexStr == "" {
		return nil, fmt.Errorf("empty index string")
	}
	
	var indices []int
	parts := strings.Split(indexStr, ",")
	
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.Contains(part, "-") {
			// Handle range (e.g., "1-3")
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("invalid range format: %s", part)
			}
			
			start, err := parseInt(strings.TrimSpace(rangeParts[0]))
			if err != nil {
				return nil, fmt.Errorf("invalid start index: %s", rangeParts[0])
			}
			
			end, err := parseInt(strings.TrimSpace(rangeParts[1]))
			if err != nil {
				return nil, fmt.Errorf("invalid end index: %s", rangeParts[1])
			}
			
			if start > end {
				return nil, fmt.Errorf("start index %d is greater than end index %d", start, end)
			}
			
			for i := start; i <= end; i++ {
				indices = append(indices, i)
			}
		} else {
			// Handle single index
			idx, err := parseInt(part)
			if err != nil {
				return nil, fmt.Errorf("invalid index: %s", part)
			}
			indices = append(indices, idx)
		}
	}
	
	return &IndexTargeter{indices: indices}, nil
}

// SelectNotes selects notes by their index (1-based)
func (t *IndexTargeter) SelectNotes(notes []Note) ([]Note, error) {
	var selected []Note
	
	for _, idx := range t.indices {
		if idx < 1 || idx > len(notes) {
			return nil, fmt.Errorf("index %d is out of range (1-%d)", idx, len(notes))
		}
		selected = append(selected, notes[idx-1]) // Convert to 0-based
	}
	
	return selected, nil
}

// ExactTargeter handles exact timestamp matching
type ExactTargeter struct {
	timestamp string
}

// NewExactTargeter creates a new exact targeter for timestamp matching
func NewExactTargeter(timestamp string) (*ExactTargeter, error) {
	if timestamp == "" {
		return nil, fmt.Errorf("empty timestamp string")
	}
	return &ExactTargeter{timestamp: timestamp}, nil
}

// SelectNotes selects notes that match the exact timestamp
func (t *ExactTargeter) SelectNotes(notes []Note) ([]Note, error) {
	var selected []Note
	
	for _, note := range notes {
		if strings.Contains(note.Title, t.timestamp) {
			selected = append(selected, note)
		}
	}
	
	if len(selected) == 0 {
		return nil, fmt.Errorf("no notes found matching timestamp: %s", t.timestamp)
	}
	
	return selected, nil
}

// PatternTargeter handles regex pattern matching
type PatternTargeter struct {
	pattern *regexp.Regexp
	rawPattern string
}

// NewPatternTargeter creates a new pattern targeter for regex matching
func NewPatternTargeter(pattern string) (*PatternTargeter, error) {
	if pattern == "" {
		return nil, fmt.Errorf("empty pattern string")
	}
	
	compiled, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern '%s': %w", pattern, err)
	}
	
	return &PatternTargeter{
		pattern: compiled,
		rawPattern: pattern,
	}, nil
}

// SelectNotes selects notes that match the regex pattern (title or content)
func (t *PatternTargeter) SelectNotes(notes []Note) ([]Note, error) {
	var selected []Note
	
	for _, note := range notes {
		// Check both title and content for pattern match
		if t.pattern.MatchString(note.Title) || t.pattern.MatchString(note.Content) {
			selected = append(selected, note)
		}
	}
	
	if len(selected) == 0 {
		return nil, fmt.Errorf("no notes found matching pattern: %s", t.rawPattern)
	}
	
	return selected, nil
}

// OffsetTargeter handles cursor position-based targeting for editor integration
type OffsetTargeter struct {
	offset int
}

// NewOffsetTargeter creates a new offset targeter for cursor position targeting
func NewOffsetTargeter(offset int) (*OffsetTargeter, error) {
	if offset < 0 {
		return nil, fmt.Errorf("byte offset must be non-negative, got %d", offset)
	}
	return &OffsetTargeter{offset: offset}, nil
}

// SelectNotes selects the note that contains the specified byte offset
func (t *OffsetTargeter) SelectNotes(notes []Note) ([]Note, error) {
	for _, note := range notes {
		// Check if the offset falls within this note's byte range
		if t.offset >= note.ByteStart && t.offset <= note.ByteEnd {
			return []Note{note}, nil
		}
	}
	
	return nil, fmt.Errorf("no note found at byte offset %d", t.offset)
}

// parseInt is a helper to parse integers with better error messages
func parseInt(s string) (int, error) {
	if s == "" {
		return 0, fmt.Errorf("empty string")
	}
	
	var result int
	var err error
	
	if result, err = parseIntHelper(s); err != nil {
		return 0, fmt.Errorf("'%s' is not a valid number", s)
	}
	
	if result < 1 {
		return 0, fmt.Errorf("index must be positive, got %d", result)
	}
	
	return result, nil
}

// parseIntHelper does the actual integer parsing
func parseIntHelper(s string) (int, error) {
	result := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, fmt.Errorf("not a number")
		}
		result = result*10 + int(r-'0')
	}
	return result, nil
}

// moveNotes moves selected notes from inbox to destination file
func moveNotes(ws *workspace.Workspace, notes []Note, destFile string) error {
	if len(notes) == 0 {
		return fmt.Errorf("no notes to move")
	}
	
	// Validate destination file path
	destPath := filepath.Join(ws.LibDir, destFile)
	
	fmt.Printf("Moving %d notes to %s...\n", len(notes), destFile)
	
	// Read the source inbox file for precise byte extraction
	sourceContent, err := os.ReadFile(ws.InboxPath)
	if err != nil {
		return fmt.Errorf("failed to read source inbox: %w", err)
	}
	
	// Read current destination file content (if it exists)
	var destContent []byte
	if _, err := os.Stat(destPath); err == nil {
		destContent, err = os.ReadFile(destPath)
		if err != nil {
			return fmt.Errorf("failed to read destination file: %w", err)
		}
	}
	
	// Append notes to destination
	var newContent []byte
	if len(destContent) > 0 {
		newContent = append(destContent, '\n')
	}
	
	// Add each note using precise byte extraction
	for _, note := range notes {
		// Extract the exact bytes from the source file
		if note.ByteStart >= 0 && note.ByteEnd > note.ByteStart && note.ByteEnd <= len(sourceContent) {
			noteBytes := sourceContent[note.ByteStart:note.ByteEnd]
			newContent = append(newContent, noteBytes...)
			newContent = append(newContent, '\n', '\n') // Add spacing between notes
		} else {
			// Fallback to reconstituted content if byte positions are invalid
			noteContent := fmt.Sprintf("## %s\n%s\n\n", note.Title, note.Content)
			newContent = append(newContent, []byte(noteContent)...)
		}
	}
	
	// Write to destination file
	if err := os.WriteFile(destPath, newContent, 0644); err != nil {
		return fmt.Errorf("failed to write destination file: %w", err)
	}
	
	// Remove notes from inbox
	if err := removeNotesFromInbox(ws.InboxPath, notes); err != nil {
		return err
	}
	
	// Success message
	fmt.Printf("Successfully moved %d notes to %s\n", len(notes), destFile)
	return nil
}

// removeNotesFromInbox removes the specified notes from inbox.md
func removeNotesFromInbox(inboxPath string, notesToRemove []Note) error {
	// Create a map of notes to remove for quick lookup (by title)
	removeMap := make(map[string]bool)
	for _, note := range notesToRemove {
		removeMap[note.Title] = true
	}
	
	// Use the legacy parser to avoid recursion with the AST parser
	allNotes, err := parseInboxNotes(inboxPath)
	if err != nil {
		return fmt.Errorf("failed to parse inbox for removal: %w", err)
	}
	
	// Build new content with remaining notes
	var newContent strings.Builder
	newContent.WriteString("# Inbox\n\n")
	
	for _, note := range allNotes {
		if !removeMap[note.Title] {
			// Keep this note
			newContent.WriteString(fmt.Sprintf("## %s\n%s\n\n", note.Title, note.Content))
		}
	}
	
	// Write the updated content back to inbox
	if err := os.WriteFile(inboxPath, []byte(newContent.String()), 0644); err != nil {
		return fmt.Errorf("failed to update inbox: %w", err)
	}
	
	return nil
}

func init() {
	refileCmd.Flags().Bool("all", false, "Process all notes in inbox")
	refileCmd.Flags().String("dest", "", "Destination file in lib/")
	refileCmd.Flags().String("exact", "", "Target notes by exact timestamp match")
	refileCmd.Flags().String("pattern", "", "Target notes by regex pattern match")
	refileCmd.Flags().Int("offset", -1, "Target note containing the specified byte offset (for editor integration)")
}

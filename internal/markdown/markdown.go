// Package markdown provides utilities for parsing and querying markdown documents
package markdown

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// HeadingPath represents a parsed path selector for navigating markdown headings
type HeadingPath struct {
	File       string   // "inbox.md" - source file name
	Segments   []string // ["meeting", "attendees"] - path segments for navigation
	SkipLevels int      // Number of leading slashes (for unusual document structures)
}

// Subtree represents a complete markdown subtree (heading + all nested content)
type Subtree struct {
	Heading     string // Original heading text
	Level       int    // Original heading level (1-6)
	Content     []byte // Full subtree content (markdown)
	StartOffset int    // Byte position in source
	EndOffset   int    // Byte position in source
}

// ParsePath parses a path selector like "file.md#path/to/heading"
func ParsePath(pathStr string) (*HeadingPath, error) {
	parts := strings.SplitN(pathStr, "#", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("path must contain '#' separator (e.g., 'file.md#heading')")
	}

	file := strings.TrimSpace(parts[0])
	pathPart := strings.TrimSpace(parts[1])

	if file == "" {
		return nil, fmt.Errorf("file name cannot be empty")
	}

	// Count leading slashes for skip levels
	skipLevels := 0
	for len(pathPart) > 0 && pathPart[0] == '/' {
		skipLevels++
		pathPart = pathPart[1:]
	}

	// Parse path segments
	var segments []string
	if pathPart != "" {
		segments = strings.Split(pathPart, "/")
		// Clean up segments
		for i, seg := range segments {
			segments[i] = strings.TrimSpace(seg)
		}
	}

	return &HeadingPath{
		File:       file,
		Segments:   segments,
		SkipLevels: skipLevels,
	}, nil
}

// ParseDocument parses markdown content and returns the AST document
func ParseDocument(content []byte) ast.Node {
	md := goldmark.New()
	reader := text.NewReader(content)
	return md.Parser().Parse(reader)
}

// FindSubtree finds a subtree matching the given path selector
func FindSubtree(doc ast.Node, content []byte, path *HeadingPath) (*Subtree, error) {
	var matches []*Subtree

	// Walk the AST to find matching headings
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if heading, ok := n.(*ast.Heading); ok {
			// Check if this heading starts a valid path match
			if subtree := tryMatchPath(heading, content, path, 0); subtree != nil {
				matches = append(matches, subtree)
			}
		}

		return ast.WalkContinue, nil
	})

	if len(matches) == 0 {
		return nil, fmt.Errorf("no headings found matching path \"%s\" in %s",
			strings.Join(path.Segments, "/"), path.File)
	}

	if len(matches) > 1 {
		var matchDetails []string
		for _, match := range matches {
			line := CalculateLineNumber(content, match.StartOffset)
			matchDetails = append(matchDetails, fmt.Sprintf("  - \"%s\" at line %d", match.Heading, line))
		}
		return nil, fmt.Errorf("multiple headings match \"%s\" in %s:\n%s\nUse a more specific path",
			strings.Join(path.Segments, "/"), path.File, strings.Join(matchDetails, "\n"))
	}

	return matches[0], nil
}

// FindAllHeadings returns all headings in the document with their paths
func FindAllHeadings(doc ast.Node, content []byte) []HeadingInfo {
	var headings []HeadingInfo
	var currentPath []string
	var levelStack []int

	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if heading, ok := n.(*ast.Heading); ok {
			headingText := ExtractHeadingText(heading, content)

			// Adjust path stack based on heading level
			for len(levelStack) > 0 && levelStack[len(levelStack)-1] >= heading.Level {
				levelStack = levelStack[:len(levelStack)-1]
				if len(currentPath) > 0 {
					currentPath = currentPath[:len(currentPath)-1]
				}
			}

			// Add current heading to path
			levelStack = append(levelStack, heading.Level)
			currentPath = append(currentPath, headingText)

			// Create heading info
			pathCopy := make([]string, len(currentPath))
			copy(pathCopy, currentPath)

			headings = append(headings, HeadingInfo{
				Text:   headingText,
				Level:  heading.Level,
				Path:   pathCopy,
				Offset: GetNodeOffset(heading, content),
			})
		}

		return ast.WalkContinue, nil
	})

	return headings
}

// HeadingInfo represents information about a heading in the document
type HeadingInfo struct {
	Text   string   // Heading text
	Level  int      // Heading level (1-6)
	Path   []string // Full path to this heading
	Offset int      // Byte offset in document
}

// tryMatchPath attempts to match a path starting from a given heading
func tryMatchPath(heading *ast.Heading, content []byte, path *HeadingPath, segmentIndex int) *Subtree {
	// Get heading text for matching
	headingText := ExtractHeadingText(heading, content)

	// Check if current segment matches (case-insensitive contains)
	if segmentIndex >= len(path.Segments) {
		return nil
	}

	segment := path.Segments[segmentIndex]
	if !strings.Contains(strings.ToLower(headingText), strings.ToLower(segment)) {
		return nil
	}

	// For single-segment paths, allow any level (contains matching)
	if len(path.Segments) == 1 {
		return extractSubtreeFromHeading(heading, content)
	}

	// For multi-segment paths, enforce hierarchical level structure
	expectedLevel := segmentIndex + 1 + path.SkipLevels
	if heading.Level != expectedLevel {
		return nil
	}

	// If this is the last segment, we found our target
	if segmentIndex == len(path.Segments)-1 {
		return extractSubtreeFromHeading(heading, content)
	}

	// Look for next level heading among siblings
	for sibling := heading.NextSibling(); sibling != nil; sibling = sibling.NextSibling() {
		if siblingHeading, ok := sibling.(*ast.Heading); ok {
			if siblingHeading.Level == expectedLevel+1 {
				if result := tryMatchPath(siblingHeading, content, path, segmentIndex+1); result != nil {
					return result
				}
			} else if siblingHeading.Level <= expectedLevel {
				// Hit a heading at same or higher level, stop looking
				break
			}
		}
	}

	return nil
}

// extractSubtreeFromHeading extracts a complete subtree starting from a heading
func extractSubtreeFromHeading(heading *ast.Heading, content []byte) *Subtree {
	headingText := ExtractHeadingText(heading, content)
	textOffset := GetNodeOffset(heading, content)

	// Find the actual start of the heading line (including ### markers)
	// Walk backwards from the text offset to find the beginning of the line
	startOffset := textOffset
	for startOffset > 0 && content[startOffset-1] != '\n' {
		startOffset--
	}

	// Find the end of this subtree
	endOffset := findSubtreeEnd(heading, content)

	// Extract content
	subtreeContent := content[startOffset:endOffset]

	return &Subtree{
		Heading:     headingText,
		Level:       heading.Level,
		Content:     subtreeContent,
		StartOffset: startOffset,
		EndOffset:   endOffset,
	}
}

// ExtractHeadingText extracts text from a heading node
func ExtractHeadingText(heading *ast.Heading, content []byte) string {
	var text strings.Builder
	ast.Walk(heading, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			if textNode, ok := n.(*ast.Text); ok {
				text.Write(textNode.Segment.Value(content))
			}
		}
		return ast.WalkContinue, nil
	})
	return strings.TrimSpace(text.String())
}

// GetNodeOffset gets the byte offset of a node in the content
func GetNodeOffset(node ast.Node, content []byte) int {
	// Try to get segment information
	if hasSegment, ok := node.(interface{ Segment() *text.Segment }); ok {
		seg := hasSegment.Segment()
		if seg != nil {
			return seg.Start
		}
	}

	// Fallback: try to get segment from Lines() method for headings
	if heading, ok := node.(*ast.Heading); ok {
		if heading.Lines().Len() > 0 {
			segment := heading.Lines().At(0)
			return segment.Start
		}
	}

	return 0
}

// FindSubtreeEnd finds where a subtree ends (before next same-level heading)
// This is now public to allow external testing and usage
func FindSubtreeEnd(heading *ast.Heading, content []byte) int {
	return findSubtreeEnd(heading, content)
}

// findSubtreeEnd finds where this subtree ends (before next same-level heading)
func findSubtreeEnd(heading *ast.Heading, content []byte) int {
	// Walk forward to find the next heading at same or higher level
	current := heading.NextSibling()
	for current != nil {
		if h, ok := current.(*ast.Heading); ok && h.Level <= heading.Level {
			// Found next same-level or higher heading
			nextHeadingOffset := GetNodeOffset(h, content)

			// Find the actual start of the heading line by looking backwards for newline
			lineStart := nextHeadingOffset
			for lineStart > 0 && content[lineStart-1] != '\n' {
				lineStart--
			}

			// If we found a newline, the line starts after it
			// Otherwise, we're at the beginning of the content
			if lineStart > 0 && content[lineStart-1] == '\n' {
				return lineStart
			}

			return lineStart
		}
		current = current.NextSibling()
	}

	// No next heading found, go to end of content
	return len(content)
}

// ValidateOffset ensures an offset is within valid bounds for the given content
func ValidateOffset(content []byte, offset int) int {
	if offset < 0 {
		return 0
	}
	if offset > len(content) {
		return len(content)
	}
	return offset
}

// CalculateLineNumber calculates line number from byte offset
func CalculateLineNumber(content []byte, offset int) int {
	offset = ValidateOffset(content, offset)

	lineNumber := 1
	for i := 0; i < offset; i++ {
		if content[i] == '\n' {
			lineNumber++
		}
	}
	return lineNumber
}

// CalculateLineColumn calculates both line and column number from byte offset
func CalculateLineColumn(content []byte, offset int) (line int, column int) {
	offset = ValidateOffset(content, offset)

	line = 1
	column = 1

	for i := 0; i < offset; i++ {
		if content[i] == '\n' {
			line++
			column = 1
		} else {
			column++
		}
	}

	return line, column
}

// FindInsertionPoint finds the best byte offset for inserting content at a specific location
func FindInsertionPoint(content []byte, targetOffset int, prepend bool) int {
	targetOffset = ValidateOffset(content, targetOffset)

	if prepend {
		// Find the end of the current line to insert after
		for i := targetOffset; i < len(content); i++ {
			if content[i] == '\n' {
				return i + 1 // Insert at beginning of next line
			}
		}
		// If no newline found, append at end
		return len(content)
	}

	// For append mode, find end of content section
	return len(content)
}

// OffsetRange represents a range of bytes in content
type OffsetRange struct {
	Start int
	End   int
}

// IsValid checks if the offset range is valid for the given content
func (r OffsetRange) IsValid(content []byte) bool {
	return r.Start >= 0 && r.End >= r.Start && r.End <= len(content)
}

// Length returns the length of the range
func (r OffsetRange) Length() int {
	if r.End < r.Start {
		return 0
	}
	return r.End - r.Start
}

// Extract returns the content within the range
func (r OffsetRange) Extract(content []byte) []byte {
	if !r.IsValid(content) {
		return nil
	}
	return content[r.Start:r.End]
}

// TransformHeadingLevels adjusts heading levels in markdown content
func TransformHeadingLevels(content []byte, levelDiff int) []byte {
	lines := bytes.Split(content, []byte("\n"))
	var result []byte

	for i, line := range lines {
		if i > 0 {
			result = append(result, '\n')
		}

		// Check if line is a heading
		if bytes.HasPrefix(line, []byte("#")) {
			// Count current level
			currentLevel := 0
			for j := 0; j < len(line) && line[j] == '#'; j++ {
				currentLevel++
			}

			if currentLevel > 0 && currentLevel < len(line) && line[currentLevel] == ' ' {
				// This is a valid heading, transform it
				newLevel := currentLevel + levelDiff
				if newLevel > 6 {
					newLevel = 6 // Markdown max heading level
				}
				if newLevel < 1 {
					newLevel = 1
				}

				// Build new heading
				newHeading := bytes.Repeat([]byte("#"), newLevel)
				newHeading = append(newHeading, line[currentLevel:]...)
				result = append(result, newHeading...)
			} else {
				result = append(result, line...)
			}
		} else {
			result = append(result, line...)
		}
	}

	return result
}

// CreateHeadingStructure creates missing heading hierarchy
func CreateHeadingStructure(headings []string, baseLevel int) []byte {
	var result []byte

	for i, heading := range headings {
		level := baseLevel + i
		if i > 0 {
			result = append(result, '\n')
		}

		// Create heading line
		levelMarker := bytes.Repeat([]byte("#"), level)
		result = append(result, levelMarker...)
		result = append(result, ' ')
		result = append(result, []byte(heading)...)
		result = append(result, '\n')
	}

	return result
}

// PathMatches checks if a path matches the given segments using contains matching
func PathMatches(actualPath []string, targetSegments []string, skipLevels int) bool {
	if len(actualPath) < len(targetSegments) {
		return false
	}

	// Adjust for skip levels
	if skipLevels >= len(actualPath) {
		return false
	}

	startIndex := skipLevels
	if len(actualPath)-startIndex < len(targetSegments) {
		return false
	}

	// Check each segment for contains match
	for i, segment := range targetSegments {
		actualIndex := startIndex + i
		if actualIndex >= len(actualPath) {
			return false
		}

		actual := strings.ToLower(actualPath[actualIndex])
		target := strings.ToLower(segment)

		if !strings.Contains(actual, target) {
			return false
		}
	}

	return true
}

// LineHeadingMap represents the mapping from line numbers to heading paths
type LineHeadingMap map[int]string

// FindNearestHeadingsForLines efficiently finds the nearest dominant heading for multiple line numbers
// by parsing the document once and tracking heading context as we go.
// It stops parsing once all target lines have been resolved.
func FindNearestHeadingsForLines(content []byte, targetLines []int) (LineHeadingMap, error) {
	if len(targetLines) == 0 {
		return make(LineHeadingMap), nil
	}

	// Sort target lines for efficient processing
	sortedLines := make([]int, len(targetLines))
	copy(sortedLines, targetLines)
	sort.Ints(sortedLines)

	// Create result map
	result := make(LineHeadingMap)

	// Track which lines we still need to resolve
	remainingLines := make(map[int]bool)
	for _, line := range targetLines {
		remainingLines[line] = true
	}

	// Parse the markdown document
	md := goldmark.New()
	doc := md.Parser().Parse(text.NewReader(content))

	// Track current heading context as we walk through the document
	var currentHeadingPath []string
	var levelStack []int
	currentLine := 1

	// Walk through the AST and track line numbers
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		// Calculate current line number from node offset
		nodeOffset := GetNodeOffset(n, content)
		newLine := CalculateLineNumber(content, nodeOffset)

		// Assign any target lines between the previous position and current position
		// to the current heading context (before we potentially update it)
		for line := currentLine; line < newLine && line <= sortedLines[len(sortedLines)-1]; line++ {
			if remainingLines[line] {
				if len(currentHeadingPath) > 0 {
					result[line] = strings.Join(currentHeadingPath, "/")
				} else {
					result[line] = "" // No heading context (top of file)
				}
				delete(remainingLines, line)
			}
		}

		currentLine = newLine

		// Handle headings - update our current context
		if heading, ok := n.(*ast.Heading); ok {
			headingText := ExtractHeadingText(heading, content)

			// Adjust path stack based on heading level
			for len(levelStack) > 0 && levelStack[len(levelStack)-1] >= heading.Level {
				levelStack = levelStack[:len(levelStack)-1]
				if len(currentHeadingPath) > 0 {
					currentHeadingPath = currentHeadingPath[:len(currentHeadingPath)-1]
				}
			}

			// Add current heading to path
			levelStack = append(levelStack, heading.Level)
			currentHeadingPath = append(currentHeadingPath, headingText)
		}

		// Assign any target lines at the current position to the current heading context
		if remainingLines[currentLine] {
			if len(currentHeadingPath) > 0 {
				result[currentLine] = strings.Join(currentHeadingPath, "/")
			} else {
				result[currentLine] = "" // No heading context (top of file)
			}
			delete(remainingLines, currentLine)
		}

		// Early exit if we've resolved all target lines
		if len(remainingLines) == 0 {
			return ast.WalkStop, nil
		}

		return ast.WalkContinue, nil
	})

	// Handle any remaining lines that weren't reached (assign last known heading context)
	lastHeadingPath := ""
	if len(currentHeadingPath) > 0 {
		lastHeadingPath = strings.Join(currentHeadingPath, "/")
	}

	for line := range remainingLines {
		result[line] = lastHeadingPath
	}

	return result, nil
}

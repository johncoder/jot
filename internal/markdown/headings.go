// headings.go: Shared heading/path navigation and selector utilities for jot

package markdown

import (
	"strings"

	"github.com/gomarkdown/markdown/ast"
)

// PathResolution represents the result of heading path navigation
// (moved from refile.go)
type PathResolution struct {
	TargetHeading   *ast.Heading
	ParentHeading   *ast.Heading
	FoundSegments   []string
	MissingSegments []string
}

// NavigateHeadingPath navigates through the heading hierarchy following the path
func NavigateHeadingPath(doc ast.Node, destPath *HeadingPath) (*PathResolution, error) {
	result := &PathResolution{
		FoundSegments:   []string{},
		MissingSegments: []string{},
	}

	if len(destPath.Segments) == 0 {
		return result, nil
	}

	allHeadings := FindAllHeadings(doc)

	var bestMatch *HeadingInfo
	var bestMatchDepth int

	for _, heading := range allHeadings {
		matchDepth := CalculatePathMatch(heading.Path, destPath.Segments, destPath.SkipLevels)
		if matchDepth == len(destPath.Segments) {
			result.TargetHeading = heading.Node
			result.FoundSegments = destPath.Segments
			return result, nil
		}
		if matchDepth > bestMatchDepth {
			bestMatchDepth = matchDepth
			bestMatch = &heading
		}
	}

	if bestMatch != nil && bestMatchDepth > 0 {
		result.ParentHeading = bestMatch.Node
		result.FoundSegments = destPath.Segments[:bestMatchDepth]
		result.MissingSegments = destPath.Segments[bestMatchDepth:]
	} else {
		result.MissingSegments = destPath.Segments
	}

	return result, nil
}

// CalculatePathMatch checks how many consecutive segments match using contains logic
func CalculatePathMatch(headingPath []string, targetSegments []string, skipLevels int) int {
	if len(headingPath) < skipLevels {
		return 0
	}
	adjustedPath := headingPath[skipLevels:]
	bestMatch := 0
	for startPos := 0; startPos <= len(adjustedPath)-1; startPos++ {
		matchCount := 0
		for i, targetSeg := range targetSegments {
			pathIndex := startPos + i
			if pathIndex >= len(adjustedPath) {
				break
			}
			headingSeg := adjustedPath[pathIndex]
			if strings.Contains(strings.ToLower(headingSeg), strings.ToLower(targetSeg)) {
				matchCount++
			} else {
				break
			}
		}
		if matchCount > bestMatch {
			bestMatch = matchCount
		}
		if matchCount == len(targetSegments) {
			break
		}
	}
	return bestMatch
}

// HeadingInfo holds heading node, text, level, and path
// Node is the *ast.Heading node
// Path is the full path of heading texts up to this heading
// Level is the heading level (1-6)
type HeadingInfo struct {
	Node  *ast.Heading
	Text  string
	Level int
	Path  []string
}

// FindAllHeadings returns all headings in a gomarkdown document, with their text, level, and path.
func FindAllHeadings(doc ast.Node) []HeadingInfo {
	var headings []HeadingInfo
	var currentPath []string
	var visit func(n ast.Node)
	visit = func(n ast.Node) {
		switch h := n.(type) {
		case *ast.Heading:
			txt := extractHeadingText(h)
			level := h.Level
			if level <= len(currentPath) {
				currentPath = currentPath[:level-1]
			}
			currentPath = append(currentPath, txt)
			headings = append(headings, HeadingInfo{
				Node:  h,
				Text:  txt,
				Level: int(level),
				Path:  append([]string{}, currentPath...),
			})
			// Visit children
			for _, child := range h.GetChildren() {
				visit(child)
			}
			// Pop path after leaving heading
			currentPath = currentPath[:len(currentPath)-1]
		default:
			for _, child := range n.GetChildren() {
				visit(child)
			}
		}
	}
	visit(doc)
	return headings
}

// extractHeadingText extracts the text from a heading node
func extractHeadingText(heading *ast.Heading) string {
	var sb strings.Builder
	for _, child := range heading.GetChildren() {
		if t, ok := child.(*ast.Text); ok {
			sb.WriteString(string(t.Literal))
		}
	}
	return sb.String()
}

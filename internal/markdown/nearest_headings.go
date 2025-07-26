package markdown

import (
	"strings"
)

// FindNearestHeadingsForLines scans the content for headings, notes their line numbers,
// and for each target line, finds the nearest heading above it.
// Returns a map[lineNumber]headingPath (e.g., "Section/Subsection").
func FindNearestHeadingsForLines(content []byte, targetLines []int) (map[int]string, error) {
	lines := strings.Split(string(content), "\n")
	headingLines := make([]struct {
		LineNum int
		Path    string
	}, 0)
	var currentPath []string

	// Scan for headings and record their line numbers and paths
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			// Count heading level
			level := 0
			for level < len(trimmed) && trimmed[level] == '#' {
				level++
			}
			headingText := strings.TrimSpace(trimmed[level:])
			// Adjust path stack for heading level
			if level <= len(currentPath) {
				currentPath = currentPath[:level-1]
			}
			currentPath = append(currentPath, headingText)
			headingLines = append(headingLines, struct {
				LineNum int
				Path    string
			}{LineNum: i + 1, Path: strings.Join(currentPath, "/")})
		}
	}

	// For each target line, find the nearest heading above it
	result := make(map[int]string)
	for _, target := range targetLines {
		nearest := ""
		for _, h := range headingLines {
			if h.LineNum <= target {
				nearest = h.Path
			} else {
				break
			}
		}
		result[target] = nearest
	}
	return result, nil
}

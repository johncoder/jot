package markdown

import "bytes"

// TransformHeadingLevels adjusts all heading levels in the given Markdown content by levelDiff.
// It clamps heading levels to the range 1-6.
func TransformHeadingLevels(content []byte, levelDiff int) []byte {
	// Naive implementation: adjust heading levels by manipulating lines starting with '#'
	lines := bytes.Split(content, []byte("\n"))
	for i, line := range lines {
		if len(line) == 0 || line[0] != '#' {
			continue
		}
		// Count leading '#' characters
		headingLevel := 0
		for headingLevel < len(line) && line[headingLevel] == '#' {
			headingLevel++
		}
		if headingLevel == 0 || headingLevel > 6 {
			continue
		}
		// Adjust level
		newLevel := headingLevel + levelDiff
		if newLevel < 1 {
			newLevel = 1
		} else if newLevel > 6 {
			newLevel = 6
		}
		// Replace with new heading
		lines[i] = append(bytes.Repeat([]byte{'#'}, newLevel), line[headingLevel:]...)
	}
	return bytes.Join(lines, []byte("\n"))
}

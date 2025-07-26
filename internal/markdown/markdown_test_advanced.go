package markdown

import (
	"strings"
	"testing"
)

// TestParsePathAdvanced tests advanced parsing scenarios for Markdown paths.
func TestParsePathAdvanced(t *testing.T) {
	tests := []struct {
		name         string
		pathStr      string
		expectedFile string
		expectedSegs []string
		expectedSkip int
		expectError  bool
	}{
		{
			name:         "basic path",
			pathStr:      "file.md#heading",
			expectedFile: "file.md",
			expectedSegs: []string{"heading"},
			expectedSkip: 0,
			expectError:  false,
		},
		{
			name:         "hierarchical path",
			pathStr:      "work.md#projects/frontend/tasks",
			expectedFile: "work.md",
			expectedSegs: []string{"projects", "frontend", "tasks"},
			expectedSkip: 0,
			expectError:  false,
		},
		{
			name:         "path with skip levels",
			pathStr:      "notes.md#/section/subsection",
			expectedFile: "notes.md",
			expectedSegs: []string{"section", "subsection"},
			expectedSkip: 1,
			expectError:  false,
		},
		{
			name:         "path with multiple skip levels",
			pathStr:      "doc.md#//deep/section",
			expectedFile: "doc.md",
			expectedSegs: []string{"deep", "section"},
			expectedSkip: 2,
			expectError:  false,
		},
		{
			name:         "empty segments after skip",
			pathStr:      "file.md#/",
			expectedFile: "file.md",
			expectedSegs: []string{""},
			expectedSkip: 1,
			expectError:  false,
		},
		{
			name:        "missing hash separator",
			pathStr:     "file.md",
			expectError: true,
		},
		{
			name:        "empty file name",
			pathStr:     "#heading",
			expectError: true,
		},
		{
			name:         "complex path with spaces",
			pathStr:      "my notes.md#meeting notes/action items",
			expectedFile: "my notes.md",
			expectedSegs: []string{"meeting notes", "action items"},
			expectedSkip: 0,
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParsePath(tt.pathStr)

			if tt.expectError {
				if err == nil {
					t.Errorf("ParsePath() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("ParsePath() unexpected error = %v", err)
			}

			if result.File != tt.expectedFile {
				t.Errorf("ParsePath() file = %q, want %q", result.File, tt.expectedFile)
			}

			if len(result.Segments) != len(tt.expectedSegs) {
				t.Errorf("ParsePath() segments length = %d, want %d", len(result.Segments), len(tt.expectedSegs))
			} else {
				for i, seg := range result.Segments {
					if seg != tt.expectedSegs[i] {
						t.Errorf("ParsePath() segment[%d] = %q, want %q", i, seg, tt.expectedSegs[i])
					}
				}
			}

			if result.SkipLevels != tt.expectedSkip {
				t.Errorf("ParsePath() skip levels = %d, want %d", result.SkipLevels, tt.expectedSkip)
			}
		})
	}
}

// PathMatches returns true if the targetSegments match the actualPath after skipping skipLevels.
func PathMatches(actualPath, targetSegments []string, skipLevels int) bool {
	if len(actualPath) < skipLevels+len(targetSegments) {
		return false
	}
	adjustedPath := actualPath[skipLevels:]
	for i, seg := range targetSegments {
		if i >= len(adjustedPath) {
			return false
		}
		if !strings.Contains(strings.ToLower(adjustedPath[i]), strings.ToLower(seg)) {
			return false
		}
	}
	return true
}

// TestPathMatches tests the PathMatches function for matching segments in a path.
func TestPathMatches(t *testing.T) {
	tests := []struct {
		name           string
		actualPath     []string
		targetSegments []string
		skipLevels     int
		expected       bool
	}{
		{
			name:           "exact match",
			actualPath:     []string{"Work", "Projects", "Frontend"},
			targetSegments: []string{"Projects", "Frontend"},
			skipLevels:     1,
			expected:       true,
		},
		{
			name:           "contains match",
			actualPath:     []string{"Work", "Projects", "Frontend Development"},
			targetSegments: []string{"projects", "frontend"},
			skipLevels:     1,
			expected:       true,
		},
		{
			name:           "partial match only",
			actualPath:     []string{"Work", "Projects", "Backend"},
			targetSegments: []string{"Projects", "Frontend"},
			skipLevels:     1,
			expected:       false,
		},
		{
			name:           "no match",
			actualPath:     []string{"Work", "Meetings", "Daily"},
			targetSegments: []string{"Projects", "Frontend"},
			skipLevels:     1,
			expected:       false,
		},
		{
			name:           "insufficient path length",
			actualPath:     []string{"Work"},
			targetSegments: []string{"Projects", "Frontend"},
			skipLevels:     1,
			expected:       false,
		},
		{
			name:           "skip levels too high",
			actualPath:     []string{"Work", "Projects"},
			targetSegments: []string{"Frontend"},
			skipLevels:     3,
			expected:       false,
		},
		{
			name:           "case insensitive matching",
			actualPath:     []string{"WORK", "Projects", "Frontend Development"},
			targetSegments: []string{"projects", "FRONTEND"},
			skipLevels:     1,
			expected:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PathMatches(tt.actualPath, tt.targetSegments, tt.skipLevels)
			if result != tt.expected {
				t.Errorf("PathMatches() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// CreateHeadingStructure generates Markdown heading structure from a slice of headings and a base level.
func CreateHeadingStructure(headings []string, baseLevel int) []byte {
	var sb strings.Builder
	for i, h := range headings {
		level := baseLevel + i
		if level > 6 {
			level = 6
		}
		sb.WriteString(strings.Repeat("#", level))
		sb.WriteString(" ")
		sb.WriteString(h)
		sb.WriteString("\n")
	}
	return []byte(sb.String())
}

// TestCreateHeadingStructure tests the CreateHeadingStructure function for generating heading Markdown.
func TestCreateHeadingStructure(t *testing.T) {
	tests := []struct {
		name      string
		headings  []string
		baseLevel int
		expected  string
	}{
		{
			name:      "single heading",
			headings:  []string{"Projects"},
			baseLevel: 1,
			expected:  "# Projects\n",
		},
		{
			name:      "multiple headings",
			headings:  []string{"Projects", "Frontend", "Components"},
			baseLevel: 2,
			expected:  "## Projects\n### Frontend\n#### Components\n",
		},
		{
			name:      "deep nesting",
			headings:  []string{"Level1", "Level2"},
			baseLevel: 4,
			expected:  "#### Level1\n##### Level2\n",
		},
		{
			name:      "empty headings",
			headings:  []string{},
			baseLevel: 1,
			expected:  "",
		},
		{
			name:      "headings with spaces",
			headings:  []string{"Meeting Notes", "Action Items"},
			baseLevel: 1,
			expected:  "# Meeting Notes\n## Action Items\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CreateHeadingStructure(tt.headings, tt.baseLevel)
			resultStr := string(result)
			if resultStr != tt.expected {
				t.Errorf("CreateHeadingStructure() = %q, want %q", resultStr, tt.expected)
			}
		})
	}
}

// TestTransformHeadingLevels tests the TransformHeadingLevels function for adjusting heading levels.
func TestTransformHeadingLevels(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		levelDiff int
		expected  string
	}{
		{
			name: "increase heading levels",
			content: `## Heading Level 2

Some content here.

### Heading Level 3

More content.

#### Heading Level 4

Deep content.
`,
			levelDiff: 1,
			expected: `### Heading Level 2

Some content here.

#### Heading Level 3

More content.

##### Heading Level 4

Deep content.
`,
		},
		{
			name: "decrease heading levels",
			content: `### Heading Level 3

Content.

#### Heading Level 4

More content.
`,
			levelDiff: -1,
			expected: `## Heading Level 3

Content.

### Heading Level 4

More content.
`,
		},
		{
			name: "no change",
			content: `## Heading

Content here.
`,
			levelDiff: 0,
			expected: `## Heading

Content here.
`,
		},
		{
			name: "mixed content with code blocks",
			content: `## Code Example

Here's some code:

` + "```go" + `
func main() {
    // This ## should not be changed
    fmt.Println("# Also not changed")
}
` + "```" + `

### Another Heading

Regular content.
`,
			levelDiff: 1,
			expected: `### Code Example

Here's some code:

` + "```go" + `
func main() {
    // This ## should not be changed
    fmt.Println("# Also not changed")
}
` + "```" + `

#### Another Heading

Regular content.
`,
		},
		{
			name:      "empty content",
			content:   "",
			levelDiff: 1,
			expected:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TransformHeadingLevels([]byte(tt.content), tt.levelDiff)
			resultStr := string(result)
			if resultStr != tt.expected {
				t.Errorf("TransformHeadingLevels() = %q, want %q", resultStr, tt.expected)
			}
		})
	}
}

package markdown

import (
	"strings"
	"testing"
)

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

func TestFindAllHeadings(t *testing.T) {
	content := `# Main Title

Some introduction text.

## Section 1

Content for section 1.

### Subsection 1.1

Nested content.

#### Deep Section

Very deep content.

### Subsection 1.2

More nested content.

## Section 2

Content for section 2.

### Subsection 2.1

Final nested content.
`

	doc := ParseDocument([]byte(content))
	headings := FindAllHeadings(doc, []byte(content))

	expectedHeadings := []struct {
		text  string
		level int
		path  []string
	}{
		{"Main Title", 1, []string{"Main Title"}},
		{"Section 1", 2, []string{"Main Title", "Section 1"}},
		{"Subsection 1.1", 3, []string{"Main Title", "Section 1", "Subsection 1.1"}},
		{"Deep Section", 4, []string{"Main Title", "Section 1", "Subsection 1.1", "Deep Section"}},
		{"Subsection 1.2", 3, []string{"Main Title", "Section 1", "Subsection 1.2"}},
		{"Section 2", 2, []string{"Main Title", "Section 2"}},
		{"Subsection 2.1", 3, []string{"Main Title", "Section 2", "Subsection 2.1"}},
	}

	if len(headings) != len(expectedHeadings) {
		t.Fatalf("FindAllHeadings() found %d headings, want %d", len(headings), len(expectedHeadings))
	}

	for i, heading := range headings {
		expected := expectedHeadings[i]

		if heading.Text != expected.text {
			t.Errorf("Heading[%d] text = %q, want %q", i, heading.Text, expected.text)
		}

		if heading.Level != expected.level {
			t.Errorf("Heading[%d] level = %d, want %d", i, heading.Level, expected.level)
		}

		if len(heading.Path) != len(expected.path) {
			t.Errorf("Heading[%d] path length = %d, want %d", i, len(heading.Path), len(expected.path))
		} else {
			for j, pathSeg := range heading.Path {
				if pathSeg != expected.path[j] {
					t.Errorf("Heading[%d] path[%d] = %q, want %q", i, j, pathSeg, expected.path[j])
				}
			}
		}

		// Verify offset is reasonable
		if heading.Offset < 0 || heading.Offset >= len(content) {
			t.Errorf("Heading[%d] offset %d out of range [0, %d)", i, heading.Offset, len(content))
		}
	}
}

func TestFindSubtreeAdvanced(t *testing.T) {
	content := `# Work

## Projects

### Frontend Development
Current frontend tasks.

#### React Components
- Button component
- Modal component

#### State Management
Redux and context setup.

### Backend API
Server-side development.

#### Authentication
User login system.

### Mobile Development
Cross-platform apps.

## Meetings

### Daily Standup
Team updates.
`

	doc := ParseDocument([]byte(content))

	tests := []struct {
		name         string
		pathStr      string
		expectError  bool
		expectedText string
		checkContent func(t *testing.T, subtree *Subtree, content []byte)
	}{
		{
			name:         "find top-level section",
			pathStr:      "test.md#Projects",
			expectError:  false,
			expectedText: "Projects",
			checkContent: func(t *testing.T, subtree *Subtree, content []byte) {
				if subtree.Level != 2 {
					t.Errorf("Expected level 2, got %d", subtree.Level)
				}
				if !strings.Contains(string(subtree.Content), "Frontend Development") {
					t.Errorf("Subtree should contain Frontend Development")
				}
				if !strings.Contains(string(subtree.Content), "Mobile Development") {
					t.Errorf("Subtree should contain Mobile Development")
				}
				if strings.Contains(string(subtree.Content), "## Meetings") {
					t.Errorf("Subtree should not contain Meetings section")
				}
			},
		},
		{
			name:         "find nested section",
			pathStr:      "test.md#Projects/Frontend",
			expectError:  false,
			expectedText: "Frontend Development",
			checkContent: func(t *testing.T, subtree *Subtree, content []byte) {
				if subtree.Level != 3 {
					t.Errorf("Expected level 3, got %d", subtree.Level)
				}
				if !strings.Contains(string(subtree.Content), "React Components") {
					t.Errorf("Subtree should contain React Components")
				}
				if !strings.Contains(string(subtree.Content), "State Management") {
					t.Errorf("Subtree should contain State Management")
				}
				if strings.Contains(string(subtree.Content), "Backend API") {
					t.Errorf("Subtree should not contain Backend API section")
				}
			},
		},
		{
			name:         "find deep nested section",
			pathStr:      "test.md#Projects/Frontend/React",
			expectError:  false,
			expectedText: "React Components",
			checkContent: func(t *testing.T, subtree *Subtree, content []byte) {
				if subtree.Level != 4 {
					t.Errorf("Expected level 4, got %d", subtree.Level)
				}
				if !strings.Contains(string(subtree.Content), "Button component") {
					t.Errorf("Subtree should contain Button component")
				}
				if strings.Contains(string(subtree.Content), "State Management") {
					t.Errorf("Subtree should not contain State Management section")
				}
			},
		},
		{
			name:        "nonexistent path",
			pathStr:     "test.md#NonExistent",
			expectError: true,
		},
		{
			name:        "ambiguous path should error",
			pathStr:     "test.md#Development", // Matches both Frontend and Mobile Development
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := ParsePath(tt.pathStr)
			if err != nil {
				t.Fatalf("ParsePath() error = %v", err)
			}

			subtree, err := FindSubtree(doc, []byte(content), path)

			if tt.expectError {
				if err == nil {
					t.Errorf("FindSubtree() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("FindSubtree() unexpected error = %v", err)
			}

			if subtree.Heading != tt.expectedText {
				t.Errorf("FindSubtree() heading = %q, want %q", subtree.Heading, tt.expectedText)
			}

			// Run additional content checks if provided
			if tt.checkContent != nil {
				tt.checkContent(t, subtree, []byte(content))
			}

			// Verify offset bounds
			if subtree.StartOffset < 0 || subtree.StartOffset >= len(content) {
				t.Errorf("StartOffset %d out of range", subtree.StartOffset)
			}
			if subtree.EndOffset <= subtree.StartOffset || subtree.EndOffset > len(content) {
				t.Errorf("EndOffset %d invalid (start: %d, content length: %d)",
					subtree.EndOffset, subtree.StartOffset, len(content))
			}

			// Verify content length matches offset difference
			expectedLength := subtree.EndOffset - subtree.StartOffset
			if len(subtree.Content) != expectedLength {
				t.Errorf("Content length %d doesn't match offset difference %d",
					len(subtree.Content), expectedLength)
			}
		})
	}
}

func TestCalculateLineNumberAdvanced(t *testing.T) {
	content := `Line 1
Line 2
Line 3
Line 4`

	tests := []struct {
		name     string
		offset   int
		expected int
	}{
		{"beginning of file", 0, 1},
		{"start of line 2", 7, 2},
		{"middle of line 2", 9, 2},
		{"start of line 3", 14, 3},
		{"end of content", len(content), 4},
		{"beyond content", len(content) + 10, 4},
		{"negative offset", -5, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateLineNumber([]byte(content), tt.offset)
			if result != tt.expected {
				t.Errorf("CalculateLineNumber() = %d, want %d", result, tt.expected)
			}
		})
	}
}

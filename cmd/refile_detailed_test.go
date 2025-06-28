package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/johncoder/jot/internal/markdown"
	"github.com/johncoder/jot/internal/workspace"
)

// TestSubtreeExtractionAndInsertion tests the detailed refile functionality
// to identify issues with subtree content handling
func TestSubtreeExtractionAndInsertion(t *testing.T) {
	tests := []struct {
		name            string
		sourceContent   string
		sourceSelector  string
		targetContent   string
		targetSelector  string
		expectedSource  string
		expectedTarget  string
		expectedErrors  []string
	}{
		{
			name: "h1_and_h2_level_preservation",
			sourceContent: `# Main Topic

Some intro content.

## Subtopic

Content under subtopic.

### Deep Topic

More nested content.

# Another Topic

Other content.
`,
			sourceSelector: "source.md#subtopic",
			targetContent: `# Work

## Projects

Existing project content.
`,
			targetSelector: "target.md#projects",
			expectedSource: `# Main Topic

Some intro content.

# Another Topic

Other content.
`,
			expectedTarget: `# Work

## Projects

Existing project content.

### Subtopic

Content under subtopic.

#### Deep Topic

More nested content.
`,
			expectedErrors: nil,
		},
		{
			name: "proper_newline_insertion",
			sourceContent: `# Source

## Task List

- Important task
- Another task

# Other Section

Content here.
`,
			sourceSelector: "source.md#task list",
			targetContent: `# Target
Content without trailing newline`,
			targetSelector: "target.md#",
			expectedSource: `# Source

# Other Section

Content here.
`,
			expectedTarget: `# Target
Content without trailing newline

## Task List

- Important task
- Another task
`,
			expectedErrors: nil,
		},
		{
			name: "no_errant_hash_tags",
			sourceContent: `# Source Doc

## Section A

Content with #hashtag in text.

### Subsection

More content with #another-tag.

## Section B

Other content.
`,
			sourceSelector: "source.md#section a",
			targetContent: `# Target

## Area

Existing content.
`,
			targetSelector: "target.md#area",
			expectedSource: `# Source Doc

## Section B

Other content.
`,
			expectedTarget: `# Target

## Area

Existing content.

### Section A

Content with #hashtag in text.

#### Subsection

More content with #another-tag.
`,
			expectedErrors: nil,
		},
		{
			name: "empty_heading_handling",
			sourceContent: `# Main

## Empty Section

## Another Section

Some content here.
`,
			sourceSelector: "source.md#empty section",
			targetContent: `# Target

Content.
`,
			targetSelector: "target.md#",
			expectedSource: `# Main

## Another Section

Some content here.
`,
			expectedTarget: `# Target

Content.

## Empty Section
`,
			expectedErrors: nil,
		},
		{
			name: "multiple_line_break_handling",
			sourceContent: `# Source

## Section


Content with multiple line breaks.


Another paragraph.

# End
`,
			sourceSelector: "source.md#section",
			targetContent: `# Target

Existing content.
`,
			targetSelector: "target.md#",
			expectedSource: `# Source

# End
`,
			expectedTarget: `# Target

Existing content.

## Section


Content with multiple line breaks.


Another paragraph.
`,
			expectedErrors: nil,
		},
		{
			name: "avoid_duplicate_content_on_same_file_refile",
			sourceContent: `# Inbox

This is your inbox for capturing new notes quickly. Use 'jot capture' to add new notes here.

---

# Test Heading

Some test content for refile.

## One - ASDF

Some content here.

## Two - ASDF

Some more content here.

## Three - ASDF

Even more content here.

## Four - ASDF

And even more content here.

## Five - ASDF

Last bit of content here.
`,
			sourceSelector: "source.md#test heading/two - asdf",
			targetContent: `# Inbox

This is your inbox for capturing new notes quickly. Use 'jot capture' to add new notes here.

---

# Test Heading

Some test content for refile.

## One - ASDF

Some content here.

## Two - ASDF

Some more content here.

## Three - ASDF

Even more content here.

## Four - ASDF

And even more content here.

## Five - ASDF

Last bit of content here.
`,
			targetSelector: "source.md#inbox", // Same file - this is the key change
			expectedSource: `# Inbox

This is your inbox for capturing new notes quickly. Use 'jot capture' to add new notes here.

---

## Two - ASDF

Some more content here.

# Test Heading

Some test content for refile.

## One - ASDF

Some content here.

## Three - ASDF

Even more content here.

## Four - ASDF

And even more content here.

## Five - ASDF

Last bit of content here.
`,
			expectedTarget: ``, // Empty because it's the same file as source
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary workspace
			tempDir, err := os.MkdirTemp("", "jot-refile-test-*")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tempDir)

			// Create workspace structure
			jotDir := filepath.Join(tempDir, ".jot")
			if err := os.MkdirAll(jotDir, 0755); err != nil {
				t.Fatalf("Failed to create .jot dir: %v", err)
			}

			// Write source file
			sourceFile := filepath.Join(tempDir, "source.md")
			if err := os.WriteFile(sourceFile, []byte(tt.sourceContent), 0644); err != nil {
				t.Fatalf("Failed to write source file: %v", err)
			}

			// Write target file (only if different from source)
			targetFile := filepath.Join(tempDir, "target.md")
			sourcePathCheck, _ := markdown.ParsePath(tt.sourceSelector)
			destPathCheck, _ := markdown.ParsePath(tt.targetSelector)
			sameFile := sourcePathCheck.File == destPathCheck.File
			
			if !sameFile {
				if err := os.WriteFile(targetFile, []byte(tt.targetContent), 0644); err != nil {
					t.Fatalf("Failed to write target file: %v", err)
				}
			} else {
				// When same file, target file path should point to source file
				targetFile = sourceFile
			}

			// Create workspace
			ws := &workspace.Workspace{
				Root:      tempDir,
				JotDir:    jotDir,
				InboxPath: filepath.Join(tempDir, "inbox.md"),
				LibDir:    filepath.Join(tempDir, "lib"),
			}

			// Parse selectors
			sourcePath, err := markdown.ParsePath(tt.sourceSelector)
			if err != nil {
				t.Fatalf("Failed to parse source selector: %v", err)
			}

			destPath, err := markdown.ParsePath(tt.targetSelector)
			if err != nil {
				t.Fatalf("Failed to parse target selector: %v", err)
			}

			// Extract subtree
			subtree, err := ExtractSubtree(ws, sourcePath)
			if err != nil {
				t.Fatalf("ExtractSubtree() error = %v", err)
			}

			// Resolve destination
			dest, err := ResolveDestination(ws, destPath, false)
			if err != nil {
				t.Fatalf("ResolveDestination() error = %v", err)
			}

			// Transform content
			transformedContent := TransformSubtreeLevel(subtree, dest.TargetLevel)

			// Debug output
			t.Logf("Original subtree content: %q", string(subtree.Content))
			t.Logf("Transformed content: %q", string(transformedContent))
			t.Logf("Target level: %d, Source level: %d", dest.TargetLevel, subtree.Level)

			// Perform refile
			err = performRefile(ws, sourcePath, subtree, dest, transformedContent)
			if err != nil {
				if tt.expectedErrors == nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				errorFound := false
				for _, expectedError := range tt.expectedErrors {
					if strings.Contains(err.Error(), expectedError) {
						errorFound = true
						break
					}
				}
				if !errorFound {
					t.Fatalf("Error %q does not match expected errors %v", err.Error(), tt.expectedErrors)
				}
				return // Test passed with expected error
			}

			// Read results
			sourceResult, err := os.ReadFile(sourceFile)
			if err != nil {
				t.Fatalf("Failed to read source file: %v", err)
			}

			// Compare results
			if strings.TrimSpace(string(sourceResult)) != strings.TrimSpace(tt.expectedSource) {
				t.Errorf("Source file mismatch:\nGot:\n%s\n\nExpected:\n%s", string(sourceResult), tt.expectedSource)
			}

			// Only check target file if it's different from source and expectedTarget is not empty
			if !sameFile && tt.expectedTarget != "" {
				targetResult, err := os.ReadFile(targetFile)
				if err != nil {
					t.Fatalf("Failed to read target file: %v", err)
				}

				if strings.TrimSpace(string(targetResult)) != strings.TrimSpace(tt.expectedTarget) {
					t.Errorf("Target file mismatch:\nGot:\n%s\n\nExpected:\n%s", string(targetResult), tt.expectedTarget)
				}
			}
		})
	}
}

// TestHeadingLevelTransformation tests the heading level transformation logic specifically
func TestHeadingLevelTransformation(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		sourceLevel    int
		targetLevel    int
		expectedResult string
	}{
		{
			name:        "h1_to_h3_transformation",
			content:     "# Main\n\n## Sub\n\nContent\n\n### Deep\n\nMore content",
			sourceLevel: 1,
			targetLevel: 3,
			expectedResult: "### Main\n\n#### Sub\n\nContent\n\n##### Deep\n\nMore content",
		},
		{
			name:        "h2_to_h1_transformation",
			content:     "## Title\n\n### Subtitle\n\nText content",
			sourceLevel: 2,
			targetLevel: 1,
			expectedResult: "# Title\n\n## Subtitle\n\nText content",
		},
		{
			name:        "preserve_non_headings_with_hashes",
			content:     "# Main\n\nText with #hashtag and #another-tag\n\n## Sub\n\nMore #tags here",
			sourceLevel: 1,
			targetLevel: 2,
			expectedResult: "## Main\n\nText with #hashtag and #another-tag\n\n### Sub\n\nMore #tags here",
		},
		{
			name:        "handle_max_heading_level",
			content:     "##### Level 5\n\n###### Level 6\n\nContent",
			sourceLevel: 5,
			targetLevel: 6,
			expectedResult: "###### Level 5\n\n###### Level 6\n\nContent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			levelDiff := tt.targetLevel - tt.sourceLevel
			result := markdown.TransformHeadingLevels([]byte(tt.content), levelDiff)
			
			if string(result) != tt.expectedResult {
				t.Errorf("Transformation mismatch:\nGot:\n%s\n\nExpected:\n%s", string(result), tt.expectedResult)
			}
		})
	}
}

// TestNewlineHandling tests specific newline and whitespace handling issues
func TestNewlineHandling(t *testing.T) {
	// Test simple insertion logic directly
	content := []byte("# Target\n\nExisting content\n")
	insertOffset := len("# Target\n")
	insertContent := []byte("\n## Inserted\n\nContent\n")
	
	result := append(content[:insertOffset], insertContent...)
	result = append(result, content[insertOffset:]...)
	
	expected := "# Target\n\n## Inserted\n\nContent\n\nExisting content\n"
	
	if string(result) != expected {
		t.Errorf("Basic insertion test failed:\nGot: %q\nExpected: %q", string(result), expected)
		t.Logf("Content before: %q", string(content[:insertOffset]))
		t.Logf("Insert content: %q", string(insertContent))
		t.Logf("Content after: %q", string(content[insertOffset:]))
	}
}

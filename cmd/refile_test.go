package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/johncoder/jot/internal/markdown"
	"github.com/johncoder/jot/internal/workspace"
	"github.com/yuin/goldmark/ast"
)

// Test data for path resolution tests
var testMarkdownContent = `# Work

## Projects

### Frontend Development
Current frontend tasks and components.

#### React Components
- Button component
- Modal component

### Backend API  
Database work and API endpoints.

#### Authentication
User login and session management.

### Mobile Development
Cross-platform mobile apps.

## Meetings

### Daily Standup
Team coordination and updates.

### Sprint Planning
Quarterly planning sessions.

## Tasks

Current work items and todos.
`

func TestCalculatePathMatch(t *testing.T) {
	tests := []struct {
		name           string
		headingPath    []string
		targetSegments []string
		skipLevels     int
		expectedMatch  int
	}{
		{
			name:           "exact match at root level",
			headingPath:    []string{"Work", "Projects"},
			targetSegments: []string{"Projects"},
			skipLevels:     1,
			expectedMatch:  1,
		},
		{
			name:           "hierarchical exact match",
			headingPath:    []string{"Work", "Projects", "Frontend Development"},
			targetSegments: []string{"Projects", "Frontend"},
			skipLevels:     1,
			expectedMatch:  2,
		},
		{
			name:           "contains matching",
			headingPath:    []string{"Work", "Projects", "Backend API"},
			targetSegments: []string{"projects", "backend"},
			skipLevels:     1,
			expectedMatch:  2,
		},
		{
			name:           "partial match only",
			headingPath:    []string{"Work", "Projects", "Frontend Development"},
			targetSegments: []string{"Projects", "Mobile"},
			skipLevels:     1,
			expectedMatch:  1,
		},
		{
			name:           "no match",
			headingPath:    []string{"Work", "Meetings", "Daily Standup"},
			targetSegments: []string{"Projects", "Backend"},
			skipLevels:     1,
			expectedMatch:  0,
		},
		{
			name:           "skip levels matching",
			headingPath:    []string{"Work", "Projects", "Frontend Development"},
			targetSegments: []string{"Frontend"},
			skipLevels:     2,
			expectedMatch:  1,
		},
		{
			name:           "nested deep match",
			headingPath:    []string{"Work", "Projects", "Frontend Development", "React Components"},
			targetSegments: []string{"Projects", "Frontend", "React"},
			skipLevels:     1,
			expectedMatch:  3,
		},
		{
			name:           "case insensitive matching",
			headingPath:    []string{"Work", "Projects", "Backend API"},
			targetSegments: []string{"PROJECTS", "backend"},
			skipLevels:     1,
			expectedMatch:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculatePathMatch(tt.headingPath, tt.targetSegments, tt.skipLevels)
			if result != tt.expectedMatch {
				t.Errorf("calculatePathMatch() = %d, want %d", result, tt.expectedMatch)
			}
		})
	}
}

func TestNavigateHeadingPath(t *testing.T) {
	// Parse the test markdown content
	doc := markdown.ParseDocument([]byte(testMarkdownContent))
	content := []byte(testMarkdownContent)

	tests := []struct {
		name              string
		pathSegments      []string
		skipLevels        int
		expectTargetFound bool
		expectedFound     []string
		expectedMissing   []string
	}{
		{
			name:              "find existing single segment",
			pathSegments:      []string{"Projects"},
			skipLevels:        1,
			expectTargetFound: true,
			expectedFound:     []string{"Projects"},
			expectedMissing:   []string{},
		},
		{
			name:              "find existing hierarchical path",
			pathSegments:      []string{"Projects", "Frontend"},
			skipLevels:        1,
			expectTargetFound: true,
			expectedFound:     []string{"Projects", "Frontend"},
			expectedMissing:   []string{},
		},
		{
			name:              "partial match with missing segment",
			pathSegments:      []string{"Projects", "Mobile", "iOS"},
			skipLevels:        1,
			expectTargetFound: false,
			expectedFound:     []string{"Projects", "Mobile"},
			expectedMissing:   []string{"iOS"},
		},
		{
			name:              "completely missing path",
			pathSegments:      []string{"Archive", "Old Projects"},
			skipLevels:        1,
			expectTargetFound: false,
			expectedFound:     []string{},
			expectedMissing:   []string{"Archive", "Old Projects"},
		},
		{
			name:              "contains matching",
			pathSegments:      []string{"proj", "backend"},
			skipLevels:        1,
			expectTargetFound: true,
			expectedFound:     []string{"proj", "backend"},
			expectedMissing:   []string{},
		},
		{
			name:              "deep nested path",
			pathSegments:      []string{"Projects", "Frontend", "React"},
			skipLevels:        1,
			expectTargetFound: true,
			expectedFound:     []string{"Projects", "Frontend", "React"},
			expectedMissing:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			destPath := &markdown.HeadingPath{
				File:       "test.md",
				Segments:   tt.pathSegments,
				SkipLevels: tt.skipLevels,
			}

			result, err := navigateHeadingPath(doc, content, destPath)
			if err != nil {
				t.Fatalf("navigateHeadingPath() error = %v", err)
			}

			// Check if target was found
			targetFound := result.TargetHeading != nil
			if targetFound != tt.expectTargetFound {
				t.Errorf("navigateHeadingPath() target found = %v, want %v", targetFound, tt.expectTargetFound)
			}

			// Check found segments
			if !sliceEqual(result.FoundSegments, tt.expectedFound) {
				t.Errorf("navigateHeadingPath() found segments = %v, want %v", result.FoundSegments, tt.expectedFound)
			}

			// Check missing segments
			if !sliceEqual(result.MissingSegments, tt.expectedMissing) {
				t.Errorf("navigateHeadingPath() missing segments = %v, want %v", result.MissingSegments, tt.expectedMissing)
			}
		})
	}
}

func TestResolveDestinationPath(t *testing.T) {
	// Parse the test markdown content
	doc := markdown.ParseDocument([]byte(testMarkdownContent))
	content := []byte(testMarkdownContent)

	tests := []struct {
		name             string
		pathSegments     []string
		skipLevels       int
		prepend          bool
		expectExists     bool
		expectedLevel    int
		expectedCreate   []string
	}{
		{
			name:           "resolve existing path",
			pathSegments:   []string{"Projects", "Frontend"},
			skipLevels:     1,
			prepend:        false,
			expectExists:   true,
			expectedLevel:  4, // Should insert at level 4 under Frontend (level 3)
			expectedCreate: []string{},
		},
		{
			name:           "resolve partial path needing creation",
			pathSegments:   []string{"Projects", "Mobile", "iOS"},
			skipLevels:     1,
			prepend:        false,
			expectExists:   false,
			expectedLevel:  5, // Mobile level 4 + iOS level 5 = content at level 5
			expectedCreate: []string{"iOS"},
		},
		{
			name:           "resolve completely new path",
			pathSegments:   []string{"Archive", "Old Projects"},
			skipLevels:     1,
			prepend:        false,
			expectExists:   false,
			expectedLevel:  3, // Archive level 1 + Old Projects level 2 = content at level 3
			expectedCreate: []string{"Archive", "Old Projects"},
		},
		{
			name:           "resolve with prepend",
			pathSegments:   []string{"Projects"},
			skipLevels:     1,
			prepend:        true,
			expectExists:   true,
			expectedLevel:  3, // Should insert at level 3 under Projects (level 2)
			expectedCreate: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			destPath := &markdown.HeadingPath{
				File:       "test.md",
				Segments:   tt.pathSegments,
				SkipLevels: tt.skipLevels,
			}

			result, err := resolveDestinationPath(doc, content, destPath, tt.prepend)
			if err != nil {
				t.Fatalf("resolveDestinationPath() error = %v", err)
			}

			// Check if path exists
			if result.Exists != tt.expectExists {
				t.Errorf("resolveDestinationPath() exists = %v, want %v", result.Exists, tt.expectExists)
			}

			// Check target level
			if result.TargetLevel != tt.expectedLevel {
				t.Errorf("resolveDestinationPath() target level = %d, want %d", result.TargetLevel, tt.expectedLevel)
			}

			// Check create path
			if !sliceEqual(result.CreatePath, tt.expectedCreate) {
				t.Errorf("resolveDestinationPath() create path = %v, want %v", result.CreatePath, tt.expectedCreate)
			}

			// Check that insert offset is reasonable
			if result.InsertOffset < 0 || result.InsertOffset > len(content) {
				t.Errorf("resolveDestinationPath() insert offset %d out of range [0, %d]", result.InsertOffset, len(content))
			}
		})
	}
}

func TestCalculateInsertionPoint(t *testing.T) {
	// Create a simple test document
	testContent := `# Work

## Projects

### Frontend
Some content here

### Backend
More content

## Meetings
Meeting notes
`

	doc := markdown.ParseDocument([]byte(testContent))
	content := []byte(testContent)

	// Find the "Projects" heading to test insertion
	var projectsHeading *ast.Heading
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if heading, ok := n.(*ast.Heading); ok {
			text := markdown.ExtractHeadingText(heading, content)
			if text == "Projects" {
				projectsHeading = heading
				return ast.WalkStop, nil
			}
		}
		return ast.WalkContinue, nil
	})

	if projectsHeading == nil {
		t.Fatal("Could not find Projects heading in test content")
	}

	tests := []struct {
		name     string
		prepend  bool
		validate func(t *testing.T, offset int, content []byte)
	}{
		{
			name:    "append mode",
			prepend: false,
			validate: func(t *testing.T, offset int, content []byte) {
				// Should be after the Backend section but before Meetings
				if offset <= 0 || offset >= len(content) {
					t.Errorf("Insert offset %d out of range", offset)
				}
				// Check that we're before "## Meetings"
				remainingContent := string(content[offset:])
				if !strings.Contains(remainingContent, "## Meetings") {
					t.Errorf("Insert point should be before Meetings section")
				}
			},
		},
		{
			name:    "prepend mode",
			prepend: true,
			validate: func(t *testing.T, offset int, content []byte) {
				// Should be right after the Projects heading line
				if offset <= 0 || offset >= len(content) {
					t.Errorf("Insert offset %d out of range", offset)
				}
				// Should be after "## Projects\n"
				beforeContent := string(content[:offset])
				if !strings.Contains(beforeContent, "## Projects") {
					t.Errorf("Insert point should be after Projects heading")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			offset := calculateInsertionPoint(projectsHeading, content, tt.prepend)
			tt.validate(t, offset, content)
		})
	}
}

func TestFindHeadingByOffset(t *testing.T) {
	doc := markdown.ParseDocument([]byte(testMarkdownContent))
	content := []byte(testMarkdownContent)

	// Find all headings and their offsets
	allHeadings := markdown.FindAllHeadings(doc, content)

	for _, headingInfo := range allHeadings {
		t.Run("find_"+headingInfo.Text, func(t *testing.T) {
			found := findHeadingByOffset(doc, headingInfo.Offset)
			if found == nil {
				t.Errorf("findHeadingByOffset() could not find heading at offset %d", headingInfo.Offset)
				return
			}

			// Verify we found the correct heading
			foundText := markdown.ExtractHeadingText(found, content)
			if foundText != headingInfo.Text {
				t.Errorf("findHeadingByOffset() found heading %q, want %q", foundText, headingInfo.Text)
			}
		})
	}
}

// Integration test for the complete refile workflow
func TestRefileIntegration(t *testing.T) {
	// Create a temporary workspace for testing
	tempDir := t.TempDir()
	
	// Create workspace structure
	jotDir := filepath.Join(tempDir, ".jot")
	libDir := filepath.Join(tempDir, "lib")
	if err := os.MkdirAll(jotDir, 0755); err != nil {
		t.Fatalf("Failed to create .jot directory: %v", err)
	}
	if err := os.MkdirAll(libDir, 0755); err != nil {
		t.Fatalf("Failed to create lib directory: %v", err)
	}

	// Create test source file
	sourceContent := `# Source

## Important Task
This is an important task that needs to be done.

### Details
- Priority: High
- Due: Tomorrow

## Another Task
Different task here.
`

	sourceFile := filepath.Join(libDir, "source.md")
	if err := os.WriteFile(sourceFile, []byte(sourceContent), 0644); err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	// Create test destination file
	destContent := `# Work

## Projects

### Frontend
Current tasks

## Archive
Old stuff
`

	destFile := filepath.Join(libDir, "work.md")
	if err := os.WriteFile(destFile, []byte(destContent), 0644); err != nil {
		t.Fatalf("Failed to write destination file: %v", err)
	}

	// Create workspace
	ws := &workspace.Workspace{
		Root:      tempDir,
		JotDir:    jotDir,
		InboxPath: filepath.Join(tempDir, "inbox.md"),
		LibDir:    libDir,
	}

	// Test the refile operation
	sourcePath := &markdown.HeadingPath{
		File:     "lib/source.md",
		Segments: []string{"Source", "Important"},
	}

	destPath := &markdown.HeadingPath{
		File:     "lib/work.md",
		Segments: []string{"Projects", "Backend"},
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

	// Perform refile
	err = performRefile(ws, sourcePath, subtree, dest, transformedContent)
	if err != nil {
		t.Fatalf("performRefile() error = %v", err)
	}

	// Verify source file
	sourceResult, err := os.ReadFile(sourceFile)
	if err != nil {
		t.Fatalf("Failed to read source file after refile: %v", err)
	}

	if strings.Contains(string(sourceResult), "Important Task") {
		t.Errorf("Source file still contains refiled content")
	}

	// Verify destination file
	destResult, err := os.ReadFile(destFile)
	if err != nil {
		t.Fatalf("Failed to read destination file after refile: %v", err)
	}

	destStr := string(destResult)
	if !strings.Contains(destStr, "Backend") {
		t.Errorf("Destination file missing created Backend section")
	}
	if !strings.Contains(destStr, "Important Task") {
		t.Errorf("Destination file missing refiled content")
	}
	if !strings.Contains(destStr, "Priority: High") {
		t.Errorf("Destination file missing nested content from refile")
	}
}

// Helper function to compare string slices
func sliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

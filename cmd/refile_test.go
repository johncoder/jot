package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/johncoder/jot/internal/markdown"
	"github.com/johncoder/jot/internal/workspace"
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
			// Use the new path-based matching logic, which is case-insensitive and supports partial/contains matches
			match := markdown.CalculatePathMatch(tt.headingPath, tt.targetSegments, tt.skipLevels)
			if match != tt.expectedMatch {
				t.Errorf("CalculatePathMatch() = %d, want %d", match, tt.expectedMatch)
			}
		})
	}
}

func TestNavigateHeadingPath(t *testing.T) {
	// ...existing code...

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

			// Use the new AST/path-based navigation logic
			doc := markdown.ParseMarkdownAST([]byte(testMarkdownContent))
			result, err := markdown.NavigateHeadingPath(doc, destPath)
			if err != nil {
				t.Fatalf("NavigateHeadingPath() error = %v", err)
			}

			// Check if target was found
			targetFound := result.TargetHeading != nil
			if targetFound != tt.expectTargetFound {
				t.Errorf("NavigateHeadingPath() target found = %v, want %v", targetFound, tt.expectTargetFound)
			}

			// Check found segments
			if !sliceEqual(result.FoundSegments, tt.expectedFound) {
				t.Errorf("NavigateHeadingPath() found segments = %v, want %v", result.FoundSegments, tt.expectedFound)
			}

			// Check missing segments
			if !sliceEqual(result.MissingSegments, tt.expectedMissing) {
				t.Errorf("NavigateHeadingPath() missing segments = %v, want %v", result.MissingSegments, tt.expectedMissing)
			}
		})
	}
}

func TestResolveDestinationPath(t *testing.T) {
	// ...existing code...

	tests := []struct {
		name           string
		pathSegments   []string
		skipLevels     int
		prepend        bool
		expectExists   bool
		expectedLevel  int
		expectedCreate []string
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

			// Use the new AST/path-based destination resolution logic
			// This logic is now in cmd/refile_pipeline.go: ResolveDestination
			ws := &workspace.Workspace{Root: ".", LibDir: "."}
			dest, err := ResolveDestination(ws, destPath, tt.prepend)
			if err != nil {
				t.Fatalf("ResolveDestination() error = %v", err)
			}

			// Check if path exists
			if dest.Exists != tt.expectExists {
				t.Errorf("ResolveDestination() exists = %v, want %v", dest.Exists, tt.expectExists)
			}

			// Check target level
			if dest.TargetLevel != tt.expectedLevel {
				t.Errorf("ResolveDestination() target level = %d, want %d", dest.TargetLevel, tt.expectedLevel)
			}

			// Check create path
			if !sliceEqual(dest.CreatePath, tt.expectedCreate) {
				t.Errorf("ResolveDestination() create path = %v, want %v", dest.CreatePath, tt.expectedCreate)
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

	// Extract subtree using the new AST/path-based logic
	subtree, err := ExtractSubtree(ws, sourcePath)
	if err != nil {
		t.Fatalf("ExtractSubtree() error = %v", err)
	}

	// Resolve destination using the new logic
	dest, err := ResolveDestination(ws, destPath, false)
	if err != nil {
		t.Fatalf("ResolveDestination() error = %v", err)
	}

	// Transform subtree level for correct insertion
	transformedContent := TransformSubtreeLevel(subtree, dest.TargetLevel)

	// Perform refile using the new logic
	err = PerformRefile(ws, sourcePath, subtree, dest, transformedContent)
	if err != nil {
		t.Fatalf("PerformRefile() error = %v", err)
	}

	// Verify source file no longer contains refiled content
	sourceResult, err := os.ReadFile(sourceFile)
	if err != nil {
		t.Fatalf("Failed to read source file after refile: %v", err)
	}

	if strings.Contains(string(sourceResult), "Important Task") {
		t.Errorf("Source file still contains refiled content")
	}

	// Verify destination file contains refiled content and created headings
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

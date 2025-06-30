package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/johncoder/jot/internal/markdown"
	"github.com/johncoder/jot/internal/workspace"
)

func TestPeekCommand(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()

	// Create workspace structure
	jotDir := filepath.Join(tempDir, ".jot")
	libDir := filepath.Join(tempDir, "lib")

	if err := os.MkdirAll(jotDir, 0755); err != nil {
		t.Fatalf("Failed to create jot directory: %v", err)
	}
	if err := os.MkdirAll(libDir, 0755); err != nil {
		t.Fatalf("Failed to create lib directory: %v", err)
	}

	// Create test content
	testContent := `# Documentation

## User Guide

### Getting Started
This section explains how to get started with jot.

#### Installation
Download and install jot from the official repository.

#### Configuration
Set up your workspace and configuration files.

### Advanced Features

#### Custom Templates
Create custom templates for different note types.

## API Reference

### Core Functions
Documentation for the main API functions.

### Utilities
Helper functions and utilities.
`

	testFile := filepath.Join(libDir, "test_peek.md")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create workspace
	ws := &workspace.Workspace{
		Root:      tempDir,
		JotDir:    jotDir,
		InboxPath: filepath.Join(tempDir, "inbox.md"),
		LibDir:    libDir,
	}

	tests := []struct {
		name            string
		selector        string
		expectedHeading string
		expectedLevel   int
		expectError     bool
	}{
		{
			name:            "peek user guide section",
			selector:        "lib/test_peek.md#doc/user",
			expectedHeading: "User Guide",
			expectedLevel:   2,
			expectError:     false,
		},
		{
			name:            "peek getting started subsection",
			selector:        "lib/test_peek.md#doc/user/getting",
			expectedHeading: "Getting Started",
			expectedLevel:   3,
			expectError:     false,
		},
		{
			name:            "peek installation with skip levels",
			selector:        "lib/test_peek.md#///install",
			expectedHeading: "Installation",
			expectedLevel:   4,
			expectError:     false,
		},
		{
			name:            "peek non-existent section",
			selector:        "lib/test_peek.md#nonexistent",
			expectedHeading: "",
			expectedLevel:   0,
			expectError:     true,
		},
		{
			name:            "peek with contains matching",
			selector:        "lib/test_peek.md#doc/api",
			expectedHeading: "API Reference",
			expectedLevel:   2,
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the selector
			sourcePath, err := markdown.ParsePath(tt.selector)
			if err != nil {
				t.Fatalf("Failed to parse selector: %v", err)
			}

			// Extract subtree
			subtree, err := ExtractSubtree(ws, sourcePath)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("ExtractSubtree() error = %v", err)
			}

			// Verify results
			if subtree.Heading != tt.expectedHeading {
				t.Errorf("Expected heading %q, got %q", tt.expectedHeading, subtree.Heading)
			}

			if subtree.Level != tt.expectedLevel {
				t.Errorf("Expected level %d, got %d", tt.expectedLevel, subtree.Level)
			}

			// Verify content is not empty (unless it's supposed to be)
			if len(subtree.Content) == 0 {
				t.Errorf("Subtree content is empty")
			}

			// Verify content starts with the heading
			contentStr := string(subtree.Content)
			if !strings.Contains(contentStr, tt.expectedHeading) {
				t.Errorf("Subtree content doesn't contain expected heading %q", tt.expectedHeading)
			}
		})
	}
}

func TestCountNestedHeadings(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		baseLevel int
		expected  int
	}{
		{
			name: "count nested headings",
			content: `## Base Heading

### Nested Level 1

#### Nested Level 2

##### Nested Level 3

### Another Nested Level 1`,
			baseLevel: 2,
			expected:  4,
		},
		{
			name: "no nested headings",
			content: `## Base Heading

Some content without nested headings.`,
			baseLevel: 2,
			expected:  0,
		},
		{
			name: "mixed levels",
			content: `# Top Level

## Level 2

### Level 3

## Another Level 2

#### Level 4

# Another Top Level`,
			baseLevel: 1,
			expected:  4, // 2 level-2, 1 level-3, 1 level-4
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countNestedHeadings([]byte(tt.content), tt.baseLevel)
			if result != tt.expected {
				t.Errorf("countNestedHeadings() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestSplitLines(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "single line",
			content:  "single line",
			expected: []string{"single line"},
		},
		{
			name:     "multiple lines",
			content:  "line 1\nline 2\nline 3",
			expected: []string{"line 1", "line 2", "line 3"},
		},
		{
			name:     "empty lines",
			content:  "line 1\n\nline 3",
			expected: []string{"line 1", "", "line 3"},
		},
		{
			name:     "trailing newline",
			content:  "line 1\nline 2\n",
			expected: []string{"line 1", "line 2", ""},
		},
		{
			name:     "empty content",
			content:  "",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitLines([]byte(tt.content))
			if len(result) != len(tt.expected) {
				t.Errorf("splitLines() length = %d, want %d", len(result), len(tt.expected))
				return
			}
			for i, line := range result {
				if line != tt.expected[i] {
					t.Errorf("splitLines()[%d] = %q, want %q", i, line, tt.expected[i])
				}
			}
		})
	}
}

func TestExtractHeadingsFromContent(t *testing.T) {
	testContent := `# Top Level

## Second Level

### Third Level
Some content here.

#### Fourth Level

## Another Second Level

### Another Third Level`

	doc := markdown.ParseDocument([]byte(testContent))
	headings := extractHeadingsFromContent(doc, []byte(testContent))

	expected := []HeadingInfo{
		{Text: "Top Level", Level: 1},
		{Text: "Second Level", Level: 2},
		{Text: "Third Level", Level: 3},
		{Text: "Fourth Level", Level: 4},
		{Text: "Another Second Level", Level: 2},
		{Text: "Another Third Level", Level: 3},
	}

	if len(headings) != len(expected) {
		t.Errorf("Expected %d headings, got %d", len(expected), len(headings))
		return
	}

	for i, heading := range headings {
		if heading.Text != expected[i].Text {
			t.Errorf("Heading %d: expected text %q, got %q", i, expected[i].Text, heading.Text)
		}
		if heading.Level != expected[i].Level {
			t.Errorf("Heading %d: expected level %d, got %d", i, expected[i].Level, heading.Level)
		}
	}
}

func TestShowTableOfContents(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()

	// Create workspace structure
	jotDir := filepath.Join(tempDir, ".jot")
	libDir := filepath.Join(tempDir, "lib")

	if err := os.MkdirAll(jotDir, 0755); err != nil {
		t.Fatalf("Failed to create jot directory: %v", err)
	}
	if err := os.MkdirAll(libDir, 0755); err != nil {
		t.Fatalf("Failed to create lib directory: %v", err)
	}

	// Create test content for TOC testing
	tocTestContent := `# Project Documentation

## Overview
High-level project overview.

## Getting Started

### Prerequisites
What you need before starting.

### Installation
How to install the project.

#### From Source
Building from source code.

#### Binary Release
Using pre-built binaries.

### Configuration

#### Basic Setup
Essential configuration steps.

#### Advanced Options
Optional advanced configuration.

## API Documentation

### Authentication
How to authenticate with the API.

### Endpoints

#### User Management
User-related API endpoints.

#### Data Operations
Data manipulation endpoints.

## Troubleshooting

### Common Issues
Frequently encountered problems.

### Support
How to get help.
`

	testFile := filepath.Join(libDir, "toc_test.md")
	if err := os.WriteFile(testFile, []byte(tocTestContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create workspace
	ws := &workspace.Workspace{
		Root:      tempDir,
		JotDir:    jotDir,
		InboxPath: filepath.Join(tempDir, "inbox.md"),
		LibDir:    libDir,
	}

	tests := []struct {
		name        string
		selector    string
		expectError bool
		checkFunc   func(t *testing.T, output string)
	}{
		{
			name:        "full file TOC",
			selector:    "lib/toc_test.md",
			expectError: false,
			checkFunc: func(t *testing.T, output string) {
				// Should contain the table of contents header
				if !strings.Contains(output, "Table of Contents: lib/toc_test.md") {
					t.Error("Expected TOC header not found")
				}

				// Should contain major headings
				if !strings.Contains(output, "# Project Documentation") {
					t.Error("Expected top-level heading not found")
				}
				if !strings.Contains(output, "## Overview") {
					t.Error("Expected second-level heading not found")
				}
				if !strings.Contains(output, "### Prerequisites") {
					t.Error("Expected third-level heading not found")
				}
				if !strings.Contains(output, "#### From Source") {
					t.Error("Expected fourth-level heading not found")
				}

				// Should contain selector hints
				if !strings.Contains(output, "jot peek") {
					t.Error("Expected selector hints not found")
				}
			},
		},
		{
			name:        "subtree TOC",
			selector:    "lib/toc_test.md#doc/getting",
			expectError: false,
			checkFunc: func(t *testing.T, output string) {
				// Should contain the subtree TOC header
				if !strings.Contains(output, "Table of Contents: lib/toc_test.md#doc/getting") {
					t.Error("Expected subtree TOC header not found")
				}

				// Should contain subtree headings but not others
				if !strings.Contains(output, "### Prerequisites") {
					t.Error("Expected subtree heading not found")
				}
				if !strings.Contains(output, "### Installation") {
					t.Error("Expected subtree heading not found")
				}
				if strings.Contains(output, "## Overview") {
					t.Error("Should not contain headings from outside subtree")
				}
				if strings.Contains(output, "## API Documentation") {
					t.Error("Should not contain headings from outside subtree")
				}
			},
		},
		{
			name:        "non-existent file",
			selector:    "nonexistent.md",
			expectError: true,
			checkFunc:   nil,
		},
		{
			name:        "invalid subtree selector",
			selector:    "lib/toc_test.md#nonexistent",
			expectError: true,
			checkFunc:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout for testing
			// For this test, we'll check that the function doesn't panic
			// and returns appropriate errors
			err := showTableOfContents(ws, tt.selector, false, false) // Use default (non-short) selectors for tests, workspace mode

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Note: For a more comprehensive test, we would need to capture
			// the actual output. For now, we're just testing that the function
			// executes without errors and handles edge cases properly.
		})
	}
}

func TestTableOfContentsEdgeCases(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()

	// Create workspace structure
	jotDir := filepath.Join(tempDir, ".jot")
	libDir := filepath.Join(tempDir, "lib")

	if err := os.MkdirAll(jotDir, 0755); err != nil {
		t.Fatalf("Failed to create jot directory: %v", err)
	}
	if err := os.MkdirAll(libDir, 0755); err != nil {
		t.Fatalf("Failed to create lib directory: %v", err)
	}

	// Create workspace
	ws := &workspace.Workspace{
		Root:      tempDir,
		JotDir:    jotDir,
		InboxPath: filepath.Join(tempDir, "inbox.md"),
		LibDir:    libDir,
	}

	tests := []struct {
		name        string
		filename    string
		content     string
		expectError bool
	}{
		{
			name:        "empty file",
			filename:    "lib/empty.md",
			content:     "",
			expectError: false, // Should handle gracefully
		},
		{
			name:        "file with no headings",
			filename:    "lib/no_headings.md",
			content:     "Just some text without any headings.\n\nMore text here.",
			expectError: false, // Should handle gracefully
		},
		{
			name:        "file with only content",
			filename:    "lib/content_only.md",
			content:     "This file has content but no markdown headings at all.",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test file (relative to workspace root)
			testFile := filepath.Join(tempDir, tt.filename)
			// Ensure parent directory exists
			if err := os.MkdirAll(filepath.Dir(testFile), 0755); err != nil {
				t.Fatalf("Failed to create parent directory: %v", err)
			}
			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			// Test TOC function
			err := showTableOfContents(ws, tt.filename, false, false) // Use default (non-short) selectors for tests, workspace mode

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestShortSelectorGeneration(t *testing.T) {
	// Create test headings that represent a document structure
	headings := []HeadingInfo{
		{Text: "Project", Level: 1, Line: 1},
		{Text: "Getting Started", Level: 2, Line: 3},
		{Text: "Installation", Level: 3, Line: 5},
		{Text: "Configuration", Level: 3, Line: 7},
		{Text: "Advanced", Level: 2, Line: 9},
		{Text: "Plugin System", Level: 3, Line: 11},
		{Text: "Performance", Level: 3, Line: 13},
		{Text: "Getting Started", Level: 2, Line: 15}, // Duplicate
		{Text: "Quick Start", Level: 3, Line: 17},
	}

	tests := []struct {
		name           string
		targetHeading  HeadingInfo
		expectedShort  string
		expectedNormal string
	}{
		{
			name:           "unique level 1 heading",
			targetHeading:  HeadingInfo{Text: "Project", Level: 1, Line: 1},
			expectedShort:  `jot peek "test.md#pr"`,
			expectedNormal: `jot peek "test.md#project"`,
		},
		{
			name:           "unique level 3 heading with short skip-level",
			targetHeading:  HeadingInfo{Text: "Performance", Level: 3, Line: 13},
			expectedShort:  `jot peek "test.md#//pe"`,
			expectedNormal: `jot peek "test.md#project/advanced/performance"`,
		},
		{
			name:           "unique level 3 heading - installation",
			targetHeading:  HeadingInfo{Text: "Installation", Level: 3, Line: 5},
			expectedShort:  `jot peek "test.md#//ins"`,
			expectedNormal: `jot peek "test.md#project/getting started/installation"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shortSelector := generateShortSelector("test.md", tt.targetHeading, headings)
			normalSelector := generateOptimalSelector("test.md", tt.targetHeading, headings)

			if shortSelector != tt.expectedShort {
				t.Errorf("generateShortSelector() = %q, want %q", shortSelector, tt.expectedShort)
			}

			if normalSelector != tt.expectedNormal {
				t.Errorf("generateOptimalSelector() = %q, want %q", normalSelector, tt.expectedNormal)
			}

			// Short selector should generally be shorter than or equal to normal selector
			if len(shortSelector) > len(normalSelector) {
				t.Errorf("Short selector %q is longer than normal selector %q", shortSelector, normalSelector)
			}
		})
	}
}

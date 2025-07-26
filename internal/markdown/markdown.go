// Package markdown provides utilities for parsing and querying markdown documents
package markdown

import (
	"fmt"
	"strings"
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

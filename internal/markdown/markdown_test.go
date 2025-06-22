package markdown

import (
	"strings"
	"testing"
)

func TestCalculateLineNumber(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		offset   int
		expected int
	}{
		{
			name:     "beginning of file",
			content:  "# Heading\nSome text",
			offset:   0,
			expected: 1,
		},
		{
			name:     "start of second line",
			content:  "# Heading\nSome text",
			offset:   10, // Right after first newline
			expected: 2,
		},
		{
			name:     "middle of second line",
			content:  "# Heading\nSome text",
			offset:   15,
			expected: 2,
		},
		{
			name:     "multiline content",
			content:  "Line 1\nLine 2\nLine 3\nLine 4",
			offset:   14, // Start of "Line 3"
			expected: 3,
		},
		{
			name:     "negative offset should be clamped to 0",
			content:  "Test content",
			offset:   -5,
			expected: 1,
		},
		{
			name:     "offset beyond content should be clamped",
			content:  "Test content",
			offset:   100,
			expected: 1,
		},
		{
			name:     "empty content",
			content:  "",
			offset:   0,
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateLineNumber([]byte(tt.content), tt.offset)
			if result != tt.expected {
				t.Errorf("CalculateLineNumber() = %d, expected %d", result, tt.expected)
			}
		})
	}
}

func TestCalculateLineColumn(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		offset         int
		expectedLine   int
		expectedColumn int
	}{
		{
			name:           "beginning of file",
			content:        "# Heading\nSome text",
			offset:         0,
			expectedLine:   1,
			expectedColumn: 1,
		},
		{
			name:           "middle of first line",
			content:        "# Heading\nSome text",
			offset:         5,
			expectedLine:   1,
			expectedColumn: 6,
		},
		{
			name:           "start of second line",
			content:        "# Heading\nSome text",
			offset:         10,
			expectedLine:   2,
			expectedColumn: 1,
		},
		{
			name:           "middle of second line",
			content:        "# Heading\nSome text",
			offset:         15,
			expectedLine:   2,
			expectedColumn: 6,
		},
		{
			name:           "negative offset",
			content:        "Test content",
			offset:         -5,
			expectedLine:   1,
			expectedColumn: 1,
		},
		{
			name:           "offset beyond content",
			content:        "Test",
			offset:         100,
			expectedLine:   1,
			expectedColumn: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line, column := CalculateLineColumn([]byte(tt.content), tt.offset)
			if line != tt.expectedLine || column != tt.expectedColumn {
				t.Errorf("CalculateLineColumn() = (%d, %d), expected (%d, %d)",
					line, column, tt.expectedLine, tt.expectedColumn)
			}
		})
	}
}

func TestValidateOffset(t *testing.T) {
	content := []byte("Hello, World!")

	tests := []struct {
		name     string
		offset   int
		expected int
	}{
		{
			name:     "valid offset",
			offset:   5,
			expected: 5,
		},
		{
			name:     "negative offset",
			offset:   -10,
			expected: 0,
		},
		{
			name:     "offset beyond content",
			offset:   100,
			expected: len(content),
		},
		{
			name:     "offset at end",
			offset:   len(content),
			expected: len(content),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateOffset(content, tt.offset)
			if result != tt.expected {
				t.Errorf("ValidateOffset() = %d, expected %d", result, tt.expected)
			}
		})
	}
}

func TestFindInsertionPoint(t *testing.T) {
	content := []byte("Line 1\nLine 2\nLine 3")

	tests := []struct {
		name     string
		offset   int
		prepend  bool
		expected int
	}{
		{
			name:     "append mode at beginning",
			offset:   0,
			prepend:  false,
			expected: len(content),
		},
		{
			name:     "append mode at middle",
			offset:   10,
			prepend:  false,
			expected: len(content),
		},
		{
			name:     "prepend mode at beginning of first line",
			offset:   0,
			prepend:  true,
			expected: 7, // After "Line 1\n"
		},
		{
			name:     "prepend mode at beginning of second line",
			offset:   7,
			prepend:  true,
			expected: 14, // After "Line 2\n"
		},
		{
			name:     "prepend mode at end of content",
			offset:   len(content),
			prepend:  true,
			expected: len(content),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindInsertionPoint(content, tt.offset, tt.prepend)
			if result != tt.expected {
				t.Errorf("FindInsertionPoint() = %d, expected %d", result, tt.expected)
			}
		})
	}
}

func TestOffsetRange(t *testing.T) {
	content := []byte("Hello, World!")

	t.Run("valid range", func(t *testing.T) {
		r := OffsetRange{Start: 0, End: 5}

		if !r.IsValid(content) {
			t.Error("Expected range to be valid")
		}

		if r.Length() != 5 {
			t.Errorf("Expected length 5, got %d", r.Length())
		}

		extracted := r.Extract(content)
		expected := "Hello"
		if string(extracted) != expected {
			t.Errorf("Expected %q, got %q", expected, string(extracted))
		}
	})

	t.Run("invalid range - negative start", func(t *testing.T) {
		r := OffsetRange{Start: -1, End: 5}

		if r.IsValid(content) {
			t.Error("Expected range to be invalid")
		}

		extracted := r.Extract(content)
		if extracted != nil {
			t.Error("Expected nil extraction from invalid range")
		}
	})

	t.Run("invalid range - end before start", func(t *testing.T) {
		r := OffsetRange{Start: 10, End: 5}

		if r.IsValid(content) {
			t.Error("Expected range to be invalid")
		}

		if r.Length() != 0 {
			t.Errorf("Expected length 0 for invalid range, got %d", r.Length())
		}
	})

	t.Run("invalid range - end beyond content", func(t *testing.T) {
		r := OffsetRange{Start: 0, End: 100}

		if r.IsValid(content) {
			t.Error("Expected range to be invalid")
		}
	})
}

func TestOffsetCalculationsWithRealMarkdown(t *testing.T) {
	// Test with realistic markdown content that might appear in jot
	content := `# Meeting Notes

## 2025-06-20 10:30

Discussed project timeline and deliverables.

### Action Items
- Review proposal by Friday
- Schedule follow-up meeting

## 2025-06-20 14:00

Team standup meeting.

### Completed
- Fixed bug #123
- Updated documentation

### In Progress  
- Feature implementation
- Code review
`

	// Test finding line numbers for various headings
	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{"main heading", "# Meeting Notes", 1},
		{"first timestamp", "## 2025-06-20 10:30", 3},
		{"action items", "### Action Items", 7},
		{"second timestamp", "## 2025-06-20 14:00", 11}, // Fixed: was 12
		{"completed section", "### Completed", 15},      // Fixed: was 16
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			offset := strings.Index(content, tt.text)
			if offset == -1 {
				t.Fatalf("Could not find text %q in content", tt.text)
			}

			line := CalculateLineNumber([]byte(content), offset)
			if line != tt.expected {
				t.Errorf("Expected line %d for %q, got %d", tt.expected, tt.text, line)
			}
		})
	}
}

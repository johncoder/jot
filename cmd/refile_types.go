// Package cmd: refile_types.go
// Types and struct definitions for refile command.
package cmd

// RefileResponse is the response structure for the refile operation.
type RefileResponse struct {
	Operation   string               `json:"operation"`
	Source      RefileSource         `json:"source"`
	Destination RefileDestination    `json:"destination"`
	Content     RefileContent        `json:"content"`
	Metadata    interface{}          `json:"metadata"`
}

// RefileSource represents the source details in a refile operation.
type RefileSource struct {
	Selector      string `json:"selector"`
	FilePath      string `json:"file_path"`
	Heading       string `json:"heading"`
	OriginalLevel int    `json:"original_level"`
}

// RefileDestination represents the destination details in a refile operation.
type RefileDestination struct {
	Selector        string   `json:"selector"`
	FilePath        string   `json:"file_path"`
	TargetLevel     int      `json:"target_level"`
	PathExists      bool     `json:"path_exists"`
	CreatedHeadings []string `json:"created_headings,omitempty"`
}

// RefileContent represents the content details in a refile operation.
type RefileContent struct {
	Content          string `json:"content"`
	CharacterCount   int    `json:"character_count"`
	LineCount        int    `json:"line_count"`
	TransformedLevel int    `json:"transformed_level"`
}

// InspectDestinationResponse is the response structure for inspecting a destination.
type InspectDestinationResponse struct {
	Operation   string                     `json:"operation"`
	Destination InspectDestinationInfo     `json:"destination"`
	Analysis    InspectDestinationAnalysis `json:"analysis"`
	Metadata    interface{}                `json:"metadata"`
}

// InspectDestinationInfo holds information about the destination being inspected.
type InspectDestinationInfo struct {
	Selector   string `json:"selector"`
	FilePath   string `json:"file_path"`
	FileExists bool   `json:"file_exists"`
}

// InspectDestinationAnalysis provides analysis details of the inspection.
type InspectDestinationAnalysis struct {
	PathExists      bool                     `json:"path_exists"`
	FoundSegments   []string                 `json:"found_segments"`
	MissingSegments []string                 `json:"missing_segments"`
	TargetLevel     int                      `json:"target_level"`
	WouldCreate     []InspectHeadingCreation `json:"would_create,omitempty"`
}

// InspectHeadingCreation represents a heading that would be created in the inspection.
type InspectHeadingCreation struct {
	Heading string `json:"heading"`
	Level   int    `json:"level"`
}

// DestinationTarget represents a resolved destination for refile operations.
type DestinationTarget struct {
	File         string   // Target file path
	TargetLevel  int      // Level where content should be inserted
	InsertOffset int      // Byte position for insertion
	CreatePath   []string // Missing headings to create
	Exists       bool     // Whether the target path exists
}

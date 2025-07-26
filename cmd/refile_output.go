package cmd

// Output and formatting helpers for refile command.

import (
	"fmt"
	"os"
	"strings"

	"github.com/johncoder/jot/internal/cmdutil"
	"github.com/johncoder/jot/internal/markdown"
	"github.com/johncoder/jot/internal/workspace"
)

// OutputRefileJSON outputs JSON response for refile operation
func OutputRefileJSON(ctx *cmdutil.CommandContext, sourcePath *markdown.HeadingPath, destPath *markdown.HeadingPath,
	subtree *markdown.Subtree, dest *DestinationTarget, transformedContent []byte) error {
	return outputRefileJSON(ctx, sourcePath, destPath, subtree, dest, transformedContent)
}

// InspectDestinationJSON outputs JSON response for destination inspection
func InspectDestinationJSON(ctx *cmdutil.CommandContext, ws *workspace.Workspace, destPath *markdown.HeadingPath) error {
	return inspectDestinationJSON(ctx, ws, destPath)
}

// PrintVerboseSubtreeInfo prints detailed information about the extracted subtree
func PrintVerboseSubtreeInfo(subtree *markdown.Subtree, filename string) {
	printVerboseSubtreeInfo(subtree, filename)
}

// PrintVerboseDestinationInfo prints detailed information about the destination
func PrintVerboseDestinationInfo(dest *DestinationTarget) {
	printVerboseDestinationInfo(dest)
}

// outputRefileJSON outputs JSON response for refile operation
func outputRefileJSON(ctx *cmdutil.CommandContext, sourcePath *markdown.HeadingPath, destPath *markdown.HeadingPath,
	subtree *markdown.Subtree, dest *DestinationTarget, transformedContent []byte) error {
	// Get source file path
	sourceFilePath := sourcePath.File
	if sourcePath.File == "inbox.md" {
		sourceFilePath = "inbox.md"
	}

	// Get destination file path
	destFilePath := dest.File

	// Count lines in transformed content
	lineCount := strings.Count(string(transformedContent), "\n") + 1
	if len(transformedContent) == 0 {
		lineCount = 0
	}

	// Build created headings list
	var createdHeadings []string
	if len(dest.CreatePath) > 0 {
		createdHeadings = dest.CreatePath
	}

	response := RefileResponse{
		Operation: "refile",
		Source: RefileSource{
			Selector:      sourcePath.File + "#" + strings.Join(sourcePath.Segments, "/"),
			FilePath:      sourceFilePath,
			Heading:       subtree.Heading,
			OriginalLevel: subtree.Level,
		},
		Destination: RefileDestination{
			Selector:        destPath.File + "#" + strings.Join(destPath.Segments, "/"),
			FilePath:        destFilePath,
			TargetLevel:     dest.TargetLevel,
			PathExists:      dest.Exists,
			CreatedHeadings: createdHeadings,
		},
		Content: RefileContent{
			Content:          string(transformedContent),
			CharacterCount:   len(transformedContent),
			LineCount:        lineCount,
			TransformedLevel: dest.TargetLevel,
		},
		Metadata: cmdutil.CreateJSONMetadata(ctx.Cmd, true, ctx.StartTime),
	}

	return outputJSON(response)
}

// inspectDestinationJSON outputs JSON response for destination inspection
func inspectDestinationJSON(ctx *cmdutil.CommandContext, ws *workspace.Workspace, destPath *markdown.HeadingPath) error {
	pathUtil := cmdutil.NewPathUtil(ws)
	// Check if file exists
	var filePath string
	if destPath.File == "inbox.md" {
		filePath = ws.InboxPath
	} else if strings.HasPrefix(destPath.File, "/") {
		filePath = destPath.File
	} else {
		filePath = pathUtil.WorkspaceJoin(destPath.File)
	}

	fileExists := true
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fileExists = false
	}

	response := InspectDestinationResponse{
		Operation: "inspect_destination",
		Destination: InspectDestinationInfo{
			Selector:   destPath.File + "#" + strings.Join(destPath.Segments, "/"),
			FilePath:   filePath,
			FileExists: fileExists,
		},
		Analysis: InspectDestinationAnalysis{
			PathExists:      false,
			FoundSegments:   []string{},
			MissingSegments: destPath.Segments,
			TargetLevel:     destPath.SkipLevels + len(destPath.Segments) + 1,
		},
		Metadata: cmdutil.CreateJSONMetadata(ctx.Cmd, true, ctx.StartTime),
	}

	// If file doesn't exist, return early with basic analysis
	if !fileExists {
		// Fill in what would be created
		baseLevel := destPath.SkipLevels + 1
		for i, heading := range destPath.Segments {
			level := baseLevel + i
			response.Analysis.WouldCreate = append(response.Analysis.WouldCreate, InspectHeadingCreation{
				Heading: heading,
				Level:   level,
			})
		}
		return outputJSON(response)
	}

	// Read and parse the file to analyze the path
	content, err := os.ReadFile(filePath)
	if err != nil {
		// Return error as JSON
		return ctx.HandleError(fmt.Errorf("error reading file: %w", err))
	}

	doc := markdown.ParseMarkdownAST(content)
	pathResolution, err := markdown.NavigateHeadingPath(doc, destPath)
	if err != nil {
		return ctx.HandleError(fmt.Errorf("error analyzing path: %w", err))
	}

	if pathResolution.TargetHeading != nil {
		// Complete path exists
		response.Analysis.PathExists = true
		response.Analysis.FoundSegments = destPath.Segments
		response.Analysis.MissingSegments = []string{}
		response.Analysis.TargetLevel = pathResolution.TargetHeading.Level + 1
	} else if len(pathResolution.FoundSegments) > 0 {
		// Partial path exists
		response.Analysis.PathExists = false
		response.Analysis.FoundSegments = pathResolution.FoundSegments
		response.Analysis.MissingSegments = pathResolution.MissingSegments

		// Create what would be created
		baseLevel := pathResolution.ParentHeading.Level + 1
		for i, heading := range pathResolution.MissingSegments {
			level := baseLevel + i
			response.Analysis.WouldCreate = append(response.Analysis.WouldCreate, InspectHeadingCreation{
				Heading: heading,
				Level:   level,
			})
		}
		response.Analysis.TargetLevel = baseLevel + len(pathResolution.MissingSegments)
	} else {
		// No path exists
		response.Analysis.PathExists = false
		response.Analysis.FoundSegments = []string{}
		response.Analysis.MissingSegments = destPath.Segments

		// Create what would be created
		baseLevel := destPath.SkipLevels + 1
		for i, heading := range destPath.Segments {
			level := baseLevel + i
			response.Analysis.WouldCreate = append(response.Analysis.WouldCreate, InspectHeadingCreation{
				Heading: heading,
				Level:   level,
			})
		}
		response.Analysis.TargetLevel = baseLevel + len(destPath.Segments)
	}

	return outputJSON(response)
}

// printVerboseSubtreeInfo prints detailed information about the extracted subtree
func printVerboseSubtreeInfo(subtree *markdown.Subtree, filename string) {
	fmt.Printf("Source subtree analysis:\n")
	fmt.Printf("  File: %s\n", filename)
	fmt.Printf("  Heading: %q\n", subtree.Heading)
	fmt.Printf("  Level: %d\n", subtree.Level)
	fmt.Printf("  Start offset: %d\n", subtree.StartOffset)
	fmt.Printf("  End offset: %d\n", subtree.EndOffset)
	fmt.Printf("  Total length: %d bytes\n", len(subtree.Content))

	// Show head and tail summary
	content := subtree.Content
	if len(content) > 100 {
		head := strings.ReplaceAll(string(content[:50]), "\n", "\\n")
		tail := strings.ReplaceAll(string(content[len(content)-50:]), "\n", "\\n")
		fmt.Printf("  Content preview: %q ... %q\n", head, tail)
	} else {
		preview := strings.ReplaceAll(string(content), "\n", "\\n")
		fmt.Printf("  Content: %q\n", preview)
	}
	fmt.Println()
}

// printVerboseDestinationInfo prints detailed information about the destination
func printVerboseDestinationInfo(dest *DestinationTarget) {
	fmt.Printf("Destination analysis:\n")
	fmt.Printf("  File: %s\n", dest.File)
	fmt.Printf("  Target level: %d\n", dest.TargetLevel)
	fmt.Printf("  Insert offset: %d\n", dest.InsertOffset)
	fmt.Printf("  Path exists: %t\n", dest.Exists)
	if len(dest.CreatePath) > 0 {
		fmt.Printf("  Will create path: %s\n", strings.Join(dest.CreatePath, " > "))
	}
	fmt.Println()
}

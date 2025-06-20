package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/johncoder/jot/internal/markdown"
	"github.com/johncoder/jot/internal/workspace"
	"github.com/spf13/cobra"
	"github.com/yuin/goldmark/ast"
)

// DestinationTarget represents a resolved destination
type DestinationTarget struct {
	File         string   // Target file path
	TargetLevel  int      // Level where content should be inserted
	InsertOffset int      // Byte position for insertion
	CreatePath   []string // Missing headings to create
	Exists       bool     // Whether the target path exists
}

var refileCmd = &cobra.Command{
	Use:   "refile [SOURCE] --to DESTINATION",
	Short: "Move markdown subtrees between files using path-based selectors",
	Long: `Move entire markdown subtrees (headings with all nested content) between files.

Path-based selector syntax with contains matching:
- Each segment uses case-insensitive contains matching
- Must match exactly one subtree
- Leading slashes handle unusual document structures

Examples:
  jot refile "inbox.md#meeting" --to "work.md#projects"
  jot refile "notes.md#research/database" --to "archive.md#technical"  
  jot refile "inbox.md#/foo/bar" --to "work.md#tasks"  # Skip level 1
  jot refile --to "work.md#projects/frontend"          # Inspect destination`,

	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := workspace.RequireWorkspace()
		if err != nil {
			return err
		}

		// Get flags
		to, _ := cmd.Flags().GetString("to")
		prepend, _ := cmd.Flags().GetBool("prepend")
		verbose, _ := cmd.Flags().GetBool("verbose")

		// No source and no destination: show usage help
		if len(args) == 0 && to == "" {
			return fmt.Errorf("provide a source file or --to destination")
		}

		if to == "" {
			// Check if this is a request to show selectors for a specific file
			if len(args) == 1 && !strings.Contains(args[0], "#") {
				return showSelectorsForFile(ws, args[0])
			}
			return fmt.Errorf("destination path required: use --to flag")
		}

		// Parse destination path
		destPath, err := markdown.ParsePath(to)
		if err != nil {
			return fmt.Errorf("invalid destination path '%s': %w", to, err)
		}

		// Source-less mode: inspect destination
		if len(args) == 0 {
			return inspectDestination(ws, destPath)
		}

		// Parse source path
		sourcePath, err := markdown.ParsePath(args[0])
		if err != nil {
			return fmt.Errorf("invalid source path '%s': %w", args[0], err)
		}

		// Extract subtree from source
		subtree, err := ExtractSubtree(ws, sourcePath)
		if err != nil {
			return fmt.Errorf("failed to extract subtree: %w", err)
		}

		if verbose {
			printVerboseSubtreeInfo(subtree, sourcePath.File)
		}

		// Resolve destination
		dest, err := ResolveDestination(ws, destPath, prepend)
		if err != nil {
			return fmt.Errorf("failed to resolve destination: %w", err)
		}

		if verbose {
			printVerboseDestinationInfo(dest)
		}

		// Transform subtree level
		transformedContent := TransformSubtreeLevel(subtree, dest.TargetLevel)

		// Perform the refile operation
		if err := performRefile(ws, sourcePath, subtree, dest, transformedContent); err != nil {
			return fmt.Errorf("refile operation failed: %w", err)
		}

		if verbose {
			fmt.Printf("Refile operation completed successfully!\n")
		}

		fmt.Printf("Successfully refiled '%s' to '%s'\n", 
			subtree.Heading, destPath.File+"#"+strings.Join(destPath.Segments, "/"))

		return nil
	},
}

// inspectDestination analyzes destination path without performing refile
func inspectDestination(ws *workspace.Workspace, destPath *markdown.HeadingPath) error {
	fmt.Printf("Destination analysis for \"%s#%s\":\n", 
		destPath.File, strings.Join(destPath.Segments, "/"))

	// Check if file exists
	filePath := filepath.Join(ws.LibDir, destPath.File)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Printf("✗ File not found: %s\n", destPath.File)
		return nil
	}
	fmt.Printf("✓ File exists: %s\n", destPath.File)

	// Resolve destination to check path
	dest, err := ResolveDestination(ws, destPath, false)
	if err != nil {
		// Check if it's an ambiguous path error
		if strings.Contains(err.Error(), "matches multiple headings") {
			fmt.Printf("✗ Ambiguous path: %s\n", err.Error())
			return nil
		}
		fmt.Printf("✗ Error: %s\n", err.Error())
		return nil
	}

	if dest.Exists {
		fmt.Printf("✓ Path exists: %s\n", strings.Join(destPath.Segments, " > "))
		fmt.Printf("Ready to receive content at level %d\n", dest.TargetLevel+1)
	} else {
		if len(dest.CreatePath) > 0 {
			fmt.Printf("✗ Missing path: %s\n", strings.Join(dest.CreatePath, " > "))
			for i, heading := range dest.CreatePath {
				level := dest.TargetLevel - len(dest.CreatePath) + i + 1
				fmt.Printf("Would create: %s %s (level %d)\n", 
					strings.Repeat("#", level), heading, level)
			}
		}
		fmt.Printf("Ready to receive content at level %d\n", dest.TargetLevel+1)
	}

	return nil
}

// ExtractSubtree extracts a subtree from the source file
func ExtractSubtree(ws *workspace.Workspace, sourcePath *markdown.HeadingPath) (*markdown.Subtree, error) {
	// Construct full file path
	var filePath string
	if sourcePath.File == "inbox.md" {
		filePath = ws.InboxPath
	} else {
		filePath = filepath.Join(ws.LibDir, sourcePath.File)
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", sourcePath.File, err)
	}

	// Parse document and find subtree
	doc := markdown.ParseDocument(content)
	subtree, err := markdown.FindSubtree(doc, content, sourcePath)
	if err != nil {
		return nil, err
	}

	return subtree, nil
}

// ResolveDestination resolves a destination path and determines insertion point
func ResolveDestination(ws *workspace.Workspace, destPath *markdown.HeadingPath, prepend bool) (*DestinationTarget, error) {
	// Construct full file path
	var filePath string
	if destPath.File == "inbox.md" {
		filePath = ws.InboxPath
	} else {
		filePath = filepath.Join(ws.LibDir, destPath.File)
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("destination file not found: %s", destPath.File)
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read destination file: %w", err)
	}

	// Parse document
	doc := markdown.ParseDocument(content)

	// Find or create the destination path
	return resolveDestinationPath(doc, content, destPath, prepend)
}

// resolveDestinationPath finds the target location for insertion
func resolveDestinationPath(doc ast.Node, content []byte, destPath *markdown.HeadingPath, prepend bool) (*DestinationTarget, error) {
	// For now, implement simple case: create at end of file
	// This is a simplified implementation - full version would search for existing paths
	
	insertOffset := len(content)
	if insertOffset > 0 && content[insertOffset-1] != '\n' {
		insertOffset-- // Insert before final newline if exists
	}
	
	targetLevel := len(destPath.Segments) + destPath.SkipLevels
	
	return &DestinationTarget{
		File:         destPath.File,
		TargetLevel:  targetLevel,
		InsertOffset: insertOffset,
		CreatePath:   destPath.Segments,
		Exists:       false,
	}, nil
}

// TransformSubtreeLevel adjusts heading levels in subtree content
func TransformSubtreeLevel(subtree *markdown.Subtree, newBaseLevel int) []byte {
	levelDiff := newBaseLevel - subtree.Level
	return markdown.TransformHeadingLevels(subtree.Content, levelDiff)
}

// performRefile executes the actual refile operation
func performRefile(ws *workspace.Workspace, sourcePath *markdown.HeadingPath, subtree *markdown.Subtree, dest *DestinationTarget, transformedContent []byte) error {
	// Read source file
	var sourceFilePath string
	if sourcePath.File == "inbox.md" {
		sourceFilePath = ws.InboxPath
	} else {
		sourceFilePath = filepath.Join(ws.LibDir, sourcePath.File)
	}
	
	sourceContent, err := os.ReadFile(sourceFilePath)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	// Read destination file
	var destFilePath string
	if dest.File == "inbox.md" {
		destFilePath = ws.InboxPath
	} else {
		destFilePath = filepath.Join(ws.LibDir, dest.File)
	}
	
	destContent, err := os.ReadFile(destFilePath)
	if err != nil {
		return fmt.Errorf("failed to read destination file: %w", err)
	}

	// Remove from source
	newSourceContent := append(sourceContent[:subtree.StartOffset], sourceContent[subtree.EndOffset:]...)
	
	// Insert into destination
	var newDestContent []byte
	insertContent := transformedContent
	
	// Add missing headings if needed
	if len(dest.CreatePath) > 0 {
		var pathContent []byte
		baseLevel := dest.TargetLevel - len(dest.CreatePath) + 1
		for i, heading := range dest.CreatePath {
			level := baseLevel + i
			if i > 0 || len(destContent) > 0 {
				pathContent = append(pathContent, '\n')
			}
			levelMarker := bytes.Repeat([]byte("#"), level)
			pathContent = append(pathContent, levelMarker...)
			pathContent = append(pathContent, ' ')
			pathContent = append(pathContent, []byte(heading)...)
			pathContent = append(pathContent, '\n')
		}
		insertContent = append(pathContent, insertContent...)
	}
	
	// Insert at the specified offset
	newDestContent = append(destContent[:dest.InsertOffset], insertContent...)
	newDestContent = append(newDestContent, destContent[dest.InsertOffset:]...)

	// Write files back
	if err := os.WriteFile(sourceFilePath, newSourceContent, 0644); err != nil {
		return fmt.Errorf("failed to write source file: %w", err)
	}
	
	if err := os.WriteFile(destFilePath, newDestContent, 0644); err != nil {
		return fmt.Errorf("failed to write destination file: %w", err)
	}

	return nil
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

func init() {
	refileCmd.Flags().String("to", "", "Destination path (e.g., 'work.md#projects/frontend')")
	refileCmd.Flags().Bool("prepend", false, "Insert content at the beginning under target heading")
	refileCmd.Flags().BoolP("verbose", "v", false, "Show detailed information about the refile operation")
}

// showSelectorsForFile displays available selectors for a specific file
func showSelectorsForFile(ws *workspace.Workspace, filename string) error {
	// Determine the full file path
	var filePath string
	if filename == "inbox.md" {
		filePath = ws.InboxPath
	} else {
		// Handle files in lib/ directory
		if !strings.HasSuffix(filename, ".md") {
			filename += ".md"
		}
		filePath = filepath.Join(ws.LibDir, filename)
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filename)
	}

	// Read and parse the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	if len(content) == 0 {
		fmt.Printf("File %s is empty (no selectors available)\n", filename)
		return nil
	}

	// Parse document and get headings
	doc := markdown.ParseDocument(content)
	headings := markdown.FindAllHeadings(doc, content)

	if len(headings) == 0 {
		fmt.Printf("No headings found in %s\n", filename)
		return nil
	}

	// Display selectors
	fmt.Printf("Available selectors in %s:\n", filename)
	
	for _, heading := range headings {
		if strings.TrimSpace(heading.Text) == "" {
			continue // Skip empty headings
		}

		// Build selector based on heading level and path
		var selector string
		if heading.Level == 1 {
			selector = fmt.Sprintf("%s#%s", filename, strings.ToLower(heading.Text))
		} else {
			// For level 2+ headings, use skip syntax
			skipPrefix := strings.Repeat("/", heading.Level-1)
			selector = fmt.Sprintf("%s#%s%s", filename, skipPrefix, strings.ToLower(heading.Text))
		}

		fmt.Printf("  %s\n", selector)
	}

	fmt.Printf("\nUsage: jot refile \"<selector>\" --to \"<destination>\"\n")
	return nil
}

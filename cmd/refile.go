package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/johncoder/jot/internal/cmdutil"
	"github.com/johncoder/jot/internal/fzf"
	"github.com/johncoder/jot/internal/hooks"
	"github.com/johncoder/jot/internal/markdown"
	"github.com/johncoder/jot/internal/workspace"
	"github.com/spf13/cobra"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// DestinationTarget represents a resolved destination
type DestinationTarget struct {
	File         string   // Target file path
	TargetLevel  int      // Level where content should be inserted
	InsertOffset int      // Byte position for insertion
	CreatePath   []string // Missing headings to create
	Exists       bool     // Whether the target path exists
}

// RefileOperation encapsulates a refile operation with atomic execution for same-file operations
type RefileOperation struct {
	SourcePath       string
	DestPath         string
	Subtree          *markdown.Subtree
	TransformedContent []byte
	InsertOffset     int
	CreatePath       []string
	TargetLevel      int
}

// IsSameFile returns true if source and destination are the same file
func (op *RefileOperation) IsSameFile() bool {
	return op.SourcePath == op.DestPath
}

// Execute performs the refile operation with proper same-file handling
func (op *RefileOperation) Execute() error {
	if op.IsSameFile() {
		return op.executeSameFile()
	}
	return op.executeCrossFile()
}

// executeSameFile handles same-file refile using simple, safe text manipulation
func (op *RefileOperation) executeSameFile() error {
	// Read the file content using unified content utilities
	content, err := cmdutil.ReadFileContent(op.SourcePath)
	if err != nil {
		return err
	}

	// Perform simple same-file refile
	newContent := op.performSimpleSameFileRefile(content)

	// Write the modified content back to file using unified content utilities
	return cmdutil.WriteFileContent(op.SourcePath, newContent)
}

// performSimpleSameFileRefile performs safe same-file refile with consistent formatting
func (op *RefileOperation) performSimpleSameFileRefile(content []byte) []byte {
	// Step 1: Prepare content to move with consistent formatting
	contentToMove := op.ensureConsistentFormatting(op.TransformedContent)
	
	// Step 2: Remove the original subtree cleanly
	beforeSubtree := content[:op.Subtree.StartOffset]
	afterSubtree := content[op.Subtree.EndOffset:]
	contentWithoutSubtree := append(beforeSubtree, afterSubtree...)
	
	// Step 3: Adjust insertion offset for removed content
	adjustedOffset := op.InsertOffset
	if op.InsertOffset > op.Subtree.StartOffset {
		removedLength := op.Subtree.EndOffset - op.Subtree.StartOffset
		adjustedOffset = op.InsertOffset - removedLength
	}
	
	// Step 4: Ensure we don't go past the content boundary
	if adjustedOffset > len(contentWithoutSubtree) {
		adjustedOffset = len(contentWithoutSubtree)
	}
	
	// Step 5: Perform insertion and normalize spacing in post-processing
	result := make([]byte, 0, len(contentWithoutSubtree)+len(contentToMove)+2)
	result = append(result, contentWithoutSubtree[:adjustedOffset]...)
	
	// Add spacing before content
	if adjustedOffset > 0 && contentWithoutSubtree[adjustedOffset-1] != '\n' {
		result = append(result, '\n', '\n')
	} else if adjustedOffset > 0 {
		result = append(result, '\n')
	}
	
	// Add content
	result = append(result, contentToMove...)
	
	// Add remaining content
	result = append(result, contentWithoutSubtree[adjustedOffset:]...)
	
	// Post-process to normalize spacing: ensure exactly one blank line between sections
	return op.normalizeMarkdownSpacing(result)
}

// executeCrossFile handles cross-file refile operations
func (op *RefileOperation) executeCrossFile() error {
	// Step 1: Read and update source file using unified content utilities
	sourceContent, err := cmdutil.ReadFileContent(op.SourcePath)
	if err != nil {
		return err
	}

	newSourceContent := append(sourceContent[:op.Subtree.StartOffset], sourceContent[op.Subtree.EndOffset:]...)
	if err := cmdutil.WriteFileContent(op.SourcePath, newSourceContent); err != nil {
		return err
	}

	// Step 2: Read and update destination file using unified content utilities
	destContent, err := cmdutil.ReadFileContent(op.DestPath)
	if err != nil {
		return err
	}

	insertContent := op.prepareInsertContent(destContent, op.InsertOffset)
	newDestContent := append(destContent[:op.InsertOffset], insertContent...)
	newDestContent = append(newDestContent, destContent[op.InsertOffset:]...)

	return cmdutil.WriteFileContent(op.DestPath, newDestContent)
}

// performInMemoryRefile performs the entire refile operation in memory for same-file operations
func (op *RefileOperation) performInMemoryRefile(content []byte) []byte {
	// Step 1: Remove the subtree from its original location
	contentWithoutSubtree := append(content[:op.Subtree.StartOffset], content[op.Subtree.EndOffset:]...)

	// Step 2: Adjust insertion offset if it's after the removed content
	adjustedOffset := op.InsertOffset
	if op.InsertOffset > op.Subtree.StartOffset {
		adjustedOffset = op.InsertOffset - (op.Subtree.EndOffset - op.Subtree.StartOffset)
	}

	// Step 3: Prepare insertion content
	insertContent := op.prepareInsertContent(contentWithoutSubtree, adjustedOffset)

	// Step 4: Insert content at the adjusted offset
	result := append(contentWithoutSubtree[:adjustedOffset], insertContent...)
	result = append(result, contentWithoutSubtree[adjustedOffset:]...)

	return result
}

// prepareInsertContent prepares the content to be inserted, including missing headings and spacing
func (op *RefileOperation) prepareInsertContent(destContent []byte, insertOffset int) []byte {
	// Ensure consistent formatting for the content being inserted
	insertContent := op.ensureConsistentFormatting(op.TransformedContent)

	// Add missing headings if needed
	if len(op.CreatePath) > 0 {
		baseLevel := op.TargetLevel - len(op.CreatePath)
		pathContent := markdown.CreateHeadingStructure(op.CreatePath, baseLevel)

		// Ensure proper spacing before path content
		if insertOffset > 0 && destContent[insertOffset-1] != '\n' {
			pathContent = append([]byte("\n\n"), pathContent...)
		} else if insertOffset > 0 {
			pathContent = append([]byte("\n"), pathContent...)
		}

		insertContent = append(pathContent, insertContent...)
	} else {
		// Add proper spacing for standalone content insertion
		if insertOffset > 0 {
			prevChar := destContent[insertOffset-1]
			if prevChar != '\n' {
				insertContent = append([]byte("\n\n"), insertContent...)
			} else {
				insertContent = append([]byte("\n"), insertContent...)
			}
		}
	}

	// Ensure spacing after the content if there's content following
	if insertOffset < len(destContent) {
		insertContent = append(insertContent, '\n')
	}

	return insertContent
}

// PathResolution represents the result of path navigation
type PathResolution struct {
	TargetHeading   *ast.Heading // The final target heading if found
	ParentHeading   *ast.Heading // The deepest parent heading found
	FoundSegments   []string     // Successfully matched segments
	MissingSegments []string     // Segments that need to be created
}

var refileNoVerify bool

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
		ctx := cmdutil.StartCommand(cmd)

		ws, err := workspace.RequireWorkspace()
		if err != nil {
			return ctx.HandleError(err)
		}

		// Get flags
		to, _ := cmd.Flags().GetString("to")
		prepend, _ := cmd.Flags().GetBool("prepend")
		verbose, _ := cmd.Flags().GetBool("verbose")
		interactive, _ := cmd.Flags().GetBool("interactive")

		// Check for interactive mode
		if fzf.ShouldUseFZF(interactive) {
			return runInteractiveRefile(ctx, args, ws)
		}

		// No source and no destination: show usage help
		if len(args) == 0 && to == "" {
			err := fmt.Errorf("provide a source file or --to destination")
			return ctx.HandleError(err)
		}

		if to == "" {
			// Check if this is a request to show selectors for a specific file
			if len(args) == 1 && !strings.Contains(args[0], "#") {
				return showSelectorsForFile(ws, args[0])
			}
			err := fmt.Errorf("destination path required: use --to flag")
			if ctx.IsJSONOutput() {
				return ctx.HandleError(err)
			}
			return err
		}

		// Parse destination path
		destPath, err := markdown.ParsePath(to)
		if err != nil {
			err := cmdutil.NewValidationError("destination path", to, err)
			if ctx.IsJSONOutput() {
				return ctx.HandleError(err)
			}
			return err
		}

		// Source-less mode: inspect destination
		if len(args) == 0 {
			if ctx.IsJSONOutput() {
				return inspectDestinationJSON(ctx, ws, destPath)
			}
			return inspectDestination(ws, destPath)
		}

		// Parse source path
		sourcePath, err := markdown.ParsePath(args[0])
		if err != nil {
			err := cmdutil.NewValidationError("source path", args[0], err)
			if ctx.IsJSONOutput() {
				return ctx.HandleError(err)
			}
			return err
		}

		// Extract subtree from source
		subtree, err := ExtractSubtree(ws, sourcePath)
		if err != nil {
			err := fmt.Errorf("failed to extract subtree: %w", err)
			if ctx.IsJSONOutput() {
				return ctx.HandleError(err)
			}
			return err
		}

		if verbose && !ctx.IsJSONOutput() {
			printVerboseSubtreeInfo(subtree, sourcePath.File)
		}

		// Resolve destination
		dest, err := ResolveDestination(ws, destPath, prepend)
		if err != nil {
			err := fmt.Errorf("failed to resolve destination: %w", err)
			if ctx.IsJSONOutput() {
				return ctx.HandleError(err)
			}
			return err
		}

		if verbose && !ctx.IsJSONOutput() {
			printVerboseDestinationInfo(dest)
		}

		// Transform subtree level
		transformedContent := TransformSubtreeLevel(subtree, dest.TargetLevel)

		// Run pre-refile hook
		hookManager := hooks.NewManager(ws)
		if !refileNoVerify {
			hookCtx := &hooks.HookContext{
				Type:        hooks.PreRefile,
				Workspace:   ws,
				SourceFile:  args[0],
				DestPath:    to,
				Timeout:     30 * time.Second,
				AllowBypass: refileNoVerify,
			}
			
			result, err := hookManager.Execute(hookCtx)
			if err != nil {
				err := cmdutil.NewExternalError("pre-refile hook", nil, err)
				if ctx.IsJSONOutput() {
					return ctx.HandleError(err)
				}
				return err
			}
			
			if result.Aborted {
				err := fmt.Errorf("pre-refile hook aborted operation")
				if ctx.IsJSONOutput() {
					return ctx.HandleError(err)
				}
				return err
			}
		}

		// Perform the refile operation
		if err := performRefile(ws, sourcePath, subtree, dest, transformedContent); err != nil {
			err := fmt.Errorf("refile operation failed: %w", err)
			if ctx.IsJSONOutput() {
				return ctx.HandleError(err)
			}
			return err
		}

		// Run post-refile hook (informational only)
		if !refileNoVerify {
			hookCtx := &hooks.HookContext{
				Type:        hooks.PostRefile,
				Workspace:   ws,
				SourceFile:  args[0],
				DestPath:    to,
				Timeout:     30 * time.Second,
				AllowBypass: refileNoVerify,
			}
			
			_, hookErr := hookManager.Execute(hookCtx)
			if hookErr != nil && !ctx.IsJSONOutput() {
				fmt.Printf("Warning: post-refile hook failed: %s\n", hookErr.Error())
			}
		}

		// Handle JSON output
		if ctx.IsJSONOutput() {
			return outputRefileJSON(ctx, sourcePath, destPath, subtree, dest, transformedContent)
		}

		// Human-readable output
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
	pathUtil := cmdutil.NewPathUtil(ws)
	fmt.Printf("Destination analysis for \"%s#%s\":\n",
		destPath.File, strings.Join(destPath.Segments, "/"))

	// Check if file exists
	var filePath string
	if destPath.File == "inbox.md" {
		filePath = ws.InboxPath
	} else if filepath.IsAbs(destPath.File) {
		filePath = destPath.File
	} else {
		// Use workspace root for relative paths, not lib/ directory
		filePath = pathUtil.WorkspaceJoin(destPath.File)
	}

	if _, err := os.Stat(filePath); err != nil {
		if cmdutil.IsFileNotFound(err) {
			cmdutil.ShowError("✗ File not found: %s", destPath.File)
			return nil
		}
		cmdutil.ShowError("✗ Error accessing file: %s", err.Error())
		return nil
	}
	cmdutil.ShowSuccess("✓ File exists: %s", destPath.File)

	// Read and parse the file to analyze the path
	content, err := os.ReadFile(filePath)
	if err != nil {
		// Use structured error inspection for better error handling
		if fileErr, ok := cmdutil.GetFileError(err); ok {
			cmdutil.ShowError("✗ Error reading file %s: %s", fileErr.Path, fileErr.Err.Error())
		} else {
			cmdutil.ShowError("✗ Error reading file: %s", err.Error())
		}
		return nil
	}

	doc := markdown.ParseDocument(content)
	pathResolution, err := navigateHeadingPath(doc, content, destPath)
	if err != nil {
		fmt.Printf("✗ Error analyzing path: %s\n", err.Error())
		return nil
	}

	if pathResolution.TargetHeading != nil {
		// Complete path exists
		fmt.Printf("✓ Path exists: %s\n", strings.Join(destPath.Segments, " > "))
		targetLevel := pathResolution.TargetHeading.Level + 1
		fmt.Printf("Ready to receive content at level %d\n", targetLevel)
	} else if len(pathResolution.FoundSegments) > 0 {
		// Partial path exists
		cmdutil.ShowSuccess("✓ Partial path exists: %s", strings.Join(pathResolution.FoundSegments, " > "))
		cmdutil.ShowError("✗ Missing path: %s", strings.Join(pathResolution.MissingSegments, " > "))

		// Show what would be created
		baseLevel := pathResolution.ParentHeading.Level + 1
		for i, heading := range pathResolution.MissingSegments {
			level := baseLevel + i
			fmt.Printf("Would create: %s %s (level %d)\n",
				strings.Repeat("#", level), heading, level)
		}
		finalLevel := baseLevel + len(pathResolution.MissingSegments)
		fmt.Printf("Ready to receive content at level %d\n", finalLevel)
	} else {
		// No path exists
		fmt.Printf("✗ Missing path: %s\n", strings.Join(destPath.Segments, " > "))

		// Show what would be created
		baseLevel := destPath.SkipLevels + 1
		for i, heading := range destPath.Segments {
			level := baseLevel + i
			fmt.Printf("Would create: %s %s (level %d)\n",
				strings.Repeat("#", level), heading, level)
		}
		finalLevel := baseLevel + len(destPath.Segments)
		fmt.Printf("Ready to receive content at level %d\n", finalLevel)
	}

	return nil
}

// ExtractSubtree extracts a subtree from the source file
func ExtractSubtree(ws *workspace.Workspace, sourcePath *markdown.HeadingPath) (*markdown.Subtree, error) {
	return ExtractSubtreeWithOptions(ws, sourcePath, false)
}

// ExtractSubtreeWithOptions extracts a subtree with optional no-workspace mode
func ExtractSubtreeWithOptions(ws *workspace.Workspace, sourcePath *markdown.HeadingPath, noWorkspace bool) (*markdown.Subtree, error) {
	// Construct full file path using the shared resolution logic
	filePath := cmdutil.ResolvePath(ws, sourcePath.File, noWorkspace)

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, cmdutil.NewFileError("read", sourcePath.File, err)
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
	pathUtil := cmdutil.NewPathUtil(ws)
	// Construct full file path
	var filePath string
	if destPath.File == "inbox.md" {
		filePath = ws.InboxPath
	} else if filepath.IsAbs(destPath.File) {
		filePath = destPath.File
	} else {
		// Use workspace root for relative paths, not lib/ directory
		filePath = pathUtil.WorkspaceJoin(destPath.File)
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
	// Try to find existing path in the document
	pathResolution, err := navigateHeadingPath(doc, content, destPath)
	if err != nil {
		return nil, err
	}

	var insertOffset int
	var targetLevel int

	if pathResolution.TargetHeading != nil {
		// Found existing target heading - insert content under it
		insertOffset = calculateInsertionPoint(pathResolution.TargetHeading, content, prepend)
		targetLevel = pathResolution.TargetHeading.Level + 1
	} else {
		// Need to create missing path
		if pathResolution.ParentHeading != nil {
			// Insert under the deepest found parent
			insertOffset = calculateInsertionPoint(pathResolution.ParentHeading, content, false)
			targetLevel = pathResolution.ParentHeading.Level + len(pathResolution.MissingSegments) + 1
		} else {
			// No parent found, append to end of file
			insertOffset = len(content)
			if insertOffset > 0 && content[insertOffset-1] != '\n' {
				insertOffset = len(content)
			}
			// For top-level insertion with empty segments, default to level 2
			if len(destPath.Segments) == 0 {
				targetLevel = 2 // Default level for top-level insertion
			} else {
				targetLevel = destPath.SkipLevels + len(destPath.Segments)
			}
		}
	}

	return &DestinationTarget{
		File:         destPath.File,
		TargetLevel:  targetLevel,
		InsertOffset: insertOffset,
		CreatePath:   pathResolution.MissingSegments,
		Exists:       pathResolution.TargetHeading != nil,
	}, nil
}

// TransformSubtreeLevel adjusts heading levels in subtree content
func TransformSubtreeLevel(subtree *markdown.Subtree, newBaseLevel int) []byte {
	levelDiff := newBaseLevel - subtree.Level
	return markdown.TransformHeadingLevels(subtree.Content, levelDiff)
}

// performRefile executes the actual refile operation
// performRefile executes the actual refile operation using RefileOperation for atomic same-file handling
func performRefile(ws *workspace.Workspace, sourcePath *markdown.HeadingPath, subtree *markdown.Subtree, dest *DestinationTarget, transformedContent []byte) error {
	// Create a RefileOperation with all necessary data
	operation := &RefileOperation{
		SourcePath:         cmdutil.ResolveWorkspaceRelativePath(ws, sourcePath.File),
		DestPath:           cmdutil.ResolveWorkspaceRelativePath(ws, dest.File),
		Subtree:           subtree,
		TransformedContent: transformedContent,
		InsertOffset:      dest.InsertOffset,
		CreatePath:        dest.CreatePath,
		TargetLevel:       dest.TargetLevel,
	}

	// Execute the operation with proper same-file handling
	return operation.Execute()
}

// executeRefile executes the refile operation using existing logic
func executeRefile(sourceSelector, targetSelector string, ctx *cmdutil.CommandContext, ws *workspace.Workspace) error {
	// Initialize hook manager
	hookManager := hooks.NewManager(ws)
	
	// Run pre-refile hook
	if !refileNoVerify {
		hookCtx := &hooks.HookContext{
			Type:        hooks.PreRefile,
			Workspace:   ws,
			SourceFile:  sourceSelector,
			DestPath:    targetSelector,
			Timeout:     30 * time.Second,
			AllowBypass: refileNoVerify,
		}
		
		result, err := hookManager.Execute(hookCtx)
		if err != nil {
			return cmdutil.NewExternalError("pre-refile hook", nil, err)
		}
		
		if result.Aborted {
			return fmt.Errorf("pre-refile hook aborted operation")
		}
	}

	// Parse paths
	sourcePath, err := markdown.ParsePath(sourceSelector)
	if err != nil {
		return cmdutil.NewValidationError("source selector", sourceSelector, err)
	}

	destPath, err := markdown.ParsePath(targetSelector)
	if err != nil {
		return fmt.Errorf("invalid target selector '%s': %w", targetSelector, err)
	}

	// Extract subtree from source
	subtree, err := ExtractSubtree(ws, sourcePath)
	if err != nil {
		return fmt.Errorf("failed to extract subtree: %w", err)
	}

	// Get flags
	prepend, _ := ctx.Cmd.Flags().GetBool("prepend")
	verbose, _ := ctx.Cmd.Flags().GetBool("verbose")

	// Resolve destination
	destTarget, err := ResolveDestination(ws, destPath, prepend)
	if err != nil {
		return fmt.Errorf("failed to resolve destination: %w", err)
	}

	// Transform subtree level
	transformedContent := TransformSubtreeLevel(subtree, destTarget.TargetLevel)

	// Perform the refile operation using existing logic
	err = performRefile(ws, sourcePath, subtree, destTarget, transformedContent)
	if err != nil {
		return fmt.Errorf("refile operation failed: %w", err)
	}

	// Run post-refile hook (informational only)
	if !refileNoVerify {
		hookCtx := &hooks.HookContext{
			Type:        hooks.PostRefile,
			Workspace:   ws,
			SourceFile:  sourceSelector,
			DestPath:    targetSelector,
			Timeout:     30 * time.Second,
			AllowBypass: refileNoVerify,
		}
		
		_, hookErr := hookManager.Execute(hookCtx)
		if hookErr != nil {
			// Check for JSON output to determine if we should show warnings
			if !ctx.IsJSONOutput() {
				fmt.Printf("Warning: post-refile hook failed: %s\n", hookErr.Error())
			}
		}
	}

	if verbose {
		cmdutil.ShowSuccess("✓ Refiled subtree from %s to %s", sourceSelector, targetSelector)
	} else {
		cmdutil.ShowSuccess("✓ Successfully refiled '%s' to '%s'",
			subtree.Heading, destPath.File+"#"+strings.Join(destPath.Segments, "/"))
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
	refileCmd.Flags().BoolP("interactive", "i", false, "Interactive mode using FZF (requires JOT_FZF=1)")
	refileCmd.Flags().BoolVar(&refileNoVerify, "no-verify", false, "Skip hooks verification")
}

// showSelectorsForFile displays available selectors for a specific file
func showSelectorsForFile(ws *workspace.Workspace, filename string) error {
	pathUtil := cmdutil.NewPathUtil(ws)
	// Determine the full file path
	var filePath string
	if filename == "inbox.md" {
		filePath = ws.InboxPath
	} else if filepath.IsAbs(filename) {
		filePath = filename
	} else {
		// Handle files in workspace root, add .md extension if needed
		if !strings.HasSuffix(filename, ".md") {
			filename += ".md"
		}
		// Use workspace root for relative paths, not lib/ directory
		filePath = pathUtil.WorkspaceJoin(filename)
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filename)
	}

	// Read and parse the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return cmdutil.NewFileError("read", filename, err)
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

// navigateHeadingPath navigates through the heading hierarchy following the path
func navigateHeadingPath(doc ast.Node, content []byte, destPath *markdown.HeadingPath) (*PathResolution, error) {
	result := &PathResolution{
		FoundSegments:   []string{},
		MissingSegments: []string{},
	}

	if len(destPath.Segments) == 0 {
		// No path segments, insert at end of file
		return result, nil
	}

	// Find all headings in the document
	allHeadings := markdown.FindAllHeadings(doc, content)

	// Try to find matches for the path segments
	var bestMatch *markdown.HeadingInfo
	var bestMatchDepth int

	for _, heading := range allHeadings {
		// Check if this heading's path contains our target segments
		matchDepth := calculatePathMatch(heading.Path, destPath.Segments, destPath.SkipLevels)

		if matchDepth == len(destPath.Segments) {
			// Found a complete match
			targetHeading := findHeadingByOffset(doc, heading.Offset)
			if targetHeading != nil {
				result.TargetHeading = targetHeading
				result.FoundSegments = destPath.Segments
				return result, nil
			}
		}

		// Track the best partial match
		if matchDepth > bestMatchDepth {
			bestMatchDepth = matchDepth
			bestMatch = &heading
		}
	}

	// Handle partial or no matches
	if bestMatch != nil && bestMatchDepth > 0 {
		parentHeading := findHeadingByOffset(doc, bestMatch.Offset)
		if parentHeading != nil {
			result.ParentHeading = parentHeading
			result.FoundSegments = destPath.Segments[:bestMatchDepth]
			result.MissingSegments = destPath.Segments[bestMatchDepth:]
		}
	} else {
		// No match found, need to create all segments
		result.MissingSegments = destPath.Segments
	}

	return result, nil
}

// calculatePathMatch checks how many consecutive segments match using contains logic
func calculatePathMatch(headingPath []string, targetSegments []string, skipLevels int) int {
	if len(headingPath) < skipLevels {
		return 0
	}

	// Adjust the heading path based on skip levels
	adjustedPath := headingPath[skipLevels:]

	// Find the best consecutive match of targetSegments within adjustedPath
	bestMatch := 0

	// Try starting from each position in the adjusted path
	for startPos := 0; startPos <= len(adjustedPath)-1; startPos++ {
		matchCount := 0

		// Try to match as many consecutive segments as possible from this position
		for i, targetSeg := range targetSegments {
			pathIndex := startPos + i
			if pathIndex >= len(adjustedPath) {
				break
			}

			headingSeg := adjustedPath[pathIndex]
			if strings.Contains(strings.ToLower(headingSeg), strings.ToLower(targetSeg)) {
				matchCount++
			} else {
				break // Stop on first non-match for consecutive matching
			}
		}

		if matchCount > bestMatch {
			bestMatch = matchCount
		}

		// If we found a complete match, we can stop
		if matchCount == len(targetSegments) {
			break
		}
	}

	return bestMatch
}

// findHeadingByOffset finds a heading node by its byte offset
func findHeadingByOffset(doc ast.Node, targetOffset int) *ast.Heading {
	var result *ast.Heading

	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if heading, ok := n.(*ast.Heading); ok {
			offset := markdown.GetNodeOffset(heading, nil) // We don't need content for offset comparison
			if offset == targetOffset {
				result = heading
				return ast.WalkStop, nil
			}
		}

		return ast.WalkContinue, nil
	})

	return result
}

// calculateInsertionPoint finds where to insert content under a heading
func calculateInsertionPoint(heading *ast.Heading, content []byte, prepend bool) int {
	if prepend {
		// Insert right after the heading line
		headingEnd := findHeadingLineEnd(heading, content)
		return headingEnd
	}

	// Find the end of this heading's subtree
	subtreeEnd := markdown.FindSubtreeEnd(heading, content)

	// Back up to find a good insertion point (before the next heading)
	insertPoint := subtreeEnd
	for insertPoint > 0 && content[insertPoint-1] == '\n' {
		insertPoint--
	}

	return insertPoint
}

// findHeadingLineEnd finds the end of the heading line (after the newline)
func findHeadingLineEnd(heading *ast.Heading, content []byte) int {
	startOffset := markdown.GetNodeOffset(heading, content)

	// Find the end of the heading line
	for i := startOffset; i < len(content); i++ {
		if content[i] == '\n' {
			return i + 1 // Return position after the newline
		}
	}

	return len(content)
}

// JSON response structures for refile command
type RefileResponse struct {
	Operation   string            `json:"operation"`
	Source      RefileSource      `json:"source"`
	Destination RefileDestination `json:"destination"`
	Content     RefileContent     `json:"content"`
	Metadata    cmdutil.JSONMetadata      `json:"metadata"`
}

type RefileSource struct {
	Selector      string `json:"selector"`
	FilePath      string `json:"file_path"`
	Heading       string `json:"heading"`
	OriginalLevel int    `json:"original_level"`
}

type RefileDestination struct {
	Selector        string   `json:"selector"`
	FilePath        string   `json:"file_path"`
	TargetLevel     int      `json:"target_level"`
	PathExists      bool     `json:"path_exists"`
	CreatedHeadings []string `json:"created_headings,omitempty"`
}

type RefileContent struct {
	Content          string `json:"content"`
	CharacterCount   int    `json:"character_count"`
	LineCount        int    `json:"line_count"`
	TransformedLevel int    `json:"transformed_level"`
}

// JSON response structures for destination inspection
type InspectDestinationResponse struct {
	Operation   string                     `json:"operation"`
	Destination InspectDestinationInfo     `json:"destination"`
	Analysis    InspectDestinationAnalysis `json:"analysis"`
	Metadata    cmdutil.JSONMetadata               `json:"metadata"`
}

type InspectDestinationInfo struct {
	Selector   string `json:"selector"`
	FilePath   string `json:"file_path"`
	FileExists bool   `json:"file_exists"`
}

type InspectDestinationAnalysis struct {
	PathExists      bool                     `json:"path_exists"`
	FoundSegments   []string                 `json:"found_segments"`
	MissingSegments []string                 `json:"missing_segments"`
	TargetLevel     int                      `json:"target_level"`
	WouldCreate     []InspectHeadingCreation `json:"would_create,omitempty"`
}

type InspectHeadingCreation struct {
	Heading string `json:"heading"`
	Level   int    `json:"level"`
}

// outputRefileJSON outputs JSON response for refile operation
func outputRefileJSON(ctx *cmdutil.CommandContext, sourcePath *markdown.HeadingPath, destPath *markdown.HeadingPath,
	subtree *markdown.Subtree, dest *DestinationTarget, transformedContent []byte) error {

	// Get source file path
	sourceFilePath := sourcePath.File
	if sourcePath.File == "inbox.md" {
		// We need the workspace to get the inbox path, but we don't have it here
		// Let's use the literal name for JSON consistency
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
	} else if filepath.IsAbs(destPath.File) {
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

	doc := markdown.ParseDocument(content)
	pathResolution, err := navigateHeadingPath(doc, content, destPath)
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

// ensureConsistentFormatting ensures content has consistent markdown formatting
func (op *RefileOperation) ensureConsistentFormatting(content []byte) []byte {
	// Trim any trailing whitespace/newlines
	trimmed := strings.TrimRight(string(content), " \t\n")
	
	// Ensure content ends with exactly one newline for consistent formatting
	if len(trimmed) > 0 {
		return []byte(trimmed + "\n")
	}
	return []byte(trimmed)
}

// preserveSpacingAfterRemoval ensures proper spacing after removing a subtree
func (op *RefileOperation) preserveSpacingAfterRemoval(beforeSubtree, afterSubtree []byte) []byte {
	// If either part is empty, just return the other
	if len(beforeSubtree) == 0 {
		return afterSubtree
	}
	if len(afterSubtree) == 0 {
		return beforeSubtree
	}
	
	// Check if afterSubtree starts with a heading (## or #)
	// Skip leading newlines to check actual content
	afterStart := 0
	for afterStart < len(afterSubtree) && afterSubtree[afterStart] == '\n' {
		afterStart++
	}
	
	isNextHeading := false
	if afterStart < len(afterSubtree) && afterSubtree[afterStart] == '#' {
		isNextHeading = true
	}
	
	// Check how beforeSubtree ends
	beforeEndsWithNewline := len(beforeSubtree) > 0 && beforeSubtree[len(beforeSubtree)-1] == '\n'
	
	// If the next content is a heading and beforeSubtree doesn't end with newline,
	// or if we need to ensure proper spacing between sections
	if isNextHeading {
		if beforeEndsWithNewline {
			// Add exactly one blank line between sections
			return append(append(beforeSubtree, '\n'), afterSubtree...)
		} else {
			// Add newline + blank line
			return append(append(beforeSubtree, '\n', '\n'), afterSubtree...)
		}
	}
	
	// Default: just concatenate
	return append(beforeSubtree, afterSubtree...)
}

// normalizeMarkdownSpacing ensures consistent spacing throughout the content
func (op *RefileOperation) normalizeMarkdownSpacing(content []byte) []byte {
	// Simple approach: replace any sequence of 3+ newlines with exactly 2 newlines (one blank line)
	result := string(content)
	
	// Replace multiple consecutive newlines with exactly two (which creates one blank line)
	for strings.Contains(result, "\n\n\n") {
		result = strings.ReplaceAll(result, "\n\n\n", "\n\n")
	}
	
	return []byte(result)
}

// SubtreeItem represents a selectable subtree for FZF interfaces
type SubtreeItem struct {
	Selector string // e.g., "inbox.md#meeting-notes"
	Title    string // Heading title for display
	Level    int    // Heading level (1-6)
	Preview  string // First few lines for display
}

// runInteractiveRefile handles the interactive refile workflow using FZF
func runInteractiveRefile(ctx *cmdutil.CommandContext, args []string, ws *workspace.Workspace) error {
	var sourceSelector, targetSelector string
	var err error

	// Get flags
	to, _ := ctx.Cmd.Flags().GetString("to")
	verbose, _ := ctx.Cmd.Flags().GetBool("verbose")

	// Stage 1 & 2: Select source (if not provided)
	if len(args) > 0 {
		providedArg := args[0]
		// Check if the argument is just a filename (no # selector)
		if !strings.Contains(providedArg, "#") {
			// Treat it as a file and proceed to subtree selection
			if verbose {
				fmt.Printf("Using provided source file: %s\n", providedArg)
			}
			sourceSelector, err = selectSourceSubtree(ws, providedArg, verbose)
			if err != nil {
				return err
			}
			if sourceSelector == "" {
				fmt.Println("Source selection cancelled.")
				return nil
			}
		} else {
			// Complete selector provided
			sourceSelector = providedArg
			if verbose {
				fmt.Printf("Using provided source: %s\n", sourceSelector)
			}
		}
	} else {
		sourceSelector, err = selectSource(ws, verbose)
		if err != nil {
			return err
		}
		if sourceSelector == "" {
			fmt.Println("Source selection cancelled.")
			return nil
		}
	}

	// Stage 3 & 4: Select target (if not provided)
	if to != "" {
		targetSelector = to
		if verbose {
			fmt.Printf("Using provided target: %s\n", targetSelector)
		}
	} else {
		targetSelector, err = selectTarget(ws, verbose)
		if err != nil {
			return err
		}
		if targetSelector == "" {
			fmt.Println("Target selection cancelled.")
			return nil
		}
	}

	// Stage 5: Confirmation
	confirmed, err := confirmRefile(sourceSelector, targetSelector, ws)
	if err != nil {
		return err
	}
	if !confirmed {
		fmt.Println("Refile cancelled.")
		return nil
	}

	// Execute refile using existing logic
	return executeRefile(sourceSelector, targetSelector, ctx, ws)
}

// selectSource handles source file and subtree selection
func selectSource(ws *workspace.Workspace, verbose bool) (string, error) {
	// Stage 1: Select source file
	sourceFile, err := selectSourceFile(ws, "inbox.md", verbose)
	if err != nil {
		return "", fmt.Errorf("source file selection failed: %w", err)
	}
	if sourceFile == "" {
		return "", nil // User cancelled
	}

	// Stage 2: Select subtree from source file
	return selectSourceSubtree(ws, sourceFile, verbose)
}

// selectTarget handles target file and location selection
func selectTarget(ws *workspace.Workspace, verbose bool) (string, error) {
	// Stage 3: Select target file
	targetFile, err := selectTargetFile(ws, verbose)
	if err != nil {
		return "", fmt.Errorf("target file selection failed: %w", err)
	}
	if targetFile == "" {
		return "", nil // User cancelled
	}

	// Stage 4: Select target location
	return selectTargetLocation(ws, targetFile, verbose)
}

// selectSourceFile shows FZF file browser for source selection
func selectSourceFile(ws *workspace.Workspace, defaultFile string, verbose bool) (string, error) {
	files, err := scanWorkspaceMarkdownFiles(ws)
	if err != nil {
		return "", err
	}

	if len(files) == 0 {
		return "", fmt.Errorf("no markdown files found in workspace")
	}

	// Move default file to front if it exists
	if defaultFile != "" {
		files = moveToFront(files, defaultFile)
	}

	if verbose {
		fmt.Printf("Found %d markdown files\n", len(files))
	}

	return runFileSelectionFZF(ws, files, "Select source file > ")
}

// selectSourceSubtree shows FZF subtree browser for the selected file
func selectSourceSubtree(ws *workspace.Workspace, sourceFile string, verbose bool) (string, error) {
	subtrees, err := extractSubtreesFromFile(ws, sourceFile)
	if err != nil {
		return "", fmt.Errorf("failed to extract subtrees: %w", err)
	}

	if len(subtrees) == 0 {
		return "", fmt.Errorf("no headings found in %s - cannot refile from a file without headings", sourceFile)
	}

	// Check for duplicate heading titles
	duplicates := findDuplicateHeadings(subtrees)
	if len(duplicates) > 0 && verbose {
		fmt.Printf("⚠️  Warning: Found duplicate headings in %s: %s\n", sourceFile, strings.Join(duplicates, ", "))
		fmt.Println("   Use the preview (TAB) to distinguish between them")
	}

	if verbose {
		fmt.Printf("Found %d subtrees in %s\n", len(subtrees), sourceFile)
	}

	selector, err := runSubtreeSelectionFZF(subtrees, "Select subtree to refile > ")
	if err != nil {
		return "", err
	}
	
	if selector == "" {
		return "", nil // User cancelled
	}

	// Validate that the selected subtree can be uniquely identified
	return validateAndDisambiguateSelector(ws, selector, subtrees)
}

// selectTargetFile shows FZF file browser for target selection
func selectTargetFile(ws *workspace.Workspace, verbose bool) (string, error) {
	files, err := scanWorkspaceMarkdownFiles(ws)
	if err != nil {
		return "", err
	}

	if len(files) == 0 {
		return "", fmt.Errorf("no markdown files found in workspace")
	}

	if verbose {
		fmt.Printf("Found %d markdown files for target\n", len(files))
	}

	return runFileSelectionFZF(ws, files, "Select target file > ")
}

// selectTargetLocation shows FZF heading browser for the selected target file
func selectTargetLocation(ws *workspace.Workspace, targetFile string, verbose bool) (string, error) {
	subtrees, err := extractSubtreesFromFile(ws, targetFile)
	if err != nil {
		return "", fmt.Errorf("failed to extract headings: %w", err)
	}

	if len(subtrees) == 0 {
		// No headings, allow top-level insertion
		if verbose {
			fmt.Printf("No headings found in %s - content will be inserted at top level\n", targetFile)
		}
		return targetFile, nil
	}

	// Check for duplicate heading titles
	duplicates := findDuplicateHeadings(subtrees)
	if len(duplicates) > 0 && verbose {
		fmt.Printf("⚠️  Warning: Found duplicate headings in %s: %s\n", targetFile, strings.Join(duplicates, ", "))
		fmt.Println("   Use the preview (TAB) to distinguish between them")
	}

	if verbose {
		fmt.Printf("Found %d target locations in %s\n", len(subtrees), targetFile)
	}

	// Add option for top-level insertion
	topLevel := SubtreeItem{
		Selector: targetFile,
		Title:    "(Top level - beginning of file)",
		Level:    0,
		Preview:  "Insert at the beginning of the file",
	}
	allTargets := append([]SubtreeItem{topLevel}, subtrees...)

	selector, err := runSubtreeSelectionFZF(allTargets, "Select target location > ")
	if err != nil {
		return "", err
	}
	
	if selector == "" {
		return "", nil // User cancelled
	}

	// For top-level insertion, convert to proper selector format
	if selector == targetFile {
		return targetFile + "#", nil
	}

	// For heading targets, validate uniqueness
	return validateAndDisambiguateSelector(ws, selector, subtrees)
}

// confirmRefile shows a confirmation dialog before executing the refile
func confirmRefile(sourceSelector, targetSelector string, ws *workspace.Workspace) (bool, error) {
	border := strings.Repeat("=", 60)
	separator := strings.Repeat("-", 60)
	
	fmt.Printf("\n%s\n", border)
	fmt.Printf("📋 REFILE OPERATION SUMMARY\n")
	fmt.Printf("%s\n", border)
	
	// Parse selectors to get more detailed info
	sourcePath, err := markdown.ParsePath(sourceSelector)
	if err != nil {
		fmt.Printf("  Source: %s (⚠️  parse error: %v)\n", sourceSelector, err)
	} else {
		if len(sourcePath.Segments) > 0 {
			fmt.Printf("  📤 Source: %s → '%s'\n", sourcePath.File, strings.Join(sourcePath.Segments, "/"))
		} else {
			fmt.Printf("  📤 Source: %s (entire file)\n", sourcePath.File)
		}
	}
	
	destPath, err := markdown.ParsePath(targetSelector)
	if err != nil {
		fmt.Printf("  Target: %s (⚠️  parse error: %v)\n", targetSelector, err)
	} else {
		if len(destPath.Segments) > 0 {
			fmt.Printf("  📥 Target: %s → '%s'\n", destPath.File, strings.Join(destPath.Segments, "/"))
		} else {
			fmt.Printf("  📥 Target: %s (top level)\n", destPath.File)
		}
	}
	
	fmt.Printf("%s\n", separator)
	
	return cmdutil.ConfirmOperation("\n🚀 Execute refile operation?")
}

// extractSubtreesFromFile extracts all headings from a markdown file
func extractSubtreesFromFile(ws *workspace.Workspace, filename string) ([]SubtreeItem, error) {
	// Determine full file path
	filePath := cmdutil.ResolveWorkspaceRelativePath(ws, filename)

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, cmdutil.NewFileError("read", filePath, err)
	}

	// Parse markdown document
	doc := goldmark.New().Parser().Parse(text.NewReader(content))

	var subtrees []SubtreeItem

	// Walk through the document and find headings
	err = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if heading, ok := n.(*ast.Heading); ok {
			// Extract heading text
			headingText := markdown.ExtractHeadingText(heading, content)
			if strings.TrimSpace(headingText) == "" {
				return ast.WalkContinue, nil
			}

			// Generate selector using proper jot format
			selector := fmt.Sprintf("%s#%s", filename, headingText)

			// Extract preview content (next few lines after heading)
			preview := extractPreviewContent(n, content, 100)

			subtrees = append(subtrees, SubtreeItem{
				Selector: selector,
				Title:    headingText,
				Level:    heading.Level,
				Preview:  preview,
			})
		}

		return ast.WalkContinue, nil
	})

	return subtrees, err
}

// scanWorkspaceMarkdownFiles returns all markdown files in the workspace
func scanWorkspaceMarkdownFiles(ws *workspace.Workspace) ([]string, error) {
	var files []string

	// Add inbox.md if it exists
	if _, err := os.Stat(ws.InboxPath); err == nil {
		files = append(files, "inbox.md")
	}

	// Scan for markdown files in workspace root
	err := filepath.Walk(ws.Root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories and .jot directory
		if info.IsDir() && (strings.HasPrefix(info.Name(), ".") || info.Name() == ".jot") {
			return filepath.SkipDir
		}

		// Only include .md files
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".md") {
			relPath, err := filepath.Rel(ws.Root, path)
			if err == nil && relPath != "inbox.md" { // Don't duplicate inbox.md
				files = append(files, relPath)
			}
		}

		return nil
	})

	return files, err
}

// moveToFront moves the specified item to the front of the slice if it exists
func moveToFront(files []string, target string) []string {
	for i, file := range files {
		if file == target {
			// Move to front
			result := make([]string, len(files))
			result[0] = target
			copy(result[1:], files[:i])
			copy(result[1+i:], files[i+1:])
			return result
		}
	}
	return files
}

// runFileSelectionFZF runs FZF for file selection
func runFileSelectionFZF(ws *workspace.Workspace, files []string, prompt string) (string, error) {
	pathUtil := cmdutil.NewPathUtil(ws)
	// Validate FZF availability
	if _, err := exec.LookPath("fzf"); err != nil {
		return "", fmt.Errorf("fzf not found in PATH. Please install fzf or set JOT_FZF=0 to disable")
	}

	// Create temporary file with file list
	tempFile, err := os.CreateTemp("", "jot-files-*.txt")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Write files to temp file with absolute paths for preview
	for _, file := range files {
		// Resolve to absolute path for preview
		var absolutePath string
		if file == "inbox.md" {
			absolutePath = ws.InboxPath
		} else if filepath.IsAbs(file) {
			absolutePath = file
		} else {
			absolutePath = pathUtil.WorkspaceJoin(file)
		}
		
		// Write both the display name and absolute path (tab-separated)
		fmt.Fprintf(tempFile, "%s\t%s\n", file, absolutePath)
	}
	tempFile.Close()

	// Use the absolute path (second field) for preview
	previewCmd := "head -20 {2}"

	// Build FZF command
	cmd := exec.Command("fzf",
		"--delimiter", "\t",
		"--with-nth", "1", // Only show the filename (first field)
		"--prompt", prompt,
		"--preview", previewCmd,
		"--preview-window", "right:50%:wrap",
		"--bind", "tab:toggle-preview",
		"--header", "ENTER:select | TAB:preview | ESC:cancel",
		"--height", "60%",
		"--border",
	)

	// Set up input from temp file
	tempFileRead, err := os.Open(tempFile.Name())
	if err != nil {
		return "", fmt.Errorf("failed to open temp file: %w", err)
	}
	defer tempFileRead.Close()

	cmd.Stdin = tempFileRead
	cmd.Stderr = os.Stderr

	// Run FZF and capture selection
	output, err := cmd.Output()
	if err != nil {
		// Check if it's a cancellation (exit code 130) vs actual error
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() == 130 {
				return "", nil // User cancelled
			}
		}
		return "", fmt.Errorf("fzf command failed: %w", err)
	}

	selected := strings.TrimSpace(string(output))
	if selected == "" {
		return "", nil
	}

	// Extract just the filename (first field) from the tab-separated result
	parts := strings.Split(selected, "\t")
	if len(parts) > 0 {
		return parts[0], nil
	}
	
	return selected, nil
}

// runSubtreeSelectionFZF runs FZF for subtree selection
func runSubtreeSelectionFZF(subtrees []SubtreeItem, prompt string) (string, error) {
	// Validate FZF availability
	if _, err := exec.LookPath("fzf"); err != nil {
		return "", fmt.Errorf("fzf not found in PATH. Please install fzf or set JOT_FZF=0 to disable")
	}

	// Create temporary file with subtree list
	tempFile, err := os.CreateTemp("", "jot-subtrees-*.txt")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Write subtrees to temp file in format: selector\ttitle\tpreview
	for _, subtree := range subtrees {
		var levelIndent string
		if subtree.Level == 0 {
			levelIndent = "" // Top-level insertion option
		} else {
			levelIndent = strings.Repeat("  ", subtree.Level-1)
		}
		displayTitle := fmt.Sprintf("%s%s", levelIndent, subtree.Title)
		line := fmt.Sprintf("%s\t%s\t%s\n", subtree.Selector, displayTitle, subtree.Preview)
		tempFile.WriteString(line)
	}
	tempFile.Close()

	// Build FZF command
	cmd := exec.Command("fzf",
		"--delimiter", "\t",
		"--with-nth", "2,3", // Show title and preview
		"--prompt", prompt,
		"--preview", "jot peek {1}", // Use first field (selector) for preview
		"--preview-window", "right:50%:wrap",
		"--bind", "tab:toggle-preview",
		"--header", "ENTER:select | TAB:preview | ESC:cancel",
		"--height", "60%",
		"--border",
	)

	// Set up input from temp file
	tempFileRead, err := os.Open(tempFile.Name())
	if err != nil {
		return "", fmt.Errorf("failed to open temp file: %w", err)
	}
	defer tempFileRead.Close()

	cmd.Stdin = tempFileRead
	cmd.Stderr = os.Stderr

	// Run FZF and capture selection
	output, err := cmd.Output()
	if err != nil {
		// Check if it's a cancellation (exit code 130) vs actual error
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() == 130 {
				return "", nil // User cancelled
			}
		}
		return "", fmt.Errorf("fzf command failed: %w", err)
	}

	selectedLine := strings.TrimSpace(string(output))
	if selectedLine == "" {
		return "", nil
	}

	// Extract the selector (first field)
	parts := strings.Split(selectedLine, "\t")
	if len(parts) > 0 {
		return parts[0], nil
	}

	return "", nil
}

// findDuplicateHeadings checks for duplicate heading titles in subtrees
func findDuplicateHeadings(subtrees []SubtreeItem) []string {
	titleCount := make(map[string]int)
	for _, subtree := range subtrees {
		titleCount[subtree.Title]++
	}

	var duplicates []string
	for title, count := range titleCount {
		if count > 1 {
			duplicates = append(duplicates, title)
		}
	}
	return duplicates
}

// validateAndDisambiguateSelector validates a selector and handles ambiguous cases
func validateAndDisambiguateSelector(ws *workspace.Workspace, selector string, subtrees []SubtreeItem) (string, error) {
	// Parse the selector to extract the heading
	parsedPath, err := markdown.ParsePath(selector)
	if err != nil {
		return "", fmt.Errorf("invalid selector format: %w", err)
	}

	if len(parsedPath.Segments) == 0 {
		return selector, nil // File-level selector
	}

	// Get the heading name from the selector
	headingName := parsedPath.Segments[len(parsedPath.Segments)-1]

	// Check how many subtrees match this heading name
	var matches []SubtreeItem
	for _, subtree := range subtrees {
		if subtree.Title == headingName {
			matches = append(matches, subtree)
		}
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("heading '%s' not found in the file", headingName)
	}

	if len(matches) == 1 {
		return selector, nil // Unique match, no ambiguity
	}

	// Multiple matches - this could be problematic for the actual refile operation
	// For now, we'll use the first match but warn the user
	fmt.Printf("⚠️  Warning: Multiple headings named '%s' found. Using the first occurrence.\n", headingName)
	fmt.Printf("   Preview showed: %s\n", matches[0].Preview)
	
	return selector, nil
}

// improvePreviewFormatting enhances the preview content extraction
func improvePreviewFormatting(content []byte, startPos, endPos int) string {
	if startPos >= len(content) || endPos > len(content) || startPos >= endPos {
		return "(empty content)"
	}
	
	lines := strings.Split(string(content[startPos:endPos]), "\n")
	
	var preview strings.Builder
	lineCount := 0
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// Skip empty lines at the beginning
		if preview.Len() == 0 && trimmed == "" {
			continue
		}
		
		// Stop at next heading
		if strings.HasPrefix(trimmed, "#") && preview.Len() > 0 {
			break
		}
		
		// Add line with proper formatting
		if trimmed != "" {
			if preview.Len() > 0 {
				preview.WriteString(" ")
			}
			preview.WriteString(trimmed)
			lineCount++
			
			// Limit preview length
			if preview.Len() > 120 || lineCount >= 3 {
				break
			}
		}
	}
	
	result := preview.String()
	if len(result) > 120 {
		result = result[:117] + "..."
	}
	if result == "" {
		result = "(empty content)"
	}
	
	return result
}

// extractPreviewContent gets some preview text after a heading
func extractPreviewContent(node ast.Node, content []byte, maxLen int) string {
	// Get the position of this heading in the content
	startPos := markdown.GetNodeOffset(node, content)
	
	// Find the heading line and skip it
	lines := strings.Split(string(content), "\n")
	var headingLineIndex int
	var currentPos int
	
	// Find which line this heading is on
	for i, line := range lines {
		if currentPos >= startPos {
			headingLineIndex = i
			break
		}
		currentPos += len(line) + 1 // +1 for newline
	}
	
	// Collect content from the next few lines after the heading
	var preview strings.Builder
	lineCount := 0
	
	for i := headingLineIndex + 1; i < len(lines) && lineCount < 3; i++ {
		line := strings.TrimSpace(lines[i])
		
		// Skip empty lines at the beginning
		if preview.Len() == 0 && line == "" {
			continue
		}
		
		// Stop if we hit another heading
		if strings.HasPrefix(line, "#") {
			break
		}
		
		// Add meaningful content
		if line != "" {
			if preview.Len() > 0 {
				preview.WriteString(" ")
			}
			preview.WriteString(line)
			lineCount++
			
			// Stop if we've gotten enough content
			if preview.Len() > maxLen {
				break
			}
		}
	}
	
	result := preview.String()
	if len(result) > maxLen {
		result = result[:maxLen-3] + "..."
	}
	
	if result == "" {
		result = "No content"
	}
	
	return result
}

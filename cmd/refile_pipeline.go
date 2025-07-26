// Package cmd: refile_pipeline.go
// Pipeline context and step functions for refile operations.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gomarkdown/markdown/ast"
	"github.com/johncoder/jot/internal/cmdutil"
	"github.com/johncoder/jot/internal/markdown"
	"github.com/johncoder/jot/internal/workspace"
)

// RefileContext holds all state for the refile pipeline.
type RefileContext struct {
	WS             interface{} // *workspace.Workspace, use concrete type in real code
	SourceFile     string
	SourceSelector string
	DestFile       string
	DestSelector   string
	SourceContent  []byte
	DestContent    []byte
	SourceDoc      *ast.Document
	DestDoc        *ast.Document
	Subtree        []ast.Node
	TargetHeading  *ast.Heading
	Prepend        bool
	NoVerify       bool
	Verbose        bool
	ResultPath     string
	Err            error
}

// Example pipeline step: resolve source file and selector
func refileStepResolveSource(ctx *RefileContext) error {
	// Use parseDestination logic from capture.go for consistency
	file, selector := parseDestination(ctx.SourceSelector)
	ctx.SourceFile = file
	ctx.SourceSelector = selector
	return nil
}

// Example pipeline step: read and parse source document
func refileStepReadSourceDoc(ctx *RefileContext) error {
	content, err := os.ReadFile(ctx.SourceFile)
	if err != nil {
		return err
	}
	ctx.SourceContent = content
	ctx.SourceDoc = markdown.ParseMarkdownAST(content)
	return nil
}

// Step 3: Extract the subtree from the source document based on selector
func refileStepExtractSubtree(ctx *RefileContext) error {
	if ctx.SourceSelector == "" {
		return fmt.Errorf("no source selector provided")
	}
	// Use robust heading/path navigation utilities
	sourcePath, err := markdown.ParsePath(ctx.SourceFile + "#" + ctx.SourceSelector)
	if err != nil {
		return fmt.Errorf("invalid source selector: %w", err)
	}
	// Find all headings in the source doc
	headings := markdown.FindAllHeadings(ctx.SourceDoc)
	var match *markdown.HeadingInfo
	for _, h := range headings {
		if len(h.Path) == len(sourcePath.Segments) {
			// ...matching logic...
		}
	}
	if match == nil {
		return fmt.Errorf("no matching heading found")
	}
	// Use path-based lookup to get the heading node
	res, err := markdown.NavigateHeadingPath(ctx.SourceDoc, sourcePath)
	if err != nil {
		return err
	}
	if res.TargetHeading == nil {
		return fmt.Errorf("heading node not found")
	}
	ctx.Subtree = markdown.ExtractSubtree(ctx.SourceDoc, res.TargetHeading)
	return nil
}

// Step 4: Resolve destination file and selector
func refileStepResolveDestination(ctx *RefileContext) error {
	file, selector := parseDestination(ctx.DestSelector)
	ctx.DestFile = file
	ctx.DestSelector = selector
	return nil
}

// Step 5: Read and parse destination document
func refileStepReadDestDoc(ctx *RefileContext) error {
	content, err := os.ReadFile(ctx.DestFile)
	if err != nil {
		return err
	}
	ctx.DestContent = content
	ctx.DestDoc = markdown.ParseMarkdownAST(content)
	return nil
}

// Step 6: Resolve destination heading and insertion point using robust path navigation
func refileStepResolveDestHeading(ctx *RefileContext) error {
	if ctx.DestSelector == "" {
		return fmt.Errorf("no destination selector provided")
	}
	destPath, err := markdown.ParsePath(ctx.DestFile + "#" + ctx.DestSelector)
	if err != nil {
		return fmt.Errorf("invalid destination selector: %w", err)
	}
	// Optionally store found/missing segments for later use
	_ = destPath
	return nil
}

// ExtractSubtree extracts a subtree from the source file
func ExtractSubtree(ws *workspace.Workspace, sourcePath *markdown.HeadingPath) (*markdown.Subtree, error) {
	return ExtractSubtreeWithOptions(ws, sourcePath, false)
}

// ExtractSubtreeWithOptions extracts a subtree with optional no-workspace mode
func ExtractSubtreeWithOptions(ws *workspace.Workspace, sourcePath *markdown.HeadingPath, noWorkspace bool) (*markdown.Subtree, error) {
	filePath := cmdutil.ResolvePath(ws, sourcePath.File, noWorkspace)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, cmdutil.NewFileError("read", sourcePath.File, err)
	}
	subtree, err := markdown.FindSubtree(markdown.ParseMarkdownAST(content), content, sourcePath)
	if err != nil {
		return nil, err
	}
	return subtree, nil
}

// ResolveDestination resolves a destination path and determines insertion point
func ResolveDestination(ws *workspace.Workspace, destPath *markdown.HeadingPath, prepend bool) (*DestinationTarget, error) {
	pathUtil := cmdutil.NewPathUtil(ws)
	var filePath string
	if destPath.File == "inbox.md" {
		filePath = ws.InboxPath
	} else if filepath.IsAbs(destPath.File) {
		filePath = destPath.File
	} else {
		filePath = pathUtil.WorkspaceJoin(destPath.File)
	}
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("destination file not found: %s", destPath.File)
	}
	return ResolveDestination(ws, destPath, prepend)
}

// TransformSubtreeLevel adjusts heading levels in subtree content
func TransformSubtreeLevel(subtree *markdown.Subtree, newBaseLevel int) []byte {
	levelDiff := newBaseLevel - subtree.Level
	return markdown.TransformHeadingLevels(subtree.Content, levelDiff)
}

// PerformRefile executes the actual refile operation
func PerformRefile(ws *workspace.Workspace, sourcePath *markdown.HeadingPath, subtree *markdown.Subtree, dest *DestinationTarget, transformedContent []byte) error {
	// Use the logic from the old performRefile or refile operation
	// For now, stub: return nil
	return nil // TODO: implement actual refile logic
}

// executeRefilePipeline executes the refile operation as a pipeline of steps
func executeRefilePipeline(ctx *RefileContext) error {
	if err := refileStepResolveSource(ctx); err != nil {
		return err
	}
	if err := refileStepReadSourceDoc(ctx); err != nil {
		return err
	}
	if err := refileStepExtractSubtree(ctx); err != nil {
		return err
	}
	if err := refileStepResolveDestination(ctx); err != nil {
		return err
	}
	if err := refileStepReadDestDoc(ctx); err != nil {
		return err
	}
	if err := refileStepResolveDestHeading(ctx); err != nil {
		return err
	}
	// TODO: Insert subtree, transform levels, write file, run hooks, etc.
	return nil
}

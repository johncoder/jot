// Package cmd: refile_interactive.go
// Interactive (FZF) mode for refile command.
package cmd

import (
	"fmt"
	"os"

	"github.com/johncoder/jot/internal/cmdutil"
	"github.com/johncoder/jot/internal/fzf"
	"github.com/johncoder/jot/internal/markdown"
	"github.com/johncoder/jot/internal/workspace"
)

// RunInteractiveRefile launches an interactive FZF-based refile UI.
func RunInteractiveRefile(ctx *cmdutil.CommandContext, args []string, ws *workspace.Workspace) error {
	// List all markdown files in the workspace (inbox.md and lib/*.md)
	files, err := workspace.ListMarkdownFiles(ws)
	if err != nil {
		return ctx.HandleError(fmt.Errorf("failed to list markdown files: %w", err))
	}

	// Prompt user to select a source file
	sourceFile, err := fzf.SelectFile("Select source file to refile from:", files)
	if err != nil {
		return ctx.HandleError(fmt.Errorf("source file selection cancelled or failed: %w", err))
	}

	// Parse headings in the source file
	sourceContent, err := os.ReadFile(sourceFile)
	if err != nil {
		return ctx.HandleError(fmt.Errorf("failed to read source file: %w", err))
	}
	sourceDoc := markdown.ParseMarkdownAST(sourceContent)
	headings := markdown.FindAllHeadings(sourceDoc)
	if len(headings) == 0 {
		return ctx.HandleError(fmt.Errorf("no headings found in source file"))
	}

	// Prompt user to select a heading (subtree) to move
	headingOptions := make([]string, len(headings))
	for i, h := range headings {
		headingOptions[i] = fmt.Sprintf("%s (L%d): %s", sourceFile, h.Level, h.Text)
	}
	_, err = fzf.SelectIndex("Select heading/subtree to refile:", headingOptions)
	if err != nil {
		return ctx.HandleError(fmt.Errorf("heading selection cancelled or failed: %w", err))
	}

	// Prompt user to select a destination file
	destFile, err := fzf.SelectFile("Select destination file:", files)
	if err != nil {
		return ctx.HandleError(fmt.Errorf("destination file selection cancelled or failed: %w", err))
	}
	destContent, err := os.ReadFile(destFile)
	if err != nil {
		return ctx.HandleError(fmt.Errorf("failed to read destination file: %w", err))
	}
	destDoc := markdown.ParseMarkdownAST(destContent)
	destHeadings := markdown.FindAllHeadings(destDoc)
	destOptions := []string{"(top of file)"}
	for _, h := range destHeadings {
		destOptions = append(destOptions, fmt.Sprintf("%s (L%d): %s", destFile, h.Level, h.Text))
	}
	_, err = fzf.SelectIndex("Select destination heading:", destOptions)
	if err != nil {
		return ctx.HandleError(fmt.Errorf("destination selection cancelled or failed: %w", err))
	}

	// Construct a RefileContext and execute the refile pipeline
	refileCtx := &RefileContext{
		SourceFile: sourceFile,
		DestFile:   destFile,
		WS:         ws,
	}
	return executeRefilePipeline(refileCtx)
}

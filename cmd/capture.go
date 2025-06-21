package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/johncoder/jot/internal/editor"
	"github.com/johncoder/jot/internal/markdown"
	"github.com/johncoder/jot/internal/template"
	"github.com/johncoder/jot/internal/workspace"
	"github.com/spf13/cobra"
)

var (
	captureNote     string
	captureStdin    bool
	captureTemplate string
	captureContent  string
)

var captureCmd = &cobra.Command{
	Use:   "capture [template]",
	Short: "Capture a new note",
	Long: `Capture a new note and add it to inbox.md.

Supports templates for structured note capture:
- Templates open in editor by default for interactive editing
- Use --content for quick append to template
- Piped content automatically appends to template body

Input methods:
- Template-based (default): Opens template in editor
- Direct content: Use --content "your note text"  
- Stdin input: Pipe content or use explicit flags

Examples:
  jot capture                              # Open editor
  jot capture meeting                      # Use meeting template in editor
  jot capture --template meeting           # Use meeting template in editor (same as above)
  jot capture standup --content "Completed API design"
  echo "Notes here" | jot capture meeting
  jot capture --content "Quick note"       # Direct append to inbox`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := workspace.RequireWorkspace()
		if err != nil {
			return err
		}

		// Use positional argument as template name if provided
		if len(args) > 0 {
			captureTemplate = args[0]
		}

		// Determine content source
		var appendContent string
		var useEditor bool = true

		// Check for piped input
		stat, _ := os.Stdin.Stat()
		hasPipedInput := (stat.Mode() & os.ModeCharDevice) == 0

		// Get content from various sources
		switch {
		case captureContent != "":
			appendContent = strings.TrimSpace(captureContent)
			useEditor = false
		case hasPipedInput:
			stdin, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read from stdin: %w", err)
			}
			appendContent = strings.TrimSpace(string(stdin))
			// If using template with piped input, still use editor unless --content specified
			if captureTemplate == "" {
				useEditor = false
			}
		case captureNote != "":
			// Legacy support for --note flag
			appendContent = strings.TrimSpace(captureNote)
			useEditor = false
		}

		var finalContent string

		// Handle template-based capture
		if captureTemplate != "" {
			tm := template.NewManager(ws)
			t, err := tm.Get(captureTemplate)
			if err != nil {
				return fmt.Errorf("template error: %w", err)
			}

			// Render template with shell commands and append content
			renderedTemplate, err := tm.Render(t, appendContent)
			if err != nil {
				return err
			}

			if useEditor {
				// Open rendered template in editor
				tempFile, err := os.CreateTemp("", "jot-capture-*.md")
				if err != nil {
					return fmt.Errorf("failed to create temp file: %w", err)
				}
				defer os.Remove(tempFile.Name())

				if _, err := tempFile.WriteString(renderedTemplate); err != nil {
					tempFile.Close()
					return fmt.Errorf("failed to write template to temp file: %w", err)
				}
				tempFile.Close()

				fmt.Printf("Opening template '%s' in editor...\n", captureTemplate)
				editedContent, err := editor.OpenEditor(renderedTemplate)
				if err != nil {
					return fmt.Errorf("failed to open editor: %w", err)
				}
				finalContent = strings.TrimSpace(editedContent)
			} else {
				finalContent = renderedTemplate
			}

			// Use DestinationFile if specified - can be either a file or selector
			destination := t.DestinationFile
			if destination == "" {
				destination = "inbox.md"
			}

			// Check if destination is a selector (contains #) or just a file
			if strings.Contains(destination, "#") {
				// Use selector-based refile logic
				if err := refileContentToDestination(ws, finalContent, destination, t.RefileMode); err != nil {
					return fmt.Errorf("failed to refile to destination '%s': %w", destination, err)
				}
				fmt.Printf("✓ Captured '%s' and refiled to '%s'\n", captureTemplate, destination)
			} else {
				// Simple file destination
				destinationPath := destination
				if destination == "inbox.md" {
					destinationPath = ws.InboxPath
				} else if !filepath.IsAbs(destination) {
					// Use workspace root for relative paths, not lib/ directory
					destinationPath = filepath.Join(ws.Root, destination)
				}

				if err := ws.AppendToFile(destinationPath, finalContent); err != nil {
					return fmt.Errorf("failed to save note: %w", err)
				}
				fmt.Printf("✓ Captured '%s' to '%s'\n", captureTemplate, destination)
			}

			return nil
		} else {
			// No template - handle as before
			if appendContent == "" && useEditor {
				// Open editor for free-form capture
				fmt.Println("Opening editor for note capture...")
				editedContent, err := editor.OpenEditor("")
				if err != nil {
					return fmt.Errorf("failed to open editor: %w", err)
				}
				finalContent = strings.TrimSpace(editedContent)
			} else {
				finalContent = appendContent
			}
		}

		if finalContent == "" {
			fmt.Println("No content captured. Note not saved.")
			return nil
		}

		// Append to inbox
		if err := ws.AppendToInbox(finalContent); err != nil {
			return fmt.Errorf("failed to save note: %w", err)
		}

		fmt.Printf("✓ Note captured (%d characters)\n", len(finalContent))
		if captureTemplate != "" {
			fmt.Printf("✓ Used template: %s\n", captureTemplate)
		}
		fmt.Printf("✓ Added to %s\n", ws.InboxPath)

		return nil
	},
}

func init() {
	captureCmd.Flags().StringVar(&captureTemplate, "template", "", "Use a named template for structured capture")
	captureCmd.Flags().StringVar(&captureContent, "content", "", "Note content to append (skips editor)")
}

// refileContentToDestination performs refile operation for captured content
func refileContentToDestination(ws *workspace.Workspace, content, destination, mode string) error {
	// Parse the destination
	destPath, err := markdown.ParsePath(destination)
	if err != nil {
		return fmt.Errorf("invalid destination '%s': %w", destination, err)
	}

	// Create a temporary subtree from the captured content
	// We'll wrap the content in a heading to make it a proper subtree
	tempContent := "# Captured Content\n\n" + content
	capturedSubtree := &markdown.Subtree{
		Heading:     "Captured Content",
		Level:       1,
		Content:     []byte(tempContent),
		StartOffset: 0,
		EndOffset:   len(tempContent),
	}

	// Use the existing refile functionality to resolve the destination
	dest, err := ResolveDestination(ws, destPath, mode == "prepend")
	if err != nil {
		return fmt.Errorf("failed to resolve destination: %w", err)
	}

	// Transform the subtree level to match the destination
	transformedContent := TransformSubtreeLevel(capturedSubtree, dest.TargetLevel)

	// Perform the direct insertion (similar to refile but without removing from source)
	return performDirectInsertion(ws, dest, transformedContent)
}

// performDirectInsertion inserts content directly into the destination file
func performDirectInsertion(ws *workspace.Workspace, dest *DestinationTarget, transformedContent []byte) error {
	// Construct destination file path
	var destFilePath string
	if dest.File == "inbox.md" {
		destFilePath = ws.InboxPath
	} else if filepath.IsAbs(dest.File) {
		destFilePath = dest.File
	} else {
		// Use workspace root for relative paths, not lib/ directory
		destFilePath = filepath.Join(ws.Root, dest.File)
	}

	// Read destination file
	destContent, err := os.ReadFile(destFilePath)
	if err != nil {
		return fmt.Errorf("failed to read destination file: %w", err)
	}

	// Prepare content to insert
	var insertContent []byte = transformedContent
	
	// Add missing headings if needed
	if len(dest.CreatePath) > 0 {
		// Calculate the base level for missing headings
		baseLevel := dest.TargetLevel - len(dest.CreatePath)
		pathContent := markdown.CreateHeadingStructure(dest.CreatePath, baseLevel)
		
		// Ensure proper spacing
		if dest.InsertOffset > 0 && destContent[dest.InsertOffset-1] != '\n' {
			pathContent = append([]byte("\n"), pathContent...)
		}
		
		insertContent = append(pathContent, insertContent...)
	}

	// Insert at the specified offset
	newDestContent := append(destContent[:dest.InsertOffset], insertContent...)
	newDestContent = append(newDestContent, destContent[dest.InsertOffset:]...)

	// Write back to destination file
	if err := os.WriteFile(destFilePath, newDestContent, 0644); err != nil {
		return fmt.Errorf("failed to write destination file: %w", err)
	}

	return nil
}

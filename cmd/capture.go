package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/johncoder/jot/internal/editor"
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
	Use:   "capture",
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
  jot capture --template meeting           # Use meeting template in editor
  jot capture --template standup --content "Completed API design"
  echo "Notes here" | jot capture --template meeting
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

			// Use DestinationFile if specified
			destinationFile := t.DestinationFile
			if destinationFile == "" {
				destinationFile = ws.InboxPath
			}

			if err := ws.AppendToFile(destinationFile, finalContent); err != nil {
				return fmt.Errorf("failed to save note: %w", err)
			}

			fmt.Printf("✓ Note captured (%d characters)\n", len(finalContent))
			fmt.Printf("✓ Used template: %s\n", captureTemplate)
			fmt.Printf("✓ Added to %s\n", destinationFile)

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
	captureCmd.Flags().StringVar(&captureNote, "note", "", "(deprecated) Use --content instead")
	captureCmd.Flags().BoolVar(&captureStdin, "stdin", false, "(deprecated) Piped input is auto-detected")
}

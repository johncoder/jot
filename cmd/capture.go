package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/johncoder/jot/internal/cmdutil"
	"github.com/johncoder/jot/internal/editor"
	"github.com/johncoder/jot/internal/hooks"
	"github.com/johncoder/jot/internal/markdown"
	"github.com/johncoder/jot/internal/template"
	"github.com/johncoder/jot/internal/workspace"
	"github.com/spf13/cobra"
)

var (
	captureNote     string
	captureTemplate string
	captureContent  string
	captureNoVerify bool
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
		ctx := cmdutil.StartCommand(cmd)

		ws, err := getWorkspace(cmd)
		if err != nil {
			return ctx.HandleError(err)
		}

		// Initialize hook manager
		hookManager := hooks.NewManager(ws)

		// Run pre-capture hook unless --no-verify is set
		if !captureNoVerify {
			hookCtx := &hooks.HookContext{
				Type:         hooks.PreCapture,
				Workspace:    ws,
				Content:      captureContent,
				TemplateName: captureTemplate,
				Timeout:      30 * time.Second,
				AllowBypass:  captureNoVerify,
			}
			
			result, err := hookManager.Execute(hookCtx)
			if err != nil {
				return ctx.HandleOperationError("pre-capture hook", fmt.Errorf("pre-capture hook failed: %s", err.Error()))
			}
			
			if result.Aborted {
				return ctx.HandleOperationError("pre-capture hook", fmt.Errorf("pre-capture hook aborted operation"))
			}
			
			// Update content if hook modified it
			if result.Content != captureContent {
				captureContent = result.Content
			}
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
				return ctx.HandleOperationError("template", fmt.Errorf("template error: %w", err))
			}

			// Render template with shell commands and append content
			renderedTemplate, err := tm.Render(t, appendContent)
			if err != nil {
				return ctx.HandleOperationError("template", err)
			}

			if useEditor {
				// Open rendered template in editor
				tempFile, err := os.CreateTemp("", "jot-capture-*.md")
				if err != nil {
					return ctx.HandleOperationError("temp file", fmt.Errorf("failed to create temp file: %w", err))
				}
				defer os.Remove(tempFile.Name())

				if _, err := tempFile.WriteString(renderedTemplate); err != nil {
					tempFile.Close()
					return ctx.HandleOperationError("temp file", fmt.Errorf("failed to write template to temp file: %w", err))
				}
				tempFile.Close()

				if !ctx.IsJSONOutput() {
					fmt.Printf("Opening template '%s' in editor...\n", captureTemplate)
				}
				editedContent, err := editor.OpenEditor(renderedTemplate)
				if err != nil {
					return ctx.HandleOperationError("editor", fmt.Errorf("failed to open editor: %w", err))
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
					return ctx.HandleOperationError("refile", fmt.Errorf("failed to refile to destination '%s': %w", destination, err))
				}

				if ctx.IsJSONOutput() {
					templateInfo := &CaptureTemplate{
						Name:            captureTemplate,
						RenderedContent: finalContent,
						DestinationFile: destination,
						RefileMode:      t.RefileMode,
					}
					lineCount := strings.Count(finalContent, "\n") + 1
					if len(finalContent) == 0 {
						lineCount = 0
					}
					
					response := CaptureResponse{
						Operation: "capture_and_refile",
						ContentInfo: CaptureContent{
							Content:        finalContent,
							CharacterCount: len(finalContent),
							LineCount:      lineCount,
							Source:         getContentSource(appendContent, useEditor),
						},
						FileInfo: CaptureFile{
							FilePath:    destination,
							IsInbox:     false,
							IsSelector:  true,
							Destination: destination,
						},
						Template: templateInfo,
						Metadata: cmdutil.CreateJSONMetadata(ctx.Cmd, true, ctx.StartTime),
					}
					return cmdutil.OutputJSON(response)
				}

				// Run post-capture hook for refile case
				if !captureNoVerify {
					hookCtx := &hooks.HookContext{
						Type:         hooks.PostCapture,
						Workspace:    ws,
						Content:      finalContent,
						TemplateName: captureTemplate,
						SourceFile:   destination,
						Timeout:      30 * time.Second,
						AllowBypass:  captureNoVerify,
					}
					
					_, err := hookManager.Execute(hookCtx)
					if err != nil && !ctx.IsJSONOutput() {
						fmt.Printf("Warning: post-capture hook failed: %s\n", err.Error())
					}
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
					return ctx.HandleOperationError("save", fmt.Errorf("failed to save note: %w", err))
				}

				if ctx.IsJSONOutput() {
					templateInfo := &CaptureTemplate{
						Name:            captureTemplate,
						RenderedContent: finalContent,
						DestinationFile: destination,
						RefileMode:      t.RefileMode,
					}
					lineCount := strings.Count(finalContent, "\n") + 1
					if len(finalContent) == 0 {
						lineCount = 0
					}
					
					response := CaptureResponse{
						Operation: "capture_to_file",
						ContentInfo: CaptureContent{
							Content:        finalContent,
							CharacterCount: len(finalContent),
							LineCount:      lineCount,
							Source:         getContentSource(appendContent, useEditor),
						},
						FileInfo: CaptureFile{
							FilePath:    destinationPath,
							IsInbox:     destination == "inbox.md",
							IsSelector:  false,
							Destination: destination,
						},
						Template: templateInfo,
						Metadata: cmdutil.CreateJSONMetadata(ctx.Cmd, true, ctx.StartTime),
					}
					return cmdutil.OutputJSON(response)
				}

				// Run post-capture hook for file destination case
				if !captureNoVerify {
					hookCtx := &hooks.HookContext{
						Type:         hooks.PostCapture,
						Workspace:    ws,
						Content:      finalContent,
						TemplateName: captureTemplate,
						SourceFile:   destinationPath,
						Timeout:      30 * time.Second,
						AllowBypass:  captureNoVerify,
					}
					
					_, err := hookManager.Execute(hookCtx)
					if err != nil && !ctx.IsJSONOutput() {
						fmt.Printf("Warning: post-capture hook failed: %s\n", err.Error())
					}
				}

				fmt.Printf("✓ Captured '%s' to '%s'\n", captureTemplate, destination)
			}

			return nil
		} else {
			// No template - handle as before
			if appendContent == "" && useEditor {
				// Open editor for free-form capture
				if !ctx.IsJSONOutput() {
					fmt.Println("Opening editor for note capture...")
				}
				editedContent, err := editor.OpenEditor("")
				if err != nil {
					return ctx.HandleOperationError("editor", fmt.Errorf("failed to open editor: %w", err))
				}
				finalContent = strings.TrimSpace(editedContent)
			} else {
				finalContent = appendContent
			}
		}

		if finalContent == "" {
			if ctx.IsJSONOutput() {
				// For JSON, we still return success but with empty content
				return ctx.Response.RespondWithSuccess(map[string]interface{}{
					"operation": "capture_empty",
					"content_info": map[string]interface{}{
						"content":         "",
						"character_count": 0,
						"line_count":      0,
						"source":          getContentSource(appendContent, useEditor),
					},
					"file_info": map[string]interface{}{
						"file_path":   ws.InboxPath,
						"is_inbox":    true,
						"is_selector": false,
						"destination": "inbox.md",
					},
				})
			}
			fmt.Println("No content captured. Note not saved.")
			return nil
		}

		// Append to inbox
		if err := ws.AppendToInbox(finalContent); err != nil {
			return ctx.HandleOperationError("save", fmt.Errorf("failed to save note: %w", err))
		}

		// Run post-capture hook unless --no-verify is set
		if !captureNoVerify {
			hookCtx := &hooks.HookContext{
				Type:         hooks.PostCapture,
				Workspace:    ws,
				Content:      finalContent,
				TemplateName: captureTemplate,
				SourceFile:   ws.InboxPath,
				Timeout:      30 * time.Second,
				AllowBypass:  captureNoVerify,
			}
			
			_, err := hookManager.Execute(hookCtx)
			if err != nil {
				// Post-capture hooks are informational only - log but don't fail
				if !ctx.IsJSONOutput() {
					fmt.Printf("Warning: post-capture hook failed: %s\n", err.Error())
				}
			}
		}

		// Handle JSON output
		if ctx.IsJSONOutput() {
			var templateInfo *CaptureTemplate
			if captureTemplate != "" {
				templateInfo = &CaptureTemplate{
					Name:            captureTemplate,
					RenderedContent: finalContent,
					DestinationFile: "inbox.md",
					RefileMode:      "append",
				}
			}
			lineCount := strings.Count(finalContent, "\n") + 1
			if len(finalContent) == 0 {
				lineCount = 0
			}
			
			response := CaptureResponse{
				Operation: "capture",
				ContentInfo: CaptureContent{
					Content:        finalContent,
					CharacterCount: len(finalContent),
					LineCount:      lineCount,
					Source:         getContentSource(appendContent, useEditor),
				},
				FileInfo: CaptureFile{
					FilePath:    ws.InboxPath,
					IsInbox:     true,
					IsSelector:  false,
					Destination: "inbox.md",
				},
				Template: templateInfo,
				Metadata: cmdutil.CreateJSONMetadata(ctx.Cmd, true, ctx.StartTime),
			}
			return cmdutil.OutputJSON(response)
		}

		// Human-readable output
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
	captureCmd.Flags().StringVar(&captureNote, "note", "", "Note content to append (legacy alias for --content)")
	captureCmd.Flags().BoolVar(&captureNoVerify, "no-verify", false, "Skip hooks verification")
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

	// Read destination file using unified content utilities
	destContent, err := cmdutil.ReadFileContent(destFilePath)
	if err != nil {
		return err
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

// JSON response structures for capture command
type CaptureResponse struct {
	Operation   string           `json:"operation"`
	ContentInfo CaptureContent   `json:"content_info"`
	FileInfo    CaptureFile      `json:"file_info"`
	Template    *CaptureTemplate `json:"template,omitempty"`
	Metadata    cmdutil.JSONMetadata     `json:"metadata"`
}

type CaptureContent struct {
	Content        string `json:"content"`
	CharacterCount int    `json:"character_count"`
	LineCount      int    `json:"line_count"`
	Source         string `json:"source"` // "editor", "stdin", "content_flag", "template"
}

type CaptureFile struct {
	FilePath    string `json:"file_path"`
	IsInbox     bool   `json:"is_inbox"`
	IsSelector  bool   `json:"is_selector"`
	Destination string `json:"destination"`
}

type CaptureTemplate struct {
	Name            string `json:"name"`
	RenderedContent string `json:"rendered_content,omitempty"`
	DestinationFile string `json:"destination_file,omitempty"`
	RefileMode      string `json:"refile_mode,omitempty"`
}

// getContentSource determines the source of content for JSON output
func getContentSource(appendContent string, useEditor bool) string {
	if appendContent != "" && !useEditor {
		return "content_flag"
	} else if appendContent != "" && useEditor {
		return "template" // Template with piped/flag content
	} else if useEditor {
		return "editor"
	}
	return "stdin"
}

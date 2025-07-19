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
	"gopkg.in/yaml.v2"
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
				return ctx.HandleExternalCommand("pre-capture hook", nil, err)
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

		// Consolidated content source handling
		appendContent, useEditor, err := getCaptureContent(captureContent, captureNote, captureTemplate, true)
		if err != nil {
			return ctx.HandleOperationError("content", err)
		}
		var finalContent string
		isTemplate := captureTemplate != ""
		if isTemplate {
			tm := template.NewManager(ws)
			t, err := tm.Get(captureTemplate)
			if err != nil {
				return ctx.HandleOperationError("template", fmt.Errorf("template error: %w", err))
			}
			renderedTemplate, err := tm.Render(t, appendContent)
			if err != nil {
				return ctx.HandleOperationError("template", err)
			}
			if useEditor {
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
			// Remove frontmatter
			finalContent = template.StripFrontmatter(finalContent)
			// Unified frontmatter parsing
			fm := parseFrontmatter(t.Content, t.DestinationFile, t.RefileMode)
			if finalContent == "" {
				return handleCaptureOutput(ctx, ws, hookManager, "", appendContent, captureTemplate, fm, true, "capture_empty")
			}
			if err := saveCapturedContent(ws, fm, finalContent); err != nil {
				return ctx.HandleOperationError("save", fmt.Errorf("failed to save note: %w", err))
			}
			if !captureNoVerify {
				hookCtx := &hooks.HookContext{
					Type:         hooks.PostCapture,
					Workspace:    ws,
					Content:      finalContent,
					TemplateName: captureTemplate,
					SourceFile:   fm.Destination,
					Timeout:      30 * time.Second,
					AllowBypass:  captureNoVerify,
				}
				_, err := hookManager.Execute(hookCtx)
				if err != nil && !ctx.IsJSONOutput() {
					cmdutil.ShowWarning("Warning: post-capture hook failed: %s", err.Error())
				}
			}
			return handleCaptureOutput(ctx, ws, hookManager, finalContent, appendContent, captureTemplate, fm, true, func() string {
				if strings.Contains(fm.Destination, "#") {
					return "capture_and_refile"
				}
				return "capture_to_file"
			}())
		} else {
			if appendContent == "" && useEditor {
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
			if finalContent == "" {
				fm := parseFrontmatter("", "inbox.md", "append")
				return handleCaptureOutput(ctx, ws, hookManager, "", appendContent, captureTemplate, fm, false, "capture_empty")
			}
			// Unified frontmatter parsing
			fm := parseFrontmatter(finalContent, "inbox.md", "append")
			finalContent = template.StripFrontmatter(finalContent)
			if err := saveCapturedContent(ws, fm, finalContent); err != nil {
				return ctx.HandleOperationError("save", fmt.Errorf("failed to save note: %w", err))
			}
			if !captureNoVerify {
				hookCtx := &hooks.HookContext{
					Type:         hooks.PostCapture,
					Workspace:    ws,
					Content:      finalContent,
					TemplateName: captureTemplate,
					SourceFile:   fm.Destination,
					Timeout:      30 * time.Second,
					AllowBypass:  captureNoVerify,
				}
				_, err := hookManager.Execute(hookCtx)
				if err != nil && !ctx.IsJSONOutput() {
					cmdutil.ShowWarning("Warning: post-capture hook failed: %s", err.Error())
				}
			}
			return handleCaptureOutput(ctx, ws, hookManager, finalContent, appendContent, captureTemplate, fm, false, "capture")
		}
	},
}

func init() {
	captureCmd.Flags().StringVar(&captureTemplate, "template", "", "Use a named template for structured capture")
	captureCmd.Flags().StringVar(&captureContent, "content", "", "Note content to append (skips editor)")
	captureCmd.Flags().StringVar(&captureNote, "note", "", "Note content to append (legacy alias for --content)")
	captureCmd.Flags().BoolVar(&captureNoVerify, "no-verify", false, "Skip hooks verification")
}

// Helper: get content from all sources (flag, piped, note, editor)
func getCaptureContent(captureContent, captureNote, captureTemplate string, useEditorDefault bool) (string, bool, error) {
	stat, _ := os.Stdin.Stat()
	hasPipedInput := (stat.Mode() & os.ModeCharDevice) == 0
	if captureContent != "" {
		return strings.TrimSpace(captureContent), false, nil
	} else if hasPipedInput {
		stdin, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", false, fmt.Errorf("failed to read from stdin: %w", err)
		}
		return strings.TrimSpace(string(stdin)), captureTemplate == "", nil
	} else if captureNote != "" {
		return strings.TrimSpace(captureNote), false, nil
	}
	return "", useEditorDefault, nil
}

// Helper: parse frontmatter from content or template
func parseFrontmatter(content string, fallbackDest, fallbackMode string) CaptureFrontmatter {
	var fm CaptureFrontmatter
	renderedFrontmatter := template.GetRenderedTemplateFrontmatter(content)
	if renderedFrontmatter != "" {
		_ = yaml.Unmarshal([]byte(strings.ReplaceAll(renderedFrontmatter, "---", "")), &fm)
	}
	if fm.Destination == "" {
		fm.Destination = fallbackDest
	}
	if fm.RefileMode == "" {
		fm.RefileMode = fallbackMode
	}
	return fm
}

// Helper: unified save logic for all destinations/modes
func saveCapturedContent(ws *workspace.Workspace, fm CaptureFrontmatter, content string) error {
	pathUtil := cmdutil.NewPathUtil(ws)
	if fm.Destination == "inbox.md" {
		if fm.RefileMode == "prepend" {
			return ws.PrependToInbox(content)
		}
		return ws.AppendToInbox(content)
	} else if strings.Contains(fm.Destination, "#") {
		return refileContentToDestination(ws, content, fm.Destination, fm.RefileMode)
	} else {
		destPath := fm.Destination
		if !filepath.IsAbs(destPath) {
			destPath = pathUtil.WorkspaceJoin(destPath)
		}
		if err := ws.EnsureFileExists(destPath); err != nil {
			return err
		}
		if fm.RefileMode == "prepend" {
			return ws.PrependToFile(destPath, content)
		}
		return ws.AppendToFile(destPath, content)
	}
}

// Helper: output and hook handling
func handleCaptureOutput(ctx *cmdutil.CommandContext, ws *workspace.Workspace, hookManager *hooks.Manager, finalContent, appendContent, captureTemplate string, fm CaptureFrontmatter, isTemplate bool, jsonOp string) error {
	if ctx.IsJSONOutput() {
		lineCount := strings.Count(finalContent, "\n") + 1
		if len(finalContent) == 0 {
			lineCount = 0
		}
		response := CaptureResponse{
			Operation: jsonOp,
			ContentInfo: CaptureContent{
				Content:        finalContent,
				CharacterCount: len(finalContent),
				LineCount:      lineCount,
				Source:         getContentSource(appendContent, !isTemplate),
			},
			FileInfo: CaptureFile{
				FilePath:    fm.Destination,
				IsInbox:     fm.Destination == "inbox.md",
				IsSelector:  strings.Contains(fm.Destination, "#"),
				Destination: fm.Destination,
			},
			Template: func() *CaptureTemplate {
				if isTemplate {
					return &CaptureTemplate{
						Name:            captureTemplate,
						RenderedContent: finalContent,
						DestinationFile: fm.Destination,
						RefileMode:      fm.RefileMode,
					}
				}
				return nil
			}(),
			Metadata: cmdutil.CreateJSONMetadata(ctx.Cmd, true, ctx.StartTime),
		}
		return cmdutil.OutputJSON(response)
	}
	// Human output
	if isTemplate {
		cmdutil.ShowSuccess("✓ Captured '%s' to '%s'", captureTemplate, fm.Destination)
	} else {
		cmdutil.ShowSuccess("✓ Note captured (%d characters)", len(finalContent))
		cmdutil.ShowSuccess("✓ Added to %s", fm.Destination)
	}
	return nil
}

// refileContentToDestination performs refile operation for captured content
func refileContentToDestination(ws *workspace.Workspace, content, destination, mode string) error {
	// Parse the destination
	destPath, err := markdown.ParsePath(destination)
	if err != nil {
		return cmdutil.NewValidationError("destination", destination, err)
	}

	// Do not wrap content in any heading or subtree
	capturedContent := []byte(content)

	// Use the existing refile functionality to resolve the destination
	pathUtil := cmdutil.NewPathUtil(ws)
	var absDestFile string
	if filepath.IsAbs(destPath.File) {
		absDestFile = destPath.File
	} else {
		absDestFile = pathUtil.WorkspaceJoin(destPath.File)
	}
	if err := ws.EnsureFileExists(absDestFile); err != nil {
		return err
	}
	dest, err := ResolveDestination(ws, destPath, mode == "prepend")
	if err != nil {
		return fmt.Errorf("failed to resolve destination: %w", err)
	}

	// Insert content directly, no subtree transformation
	return performDirectInsertion(ws, dest, capturedContent)
}

// performDirectInsertion inserts content directly into the destination file
func performDirectInsertion(ws *workspace.Workspace, dest *DestinationTarget, transformedContent []byte) error {
	// Construct destination file path
	pathUtil := cmdutil.NewPathUtil(ws)
	var destFilePath string
	if dest.File == "inbox.md" {
		destFilePath = ws.InboxPath
	} else if filepath.IsAbs(dest.File) {
		destFilePath = dest.File
	} else {
		// Use workspace root for relative paths, not lib/ directory
		destFilePath = pathUtil.WorkspaceJoin(dest.File)
	}

	if err := ws.EnsureFileExists(destFilePath); err != nil {
		return err
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
		var headingLines []string
		for i, heading := range dest.CreatePath {
			level := baseLevel + i + 1
			headingLines = append(headingLines, strings.Repeat("#", level)+" "+heading)
		}
		insertContent = append([]byte(strings.Join(headingLines, "\n\n")+"\n\n"), insertContent...)
	}

	// Insert at the specified offset
	newDestContent := append(destContent[:dest.InsertOffset], insertContent...)
	newDestContent = append(newDestContent, destContent[dest.InsertOffset:]...)

	// Write back to destination file
	if err := pathUtil.SafeWriteFile(destFilePath, newDestContent); err != nil {
		return fmt.Errorf("failed to write destination file: %w", err)
	}

	return nil
}

// JSON response structures for capture command
type CaptureResponse struct {
	Operation   string               `json:"operation"`
	ContentInfo CaptureContent       `json:"content_info"`
	FileInfo    CaptureFile          `json:"file_info"`
	Template    *CaptureTemplate     `json:"template,omitempty"`
	Metadata    cmdutil.JSONMetadata `json:"metadata"`
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

// Helper struct for frontmatter
type CaptureFrontmatter struct {
	Destination string `yaml:"destination"`
	RefileMode  string `yaml:"refile_mode"`
}

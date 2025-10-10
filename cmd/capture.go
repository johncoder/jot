package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/gomarkdown/markdown/ast"
	"github.com/johncoder/jot/internal/editor"
	"github.com/johncoder/jot/internal/hooks"
	"github.com/johncoder/jot/internal/markdown"
	"github.com/johncoder/jot/internal/template"
	"github.com/johncoder/jot/internal/workspace"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// captureCmd is the primary capture command, using robust Markdown AST operations.
var captureCmd = &cobra.Command{
	Use:   "capture [template]",
	Short: "Capture a note to inbox or a specific destination (AST-based)",
	Long: `Capture a new note and add it to inbox.md or a specific destination.

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
  jot capture --content "Quick note"       # Direct append to inbox

Flags:
  --template, -t     Use a named template for structured capture
  --content          Note content to append (skips editor)
  --editor           Edit the note in $EDITOR before saving
  --prepend          Prepend the note instead of appending
  --destination, -d  Destination file[#selector] for the note
  --json             Output result as JSON
  --no-verify        Skip hooks verification
`,
	Args: cobra.MaximumNArgs(1),
	RunE: runCapture,
}

func init() {
	captureCmd.Flags().StringP("destination", "d", "inbox.md", "Destination file[#selector] for the note")
	captureCmd.Flags().Bool("editor", false, "Edit the note in $EDITOR before saving")
	captureCmd.Flags().Bool("prepend", false, "Prepend the note instead of appending")
	captureCmd.Flags().Bool("json", false, "Output result as JSON")
	captureCmd.Flags().String("template", "", "Template to use for new notes (file path or template name)")
	captureCmd.Flags().String("content", "", "Note content (overrides stdin/template)")
	captureCmd.Flags().Bool("no-verify", false, "Skip hooks verification")
}

// CaptureContext holds all state for the capture pipeline.
type CaptureContext struct {
	WS         *workspace.Workspace
	Content    string
	File       string
	Selector   string
	Doc        *ast.Document
	Heading    *ast.Heading
	Subtree    []ast.Node
	Prepend    bool
	Editor     bool
	JSONOut    bool
	Template   string
	ResultPath string
	Err        error
	NoVerify   bool // skip hooks if true
}

func runCapture(cmd *cobra.Command, args []string) error {
	ctx := &CaptureContext{}

	// Parse flags and positional template
	ctx.Editor, _ = cmd.Flags().GetBool("editor")
	ctx.Prepend, _ = cmd.Flags().GetBool("prepend")
	ctx.JSONOut, _ = cmd.Flags().GetBool("json")
	ctx.NoVerify, _ = cmd.Flags().GetBool("no-verify")
	flagTemplate, _ := cmd.Flags().GetString("template")
	if len(args) > 0 && args[0] != "" {
		ctx.Template = args[0]
	} else {
		ctx.Template = flagTemplate
	}
	dest, _ := cmd.Flags().GetString("destination")
	if dest == "" {
		dest = "inbox.md"
	}
	ctx.File, ctx.Selector = parseDestination(dest)

	// --- Default to editor if template is used and --editor is not set ---
	if ctx.Template != "" && !ctx.Editor {
		ctx.Editor = true
	}

	// Pipeline steps
	if err := captureStepFindWorkspace(ctx); err != nil {
		return err
	}

	// --- Pre-capture hook ---
	if !ctx.NoVerify {
		preHook := hooks.NewPreCaptureHook(ctx.WS)
		newContent, err := preHook.Execute(ctx.Content, ctx.Template, false)
		if err != nil {
			return fmt.Errorf("pre-capture hook failed: %w", err)
		}
		ctx.Content = newContent
	}

	if err := captureStepGetContent(cmd, ctx); err != nil {
		return err
	}

	if err := captureStepApplyFrontmatter(ctx); err != nil {
		return err
	}
	if err := captureStepResolveFile(ctx); err != nil {
		return err
	}
	if err := captureStepReadOrInitDoc(ctx); err != nil {
		return err
	}
	if err := captureStepPrepareSubtree(ctx); err != nil {
		return err
	}
	if err := captureStepInsertSubtree(ctx); err != nil {
		return err
	}
	if err := captureStepWriteDoc(ctx); err != nil {
		return err
	}

	// --- Post-capture hook ---
	if !ctx.NoVerify {
		postHook := hooks.NewPostCaptureHook(ctx.WS)
		err := postHook.Execute(ctx.Content, ctx.Template, ctx.File, false)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: post-capture hook failed: %v\n", err)
		}
	}

	if ctx.JSONOut {
		captureStepOutputJSON(ctx)
	}
	return ctx.Err
}

// Apply frontmatter: override destination and prepend mode if present
func captureStepApplyFrontmatter(ctx *CaptureContext) error {
	fm := parseFrontmatter(ctx.Content, ctx.File, func() string {
		if ctx.Prepend {
			return "prepend"
		}
		return "append"
	}())
	if fm.Destination != "" && fm.Destination != ctx.File {
		ctx.File, ctx.Selector = parseDestination(fm.Destination)
	}
	if fm.RefileMode == "prepend" {
		ctx.Prepend = true
	} else if fm.RefileMode == "append" {
		ctx.Prepend = false
	}
	return nil
}

// --- Pipeline step functions (to be implemented/filled in) ---
func captureStepFindWorkspace(ctx *CaptureContext) error {
	ws, err := workspace.FindWorkspace()
	if err != nil {
		return err
	}
	ctx.WS = ws
	return nil
}

func captureStepGetContent(cmd *cobra.Command, ctx *CaptureContext) error {
	// 1. --content flag takes highest precedence
	contentFlag, _ := cmd.Flags().GetString("content")
	if contentFlag != "" {
		ctx.Content = contentFlag
		return nil
	}

	// 2. If a template is specified, load and render it
	if ctx.Template != "" {
		tm := template.NewManager(ctx.WS)
		t, err := tm.Get(ctx.Template)
		if err != nil {
			return fmt.Errorf("template error: %w", err)
		}
		rendered, err := tm.Render(t, "")
		if err != nil {
			return fmt.Errorf("template render error: %w", err)
		}
		// Strip frontmatter before editor
		ctx.Content = template.StripFrontmatter(rendered)
	} else {
		// 3. Otherwise, use stdin (piped input or prompt)
		stat, _ := os.Stdin.Stat()
		hasPipedInput := (stat.Mode() & os.ModeCharDevice) == 0
		if hasPipedInput {
			b, err := os.ReadFile("/dev/stdin")
			if err != nil {
				return err
			}
			ctx.Content = string(b)
			return nil // skip editor if piped input
		}
		content, err := getCaptureContent()
		if err != nil {
			return err
		}
		ctx.Content = content
	}

	// 4. Open editor if requested (after content is loaded)
	if ctx.Editor {
		edited, err := editor.OpenEditor(ctx.Content)
		if err != nil {
			return fmt.Errorf("failed to open editor: %w", err)
		}
		ctx.Content = edited
	}

	return nil
}

func captureStepResolveFile(ctx *CaptureContext) error {
	file, err := ctx.WS.ResolveFilePath(ctx.File, true)
	if err != nil {
		return err
	}
	ctx.File = file
	if err := ctx.WS.EnsureFileExists(file); err != nil {
		return err
	}
	return nil
}

func captureStepReadOrInitDoc(ctx *CaptureContext) error {
	destContent, err := os.ReadFile(ctx.File)
	if err != nil {
		return err
	}
	ctx.Doc = markdown.ParseMarkdownAST(destContent)
	return nil
}

func captureStepPrepareSubtree(ctx *CaptureContext) error {
	ctx.Subtree = markdown.ParseContentAsSubtree(ctx.Content)
	return nil
}

func captureStepInsertSubtree(ctx *CaptureContext) error {
	if ctx.Selector == "" {
		if ctx.Prepend {
			markdown.InsertSubtree(ctx.Doc, 0, ctx.Subtree)
		} else {
			markdown.InsertSubtree(ctx.Doc, len(ctx.Doc.GetChildren()), ctx.Subtree)
		}
	} else {
		sp := markdown.ParseSelector(ctx.Selector, false)
		h, _ := sp.FindOrCreateHeadingNode(ctx.Doc, 1)
		ctx.Heading = h
		// Determine parent heading level
		headingLevel := 1
		if h != nil {
			headingLevel = h.Level
		}
		if len(ctx.Subtree) > 0 {
			// Find minimum heading level in subtree
			minLevel := 7
			for _, n := range ctx.Subtree {
				if hd, ok := n.(*ast.Heading); ok {
					if hd.Level < minLevel {
						minLevel = hd.Level
					}
				}
			}
			if minLevel <= 6 {
				// Calculate level shift so top-level heading is one deeper than parent
				shift := headingLevel + 1 - minLevel
				if shift != 0 {
					markdown.RelevelSubtree(ctx.Subtree, shift)
				}
				// Clamp all heading levels to max 6
				for _, n := range ctx.Subtree {
					if hd, ok := n.(*ast.Heading); ok {
						if hd.Level > 6 {
							hd.Level = 6
						}
					}
				}
			}
		}
		if ctx.Prepend {
			markdown.InsertSubtree(h, 0, ctx.Subtree)
		} else {
			markdown.InsertSubtree(h, len(h.GetChildren()), ctx.Subtree)
		}
	}
	return nil
}

func captureStepWriteDoc(ctx *CaptureContext) error {
	newContent := markdown.RenderMarkdownAST(ctx.Doc)
	ctx.ResultPath = ctx.File
	return os.WriteFile(ctx.File, []byte(newContent), 0644)
}

func captureStepOutputJSON(ctx *CaptureContext) {
	fmt.Printf("{\"file\":%q,\"ok\":%v}\n", ctx.ResultPath, ctx.Err == nil)
}

func getCaptureContent() (string, error) {
	fmt.Print("Enter note content (end with Ctrl+D):\n")
	b, err := os.ReadFile("/dev/stdin")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func parseDestination(dest string) (file, selector string) {
	parts := strings.SplitN(dest, "#", 2)
	file = parts[0]
	if len(parts) > 1 {
		selector = parts[1]
	}
	return
}

// --- Frontmatter utilities ---

// CaptureFrontmatter holds supported frontmatter fields for capture
type CaptureFrontmatter struct {
	Destination string `yaml:"destination"`
	RefileMode  string `yaml:"refile_mode"`
}

// parseFrontmatter extracts supported fields from YAML frontmatter in content
func parseFrontmatter(content, fallbackDest, fallbackMode string) CaptureFrontmatter {
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

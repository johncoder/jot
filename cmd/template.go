package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/johncoder/jot/internal/cmdutil"
	"github.com/johncoder/jot/internal/editor"
	"github.com/johncoder/jot/internal/template"
	"github.com/johncoder/jot/internal/workspace"
	"github.com/spf13/cobra"
)

var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Manage note templates",
	Long: `Manage note templates for structured capture.

Templates are stored in .jot/templates/ and can contain shell commands
for dynamic content generation. Templates require explicit approval
before they can execute shell commands for security.

Examples:
  jot template list                # List all templates
  jot template new meeting         # Create new template
  jot template edit meeting        # Edit existing template
  jot template approve meeting     # Approve template for execution
  jot template render meeting      # Render template content`,
}

var templateListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available templates",
	Long:  `List all available templates and their approval status.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmdutil.StartCommand(cmd)

		ws, err := workspace.RequireWorkspace()
		if err != nil {
			if ctx.IsJSONOutput() {
				return ctx.HandleError(err)
			}
			return err
		}

		tm := template.NewManager(ws)
		templates, err := tm.List()
		if err != nil {
			err := fmt.Errorf("failed to list templates: %w", err)
			if ctx.IsJSONOutput() {
				return ctx.HandleError(err)
			}
			return err
		}

		if ctx.IsJSONOutput() {
			var templateItems []TemplateItem
			for _, t := range templates {
				templateItems = append(templateItems, TemplateItem{
					Name:     t.Name,
					Approved: t.Approved,
					Hash:     t.Hash,
				})
			}

			response := TemplateListResponse{
				Operation: "template_list",
				Templates: templateItems,
				Summary: TemplateListSummary{
					TotalTemplates:    len(templates),
					ApprovedTemplates: countApproved(templates),
				},
				Metadata: cmdutil.CreateJSONMetadata(ctx.Cmd, true, ctx.StartTime),
			}

			return cmdutil.OutputJSON(response)
		}

		if len(templates) == 0 {
			fmt.Println("No templates found. Create one with: jot template new <n>")
			return nil
		}

		fmt.Printf("Available templates:\n\n")
		for _, t := range templates {
			status := "✗ needs approval"
			if t.Approved {
				status = "✓ approved"
			}
			fmt.Printf("  %s (%s)\n", t.Name, status)
		}

		return nil
	},
}

var templateNewCmd = &cobra.Command{
	Use:   "new <n>",
	Short: "Create a new template",
	Long: `Create a new template and open it in your editor.

The template can contain shell commands using $(command) syntax:
  # Meeting Notes - $(date '+%Y-%m-%d')
  **Project:** $(git branch --show-current)

Templates require approval before shell commands can execute.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmdutil.StartCommand(cmd)

		ws, err := workspace.RequireWorkspace()
		if err != nil {
			if ctx.IsJSONOutput() {
				return ctx.HandleError(err)
			}
			return err
		}

		name := args[0]
		tm := template.NewManager(ws)

		// Default template content with frontmatter
		defaultContent := fmt.Sprintf(`---
destination: inbox.md
refile_mode: append
---
## %s - $(date '+%%Y-%%m-%%d %%H:%%M')


`, strings.ToTitle(name))

		// Create template
		pathUtil := cmdutil.NewPathUtil(ws)
		err = tm.Create(name, defaultContent)
		if err != nil {
			err := fmt.Errorf("failed to create template: %w", err)
			if ctx.IsJSONOutput() {
				return ctx.HandleError(err)
			}
			return err
		}

		templatePath := pathUtil.JotDirJoin(filepath.Join("templates", name+".md"))
		editorError := ""
		edited := false

		if !ctx.IsJSONOutput() {
			fmt.Printf("Created template '%s'\n", name)
		}

		// Open in editor (skip for JSON output to avoid interactive prompt)
		if !ctx.IsJSONOutput() {
			// Read the current template content using unified content utilities
			content, err := cmdutil.ReadFileContent(templatePath)
			if err != nil {
				if ctx.IsJSONOutput() {
					return ctx.HandleError(err)
				}
				return err
			}

			// Open in editor
			editedContent, err := editor.OpenEditor(string(content))
			if err != nil {
				editorError = err.Error()
				fmt.Printf("Template created but failed to open editor: %v\n", err)
				fmt.Printf("Edit manually: %s\n", templatePath)
			} else {
				// Write back the edited content using unified content utilities
				err = cmdutil.WriteFileContent(templatePath, []byte(editedContent))
				if err != nil {
					if ctx.IsJSONOutput() {
						return ctx.HandleError(err)
					}
					return err
				}
				edited = true
			}

			fmt.Printf("\nTo use this template, first approve it:\n")
			fmt.Printf("  jot template approve %s\n", name)
		}

		if ctx.IsJSONOutput() {
			nextSteps := []string{
				fmt.Sprintf("jot template approve %s", name),
				fmt.Sprintf("jot template edit %s", name),
			}

			response := TemplateCreateResponse{
				Operation:    "template_new",
				TemplateName: name,
				TemplatePath: templatePath,
				Created:      true,
				Edited:       edited,
				EditorError:  editorError,
				NextSteps:    nextSteps,
				Metadata:     cmdutil.CreateJSONMetadata(ctx.Cmd, true, ctx.StartTime),
			}

			return cmdutil.OutputJSON(response)
		}

		return nil
	},
}

var templateEditCmd = &cobra.Command{
	Use:   "edit <n>",
	Short: "Edit an existing template",
	Long:  `Edit an existing template in your editor. Changes will require re-approval.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmdutil.StartCommand(cmd)

		ws, err := workspace.RequireWorkspace()
		if err != nil {
			if ctx.IsJSONOutput() {
				return ctx.HandleError(err)
			}
			return err
		}

		name := args[0]
		tm := template.NewManager(ws)

		// Check if template exists
		_, err = tm.Get(name)
		pathUtil := cmdutil.NewPathUtil(ws)
		if err != nil {
			if ctx.IsJSONOutput() {
				return ctx.HandleError(err)
			}
			return err
		}

		templatePath := pathUtil.JotDirJoin(filepath.Join("templates", name+".md"))

		if ctx.IsJSONOutput() {
			// For JSON output, skip interactive editor and return success
			nextSteps := []string{
				fmt.Sprintf("jot template approve %s", name),
				fmt.Sprintf("Edit manually: %s", templatePath),
			}

			response := TemplateEditResponse{
				Operation:    "template_edit",
				TemplateName: name,
				TemplatePath: templatePath,
				Updated:      false,
				EditorError:  "Editor skipped in JSON mode",
				NextSteps:    nextSteps,
				Metadata:     cmdutil.CreateJSONMetadata(ctx.Cmd, true, ctx.StartTime),
			}

			return cmdutil.OutputJSON(response)
		}

		// Open in editor
		// Read the current template content using unified content utilities
		content, err := cmdutil.ReadFileContent(templatePath)
		if err != nil {
			return err
		}

		// Open in editor
		editedContent, err := editor.OpenEditor(string(content))
		if err != nil {
			return fmt.Errorf("failed to open editor: %w", err)
		}

		// Write back the edited content using unified content utilities
		err = cmdutil.WriteFileContent(templatePath, []byte(editedContent))
		if err != nil {
			return fmt.Errorf("failed to save template: %w", err)
		}

		fmt.Printf("Template '%s' updated. Re-approve if needed:\n", name)
		fmt.Printf("  jot template approve %s\n", name)

		return nil
	},
}

var templateApproveCmd = &cobra.Command{
	Use:   "approve <n>",
	Short: "Approve a template for execution",
	Long: `Approve a template to allow shell command execution.

This grants permission for the template to execute shell commands
like $(date) or $(git status). Approval is based on the template's
current content hash - any changes will require re-approval.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmdutil.StartCommand(cmd)

		ws, err := workspace.RequireWorkspace()
		if err != nil {
			if ctx.IsJSONOutput() {
				return ctx.HandleError(err)
			}
			return err
		}

		name := args[0]
		tm := template.NewManager(ws)

		// Get template to show what we're approving
		t, err := tm.Get(name)
		if err != nil {
			if ctx.IsJSONOutput() {
				return ctx.HandleError(err)
			}
			return err
		}

		if ctx.IsJSONOutput() {
			// For JSON output, we can't do interactive approval
			// Return an error or require a --force flag
			err := fmt.Errorf("interactive approval not supported in JSON mode")
			return ctx.HandleError(err)
		}

		// Show template content for review
		fmt.Printf("Approving template '%s':\n\n", name)
		fmt.Println(strings.Repeat("-", 50))
		fmt.Println(t.Content)
		fmt.Println(strings.Repeat("-", 50))
		fmt.Printf("\nThis will allow the template to execute shell commands.\n")
		fmt.Printf("Template hash: %s\n\n", t.Hash[:16]+"...")

		// Confirm approval
		confirmed, err := cmdutil.ConfirmOperation("Approve this template?")
		if err != nil {
			return err
		}

		if !confirmed {
			cmdutil.ShowInfo("Template not approved.")
			return nil
		}

		// Approve template
		err = tm.Approve(name)
		if err != nil {
			return fmt.Errorf("failed to approve template: %w", err)
		}

		fmt.Printf("Template '%s' approved for execution.\n", name)
		return nil
	},
}

var templateViewCmd = &cobra.Command{
	Use:   "view <n>",
	Short: "View the raw content of a template",
	Long:  `Display the raw content of a specified template.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmdutil.StartCommand(cmd)

		ws, err := workspace.RequireWorkspace()
		if err != nil {
			if ctx.IsJSONOutput() {
				return ctx.HandleError(err)
			}
			return err
		}

		name := args[0]
		tm := template.NewManager(ws)

		t, err := tm.Get(name)
		if err != nil {
			err := fmt.Errorf("failed to retrieve template: %w", err)
			if ctx.IsJSONOutput() {
				return ctx.HandleError(err)
			}
			return err
		}

		if ctx.IsJSONOutput() {
			response := TemplateViewResponse{
				Operation:    "template_view",
				TemplateName: name,
				Content:      t.Content,
				Approved:     t.Approved,
				Hash:         t.Hash,
				Metadata:     cmdutil.CreateJSONMetadata(ctx.Cmd, true, ctx.StartTime),
			}

			return cmdutil.OutputJSON(response)
		}

		fmt.Println(t.Content)
		return nil
	},
}

var templateRenderCmd = &cobra.Command{
	Use:   "render <n>",
	Short: "Render a template with shell command execution",
	Long: `Render a template and execute any shell commands within it.

This command outputs the fully rendered template content, executing
any shell commands like $(date) or $(git status). The template must
be approved before shell commands can execute.

Examples:
  jot template render meeting      # Render meeting template
  jot template render meeting --json  # Output rendered content as JSON`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmdutil.StartCommand(cmd)

		ws, err := workspace.RequireWorkspace()
		if err != nil {
			if ctx.IsJSONOutput() {
				return ctx.HandleError(err)
			}
			return err
		}

		name := args[0]
		tm := template.NewManager(ws)

		t, err := tm.Get(name)
		if err != nil {
			err := fmt.Errorf("failed to retrieve template: %w", err)
			if ctx.IsJSONOutput() {
				return ctx.HandleError(err)
			}
			return err
		}

		// Render the template (this will respect approval status)
		renderedContent, err := tm.Render(t, "")
		if err != nil {
			err := fmt.Errorf("failed to render template: %w", err)
			if ctx.IsJSONOutput() {
				return ctx.HandleError(err)
			}
			return err
		}

		if ctx.IsJSONOutput() {
			response := TemplateRenderResponse{
				Operation:        "template_render",
				TemplateName:     name,
				RenderedContent:  renderedContent,
				Approved:         t.Approved,
				ExecutionAllowed: t.Approved,
				Metadata:         cmdutil.CreateJSONMetadata(ctx.Cmd, true, ctx.StartTime),
			}

			return cmdutil.OutputJSON(response)
		}

		fmt.Print(renderedContent)
		return nil
	},
}

// Helper function to count approved templates
func countApproved(templates []template.Template) int {
	count := 0
	for _, t := range templates {
		if t.Approved {
			count++
		}
	}
	return count
}

// JSON response structures for template commands
type TemplateListResponse struct {
	Operation string               `json:"operation"`
	Templates []TemplateItem       `json:"templates"`
	Summary   TemplateListSummary  `json:"summary"`
	Metadata  cmdutil.JSONMetadata `json:"metadata"`
}

type TemplateItem struct {
	Name     string `json:"name"`
	Approved bool   `json:"approved"`
	Hash     string `json:"hash"`
}

type TemplateListSummary struct {
	TotalTemplates    int `json:"total_templates"`
	ApprovedTemplates int `json:"approved_templates"`
}

type TemplateCreateResponse struct {
	Operation    string               `json:"operation"`
	TemplateName string               `json:"template_name"`
	TemplatePath string               `json:"template_path"`
	Created      bool                 `json:"created"`
	Edited       bool                 `json:"edited"`
	EditorError  string               `json:"editor_error,omitempty"`
	NextSteps    []string             `json:"next_steps"`
	Metadata     cmdutil.JSONMetadata `json:"metadata"`
}

type TemplateEditResponse struct {
	Operation    string               `json:"operation"`
	TemplateName string               `json:"template_name"`
	TemplatePath string               `json:"template_path"`
	Updated      bool                 `json:"updated"`
	EditorError  string               `json:"editor_error,omitempty"`
	NextSteps    []string             `json:"next_steps"`
	Metadata     cmdutil.JSONMetadata `json:"metadata"`
}

type TemplateApproveResponse struct {
	Operation     string               `json:"operation"`
	TemplateName  string               `json:"template_name"`
	Approved      bool                 `json:"approved"`
	Hash          string               `json:"hash"`
	UserConfirmed bool                 `json:"user_confirmed"`
	Metadata      cmdutil.JSONMetadata `json:"metadata"`
}

type TemplateViewResponse struct {
	Operation    string               `json:"operation"`
	TemplateName string               `json:"template_name"`
	Content      string               `json:"content"`
	Approved     bool                 `json:"approved"`
	Hash         string               `json:"hash"`
	Metadata     cmdutil.JSONMetadata `json:"metadata"`
}

type TemplateRenderResponse struct {
	Operation        string               `json:"operation"`
	TemplateName     string               `json:"template_name"`
	RenderedContent  string               `json:"rendered_content"`
	Approved         bool                 `json:"approved"`
	ExecutionAllowed bool                 `json:"execution_allowed"`
	Metadata         cmdutil.JSONMetadata `json:"metadata"`
}

func init() {
	templateCmd.AddCommand(templateListCmd)
	templateCmd.AddCommand(templateNewCmd)
	templateCmd.AddCommand(templateEditCmd)
	templateCmd.AddCommand(templateApproveCmd)
	templateCmd.AddCommand(templateViewCmd)
	templateCmd.AddCommand(templateRenderCmd)
}

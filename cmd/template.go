package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
  jot template approve meeting     # Approve template for execution`,
}

var templateListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available templates",
	Long:  `List all available templates and their approval status.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := workspace.RequireWorkspace()
		if err != nil {
			return err
		}

		tm := template.NewManager(ws)
		templates, err := tm.List()
		if err != nil {
			return fmt.Errorf("failed to list templates: %w", err)
		}

		if len(templates) == 0 {
			fmt.Println("No templates found. Create one with: jot template new <name>")
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
	Use:   "new <name>",
	Short: "Create a new template",
	Long: `Create a new template and open it in your editor.

The template can contain shell commands using $(command) syntax:
  # Meeting Notes - $(date '+%Y-%m-%d')
  **Project:** $(git branch --show-current)

Templates require approval before shell commands can execute.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := workspace.RequireWorkspace()
		if err != nil {
			return err
		}

		name := args[0]
		tm := template.NewManager(ws)

		// Default template content
		defaultContent := fmt.Sprintf(`# %s - $(date '+%%Y-%%m-%%d %%H:%%M')

**Created:** $(date '+%%Y-%%m-%%d')

## Notes


## Action Items

`, strings.Title(name))

		// Create template
		err = tm.Create(name, defaultContent)
		if err != nil {
			return fmt.Errorf("failed to create template: %w", err)
		}

		fmt.Printf("Created template '%s'\n", name)

		// Open in editor
		templatePath := filepath.Join(ws.JotDir, "templates", name+".md")
		
		// Read the current template content
		content, err := os.ReadFile(templatePath)
		if err != nil {
			return fmt.Errorf("failed to read template file: %w", err)
		}
		
		// Open in editor
		editedContent, err := editor.OpenEditor(string(content))
		if err != nil {
			fmt.Printf("Template created but failed to open editor: %v\n", err)
			fmt.Printf("Edit manually: %s\n", templatePath)
		} else {
			// Write back the edited content
			err = os.WriteFile(templatePath, []byte(editedContent), 0644)
			if err != nil {
				return fmt.Errorf("failed to save template: %w", err)
			}
		}

		fmt.Printf("\nTo use this template, first approve it:\n")
		fmt.Printf("  jot template approve %s\n", name)

		return nil
	},
}

var templateEditCmd = &cobra.Command{
	Use:   "edit <name>",
	Short: "Edit an existing template",
	Long:  `Edit an existing template in your editor. Changes will require re-approval.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := workspace.RequireWorkspace()
		if err != nil {
			return err
		}

		name := args[0]
		tm := template.NewManager(ws)

		// Check if template exists
		_, err = tm.Get(name)
		if err != nil {
			return err
		}

		// Open in editor
		templatePath := filepath.Join(ws.JotDir, "templates", name+".md")
		
		// Read the current template content
		content, err := os.ReadFile(templatePath)
		if err != nil {
			return fmt.Errorf("failed to read template file: %w", err)
		}
		
		// Open in editor
		editedContent, err := editor.OpenEditor(string(content))
		if err != nil {
			return fmt.Errorf("failed to open editor: %w", err)
		}
		
		// Write back the edited content
		err = os.WriteFile(templatePath, []byte(editedContent), 0644)
		if err != nil {
			return fmt.Errorf("failed to save template: %w", err)
		}

		fmt.Printf("Template '%s' updated. Re-approve if needed:\n", name)
		fmt.Printf("  jot template approve %s\n", name)

		return nil
	},
}

var templateApproveCmd = &cobra.Command{
	Use:   "approve <name>",
	Short: "Approve a template for execution",
	Long: `Approve a template to allow shell command execution.

This grants permission for the template to execute shell commands
like $(date) or $(git status). Approval is based on the template's
current content hash - any changes will require re-approval.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := workspace.RequireWorkspace()
		if err != nil {
			return err
		}

		name := args[0]
		tm := template.NewManager(ws)

		// Get template to show what we're approving
		t, err := tm.Get(name)
		if err != nil {
			return err
		}

		// Show template content for review
		fmt.Printf("Approving template '%s':\n\n", name)
		fmt.Println(strings.Repeat("-", 50))
		fmt.Println(t.Content)
		fmt.Println(strings.Repeat("-", 50))
		fmt.Printf("\nThis will allow the template to execute shell commands.\n")
		fmt.Printf("Template hash: %s\n\n", t.Hash[:16]+"...")

		// Confirm approval
		fmt.Print("Approve this template? [y/N]: ")
		var response string
		fmt.Scanln(&response)

		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("Template not approved.")
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
	Use:   "view <name>",
	Short: "View the raw content of a template",
	Long:  `Display the raw content of a specified template.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := workspace.RequireWorkspace()
		if err != nil {
			return err
		}

		name := args[0]
		tm := template.NewManager(ws)

		t, err := tm.Get(name)
		if err != nil {
			return fmt.Errorf("failed to retrieve template: %w", err)
		}

		fmt.Println(t.Content)
		return nil
	},
}

func init() {
	templateCmd.AddCommand(templateListCmd)
	templateCmd.AddCommand(templateNewCmd)
	templateCmd.AddCommand(templateEditCmd)
	templateCmd.AddCommand(templateApproveCmd)
	templateCmd.AddCommand(templateViewCmd)
}

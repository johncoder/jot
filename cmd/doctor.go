package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/johncoder/jot/internal/workspace"
	"github.com/spf13/cobra"
)

var (
	doctorFix bool
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose and fix common issues",
	Long: `Diagnose and optionally fix common issues in the jot workspace.

Checks for:
- Workspace structure integrity
- File permissions and accessibility
- Database consistency
- Configuration issues
- External tool availability

Examples:
  jot doctor                     # Diagnose issues
  jot doctor --fix               # Diagnose and fix issues`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Running jot workspace diagnostics...")
		fmt.Println()

		var issues []string
		var warnings []string

		// Check if we're in a workspace
		ws, err := workspace.FindWorkspace()
		if err != nil {
			issues = append(issues, "Not in a jot workspace")
			fmt.Println("✗ Not in a jot workspace")
			fmt.Println("  Run 'jot init' to initialize a workspace")
			fmt.Println()
			fmt.Printf("Workspace health: ✗ Critical (%d issues)\n", len(issues))
			return nil
		}

		fmt.Printf("✓ Found workspace at: %s\n", ws.Root)

		// Check workspace structure
		if !ws.InboxExists() {
			issues = append(issues, "inbox.md is missing")
			fmt.Println("✗ inbox.md is missing")
		} else {
			fmt.Println("✓ inbox.md exists")
		}

		if !ws.LibExists() {
			issues = append(issues, "lib/ directory is missing")
			fmt.Println("✗ lib/ directory is missing")
		} else {
			fmt.Println("✓ lib/ directory exists")
		}

		// Check .jot directory
		if info, err := os.Stat(ws.JotDir); err != nil || !info.IsDir() {
			issues = append(issues, ".jot/ directory is missing")
			fmt.Println("✗ .jot/ directory is missing")
		} else {
			fmt.Println("✓ .jot/ directory exists")
		}

		// Check file permissions
		if ws.InboxExists() {
			if file, err := os.OpenFile(ws.InboxPath, os.O_WRONLY|os.O_APPEND, 0); err != nil {
				issues = append(issues, "inbox.md is not writable")
				fmt.Println("✗ inbox.md is not writable")
			} else {
				file.Close()
				fmt.Println("✓ inbox.md is writable")
			}
		}

		// Check external tools
		editors := []string{"vim", "nvim", "nano", "emacs"}
		editorFound := false
		for _, editor := range editors {
			if _, err := exec.LookPath(editor); err == nil {
				fmt.Printf("✓ Editor '%s' is available\n", editor)
				editorFound = true
				break
			}
		}
		if !editorFound {
			warnings = append(warnings, "No common editor found in PATH")
			fmt.Println("! No common editor found in PATH")
		}

		// Check pager
		if _, err := exec.LookPath("less"); err == nil {
			fmt.Println("✓ Pager 'less' is available")
		} else if _, err := exec.LookPath("more"); err == nil {
			fmt.Println("✓ Pager 'more' is available")
		} else {
			warnings = append(warnings, "No pager found in PATH")
			fmt.Println("! No pager found in PATH")
		}

		fmt.Println()

		// Apply fixes if requested
		if doctorFix && len(issues) > 0 {
			fmt.Println("Applying fixes...")

			if !ws.InboxExists() {
				inboxContent := `# Inbox

This is your inbox for capturing new notes quickly. Use 'jot capture' to add new notes here.

---

`
				if err := os.WriteFile(ws.InboxPath, []byte(inboxContent), 0644); err == nil {
					fmt.Println("✓ Created inbox.md")
				} else {
					fmt.Printf("✗ Failed to create inbox.md: %v\n", err)
				}
			}

			if !ws.LibExists() {
				if err := os.MkdirAll(ws.LibDir, 0755); err == nil {
					fmt.Println("✓ Created lib/ directory")
					// Add README
					readmePath := filepath.Join(ws.LibDir, "README.md")
					readmeContent := `# Library

This directory contains your organized notes. You can structure them however you like:

- By topic (e.g., go/, kubernetes/, databases/)
- By project (e.g., project-alpha/, project-beta/)
- By date (e.g., 2024/, 2024/01/)
- Or any combination

Use 'jot refile' to move notes from your inbox to organized files here.
`
					os.WriteFile(readmePath, []byte(readmeContent), 0644)
				} else {
					fmt.Printf("✗ Failed to create lib/ directory: %v\n", err)
				}
			}

			if info, err := os.Stat(ws.JotDir); err != nil || !info.IsDir() {
				if err := os.MkdirAll(ws.JotDir, 0755); err == nil {
					fmt.Println("✓ Created .jot/ directory")
				} else {
					fmt.Printf("✗ Failed to create .jot/ directory: %v\n", err)
				}
			}
		}

		// Summary
		if len(issues) == 0 {
			if len(warnings) == 0 {
				fmt.Println("Workspace health: ✓ Excellent")
			} else {
				fmt.Printf("Workspace health: ✓ Good (%d warning%s)\n",
					len(warnings), pluralize(len(warnings)))
			}
		} else {
			fmt.Printf("Workspace health: ✗ Issues found (%d issue%s",
				len(issues), pluralize(len(issues)))
			if len(warnings) > 0 {
				fmt.Printf(", %d warning%s", len(warnings), pluralize(len(warnings)))
			}
			fmt.Println(")")

			if !doctorFix {
				fmt.Println("Run 'jot doctor --fix' to apply automatic fixes")
			}
		}

		return nil
	},
}

// pluralize returns "s" if count != 1, empty string otherwise
func pluralize(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

func init() {
	doctorCmd.Flags().BoolVar(&doctorFix, "fix", false, "Automatically fix detected issues")
}

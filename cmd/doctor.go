package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

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
		startTime := time.Now()

		if !isJSONOutput(cmd) {
			fmt.Println("Running jot workspace diagnostics...")
			fmt.Println()
		}

		var issues []DoctorIssue
		var warnings []DoctorIssue
		var checks []DoctorCheck
		var fixes []DoctorFix

		// Check if we're in a workspace
		ws, err := workspace.FindWorkspace()
		if err != nil {
			issues = append(issues, DoctorIssue{
				Type:        "workspace",
				Message:     "Not in a jot workspace",
				Description: "Run 'jot init' to initialize a workspace",
				Severity:    "critical",
			})
			checks = append(checks, DoctorCheck{
				Name:    "workspace_detection",
				Status:  "failed",
				Message: "Not in a jot workspace",
			})

			if !isJSONOutput(cmd) {
				fmt.Println("✗ Not in a jot workspace")
				fmt.Println("  Run 'jot init' to initialize a workspace")
				fmt.Println()
				fmt.Printf("Workspace health: ✗ Critical (%d issues)\n", len(issues))
			} else {
				response := DoctorResponse{
					Operation:      "doctor",
					WorkspaceFound: false,
					HealthStatus:   "critical",
					Checks:         checks,
					Issues:         issues,
					Warnings:       warnings,
					FixesApplied:   fixes,
					Summary: DoctorSummary{
						TotalChecks:   len(checks),
						PassedChecks:  0,
						FailedChecks:  len(checks),
						IssuesFound:   len(issues),
						WarningsFound: len(warnings),
						FixesApplied:  len(fixes),
						OverallHealth: "critical",
					},
					Metadata: createJSONMetadata(cmd, true, startTime),
				}
				return outputJSON(response)
			}
			return nil
		}

		if !isJSONOutput(cmd) {
			fmt.Printf("✓ Found workspace at: %s\n", ws.Root)
		}
		checks = append(checks, DoctorCheck{
			Name:    "workspace_detection",
			Status:  "passed",
			Message: fmt.Sprintf("Found workspace at: %s", ws.Root),
		})

		// Check workspace structure
		if !ws.InboxExists() {
			issues = append(issues, DoctorIssue{
				Type:        "structure",
				Message:     "inbox.md is missing",
				Description: "The main inbox file for capturing notes is missing",
				Severity:    "high",
				Fixable:     true,
			})
			checks = append(checks, DoctorCheck{
				Name:    "inbox_exists",
				Status:  "failed",
				Message: "inbox.md is missing",
			})
			if !isJSONOutput(cmd) {
				fmt.Println("✗ inbox.md is missing")
			}
		} else {
			checks = append(checks, DoctorCheck{
				Name:    "inbox_exists",
				Status:  "passed",
				Message: "inbox.md exists",
			})
			if !isJSONOutput(cmd) {
				fmt.Println("✓ inbox.md exists")
			}
		}

		if !ws.LibExists() {
			issues = append(issues, DoctorIssue{
				Type:        "structure",
				Message:     "lib/ directory is missing",
				Description: "The directory for organized notes is missing",
				Severity:    "high",
				Fixable:     true,
			})
			checks = append(checks, DoctorCheck{
				Name:    "lib_exists",
				Status:  "failed",
				Message: "lib/ directory is missing",
			})
			if !isJSONOutput(cmd) {
				fmt.Println("✗ lib/ directory is missing")
			}
		} else {
			checks = append(checks, DoctorCheck{
				Name:    "lib_exists",
				Status:  "passed",
				Message: "lib/ directory exists",
			})
			if !isJSONOutput(cmd) {
				fmt.Println("✓ lib/ directory exists")
			}
		}

		// Check .jot directory
		if info, err := os.Stat(ws.JotDir); err != nil || !info.IsDir() {
			issues = append(issues, DoctorIssue{
				Type:        "structure",
				Message:     ".jot/ directory is missing",
				Description: "The internal data directory is missing",
				Severity:    "high",
				Fixable:     true,
			})
			checks = append(checks, DoctorCheck{
				Name:    "jot_dir_exists",
				Status:  "failed",
				Message: ".jot/ directory is missing",
			})
			if !isJSONOutput(cmd) {
				fmt.Println("✗ .jot/ directory is missing")
			}
		} else {
			checks = append(checks, DoctorCheck{
				Name:    "jot_dir_exists",
				Status:  "passed",
				Message: ".jot/ directory exists",
			})
			if !isJSONOutput(cmd) {
				fmt.Println("✓ .jot/ directory exists")
			}
		}

		// Check file permissions
		if ws.InboxExists() {
			if file, err := os.OpenFile(ws.InboxPath, os.O_WRONLY|os.O_APPEND, 0); err != nil {
				issues = append(issues, DoctorIssue{
					Type:        "permissions",
					Message:     "inbox.md is not writable",
					Description: "Cannot write to the inbox file",
					Severity:    "medium",
					Fixable:     false,
				})
				checks = append(checks, DoctorCheck{
					Name:    "inbox_writable",
					Status:  "failed",
					Message: "inbox.md is not writable",
				})
				if !isJSONOutput(cmd) {
					fmt.Println("✗ inbox.md is not writable")
				}
			} else {
				file.Close()
				checks = append(checks, DoctorCheck{
					Name:    "inbox_writable",
					Status:  "passed",
					Message: "inbox.md is writable",
				})
				if !isJSONOutput(cmd) {
					fmt.Println("✓ inbox.md is writable")
				}
			}
		}

		// Check external tools
		editors := []string{"vim", "nvim", "nano", "emacs"}
		editorFound := false
		var foundEditor string
		for _, editor := range editors {
			if _, err := exec.LookPath(editor); err == nil {
				foundEditor = editor
				editorFound = true
				break
			}
		}

		if editorFound {
			checks = append(checks, DoctorCheck{
				Name:    "editor_available",
				Status:  "passed",
				Message: fmt.Sprintf("Editor '%s' is available", foundEditor),
			})
			if !isJSONOutput(cmd) {
				fmt.Printf("✓ Editor '%s' is available\n", foundEditor)
			}
		} else {
			warnings = append(warnings, DoctorIssue{
				Type:        "external_tools",
				Message:     "No common editor found in PATH",
				Description: "Consider installing vim, nvim, nano, or emacs",
				Severity:    "low",
				Fixable:     false,
			})
			checks = append(checks, DoctorCheck{
				Name:    "editor_available",
				Status:  "warning",
				Message: "No common editor found in PATH",
			})
			if !isJSONOutput(cmd) {
				fmt.Println("! No common editor found in PATH")
			}
		}

		// Check pager
		var foundPager string
		if _, err := exec.LookPath("less"); err == nil {
			foundPager = "less"
		} else if _, err := exec.LookPath("more"); err == nil {
			foundPager = "more"
		}

		if foundPager != "" {
			checks = append(checks, DoctorCheck{
				Name:    "pager_available",
				Status:  "passed",
				Message: fmt.Sprintf("Pager '%s' is available", foundPager),
			})
			if !isJSONOutput(cmd) {
				fmt.Printf("✓ Pager '%s' is available\n", foundPager)
			}
		} else {
			warnings = append(warnings, DoctorIssue{
				Type:        "external_tools",
				Message:     "No pager found in PATH",
				Description: "Consider installing 'less' or ensure 'more' is available",
				Severity:    "low",
				Fixable:     false,
			})
			checks = append(checks, DoctorCheck{
				Name:    "pager_available",
				Status:  "warning",
				Message: "No pager found in PATH",
			})
			if !isJSONOutput(cmd) {
				fmt.Println("! No pager found in PATH")
			}
		}

		if !isJSONOutput(cmd) {
			fmt.Println()
		}

		// Apply fixes if requested
		if doctorFix && len(issues) > 0 {
			if !isJSONOutput(cmd) {
				fmt.Println("Applying fixes...")
			}

			// Fix missing inbox
			for _, issue := range issues {
				if issue.Type == "structure" && issue.Message == "inbox.md is missing" && issue.Fixable {
					inboxContent := `# Inbox

This is your inbox for capturing new notes quickly. Use 'jot capture' to add new notes here.

---

`
					if err := os.WriteFile(ws.InboxPath, []byte(inboxContent), 0644); err == nil {
						fixes = append(fixes, DoctorFix{
							Type:        "structure",
							Description: "Created inbox.md",
							Success:     true,
						})
						if !isJSONOutput(cmd) {
							fmt.Println("✓ Created inbox.md")
						}
					} else {
						fixes = append(fixes, DoctorFix{
							Type:        "structure",
							Description: "Failed to create inbox.md",
							Success:     false,
							Error:       err.Error(),
						})
						if !isJSONOutput(cmd) {
							fmt.Printf("✗ Failed to create inbox.md: %v\n", err)
						}
					}
				}

				// Fix missing lib directory
				if issue.Type == "structure" && issue.Message == "lib/ directory is missing" && issue.Fixable {
					if err := os.MkdirAll(ws.LibDir, 0755); err == nil {
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
						fixes = append(fixes, DoctorFix{
							Type:        "structure",
							Description: "Created lib/ directory",
							Success:     true,
						})
						if !isJSONOutput(cmd) {
							fmt.Println("✓ Created lib/ directory")
						}
					} else {
						fixes = append(fixes, DoctorFix{
							Type:        "structure",
							Description: "Failed to create lib/ directory",
							Success:     false,
							Error:       err.Error(),
						})
						if !isJSONOutput(cmd) {
							fmt.Printf("✗ Failed to create lib/ directory: %v\n", err)
						}
					}
				}

				// Fix missing .jot directory
				if issue.Type == "structure" && issue.Message == ".jot/ directory is missing" && issue.Fixable {
					if err := os.MkdirAll(ws.JotDir, 0755); err == nil {
						fixes = append(fixes, DoctorFix{
							Type:        "structure",
							Description: "Created .jot/ directory",
							Success:     true,
						})
						if !isJSONOutput(cmd) {
							fmt.Println("✓ Created .jot/ directory")
						}
					} else {
						fixes = append(fixes, DoctorFix{
							Type:        "structure",
							Description: "Failed to create .jot/ directory",
							Success:     false,
							Error:       err.Error(),
						})
						if !isJSONOutput(cmd) {
							fmt.Printf("✗ Failed to create .jot/ directory: %v\n", err)
						}
					}
				}
			}
		}

		// Calculate summary statistics
		passedChecks := 0
		failedChecks := 0
		for _, check := range checks {
			if check.Status == "passed" {
				passedChecks++
			} else if check.Status == "failed" {
				failedChecks++
			}
		}

		// Determine overall health
		var healthStatus string
		if len(issues) == 0 {
			if len(warnings) == 0 {
				healthStatus = "excellent"
			} else {
				healthStatus = "good"
			}
		} else {
			if failedChecks > 0 {
				healthStatus = "critical"
			} else {
				healthStatus = "issues"
			}
		}

		// Output results
		if isJSONOutput(cmd) {
			response := DoctorResponse{
				Operation:      "doctor",
				WorkspaceFound: true,
				WorkspaceRoot:  ws.Root,
				HealthStatus:   healthStatus,
				Checks:         checks,
				Issues:         issues,
				Warnings:       warnings,
				FixesApplied:   fixes,
				Summary: DoctorSummary{
					TotalChecks:   len(checks),
					PassedChecks:  passedChecks,
					FailedChecks:  failedChecks,
					IssuesFound:   len(issues),
					WarningsFound: len(warnings),
					FixesApplied:  len(fixes),
					OverallHealth: healthStatus,
				},
				Metadata: createJSONMetadata(cmd, true, startTime),
			}
			return outputJSON(response)
		}

		// Summary for non-JSON output
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

// JSON response structures for doctor command
type DoctorResponse struct {
	Operation      string        `json:"operation"`
	WorkspaceFound bool          `json:"workspace_found"`
	WorkspaceRoot  string        `json:"workspace_root,omitempty"`
	HealthStatus   string        `json:"health_status"` // "excellent", "good", "issues", "critical"
	Checks         []DoctorCheck `json:"checks"`
	Issues         []DoctorIssue `json:"issues"`
	Warnings       []DoctorIssue `json:"warnings"`
	FixesApplied   []DoctorFix   `json:"fixes_applied"`
	Summary        DoctorSummary `json:"summary"`
	Metadata       JSONMetadata  `json:"metadata"`
}

type DoctorCheck struct {
	Name    string `json:"name"`
	Status  string `json:"status"` // "passed", "failed", "warning"
	Message string `json:"message"`
}

type DoctorIssue struct {
	Type        string `json:"type"` // "workspace", "structure", "permissions", "external_tools"
	Message     string `json:"message"`
	Description string `json:"description"`
	Severity    string `json:"severity"` // "critical", "high", "medium", "low"
	Fixable     bool   `json:"fixable"`
}

type DoctorFix struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Success     bool   `json:"success"`
	Error       string `json:"error,omitempty"`
}

type DoctorSummary struct {
	TotalChecks   int    `json:"total_checks"`
	PassedChecks  int    `json:"passed_checks"`
	FailedChecks  int    `json:"failed_checks"`
	IssuesFound   int    `json:"issues_found"`
	WarningsFound int    `json:"warnings_found"`
	FixesApplied  int    `json:"fixes_applied"`
	OverallHealth string `json:"overall_health"`
}

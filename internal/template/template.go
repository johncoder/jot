package template

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/johncoder/jot/internal/workspace"
)

// Template represents a note template
// Added DestinationFile field to specify where notes should be stored
type Template struct {
	Name            string
	Path            string
	Content         string
	Hash            string
	Approved        bool
	DestinationFile string
}

// Manager handles template operations
type Manager struct {
	ws *workspace.Workspace
}

// NewManager creates a new template manager
func NewManager(ws *workspace.Workspace) *Manager {
	return &Manager{ws: ws}
}

// List returns all available templates
func (m *Manager) List() ([]Template, error) {
	templatesDir := filepath.Join(m.ws.JotDir, "templates")

	// Create templates directory if it doesn't exist
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create templates directory: %w", err)
	}

	var templates []Template

	err := filepath.Walk(templatesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't read
		}

		if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".md") {
			name := strings.TrimSuffix(info.Name(), ".md")
			content, err := os.ReadFile(path)
			if err != nil {
				return nil // Skip files we can't read
			}

			hash := calculateHash(string(content))
			approved := m.isApproved(hash)

			metadata := parseMetadata(string(content))
			templates = append(templates, Template{
				Name:            name,
				Path:            path,
				Content:         string(content),
				Hash:            hash,
				Approved:        approved,
				DestinationFile: metadata["destination_file"],
			})
		}
		return nil
	})

	return templates, err
}

// Get retrieves a specific template by name
func (m *Manager) Get(name string) (*Template, error) {
	templatePath := filepath.Join(m.ws.JotDir, "templates", name+".md")

	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("template '%s' not found", name)
	}

	content, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template: %w", err)
	}

	hash := calculateHash(string(content))
	approved := m.isApproved(hash)

	metadata := parseMetadata(string(content))
	destinationFile := metadata["destination_file"]
	if destinationFile == "" {
		destinationFile = "inbox.md"
	}
	return &Template{
		Name:            name,
		Path:            templatePath,
		Content:         string(content),
		Hash:            hash,
		Approved:        approved,
		DestinationFile: destinationFile,
	}, nil
}

// Create creates a new template
func (m *Manager) Create(name, content string) error {
	templatesDir := filepath.Join(m.ws.JotDir, "templates")

	// Create templates directory if it doesn't exist
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		return fmt.Errorf("failed to create templates directory: %w", err)
	}

	templatePath := filepath.Join(templatesDir, name+".md")

	// Check if template already exists
	if _, err := os.Stat(templatePath); !os.IsNotExist(err) {
		return fmt.Errorf("template '%s' already exists", name)
	}

	metadata := parseMetadata(content)
	destinationFile := metadata["destination_file"]
	if destinationFile == "" {
		destinationFile = "inbox.md"
	}

	// Update the metadata to include the default destination_file
	metadata["destination_file"] = destinationFile

	return os.WriteFile(templatePath, []byte(content), 0644)
}

// Approve grants permission for a template to execute shell commands
func (m *Manager) Approve(name string) error {
	template, err := m.Get(name)
	if err != nil {
		return err
	}

	permissionsFile := filepath.Join(m.ws.JotDir, "template_permissions")

	// Read existing permissions
	permissions := make(map[string]bool)
	if content, err := os.ReadFile(permissionsFile); err == nil {
		scanner := bufio.NewScanner(strings.NewReader(string(content)))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" && !strings.HasPrefix(line, "#") {
				permissions[line] = true
			}
		}
	}

	// Add new permission
	permissions[template.Hash] = true

	// Write updated permissions
	var lines []string
	lines = append(lines, "# Template permissions - SHA256 hashes of approved templates")
	lines = append(lines, fmt.Sprintf("# Template: %s", name))
	lines = append(lines, template.Hash)

	for hash := range permissions {
		if hash != template.Hash {
			lines = append(lines, hash)
		}
	}

	content := strings.Join(lines, "\n") + "\n"
	return os.WriteFile(permissionsFile, []byte(content), 0644)
}

// Render processes a template with shell command execution and content injection
func (m *Manager) Render(template *Template, appendContent string) (string, error) {
	if !template.Approved {
		return "", fmt.Errorf("template '%s' requires approval before use. Run: jot template approve %s", template.Name, template.Name)
	}

	content := template.Content

	// Execute shell commands
	content, err := m.executeShellCommands(content)
	if err != nil {
		return "", fmt.Errorf("failed to execute shell commands in template: %w", err)
	}

	// Append content if provided
	if appendContent != "" {
		content += "\n\n" + appendContent
	}

	return content, nil
}

// executeShellCommands finds and executes shell commands in the template
func (m *Manager) executeShellCommands(content string) (string, error) {
	// Match shell command syntax: $(command)
	re := regexp.MustCompile(`\$\(([^)]+)\)`)

	result := re.ReplaceAllStringFunc(content, func(match string) string {
		// Extract command (remove $( and ))
		command := match[2 : len(match)-1]

		// Execute command
		cmd := exec.Command("sh", "-c", command)
		cmd.Dir = m.ws.Root

		output, err := cmd.Output()
		if err != nil {
			// Return original if command fails
			return match
		}

		return strings.TrimSpace(string(output))
	})

	return result, nil
}

// isApproved checks if a template hash is approved
func (m *Manager) isApproved(hash string) bool {
	permissionsFile := filepath.Join(m.ws.JotDir, "template_permissions")

	content, err := os.ReadFile(permissionsFile)
	if err != nil {
		return false
	}

	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == hash {
			return true
		}
	}

	return false
}

// calculateHash computes SHA256 hash of template content
func calculateHash(content string) string {
	hash := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", hash)
}

// parseMetadata extracts metadata from template content
func parseMetadata(content string) map[string]string {
	metadata := make(map[string]string)
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			metadata[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return metadata
}

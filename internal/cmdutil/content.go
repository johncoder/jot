package cmdutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/johncoder/jot/internal/workspace"
	"gopkg.in/yaml.v3"
)

// MarkdownContent represents parsed markdown content with metadata
type MarkdownContent struct {
	Raw      string                 `json:"raw"`
	Metadata map[string]interface{} `json:"metadata"`
	Body     string                 `json:"body"`
	FilePath string                 `json:"file_path"`
	ModTime  time.Time              `json:"mod_time"`
}

// FileOperator handles file operations with workspace context
type FileOperator struct {
	workspace   *workspace.Workspace
	noWorkspace bool
}

// NewFileOperator creates a new file operator
func NewFileOperator(ws *workspace.Workspace, noWorkspace bool) *FileOperator {
	return &FileOperator{
		workspace:   ws,
		noWorkspace: noWorkspace,
	}
}

// ReadMarkdownFile reads and parses a markdown file with metadata
func (f *FileOperator) ReadMarkdownFile(path string) (*MarkdownContent, error) {
	// Resolve path using existing path resolution
	resolvedPath := ResolvePath(f.workspace, path, f.noWorkspace)

	// Read file content
	content, err := os.ReadFile(resolvedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", resolvedPath, err)
	}

	// Get file info for mod time
	info, err := os.Stat(resolvedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info for %s: %w", resolvedPath, err)
	}

	// Parse metadata and body
	metadata, body := ParseMarkdownMetadata(string(content))

	return &MarkdownContent{
		Raw:      string(content),
		Metadata: metadata,
		Body:     body,
		FilePath: resolvedPath,
		ModTime:  info.ModTime(),
	}, nil
}

// WriteMarkdownFile writes markdown content to a file
func (f *FileOperator) WriteMarkdownFile(path string, content *MarkdownContent) error {
	// Resolve path using existing path resolution
	resolvedPath := ResolvePath(f.workspace, path, f.noWorkspace)

	// Create directory if it doesn't exist
	dir := filepath.Dir(resolvedPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write file
	if err := os.WriteFile(resolvedPath, []byte(content.Raw), 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", resolvedPath, err)
	}

	return nil
}

// BackupFile creates a backup of a file before modification
func (f *FileOperator) BackupFile(path string) (string, error) {
	resolvedPath := ResolvePath(f.workspace, path, f.noWorkspace)

	// Check if file exists
	if _, err := os.Stat(resolvedPath); os.IsNotExist(err) {
		return "", nil // No backup needed for non-existent file
	}

	// Create backup filename with timestamp
	timestamp := time.Now().Format("20060102-150405")
	backupPath := resolvedPath + ".backup." + timestamp

	// Read original content
	content, err := os.ReadFile(resolvedPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file for backup %s: %w", resolvedPath, err)
	}

	// Write backup
	if err := os.WriteFile(backupPath, content, 0644); err != nil {
		return "", fmt.Errorf("failed to create backup %s: %w", backupPath, err)
	}

	return backupPath, nil
}

// AppendToFile appends content to a file, creating it if it doesn't exist
func (f *FileOperator) AppendToFile(path, content string) error {
	resolvedPath := ResolvePath(f.workspace, path, f.noWorkspace)

	// Create directory if it doesn't exist
	dir := filepath.Dir(resolvedPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Open file for appending, create if doesn't exist
	file, err := os.OpenFile(resolvedPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file for appending %s: %w", resolvedPath, err)
	}
	defer file.Close()

	// Write content
	if _, err := file.WriteString(content); err != nil {
		return fmt.Errorf("failed to append to file %s: %w", resolvedPath, err)
	}

	return nil
}

// FileExists checks if a file exists
func (f *FileOperator) FileExists(path string) bool {
	resolvedPath := ResolvePath(f.workspace, path, f.noWorkspace)
	_, err := os.Stat(resolvedPath)
	return err == nil
}

// ParseMarkdownMetadata extracts YAML frontmatter and body from markdown content
func ParseMarkdownMetadata(content string) (map[string]interface{}, string) {
	metadata := make(map[string]interface{})

	// Check for YAML frontmatter
	if strings.HasPrefix(content, "---\n") {
		parts := strings.SplitN(content, "\n---\n", 2)
		if len(parts) == 2 {
			// Parse YAML frontmatter
			yamlContent := parts[0][4:] // Remove the leading "---\n"
			if err := yaml.Unmarshal([]byte(yamlContent), &metadata); err == nil {
				return metadata, strings.TrimSpace(parts[1])
			}
		}
	}

	// Fallback to simple key:value parsing for backward compatibility
	lines := strings.Split(content, "\n")
	bodyStartIndex := 0

	for i, line := range lines {
		if strings.HasPrefix(line, "#") {
			bodyStartIndex = i
			break
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if key != "" && value != "" {
				metadata[key] = value
				bodyStartIndex = i + 1
			}
		} else if strings.TrimSpace(line) == "" {
			continue
		} else {
			bodyStartIndex = i
			break
		}
	}

	// Return remaining content as body
	body := strings.Join(lines[bodyStartIndex:], "\n")
	return metadata, strings.TrimSpace(body)
}

// ValidateMarkdownFile performs basic validation on a markdown file
func ValidateMarkdownFile(path string) error {
	// Check file extension
	if !strings.HasSuffix(strings.ToLower(path), ".md") {
		return fmt.Errorf("file %s is not a markdown file", path)
	}

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("file %s does not exist", path)
	}

	// Check if file is readable
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("cannot read file %s: %w", path, err)
	}
	file.Close()

	return nil
}

// IsMarkdownFile checks if a path appears to be a markdown file
func IsMarkdownFile(path string) bool {
	return strings.HasSuffix(strings.ToLower(path), ".md")
}

// CreateMarkdownContent creates a new MarkdownContent struct
func CreateMarkdownContent(raw string, metadata map[string]interface{}, body string) *MarkdownContent {
	return &MarkdownContent{
		Raw:      raw,
		Metadata: metadata,
		Body:     body,
		FilePath: "",
		ModTime:  time.Now(),
	}
}

// BuildMarkdownWithFrontmatter builds markdown content with YAML frontmatter
func BuildMarkdownWithFrontmatter(metadata map[string]interface{}, body string) (string, error) {
	var content strings.Builder

	if len(metadata) > 0 {
		content.WriteString("---\n")
		yamlData, err := yaml.Marshal(metadata)
		if err != nil {
			return "", fmt.Errorf("failed to marshal metadata: %w", err)
		}
		content.Write(yamlData)
		content.WriteString("---\n\n")
	}

	content.WriteString(body)
	return content.String(), nil
}

// ContentProcessor provides utilities for processing different types of content
type ContentProcessor struct {
	fileOp *FileOperator
}

// NewContentProcessor creates a new content processor
func NewContentProcessor(ws *workspace.Workspace, noWorkspace bool) *ContentProcessor {
	return &ContentProcessor{
		fileOp: NewFileOperator(ws, noWorkspace),
	}
}

// ProcessTemplateContent processes template content for rendering
func (cp *ContentProcessor) ProcessTemplateContent(templatePath string) (*MarkdownContent, error) {
	content, err := cp.fileOp.ReadMarkdownFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template: %w", err)
	}

	// Validate template content
	if err := cp.validateTemplateContent(content); err != nil {
		return nil, fmt.Errorf("invalid template content: %w", err)
	}

	return content, nil
}

// validateTemplateContent performs template-specific validation
func (cp *ContentProcessor) validateTemplateContent(content *MarkdownContent) error {
	// Check for required template metadata
	if content.Metadata == nil {
		return fmt.Errorf("template missing metadata")
	}

	// Validate template has some content
	if strings.TrimSpace(content.Body) == "" {
		return fmt.Errorf("template body is empty")
	}

	return nil
}

// ReadFileContent reads file content with unified error handling
func ReadFileContent(path string) ([]byte, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}
	return content, nil
}

// WriteFileContent writes file content with unified error handling
func WriteFileContent(path string, content []byte) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	if err := os.WriteFile(path, content, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}
	return nil
}

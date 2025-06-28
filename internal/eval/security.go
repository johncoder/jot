package eval

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/johncoder/jot/internal/workspace"
	"github.com/spf13/viper"
)

// SecurityConfig represents a path-based security configuration
type SecurityConfig struct {
	Path            string `json:"path"`
	RequireApproval bool   `json:"require_approval"`
	DefaultMode     string `json:"default_mode"` // "hash", "prompt", "always"
}

// ApprovalMode represents the different approval modes
type ApprovalMode string

const (
	ApprovalModeHash   ApprovalMode = "hash"
	ApprovalModePrompt ApprovalMode = "prompt"
	ApprovalModeAlways ApprovalMode = "always"
)

// ApprovalRecord represents an approved code block
type ApprovalRecord struct {
	Hash       string       `json:"hash"`
	Mode       ApprovalMode `json:"mode"`
	FilePath   string       `json:"file_path"`
	BlockName  string       `json:"block_name"`
	ApprovedAt string       `json:"approved_at"`
}

// DocumentApprovalRecord represents an approved document
type DocumentApprovalRecord struct {
	FilePath   string       `json:"file_path"`
	Mode       ApprovalMode `json:"mode"`
	ApprovedAt string       `json:"approved_at"`
}

// SecurityManager manages code block approvals and security policies
type SecurityManager struct {
	configPath    string
	approvals     map[string]*ApprovalRecord
	docApprovals  map[string]*DocumentApprovalRecord
	docConfigPath string
}

// NewSecurityManager creates a new security manager
func NewSecurityManager() (*SecurityManager, error) {
	sm := &SecurityManager{
		approvals:    make(map[string]*ApprovalRecord),
		docApprovals: make(map[string]*DocumentApprovalRecord),
	}

	// Get workspace using the standard workspace resolution
	ws, err := workspace.RequireWorkspace()
	if err != nil {
		return nil, fmt.Errorf("could not find workspace: %w", err)
	}

	sm.configPath = filepath.Join(ws.JotDir, "eval_permissions")
	sm.docConfigPath = filepath.Join(ws.JotDir, "eval_document_permissions")

	// Ensure .jot directory exists
	if err := os.MkdirAll(ws.JotDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create .jot directory: %w", err)
	}

	// Load existing approvals
	if err := sm.loadApprovals(); err != nil {
		return nil, fmt.Errorf("failed to load approvals: %w", err)
	}

	if err := sm.loadDocumentApprovals(); err != nil {
		return nil, fmt.Errorf("failed to load document approvals: %w", err)
	}

	return sm, nil
}

// loadApprovals loads approval records from disk
func (sm *SecurityManager) loadApprovals() error {
	if _, err := os.Stat(sm.configPath); os.IsNotExist(err) {
		return nil // No approvals file yet, that's OK
	}

	data, err := os.ReadFile(sm.configPath)
	if err != nil {
		return err
	}

	var approvals []*ApprovalRecord
	if err := json.Unmarshal(data, &approvals); err != nil {
		return err
	}

	for _, approval := range approvals {
		key := sm.makeApprovalKey(approval.FilePath, approval.BlockName)
		sm.approvals[key] = approval
	}

	return nil
}

// saveApprovals saves approval records to disk
func (sm *SecurityManager) saveApprovals() error {
	var approvals []*ApprovalRecord
	for _, approval := range sm.approvals {
		approvals = append(approvals, approval)
	}

	data, err := json.MarshalIndent(approvals, "", "  ")
	if err != nil {
		return err
	}

	// Ensure .jot directory exists
	if err := os.MkdirAll(filepath.Dir(sm.configPath), 0755); err != nil {
		return err
	}

	return os.WriteFile(sm.configPath, data, 0644)
}

// makeApprovalKey creates a unique key for an approval record
func (sm *SecurityManager) makeApprovalKey(filePath, blockName string) string {
	return fmt.Sprintf("%s:%s", filePath, blockName)
}

// hashCodeBlock creates a SHA256 hash of the code block content
func (sm *SecurityManager) hashCodeBlock(block *CodeBlock) string {
	content := strings.Join(block.Code, "\n")
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}

// getSecurityConfig gets the security configuration for a given file path
func (sm *SecurityManager) getSecurityConfig(filePath string) (*SecurityConfig, error) {
	configs, err := sm.loadSecurityConfigs(filePath)
	if err != nil {
		return nil, err
	}

	// Find the most specific matching configuration
	var bestMatch *SecurityConfig
	var bestMatchLen int

	for _, config := range configs {
		if matched, err := filepath.Match(config.Path, filePath); err == nil && matched {
			if len(config.Path) > bestMatchLen {
				bestMatch = config
				bestMatchLen = len(config.Path)
			}
		}
	}

	// Return default config if no match found
	if bestMatch == nil {
		return &SecurityConfig{
			Path:            "*",
			RequireApproval: true,
			DefaultMode:     string(ApprovalModeHash),
		}, nil
	}

	return bestMatch, nil
}

// loadSecurityConfigs loads security configurations from the directory hierarchy
func (sm *SecurityManager) loadSecurityConfigs(filePath string) ([]*SecurityConfig, error) {
	var configs []*SecurityConfig

	// Get absolute path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, err
	}

	// Walk up the directory tree looking for .jotrc files
	dir := filepath.Dir(absPath)
	for {
		jotrcPath := filepath.Join(dir, ".jotrc")
		if _, err := os.Stat(jotrcPath); err == nil {
			dirConfigs, err := sm.loadSecurityConfigFromFile(jotrcPath)
			if err == nil {
				// Prepend to give precedence to closer files
				configs = append(dirConfigs, configs...)
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break // reached root
		}
		dir = parent
	}

	return configs, nil
}

// loadSecurityConfigFromFile loads security config from a single .jotrc file
func (sm *SecurityManager) loadSecurityConfigFromFile(configPath string) ([]*SecurityConfig, error) {
	v := viper.New()
	v.SetConfigFile(configPath)

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var configs []*SecurityConfig
	if err := v.UnmarshalKey("eval.security", &configs); err != nil {
		return nil, err
	}

	// Set defaults for any missing fields
	for _, config := range configs {
		if config.DefaultMode == "" {
			config.DefaultMode = string(ApprovalModeHash)
		}
	}

	return configs, nil
}

// CheckApproval checks if a code block is approved for execution
func (sm *SecurityManager) CheckApproval(filePath string, block *CodeBlock) (bool, error) {
	if block.Eval == nil || block.Eval.Params["name"] == "" {
		return false, fmt.Errorf("code block has no name")
	}

	blockName := block.Eval.Params["name"]

	// First check if the entire document is approved
	docApproved, docMode, err := sm.CheckDocumentApproval(filePath)
	if err != nil {
		return false, err
	}

	if docApproved {
		// For document approval, we bypass individual block checks
		// but still respect the document's approval mode
		if docMode == ApprovalModeAlways {
			return true, nil
		}
		// For hash and prompt modes at document level, we still check individual blocks
		// but they are auto-approved if the document is approved
	}

	// Get security configuration
	secConfig, err := sm.getSecurityConfig(filePath)
	if err != nil {
		return false, err
	}

	// If approval is not required, allow execution
	if !secConfig.RequireApproval {
		return true, nil
	}

	// If document is approved, individual blocks are automatically approved
	if docApproved {
		return true, nil
	}

	// Check if block is approved
	key := sm.makeApprovalKey(filePath, blockName)
	approval, exists := sm.approvals[key]
	if !exists {
		return false, nil
	}

	// For hash mode, verify the content hasn't changed
	if approval.Mode == ApprovalModeHash {
		currentHash := sm.hashCodeBlock(block)
		if currentHash != approval.Hash {
			return false, nil // Content changed, re-approval needed
		}
	}

	return true, nil
}

// ApproveBlock approves a code block for execution
func (sm *SecurityManager) ApproveBlock(filePath string, block *CodeBlock, mode ApprovalMode) error {
	if block.Eval == nil || block.Eval.Params["name"] == "" {
		return fmt.Errorf("code block has no name")
	}

	blockName := block.Eval.Params["name"]
	hash := sm.hashCodeBlock(block)

	approval := &ApprovalRecord{
		Hash:       hash,
		Mode:       mode,
		FilePath:   filePath,
		BlockName:  blockName,
		ApprovedAt: time.Now().Format(time.RFC3339),
	}

	key := sm.makeApprovalKey(filePath, blockName)
	sm.approvals[key] = approval

	return sm.saveApprovals()
}

// RevokeApproval removes approval for a code block
func (sm *SecurityManager) RevokeApproval(filePath, blockName string) error {
	key := sm.makeApprovalKey(filePath, blockName)
	delete(sm.approvals, key)
	return sm.saveApprovals()
}

// ListApprovals returns all approval records
func (sm *SecurityManager) ListApprovals() []*ApprovalRecord {
	var approvals []*ApprovalRecord
	for _, approval := range sm.approvals {
		approvals = append(approvals, approval)
	}
	return approvals
}

// loadDocumentApprovals loads document approval records from disk
func (sm *SecurityManager) loadDocumentApprovals() error {
	if _, err := os.Stat(sm.docConfigPath); os.IsNotExist(err) {
		return nil // No document approvals file yet, that's OK
	}

	data, err := os.ReadFile(sm.docConfigPath)
	if err != nil {
		return err
	}

	var docApprovals []*DocumentApprovalRecord
	if err := json.Unmarshal(data, &docApprovals); err != nil {
		return err
	}

	for _, approval := range docApprovals {
		sm.docApprovals[approval.FilePath] = approval
	}

	return nil
}

// saveDocumentApprovals saves document approval records to disk
func (sm *SecurityManager) saveDocumentApprovals() error {
	var docApprovals []*DocumentApprovalRecord
	for _, approval := range sm.docApprovals {
		docApprovals = append(docApprovals, approval)
	}

	data, err := json.MarshalIndent(docApprovals, "", "  ")
	if err != nil {
		return err
	}

	// Ensure .jot directory exists
	if err := os.MkdirAll(filepath.Dir(sm.docConfigPath), 0755); err != nil {
		return err
	}

	return os.WriteFile(sm.docConfigPath, data, 0644)
}

// CheckDocumentApproval checks if a document is approved for execution
func (sm *SecurityManager) CheckDocumentApproval(filePath string) (bool, ApprovalMode, error) {
	approval, exists := sm.docApprovals[filePath]
	if !exists {
		return false, ApprovalModeHash, nil
	}

	return true, approval.Mode, nil
}

// ApproveDocument approves an entire document for execution
func (sm *SecurityManager) ApproveDocument(filePath string, mode ApprovalMode) error {
	approval := &DocumentApprovalRecord{
		FilePath:   filePath,
		Mode:       mode,
		ApprovedAt: time.Now().Format(time.RFC3339),
	}

	sm.docApprovals[filePath] = approval
	return sm.saveDocumentApprovals()
}

// RevokeDocumentApproval removes document approval
func (sm *SecurityManager) RevokeDocumentApproval(filePath string) error {
	delete(sm.docApprovals, filePath)
	return sm.saveDocumentApprovals()
}

// ListDocumentApprovals returns all document approval records
func (sm *SecurityManager) ListDocumentApprovals() []*DocumentApprovalRecord {
	var approvals []*DocumentApprovalRecord
	for _, approval := range sm.docApprovals {
		approvals = append(approvals, approval)
	}
	return approvals
}

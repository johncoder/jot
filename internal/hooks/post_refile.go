package hooks

import (
	"time"

	"github.com/johncoder/jot/internal/workspace"
)

// PostRefileHook handles post-refile hook execution
type PostRefileHook struct {
	manager *Manager
}

// NewPostRefileHook creates a new post-refile hook handler
func NewPostRefileHook(ws *workspace.Workspace) *PostRefileHook {
	return &PostRefileHook{
		manager: NewManager(ws),
	}
}

// Execute runs the post-refile hook with the given context
func (h *PostRefileHook) Execute(content, sourceFile, destPath string, allowBypass bool) error {
	ctx := &HookContext{
		Type:        PostRefile,
		Workspace:   h.manager.workspace,
		Content:     content,
		SourceFile:  sourceFile,
		DestPath:    destPath,
		Timeout:     h.manager.timeout,
		AllowBypass: allowBypass,
		ExtraEnv: map[string]string{
			"JOT_REFILE_SOURCE": sourceFile,
			"JOT_REFILE_DEST":   destPath,
		},
	}

	result, err := h.manager.Execute(ctx)
	if err != nil {
		return err
	}

	// Post hooks can't abort the operation since it's already completed
	// But we can log errors or warnings
	if result.Aborted && result.Error != nil {
		// Log the error but don't fail the refile
		// TODO: Add logging framework
		return nil
	}

	return nil
}

// ExecuteWithTimeout runs the post-refile hook with a custom timeout
func (h *PostRefileHook) ExecuteWithTimeout(content, sourceFile, destPath string, timeout time.Duration, allowBypass bool) error {
	ctx := &HookContext{
		Type:        PostRefile,
		Workspace:   h.manager.workspace,
		Content:     content,
		SourceFile:  sourceFile,
		DestPath:    destPath,
		Timeout:     timeout,
		AllowBypass: allowBypass,
		ExtraEnv: map[string]string{
			"JOT_REFILE_SOURCE": sourceFile,
			"JOT_REFILE_DEST":   destPath,
		},
	}

	_, err := h.manager.Execute(ctx)
	if err != nil {
		return err
	}

	return nil
}

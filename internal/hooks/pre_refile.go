package hooks

import (
	"time"

	"github.com/johncoder/jot/internal/workspace"
)

// PreRefileHook handles pre-refile hook execution
type PreRefileHook struct {
	manager *Manager
}

// NewPreRefileHook creates a new pre-refile hook handler
func NewPreRefileHook(ws *workspace.Workspace) *PreRefileHook {
	return &PreRefileHook{
		manager: NewManager(ws),
	}
}

// Execute runs the pre-refile hook with the given content and context
func (h *PreRefileHook) Execute(content, sourceFile, destPath string, allowBypass bool) (string, error) {
	ctx := &HookContext{
		Type:        PreRefile,
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
		return content, err
	}

	if result.Aborted {
		return content, result.Error
	}

	return result.Content, nil
}

// ExecuteWithTimeout runs the pre-refile hook with a custom timeout
func (h *PreRefileHook) ExecuteWithTimeout(content, sourceFile, destPath string, timeout time.Duration, allowBypass bool) (string, error) {
	ctx := &HookContext{
		Type:        PreRefile,
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

	result, err := h.manager.Execute(ctx)
	if err != nil {
		return content, err
	}

	if result.Aborted {
		return content, result.Error
	}

	return result.Content, nil
}

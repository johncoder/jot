package hooks

import (
	"time"

	"github.com/johncoder/jot/internal/workspace"
)

// PreCaptureHook handles pre-capture hook execution
type PreCaptureHook struct {
	manager *Manager
}

// NewPreCaptureHook creates a new pre-capture hook handler
func NewPreCaptureHook(ws *workspace.Workspace) *PreCaptureHook {
	return &PreCaptureHook{
		manager: NewManager(ws),
	}
}

// Execute runs the pre-capture hook with the given content and context
func (h *PreCaptureHook) Execute(content, templateName string, allowBypass bool) (string, error) {
	ctx := &HookContext{
		Type:         PreCapture,
		Workspace:    h.manager.workspace,
		Content:      content,
		TemplateName: templateName,
		Timeout:      h.manager.timeout,
		AllowBypass:  allowBypass,
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

// ExecuteWithTimeout runs the pre-capture hook with a custom timeout
func (h *PreCaptureHook) ExecuteWithTimeout(content, templateName string, timeout time.Duration, allowBypass bool) (string, error) {
	ctx := &HookContext{
		Type:         PreCapture,
		Workspace:    h.manager.workspace,
		Content:      content,
		TemplateName: templateName,
		Timeout:      timeout,
		AllowBypass:  allowBypass,
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

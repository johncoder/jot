package hooks

import (
	"time"

	"github.com/johncoder/jot/internal/workspace"
)

// PostCaptureHook handles post-capture hook execution
type PostCaptureHook struct {
	manager *Manager
}

// NewPostCaptureHook creates a new post-capture hook handler
func NewPostCaptureHook(ws *workspace.Workspace) *PostCaptureHook {
	return &PostCaptureHook{
		manager: NewManager(ws),
	}
}

// Execute runs the post-capture hook with the given context
func (h *PostCaptureHook) Execute(content, templateName, targetFile string, allowBypass bool) error {
	ctx := &HookContext{
		Type:         PostCapture,
		Workspace:    h.manager.workspace,
		Content:      content,
		TemplateName: templateName,
		DestPath:     targetFile,
		Timeout:      h.manager.timeout,
		AllowBypass:  allowBypass,
		ExtraEnv: map[string]string{
			"JOT_CAPTURED_FILE": targetFile,
			"JOT_CONTENT_SIZE":  string(rune(len(content))),
		},
	}

	result, err := h.manager.Execute(ctx)
	if err != nil {
		return err
	}

	// Post hooks can't abort the operation since it's already completed
	// But we can log errors or warnings
	if result.Aborted && result.Error != nil {
		// Log the error but don't fail the capture
		// TODO: Add logging framework
		return nil
	}

	return nil
}

// ExecuteWithTimeout runs the post-capture hook with a custom timeout
func (h *PostCaptureHook) ExecuteWithTimeout(content, templateName, targetFile string, timeout time.Duration, allowBypass bool) error {
	ctx := &HookContext{
		Type:         PostCapture,
		Workspace:    h.manager.workspace,
		Content:      content,
		TemplateName: templateName,
		DestPath:     targetFile,
		Timeout:      timeout,
		AllowBypass:  allowBypass,
		ExtraEnv: map[string]string{
			"JOT_CAPTURED_FILE": targetFile,
			"JOT_CONTENT_SIZE":  string(rune(len(content))),
		},
	}

	_, err := h.manager.Execute(ctx)
	if err != nil {
		return err
	}

	return nil
}

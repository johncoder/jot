package hooks

import (
	"time"

	"github.com/johncoder/jot/internal/workspace"
)

// WorkspaceChangeHook handles workspace-change hook execution
type WorkspaceChangeHook struct {
	manager *Manager
}

// NewWorkspaceChangeHook creates a new workspace-change hook handler
func NewWorkspaceChangeHook(ws *workspace.Workspace) *WorkspaceChangeHook {
	return &WorkspaceChangeHook{
		manager: NewManager(ws),
	}
}

// Execute runs the workspace-change hook with the given context
func (h *WorkspaceChangeHook) Execute(oldWorkspace, newWorkspace string, allowBypass bool) error {
	ctx := &HookContext{
		Type:        WorkspaceChange,
		Workspace:   h.manager.workspace,
		Timeout:     h.manager.timeout,
		AllowBypass: allowBypass,
		ExtraEnv: map[string]string{
			"JOT_OLD_WORKSPACE": oldWorkspace,
			"JOT_NEW_WORKSPACE": newWorkspace,
		},
	}

	_, err := h.manager.Execute(ctx)
	if err != nil {
		return err
	}

	return nil
}

// ExecuteWithTimeout runs the workspace-change hook with a custom timeout
func (h *WorkspaceChangeHook) ExecuteWithTimeout(oldWorkspace, newWorkspace string, timeout time.Duration, allowBypass bool) error {
	ctx := &HookContext{
		Type:        WorkspaceChange,
		Workspace:   h.manager.workspace,
		Timeout:     timeout,
		AllowBypass: allowBypass,
		ExtraEnv: map[string]string{
			"JOT_OLD_WORKSPACE": oldWorkspace,
			"JOT_NEW_WORKSPACE": newWorkspace,
		},
	}

	_, err := h.manager.Execute(ctx)
	if err != nil {
		return err
	}

	return nil
}

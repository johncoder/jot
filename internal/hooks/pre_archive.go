package hooks

import (
	"time"

	"github.com/johncoder/jot/internal/workspace"
)

// PreArchiveHook handles pre-archive hook execution
type PreArchiveHook struct {
	manager *Manager
}

// NewPreArchiveHook creates a new pre-archive hook handler
func NewPreArchiveHook(ws *workspace.Workspace) *PreArchiveHook {
	return &PreArchiveHook{
		manager: NewManager(ws),
	}
}

// Execute runs the pre-archive hook with the given context
func (h *PreArchiveHook) Execute(content, sourceFile, archiveLocation string, allowBypass bool) error {
	ctx := &HookContext{
		Type:        PreArchive,
		Workspace:   h.manager.workspace,
		Content:     content,
		SourceFile:  sourceFile,
		DestPath:    archiveLocation,
		Timeout:     h.manager.timeout,
		AllowBypass: allowBypass,
		ExtraEnv: map[string]string{
			"JOT_ARCHIVE_SOURCE":   sourceFile,
			"JOT_ARCHIVE_LOCATION": archiveLocation,
		},
	}

	result, err := h.manager.Execute(ctx)
	if err != nil {
		return err
	}

	if result.Aborted {
		return result.Error
	}

	return nil
}

// ExecuteWithTimeout runs the pre-archive hook with a custom timeout
func (h *PreArchiveHook) ExecuteWithTimeout(content, sourceFile, archiveLocation string, timeout time.Duration, allowBypass bool) error {
	ctx := &HookContext{
		Type:        PreArchive,
		Workspace:   h.manager.workspace,
		Content:     content,
		SourceFile:  sourceFile,
		DestPath:    archiveLocation,
		Timeout:     timeout,
		AllowBypass: allowBypass,
		ExtraEnv: map[string]string{
			"JOT_ARCHIVE_SOURCE":   sourceFile,
			"JOT_ARCHIVE_LOCATION": archiveLocation,
		},
	}

	result, err := h.manager.Execute(ctx)
	if err != nil {
		return err
	}

	if result.Aborted {
		return result.Error
	}

	return nil
}

package hooks

import (
	"time"

	"github.com/johncoder/jot/internal/workspace"
)

// PostArchiveHook handles post-archive hook execution
type PostArchiveHook struct {
	manager *Manager
}

// NewPostArchiveHook creates a new post-archive hook handler
func NewPostArchiveHook(ws *workspace.Workspace) *PostArchiveHook {
	return &PostArchiveHook{
		manager: NewManager(ws),
	}
}

// Execute runs the post-archive hook with the given context
func (h *PostArchiveHook) Execute(content, sourceFile, archiveLocation string, allowBypass bool) error {
	ctx := &HookContext{
		Type:        PostArchive,
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

	_, err := h.manager.Execute(ctx)
	if err != nil {
		return err
	}

	return nil
}

// ExecuteWithTimeout runs the post-archive hook with a custom timeout
func (h *PostArchiveHook) ExecuteWithTimeout(content, sourceFile, archiveLocation string, timeout time.Duration, allowBypass bool) error {
	ctx := &HookContext{
		Type:        PostArchive,
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

	_, err := h.manager.Execute(ctx)
	if err != nil {
		return err
	}

	return nil
}

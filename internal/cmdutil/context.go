package cmdutil

import (
	"time"

	"github.com/spf13/cobra"
)

// CommandContext holds common command execution data.
type CommandContext struct {
	Cmd       *cobra.Command
	StartTime time.Time
	Response  *ResponseManager
}

// StartCommand initializes command execution with timing and response management.
func StartCommand(cmd *cobra.Command) *CommandContext {
	startTime := time.Now()
	return &CommandContext{
		Cmd:       cmd,
		StartTime: startTime,
		Response:  NewResponseManager(cmd, startTime),
	}
}

// HandleError provides convenient error handling using the command context.
func (ctx *CommandContext) HandleError(err error) error {
	return HandleError(ctx.Cmd, err, ctx.StartTime)
}

// HandleErrorf provides convenient formatted error handling using the command context.
func (ctx *CommandContext) HandleErrorf(format string, args ...interface{}) error {
	return HandleErrorf(ctx.Cmd, ctx.StartTime, format, args...)
}

// HandleOperationError provides convenient operation error handling using the command context.
func (ctx *CommandContext) HandleOperationError(operation string, err error) error {
	return HandleOperationError(ctx.Cmd, ctx.StartTime, operation, err)
}

// IsJSONOutput checks if JSON output is requested.
func (ctx *CommandContext) IsJSONOutput() bool {
	return IsJSONOutput(ctx.Cmd)
}

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

// Modern error handling context methods

// WrapFileError wraps file operation errors with context
func (ctx *CommandContext) WrapFileError(op, path string, err error) error {
	fileErr := NewFileError(op, path, err)
	return ctx.HandleError(fileErr)
}

// WrapValidationError wraps validation errors with context
func (ctx *CommandContext) WrapValidationError(field, value string, err error) error {
	validationErr := NewValidationError(field, value, err)
	return ctx.HandleError(validationErr)
}

// WrapExternalError wraps external command errors with context
func (ctx *CommandContext) WrapExternalError(command string, args []string, err error) error {
	extErr := NewExternalError(command, args, err)
	return ctx.HandleError(extErr)
}

// HandleFileOperation provides specialized file operation error handling
func (ctx *CommandContext) HandleFileOperation(op, path string, err error) error {
	if err == nil {
		return nil
	}
	return ctx.WrapFileError(op, path, err)
}

// HandleValidation provides specialized validation error handling
func (ctx *CommandContext) HandleValidation(field, value string, err error) error {
	if err == nil {
		return nil
	}
	return ctx.WrapValidationError(field, value, err)
}

// HandleExternalCommand provides specialized external command error handling
func (ctx *CommandContext) HandleExternalCommand(command string, args []string, err error) error {
	if err == nil {
		return nil
	}
	return ctx.WrapExternalError(command, args, err)
}

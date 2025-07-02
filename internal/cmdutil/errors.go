package cmdutil

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

// HandleError provides unified error handling for commands.
// It automatically detects JSON output mode and formats errors appropriately.
func HandleError(cmd *cobra.Command, err error, startTime time.Time) error {
	if err == nil {
		return nil
	}
	
	if IsJSONOutput(cmd) {
		return OutputJSONError(cmd, err, startTime)
	}
	return err
}

// HandleErrorf provides unified error handling with formatted error messages.
func HandleErrorf(cmd *cobra.Command, startTime time.Time, format string, args ...interface{}) error {
	err := fmt.Errorf(format, args...)
	return HandleError(cmd, err, startTime)
}

// WrapError wraps an error with additional context using fmt.Errorf.
func WrapError(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf(format+": %w", append(args, err)...)
}

// OperationError represents a failed operation with context.
type OperationError struct {
	Operation string
	Err       error
}

func (e *OperationError) Error() string {
	return fmt.Sprintf("failed to %s: %v", e.Operation, e.Err)
}

func (e *OperationError) Unwrap() error {
	return e.Err
}

// NewOperationError creates a new operation error.
func NewOperationError(operation string, err error) *OperationError {
	return &OperationError{
		Operation: operation,
		Err:       err,
	}
}

// HandleOperationError handles operation-specific errors with consistent formatting.
func HandleOperationError(cmd *cobra.Command, startTime time.Time, operation string, err error) error {
	if err == nil {
		return nil
	}
	
	opErr := NewOperationError(operation, err)
	return HandleError(cmd, opErr, startTime)
}

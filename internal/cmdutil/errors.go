package cmdutil

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

// Standard error variables for error inspection
var (
	ErrFileNotFound    = errors.New("file not found")
	ErrInvalidInput    = errors.New("invalid input")
	ErrOperationFailed = errors.New("operation failed")
	ErrExternalCommand = errors.New("external command failed")
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

// FileError represents file operation errors with rich context
type FileError struct {
	Op   string // "read", "write", "create", "delete"
	Path string
	Err  error
}

func (e *FileError) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("failed to %s file %s: %v", e.Op, e.Path, e.Err)
	}
	return fmt.Sprintf("failed to %s file: %v", e.Op, e.Err)
}

func (e *FileError) Unwrap() error { return e.Err }

// Is implements error matching for FileError
func (e *FileError) Is(target error) bool {
	if target == ErrFileNotFound && errors.Is(e.Err, os.ErrNotExist) {
		return true
	}
	return false
}

// ValidationError represents input validation errors
type ValidationError struct {
	Field string
	Value string
	Err   error
}

func (e *ValidationError) Error() string {
	if e.Value != "" {
		return fmt.Sprintf("invalid %s '%s': %v", e.Field, e.Value, e.Err)
	}
	return fmt.Sprintf("invalid %s: %v", e.Field, e.Err)
}

func (e *ValidationError) Unwrap() error { return e.Err }

// Is implements error matching for ValidationError
func (e *ValidationError) Is(target error) bool {
	return target == ErrInvalidInput
}

// ExternalError represents external command failures
type ExternalError struct {
	Command string
	Args    []string
	Err     error
}

func (e *ExternalError) Error() string {
	if len(e.Args) > 0 {
		return fmt.Sprintf("%s %v failed: %v", e.Command, e.Args, e.Err)
	}
	return fmt.Sprintf("%s failed: %v", e.Command, e.Err)
}

func (e *ExternalError) Unwrap() error { return e.Err }

// Is implements error matching for ExternalError
func (e *ExternalError) Is(target error) bool {
	return target == ErrExternalCommand
}

// Constructor functions for structured errors

// NewFileError creates a file operation error
func NewFileError(op, path string, err error) *FileError {
	return &FileError{Op: op, Path: path, Err: err}
}

// NewValidationError creates a validation error
func NewValidationError(field, value string, err error) *ValidationError {
	return &ValidationError{Field: field, Value: value, Err: err}
}

// NewExternalError creates an external command error
func NewExternalError(command string, args []string, err error) *ExternalError {
	return &ExternalError{Command: command, Args: args, Err: err}
}

// Error inspection utilities

// IsFileNotFound checks if error indicates a missing file
func IsFileNotFound(err error) bool {
	return errors.Is(err, ErrFileNotFound) || errors.Is(err, os.ErrNotExist)
}

// IsValidationError checks if error is a validation failure
func IsValidationError(err error) bool {
	return errors.Is(err, ErrInvalidInput)
}

// IsExternalCommandError checks if error is from external command failure
func IsExternalCommandError(err error) bool {
	return errors.Is(err, ErrExternalCommand)
}

// GetFileError extracts FileError details if present
func GetFileError(err error) (*FileError, bool) {
	var fileErr *FileError
	ok := errors.As(err, &fileErr)
	return fileErr, ok
}

// GetValidationError extracts ValidationError details if present
func GetValidationError(err error) (*ValidationError, bool) {
	var validationErr *ValidationError
	ok := errors.As(err, &validationErr)
	return validationErr, ok
}

// GetExternalError extracts ExternalError details if present
func GetExternalError(err error) (*ExternalError, bool) {
	var extErr *ExternalError
	ok := errors.As(err, &extErr)
	return extErr, ok
}

// Enhanced OperationError with target support
func (e *OperationError) WithTarget(target string) *OperationError {
	if target != "" {
		e.Operation = e.Operation + " " + target
	}
	return e
}

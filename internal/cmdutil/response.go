package cmdutil

import (
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// JSONMetadata represents common metadata included in JSON responses.
// Compatible with existing cmd/json.go format.
type JSONMetadata struct {
	Success       bool      `json:"success"`
	Command       string    `json:"command"`
	ExecutionTime int64     `json:"execution_time_ms"`
	Timestamp     time.Time `json:"timestamp"`
}

// JSONError represents an error in JSON format.
// Compatible with existing cmd/json.go format.
type JSONError struct {
	Message string                 `json:"message"`
	Code    string                 `json:"code"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// ResponseManager handles command output in both JSON and text formats.
type ResponseManager struct {
	cmd       *cobra.Command
	startTime time.Time
}

// NewResponseManager creates a new response manager for a command.
func NewResponseManager(cmd *cobra.Command, startTime time.Time) *ResponseManager {
	return &ResponseManager{
		cmd:       cmd,
		startTime: startTime,
	}
}

// IsJSONOutput checks if the command should output JSON format.
func IsJSONOutput(cmd *cobra.Command) bool {
	if cmd == nil {
		return false
	}
	
	jsonFlag, err := cmd.Flags().GetBool("json")
	if err != nil {
		return false
	}
	return jsonFlag
}

// CreateJSONMetadata creates standard metadata for JSON responses.
// Compatible with existing cmd/json.go format.
func CreateJSONMetadata(cmd *cobra.Command, success bool, startTime time.Time) JSONMetadata {
	return JSONMetadata{
		Success:       success,
		Command:       cmd.CommandPath(),
		ExecutionTime: time.Since(startTime).Milliseconds(),
		Timestamp:     time.Now(),
	}
}

// OutputJSON outputs a JSON response to stdout.
func OutputJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// OutputJSONError outputs an error in JSON format.
// Compatible with existing cmd/json.go format.
func OutputJSONError(cmd *cobra.Command, err error, startTime time.Time) error {
	errorCode := "unknown_error"
	details := map[string]interface{}{}

	// Extract error details based on error type
	if strings.Contains(err.Error(), "not found") {
		errorCode = "not_found"
	} else if strings.Contains(err.Error(), "workspace") {
		errorCode = "workspace_error"
	}

	response := map[string]interface{}{
		"error": JSONError{
			Message: err.Error(),
			Code:    errorCode,
			Details: details,
		},
		"metadata": CreateJSONMetadata(cmd, false, startTime),
	}
	
	if jsonErr := OutputJSON(response); jsonErr != nil {
		// If JSON output fails, fall back to regular error
		return err
	}
	
	// Return the original error so the command exits with proper code
	return err
}

// RespondWithSuccess outputs a successful response in the appropriate format.
func (rm *ResponseManager) RespondWithSuccess(data interface{}) error {
	if IsJSONOutput(rm.cmd) {
		// If data doesn't already have metadata, add it
		if responseMap, ok := data.(map[string]interface{}); ok {
			if _, hasMetadata := responseMap["metadata"]; !hasMetadata {
				responseMap["metadata"] = CreateJSONMetadata(rm.cmd, true, rm.startTime)
			}
		}
		return OutputJSON(data)
	}
	
	// For non-JSON output, data should be handled by the caller
	// since text output is command-specific
	return nil
}

// RespondWithError outputs an error response in the appropriate format.
func (rm *ResponseManager) RespondWithError(err error) error {
	if IsJSONOutput(rm.cmd) {
		return OutputJSONError(rm.cmd, err, rm.startTime)
	}
	return err
}

// RespondWithOperation creates a standardized operation response.
func (rm *ResponseManager) RespondWithOperation(operation string, result string, details interface{}) error {
	if IsJSONOutput(rm.cmd) {
		response := map[string]interface{}{
			"operations": []map[string]interface{}{
				{
					"operation": operation,
					"result":    result,
					"details":   details,
				},
			},
			"metadata": CreateJSONMetadata(rm.cmd, true, rm.startTime),
		}
		return OutputJSON(response)
	}
	
	// For text output, the caller should handle the specific formatting
	return nil
}

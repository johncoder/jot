package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// JSONMetadata contains standard metadata for all JSON responses
type JSONMetadata struct {
	Success       bool      `json:"success"`
	Command       string    `json:"command"`
	ExecutionTime int64     `json:"execution_time_ms"`
	Timestamp     time.Time `json:"timestamp"`
}

// JSONError represents an error in JSON format
type JSONError struct {
	Message string                 `json:"message"`
	Code    string                 `json:"code"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// JSONResponse is the base structure for all JSON responses
type JSONResponse struct {
	Data     interface{}  `json:",inline"`
	Error    *JSONError   `json:"error,omitempty"`
	Metadata JSONMetadata `json:"metadata"`
}

// JSONOperation represents a single operation in action commands
type JSONOperation struct {
	Operation string                 `json:"operation"`
	Result    string                 `json:"result"`
	Details   map[string]interface{} `json:"details"`
}

// JSONOperationSummary provides summary statistics for operations
type JSONOperationSummary struct {
	TotalOperations int `json:"total_operations"`
	Successful      int `json:"successful"`
	Failed          int `json:"failed"`
}

// JSONOperationsResponse is the standard format for action commands
type JSONOperationsResponse struct {
	Operations []JSONOperation      `json:"operations"`
	Summary    JSONOperationSummary `json:"summary"`
}

// Global variable to track if JSON output is enabled
var jsonOutput bool

// Global variable to track command execution start time
var commandStartTime time.Time

// JSONOutputEnabled returns true if --json flag is set
func JSONOutputEnabled() bool {
	return jsonOutput
}

// StartTiming initializes command execution timing
func StartTiming() {
	commandStartTime = time.Now()
}

// BuildJSONResponse creates a JSON response with metadata
func BuildJSONResponse(data interface{}, commandName string) JSONResponse {
	executionTime := time.Since(commandStartTime).Milliseconds()

	return JSONResponse{
		Data: data,
		Metadata: JSONMetadata{
			Success:       true,
			Command:       commandName,
			ExecutionTime: executionTime,
			Timestamp:     time.Now(),
		},
	}
}

// BuildJSONError creates a JSON error response
func BuildJSONError(message, code string, details map[string]interface{}, commandName string) JSONResponse {
	executionTime := time.Since(commandStartTime).Milliseconds()

	return JSONResponse{
		Error: &JSONError{
			Message: message,
			Code:    code,
			Details: details,
		},
		Metadata: JSONMetadata{
			Success:       false,
			Command:       commandName,
			ExecutionTime: executionTime,
			Timestamp:     time.Now(),
		},
	}
}

// OutputJSON outputs a JSON response and exits
func OutputJSON(response JSONResponse) {
	jsonBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(jsonBytes))

	// Exit with appropriate code
	if response.Error != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

// BuildOperationSuccess creates a successful operation
func BuildOperationSuccess(operation string, details map[string]interface{}) JSONOperation {
	return JSONOperation{
		Operation: operation,
		Result:    "success",
		Details:   details,
	}
}

// BuildOperationFailure creates a failed operation
func BuildOperationFailure(operation string, errorMsg string, details map[string]interface{}) JSONOperation {
	if details == nil {
		details = make(map[string]interface{})
	}
	details["error"] = errorMsg

	return JSONOperation{
		Operation: operation,
		Result:    "failure",
		Details:   details,
	}
}

// BuildOperationsSummary creates a summary from a list of operations
func BuildOperationsSummary(operations []JSONOperation) JSONOperationSummary {
	summary := JSONOperationSummary{
		TotalOperations: len(operations),
	}

	for _, op := range operations {
		if op.Result == "success" {
			summary.Successful++
		} else {
			summary.Failed++
		}
	}

	return summary
}

// outputJSON outputs a JSON response
func outputJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

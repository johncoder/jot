package cmdutil

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ConfirmOperation prompts the user with a yes/no question and returns their response.
// The prompt should be a complete question without the [y/N] suffix, which is added automatically.
func ConfirmOperation(prompt string) (bool, error) {
	fmt.Printf("%s [y/N]: ", prompt)

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read user input: %w", err)
	}

	return IsConfirmationYes(response), nil
}

// ShowSuccess displays a success message with consistent formatting.
// Message can include icons and will be printed as-is with formatting.
func ShowSuccess(message string, args ...interface{}) {
	fmt.Printf(message+"\n", args...)
}

// ShowError displays an error message with consistent formatting.
// Message can include icons and will be printed as-is with formatting.
func ShowError(message string, args ...interface{}) {
	fmt.Printf(message+"\n", args...)
}

// ShowWarning displays a warning message with consistent formatting.
// Message can include icons and will be printed as-is with formatting.
func ShowWarning(message string, args ...interface{}) {
	fmt.Printf(message+"\n", args...)
}

// ShowInfo displays an informational message with consistent formatting.
// Message can include icons and will be printed as-is with formatting.
func ShowInfo(message string, args ...interface{}) {
	fmt.Printf(message+"\n", args...)
}

// ShowProgress displays a progress message with consistent formatting.
// Message can include icons and will be printed as-is with formatting.
func ShowProgress(message string, args ...interface{}) {
	fmt.Printf(message+"\n", args...)
}

// NormalizeUserInput trims whitespace and converts input to lowercase for consistent processing.
func NormalizeUserInput(input string) string {
	return strings.ToLower(strings.TrimSpace(input))
}

// IsConfirmationYes checks if the given input represents a positive confirmation.
// Accepts "y", "yes" (case-insensitive) as positive responses.
func IsConfirmationYes(input string) bool {
	normalized := NormalizeUserInput(input)
	return normalized == "y" || normalized == "yes"
}

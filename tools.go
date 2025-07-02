//go:build tools
// +build tools

// Package tools is used to pin development tool versions
// This ensures all developers use the same versions of tools
// Run: go mod tidy to download these tools
package tools

import (
	// Static analysis
	_ "honnef.co/go/tools/cmd/staticcheck"
	// Import formatting (optional)
	// _ "golang.org/x/tools/cmd/goimports"
	// Vulnerability checking (optional)
	// _ "golang.org/x/vuln/cmd/govulncheck"
	// Linting (optional)
	// _ "golang.org/x/lint/golint"
)

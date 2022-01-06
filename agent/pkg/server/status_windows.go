//go:build windows

package server

import (
	"path/filepath"
)

var (
	// PathStatus is filepath for status
	PathStatus = filepath.Join("C", "ProgramData", "shoes-agent")
)

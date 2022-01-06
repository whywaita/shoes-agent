//go:build windows

package server

import (
	"path/filepath"
)

var (
	PathStatus = filepath.Join("C", "ProgramData", "shoes-agent")
)

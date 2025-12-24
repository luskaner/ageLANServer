//go:build !windows && !darwin && !linux

// This file provides fallback implementations for unsupported Unix systems.
// It implements GetProcessStartTime, WaitForProcess, and ProcessesPID with
// minimal functionality to allow the package to compile and run on any Unix system.

package process

import (
	"os"
	"time"
)

func GetProcessStartTime(_ int) (int64, error) {
	// Fallback for unsupported Unix systems - always return 0
	// This disables startTime validation but maintains basic functionality
	return 0, nil
}

func WaitForProcess(_ *os.Process, _ *time.Duration) bool {
	return true
}

// ProcessesPID returns a map of process names to their PIDs.
// Note: If multiple processes share the same name, only one PID is stored per name.
func ProcessesPID(_ []string) map[string]uint32 {
	return make(map[string]uint32)
}

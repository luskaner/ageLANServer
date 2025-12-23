//go:build !windows && !darwin && !linux

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

func ProcessesPID(_ []string) map[string]uint32 {
	return make(map[string]uint32)
}

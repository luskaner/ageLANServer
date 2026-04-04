//go:build linux || darwin

// This file provides shared Unix implementations for Linux and Darwin.
// It works in conjunction with:
//   - process_linux.go: provides GetProcessStartTime and ProcessesPID for Linux
//   - process_darwin.go: provides GetProcessStartTime and ProcessesPID for Darwin
//   - process_notwin.go: provides FindProcess and FindProcessWithStartTime for all non-Windows

package process

import (
	"errors"
	"os"
	"time"

	"golang.org/x/sys/unix"
)

func alive(proc *os.Process) bool {
	err := proc.Signal(unix.Signal(0))
	if err == nil {
		return true
	}
	if errors.Is(err, unix.EPERM) {
		return true
	}
	return false
}

func WaitForProcess(proc *os.Process, duration *time.Duration) bool {
	pollInterval := 100 * time.Millisecond
	if duration == nil {
		pollInterval *= 10
	}
	waitByPolling := func(timeoutChan <-chan time.Time) bool {
		ticker := time.NewTicker(pollInterval)
		defer ticker.Stop()
		for {
			if !alive(proc) {
				return true
			}
			select {
			case <-timeoutChan:
				return false
			case <-ticker.C:
				if !alive(proc) {
					return true
				}
			}
		}
	}
	done := make(chan error, 1)
	go func() {
		_, err := proc.Wait()
		done <- err
	}()
	var timeoutChan <-chan time.Time
	if duration != nil {
		timer := time.NewTimer(*duration)
		defer timer.Stop()
		timeoutChan = timer.C
	}
	select {
	case err := <-done:
		if err == nil {
			return true
		}
		return waitByPolling(timeoutChan)
	case <-timeoutChan:
		return false
	}
}

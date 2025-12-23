//go:build linux || darwin

package process

import (
	"os"
	"time"

	"golang.org/x/sys/unix"
)

func WaitForProcess(proc *os.Process, duration *time.Duration) bool {
	t := 100 * time.Millisecond
	if duration == nil {
		t *= 10
	}
	processExists := func() bool {
		return proc.Signal(unix.Signal(0)) != nil
	}
	if duration == nil {
		for {
			if processExists() {
				return true
			}
			time.Sleep(t)
		}
	} else {
		timeout := time.After(*duration)
		for {
			select {
			case <-timeout:
				return false
			default:
				if processExists() {
					return true
				}
				time.Sleep(t)
			}
		}
	}
}

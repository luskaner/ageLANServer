//go:build !windows

package process

import (
	"errors"
	"os"

	"golang.org/x/sys/unix"
)

func FindProcess(pid int) (proc *os.Process, err error) {
	return FindProcessWithStartTime(pid, 0)
}

func FindProcessWithStartTime(pid int, expectedStartTime int64) (proc *os.Process, err error) {
	proc, err = os.FindProcess(pid)
	if err != nil {
		return
	}
	if err = proc.Signal(unix.Signal(0)); err != nil {
		if errors.Is(err, unix.EPERM) {
			err = nil
		} else {
			proc = nil
			return
		}
	}
	if expectedStartTime != 0 {
		actualStartTime, startErr := GetProcessStartTime(pid)
		if startErr != nil {
			proc = nil
			err = startErr
			return
		}
		if actualStartTime != expectedStartTime {
			proc = nil
			err = errors.New("process start time mismatch")
		}
	}
	return
}

//go:build !windows

package process

import (
	"errors"
	"golang.org/x/sys/unix"
	"os"
)

func FindProcess(pid int) (proc *os.Process, err error) {
	proc, err = os.FindProcess(pid)
	if err != nil {
		return
	}
	if err = proc.Signal(unix.Signal(0)); err != nil {
		if errors.Is(err, unix.EPERM) {
			err = nil
		} else {
			proc = nil
		}
	}
	return
}

//go:build !windows

package process

import (
	"os"
	"syscall"
)

func FindProcess(pid int) (proc *os.Process, err error) {
	proc, err = os.FindProcess(pid)
	if err != nil {
		return
	}
	if err = proc.Signal(syscall.Signal(0)); err == nil {
		proc = nil
	}
	return
}

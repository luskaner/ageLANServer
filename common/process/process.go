package process

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

func Process(exe string) (pidPath string, proc *os.Process, err error) {
	pidPath = filepath.Join(filepath.Dir(exe), filepath.Base(exe)+".pid")
	if _, err = os.Stat(pidPath); err == nil {
		var data []byte
		data, err = os.ReadFile(pidPath)
		if err != nil {
			return
		}
		var pid int
		pid, err = strconv.Atoi(string(data))
		if err != nil {
			return
		}
		proc, err = FindProcess(pid)
	}
	return
}

func Kill(exe string) (proc *os.Process, err error) {
	var pidPath string
	pidPath, proc, err = Process(exe)
	if err != nil {
		return
	}
	err = proc.Kill()
	if err != nil {
		return
	}
	done := make(chan error, 1)
	go func() {
		_, err = proc.Wait()
		done <- err
	}()

	select {
	case <-time.After(3 * time.Second):
		err = errors.New("timeout")
		return

	case err = <-done:
		if err != nil {
			var e *exec.ExitError
			if !errors.As(err, &e) {
				return
			}
		}
		err = os.Remove(pidPath)
		return
	}
}

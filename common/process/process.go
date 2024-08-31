package process

import (
	"errors"
	"github.com/luskaner/aoe2DELanServer/common"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

func getPidPaths(exePath string) (paths []string) {
	name := common.Name + "-" + filepath.Base(exePath) + ".pid"
	tmp := os.TempDir()
	if tmp != "" {
		if d, e := os.Stat(tmp); e == nil && d.IsDir() {
			paths = append(paths, filepath.Join(tmp, name))
		}
	}
	paths = append(paths, filepath.Join(filepath.Dir(exePath), name))
	return
}

func Process(exe string) (pidPath string, proc *os.Process, err error) {
	pidPaths := getPidPaths(exe)
	var pid int
	for _, pidPath = range pidPaths {
		var data []byte
		data, err = os.ReadFile(pidPath)
		if err != nil {
			continue
		}
		pid, err = strconv.Atoi(string(data))
		if err != nil {
			continue
		}
		proc, err = FindProcess(pid)
		return
	}
	pidPath = pidPaths[0]
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

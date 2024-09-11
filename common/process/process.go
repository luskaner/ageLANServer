package process

import (
	"errors"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/aoe2DELanServer/common"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

const steamProcess = "AoE2DE_s.exe"
const microsoftStoreProcess = "AoE2DE.exe"

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

func GameProcesses(steam bool, microsoftStore bool) []string {
	processes := mapset.NewSet[string]()
	if steam {
		processes.Add(steamProcess)
	}
	if microsoftStore {
		processes.Add(microsoftStoreProcess)
	}
	return processes.ToSlice()
}

func AnyProcessExists(names []string) bool {
	processes := ProcessesPID(names)
	return len(processes) > 0
}

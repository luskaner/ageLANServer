package process

import (
	"errors"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

func steamProcess(gameTitle common.GameTitle) string {
	switch gameTitle {
	case common.AoE1:
		return "AoEDE_s.exe"
	case common.AoE2:
		return "AoE2DE_s.exe"
	case common.AoE3:
		return "AoE3DE_s.exe"
	case common.AoE4:
		return "RelicCardinal.exe"
	case common.AoM:
		return "AoMRT_s.exe"
	default:
		return ""
	}
}

func xboxProcess(gameTitle common.GameTitle) string {
	switch gameTitle {
	case common.AoE1:
		return "AoEDE.exe"
	case common.AoE2:
		return "AoE2DE.exe"
	case common.AoE3:
		return "AoE3DE.exe"
	case common.AoE4:
		return "RelicCardinal_ws.exe"
	case common.AoM:
		return "AoMRT.exe"
	default:
		return ""
	}
}

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

func KillPid(pid int) (err error) {
	var proc *os.Process
	proc, err = FindProcess(pid)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = nil
		}
		return
	}
	return KillProc("", proc)
}

func KillProc(pidPath string, proc *os.Process) (err error) {
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
		if pidPath != "" {
			err = os.Remove(pidPath)
		}
		return
	}
}

func Kill(exe string) (proc *os.Process, err error) {
	var pidPath string
	pidPath, proc, err = Process(exe)
	if err != nil {
		return
	}
	err = KillProc(pidPath, proc)
	return
}

func GameProcesses(gameTitle common.GameTitle, steam bool, xbox bool) []string {
	processes := mapset.NewThreadUnsafeSet[string]()
	if steam {
		processes.Add(steamProcess(gameTitle))
	}
	if xbox {
		processes.Add(xboxProcess(gameTitle))
	}
	return processes.ToSlice()
}

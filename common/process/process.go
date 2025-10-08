package process

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common"
)

func steamProcess(gameId string) string {
	switch gameId {
	case common.GameAoE1:
		return "AoEDE_s.exe"
	case common.GameAoE2:
		return "AoE2DE_s.exe"
	case common.GameAoE3:
		return "AoE3DE_s.exe"
	case common.GameAoE4:
		return "RelicCardinal.exe"
	case common.GameAoM:
		return "AoMRT_s.exe"
	default:
		return ""
	}
}

func xboxProcess(gameId string) string {
	switch gameId {
	case common.GameAoE1:
		return "AoEDE.exe"
	case common.GameAoE2:
		return "AoE2DE.exe"
	case common.GameAoE3:
		return "AoE3DE.exe"
	case common.GameAoE4:
		return "RelicCardinal_ws.exe"
	case common.GameAoM:
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

func KillPidProc(pidPath string, proc *os.Process) (err error) {
	err = KillProc(proc)
	if err != nil {
		return
	}
	return os.Remove(pidPath)
}

func KillProc(proc *os.Process) (err error) {
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
		return
	}
}

func Kill(exe string) (proc *os.Process, err error) {
	var pidPath string
	pidPath, proc, err = Process(exe)
	if err != nil {
		return
	}
	err = KillPidProc(pidPath, proc)
	return
}

func GameProcesses(gameId string, steam bool, xbox bool) []string {
	processes := mapset.NewThreadUnsafeSet[string]()
	if steam {
		processes.Add(steamProcess(gameId))
	}
	if xbox {
		processes.Add(xboxProcess(gameId))
	}
	return processes.ToSlice()
}

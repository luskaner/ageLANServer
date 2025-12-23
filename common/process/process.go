package process

import (
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common"
)

const PidFileSize = 16 // int64 PID + int64 StartTime

var waitDuration = 3 * time.Second

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
	for _, pidPath = range pidPaths {
		var data []byte
		var localErr error
		data, localErr = os.ReadFile(pidPath)
		if localErr != nil {
			continue
		}
		if len(data) != PidFileSize {
			// Invalid format (old or corrupted), remove orphan file
			_ = os.Remove(pidPath)
			continue
		}
		pid := int(binary.LittleEndian.Uint64(data[0:8]))
		startTime := int64(binary.LittleEndian.Uint64(data[8:16]))
		proc, err = FindProcessWithStartTime(pid, startTime)
		if proc == nil && err == nil {
			// Process doesn't exist or startTime doesn't match, remove orphan file
			_ = os.Remove(pidPath)
			continue
		}
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
	if err = proc.Signal(os.Interrupt); err == nil && WaitForProcess(proc, &waitDuration) {
		return
	}
	err = proc.Kill()
	if err != nil {
		return
	}
	if !WaitForProcess(proc, &waitDuration) {
		err = errors.New("timeout")
	}
	return
}

func Kill(exe string) error {
	pidPath, proc, err := Process(exe)
	if err != nil {
		return err
	} else if proc != nil {
		return KillPidProc(pidPath, proc)
	}
	return nil
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

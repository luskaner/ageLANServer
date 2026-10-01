package process

import (
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/luskaner/ageLANServer/common"
)

const PidFileSize = 16 // uint64 PID + uint64 StartTime

var waitDuration = 3 * time.Second

var osTempDirFn = os.TempDir
var osStatFn = os.Stat
var osReadFileFn = os.ReadFile
var osRemoveFn = os.Remove
var waitForProcessFn = WaitForProcess
var killProcFn = KillProc
var processFn = Process
var findProcessWithStartTimeFn = FindProcessWithStartTime
var procSignalFn = func(p *os.Process, sig os.Signal) error { return p.Signal(sig) }
var procKillFn = func(p *os.Process) error { return p.Kill() }
var filepathAbsFn = filepath.Abs

// getPidPaths returns the candidate pid file locations for an executable.
//
// The temp candidate is keyed on the base name, so it does not care how the
// caller spells the path. The exe-dir candidate used to be filepath.Dir of
// whatever the caller happened to pass, which does care, and the callers
// disagree: every binary chdirs to its own directory on startup
// (common.ChdirToExe), so a process already sitting in bin/ passes the bare
// "config-admin-agent.exe" while the launcher, one level up, passes
// "bin\config-admin-agent.exe". Those resolved to bin\ and bin\bin\ for the same
// file, so whenever the temp location was unavailable two callers could look at
// two different pid files and each conclude the other process was not running.
//
// Resolving to an absolute path first anchors the candidate to the target's
// real directory whatever the current working directory is, which is what
// PidLock writes to.
func getPidPaths(exePath string) (paths []string) {
	name := common.Name + "-" + filepath.Base(exePath) + ".pid"
	tmp := osTempDirFn()
	if tmp != "" {
		if d, e := osStatFn(tmp); e == nil && d.IsDir() {
			paths = append(paths, filepath.Join(tmp, name))
		}
	}
	dir := filepath.Dir(exePath)
	if abs, e := filepathAbsFn(exePath); e == nil {
		dir = filepath.Dir(abs)
	}
	paths = append(paths, filepath.Join(dir, name))
	return
}

func Process(exe string) (pidPath string, proc *os.Process, err error) {
	pidPaths := getPidPaths(exe)
	for _, pidPath = range pidPaths {
		var data []byte
		var localErr error
		data, localErr = osReadFileFn(pidPath)
		if localErr != nil {
			continue
		}
		if len(data) != PidFileSize {
			// Invalid format (old or corrupted), remove orphan file
			// Error ignored: file may have been removed by concurrent process (race condition)
			_ = osRemoveFn(pidPath)
			continue
		}
		pid := int(binary.LittleEndian.Uint64(data[0:8]))
		startTime := int64(binary.LittleEndian.Uint64(data[8:16]))
		proc, err = findProcessWithStartTimeFn(pid, startTime)
		if proc == nil {
			err = nil
			// Process doesn't exist or startTime doesn't match, remove orphan file
			// Error ignored: file may have been removed by concurrent process (race condition)
			_ = osRemoveFn(pidPath)
			continue
		}
		return
	}
	pidPath = pidPaths[0]
	return
}

func KillPidProc(pidPath string, proc *os.Process) (err error) {
	err = killProcFn(proc)
	if err != nil {
		return
	}
	if _, err = osStatFn(pidPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = nil
		}
		return err
	}
	return osRemoveFn(pidPath)
}

func KillProc(proc *os.Process) (err error) {
	if err = procSignalFn(proc, os.Interrupt); err == nil && waitForProcessFn(proc, &waitDuration) {
		return
	}
	err = procKillFn(proc)
	if err != nil {
		return
	}
	if !waitForProcessFn(proc, &waitDuration) {
		err = errors.New("timeout")
	}
	return
}

func Kill(exe string) error {
	pidPath, proc, err := processFn(exe)
	if err != nil {
		return err
	} else if proc != nil {
		return KillPidProc(pidPath, proc)
	}
	return nil
}

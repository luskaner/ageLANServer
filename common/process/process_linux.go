package process

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
)

func parseCmdline(cmdline []byte) []string {
	cmdline = bytes.TrimSuffix(cmdline, []byte("\x00"))
	if len(cmdline) == 0 {
		return nil
	}
	cmdline = bytes.ReplaceAll(cmdline, []byte(`\`), []byte(`/`))
	parts := bytes.Split(cmdline, []byte("\x00"))
	args := make([]string, len(parts))
	for i, p := range parts {
		args[i] = string(p)
	}
	return args
}

// GetProcessStartTime returns the process start time as clock ticks since system boot.
//
// IMPORTANT: The returned value is NOT comparable across platforms:
//   - Linux: clock ticks since boot (this implementation)
//   - Windows: nanoseconds since epoch (absolute time)
//   - Darwin: nanoseconds since epoch (absolute time)
//
// Values are only meaningful when compared within the same platform and boot session.
// This is sufficient for detecting PID reuse, as the primary use case is comparing
// a stored start time against the current start time of a process with the same PID.
func GetProcessStartTime(pid int) (int64, error) {
	statPath := fmt.Sprintf("/proc/%d/stat", pid)
	data, err := os.ReadFile(statPath)
	if err != nil {
		return 0, err
	}
	// Format: pid (comm) state ppid ... field 22 is starttime
	// Find the last ')' to skip the comm field which may contain spaces/parentheses
	statStr := string(data)
	lastParen := strings.LastIndex(statStr, ")")
	if lastParen == -1 {
		return 0, errors.New("invalid stat format")
	}
	fields := strings.Fields(statStr[lastParen+1:])
	// After (comm), fields are: state(0), ppid(1), pgrp(2), session(3), tty_nr(4), tpgid(5),
	// flags(6), minflt(7), cminflt(8), majflt(9), cmajflt(10), utime(11), stime(12),
	// cutime(13), cstime(14), priority(15), nice(16), num_threads(17), itrealvalue(18),
	// starttime(19) - in clock ticks since boot
	if len(fields) < 20 {
		return 0, errors.New("insufficient stat fields")
	}
	startTime, err := strconv.ParseInt(fields[19], 10, 64)
	if err != nil {
		return 0, err
	}
	return startTime, nil
}

// ProcessesByNames returns a map of process names to their procs.
// Note: If multiple processes share the same name, only one PID is stored per name.
func ProcessesByNames(names []string) map[string]*os.Process {
	processesPid := make(map[string]*os.Process)
	procs, err := os.ReadDir("/proc")
	if err != nil {
		return processesPid
	}
	namesLeft := mapset.NewThreadUnsafeSet[string](names...)
	for _, proc := range procs {
		if namesLeft.IsEmpty() {
			break
		}
		var pid uint64
		if pid, err = strconv.ParseUint(proc.Name(), 10, 32); err == nil {
			var cmdline []byte
			cmdline, err = os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
			if err != nil {
				continue
			}
			args := parseCmdline(cmdline)
			if len(args) == 0 {
				continue
			}
			name := filepath.Base(args[0])
			if namesLeft.Contains(name) {
				if localProc, err := FindProcess(int(pid)); err == nil {
					processesPid[name] = localProc
					namesLeft.Remove(name)
				}
			}
		}
	}
	return processesPid
}

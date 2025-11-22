//go:build !windows && !darwin

package process

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"mvdan.cc/sh/v3/shell"
)

func WaitForProcess(proc *os.Process, duration *time.Duration) bool {
	t := 100 * time.Millisecond
	if duration == nil {
		t *= 10
	}
	procPath := fmt.Sprintf("/proc/%d", proc.Pid)
	processExists := func(path string) bool {
		if _, err := os.Stat(procPath); os.IsNotExist(err) {
			return true
		}
		return false
	}
	if duration == nil {
		for {
			if processExists(procPath) {
				return true
			}
			time.Sleep(t)
		}
	} else {
		timeout := time.After(*duration)
		for {
			select {
			case <-timeout:
				return false
			default:
				if processExists(procPath) {
					return true
				}
				time.Sleep(t)
			}
		}
	}
}

func ProcessesPID(names []string) map[string]uint32 {
	processesPid := make(map[string]uint32)
	procs, err := os.ReadDir("/proc")
	if err != nil {
		return processesPid
	}

	for _, proc := range procs {
		var pid uint64
		if pid, err = strconv.ParseUint(proc.Name(), 10, 32); err == nil {
			var cmdline []byte
			cmdline, err = os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
			if err != nil {
				continue
			}
			cmdlineStr := strings.TrimSpace(
				strings.ReplaceAll(strings.ReplaceAll(string(cmdline), "\x00", " "), "\\", "/"),
			)
			var args []string
			args, err = shell.Fields(cmdlineStr, nil)
			if err != nil || len(args) == 0 {
				continue
			}
			cmdlineName := filepath.Base(args[0])
			if slices.Contains(names, cmdlineName) {
				processesPid[cmdlineName] = uint32(pid)
			}
		}
	}
	return processesPid
}

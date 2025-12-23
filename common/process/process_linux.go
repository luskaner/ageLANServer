package process

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"mvdan.cc/sh/v3/shell"
)

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
	// starttime(19)
	if len(fields) < 20 {
		return 0, errors.New("insufficient stat fields")
	}
	startTime, err := strconv.ParseInt(fields[19], 10, 64)
	if err != nil {
		return 0, err
	}
	return startTime, nil
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

//go:build !windows && !darwin

package process

import (
	"fmt"
	"mvdan.cc/sh/v3/shell"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

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

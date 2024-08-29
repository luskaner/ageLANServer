//go:build !windows && !darwin

package process

import (
	"fmt"
	"mvdan.cc/sh/v3/shell"
	"os"
	"path/filepath"
	"strconv"
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
			if err == nil {
				cmdlineStr := string(cmdline)
				var args []string
				args, err = shell.Fields(cmdlineStr, nil)
				cmdlineName := filepath.Base(args[0])
				for _, name := range names {
					if cmdlineName == name {
						processesPid[name] = uint32(pid)
					}
				}
			}
		}
	}
	return processesPid
}

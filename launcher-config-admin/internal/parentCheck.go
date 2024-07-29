package internal

import (
	commonProcess "common/process"
	"golang.org/x/sys/windows"
	"os"
	"path/filepath"
)

func parentProcessID(pid int) (int, error) {
	processes := commonProcess.ProcessesEntry(func(entry *windows.ProcessEntry32) bool {
		return int(entry.ProcessID) == pid
	}, true)
	if len(processes) == 0 {
		return 0, nil
	}
	return int(processes[0].ParentProcessID), nil
}

func exePathFromPID(pid int) (string, error) {
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_INFORMATION|windows.PROCESS_VM_READ, false, uint32(pid))
	if err != nil {
		return "", err
	}
	defer func(handle windows.Handle) {
		_ = windows.CloseHandle(handle)
	}(h)

	var buf [windows.MAX_PATH]uint16
	size := uint32(len(buf))
	err = windows.QueryFullProcessImageName(h, 0, &buf[0], &size)
	if err != nil {
		return "", err
	}

	return windows.UTF16ToString(buf[:size]), nil
}

func ParentMatches(path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	pid := os.Getpid()
	ppid, err := parentProcessID(pid)
	if err != nil {
		return false
	}

	exePath, err := exePathFromPID(ppid)
	if err != nil {
		return false
	}

	return exePath == absPath
}

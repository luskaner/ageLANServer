package internal

import (
	"golang.org/x/sys/windows"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

func parentProcessID(pid int) (int, error) {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return 0, err
	}
	defer func(handle windows.Handle) {
		_ = windows.CloseHandle(handle)
	}(snapshot)

	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))

	if err = windows.Process32First(snapshot, &entry); err != nil {
		return 0, err
	}

	for {
		if entry.ProcessID == uint32(pid) {
			return int(entry.ParentProcessID), nil
		}
		err = windows.Process32Next(snapshot, &entry)
		if err != nil {
			return 0, err
		}
	}
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

	return syscall.UTF16ToString(buf[:size]), nil
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

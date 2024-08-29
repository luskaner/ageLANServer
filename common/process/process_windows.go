package process

import (
	"golang.org/x/sys/windows"
	"os"
	"unsafe"
)

func ProcessesPID(names []string) map[string]uint32 {
	name := func(entry *windows.ProcessEntry32) string {
		return windows.UTF16ToString(entry.ExeFile[:])
	}
	entries := processesEntry(func(entry *windows.ProcessEntry32) bool {
		for _, n := range names {
			if name(entry) == n {
				return true
			}
		}
		return false
	}, false)
	processes := make(map[string]uint32)
	for _, entry := range entries {
		processes[name(&entry)] = entry.ProcessID
	}
	return processes
}

func processesEntry(matches func(entry *windows.ProcessEntry32) bool, firstOnly bool) []windows.ProcessEntry32 {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil
	}
	defer func(handle windows.Handle) {
		_ = windows.CloseHandle(handle)
	}(snapshot)

	var procEntry windows.ProcessEntry32
	procEntry.Size = uint32(unsafe.Sizeof(procEntry))

	err = windows.Process32First(snapshot, &procEntry)
	if err != nil {
		return nil
	}

	var entries []windows.ProcessEntry32

	for {
		if matches(&procEntry) {
			entries = append(entries, procEntry)
			if firstOnly {
				break
			}
		}
		err = windows.Process32Next(snapshot, &procEntry)
		if err != nil {
			break
		}
	}

	return entries
}

func FindProcess(pid int) (proc *os.Process, err error) {
	proc, err = os.FindProcess(pid)
	if err != nil {
		return
	}
	entries := processesEntry(func(entry *windows.ProcessEntry32) bool {
		return int(entry.ProcessID) == pid
	}, true)
	if len(entries) == 0 {
		proc = nil
	}
	return
}

package process

import (
	mapset "github.com/deckarep/golang-set/v2"
	"golang.org/x/sys/windows"
	"os"
	"unsafe"
)

const steamProcess = "AoE2DE_s.exe"
const microsoftStoreProcess = "AoE2DE.exe"

func AnyProcessExists(names []string) bool {
	processes := ProcessesEntryNames(names)
	return len(processes) > 0
}

func GameProcesses(steam bool, microsoftStore bool) []string {
	processes := mapset.NewSet[string]()
	if steam {
		processes.Add(steamProcess)
	}
	if microsoftStore {
		processes.Add(microsoftStoreProcess)
	}
	return processes.ToSlice()
}

func ProcessesEntryNames(names []string) map[string]windows.ProcessEntry32 {
	name := func(entry *windows.ProcessEntry32) string {
		return windows.UTF16ToString(entry.ExeFile[:])
	}
	entries := ProcessesEntry(func(entry *windows.ProcessEntry32) bool {
		for _, n := range names {
			if name(entry) == n {
				return true
			}
		}
		return false
	}, false)
	processes := make(map[string]windows.ProcessEntry32)
	for _, entry := range entries {
		processes[name(&entry)] = entry
	}
	return processes
}

func ProcessesEntry(matches func(entry *windows.ProcessEntry32) bool, firstOnly bool) []windows.ProcessEntry32 {
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

	var processesEntry []windows.ProcessEntry32

	for {
		if matches(&procEntry) {
			processesEntry = append(processesEntry, procEntry)
			if firstOnly {
				break
			}
		}
		err = windows.Process32Next(snapshot, &procEntry)
		if err != nil {
			break
		}
	}

	return processesEntry
}

func FindProcess(pid int) (proc *os.Process, err error) {
	proc, err = os.FindProcess(pid)
	if err != nil {
		return
	}
	entries := ProcessesEntry(func(entry *windows.ProcessEntry32) bool {
		return int(entry.ProcessID) == pid
	}, true)
	if len(entries) == 0 {
		proc = nil
	}
	return
}

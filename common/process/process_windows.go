package process

import (
	"errors"
	"os"
	"slices"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

func WaitForProcess(proc *os.Process, duration *time.Duration) bool {
	handle, err := windows.OpenProcess(windows.SYNCHRONIZE, true, uint32(proc.Pid))
	if err != nil {
		return false
	}

	defer func(handle windows.Handle) {
		_ = windows.CloseHandle(handle)
	}(handle)

	var event uint32
	var waitMilliseconds uint32
	if duration == nil {
		waitMilliseconds = windows.INFINITE
	} else {
		waitMilliseconds = uint32(duration.Milliseconds())
	}
	event, err = windows.WaitForSingleObject(handle, waitMilliseconds)
	return err == nil && event == uint32(windows.WAIT_OBJECT_0)
}

// ProcessesPID returns a map of process names to their PIDs.
// Note: If multiple processes share the same name, only one PID is stored per name.
func ProcessesPID(names []string) map[string]uint32 {
	name := func(entry *windows.ProcessEntry32) string {
		return windows.UTF16ToString(entry.ExeFile[:])
	}
	entries := processesEntry(func(entry *windows.ProcessEntry32) bool {
		return slices.Contains(names, name(entry))
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

func GetProcessStartTime(pid int) (int64, error) {
	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		return 0, err
	}
	defer func(handle windows.Handle) {
		_ = windows.CloseHandle(handle)
	}(handle)

	var creationTime, exitTime, kernelTime, userTime windows.Filetime
	err = windows.GetProcessTimes(handle, &creationTime, &exitTime, &kernelTime, &userTime)
	if err != nil {
		return 0, err
	}
	return creationTime.Nanoseconds(), nil
}

func FindProcess(pid int) (proc *os.Process, err error) {
	return FindProcessWithStartTime(pid, 0)
}

func FindProcessWithStartTime(pid int, expectedStartTime int64) (proc *os.Process, err error) {
	proc, err = os.FindProcess(pid)
	if errors.Is(err, windows.ERROR_INVALID_PARAMETER) {
		err = nil
	}
	entries := processesEntry(func(entry *windows.ProcessEntry32) bool {
		return int(entry.ProcessID) == pid
	}, true)
	if len(entries) == 0 {
		proc = nil
		return
	}
	if err != nil {
		proc = &os.Process{Pid: pid}
		err = nil
	}
	if expectedStartTime != 0 {
		actualStartTime, startErr := GetProcessStartTime(pid)
		if startErr != nil {
			proc = nil
			err = startErr
			return
		}
		if actualStartTime != expectedStartTime {
			proc = nil
			err = errors.New("process start time mismatch")
			return
		}
	}
	return
}

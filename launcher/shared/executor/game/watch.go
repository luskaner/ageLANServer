package game

import (
	"syscall"
	"time"
	"unsafe"
)

const (
	Th32csSnapprocess = 0x00000002
	MaxPath           = 260
)

type PROCESSENTRY32 struct {
	dwSize              uint32
	cntUsage            uint32
	th32ProcessID       uint32
	th32DefaultHeapID   uintptr
	th32ModuleID        uint32
	cntThreads          uint32
	th32ParentProcessID uint32
	pcPriClassBase      int32
	dwFlags             uint32
	szExeFile           [MaxPath]uint16
}

var processes = []string{"AoE2DE.exe", "AoE2DE_s.exe"}

func processesExists() bool {
	kernel32 := syscall.MustLoadDLL("kernel32.dll")
	procProcess32First := kernel32.MustFindProc("Process32FirstW")
	procProcess32Next := kernel32.MustFindProc("Process32NextW")

	snapshot, err := syscall.CreateToolhelp32Snapshot(Th32csSnapprocess, 0)
	if err != nil {
		return false
	}
	defer func(handle syscall.Handle) {
		_ = syscall.CloseHandle(handle)
	}(snapshot)

	var procEntry PROCESSENTRY32
	procEntry.dwSize = uint32(unsafe.Sizeof(procEntry))

	r1, _, err := procProcess32First.Call(uintptr(snapshot), uintptr(unsafe.Pointer(&procEntry)))
	if r1 == 0 {
		return false
	}

	for {
		exeName := syscall.UTF16ToString(procEntry.szExeFile[:])
		for _, name := range processes {
			if exeName == name {
				return true
			}
		}

		r2, _, _ := procProcess32Next.Call(uintptr(snapshot), uintptr(unsafe.Pointer(&procEntry)))
		if r2 == 0 {
			break
		}
	}

	return false
}

func WaitUntilProcessesEnd(sleep time.Duration) {
	for {
		if !processesExists() {
			break
		}
		time.Sleep(sleep)
	}
}

func WaitUntilProcessesStart(sleep time.Duration, maxTimes int) bool {
	for i := 0; i < maxTimes; i++ {
		if processesExists() {
			return true
		}
		time.Sleep(sleep)
	}
	return false
}

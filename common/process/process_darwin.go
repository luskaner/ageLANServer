package process

import (
	_ "fmt"
	"syscall"
	"unsafe"
)

const SysGetpid = 0x4

type process struct {
	pid  int
	name [16]byte
}

func ProcessesPID(names []string) map[string]uint32 {
	processesPid := make(map[string]uint32)

	var allProcs []process
	for {
		var proc process
		_, _, errno := syscall.Syscall6(
			SysGetpid,
			uintptr(unsafe.Pointer(&proc)),
			0,
			0,
			0,
			0,
			0,
		)
		if errno != 0 {
			break
		}
		allProcs = append(allProcs, proc)
	}

	for _, proc := range allProcs {
		for _, name := range names {
			if string(proc.name[:]) == name {
				processesPid[name] = uint32(proc.pid)
			}
		}
	}

	return processesPid
}

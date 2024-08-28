package parentCheck

import (
	"fmt"
	"syscall"
	"unsafe"
)

const (
	CtlKern                = 1
	KernProc               = 14
	KernProcPathname       = 12
	ProcPidpathinfoMaxsize = 4096
)

func procPidPath(pid int, pathbuf []byte, buffersize int) (int, error) {
	mib := []int32{CtlKern, KernProc, KernProcPathname, int32(pid)}
	size := uintptr(buffersize)
	_, _, errno := syscall.Syscall6(syscall.SYS___SYSCTL, uintptr(unsafe.Pointer(&mib[0])), uintptr(len(mib)), uintptr(unsafe.Pointer(&pathbuf[0])), uintptr(unsafe.Pointer(&size)), 0, 0)
	if errno != 0 {
		return 0, errno
	}
	return int(size), nil
}

func exePathFromPID(pid int) (string, error) {
	pathbuf := make([]byte, ProcPidpathinfoMaxsize)
	ret, err := procPidPath(pid, pathbuf, ProcPidpathinfoMaxsize)
	if err != nil {
		return "", fmt.Errorf("PID %d: proc_pidpath: %s", pid, err.Error())
	}
	return string(pathbuf[:ret]), nil
}

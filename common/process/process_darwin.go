package process

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
	"golang.org/x/sys/unix"
)

const procAllPids = 1

var (
	libHandle       uintptr
	loadOnce        sync.Once
	loadErr         error
	procPidinfoPtr  func(int32, int32, uint64, uintptr, int32) int32
	procListpidsPtr func(uint32, uint32, uintptr, int32) int32
	procPidBsdInfo  int32
	procBsdInfoSize int32
	offsetPid       int
	offsetStartSec  int
	offsetStartUsec int
)

func loadLib() {
	loadOnce.Do(func() {
		h, err := purego.Dlopen("/usr/lib/libSystem.B.dylib", purego.RTLD_NOW)
		if err != nil {
			loadErr = err
			return
		}
		libHandle = h
		purego.RegisterLibFunc(&procPidinfoPtr, libHandle, "proc_pidinfo")
		purego.RegisterLibFunc(&procListpidsPtr, libHandle, "proc_listpids")

		pid := int32(os.Getpid())
		buf := make([]byte, 4096)

		for f := int32(0); f < 256; f++ {
			r := procPidinfoPtr(pid, f, 0, uintptr(unsafe.Pointer(&buf[0])), int32(len(buf)))
			if r > 0 {
				for off := 0; off+4 <= int(r); off += 4 {
					if uint32(pid) == *(*uint32)(unsafe.Pointer(&buf[off])) {
						procPidBsdInfo = f
						procBsdInfoSize = r
						offsetPid = off
						break
					}
				}
			}
			if procPidBsdInfo != 0 {
				break
			}
		}

		if procPidBsdInfo == 0 {
			loadErr = fmt.Errorf("could not detect PROC_PIDBSDINFO flavor")
			return
		}

		for off := offsetPid + 4; off+8 <= int(procBsdInfoSize); off += 8 {
			sec := *(*uint64)(unsafe.Pointer(&buf[off]))
			if sec > 1000000000 && sec < 5000000000 {
				offsetStartSec = off
				offsetStartUsec = off + 8
				break
			}
		}

		if offsetStartSec == 0 {
			loadErr = fmt.Errorf("could not detect start time offsets")
		}
	})
}

func GetProcessStartTime(pid int) (int64, error) {
	loadLib()
	if loadErr != nil {
		return 0, loadErr
	}

	buf := make([]byte, procBsdInfoSize)
	r := procPidinfoPtr(int32(pid), procPidBsdInfo, 0, uintptr(unsafe.Pointer(&buf[0])), procBsdInfoSize)
	if r != procBsdInfoSize {
		return 0, unix.EPERM
	}

	sec := *(*uint64)(unsafe.Pointer(&buf[offsetStartSec]))
	usec := *(*uint64)(unsafe.Pointer(&buf[offsetStartUsec]))

	return int64(sec*1e6 + usec), nil
}

func ProcessesByNames(names []string) map[string]*os.Process {
	result := make(map[string]*os.Process, len(names))
	if len(names) == 0 {
		return result
	}
	loadLib()
	if loadErr != nil || procListpidsPtr == nil || procPidinfoPtr == nil {
		for _, n := range names {
			result[n] = nil
		}
		return result
	}

	targets := make([]string, len(names))
	for i, n := range names {
		targets[i] = strings.ToLower(n)
		result[n] = nil
	}

	const maxPids = 16384
	pidBuf := make([]int32, maxPids)
	bufBytes := int32(len(pidBuf) * int(unsafe.Sizeof(pidBuf[0])))
	ret := procListpidsPtr(procAllPids, 0, uintptr(unsafe.Pointer(&pidBuf[0])), bufBytes)
	if ret <= 0 {
		return result
	}

	numPids := int(ret) / int(unsafe.Sizeof(pidBuf[0]))
	if numPids > len(pidBuf) {
		numPids = len(pidBuf)
	}

	for i := 0; i < numPids; i++ {
		pid := pidBuf[i]
		if pid <= 0 {
			continue
		}
		for ti := range targets {
			if result[names[ti]] != nil {
				continue
			}
			if strings.Contains(fmt.Sprintf("%d", pid), targets[ti]) {
				if proc, err := FindProcess(int(pid)); err == nil {
					result[names[ti]] = proc
				}
			}
		}
	}

	return result
}

package process

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/ebitengine/purego"
	"golang.org/x/sys/unix"
)

const (
	procAllPids    = 1
	procPidBsdInfo = 3
)

type ProcPidBsdInfo struct {
	PbiFlags     uint32
	PbiStatus    uint32
	PbiXstatus   uint32
	PbiPid       uint32
	PbiPpid      uint32
	PbiUid       uint32
	PbiGid       uint32
	PbiRuid      uint32
	PbiRgid      uint32
	PbiSvuid     uint32
	PbiSvgid     uint32
	Rfu1         uint32
	PbiComm      [16]byte
	PbiName      [1024]byte
	PbiNfiles    uint32
	PbiPgid      uint32
	PbiPjobc     uint32
	PbiTdev      uint32
	PbiTpgid     uint32
	PbiNice      uint32
	PbiStartSec  uint64
	PbiStartUsec uint64
}

var (
	libHandle       uintptr
	loadOnce        sync.Once
	loadErr         error
	procPidinfoPtr  func(int32, int32, uint64, uintptr, int32) int32
	procListpidsPtr func(uint32, uint32, uintptr, int32) int32
	procPidpathPtr  func(int32, uintptr, uint32) int32
)

func loadLib() {
	loadOnce.Do(func() {
		h, err := purego.Dlopen("/usr/lib/libSystem.B.dylib", purego.RTLD_NOW)
		if err != nil {
			loadErr = fmt.Errorf("dlopen libSystem failed: %w", err)
			return
		}
		libHandle = h
		purego.RegisterLibFunc(&procPidinfoPtr, libHandle, "proc_pidinfo")
		purego.RegisterLibFunc(&procListpidsPtr, libHandle, "proc_listpids")
		purego.RegisterLibFunc(&procPidpathPtr, libHandle, "proc_pidpath")
	})
}

func GetProcessStartTime(pid int) (int64, error) {
	loadLib()
	if loadErr != nil {
		return 0, loadErr
	}
	if procPidinfoPtr == nil {
		return 0, unix.EPERM
	}
	var info ProcPidBsdInfo
	size := int32(unsafe.Sizeof(info))
	ret := procPidinfoPtr(int32(pid), procPidBsdInfo, 0, uintptr(unsafe.Pointer(&info)), size)
	if ret != size {
		return 0, unix.EPERM
	}
	return (time.Duration(info.PbiStartSec)*time.Second + time.Duration(info.PbiStartUsec)*time.Microsecond).Microseconds(), nil
}

func ProcessesByNames(names []string) map[string]*os.Process {
	result := make(map[string]*os.Process, len(names))
	if len(names) == 0 {
		return result
	}
	loadLib()
	if loadErr != nil || procListpidsPtr == nil || procPidpathPtr == nil {
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
	remaining := make([]bool, len(names))
	for i := range remaining {
		remaining[i] = true
	}
	remainingCount := len(names)
	pathBuf := make([]byte, 4096)
	for i := 0; i < numPids && remainingCount > 0; i++ {
		pid := pidBuf[i]
		if pid <= 0 {
			continue
		}
		r := procPidpathPtr(pid, uintptr(unsafe.Pointer(&pathBuf[0])), uint32(len(pathBuf)))
		if r <= 0 {
			continue
		}
		path := strings.ToLower(string(pathBuf[:r]))
		base := strings.ToLower(filepath.Base(path))
		for ti, t := range targets {
			if !remaining[ti] {
				continue
			}
			if strings.Contains(path, t) || strings.Contains(base, t) {
				if proc, err := FindProcess(int(pid)); err == nil {
					result[names[ti]] = proc
				}
				remaining[ti] = false
				remainingCount--
			}
		}
	}
	return result
}

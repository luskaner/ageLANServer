package process

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
	"golang.org/x/sys/unix"
)

const (
	procAllPids = 1
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
	procPidBsdInfo  int32
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

		var info ProcPidBsdInfo
		size := int32(unsafe.Sizeof(info))
		buf := uintptr(unsafe.Pointer(&info))

		for f := int32(0); f < 20; f++ {
			r := procPidinfoPtr(int32(os.Getpid()), f, 0, buf, size)
			if r == size {
				procPidBsdInfo = f
				break
			}
		}

		if procPidBsdInfo == 0 {
			loadErr = fmt.Errorf("could not detect PROC_PIDBSDINFO flavor")
		}
	})
}

func GetProcessStartTime(pid int) (int64, error) {
	loadLib()
	if loadErr != nil {
		return 0, loadErr
	}
	var info ProcPidBsdInfo
	size := int32(unsafe.Sizeof(info))
	ret := procPidinfoPtr(int32(pid), procPidBsdInfo, 0, uintptr(unsafe.Pointer(&info)), size)
	if ret != size {
		return 0, unix.EPERM
	}
	return int64(info.PbiStartSec*1e6 + info.PbiStartUsec), nil
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
	remaining := make([]bool, len(names))
	for i := range remaining {
		remaining[i] = true
	}
	remainingCount := len(names)
	for i := 0; i < numPids && remainingCount > 0; i++ {
		pid := pidBuf[i]
		if pid <= 0 {
			continue
		}
		var info ProcPidBsdInfo
		size := int32(unsafe.Sizeof(info))
		r := procPidinfoPtr(pid, procPidBsdInfo, 0, uintptr(unsafe.Pointer(&info)), size)
		if r != size {
			continue
		}
		nameBytes := info.PbiName[:]
		if idx := bytes.IndexByte(nameBytes, 0); idx >= 0 {
			nameBytes = nameBytes[:idx]
		}
		name := strings.ToLower(string(nameBytes))
		for ti, t := range targets {
			if !remaining[ti] {
				continue
			}
			if strings.Contains(name, t) {
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

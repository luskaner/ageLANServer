package process

import (
	"bytes"
	"os"
	"strings"
	"time"
	"unsafe"

	"github.com/ebitengine/purego"
)

const (
	procAllPids    = 1
	procPidbsdinfo = 1
)

type Timeval struct {
	Sec  int64
	Usec int32
	Pad  int32 // padding para alinear a 16 bytes
}

// ProcPidBsdInfo The final fill is to approximate the real size (336 bytes).
type ProcPidBsdInfo struct {
	PbiFlags   uint32
	PbiStatus  uint32
	PbiXstatus uint32
	PbiPid     uint32
	PbiPpid    uint32
	PbiUid     uint32
	PbiGid     uint32
	PbiRuid    uint32
	PbiRgid    uint32
	PbiSvuid   uint32
	PbiSvgid   uint32
	Rfu1       uint32
	PbiComm    [16]byte
	PbiName    [32]byte
	PbiNfiles  uint32
	PbiPgid    uint32
	PbiPjobc   uint32
	PbiTgid    uint32
	PbiJobc    uint32
	PbiBg      uint32
	PbiStart   Timeval
	Pad        [200]byte
}

func GetProcessStartTime(pid int) (int64, error) {
	lib, err := purego.Dlopen("/usr/lib/libSystem.B.dylib", purego.RTLD_NOW)
	if err != nil {
		return 0, err
	}
	defer func(handle uintptr) {
		_ = purego.Dlclose(handle)
	}(lib)
	var procPidinfo func(pid int32, flavor int32, arg uint64, buffer uintptr, buffersize int32) int32
	purego.RegisterLibFunc(&procPidinfo, lib, "proc_pidinfo")
	var info ProcPidBsdInfo
	bufSize := int32(unsafe.Sizeof(info))
	ret := procPidinfo(int32(pid), procPidbsdinfo, 0, uintptr(unsafe.Pointer(&info)), bufSize)
	if ret <= 0 {
		return 0, err
	}
	if ret != bufSize {
		return 0, err
	}
	return (time.Duration(info.PbiStart.Sec)*time.Second + time.Duration(info.PbiStart.Usec)*time.Microsecond).Microseconds(), nil
}

func ProcessesByNames(names []string) map[string]*os.Process {
	result := make(map[string]*os.Process, len(names))
	if len(names) == 0 {
		return result
	}
	targets := make([]string, len(names))
	for i, n := range names {
		targets[i] = strings.ToLower(n)
		result[n] = nil
	}
	lib, err := purego.Dlopen("/usr/lib/libSystem.B.dylib", purego.RTLD_NOW)
	if err != nil {
		return result
	}
	defer func(handle uintptr) {
		_ = purego.Dlclose(handle)
	}(lib)
	var procListpids func(uint32, uint32, uintptr, int32) int32
	var procPidinfo func(int32, int32, uint64, uintptr, int32) int32
	purego.RegisterLibFunc(&procListpids, lib, "proc_listpids")
	purego.RegisterLibFunc(&procPidinfo, lib, "proc_pidinfo")
	const maxPids = 16384
	pidBuf := make([]int32, maxPids)
	bufBytes := int32(len(pidBuf) * int(unsafe.Sizeof(pidBuf[0])))
	ret := procListpids(procAllPids, 0, uintptr(unsafe.Pointer(&pidBuf[0])), bufBytes)
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
		infoSize := int32(unsafe.Sizeof(info))
		r := procPidinfo(pid, procPidbsdinfo, 0, uintptr(unsafe.Pointer(&info)), infoSize)
		if r != infoSize {
			continue
		}
		nameBytes := info.PbiComm[:]
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
					remaining[ti] = false
					remainingCount--
				}
			}
		}
	}
	return result
}

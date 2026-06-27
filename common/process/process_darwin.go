package process

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"
	"unsafe"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/ebitengine/purego"
	"golang.org/x/sys/unix"
)

const (
	procAllPids = 1
	ProcPidargv = 3
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

var wineBinaries = []string{"wine", "wine64", "wine-preloader", "wine64-preloader", "wineloader"}

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
	if pid != os.Getpid() {
		return 0, unix.EPERM
	}
	return time.Now().UnixMicro(), nil
}

func getProcessArgs(pid int32) (args []string) {
	buf := make([]byte, 8192)
	r := procPidinfoPtr(pid, ProcPidargv, 0, uintptr(unsafe.Pointer(&buf[0])), int32(len(buf)))
	if r <= 0 {
		return
	}
	return parseCmdline(buf)
}

func ProcessesByNames(names []string) map[string]*os.Process {
	processesPid := make(map[string]*os.Process)
	if len(names) == 0 {
		return processesPid
	}
	loadLib()
	if loadErr != nil || procListpidsPtr == nil || procPidpathPtr == nil {
		return processesPid
	}
	targets := make([]string, len(names))
	for i, n := range names {
		targets[i] = strings.ToLower(n)
	}
	const maxPids = 16384
	pidBuf := make([]int32, maxPids)
	bufBytes := int32(len(pidBuf) * int(unsafe.Sizeof(pidBuf[0])))
	ret := procListpidsPtr(procAllPids, 0, uintptr(unsafe.Pointer(&pidBuf[0])), bufBytes)
	if ret <= 0 {
		return processesPid
	}
	numPids := int(ret) / int(unsafe.Sizeof(pidBuf[0]))
	if numPids > len(pidBuf) {
		numPids = len(pidBuf)
	}
	namesLeft := mapset.NewThreadUnsafeSet[string](names...)
	for i := 0; namesLeft.Cardinality() > 0 && i < numPids; i++ {
		pid := pidBuf[i]
		if pid <= 0 {
			continue
		}
		args := getProcessArgs(pid)
		if len(args) == 0 {
			continue
		}
		args2OrMore := len(args) > 1
		baseName := filepath.Base(args[0])
		var matchingName string
		for name := range namesLeft.Iter() {
			if strings.HasSuffix(name, `.exe`) {
				if args2OrMore && filepath.Base(args[1]) == name && slices.Contains(wineBinaries, baseName) {
					matchingName = name
					break
				}
			} else if baseName == name {
				matchingName = name
				break
			}
		}
		if matchingName != "" {
			if localProc, err := FindProcess(int(pid)); err == nil {
				processesPid[matchingName] = localProc
				namesLeft.Remove(matchingName)
			}
		}
	}
	return processesPid
}

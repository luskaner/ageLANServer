package process

import (
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

func GetProcessStartTime(pid int) (int64, error) {
	var info struct {
		_           [40]byte // kp_proc offset
		P_starttime syscall.Timeval
		_           [476]byte // remaining fields
	}
	size := unsafe.Sizeof(info)
	mib := []int32{syscall.CTL_KERN, syscall.KERN_PROC, syscall.KERN_PROC_PID, int32(pid)}
	_, _, errno := syscall.Syscall6(
		syscall.SYS___SYSCTL,
		uintptr(unsafe.Pointer(&mib[0])),
		uintptr(len(mib)),
		uintptr(unsafe.Pointer(&info)),
		uintptr(unsafe.Pointer(&size)),
		0,
		0,
	)
	if errno != 0 {
		return 0, errno
	}
	return info.P_starttime.Nano(), nil
}

func ProcessesPID(names []string) map[string]uint32 {
	processesPid := make(map[string]uint32)

	output, err := exec.Command("ps", "-axo", "pid,comm").Output()
	if err != nil {
		return processesPid
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines[1:] { // Skip header
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		pid, err := strconv.ParseUint(fields[0], 10, 32)
		if err != nil {
			continue
		}
		cmdlineName := filepath.Base(fields[1])
		if slices.Contains(names, cmdlineName) {
			processesPid[cmdlineName] = uint32(pid)
		}
	}

	return processesPid
}

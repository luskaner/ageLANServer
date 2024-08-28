package parentCheck

import (
	"golang.org/x/sys/windows"
)

func exePathFromPID(pid int) (string, error) {
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_INFORMATION|windows.PROCESS_VM_READ, false, uint32(pid))
	if err != nil {
		return "", err
	}
	defer func(handle windows.Handle) {
		_ = windows.CloseHandle(handle)
	}(h)

	var buf [windows.MAX_PATH]uint16
	size := uint32(len(buf))
	err = windows.QueryFullProcessImageName(h, 0, &buf[0], &size)
	if err != nil {
		return "", err
	}

	return windows.UTF16ToString(buf[:size]), nil
}

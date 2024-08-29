package watch

import (
	"golang.org/x/sys/windows"
)

func waitForProcess(PID uint32) bool {
	handle, err := windows.OpenProcess(windows.SYNCHRONIZE, true, PID)

	if err != nil {
		return false
	}

	defer func(handle windows.Handle) {
		_ = windows.CloseHandle(handle)
	}(handle)

	_, err = windows.WaitForSingleObject(handle, windows.INFINITE)

	if err != nil {
		return false
	}

	return true
}

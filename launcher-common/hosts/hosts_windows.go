package hosts

import (
	"golang.org/x/sys/windows"
	"os"
)

const maxHostsPerLine = 9

var Lock *windows.Overlapped

func LockFile(file *os.File) (err error) {
	fileHandle := windows.Handle(file.Fd())
	Lock = &windows.Overlapped{}
	err = windows.LockFileEx(
		fileHandle,
		windows.LOCKFILE_EXCLUSIVE_LOCK,
		0,
		1,
		0,
		Lock,
	)
	return
}

func UnlockFile(file *os.File) (err error) {
	fileHandle := windows.Handle(file.Fd())
	err = windows.UnlockFileEx(fileHandle, 0, 1, 0, Lock)
	if err == nil {
		Lock = nil
	}
	return
}

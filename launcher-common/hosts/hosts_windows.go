package hosts

import (
	"os"
	"path/filepath"

	"golang.org/x/sys/windows"
)

const maxHostsPerLine = 9
const LineEnding = WindowsLineEnding

var Lock *windows.Overlapped

func Path() string {
	return filepath.Join(os.Getenv("WINDIR"), "System32", "drivers", "etc", "hosts")
}

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

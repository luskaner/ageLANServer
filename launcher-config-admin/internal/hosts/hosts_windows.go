package hosts

import (
	"github.com/luskaner/ageLANServer/launcher-common/executor/exec"
	"golang.org/x/sys/windows"
	"os"
	"path/filepath"
)

var lock *windows.Overlapped

func lockFile(file *os.File) (err error) {
	fileHandle := windows.Handle(file.Fd())
	lock = &windows.Overlapped{}
	err = windows.LockFileEx(
		fileHandle,
		windows.LOCKFILE_EXCLUSIVE_LOCK,
		0,
		1,
		0,
		lock,
	)
	return
}

func unlockFile(file *os.File) (err error) {
	fileHandle := windows.Handle(file.Fd())
	err = windows.UnlockFileEx(fileHandle, 0, 1, 0, lock)
	if err == nil {
		lock = nil
	}
	return
}

func flushDns() (result *exec.Result) {
	result = exec.Options{File: "ipconfig", SpecialFile: true, UseWorkingPath: true, ExitCode: true, Wait: true, Args: []string{"/flushdns"}}.Exec()
	return
}

func hostsPath() string {
	return filepath.Join(os.Getenv("WINDIR"), "System32", "drivers", "etc", "hosts")
}

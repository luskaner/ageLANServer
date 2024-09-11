//go:build !windows

package hosts

import (
	"github.com/luskaner/aoe2DELanServer/launcher-common/executor/exec"
	"golang.org/x/sys/unix"
	"os"
	"runtime"
)

var lock *unix.Flock_t

func lockFile(file *os.File) (err error) {
	lock = &unix.Flock_t{
		Type:   unix.F_WRLCK,
		Whence: 0,
		Start:  0,
		Len:    0,
	}
	err = unix.FcntlFlock(file.Fd(), unix.F_SETLK, lock)
	if err != nil {
		lock = &unix.Flock_t{}
	}
	return
}

func unlockFile(file *os.File) (err error) {
	lock.Type = unix.F_UNLCK
	err = unix.FcntlFlock(file.Fd(), unix.F_SETLK, lock)
	if err == nil {
		lock = &unix.Flock_t{}
	} else {
		lock.Type = unix.F_WRLCK
	}
	return
}

func flushDns() (result *exec.Result) {
	if runtime.GOOS == "darwin" {
		result = exec.Options{File: "dscacheutil", ExitCode: true, Wait: true, Args: []string{"-flushcache && killall -HUP mDNSResponder"}}.Exec()
	}
	return
}

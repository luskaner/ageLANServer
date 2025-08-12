//go:build !windows

package hosts

import (
	"golang.org/x/sys/unix"
	"os"
)

const maxHostsPerLine = 256

var Lock *unix.Flock_t

func LockFile(file *os.File) (err error) {
	Lock = &unix.Flock_t{
		Type:   unix.F_WRLCK,
		Whence: 0,
		Start:  0,
		Len:    0,
	}
	err = unix.FcntlFlock(file.Fd(), unix.F_SETLK, Lock)
	if err != nil {
		Lock = &unix.Flock_t{}
	}
	return
}

func UnlockFile(file *os.File) (err error) {
	Lock.Type = unix.F_UNLCK
	err = unix.FcntlFlock(file.Fd(), unix.F_SETLK, Lock)
	if err == nil {
		Lock = nil
	} else {
		Lock.Type = unix.F_WRLCK
	}
	return
}

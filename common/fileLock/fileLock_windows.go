package fileLock

import (
	"os"

	"golang.org/x/sys/windows"
)

type Lock struct {
	*BaseLock
	lock   *windows.Overlapped
	handle windows.Handle
}

func (l *Lock) Lock(f *os.File) error {
	var err error
	l.BaseLock = NewBaseLock(f)
	l.handle = windows.Handle(l.BaseLock.File.Fd())
	l.lock = &windows.Overlapped{}
	err = windows.LockFileEx(
		l.handle,
		windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY,
		0,
		0,
		0,
		l.lock,
	)
	if err != nil {
		_ = l.File.Close()
		l.clean()
		return err
	}
	return nil
}

func (l *Lock) Unlock() error {
	err := windows.UnlockFileEx(l.handle, 0, 0, 0, l.lock)
	if err != nil {
		return err
	}
	l.clean()
	return nil
}

func (l *Lock) clean() {
	l.close()
	l.handle = 0
	l.lock = nil
}

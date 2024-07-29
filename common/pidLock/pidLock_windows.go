package pidLock

import (
	"golang.org/x/sys/windows"
)

type Lock struct {
	Data
	lock   *windows.Overlapped
	handle windows.Handle
}

func (l *Lock) Lock() error {
	var err error
	err, l.file = openFile()
	if err != nil {
		return err
	}
	l.handle = windows.Handle(l.file.Fd())
	l.lock = &windows.Overlapped{}
	err = windows.LockFileEx(
		l.handle,
		windows.LOCKFILE_EXCLUSIVE_LOCK,
		0,
		0,
		0,
		l.lock,
	)
	if err != nil {
		_ = l.file.Close()
		l.file = nil
		l.lock = nil
		l.handle = 0
		return err
	}
	err = writePid(l.file)
	if err != nil {
		_ = l.file.Close()
		l.file = nil
		l.lock = nil
		l.handle = 0
		return err
	}
	return nil
}

func (l *Lock) Unlock() error {
	err := windows.UnlockFileEx(l.handle, 0, 0, 0, l.lock)
	if err != nil {
		return err
	}
	err = removeFile(l.file)
	if err != nil {
		return err
	}
	l.handle = 0
	l.lock = nil
	l.file = nil
	return nil
}

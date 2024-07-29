//go:build darwin || dragonfly || freebsd || illumos || linux || netbsd || openbsd || solaris

package pidLock

import "syscall"

type Lock struct {
	Data
	fd int
}

func (l *Lock) Lock() error {
	var err error
	err, l.file = openFile()
	if err != nil {
		return err
	}
	l.fd = int(l.file.Fd())
	err = syscall.Flock(l.fd, syscall.LOCK_SH|syscall.LOCK_NB)
	if err != nil {
		_ = l.file.Close()
		l.file = nil
		l.fd = 0
		return err
	}
	err = writePid(l.file)
	if err != nil {
		_ = l.file.Close()
		l.file = nil
		l.fd = 0
		return err
	}
	return nil
}

func (l *Lock) Unlock() error {
	err := syscall.Flock(l.fd, syscall.LOCK_UN)
	if err != nil {
		return err
	}
	err = removeFile(l.file)
	if err != nil {
		return err
	}
	l.fd = 0
	l.file = nil
	return nil
}

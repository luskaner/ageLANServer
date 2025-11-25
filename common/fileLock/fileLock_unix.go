//go:build darwin || dragonfly || freebsd || illumos || linux || netbsd || openbsd || solaris

package fileLock

import (
	"os"
	"syscall"
)

type Lock struct {
	*BaseLock
	fd int
}

func (l *Lock) Lock(f *os.File) error {
	var err error
	l.BaseLock = NewBaseLock(f)
	l.fd = int(l.File.Fd())
	err = syscall.Flock(l.fd, syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		l.clean()
		return err
	}
	return nil
}

func (l *Lock) Unlock() error {
	err := syscall.Flock(l.fd, syscall.LOCK_UN)
	if err != nil {
		return err
	}
	l.clean()
	return nil
}

func (l *Lock) clean() {
	l.close()
	l.fd = 0
}

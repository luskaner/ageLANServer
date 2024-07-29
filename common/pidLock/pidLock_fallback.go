//go:build !darwin && !dragonfly && !freebsd && !illumos && !linux && !netbsd && !openbsd && !solaris && !windows

package pidLock

type Lock struct {
	Data
}

func (l *Lock) Lock() error {
	var err error
	err, l.file = openFile()
	if err != nil {
		return err
	}
	err = writePid(l.file)
	if err != nil {
		return err
	}
	return nil
}

func (l *Lock) Unlock() error {
	err := removeFile(l.file)
	if err != nil {
		return err
	}
	l.file = nil
	return nil
}

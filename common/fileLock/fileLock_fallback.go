//go:build !darwin && !dragonfly && !freebsd && !illumos && !linux && !netbsd && !openbsd && !solaris && !windows

package fileLock

import (
	"os"
)

// Lock The fallback does nothing.
type Lock struct{}

func (l *Lock) Lock(_ *os.File) error {
	return nil
}

func (l *Lock) Unlock() error {
	return nil
}

func (l *Lock) clean() {}

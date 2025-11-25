package fileLock

import (
	"os"
)

type BaseLock struct {
	File *os.File
}

func (l *BaseLock) close() {
	if l.File != nil {
		_ = l.File.Close()
		l.File = nil
	}
}

func NewBaseLock(f *os.File) *BaseLock {
	return &BaseLock{File: f}
}

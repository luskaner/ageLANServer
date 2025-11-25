package fileLock

import (
	"os"
	"strconv"

	"github.com/luskaner/ageLANServer/common/process"
)

func openFile() (err error, f *os.File) {
	var exe string
	exe, err = os.Executable()
	if err != nil {
		return
	}
	var pidPath string
	var proc *os.Process
	pidPath, proc, err = process.Process(exe)
	if err == nil && proc != nil {
		return
	}
	f, err = os.OpenFile(pidPath, os.O_CREATE|os.O_WRONLY, 0644)
	return
}

func writePid(f *os.File) error {
	str := strconv.Itoa(os.Getpid())
	err := f.Truncate(int64(len(str)))
	if err != nil {
		return err
	}
	_, err = f.WriteString(str)
	if err != nil {
		return err
	}
	return f.Sync()
}

func removeFile(f *os.File) error {
	err := f.Close()
	if err != nil {
		return err
	}
	err = os.Remove(f.Name())
	if err != nil {
		return err
	}
	return nil
}

type PidLock struct {
	fileLock Lock
}

func (l *PidLock) Lock() error {
	err, file := openFile()
	if err != nil {
		return err
	}
	err = l.fileLock.Lock(file)
	if err != nil {
		return err
	}
	err = writePid(file)
	if err != nil {
		l.fileLock.Clean()
		return err
	}
	return nil
}

func (l *PidLock) Unlock() error {
	err := l.fileLock.Unlock()
	if err != nil {
		return err
	}
	err = removeFile(l.fileLock.BaseLock.File)
	if err != nil {
		return err
	}
	l.fileLock.Clean()
	return nil
}

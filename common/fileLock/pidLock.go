package fileLock

import (
	"encoding/binary"
	"os"

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
	pid := os.Getpid()
	startTime, err := process.GetProcessStartTime(pid)
	if err != nil {
		return err
	}

	data := make([]byte, process.PidFileSize)
	binary.LittleEndian.PutUint64(data[0:8], uint64(pid))
	binary.LittleEndian.PutUint64(data[8:16], uint64(startTime))

	err = f.Truncate(int64(len(data)))
	if err != nil {
		return err
	}
	_, err = f.Write(data)
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
	//goland:noinspection ALL
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
		l.fileLock.clean()
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
	l.fileLock.clean()
	return nil
}

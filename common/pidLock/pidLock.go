package pidLock

import (
	"common/process"
	"os"
	"strconv"
)

type lock interface {
	Lock() error
	Unlock() error
}

type Data struct {
	file *os.File
}

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
	_, err := f.WriteString(strconv.Itoa(os.Getpid()))
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

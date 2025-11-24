package pidLock

import (
	"os"
	"strconv"

	"github.com/luskaner/ageLANServer/common/process"
)

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

package executor

import (
	"golang.org/x/sys/windows"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

func ElevateCustomExecutable(executable string, arg ...string) bool {
	return ShellExecuteAndWait("runas", executable, arg...)
}

func ShellExecute(verb string, file string, executableWorkingPath bool, showWindow int, arg ...string) bool {
	verbPtr, _ := windows.UTF16PtrFromString(verb)
	exe, _ := windows.UTF16PtrFromString(file)
	args, _ := windows.UTF16PtrFromString(strings.Join(arg, " "))
	var workingDirPtr *uint16
	if executableWorkingPath {
		workingDirPtr, _ = syscall.UTF16PtrFromString(filepath.Dir(file))
	} else {
		workingDirPtr = nil
	}

	show := showWindow

	err := windows.ShellExecute(0, verbPtr, exe, args, workingDirPtr, int32(show))
	return err == nil
}

func StartCustomExecutable(executable string, executableWorkingPath bool, arg ...string) *exec.Cmd {
	cmd := exec.Command(executable, arg...)
	if executableWorkingPath {
		cmd.Dir = filepath.Dir(executable)
	}
	err := cmd.Start()
	if err != nil {
		return nil
	}
	return cmd
}

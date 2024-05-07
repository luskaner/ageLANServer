package executor

import (
	"golang.org/x/sys/windows"
	"os/exec"
	"path/filepath"
)

func RunCustomExecutable(executable string, arg ...string) bool {
	cmd := exec.Command(executable, arg...)
	err := cmd.Run()
	if err != nil {
		return false
	}
	return true
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

func IsAdmin() bool {
	var token windows.Token
	err := windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_QUERY, &token)
	if err != nil {
		return false
	}
	return token.IsElevated()
}

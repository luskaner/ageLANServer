package executor

import (
	"golang.org/x/sys/windows"
	"os/exec"
)

func RunCustomExecutable(executable string, arg ...string) bool {
	cmd := exec.Command(executable, arg...)
	err := cmd.Run()
	if err != nil {
		return false
	}
	return true
}

func IsAdmin() bool {
	var token windows.Token
	err := windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_QUERY, &token)
	if err != nil {
		return false
	}
	return token.IsElevated()
}

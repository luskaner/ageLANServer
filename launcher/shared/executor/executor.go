package executor

import (
	"golang.org/x/sys/windows"
	"golang.org/x/text/encoding/charmap"
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

func RunCustomExecutableOutput(executable string, arg ...string) *string {
	cmd := exec.Command(executable, arg...)
	output, err := cmd.Output()
	if err != nil {
		return nil
	}
	decoder := charmap.CodePage437.NewDecoder()
	out, err := decoder.Bytes(output)
	if err != nil {
		return nil
	}
	outStr := string(out)
	return &outStr
}

func IsAdmin() bool {
	var token windows.Token
	err := windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_QUERY, &token)
	if err != nil {
		return false
	}
	return token.IsElevated()
}

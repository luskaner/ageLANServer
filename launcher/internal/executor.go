package internal

import "os/exec"

func RunCustomExecutable(executable string, arg ...string) bool {
	cmd := exec.Command(executable, arg...)
	err := cmd.Run()
	if err != nil {
		return false
	}
	return true
}

func StartCustomExecutable(executable string, arg ...string) *exec.Cmd {
	cmd := exec.Command(executable, arg...)
	err := cmd.Start()
	if err != nil {
		return nil
	}
	return cmd
}

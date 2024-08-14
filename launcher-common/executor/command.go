package executor

import "golang.org/x/sys/windows"

func RunCommand(cmd []string) (err error) {
	var args []string
	if len(cmd) > 1 {
		args = cmd[1:]
	}
	result := ExecOptions{
		File:           cmd[0],
		WindowState:    windows.SW_NORMAL,
		SpecialFile:    true,
		Shell:          true,
		UseWorkingPath: true,
		Args:           args,
	}.Exec()
	err = result.Err
	return
}

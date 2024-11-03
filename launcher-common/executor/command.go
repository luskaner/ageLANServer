package executor

import (
	"github.com/luskaner/ageLANServer/launcher-common/executor/exec"
)

func RunRevertCommand(cmd []string) (err error) {
	var args []string
	if len(cmd) > 1 {
		args = cmd[1:]
	}
	result := exec.Options{
		File:           cmd[0],
		Wait:           true,
		UseWorkingPath: true,
		Args:           args,
	}.Exec()
	err = result.Err
	return
}

package executor

import (
	"github.com/luskaner/aoe2DELanServer/launcher-common/executor/exec"
)

func RunRevertCommand(cmd []string) (err error) {
	var args []string
	if len(cmd) > 1 {
		args = cmd[1:]
	}
	result := exec.Options{
		File:           cmd[0],
		SpecialFile:    true,
		Shell:          true,
		UseWorkingPath: true,
		Args:           args,
	}.Exec()
	err = result.Err
	return
}

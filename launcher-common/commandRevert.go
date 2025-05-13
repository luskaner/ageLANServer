package launcher_common

import (
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/launcher-common/executor/exec"
	"os"
	"path/filepath"
)

var RevertCommandStore = NewArgsStore(filepath.Join(os.TempDir(), common.Name+"_command_revert.txt"))

func RunRevertCommand() (err error) {
	var args []string
	var cmd []string
	err, cmd = RevertCommandStore.Load()
	if err != nil || len(cmd) == 0 {
		return
	}
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
	_ = RevertCommandStore.Delete()
	return
}

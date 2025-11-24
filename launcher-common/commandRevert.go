package launcher_common

import (
	"io"
	"os"
	"path/filepath"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executor/exec"
)

var RevertCommandStore = NewArgsStore(filepath.Join(os.TempDir(), common.Name+"_command_revert.txt"))

func RunRevertCommand(out io.Writer, optionsFn func(options exec.Options)) (err error) {
	var args []string
	var cmd []string
	err, cmd = RevertCommandStore.Load()
	if err != nil || len(cmd) == 0 {
		return
	}
	if len(cmd) > 1 {
		args = cmd[1:]
	}
	options := exec.Options{
		File:           cmd[0],
		Wait:           true,
		UseWorkingPath: true,
		Args:           args,
	}
	if optionsFn != nil {
		optionsFn(options)
	}
	if out != nil {
		options.Stdout = out
		options.Stderr = out
	}
	result := options.Exec()
	err = result.Err
	_ = RevertCommandStore.Delete()
	return
}

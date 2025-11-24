package launcher_common

import (
	"io"

	"github.com/luskaner/ageLANServer/common/executor/exec"
)

func RemoveBattleServerRegion(exe string, gameId string, region string, out io.Writer, optionsFn func(options exec.Options)) *exec.Result {
	options := exec.Options{
		File:     exe,
		Wait:     true,
		ExitCode: true,
		Args:     []string{"remove", "-e", gameId, "-r", region},
	}
	if optionsFn != nil {
		optionsFn(options)
	}
	if out != nil {
		options.Stdout = out
		options.Stderr = out
	}
	return options.Exec()
}

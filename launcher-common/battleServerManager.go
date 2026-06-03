package launcher_common

import (
	"io"

	"github.com/luskaner/ageLANServer/common/cmd"
	"github.com/luskaner/ageLANServer/common/cmd/bsManager"
	"github.com/luskaner/ageLANServer/common/executor/exec"
)

func RemoveBattleServerRegion(exe string, gameId string, region string, out io.Writer, optionsFn func(options exec.Options)) *exec.Result {
	values, flags := bsManager.RemoveFlagSet()
	values.GameIds = []string{gameId}
	values.Region = region
	options := exec.Options{
		File:     exe,
		Wait:     true,
		ExitCode: true,
		Args:     cmd.FlagSetToArgs(flags, true),
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

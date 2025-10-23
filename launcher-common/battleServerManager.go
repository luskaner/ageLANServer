package launcher_common

import "github.com/luskaner/ageLANServer/common/executor/exec"

func RemoveBattleServerRegion(exe string, gameId string, region string, optionsFn func(options exec.Options)) *exec.Result {
	options := exec.Options{
		File:     exe,
		Wait:     true,
		ExitCode: true,
		Args:     []string{"remove", "-e", gameId, "-r", region},
	}
	if optionsFn != nil {
		optionsFn(options)
	}
	return options.Exec()
}

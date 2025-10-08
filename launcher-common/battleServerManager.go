package launcher_common

import "github.com/luskaner/ageLANServer/common/executor/exec"

func RemoveBattleServerRegion(exe string, gameId string, region string) *exec.Result {
	return exec.Options{
		File:     exe,
		Wait:     true,
		ExitCode: true,
		Args:     []string{"remove", "-e", gameId, "-r", region},
	}.Exec()
}

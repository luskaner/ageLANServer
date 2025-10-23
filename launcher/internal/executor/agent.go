package executor

import (
	"strconv"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executor/exec"
)

func RunAgent(game string, steamProcess bool, xboxProcess bool, serverExe string, broadcastBattleServer bool,
	battleServerExe string, battleServerRegion string, optionsFn func(options exec.Options)) (result *exec.Result) {
	if serverExe == "" {
		serverExe = "-"
	}
	if battleServerExe == "" {
		battleServerExe = "-"
	}
	if battleServerRegion == "" {
		battleServerRegion = "-"
	}
	args := []string{
		strconv.FormatBool(steamProcess),
		strconv.FormatBool(xboxProcess),
		serverExe,
		strconv.FormatBool(broadcastBattleServer),
		game,
		battleServerExe,
		battleServerRegion,
	}
	options := exec.Options{File: common.GetExeFileName(false, common.LauncherAgent), Pid: true, Args: args}
	optionsFn(options)
	result = options.Exec()
	return
}

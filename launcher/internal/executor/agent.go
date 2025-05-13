package executor

import (
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/launcher-common/executor/exec"
	"strconv"
)

func RunAgent(game string, steamProcess bool, xboxProcess bool, serverExe string, broadCastBattleServer bool, revertCommand []string) (result *exec.Result) {
	if serverExe == "" {
		serverExe = "-"
	}
	args := []string{
		strconv.FormatBool(steamProcess),
		strconv.FormatBool(xboxProcess),
		serverExe,
		strconv.FormatBool(broadCastBattleServer),
		game,
		strconv.FormatUint(uint64(len(revertCommand)), 10),
	}
	args = append(args, revertCommand...)
	result = exec.Options{File: common.GetExeFileName(false, common.LauncherAgent), Pid: true, Args: args}.Exec()
	return
}

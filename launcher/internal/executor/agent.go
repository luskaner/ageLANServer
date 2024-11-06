package executor

import (
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/launcher-common/executor/exec"
	"strconv"
)

func RunAgent(game string, steamProcess bool, microsoftStoreProcess bool, serverExe string, broadCastBattleServer bool, revertCommand []string, revertFlags []string) (result *exec.Result) {
	if serverExe == "" {
		serverExe = "-"
	}
	args := []string{
		strconv.FormatBool(steamProcess),
		strconv.FormatBool(microsoftStoreProcess),
		serverExe,
		strconv.FormatBool(broadCastBattleServer),
		game,
		strconv.FormatUint(uint64(len(revertCommand)), 10),
	}
	args = append(args, revertCommand...)
	args = append(args, revertFlags...)
	result = exec.Options{ShowWindow: true, File: common.GetExeFileName(false, common.LauncherAgent), Pid: true, Args: args}.Exec()
	return
}

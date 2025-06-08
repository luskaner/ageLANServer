package executor

import (
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/launcher-common/executor/exec"
	"strconv"
)

func RunAgent(game string, steamProcess bool, xboxProcess bool, serverPid int, broadCastBattleServer bool) (result *exec.Result) {
	args := []string{
		strconv.FormatBool(steamProcess),
		strconv.FormatBool(xboxProcess),
		strconv.FormatInt(int64(serverPid), 10),
		strconv.FormatBool(broadCastBattleServer),
		game,
	}
	result = exec.Options{File: common.GetExeFileName(false, common.LauncherAgent), Pid: true, Args: args}.Exec()
	return
}

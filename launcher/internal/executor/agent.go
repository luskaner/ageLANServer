package executor

import (
	"common"
	"launcherCommon/executor"
	"strconv"
)

func RunAgent(processName string, serverExe string, broadCastBattleServer bool, unmapIPs bool, removeUserCert bool, removeLocalCert bool, restoreMetadata bool, restoreProfiles bool) (result *executor.ExecResult) {
	if serverExe == "" {
		serverExe = "-"
	}
	args := []string{processName, serverExe, strconv.FormatBool(broadCastBattleServer)}
	args = append(args, RevertFlags(unmapIPs, removeUserCert, removeLocalCert, restoreMetadata, restoreProfiles)...)
	result = executor.ExecOptions{File: common.GetExeFileName(false, common.LauncherAgent), Pid: true, Args: args}.Exec()
	return
}

package executor

import (
	"common"
	"launcherCommon/executor"
)

func RunAgent(processName string, serverExe string, canBroadCastBattleServer string, unmapIPs bool, removeUserCert bool, removeLocalCert bool, restoreMetadata bool, restoreProfiles bool) (result *executor.ExecResult) {
	if serverExe == "" {
		serverExe = "-"
	}
	args := []string{processName, serverExe, canBroadCastBattleServer}
	args = append(args, RevertFlags(unmapIPs, removeUserCert, removeLocalCert, restoreMetadata, restoreProfiles)...)
	result = executor.ExecOptions{File: common.GetExeFileName(false, common.LauncherAgent), Pid: true, Args: args}.Exec()
	return
}

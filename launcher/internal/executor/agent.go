package executor

import (
	"common"
	"launcherCommon/executor"
	"strconv"
)

func RunAgent(steamProcess bool, microsoftStoreProcess bool, serverExe string, broadCastBattleServer bool, unmapIPs bool, removeUserCert bool, removeLocalCert bool, restoreMetadata bool, restoreProfiles bool) (result *executor.ExecResult) {
	if serverExe == "" {
		serverExe = "-"
	}
	args := []string{strconv.FormatBool(steamProcess), strconv.FormatBool(microsoftStoreProcess), serverExe, strconv.FormatBool(broadCastBattleServer)}
	args = append(args, RevertFlags(unmapIPs, removeUserCert, removeLocalCert, restoreMetadata, restoreProfiles)...)
	result = executor.ExecOptions{File: common.GetExeFileName(false, common.LauncherAgent), Pid: true, Args: args}.Exec()
	return
}

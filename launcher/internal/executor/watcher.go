package executor

import (
	"common"
	"launcher-common/executor"
)

func RunWatcher(processName string, serverExe string, unmapIPs bool, removeUserCert bool, removeLocalCert bool, restoreMetadata bool, restoreProfiles bool) (result *executor.ExecResult) {
	args := []string{processName, serverExe}
	args = append(args, RevertFlags(unmapIPs, removeUserCert, removeLocalCert, restoreMetadata, restoreProfiles)...)
	result = executor.ExecOptions{File: common.GetExeFileName(false, common.LauncherWatcher), Pid: true, Args: args}.Exec()
	return
}

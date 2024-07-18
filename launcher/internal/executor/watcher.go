package executor

import (
	"common"
	"launcher-common/executor"
	"strconv"
)

func RunWatcher(processName string, serverPid uint32, unmapIPs bool, removeUserCert bool, removeLocalCert bool, restoreMetadata bool, restoreProfiles bool) (result *executor.ExecResult) {
	args := []string{processName, strconv.FormatUint(uint64(serverPid), 10)}
	args = append(args, RevertFlags(unmapIPs, removeUserCert, removeLocalCert, restoreMetadata, restoreProfiles)...)
	result = executor.ExecOptions{File: common.GetExeFileName(common.LauncherWatcher), Pid: true, Args: args}.Exec()
	return
}

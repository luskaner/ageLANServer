package executor

import (
	"github.com/luskaner/aoe2DELanServer/common"
	"github.com/luskaner/aoe2DELanServer/launcher-common/executor/exec"
	"strconv"
)

func RunAgent(steamProcess bool, microsoftStoreProcess bool, serverExe string, broadCastBattleServer bool, revertCommand []string, unmapIPs bool, removeUserCert bool, removeLocalCert bool, restoreMetadata bool, restoreProfiles bool, unmapCDN bool) (result *exec.Result) {
	if serverExe == "" {
		serverExe = "-"
	}
	args := []string{strconv.FormatBool(steamProcess), strconv.FormatBool(microsoftStoreProcess), serverExe, strconv.FormatBool(broadCastBattleServer), strconv.FormatUint(uint64(len(revertCommand)), 10)}
	args = append(args, revertCommand...)
	args = append(args, RevertFlags(unmapIPs, removeUserCert, removeLocalCert, restoreMetadata, restoreProfiles, unmapCDN)...)
	result = exec.Options{File: common.GetExeFileName(false, common.LauncherAgent), Pid: true, Args: args}.Exec()
	return
}

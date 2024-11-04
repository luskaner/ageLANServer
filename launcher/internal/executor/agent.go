package executor

import (
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/launcher-common/executor/exec"
	"strconv"
)

func RunAgent(game string, steamProcess bool, microsoftStoreProcess bool, serverExe string, broadCastBattleServer bool, revertCommand []string, unmapIPs bool, removeUserCert bool, removeLocalCert bool, restoreMetadata bool, restoreProfiles bool, unmapCDN bool) (result *exec.Result) {
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
	if unmapCDN || unmapIPs || removeUserCert || removeLocalCert || restoreMetadata || restoreProfiles {
		args = append(
			args,
			RevertFlags(game, unmapIPs, removeUserCert, removeLocalCert, restoreMetadata, restoreProfiles, unmapCDN)...,
		)
	}
	result = exec.Options{File: common.GetExeFileName(false, common.LauncherAgent), Pid: true, Args: args}.Exec()
	return
}

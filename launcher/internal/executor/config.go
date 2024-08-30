package executor

import (
	"encoding/base64"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/aoe2DELanServer/common"
	launcherCommon "github.com/luskaner/aoe2DELanServer/launcher-common"
	"github.com/luskaner/aoe2DELanServer/launcher-common/executor/exec"
)

func RunSetUp(mapIps mapset.Set[string], addUserCertData []byte, addLocalCertData []byte, backupMetadata bool, backupProfiles bool, mapCDN bool, exitAgentOnError bool) (result *exec.Result) {
	args := make([]string, 0)
	args = append(args, "setup")
	if !exec.IsAdmin() {
		args = append(args, "-g")
		if exitAgentOnError {
			args = append(args, "-e")
		}
	}
	if mapIps != nil {
		for ip := range mapIps.Iter() {
			args = append(args, "-i")
			args = append(args, ip)
		}
	}
	if addLocalCertData != nil {
		args = append(args, "-l")
		args = append(args, base64.StdEncoding.EncodeToString(addLocalCertData))
	}
	if addUserCertData != nil {
		args = append(args, "-u")
		args = append(args, base64.StdEncoding.EncodeToString(addUserCertData))
	}
	if backupMetadata {
		args = append(args, "-m")
	}
	if backupProfiles {
		args = append(args, "-p")
	}
	if mapCDN {
		args = append(args, "-c")
	}
	result = exec.Options{File: common.GetExeFileName(false, common.LauncherConfig), Wait: true, Args: args, ExitCode: true}.Exec()
	return
}

func RunRevert(unmapIPs bool, removeUserCert bool, removeLocalCert bool, restoreMetadata bool, restoreProfiles bool, unmapCDN bool) (result *exec.Result) {
	args := []string{launcherCommon.ConfigRevertCmd}
	args = append(args, RevertFlags(unmapIPs, removeUserCert, removeLocalCert, restoreMetadata, restoreProfiles, unmapCDN)...)
	result = exec.Options{File: common.GetExeFileName(false, common.LauncherConfig), Wait: true, Args: args, ExitCode: true}.Exec()
	return
}

func RevertFlags(unmapIPs bool, removeUserCert bool, removeLocalCert bool, restoreMetadata bool, restoreProfiles bool, unmapCDN bool) []string {
	args := make([]string, 0)
	if !exec.IsAdmin() {
		args = append(args, "-g")
	}
	if unmapIPs {
		args = append(args, "-i")
	}
	if removeUserCert {
		args = append(args, "-u")
	}
	if removeLocalCert {
		args = append(args, "-l")
	}
	if restoreMetadata {
		args = append(args, "-m")
	}
	if restoreProfiles {
		args = append(args, "-p")
	}
	if unmapCDN {
		args = append(args, "-c")
	}
	return args
}

package executor

import (
	"encoding/base64"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/aoe2DELanServer/common"
	launcherCommon "github.com/luskaner/aoe2DELanServer/launcherCommon"
	"github.com/luskaner/aoe2DELanServer/launcherCommon/executor"
)

func RunSetUp(mapIps mapset.Set[string], addUserCertData []byte, addLocalCertData []byte, backupMetadata bool, backupProfiles bool, exitAgentOnError bool) (result *executor.ExecResult) {
	args := make([]string, 0)
	args = append(args, "setup")
	if !executor.IsAdmin() {
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
	result = executor.ExecOptions{File: common.GetExeFileName(false, common.LauncherConfig), Wait: true, Args: args, ExitCode: true}.Exec()
	return
}

func RunRevert(unmapIPs bool, removeUserCert bool, removeLocalCert bool, restoreMetadata bool, restoreProfiles bool) (result *executor.ExecResult) {
	args := []string{launcherCommon.ConfigRevertCmd}
	args = append(args, RevertFlags(unmapIPs, removeUserCert, removeLocalCert, restoreMetadata, restoreProfiles)...)
	result = executor.ExecOptions{File: common.GetExeFileName(false, common.LauncherConfig), Wait: true, Args: args, ExitCode: true}.Exec()
	return
}

func RevertFlags(unmapIPs bool, removeUserCert bool, removeLocalCert bool, restoreMetadata bool, restoreProfiles bool) []string {
	args := make([]string, 0)
	if !executor.IsAdmin() {
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
	return args
}

package executor

import (
	"encoding/base64"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executor"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/luskaner/ageLANServer/launcher-common/executor/exec"
	"github.com/luskaner/ageLANServer/launcher/internal/server/certStore"
	"runtime"
)

func RunSetUp(game string, mapIps mapset.Set[string], addUserCertData []byte, addLocalCertData []byte, backupMetadata bool, backupProfiles bool, mapCDN bool, exitAgentOnError bool, hostFilePath string, certFilePath string) (result *exec.Result) {
	reloadSystemCertificates := false
	reloadHostMappings := false
	args := make([]string, 0)
	args = append(args, "setup")
	if game != "" {
		args = append(args, "-e")
		args = append(args, game)
	}
	if !executor.IsAdmin() {
		args = append(args, "-g")
		if exitAgentOnError {
			args = append(args, "-r")
		}
	}
	if mapIps != nil {
		for ip := range mapIps.Iter() {
			args = append(args, "-i")
			args = append(args, ip)
			reloadHostMappings = true
		}
	}
	if addLocalCertData != nil {
		reloadSystemCertificates = true
		args = append(args, "-l")
		args = append(args, base64.StdEncoding.EncodeToString(addLocalCertData))
	}
	if addUserCertData != nil {
		reloadSystemCertificates = true
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
		reloadHostMappings = true
	}
	if hostFilePath != "" {
		args = append(args, "-o")
		args = append(args, hostFilePath)
	}
	if certFilePath != "" {
		args = append(args, "-t")
		args = append(args, certFilePath)
	}
	result = exec.Options{File: common.GetExeFileName(false, common.LauncherConfig), Wait: true, Args: args, ExitCode: true}.Exec()
	if reloadSystemCertificates {
		certStore.ReloadSystemCertificates()
	}
	if reloadHostMappings {
		launcherCommon.ClearCache()
	}
	return
}

func RunRevert(game string, unmapIPs bool, removeUserCert bool, removeLocalCert bool, restoreMetadata bool, restoreProfiles bool, unmapCDN bool, hostFilePath string, certFilePath string, failfast bool) (result *exec.Result) {
	args := []string{launcherCommon.ConfigRevertCmd}
	args = append(args, RevertFlags(game, unmapIPs, removeUserCert, removeLocalCert, restoreMetadata, restoreProfiles, unmapCDN, hostFilePath, certFilePath, failfast)...)
	result = exec.Options{File: common.GetExeFileName(false, common.LauncherConfig), Wait: true, Args: args, ExitCode: true}.Exec()
	if removeUserCert || removeLocalCert {
		certStore.ReloadSystemCertificates()
	}
	if unmapIPs || unmapCDN {
		launcherCommon.ClearCache()
	}
	return
}

func RevertFlags(game string, unmapIPs bool, removeUserCert bool, removeLocalCert bool, restoreMetadata bool, restoreProfiles bool, unmapCDN bool, hostFilePath string, certFilePath string, failfast bool) []string {
	args := make([]string, 0)
	args = append(args, "-e")
	args = append(args, game)
	if !executor.IsAdmin() {
		args = append(args, "-g")
	}
	if !failfast && unmapIPs && (runtime.GOOS == "linux" || removeLocalCert) && removeLocalCert && restoreMetadata && restoreProfiles && unmapCDN {
		args = append(args, "-a")
	} else {
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
	}
	if hostFilePath != "" {
		args = append(args, "-o")
		args = append(args, hostFilePath)
	}
	if certFilePath != "" {
		args = append(args, "-t")
		args = append(args, certFilePath)
	}
	return args
}

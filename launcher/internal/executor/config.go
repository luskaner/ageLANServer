package executor

import (
	"encoding/base64"
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executor"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/luskaner/ageLANServer/launcher-common/executor/exec"
	"github.com/luskaner/ageLANServer/launcher/internal/server/certStore"
	"slices"
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
	if result.Success() {
		revertArgs := launcherCommon.RevertFlags(
			game,
			mapIps != nil && mapIps.Cardinality() > 0,
			addUserCertData != nil,
			addLocalCertData != nil,
			backupMetadata,
			backupProfiles,
			mapCDN,
			hostFilePath,
			certFilePath,
			launcherCommon.RequiresStopConfigAgent(args),
			true,
		)
		if err := launcherCommon.StoreRevertArgs(revertArgs); err != nil {
			fmt.Println("Failed to store revert arguments, reverting setup...")
			result = RunRevert(revertArgs, false)
			if !result.Success() {
				fmt.Println("Failed to revert setup.")
			}
			result.Err = err
		}
	}
	return
}

func RunRevert(flags []string, bin bool) (result *exec.Result) {
	result = launcherCommon.RunRevert(flags, bin)
	if slices.Contains(flags, "-a") || slices.Contains(flags, "-u") || slices.Contains(flags, "-l") {
		certStore.ReloadSystemCertificates()
	}
	if slices.Contains(flags, "-a") || slices.Contains(flags, "-i") || slices.Contains(flags, "-c") {
		launcherCommon.ClearCache()
	}
	return
}

package executor

import (
	"encoding/base64"
	"fmt"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executor"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/luskaner/ageLANServer/launcher-common/executor/exec"
	"github.com/luskaner/ageLANServer/launcher/internal/server/certStore"
	"slices"
)

type RunSetUpOptions struct {
	Game                   string
	HostFilePath           string
	MapIp                  string
	MapCDN                 bool
	CertFilePath           string
	AddUserCertData        []byte
	AddLocalCertData       []byte
	WindowsUserProfilePath string
	BackupMetadata         bool
	ExitAgentOnError       bool
}

func (options *RunSetUpOptions) revertFlagsOptions() *launcherCommon.RevertFlagsOptions {
	var stopAgent bool
	if _, _, err := launcherCommon.ConfigAdminAgent(false); err == nil {
		stopAgent = true
	}
	return &launcherCommon.RevertFlagsOptions{
		Game:                   options.Game,
		HostFilePath:           options.HostFilePath,
		UnmapIP:                options.MapIp != "",
		UnmapCDN:               options.MapCDN,
		CertFilePath:           options.CertFilePath,
		RemoveUserCert:         options.AddUserCertData != nil,
		RemoveLocalCert:        options.AddLocalCertData != nil,
		WindowsUserProfilePath: options.WindowsUserProfilePath,
		RestoreMetadata:        options.BackupMetadata,
		StopAgent:              stopAgent,
		Failfast:               true,
	}
}

func RunSetUp(options *RunSetUpOptions) (result *exec.Result) {
	reloadSystemCertificates := false
	reloadHostMappings := false
	args := make([]string, 0)
	args = append(args, "setup")
	if options.Game != "" {
		args = append(args, "-e")
		args = append(args, options.Game)
	}
	if !executor.IsAdmin() {
		args = append(args, "-g")
		if options.ExitAgentOnError {
			args = append(args, "-r")
		}
	}
	if options.MapIp != "" {
		args = append(args, "-i")
		args = append(args, options.MapIp)
		reloadHostMappings = true
	}
	if options.AddLocalCertData != nil {
		reloadSystemCertificates = true
		args = append(args, "-l")
		args = append(args, base64.StdEncoding.EncodeToString(options.AddLocalCertData))
	}
	if options.AddUserCertData != nil {
		reloadSystemCertificates = true
		args = append(args, "-u")
		args = append(args, base64.StdEncoding.EncodeToString(options.AddUserCertData))
	}
	if options.BackupMetadata {
		args = append(args, "-m")
	}
	if options.WindowsUserProfilePath != "" {
		args = append(args, "-s")
		args = append(args, options.WindowsUserProfilePath)
	}
	if options.MapCDN {
		args = append(args, "-c")
		reloadHostMappings = true
	}
	if options.HostFilePath != "" {
		args = append(args, "-o")
		args = append(args, options.HostFilePath)
	}
	if options.CertFilePath != "" {
		args = append(args, "-t")
		args = append(args, options.CertFilePath)
	}
	result = exec.Options{File: common.GetExeFileName(false, common.LauncherConfig), Wait: true, Args: args, ExitCode: true}.Exec()
	if reloadSystemCertificates {
		certStore.ReloadSystemCertificates()
	}
	if reloadHostMappings {
		launcherCommon.ClearCache()
	}
	if result.Success() {
		revertArgs := launcherCommon.RevertFlags(options.revertFlagsOptions())
		if err := launcherCommon.RevertConfigStore.Store(revertArgs); err != nil {
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

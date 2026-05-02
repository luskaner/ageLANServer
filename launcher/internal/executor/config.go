package executor

import (
	"encoding/base64"
	"io"
	"slices"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executables"
	"github.com/luskaner/ageLANServer/common/executor"
	"github.com/luskaner/ageLANServer/common/executor/exec"
	commonLogger "github.com/luskaner/ageLANServer/common/logger"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/luskaner/ageLANServer/launcher/internal/cmdUtils/logger"
	"github.com/luskaner/ageLANServer/launcher/internal/server/certStore"
)

type ConfigSetupOptions struct {
	GameId           string
	MapIp            string
	AddUserCertData  []byte
	AddLocalCertData []byte
	AddGameCertData  []byte
	BackupMetadata   bool
	BackupProfiles   bool
	ExitAgentOnError bool
	HostFilePath     string
	CertFilePath     string
	GameBinPath      string
	GameDataPath     string
	Out              io.Writer
	OptionsFn        func(options exec.Options)
}

func (c *ConfigSetupOptions) ConfigRevertFlagOptions(args []string) *launcherCommon.ConfigRevertFlagOptions {
	return &launcherCommon.ConfigRevertFlagOptions{
		GameId:          c.GameId,
		UnmapIPs:        c.MapIp != "",
		RemoveUserCert:  c.AddUserCertData != nil,
		RemoveLocalCert: c.AddLocalCertData != nil,
		RestoreGameCert: c.AddGameCertData != nil,
		RestoreMetadata: c.BackupMetadata,
		RestoreProfiles: c.BackupProfiles,
		HostFilePath:    c.HostFilePath,
		CertFilePath:    c.CertFilePath,
		GameBinPath:     c.GameBinPath,
		GameDataPath:    c.GameDataPath,
		LogRoot:         commonLogger.FileLogger.Folder(),
		StopAgent:       launcherCommon.RequiresStopConfigAgent(args),
		FailFast:        true,
	}
}

func (c *ConfigSetupOptions) RunSetUp() (result *exec.Result) {
	reloadSystemCertificates := false
	reloadHostMappings := false
	args := make([]string, 0)
	args = append(args, "setup")
	if c.GameId != "" {
		args = append(args, "-e")
		args = append(args, c.GameId)
	}
	if !executor.IsAdmin() {
		args = append(args, "-g")
		if c.ExitAgentOnError {
			args = append(args, "-r")
		}
	}
	if c.MapIp != "" {
		args = append(args, "-i")
		args = append(args, c.MapIp)
		reloadHostMappings = true
	}
	if c.AddLocalCertData != nil {
		reloadSystemCertificates = true
		args = append(args, "-l")
		args = append(args, base64.StdEncoding.EncodeToString(c.AddLocalCertData))
	}
	if c.AddUserCertData != nil {
		reloadSystemCertificates = true
		args = append(args, "-u")
		args = append(args, base64.StdEncoding.EncodeToString(c.AddUserCertData))
	}
	if c.AddGameCertData != nil {
		args = append(args, "-s")
		args = append(args, base64.StdEncoding.EncodeToString(c.AddGameCertData))
	}
	if c.BackupMetadata {
		args = append(args, "-m")
	}
	if c.BackupProfiles {
		args = append(args, "-p")
	}
	if c.HostFilePath != "" {
		args = append(args, "-o")
		args = append(args, c.HostFilePath)
	}
	if c.CertFilePath != "" {
		args = append(args, "-t")
		args = append(args, c.CertFilePath)
	}
	if c.GameDataPath != "" {
		args = append(args, "--dataPath")
		args = append(args, c.GameDataPath)
	}
	if c.GameBinPath != "" {
		args = append(args, "--gamePath")
		args = append(args, c.GameBinPath)
	}
	if logRoot := commonLogger.FileLogger.Folder(); logRoot != "" {
		args = append(args, "--logRoot", logRoot)
	}
	options := exec.Options{File: executables.NativeFileName(false, executables.LauncherConfig), Wait: true, Args: args, ExitCode: true}
	c.OptionsFn(options)
	if c.Out != nil {
		options.Stdout = c.Out
		options.Stderr = c.Out
	}
	result = options.Exec()
	if reloadSystemCertificates {
		certStore.ReloadSystemCertificates()
	}
	if reloadHostMappings {
		common.ClearCache()
	}
	if result.Success() {
		revertArgs := c.ConfigRevertFlagOptions(args).Flags()
		if err := launcherCommon.RevertConfigStore.Store(revertArgs); err != nil {
			logger.Println("Failed to store revert arguments, reverting setup...")
			result = RunRevert(revertArgs, false, c.Out, c.OptionsFn)
			if !result.Success() {
				logger.Println("Failed to revert setup.")
			}
			result.Err = err
		}
	}
	return
}

func RunRevert(flags []string, bin bool, out io.Writer, optionFn func(options exec.Options)) (result *exec.Result) {
	result = launcherCommon.RunRevert(flags, bin, out, optionFn)
	if slices.Contains(flags, "-a") || slices.Contains(flags, "-u") || slices.Contains(flags, "-l") {
		certStore.ReloadSystemCertificates()
	}
	if slices.Contains(flags, "-a") || slices.Contains(flags, "-i") || slices.Contains(flags, "-c") {
		common.ClearCache()
	}
	return
}

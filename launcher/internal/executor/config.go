package executor

import (
	"io"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/cmd"
	"github.com/luskaner/ageLANServer/common/executables"
	"github.com/luskaner/ageLANServer/common/executor/exec"
	commonLogger "github.com/luskaner/ageLANServer/common/logger"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/luskaner/ageLANServer/launcher-common/cmd/config"
	"github.com/luskaner/ageLANServer/launcher/internal/cmdUtils/logger"
	"github.com/luskaner/ageLANServer/launcher/internal/server/certStore"
	"github.com/spf13/pflag"
)

type ConfigSetupOptions struct {
	*config.SetupValues
	flags     *pflag.FlagSet
	Out       io.Writer
	OptionsFn func(options exec.Options)
}

func NewConfigSetupOptions() *ConfigSetupOptions {
	setupValues, flags := config.RegularSetUpFlagSet()
	return &ConfigSetupOptions{
		flags:       flags,
		SetupValues: setupValues,
	}
}

func (c *ConfigSetupOptions) ConfigRevertFlagOptions() *launcherCommon.ConfigRevertFlagOptions {
	options := launcherCommon.NewConfigRevertFlagOptions()
	options.UnmapIPs = c.MapIp != nil
	options.RemoveLocalCert = c.AddLocalCertData != nil
	options.RemoveUserCert = c.AddUserCertData != nil
	options.RestoreCAStoreCert = c.AddCACertData != nil
	options.StopAgent = c.AgentStart
	options.GameId = c.GameId
	options.LogRoot = c.LogRoot
	options.CertFilePath = c.CertFilePath
	options.HostFilePath = c.HostFilePath
	options.DataPath = c.DataPath
	options.GamePath = c.GamePath
	options.Metadata = c.Metadata
	options.Profiles = c.Profiles
	return options
}

func (c *ConfigSetupOptions) shouldStartAgent() bool {
	if !launcherCommon.RequiresAdminElevation(false) {
		return false
	}
	return (c.AddLocalCertData != nil && c.CertFilePath == "") ||
		(c.MapIp != nil && c.HostFilePath == "")
}

func (c *ConfigSetupOptions) RunSetUp() (result *exec.Result) {
	reloadSystemCertificates := false
	reloadHostMappings := false
	if logRoot := commonLogger.FileLogger.Folder(); logRoot != "" {
		c.LogRoot = logRoot
	}
	if c.AgentStart = c.shouldStartAgent(); c.AgentStart {
		c.AgentEndOnError = true
	}
	args := cmd.FlagSetToArgs(c.flags, true)
	if c.MapIp != nil {
		reloadHostMappings = true
	}
	if c.AddLocalCertData != nil || c.AddUserCertData != nil {
		reloadSystemCertificates = true
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
		common.ClearDNSCache()
	}
	if result.Success() {
		revertArgs := c.ConfigRevertFlagOptions().Flags()
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
	values, flagSet := config.RegularRevertFlagSet()
	if err := flagSet.Parse(flags); err != nil {
		return &exec.Result{Err: err}
	}
	result = launcherCommon.RunRevert(flags, bin, out, optionFn)
	if values.RemoveAll || values.RemoveUserCert || values.RemoveLocalCert {
		certStore.ReloadSystemCertificates()
	}
	if values.RemoveAll || values.UnmapIPs {
		common.ClearDNSCache()
	}
	return
}

package executor

import (
	"io"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/cmd"
	"github.com/luskaner/ageLANServer/common/executables"
	"github.com/luskaner/ageLANServer/common/executor"
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
	setupValues, flags := config.SetUpFlagSet()
	return &ConfigSetupOptions{
		flags:       flags,
		SetupValues: setupValues,
	}
}

func (c *ConfigSetupOptions) ConfigRevertFlagOptions() *launcherCommon.ConfigRevertFlagOptions {
	options := launcherCommon.NewConfigRevertFlagOptions()
	options.IPs = c.MapIp != nil
	options.Certs = c.AddLocalCertData != nil
	options.RemoveUserCert = c.AddUserCertData != nil
	options.RestoreCAStoreCert = c.AddCACertData != nil
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

func (c *ConfigSetupOptions) RunSetUp() (result *exec.Result) {
	reloadSystemCertificates := false
	reloadHostMappings := false
	if logRoot := commonLogger.FileLogger.Folder(); logRoot != "" {
		c.LogRoot = logRoot
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
	values, flagSet := config.RevertFlagSet()
	if err := flagSet.Parse(flags); err != nil {
		return &exec.Result{Err: err}
	}
	result = launcherCommon.RunRevert(flags, bin, out, optionFn)
	if values.RemoveAll || values.RemoveUserCert || values.Certs {
		certStore.ReloadSystemCertificates()
	}
	if values.RemoveAll || values.IPs {
		common.ClearDNSCache()
	}
	return
}

type ConfigFlushCacheOptions struct {
	*config.FlushCacheValues
	flags *pflag.FlagSet
}

func NewConfigFlushCacheOptions(canAddHost bool, canTrustCertificate string, customHostFile bool, customCertFile bool) *ConfigFlushCacheOptions {
	ips := !customHostFile && canAddHost
	certs := !customCertFile && canTrustCertificate != "false"
	if !ips && !certs {
		return nil
	}
	flushCacheValues, flags := config.FlushCacheFlagSet()
	flushCacheValues.IPs = ips
	flushCacheValues.Certs = certs
	return &ConfigFlushCacheOptions{
		flags:            flags,
		FlushCacheValues: flushCacheValues,
	}
}

func (c *ConfigFlushCacheOptions) RunFlushCache() (result *exec.Result) {
	reloadSystemCertificates := false
	reloadHostMappings := false
	if logRoot := commonLogger.FileLogger.Folder(); logRoot != "" {
		c.LogRoot = logRoot
	}
	if c.IPs {
		reloadHostMappings = true
	}
	if c.Certs {
		reloadSystemCertificates = true
	}
	str := "Flushing cache"
	options := exec.Options{File: executables.NativeFileName(false, executables.LauncherConfig), Wait: true, Args: cmd.FlagSetToArgs(c.flags, true), ExitCode: true}
	if executor.IsAdmin() {
		if err := commonLogger.FileLogger.Buffer("config_flushCache", func(writer io.Writer) {
			options.Stdout = writer
			options.Stderr = writer
		}); err != nil {
			return &exec.Result{ExitCode: common.ErrFileLog}
		}
	} else {
		str += ", authorize 'config-admin-agent' if needed"
	}
	str += "..."
	logger.Println(str)
	commonLogger.Println("run config flushCache", options.String())
	result = options.Exec()
	if reloadSystemCertificates {
		certStore.ReloadSystemCertificates()
	}
	if reloadHostMappings {
		common.ClearDNSCache()
	}
	if !result.Success() {
		commonLogger.Println("Failed to flush cache")
		if result.Err != nil {
			commonLogger.Println("Received error:")
			commonLogger.Println(result.Err)
		}
		if result.ExitCode != common.ErrSuccess {
			commonLogger.Println("Received exit code:")
			commonLogger.Println(result.ExitCode)
		}
	}
	return
}

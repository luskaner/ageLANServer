package executor

import (
	"crypto/x509"
	"io"
	"net"
	"runtime"

	commonCmd "github.com/luskaner/ageLANServer/common/cmd"
	"github.com/luskaner/ageLANServer/common/executables"
	commonExecutor "github.com/luskaner/ageLANServer/common/executor"
	"github.com/luskaner/ageLANServer/common/executor/exec"
	"github.com/luskaner/ageLANServer/launcher-common/cmd/config"
	"github.com/luskaner/ageLANServer/launcher-common/cmd/config/admin"
	"github.com/spf13/pflag"
)

func RunSetUp(gameId string, IP net.IP, macOsExclusiveMappings bool, certificate *x509.Certificate, logRoot string, out io.Writer, optionsFn func(options exec.Options)) (result *exec.Result) {
	values, flags := admin.SetupFlagSet()
	values.GameId = gameId
	values.MapIp = IP
	values.MacOsExclusiveMappings = macOsExclusiveMappings
	values.LogRoot = logRoot
	if certificate != nil {
		values.AddLocalCertData = certificate.Raw
	}
	options := exec.Options{File: executables.NativeFileName(true, executables.LauncherConfigAdmin), AsAdmin: true, Wait: true, ExitCode: true, Args: commonCmd.FlagSetToArgs(flags, true)}
	if optionsFn != nil {
		optionsFn(options)
	}
	if out != nil && (runtime.GOOS != "windows" || commonExecutor.IsAdmin() || !options.AsAdmin) {
		options.Stdout = out
		options.Stderr = out
	}
	result = options.Exec()
	return
}

func RunRevert(IPs bool, certificate bool, failfast bool, logRoot string, out io.Writer, optionsFn func(options exec.Options)) (result *exec.Result) {
	values, flags := admin.RevertFlagSet()
	values.LogRoot = logRoot
	if failfast {
		values.IPs = IPs
		values.Certs = certificate
	} else {
		values.RemoveAll = true
	}
	options := exec.Options{File: executables.NativeFileName(true, executables.LauncherConfigAdmin), AsAdmin: true, Wait: true, ExitCode: true, Args: commonCmd.FlagSetToArgs(flags, true)}
	if optionsFn != nil {
		optionsFn(options)
	}
	if out != nil && (runtime.GOOS != "windows" || commonExecutor.IsAdmin() || !options.AsAdmin) {
		options.Stdout = out
		options.Stderr = out
	}
	result = options.Exec()
	return
}

func runFlushCache(executableName string, wait bool, IPs bool, certificate bool, logRoot string, out io.Writer, optionsFn func(options exec.Options), values *config.FlushCacheValues, flags *pflag.FlagSet) (file string, result *exec.Result) {
	values.IPs = IPs
	values.Certs = certificate
	values.LogRoot = logRoot
	file = executables.NativeFileName(true, executableName)
	options := exec.Options{File: file, AsAdmin: true, Args: commonCmd.FlagSetToArgs(flags, wait)}
	if wait {
		options.Wait = true
		options.ExitCode = true
	} else {
		options.Pid = true
	}
	if optionsFn != nil {
		optionsFn(options)
	}
	if out != nil && (runtime.GOOS != "windows" || commonExecutor.IsAdmin() || !options.AsAdmin) {
		options.Stdout = out
		options.Stderr = out
	}
	result = options.Exec()
	return
}

func RunFlushCacheAgent(IPs bool, certificate bool, logRoot string, out io.Writer, optionsFn func(options exec.Options)) (file string, result *exec.Result) {
	values, singleFs := config.FlushCacheSingleFlagSet("", nil)
	return runFlushCache(executables.LauncherConfigAdminAgent, false, IPs, certificate, logRoot, out, optionsFn, values, singleFs.Fs())
}

func RunFlushCache(IPs bool, certificate bool, logRoot string, out io.Writer, optionsFn func(options exec.Options)) (file string, result *exec.Result) {
	values, flags := config.FlushCacheFlagSet()
	return runFlushCache(executables.LauncherConfigAdmin, true, IPs, certificate, logRoot, out, optionsFn, values, flags)
}

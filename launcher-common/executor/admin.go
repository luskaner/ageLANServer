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
)

func RunSetUp(gameId string, IP net.IP, certificate *x509.Certificate, logRoot string, out io.Writer, optionsFn func(options exec.Options)) (result *exec.Result) {
	values, flags := config.AdminSetupFlagSet()
	values.GameId = gameId
	values.MapIp = IP
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
	values, flags := config.AdminRevertFlagSet()
	values.LogRoot = logRoot
	if failfast {
		values.UnmapIPs = IPs
		values.RemoveLocalCert = certificate
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

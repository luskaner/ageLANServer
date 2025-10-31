package executor

import (
	"crypto/x509"
	"encoding/base64"
	"io"
	"net"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executor/exec"
)

func RunSetUp(gameId string, IP net.IP, certificate *x509.Certificate, CDN bool, logRoot string, out io.Writer, optionsFn func(options exec.Options)) (result *exec.Result) {
	args := make([]string, 0)
	args = append(args, "setup", "-e", gameId)
	if len(IP) > 0 {
		args = append(args, "-i")
		args = append(args, IP.String())
	}
	if certificate != nil {
		args = append(args, "-l")
		args = append(args, base64.StdEncoding.EncodeToString(certificate.Raw))
	}
	if CDN {
		args = append(args, "-c")
	}
	if logRoot != "" {
		args = append(args, "--logRoot", logRoot)
	}
	options := exec.Options{File: common.GetExeFileName(true, common.LauncherConfigAdmin), AsAdmin: true, Wait: true, ExitCode: true, Args: args}
	optionsFn(options)
	if out != nil {
		options.Stdout = out
		options.Stderr = out
	}
	result = options.Exec()
	return
}

func RunRevert(IPs bool, certificate bool, failfast bool, logRoot string, out io.Writer, optionsFn func(options exec.Options)) (result *exec.Result) {
	args := make([]string, 0)
	args = append(args, "revert")
	if failfast {
		if IPs {
			args = append(args, "-i")
		}
		if certificate {
			args = append(args, "-l")
		}
	} else {
		args = append(args, "-a")
	}
	if logRoot != "" {
		args = append(args, "--logRoot", logRoot)
	}
	options := exec.Options{File: common.GetExeFileName(true, common.LauncherConfigAdmin), AsAdmin: true, Wait: true, ExitCode: true, Args: args}
	optionsFn(options)
	if out != nil {
		options.Stdout = out
		options.Stderr = out
	}
	result = options.Exec()
	return
}

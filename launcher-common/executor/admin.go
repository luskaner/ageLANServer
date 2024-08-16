package executor

import (
	"crypto/x509"
	"encoding/base64"
	"github.com/luskaner/aoe2DELanServer/common"
	"net"
)

func RunSetUp(IPs []net.IP, certificate *x509.Certificate, CDN bool) (result *ExecResult) {
	args := make([]string, 0)
	args = append(args, "setup")
	if IPs != nil {
		for _, ip := range IPs {
			args = append(args, "-i")
			args = append(args, ip.String())
		}
	}
	if certificate != nil {
		args = append(args, "-l")
		args = append(args, base64.StdEncoding.EncodeToString(certificate.Raw))
	}
	if CDN {
		args = append(args, "-c")
	}
	result = ExecOptions{File: common.GetExeFileName(true, common.LauncherConfigAdmin), AsAdmin: true, Wait: true, ExitCode: true, Args: args}.Exec()
	return
}

func RunRevert(IPs bool, certificate bool, CDN bool, failfast bool) (result *ExecResult) {
	args := make([]string, 0)
	args = append(args, "revert")
	if failfast {
		if IPs {
			args = append(args, "-i")
		}
		if certificate {
			args = append(args, "-l")
		}
		if CDN {
			args = append(args, "-c")
		}
	} else {
		args = append(args, "-a")
	}
	result = ExecOptions{File: common.GetExeFileName(true, common.LauncherConfigAdmin), AsAdmin: true, Wait: true, ExitCode: true, Args: args}.Exec()
	return
}

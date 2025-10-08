package executor

import (
	"crypto/x509"
	"encoding/base64"
	"net"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executor/exec"
)

func RunSetUp(gameId string, IP net.IP, certificate *x509.Certificate, CDN bool) (result *exec.Result) {
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
	result = exec.Options{File: common.GetExeFileName(true, common.LauncherConfigAdmin), AsAdmin: true, Wait: true, ExitCode: true, Args: args}.Exec()
	return
}

func RunRevert(IPs bool, certificate bool, failfast bool) (result *exec.Result) {
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
	result = exec.Options{File: common.GetExeFileName(true, common.LauncherConfigAdmin), AsAdmin: true, Wait: true, ExitCode: true, Args: args}.Exec()
	return
}

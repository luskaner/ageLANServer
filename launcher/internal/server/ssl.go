package server

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/luskaner/aoe2DELanServer/common"
	"github.com/luskaner/aoe2DELanServer/launcher-common/executor/exec"
	"net"
	"os"
	"path/filepath"
)

func CheckConnectionFromServer(host string, insecureSkipVerify bool) bool {
	// 22 exit code means the host could be accessed and ssl certificate was tested (if specified)
	return HttpGet(fmt.Sprintf("https://%s", host), insecureSkipVerify) == 22
}

func ReadCertificateFromServer(host string) *x509.Certificate {
	conf := &tls.Config{
		InsecureSkipVerify: true,
	}
	conn, err := tls.Dial("tcp4", net.JoinHostPort(host, "443"), conf)
	if err != nil {
		return nil
	}
	defer func() {
		_ = conn.Close()
	}()
	certificates := conn.ConnectionState().PeerCertificates
	if len(certificates) > 0 {
		return certificates[0]
	}
	return nil
}

func GenerateCertificatePair(certificateFolder string) (result *exec.Result) {
	baseFolder := filepath.Join(certificateFolder, "..", "..")
	exePath := filepath.Join(baseFolder, common.GetExeFileName(false, common.ServerGenCert))
	if _, err := os.Stat(exePath); err != nil {
		return nil
	}
	result = exec.Options{File: exePath, Wait: true, ExitCode: true}.Exec()
	return
}

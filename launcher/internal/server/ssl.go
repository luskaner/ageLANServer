package server

import (
	"common"
	"crypto/tls"
	"crypto/x509"
	"launcher-common/executor"
	"net"
	"os"
	"path/filepath"
	"strings"
)

func connectToServer(host string, insecureSkipVerify bool) *tls.Conn {
	conf := &tls.Config{
		InsecureSkipVerify: insecureSkipVerify,
	}
	conn, err := tls.Dial("tcp", net.JoinHostPort(host, "443"), conf)
	if err != nil {
		return nil
	}
	return conn
}

func CheckConnectionFromServer(host string, insecureSkipVerify bool) bool {
	conn := connectToServer(host, insecureSkipVerify)
	if conn == nil {
		return false
	}
	defer func() {
		_ = conn.Close()
	}()
	return conn != nil
}

func ReadCertificateFromServer(host string) *x509.Certificate {
	conn := connectToServer(host, true)
	if conn == nil {
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

func GenerateCertificatePair(certificateFolder string) (result *executor.ExecResult) {
	baseFolder := filepath.Join(certificateFolder, "..", "..")
	batchPath := filepath.Join(baseFolder, common.GetScriptFileName(common.ServerGenCert))
	var path string
	if _, err := os.Stat(batchPath); err == nil {
		path = batchPath
	} else {
		exePath := filepath.Join(baseFolder, common.GetExeFileName(common.ServerGenCert))
		if _, err = os.Stat(exePath); err == nil {
			path = exePath
		} else {
			return nil
		}
	}
	result = executor.ExecOptions{File: path, Wait: true, SpecialFile: strings.HasSuffix(path, ".bat"), ExitCode: true}.Exec()
	return
}

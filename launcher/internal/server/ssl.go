package server

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/launcher-common/executor/exec"
	"net"
	"os"
	"path/filepath"
	"time"
)

func TlsConfig(insecureSkipVerify bool) *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: insecureSkipVerify,
	}
}

func connectToServer(host string, insecureSkipVerify bool) *tls.Conn {
	conn, err := tls.Dial("tcp4", net.JoinHostPort(host, "443"), TlsConfig(insecureSkipVerify))
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

func GenerateCertificatePair(certificateFolder string) (result *exec.Result) {
	baseFolder := filepath.Join(certificateFolder, "..", "..")
	exePath := filepath.Join(baseFolder, common.GetExeFileName(false, common.ServerGenCert))
	if _, err := os.Stat(exePath); err != nil {
		return nil
	}
	result = exec.Options{File: exePath, Wait: true, Args: []string{"-r"}, ExitCode: true}.Exec()
	return
}

func CertificateSoonExpired(cert string) bool {
	if cert == "" {
		return true
	}

	certPEM, err := os.ReadFile(cert)
	if err != nil {
		return true
	}

	block, _ := pem.Decode(certPEM)
	if block == nil {
		return true
	}

	crt, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return true
	}

	return time.Now().Add(24 * time.Hour).After(crt.NotAfter)
}

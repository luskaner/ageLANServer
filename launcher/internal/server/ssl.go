package server

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executor/exec"
)

func TlsConfig(serverName string, insecureSkipVerify bool) *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: insecureSkipVerify,
		ServerName:         serverName,
	}
}

func connectToServer(host string, insecureSkipVerify bool) *tls.Conn {
	ips := common.HostOrIpToIps(host)
	var ip string
	if len(ips) == 0 {
		ip = host
	} else {
		ip = ips[0]
	}
	conn, err := tls.Dial("tcp4", net.JoinHostPort(ip, "443"), TlsConfig(host, insecureSkipVerify))
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

func ReadCACertificateFromServer(host string, gameId string) *x509.Certificate {
	tr := &http.Transport{
		TLSClientConfig: TlsConfig(host, true),
	}
	ips := common.HostOrIpToIps(host)
	var ip string
	if len(ips) == 0 {
		ip = host
	} else {
		ip = ips[0]
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Get(fmt.Sprintf("https://%s/cacert.pem?gameId=%s", ip, gameId))
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
	block, _ := pem.Decode(bodyBytes)
	if block == nil || block.Type != "CERTIFICATE" {
		return nil
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil
	}
	return cert
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

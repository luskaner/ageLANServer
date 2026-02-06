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
	"github.com/luskaner/ageLANServer/common/executables"
	"github.com/luskaner/ageLANServer/common/executor/exec"
	commonLogger "github.com/luskaner/ageLANServer/common/logger"
)

func TlsConfig(serverName string, insecureSkipVerify bool, rootCAs *x509.CertPool) *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: insecureSkipVerify,
		ServerName:         serverName,
		RootCAs:            rootCAs,
	}
}

func connectToServer(host string, insecureSkipVerify bool, rootCAs *x509.CertPool) *tls.Conn {
	ips := common.HostOrIpToIps(host)
	var ip string
	if len(ips) == 0 {
		ip = host
	} else {
		ip = ips[0]
	}
	conn, err := tls.Dial("tcp4", net.JoinHostPort(ip, "443"), TlsConfig(host, insecureSkipVerify, rootCAs))
	if err != nil {
		return nil
	}
	return conn
}

func CheckConnectionFromServer(host string, insecureSkipVerify bool, rootCAs *x509.CertPool) bool {
	conn := connectToServer(host, insecureSkipVerify, rootCAs)
	if conn == nil {
		return false
	}
	defer func() {
		_ = conn.Close()
	}()
	return conn != nil
}

func ReadCACertificateFromServer(host string) *x509.Certificate {
	tr := &http.Transport{
		TLSClientConfig: TlsConfig(host, true, nil),
	}
	ips := common.HostOrIpToIps(host)
	var ip string
	if len(ips) == 0 {
		ip = host
	} else {
		ip = ips[0]
	}
	client := &http.Client{Transport: tr, Timeout: 1 * time.Second}
	//goland:noinspection ALL
	resp, err := client.Get(fmt.Sprintf("https://%s/cacert.pem", ip))
	if err != nil {
		commonLogger.Println("ReadCACertificateFromServer error:", err)
		return nil
	}
	if resp.StatusCode != http.StatusOK {
		commonLogger.Println("ReadCACertificateFromServer status code:", resp.StatusCode)
		return nil
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		commonLogger.Println("ReadCACertificateFromServer read error:", err)
		return nil
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
	block, _ := pem.Decode(bodyBytes)
	if block == nil || block.Type != "CERTIFICATE" {
		commonLogger.Println("ReadCACertificateFromServer: no certificate found")
		return nil
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		commonLogger.Println("ReadCACertificateFromServer parse error:", err)
		return nil
	}
	return cert
}

func GenerateCertificatePair(certificateFolder string, optionsFn func(options exec.Options)) (result *exec.Result) {
	baseFolder := filepath.Join(certificateFolder, "..", "..")
	exePath := filepath.Join(baseFolder, executables.Filename(false, executables.ServerGenCert))
	if _, err := os.Stat(exePath); err != nil {
		return nil
	}
	options := exec.Options{File: exePath, Wait: true, Args: []string{"-r"}, ExitCode: true}
	optionsFn(options)
	result = options.Exec()
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

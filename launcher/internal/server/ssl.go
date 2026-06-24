package server

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executables"
	"github.com/luskaner/ageLANServer/common/executor/exec"
	commonLogger "github.com/luskaner/ageLANServer/common/logger"
	"github.com/luskaner/ageLANServer/common/server"
)

func ReadCACertificateFromServer(host string) *x509.Certificate {
	tr := &http.Transport{
		TLSClientConfig: server.TlsConfig(host, true, nil),
	}
	ips := common.HostOrIpToIps(host)
	var ip string
	if len(ips) == 0 {
		ip = host
	} else {
		ip = ips[0]
	}
	client := &http.Client{Transport: tr, Timeout: 1 * time.Second}
	req, err := http.NewRequest("GET", fmt.Sprintf("https://%s/cacert.pem", ip), nil)
	if err != nil {
		commonLogger.Println("ReadCACertificateFromServer error:", err)
		return nil
	}
	req.Header.Set("User-Agent", common.UserAgent())
	req.Host = host
	//goland:noinspection ALL
	resp, err := client.Do(req)
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
	exePath := filepath.Join(baseFolder, executables.NativeFileName(false, executables.ServerGenCert))
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

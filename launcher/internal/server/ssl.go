package server

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"shared/executor"
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

func saveCertificateFromServer(host string) *os.File {
	conn := connectToServer(host, true)
	if conn == nil {
		return nil
	}
	defer func() {
		_ = conn.Close()
	}()
	certificates := conn.ConnectionState().PeerCertificates
	if len(certificates) > 0 {
		certOut, err := os.CreateTemp("", "cert_*.pem")
		if err != nil {
			return nil
		}
		defer func(certOut *os.File) {
			_ = certOut.Close()
		}(certOut)

		err = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certificates[0].Raw})
		if err != nil {
			return nil
		}
		return certOut
	}
	return nil
}

func readCertificateFromServer(host string) *x509.Certificate {
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

func TrustCertificateFromServer(host string) bool {
	certOut := saveCertificateFromServer(host)
	if certOut == nil {
		return false
	}
	defer func(certOut *os.File) {
		_ = os.Remove(certOut.Name())
	}(certOut)

	return executor.RunCustomExecutable("powershell", "-Command", fmt.Sprintf(`Import-Certificate -FilePath "%s" -CertStoreLocation Cert:\CurrentUser\Root`, certOut.Name()))
}

func UntrustCertificateFromServer(host string) bool {
	certOut := saveCertificateFromServer(host)
	if certOut == nil {
		return false
	}
	defer func(certOut *os.File) {
		_ = os.Remove(certOut.Name())
	}(certOut)
	return executor.RunCustomExecutable("powershell", "-Command", fmt.Sprintf(`$cert = Get-PfxCertificate "%s"; Get-ChildItem -Path Cert:\CurrentUser\Root | Where-Object { $_.Thumbprint -eq $cert.Thumbprint } | Remove-Item`, certOut.Name()))
}

func GenerateCertificatePair(certificateFolder string) bool {
	return executor.RunCustomExecutable(certificateFolder + `\..\..\genCert.exe`)
}

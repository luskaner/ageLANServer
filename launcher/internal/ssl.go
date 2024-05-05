package internal

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"time"
)

const Cert = "cert.pem"
const Key = "key.pem"

func GenerateCertificatePair(folder string) bool {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return false
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   Domain,
			Organization: []string{"github.com/luskaner/aoe2DELanServer"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:  x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
		},
		DNSNames: []string{Domain},
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return false
	}

	certFile, err := os.Create(folder + Cert)
	if err != nil {
		return false
	}

	defer func() {
		_ = certFile.Close()
	}()

	err = pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})

	if err != nil {
		return false
	}

	keyFile, err := os.Create(folder + Key)

	if err != nil {
		return false
	}

	defer func() {
		_ = keyFile.Close()
	}()

	err = pem.Encode(keyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
	if err != nil {
		return false
	}

	return true
}

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

func TrustCertificateFromServer(host string) bool {
	certOut := saveCertificateFromServer(host)
	if certOut == nil {
		return false
	}
	defer func(certOut *os.File) {
		_ = os.Remove(certOut.Name())
	}(certOut)
	return RunCustomExecutable("powershell", "-Command", fmt.Sprintf(`Import-Certificate -FilePath "%s" -CertStoreLocation Cert:\CurrentUser\Root`, certOut.Name()))
}

func UntrustCertificateFromServer(host string) bool {
	certOut := saveCertificateFromServer(host)
	if certOut == nil {
		return false
	}
	defer func(certOut *os.File) {
		_ = os.Remove(certOut.Name())
	}(certOut)
	return RunCustomExecutable("powershell", "-Command", fmt.Sprintf(`$cert = Get-PfxCertificate "%s"; Get-ChildItem -Path Cert:\CurrentUser\Root | Where-Object { $_.Thumbprint -eq $cert.Thumbprint } | Remove-Item`, certOut.Name()))
}

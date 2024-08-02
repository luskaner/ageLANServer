package internal

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"github.com/luskaner/aoe2DELanServer/common"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

func GenerateCertificatePair(folder string) bool {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return false
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   common.Domain,
			Organization: []string{common.CertSubjectOrganization},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:  x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
		},
		DNSNames: []string{common.Domain},
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return false
	}

	certFile, err := os.Create(filepath.Join(folder, common.Cert))
	if err != nil {
		return false
	}
	var keyFile *os.File
	delCertFile := false
	delKeyFile := false
	defer func() {
		_ = certFile.Close()
		if delCertFile {
			_ = os.Remove(filepath.Join(folder, common.Cert))
		}
		if keyFile != nil {
			_ = keyFile.Close()
			if delKeyFile {
				_ = os.Remove(filepath.Join(folder, common.Key))
			}
		}
	}()

	err = pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})

	if err != nil {
		delCertFile = true
		return false
	}

	keyFile, err = os.Create(filepath.Join(folder, common.Key))

	if err != nil {
		delCertFile = true
		return false
	}

	err = pem.Encode(keyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
	if err != nil {
		delCertFile = true
		delKeyFile = true
		return false
	}

	return true
}

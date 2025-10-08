package internal

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"github.com/luskaner/ageLANServer/common"
)

func generateSelfSignedCertificate(folder string) bool {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return false
	}
	template := getTemplate("selfsigned")
	certBytes, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return false
	}

	certFile, err := os.Create(filepath.Join(folder, common.SelfSignedCert))
	if err != nil {
		return false
	}
	var keyFile *os.File
	delCertFile := false
	delKeyFile := false
	defer func() {
		_ = certFile.Close()
		if delCertFile {
			_ = os.Remove(filepath.Join(folder, common.SelfSignedCert))
		}
		if keyFile != nil {
			_ = keyFile.Close()
			if delKeyFile {
				_ = os.Remove(filepath.Join(folder, common.SelfSignedKey))
			}
		}
	}()

	err = pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})

	if err != nil {
		delCertFile = true
		return false
	}

	keyFile, err = os.Create(filepath.Join(folder, common.SelfSignedKey))

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

func getTemplate(typ string) *x509.Certificate {
	template := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			CommonName:   common.Name,
			Organization: []string{common.CertSubjectOrganization},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		BasicConstraintsValid: true,
	}
	if typ != "normal" {
		template.IsCA = true
	} else {
		template.DNSNames = common.CertDomains()
	}
	if typ == "selfsigned" {
		template.Subject.CommonName += " Self-signed"
		template.DNSNames = common.SelfSignedCertDomains
		template.MaxPathLenZero = true
	} else if typ == "ca" {
		template.Subject.CommonName += " CA"
	}
	if typ != "ca" {
		template.KeyUsage = x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature
		template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
	}
	if typ == "selfsigned" {
		template.ExtKeyUsage = append(template.ExtKeyUsage, x509.ExtKeyUsageClientAuth)
	}
	return template
}

func generateCertificatePairs(folder string, certName string, keyName string, parent *x509.Certificate, parentKey *rsa.PrivateKey) (ok bool, caKey *bytes.Buffer) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return
	}
	var typ string
	if parent == nil && parentKey == nil {
		typ = "ca"
	} else {
		typ = "normal"
	}
	template := getTemplate(typ)
	if typ == "ca" {
		parent = template
		parentKey = key
	}
	certBytes, err := x509.CreateCertificate(rand.Reader, template, parent, &key.PublicKey, parentKey)
	if err != nil {
		return
	}
	certFile, err := os.Create(filepath.Join(folder, certName))
	if err != nil {
		return
	}
	var keyFile *os.File
	delCertFile := false
	delKeyFile := false
	defer func() {
		_ = certFile.Close()
		if delCertFile {
			_ = os.Remove(filepath.Join(folder, certName))
		}
		if keyFile != nil {
			_ = keyFile.Close()
			if delKeyFile {
				_ = os.Remove(filepath.Join(folder, keyName))
			}
		}
	}()

	err = pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})

	if err != nil {
		delCertFile = true
		return
	}

	var outputKey io.Writer
	if keyName != "" {
		keyFile, err = os.Create(filepath.Join(folder, keyName))
		if err != nil {
			delCertFile = true
			return
		}
		outputKey = keyFile
	} else {
		caKey = &bytes.Buffer{}
		outputKey = caKey
	}

	err = pem.Encode(outputKey, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	if err != nil {
		delCertFile = true
		delKeyFile = true
		return
	}

	ok = true
	return
}

func GenerateCertificatePairs(folder string) (ok bool) {
	var caKey *bytes.Buffer
	if ok, caKey = generateCertificatePairs(folder, common.CACert, "", nil, nil); !ok {
		return
	}
	ok = false
	defer func() {
		if !ok {
			_ = os.Remove(filepath.Join(folder, common.Key))
			_ = os.Remove(filepath.Join(folder, common.Cert))
		}
	}()

	certFile, err := os.ReadFile(filepath.Join(folder, common.CACert))
	if err != nil {
		return
	}

	fullCert, err := tls.X509KeyPair(certFile, caKey.Bytes())
	if err != nil {
		return
	}

	fullCert.Leaf, err = x509.ParseCertificate(fullCert.Certificate[0])
	if err != nil || fullCert.Leaf == nil {
		return
	}

	if ok, _ = generateCertificatePairs(
		folder, common.Cert, common.Key, fullCert.Leaf, fullCert.PrivateKey.(*rsa.PrivateKey)); !ok {
		_ = os.Remove(filepath.Join(folder, common.CACert))
		return
	}

	if ok = generateSelfSignedCertificate(folder); !ok {
		_ = os.Remove(filepath.Join(folder, common.CACert))
		return
	}

	ok = true
	return
}

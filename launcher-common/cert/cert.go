package cert

import (
	"crypto/x509"
	"encoding/pem"
	"os"
)

func BytesToCertificate(data []byte) *x509.Certificate {
	cert, _ := x509.ParseCertificate(data)
	return cert
}

func WriteAsPem(data []byte, file *os.File) error {
	pemData := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: data,
	})
	_, err := file.Write(pemData)
	return err
}

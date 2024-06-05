package internal

import (
	"crypto/x509"
	"encoding/base64"
)

func Base64ToCertificate(data string) *x509.Certificate {
	decodedBytes, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil
	}
	cert, err := x509.ParseCertificate(decodedBytes)
	if err != nil {
		return nil
	}
	return cert
}

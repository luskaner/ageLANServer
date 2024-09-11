package cert

import "crypto/x509"

func BytesToCertificate(data []byte) *x509.Certificate {
	cert, _ := x509.ParseCertificate(data)
	return cert
}

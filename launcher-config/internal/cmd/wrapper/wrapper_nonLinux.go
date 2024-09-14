//go:build !linux

package wrapper

import (
	"crypto/x509"
	"github.com/luskaner/aoe2DELanServer/launcher-common/cert"
)

func RemoveUserCert() (crt *x509.Certificate, err error) {
	return cert.UntrustCertificate(true)
}

func AddUserCert(crt *x509.Certificate) error {
	return cert.TrustCertificate(true, crt)
}

func BytesToCertificate(data []byte) *x509.Certificate {
	return cert.BytesToCertificate(data)
}

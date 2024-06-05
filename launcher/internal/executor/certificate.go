package executor

import (
	"crypto/x509"
	"shared"
)

func AddCertificateInternal(elevate bool, cert *x509.Certificate) bool {
	if elevate {
		return AddCertificate(elevate, *cert)
	}
	return shared.TrustCertificate(cert)
}

func RemoveCertificateInternal(elevate bool) bool {
	if elevate {
		return RemoveCertificate(elevate)
	}
	return shared.UntrustCertificate()
}

package wrapper

import "crypto/x509"

func RemoveUserCert() (crt *x509.Certificate, err error) {
	// Must not be called
	return nil, nil
}

func AddUserCert(_ any) error {
	// Must not be called
	return nil
}

func BytesToCertificate(_ any) *x509.Certificate {
	// Must not be called
	return nil
}

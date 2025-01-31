package wrapper

import "crypto/x509"

func RemoveUserCerts() (crts []*x509.Certificate, err error) {
	// Must not be called
	return nil, nil
}

func AddUserCerts(_ any) error {
	// Must not be called
	return nil
}

func BytesToCertificate(_ any) *x509.Certificate {
	// Must not be called
	return nil
}

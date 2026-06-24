package certStore

import "crypto/x509"

// ReloadSystemCertificates No need to reload certificates as the validity is checked by the OS
func ReloadSystemCertificates() {}

func CertPool() *x509.CertPool { return nil }

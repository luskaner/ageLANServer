//go:build !windows

package certStore

import (
	"crypto/x509"
	_ "unsafe"
)

// TODO: Check on every minor version release if there is a better way to do it, or, at least, it is compatible

//go:linkname systemRoots crypto/x509.systemRoots
var systemRoots *x509.CertPool

func ReloadSystemCertificates() {
	systemRoots, _ = loadSystemRoots()
}

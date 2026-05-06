//go:build windows || darwin

package certStore

// ReloadSystemCertificates No need to reload certificates as the validity is checked by the OS
func ReloadSystemCertificates() {

}

package common

import (
	"os"
	"path/filepath"

	"github.com/luskaner/ageLANServer/common/paths"
)

func CertificatePairFolder(executablePath string) string {
	if executablePath == "" {
		return ""
	}
	parentDir := filepath.Dir(executablePath)
	if parentDir == "" {
		return ""
	}
	folder := filepath.Join(parentDir, paths.ResourcesDir, "certificates")
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		if os.Mkdir(folder, 0755) != nil {
			return ""
		}
	}
	return folder
}

func CertificatePairs(executablePath string) (ok bool, parentDir string, cert string, key string, caCert string, selfSignedCert string, selfSignedKey string) {
	parentDir = CertificatePairFolder(executablePath)
	if parentDir == "" {
		return
	}
	cert = filepath.Join(parentDir, Cert)
	if _, err := os.Stat(cert); os.IsNotExist(err) {
		return
	}
	key = filepath.Join(parentDir, Key)
	if _, err := os.Stat(filepath.Join(parentDir, Key)); os.IsNotExist(err) {
		return
	}
	caCert = filepath.Join(parentDir, CACert)
	if _, err := os.Stat(caCert); os.IsNotExist(err) {
		return
	}
	selfSignedCert = filepath.Join(parentDir, SelfSignedCert)
	if _, err := os.Stat(selfSignedCert); os.IsNotExist(err) {
		return
	}
	selfSignedKey = filepath.Join(parentDir, SelfSignedKey)
	if _, err := os.Stat(selfSignedKey); os.IsNotExist(err) {
		return
	}
	ok = true
	return
}

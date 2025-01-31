package common

import (
	"os"
	"path/filepath"
)

func CertificatePairFolder(executablePath string) string {
	if executablePath == "" {
		return ""
	}
	parentDir := filepath.Dir(executablePath)
	if parentDir == "" {
		return ""
	}
	folder := filepath.Join(parentDir, "resources", "certificates")
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		if os.Mkdir(folder, 0755) != nil {
			return ""
		}
	}
	return folder
}

func CertificatePair(executablePath string) (ok bool, parentDir string, cert string) {
	parentDir = CertificatePairFolder(executablePath)
	if parentDir == "" {
		return
	}
	cert = filepath.Join(parentDir, Cert)
	if _, err := os.Stat(cert); os.IsNotExist(err) {
		return
	}
	if _, err := os.Stat(filepath.Join(parentDir, Key)); os.IsNotExist(err) {
		return
	}
	ok = true
	return
}

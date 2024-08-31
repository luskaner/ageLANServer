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

func HasCertificatePair(executablePath string) bool {
	parentDir := CertificatePairFolder(executablePath)
	if parentDir == "" {
		return false
	}
	if _, err := os.Stat(filepath.Join(parentDir, Cert)); os.IsNotExist(err) {
		return false
	}
	if _, err := os.Stat(filepath.Join(parentDir, Key)); os.IsNotExist(err) {
		return false
	}
	return true
}

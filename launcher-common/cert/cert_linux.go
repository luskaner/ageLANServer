//go:build linux

package cert

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/hairyhenderson/go-which"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/launcher-common/executor/exec"
	"io"
	"os"
	"path"
	"path/filepath"
)

var updateStoreBinaries = []string{
	// Debian, OpenSUSE
	"update-ca-certificates",
	// Fedora, Arch Linux
	"update-ca-trust",
}

var certStorePaths = []string{
	// Arch
	"/etc/ca-certificates/trust-source/anchors",
	// Debian
	"/usr/local/share/ca-certificates",
	// Fedora
	"/etc/pki/ca-trust/source/anchors",
	// OpenSUSE
	"/etc/pki/trust/anchors",
}

func updateStore() error {
	binary := which.Which(updateStoreBinaries...)
	if binary == "" {
		return fmt.Errorf("update store binary not found")
	}
	return exec.Options{
		File:        binary,
		SpecialFile: true,
		AsAdmin:     true,
		Wait:        true,
		ExitCode:    true,
	}.Exec().Err
}

func getCertPath() (err error, certPath string) {
	var stat os.FileInfo
	var foundPath string
	for _, dir := range certStorePaths {
		if stat, err = os.Stat(dir); err == nil && stat.IsDir() {
			foundPath = dir
			break
		}
	}
	if foundPath == "" {
		err = fmt.Errorf("cert store not found")
		return
	}
	certPath = path.Join(foundPath, fmt.Sprintf("%s.crt", common.Name))
	return
}

func TrustCertificates(_ bool, certs []*x509.Certificate) error {
	err, certPath := getCertPath()
	if err != nil {
		return err
	}

	for _, cert := range certs {
		var certFile *os.File

		certFile, err = os.CreateTemp("", "*")
		if err != nil {
			return err
		}

		err = WriteAsPem(cert.Raw, certFile)

		if err != nil {
			return err
		}

		err = certFile.Close()
		if err != nil {
			return err
		}

		err = os.Rename(certFile.Name(), certPath)
		if err != nil {
			// Likely certFile does not share filesystem with certPath
			var certFileTmp *os.File
			certFileTmp, err = os.CreateTemp(filepath.Dir(certPath), ".*")
			if err != nil {
				_ = os.Remove(certFile.Name())
				return err
			}

			var newCertFile *os.File
			newCertFile, err = os.Open(certFile.Name())
			if err != nil {
				_ = os.Remove(certFile.Name())
				_ = os.Remove(certFileTmp.Name())
				return err
			}

			_, err = io.Copy(certFileTmp, newCertFile)
			_ = newCertFile.Close()
			_ = os.Remove(newCertFile.Name())
			if err != nil {
				_ = os.Remove(certFileTmp.Name())
				return err
			}

			err = certFileTmp.Close()
			if err != nil {
				_ = os.Remove(certFileTmp.Name())
				return err
			}

			err = os.Rename(certFileTmp.Name(), certPath)

			if err != nil {
				_ = os.Remove(certFileTmp.Name())
				return err
			}
		}

		err = os.Chmod(certPath, 0644)
		if err != nil {
			return err
		}
	}

	return updateStore()
}

func UntrustCertificates(_ bool) (certs []*x509.Certificate, err error) {
	var certPath string
	err, certPath = getCertPath()
	if err != nil {
		return
	}

	if _, err = os.Stat(certPath); os.IsNotExist(err) {
		return
	}

	var certFile *os.File
	certFile, err = os.Open(certPath)

	if err != nil {
		return
	}

	var certBytes []byte
	certBytes, err = io.ReadAll(certFile)

	if err != nil {
		return
	}

	block, _ := pem.Decode(certBytes)
	var cert *x509.Certificate
	cert, err = x509.ParseCertificate(block.Bytes)

	if err != nil {
		return
	}

	err = os.Remove(certFile.Name())
	if err != nil {
		return
	}

	err = updateStore()
	if err != nil {
		certs = []*x509.Certificate{cert}
	}

	return
}

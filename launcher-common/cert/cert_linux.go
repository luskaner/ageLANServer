//go:build linux

package cert

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/hairyhenderson/go-which"
	"github.com/luskaner/aoe2DELanServer/common"
	"github.com/luskaner/aoe2DELanServer/launcher-common/executor/exec"
	"io"
	"os"
	"path"
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
	certPath = path.Join(foundPath, fmt.Sprintf("%s.crt", common.Domain))
	return
}

func TrustCertificate(_ bool, cert *x509.Certificate) error {
	err, certPath := getCertPath()
	if err != nil {
		return err
	}

	var certFile *os.File

	certFile, err = os.CreateTemp("", "*")
	if err != nil {
		return err
	}

	pemData := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	})

	_, err = certFile.Write(pemData)
	if err != nil {
		return err
	}

	err = certFile.Close()
	if err != nil {
		return err
	}

	err = os.Rename(certFile.Name(), certPath)
	if err != nil {
		return err
	}

	err = os.Chmod(certPath, 0644)
	if err != nil {
		return err
	}

	return updateStore()
}

func UntrustCertificate(_ bool) (cert *x509.Certificate, err error) {
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
	cert, err = x509.ParseCertificate(block.Bytes)

	if err != nil {
		return
	}

	err = os.Remove(certFile.Name())
	if err != nil {
		return
	}

	err = updateStore()
	return
}

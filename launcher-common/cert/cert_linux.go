//go:build linux

package cert

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/luskaner/aoe2DELanServer/common"
	"github.com/luskaner/aoe2DELanServer/launcher-common"
	"github.com/luskaner/aoe2DELanServer/launcher-common/executor/exec"
	"io"
	"os"
)

func updateStore() error {
	options := exec.Options{
		SpecialFile: true,
		AsAdmin:     true,
		Wait:        true,
		ExitCode:    true,
	}
	switch {
	case launcher_common.Ubuntu():
		options.File = "update-ca-certificates"
	case launcher_common.SteamOS():
		options.File = "trust extract-compat"
	}
	return options.Exec().Err
}

func getCertPath() (err error, certPath string) {
	var storePath string
	switch {
	case launcher_common.Ubuntu():
		storePath = "/usr/local/share/ca-certificates"
	case launcher_common.SteamOS():
		storePath = "/etc/ca-certificates/trust-source/anchors"
	}

	if storePath == "" {
		err = fmt.Errorf("unsupported OS")
		return
	}

	if _, err = os.Stat(storePath); os.IsNotExist(err) {
		err = fmt.Errorf("store path does not exist")
		return
	}

	certPath = fmt.Sprintf("%s/%s.crt", storePath, common.Domain)
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

func SupportsUserStore() bool {
	return false
}

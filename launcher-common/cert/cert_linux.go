//go:build linux

package cert

import (
	"crypto/x509"
	"fmt"
	"github.com/luskaner/aoe2DELanServer/launcher-common"
	"github.com/luskaner/aoe2DELanServer/launcher-common/executor/exec"
	"os"
)

func updateStore() error {
	options := exec.Options{
		AsAdmin:  true,
		Wait:     true,
		ExitCode: true,
	}
	switch {
	case launcher_common.Ubuntu():
		options.File = "update-ca-certificates"
	case launcher_common.SteamOS():
		options.File = "trust extract-compat"
	}
	return options.Exec().Err
}

func getCertPath(cert *x509.Certificate) (err error, certPath string) {
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

	certPath = fmt.Sprintf("%s/%s.crt", storePath, cert.Subject.CommonName)
	return
}

func TrustCertificate(_ bool, cert *x509.Certificate) error {
	err, certPath := getCertPath(cert)
	if err != nil {
		return err
	}

	var certFile *os.File

	certFile, err = os.CreateTemp("", "*")
	if err != nil {
		return err
	}

	_, err = certFile.Write(cert.Raw)
	if err != nil {
		return err
	}

	err = certFile.Close()
	if err != nil {
		return err
	}

	result := exec.Options{
		AsAdmin:  true,
		Wait:     true,
		ExitCode: true,
		Shell:    true,
		File:     "mv",
		Args:     []string{"-f", certFile.Name(), certPath},
	}.Exec()

	if !result.Success() {
		_ = os.Remove(certFile.Name())
		return result.Err
	}

	err = updateStore()
	if err != nil {
		return err
	}

	return nil
}

func UntrustCertificate(_ bool) (cert *x509.Certificate, err error) {
	var certPath string
	err, certPath = getCertPath(cert)
	if err != nil {
		return
	}

	if _, err = os.Stat(certPath); os.IsNotExist(err) {
		return
	}

	var certFile *os.File
	certFile, err = os.Open(certFile.Name())

	if err != nil {
		return
	}

	defer func() {
		_ = os.Remove(certFile.Name())
	}()

	var certBytes []byte
	_, err = certFile.Read(certBytes)

	if err != nil {
		return
	}

	cert, err = x509.ParseCertificate(certBytes)

	if err != nil {
		return
	}

	err = exec.Options{
		AsAdmin:  true,
		Wait:     true,
		ExitCode: true,
		Shell:    true,
		File:     "rm",
		Args:     []string{certPath},
	}.Exec().Err

	if err != nil {
		return
	}

	err = updateStore()
	if err != nil {
		return
	}

	return
}

func SupportsUserStore() bool {
	return false
}

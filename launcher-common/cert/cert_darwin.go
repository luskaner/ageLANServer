package cert

import (
	"crypto/x509"
	"fmt"
	"github.com/luskaner/aoe2DELanServer/common"
	"github.com/luskaner/aoe2DELanServer/launcher-common/executor/exec"
	"io"
	"os"
)

func getKeychain(userStore bool) (path string, err error) {
	if userStore {
		path = fmt.Sprintf("%s/Library/Keychains/login.keychain-db", os.Getenv("HOME"))
	} else {
		path = "/Library/Keychains/System.keychain"
	}
	return
}

func TrustCertificate(userStore bool, cert *x509.Certificate) error {
	keychain, err := getKeychain(userStore)
	if err != nil {
		return err
	}
	var certFile *os.File

	certFile, err = os.CreateTemp("", "*")
	if err != nil {
		return err
	}

	defer func() {
		_ = os.Remove(certFile.Name())
	}()

	_, err = certFile.Write(cert.Raw)
	if err != nil {
		return err
	}

	err = certFile.Close()
	if err != nil {
		return err
	}

	return exec.Options{
		AsAdmin:  !userStore,
		Wait:     true,
		ExitCode: true,
		File:     "security",
		Args:     []string{"add-trusted-cert", "-d", "-r", "trustRoot", "-k", keychain, certFile.Name()},
	}.Exec().Err
}

func UntrustCertificate(userStore bool) (cert *x509.Certificate, err error) {
	var keychain string
	keychain, err = getKeychain(userStore)
	if err != nil {
		return
	}
	var certFile *os.File

	certFile, err = os.CreateTemp("", "*")
	if err != nil {
		return
	}

	err = certFile.Close()
	if err != nil {
		return
	}

	result := exec.Options{
		AsAdmin:  !userStore,
		Wait:     true,
		ExitCode: true,
		File:     "security",
		Args:     []string{"export", "-k", keychain, "-t", "certs", "-f", "x509", "-o", certFile.Name(), "-c", common.Domain},
	}.Exec()

	if !result.Success() {
		_ = os.Remove(certFile.Name())
		err = result.Err
	}

	if !userStore && !exec.IsAdmin() {
		result = exec.Options{
			SpecialFile: true,
			AsAdmin:     true,
			Wait:        true,
			ExitCode:    true,
			Shell:       true,
			File:        "chmod",
			Args:        []string{"666", certFile.Name()},
		}.Exec()

		if !result.Success() {
			err = result.Err
			return
		}
	}
	certFile, err = os.Open(certFile.Name())

	if err != nil {
		return
	}

	defer func() {
		_ = os.Remove(certFile.Name())
	}()

	var certBytes []byte
	certBytes, err = io.ReadAll(certFile)

	if err != nil {
		return
	}

	cert, err = x509.ParseCertificate(certBytes)

	if err != nil {
		return
	}

	result = exec.Options{
		AsAdmin:  !userStore,
		Wait:     true,
		ExitCode: true,
		File:     "security",
		Args:     []string{"delete-certificate", "-c", common.Domain, keychain},
	}.Exec()

	if !result.Success() {
		err = result.Err
	}

	return
}

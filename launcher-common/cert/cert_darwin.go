package cert

import (
	"bytes"
	"crypto/sha1"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executor/exec"
)

// TODO: Test in mac

func TrustCertificates(userStore bool, certs []*x509.Certificate) error {
	keychain, asAdmin, err := keychainPath(userStore)
	if err != nil {
		return err
	}
	for _, cert := range certs {
		certFile, err := os.CreateTemp("", common.Name+"_cert_*.pem")
		if err != nil {
			return err
		}
		certPath := certFile.Name()
		if err = WriteAsPem(cert.Raw, certFile); err != nil {
			_ = certFile.Close()
			_ = os.Remove(certPath)
			return err
		}
		if err = certFile.Close(); err != nil {
			_ = os.Remove(certPath)
			return err
		}
		args := []string{"add-trusted-cert"}
		if !userStore {
			args = append(args, "-d")
		}
		args = append(args, "-k", keychain, certPath)
		_, err = runCommand(asAdmin, args...)
		_ = os.Remove(certPath)
		if err != nil {
			return err
		}
	}
	return nil
}

// TODO: Test in mac

func UntrustCertificates(userStore bool) (certs []*x509.Certificate, err error) {
	var existing []*x509.Certificate
	existing, err = EnumCertificates(userStore)
	if err != nil {
		return
	}
	if len(existing) == 0 {
		return
	}
	keychain, asAdmin, err := keychainPath(userStore)
	if err != nil {
		return nil, err
	}
	certs = make([]*x509.Certificate, 0, len(existing))
	for _, cert := range existing {
		match := false
		for _, org := range cert.Subject.Organization {
			if org == common.CertSubjectOrganization {
				match = true
				break
			}
		}
		if !match {
			continue
		}
		fingerprint := sha1.Sum(cert.Raw)
		fingerprintHex := strings.ToUpper(hex.EncodeToString(fingerprint[:]))
		_, err = runCommand(asAdmin, "delete-certificate", "-Z", fingerprintHex, keychain)
		if err != nil {
			return
		}
		certs = append(certs, cert)
	}
	return
}

// TODO: Test in mac

func EnumCertificates(userStore bool) (certs []*x509.Certificate, err error) {
	var keychain string
	keychain, _, err = keychainPath(userStore)
	if err != nil {
		return
	}
	var output string
	output, err = runCommand(false, "find-certificate", "-a", "-p", keychain)
	if err != nil {
		return
	}
	data := []byte(output)
	_, _, certs, err = ReadFromData(data)
	return
}

func keychainPath(userStore bool) (path string, asAdmin bool, err error) {
	if userStore {
		path, err = defaultUserKeychain()
		return path, false, err
	}
	return "/Library/Keychains/System.keychain", true, nil
}

func defaultUserKeychain() (value string, err error) {
	var output string
	output, err = runCommand(false, "default-keychain")
	if err != nil {
		return
	}
	path := strings.TrimSpace(output)
	value = strings.Trim(path, `"`)
	return
}

func runCommand(asAdmin bool, args ...string) (string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	result := exec.Options{
		File:        "security",
		SpecialFile: true,
		Args:        args,
		AsAdmin:     asAdmin,
		Wait:        true,
		ExitCode:    true,
		Stdout:      &stdout,
		Stderr:      &stderr,
	}.Exec()
	if result.Err != nil {
		return "", result.Err
	}
	if result.ExitCode != 0 {
		errText := stderr.String()
		if errText == "" {
			errText = stdout.String()
		}
		errText = strings.TrimSpace(errText)
		if errText == "" {
			errText = fmt.Sprintf("exit code %d", result.ExitCode)
		}
		return "", fmt.Errorf("command failed: %s", errText)
	}
	return stdout.String(), nil
}

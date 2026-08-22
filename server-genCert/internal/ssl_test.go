package internal

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/luskaner/ageLANServer/common"
)

func generateIntoTemp(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if !GenerateCertificatePairs(dir) {
		t.Fatal("GenerateCertificatePairs reported failure")
	}
	return dir
}

func readCertificate(t *testing.T, dir string, name string) *x509.Certificate {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		t.Fatalf("%s: no PEM block", name)
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("%s: parse: %v", name, err)
	}
	return cert
}

// Regression: certificates used to be valid starting from exactly time.Now(),
// so any client with a slightly ahead-set clock rejected them as
// "not yet valid" right after generation.
func TestGeneratedCertificatesAreBackdated(t *testing.T) {
	dir := generateIntoTemp(t)
	skewLimit := time.Now().Add(-30 * time.Second)

	for _, name := range []string{common.CACert, common.Cert, common.SelfSignedCert} {
		cert := readCertificate(t, dir, name)
		if cert.NotBefore.After(skewLimit) {
			t.Errorf("%s NotBefore = %v, want backdated at least 30s", name, cert.NotBefore)
		}
		if !cert.NotAfter.After(time.Now()) {
			t.Errorf("%s already expired", name)
		}
	}
}

func TestGeneratedChainVerifiesAgainstCA(t *testing.T) {
	dir := generateIntoTemp(t)
	ca := readCertificate(t, dir, common.CACert)
	leaf := readCertificate(t, dir, common.Cert)

	roots := x509.NewCertPool()
	roots.AddCert(ca)
	if _, err := leaf.Verify(x509.VerifyOptions{
		Roots:       roots,
		CurrentTime: time.Now(),
		KeyUsages:   []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}); err != nil {
		t.Fatalf("leaf does not verify against CA: %v", err)
	}
}

// Regression: private keys were created world-readable (0666 & umask).
func TestGeneratedPrivateKeysAreOwnerOnly(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission bits do not apply on Windows")
	}
	dir := generateIntoTemp(t)
	for _, name := range []string{common.Key, common.SelfSignedKey} {
		fi, err := os.Stat(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("stat %s: %v", name, err)
		}
		if perm := fi.Mode().Perm(); perm&0o077 != 0 {
			t.Errorf("%s permissions = %#o, group/other bits must be unset", name, perm)
		}
	}
}

func TestGeneratedKeyPairsLoadAsTLSKeyPair(t *testing.T) {
	dir := generateIntoTemp(t)
	for _, pair := range [][2]string{
		{common.Cert, common.Key},
		{common.SelfSignedCert, common.SelfSignedKey},
	} {
		certPEM, err := os.ReadFile(filepath.Join(dir, pair[0]))
		if err != nil {
			t.Fatal(err)
		}
		keyPEM, err := os.ReadFile(filepath.Join(dir, pair[1]))
		if err != nil {
			t.Fatal(err)
		}
		if _, err = tls.X509KeyPair(certPEM, keyPEM); err != nil {
			t.Errorf("%s/%s: %v", pair[0], pair[1], err)
		}
	}
}

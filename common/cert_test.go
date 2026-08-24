package common

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/luskaner/ageLANServer/common/paths"
)

func TestBytesToCertificateInvalidReturnsNil(t *testing.T) {
	if cert := BytesToCertificate(nil); cert != nil {
		t.Fatal("nil input must yield nil certificate")
	}
	if cert := BytesToCertificate([]byte("garbage")); cert != nil {
		t.Fatal("garbage input must yield nil certificate")
	}
}

func generateTestCertificate(t *testing.T, name string) ([]byte, *x509.Certificate) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("key generation: %v", err)
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: name,
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{name},
	}
	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("certificate creation: %v", err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatalf("certificate parsing: %v", err)
	}
	return der, cert
}

func TestWriteAsPemAndReadFromFileRoundTrip(t *testing.T) {
	der1, cert1 := generateTestCertificate(t, "a.test")
	der2, cert2 := generateTestCertificate(t, "b.test")

	f, err := os.CreateTemp(t.TempDir(), "certs")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	for _, der := range [][]byte{der1, der2} {
		if err = WriteAsPem(der, f); err != nil {
			t.Fatalf("WriteAsPem: %v", err)
		}
	}

	keys, keyToIndex, values, err := ReadFromFile(f.Name())
	if err != nil {
		t.Fatalf("ReadFromFile: %v", err)
	}
	if len(values) != 2 || len(keys) != 2 {
		t.Fatalf("keys=%d values=%d, want 2/2", len(keys), len(values))
	}
	if !values[0].Equal(cert1) || !values[1].Equal(cert2) {
		t.Fatal("certificates do not round-trip in order")
	}
	for i, key := range keys {
		if key == "" {
			t.Fatalf("key %d is empty", i)
		}
		if idx, ok := keyToIndex[key]; !ok || idx != i {
			t.Fatalf("keyToIndex[%q] = %d, %v; want %d, true", key, idx, ok, i)
		}
	}
}

func TestReadFromDataEmpty(t *testing.T) {
	keys, keyToIndex, values, err := ReadFromData(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys) != 0 || len(values) != 0 || len(keyToIndex) != 0 {
		t.Fatalf("expected empty results, got %d/%d", len(keys), len(values))
	}
}

func TestKoanfFileLoadErrorFormatting(t *testing.T) {
	err := &KoanfFileLoadError{Path: "/some/file.toml"}
	if err.Error() != "/some/file.toml: <nil>" {
		t.Fatalf("Error() = %q", err.Error())
	}
}

func TestReadCertsPool(t *testing.T) {
	der, _ := generateTestCertificate(t, "pool.test")
	f, err := os.CreateTemp(t.TempDir(), "pool")
	if err != nil {
		t.Fatal(err)
	}
	if err = WriteAsPem(der, f); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	pool, err := ReadCertsPool(f.Name())
	if err != nil {
		t.Fatalf("ReadCertsPool: %v", err)
	}
	if pool == nil {
		t.Fatal("pool should not be nil")
	}
}

func TestReadCertsPool_InvalidFile(t *testing.T) {
	_, err := ReadCertsPool("/nonexistent/path.pem")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestCertificatePairs_AllPresent(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{Cert, Key, CACert, SelfSignedCert, SelfSignedKey} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("data"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	ok, cert, key, ca, ssCert, ssKey := CertificatePairs(dir)
	if !ok {
		t.Fatal("CertificatePairs should return ok=true when all files present")
	}
	if cert != filepath.Join(dir, Cert) {
		t.Errorf("cert = %q", cert)
	}
	if key != filepath.Join(dir, Key) {
		t.Errorf("key = %q", key)
	}
	if ca != filepath.Join(dir, CACert) {
		t.Errorf("caCert = %q", ca)
	}
	if ssCert != filepath.Join(dir, SelfSignedCert) {
		t.Errorf("selfSignedCert = %q", ssCert)
	}
	if ssKey != filepath.Join(dir, SelfSignedKey) {
		t.Errorf("selfSignedKey = %q", ssKey)
	}
}

func TestCertificatePairs_EmptyDir(t *testing.T) {
	ok, _, _, _, _, _ := CertificatePairs("")
	if ok {
		t.Error("CertificatePairs should return ok=false for empty dir")
	}
}

func TestCertificatePairs_MissingCert(t *testing.T) {
	dir := t.TempDir()
	// Only create key, ca, ssCert, ssKey — skip cert.pem
	for _, name := range []string{Key, CACert, SelfSignedCert, SelfSignedKey} {
		os.WriteFile(filepath.Join(dir, name), []byte("data"), 0644)
	}
	ok, _, _, _, _, _ := CertificatePairs(dir)
	if ok {
		t.Error("CertificatePairs should return ok=false when cert.pem is missing")
	}
}

func TestCertificatePairs_MissingKey(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{Cert, CACert, SelfSignedCert, SelfSignedKey} {
		os.WriteFile(filepath.Join(dir, name), []byte("data"), 0644)
	}
	ok, _, _, _, _, _ := CertificatePairs(dir)
	if ok {
		t.Error("CertificatePairs should return ok=false when key.pem is missing")
	}
}

func TestCertificatePairFolderPath_Empty(t *testing.T) {
	if got := certificatePairFolderPath(""); got != "" {
		t.Errorf("certificatePairFolderPath(\"\") = %q, want empty", got)
	}
}

func TestCertificatePairFolder(t *testing.T) {
	dir := t.TempDir()
	exe := filepath.Join(dir, "server.exe")
	folder := CertificatePairFolder(exe)
	expected := filepath.Join(dir, paths.ResourcesDir, "certificates")
	if folder != expected {
		t.Errorf("CertificatePairFolder = %q, want %q", folder, expected)
	}
	// Verify directory was created
	if _, err := os.Stat(expected); os.IsNotExist(err) {
		t.Error("CertificatePairFolder should create the directory")
	}
}

func TestCertificatePairFolder_AlreadyExists(t *testing.T) {
	dir := t.TempDir()
	exe := filepath.Join(dir, "server.exe")
	CertificatePairFolder(exe) // create
	folder := CertificatePairFolder(exe) // should not fail
	if folder == "" {
		t.Error("CertificatePairFolder should succeed when directory already exists")
	}
}

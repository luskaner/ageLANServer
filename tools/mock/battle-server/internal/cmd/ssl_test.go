package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRootCmdInvalidSSLKey(t *testing.T) {
	oldBsPort := bsPort
	oldCert := sslCert
	oldKey := sslKey
	defer func() { bsPort = oldBsPort; sslCert = oldCert; sslKey = oldKey }()
	bsPort = 8080
	sslCert = "/tmp/cert.pem"
	sslKey = "/nonexistent/key.pem"
	if err := rootCmd(); err == nil {
		t.Fatal("should fail with invalid key")
	}
}

func TestRootCmdInvalidSSLCert(t *testing.T) {
	oldBsPort := bsPort
	oldCert := sslCert
	oldKey := sslKey
	defer func() { bsPort = oldBsPort; sslCert = oldCert; sslKey = oldKey }()
	bsPort = 8080
	sslCert = "/nonexistent/cert.pem"
	sslKey = "/tmp/key.pem"
	// Create temp key file so key exists but cert doesn't
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "key.pem")
	os.WriteFile(keyPath, []byte("key"), 0644)
	sslKey = keyPath
	if err := rootCmd(); err == nil {
		t.Fatal("should fail with invalid cert")
	}
}

func TestRootCmdBothSSLExist(t *testing.T) {
	oldBsPort := bsPort
	oldCert := sslCert
	oldKey := sslKey
	oldName := name
	defer func() { bsPort = oldBsPort; sslCert = oldCert; sslKey = oldKey; name = oldName }()
	bsPort = 8080
	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "cert.pem")
	keyPath := filepath.Join(tmpDir, "key.pem")
	os.WriteFile(certPath, []byte("cert"), 0644)
	os.WriteFile(keyPath, []byte("key"), 0644)
	sslCert = certPath
	sslKey = keyPath
	name = "test"
	// This will try to listen on TCP and broadcast, which will fail or succeed but we can check it doesn't return the earlier validation errors
	// We can mock the internal.ListenTCP etc by not actually calling rootCmd fully? Instead we just test that it passes the validation stage
	// For now, we just ensure it doesn't fail on the SSL validation
	// The full rootCmd will try to listen and then block forever (select{}), so we can't call it fully without timeout
	// Instead we just test validateTLS part
	if err := validateTLS(certPath, keyPath); err != nil {
		t.Fatalf("both exist should pass, got %v", err)
	}
}

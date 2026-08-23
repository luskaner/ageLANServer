package server

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"net/http"
	"strconv"
	"testing"
	"time"
)

func generateTestCertAndKey(t *testing.T) (certPem []byte, keyPem []byte) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	template := x509.Certificate{
		SerialNumber:          big.NewInt(time.Now().UnixNano()),
		Subject:               pkix.Name{CommonName: "test"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	keyDer, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	certPem = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPem = pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDer})
	return
}

func startTLSServer(t *testing.T, handler http.Handler) string {
	t.Helper()
	certPem, keyPem := generateTestCertAndKey(t)
	pair, err := tls.X509KeyPair(certPem, keyPem)
	if err != nil {
		t.Fatal(err)
	}
	srv := &http.Server{
		Handler:   handler,
		TLSConfig: &tls.Config{Certificates: []tls.Certificate{pair}},
	}
	ln, err := tls.Listen("tcp", "127.0.0.1:0", srv.TLSConfig)
	if err != nil {
		t.Skipf("cannot bind: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	go func() { _ = srv.Serve(ln) }()
	t.Cleanup(func() { _ = ln.Close() })
	return "127.0.0.1:" + strconv.Itoa(port)
}

func TestReadCACertificateFromServerSuccess(t *testing.T) {
	certPem, _ := generateTestCertAndKey(t)

	host := startTLSServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(certPem)
	}))

	result := ReadCACertificateFromServer(host)
	if result == nil {
		t.Fatal("expected valid certificate")
	}
}

// Regression: when the server returned non-200, resp.Body was never closed
// because the defer was registered only after io.ReadAll succeeded.
func TestReadCACertificateFromServerNon200CleansUp(t *testing.T) {
	host := startTLSServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))

	result := ReadCACertificateFromServer(host)
	if result != nil {
		t.Fatal("non-200 must return nil")
	}
}

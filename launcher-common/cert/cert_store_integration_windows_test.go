//go:build windows

package cert

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"os"
	"testing"
	"time"
)

// Opt-in integration test (AGELANSERVER_CERT_STORE_INTEGRATION=1): exercises
// the real Windows certificate store round-trip against the CURRENT_USER
// intermediate "CA" store, which accepts additions silently (unlike ROOT).
// It guards the iterateContext ownership contract: contexts consumed by
// CertDeleteCertificateFromStore must not be freed again by the caller.
func TestWindowsCertStoreRoundTrip(t *testing.T) {
	if os.Getenv("AGELANSERVER_CERT_STORE_INTEGRATION") != "1" {
		t.Skip("set AGELANSERVER_CERT_STORE_INTEGRATION=1 to touch the user certificate store")
	}

	const storeName = "CA"
	der, err := generateUniqueTestCert()
	if err != nil {
		t.Fatal(err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}

	if err = trustCertificatesInStore(true, storeName, []*x509.Certificate{cert}); err != nil {
		t.Fatalf("trust: %v", err)
	}

	found, err := enumCertificatesInStore(true, storeName)
	if err != nil {
		t.Fatalf("enum after trust: %v", err)
	}
	if !containsCert(found, cert) {
		t.Fatal("trusted certificate not found in store")
	}

	removed, err := untrustCertificatesFromStore(true, storeName)
	if err != nil {
		t.Fatalf("untrust: %v", err)
	}
	if !containsCert(removed, cert) {
		t.Fatal("untrust did not report removing the certificate")
	}

	found, err = enumCertificatesInStore(true, storeName)
	if err != nil {
		t.Fatalf("enum after untrust: %v", err)
	}
	if containsCert(found, cert) {
		t.Fatal("certificate still present after untrust")
	}
}

func generateUniqueTestCert() ([]byte, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	template := x509.Certificate{
		SerialNumber:          big.NewInt(time.Now().UnixNano()),
		Subject:               pkix.Name{CommonName: "ageLANServer store integration test"},
		NotBefore:             time.Now().Add(-time.Minute),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}
	return x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
}

func containsCert(certs []*x509.Certificate, want *x509.Certificate) bool {
	for _, c := range certs {
		if c.Equal(want) {
			return true
		}
	}
	return false
}

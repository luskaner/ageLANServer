package server

import (
	"common"
	"crypto/tls"
	"crypto/x509"
	"golang.org/x/sys/windows"
	"net"
	"shared/executor"
	"unsafe"
)

func connectToServer(host string, insecureSkipVerify bool) *tls.Conn {
	conf := &tls.Config{
		InsecureSkipVerify: insecureSkipVerify,
	}
	conn, err := tls.Dial("tcp", net.JoinHostPort(host, "443"), conf)
	if err != nil {
		return nil
	}
	return conn
}

func CheckConnectionFromServer(host string, insecureSkipVerify bool) bool {
	conn := connectToServer(host, insecureSkipVerify)
	if conn == nil {
		return false
	}
	defer func() {
		_ = conn.Close()
	}()
	return conn != nil
}

func readCertificateFromServer(host string) *x509.Certificate {
	conn := connectToServer(host, true)
	if conn == nil {
		return nil
	}
	defer func() {
		_ = conn.Close()
	}()
	certificates := conn.ConnectionState().PeerCertificates
	if len(certificates) > 0 {
		return certificates[0]
	}
	return nil
}

func TrustCertificateFromServer(host string) bool {
	cert := readCertificateFromServer(host)
	if cert == nil {
		return false
	}
	certBytes := cert.Raw
	certContext, err := windows.CertCreateCertificateContext(windows.X509_ASN_ENCODING|windows.PKCS_7_ASN_ENCODING, &certBytes[0], uint32(len(certBytes)))
	if err != nil {
		return false
	}

	defer func(ctx *windows.CertContext) {
		_ = windows.CertFreeCertificateContext(ctx)
	}(certContext)

	store, err := windows.CertOpenSystemStore(0, windows.StringToUTF16Ptr("ROOT"))
	if err != nil {
		return false
	}

	defer func(store windows.Handle, flags uint32) {
		_ = windows.CertCloseStore(store, flags)
	}(store, 0)

	err = windows.CertAddCertificateContextToStore(store, certContext, windows.CERT_STORE_ADD_NEW, nil)
	return err == nil
}

func UntrustCertificate() bool {
	store, err := windows.CertOpenSystemStore(0, windows.StringToUTF16Ptr("ROOT"))
	if err != nil {
		return false
	}
	defer func(store windows.Handle, flags uint32) {
		_ = windows.CertCloseStore(store, flags)
	}(store, 0)

	var certContext *windows.CertContext
	certContext, err = windows.CertFindCertificateInStore(store, windows.X509_ASN_ENCODING|windows.PKCS_7_ASN_ENCODING, 0, windows.CERT_FIND_SUBJECT_STR, unsafe.Pointer(windows.StringToUTF16Ptr(common.Domain)), nil)
	if err != nil {
		return false
	}

	err = windows.CertDeleteCertificateFromStore(certContext)
	return err == nil
}

func GenerateCertificatePair(certificateFolder string) bool {
	return executor.RunCustomExecutable(certificateFolder + `\..\..\genCert.exe`)
}

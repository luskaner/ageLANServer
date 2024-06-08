package shared

import (
	"common"
	"crypto/x509"
	"golang.org/x/sys/windows"
	"unsafe"
)

func openStore(userStore bool) (windows.Handle, error) {
	rootStr := windows.StringToUTF16Ptr("ROOT")
	if userStore {
		return windows.CertOpenSystemStore(0, rootStr)
	}
	return windows.CertOpenStore(windows.CERT_STORE_PROV_SYSTEM, 0, 0, windows.CERT_SYSTEM_STORE_LOCAL_MACHINE, uintptr(unsafe.Pointer(rootStr)))
}

func TrustCertificate(userStore bool, cert *x509.Certificate) bool {
	certBytes := cert.Raw
	certContext, err := windows.CertCreateCertificateContext(windows.X509_ASN_ENCODING|windows.PKCS_7_ASN_ENCODING, &certBytes[0], uint32(len(certBytes)))
	if err != nil {
		return false
	}

	defer func(ctx *windows.CertContext) {
		_ = windows.CertFreeCertificateContext(ctx)
	}(certContext)

	store, err := openStore(userStore)

	if err != nil {
		return false
	}

	defer func(store windows.Handle, flags uint32) {
		_ = windows.CertCloseStore(store, flags)
	}(store, 0)

	err = windows.CertAddCertificateContextToStore(store, certContext, windows.CERT_STORE_ADD_NEW, nil)
	return err == nil
}

func UntrustCertificate(userStore bool) bool {
	store, err := openStore(userStore)
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

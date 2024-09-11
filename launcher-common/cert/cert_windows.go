package cert

import (
	"crypto/x509"
	"github.com/luskaner/aoe2DELanServer/common"
	"golang.org/x/sys/windows"
	"unsafe"
)

func openStore(userStore bool) (windows.Handle, error) {
	rootStr := windows.StringToUTF16Ptr("ROOT")
	var flags uint32
	if userStore {
		flags = windows.CERT_SYSTEM_STORE_CURRENT_USER
	} else {
		flags = windows.CERT_SYSTEM_STORE_LOCAL_MACHINE
	}
	return windows.CertOpenStore(windows.CERT_STORE_PROV_SYSTEM, 0, 0, flags, uintptr(unsafe.Pointer(rootStr)))
}

func TrustCertificate(userStore bool, cert *x509.Certificate) error {
	certBytes := cert.Raw
	certContext, err := windows.CertCreateCertificateContext(windows.X509_ASN_ENCODING|windows.PKCS_7_ASN_ENCODING, &certBytes[0], uint32(len(certBytes)))
	if err != nil {
		return err
	}

	defer func(ctx *windows.CertContext) {
		_ = windows.CertFreeCertificateContext(ctx)
	}(certContext)

	store, err := openStore(userStore)

	if err != nil {
		return err
	}

	defer func(store windows.Handle, flags uint32) {
		_ = windows.CertCloseStore(store, flags)
	}(store, 0)

	return windows.CertAddCertificateContextToStore(store, certContext, windows.CERT_STORE_ADD_NEW, nil)
}

func UntrustCertificate(userStore bool) (cert *x509.Certificate, err error) {
	var store windows.Handle
	store, err = openStore(userStore)
	if err != nil {
		return
	}
	defer func(store windows.Handle, flags uint32) {
		_ = windows.CertCloseStore(store, flags)
	}(store, 0)

	var certContext *windows.CertContext
	certContext, err = windows.CertFindCertificateInStore(store, windows.X509_ASN_ENCODING|windows.PKCS_7_ASN_ENCODING, 0, windows.CERT_FIND_SUBJECT_STR, unsafe.Pointer(windows.StringToUTF16Ptr(common.CertSubjectOrganization)), nil)

	if err != nil {
		return
	}
	certBytes := make([]byte, certContext.Length)
	for i := range certBytes {
		certBytes[i] = *(*byte)(unsafe.Pointer(uintptr(unsafe.Pointer(certContext.EncodedCert)) + uintptr(i)))
	}
	cert, err = x509.ParseCertificate(certBytes)
	if err != nil {
		return
	}
	err = windows.CertDeleteCertificateFromStore(certContext)
	if err != nil {
		return
	}
	return
}

func SupportsUserStore() bool {
	return true
}

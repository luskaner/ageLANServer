package server

import (
	"crypto/tls"
	"crypto/x509"
	"net"
	"path/filepath"
	"shared/executor"
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

func ReadCertificateFromServer(host string) *x509.Certificate {
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

func GenerateCertificatePair(certificateFolder string) bool {
	return executor.RunCustomExecutable(filepath.Join(certificateFolder, "..", "..", "genCert.exe"))
}

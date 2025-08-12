package server

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"github.com/luskaner/ageLANServer/common"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/luskaner/ageLANServer/launcher-common/executor/exec"
	"net"
	"os"
	"path/filepath"
	"time"
)

func TlsConfig(serverName string, insecureSkipVerify bool) *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: insecureSkipVerify,
		ServerName:         serverName,
	}
}

func connectToServer(host string, IPProtocol *common.IPProtocol, insecureSkipVerify bool) *tls.Conn {
	ipAddrs := launcherCommon.AddrToIpAddrs(host, IPProtocol.IPv4(), IPProtocol.IPv6())
	if ipAddrs.IsEmpty() {
		return nil
	}
	for ipAddr := range ipAddrs.Iter() {
		network := "tcp"
		if *IPProtocol != common.IPvDual {
			if ipAddr.Is4() {
				network += "4"
			} else {
				network += "6"
			}
		}
		conn, err := tls.Dial(
			network,
			net.JoinHostPort(ipAddr.String(), "https"), TlsConfig(host, insecureSkipVerify),
		)
		if err == nil {
			return conn
		}
	}
	return nil
}

func CheckConnectionFromServer(host string, IPProtocol *common.IPProtocol, insecureSkipVerify bool) bool {
	conn := connectToServer(host, IPProtocol, insecureSkipVerify)
	if conn == nil {
		return false
	}
	defer func() {
		_ = conn.Close()
	}()
	return conn != nil
}

func ReadCertificateFromServer(host string, IPProtocol *common.IPProtocol) *x509.Certificate {
	conn := connectToServer(host, IPProtocol, true)
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

func GenerateCertificatePair(certificateFolder string) (result *exec.Result) {
	baseFolder := filepath.Join(certificateFolder, "..", "..")
	exePath := filepath.Join(baseFolder, common.GetExeFileName(false, common.ServerGenCert))
	if _, err := os.Stat(exePath); err != nil {
		return nil
	}
	result = exec.Options{File: exePath, Wait: true, Args: []string{"-r"}, ExitCode: true}.Exec()
	return
}

func CertificateSoonExpired(cert string) bool {
	if cert == "" {
		return true
	}

	certPEM, err := os.ReadFile(cert)
	if err != nil {
		return true
	}

	block, _ := pem.Decode(certPEM)
	if block == nil {
		return true
	}

	crt, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return true
	}

	return time.Now().Add(24 * time.Hour).After(crt.NotAfter)
}

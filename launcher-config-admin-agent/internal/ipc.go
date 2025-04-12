package internal

import (
	"crypto/x509"
	"encoding/gob"
	"github.com/luskaner/ageLANServer/common"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/luskaner/ageLANServer/launcher-common/executor"
	"net"
	"slices"
)

var mappedCdn = false
var mappedIps = false
var addedCert = false

func handleClient(c net.Conn) (exit bool) {
	exit = false
	decoder := gob.NewDecoder(c)
	encoder := gob.NewEncoder(c)
	var action byte
	var exitCode = ErrNonExistingAction
	var err error

	for !exit {
		if err = decoder.Decode(&action); err != nil {
			_ = encoder.Encode(ErrDecode)
			return
		}

		switch action {
		case launcherCommon.ConfigAdminIpcRevert:
			_ = encoder.Encode(common.ErrSuccess)
			exitCode = handleRevert(decoder)
		case launcherCommon.ConfigAdminIpcSetup:
			_ = encoder.Encode(common.ErrSuccess)
			exitCode = handleSetUp(decoder)
		case launcherCommon.ConfigAdminIpcExit:
			err = c.Close()
			if err != nil {
				exitCode = ErrConnectionClosing
			} else {
				exit = true
			}
		}

		_ = encoder.Encode(exitCode)
	}

	return
}

func checkCertificateValidity(cert *x509.Certificate) bool {
	if cert == nil {
		return false
	}
	hosts := common.AllHosts()
	if cert.Subject.CommonName != common.Name {
		return false
	}

	if !slices.Equal(cert.DNSNames, hosts) {
		return false
	}
	return true
}

func checkIps(ips []net.IP) bool {
	return len(ips) < 10
}

func handleSetUp(decoder *gob.Decoder) int {
	var msg launcherCommon.ConfigAdminIpcSetupCommand
	if err := decoder.Decode(&msg); err != nil {
		return ErrDecode
	}
	if len(msg.IPs) > 0 && mappedIps {
		return ErrIpsAlreadyMapped
	}
	if msg.CDN && mappedCdn {
		return ErrCDNAlreadyMapped
	}
	if !checkIps(msg.IPs) {
		return ErrIpsInvalid
	}
	var cert *x509.Certificate
	if msg.Certificate != nil {
		if addedCert {
			return ErrCertAlreadyAdded
		}
		cert, _ = x509.ParseCertificate(msg.Certificate)
		if !checkCertificateValidity(cert) {
			return ErrCertInvalid
		}
	}
	result := executor.RunSetUp(msg.IPs, cert, msg.CDN)
	if result.Success() {
		mappedIps = mappedIps || len(msg.IPs) > 0
		mappedCdn = mappedCdn || msg.CDN
		addedCert = addedCert || cert != nil
	}
	return result.ExitCode
}

func handleRevert(decoder *gob.Decoder) int {
	var msg launcherCommon.ConfigAdminIpcRevertCommand
	if err := decoder.Decode(&msg); err != nil {
		return ErrDecode
	}
	revertIps := msg.IPs && mappedIps
	revertCert := msg.Certificate && addedCert
	revertCdn := msg.CDN && mappedCdn
	if !revertIps && !revertCert {
		return common.ErrSuccess
	}
	result := executor.RunRevert(revertIps, revertCert, revertCdn, true)
	if result.Success() {
		mappedCdn = !revertCdn
		mappedIps = !revertIps
		addedCert = !revertCert
	}
	return result.ExitCode
}

func RunIpcServer() (errorCode int) {
	l, err := SetupIpcServer()
	if err != nil {
		errorCode = ErrListen
		return
	}
	defer func(l net.Listener) {
		_ = l.Close()
		RevertIpcServer()
	}(l)

	var conn net.Conn
	for {
		conn, err = l.Accept()
		if err != nil {
			continue
		}
		if handleClient(conn) {
			break
		}
	}
	return
}

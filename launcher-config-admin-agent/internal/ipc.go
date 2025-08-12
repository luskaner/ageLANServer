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
var mappedIp = false
var addedCert = false

func handleClient(c net.Conn) (exit bool) {
	exit = false
	decoder := gob.NewDecoder(c)
	encoder := gob.NewEncoder(c)
	var action byte
	var err error

	for !exit {
		if err = decoder.Decode(&action); err != nil {
			_ = encoder.Encode(ErrDecode)
			return
		}

		var exitCode = ErrNonExistingAction

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
				exitCode = common.ErrSuccess
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

func handleSetUp(decoder *gob.Decoder) int {
	var msg launcherCommon.ConfigAdminIpcSetupCommand
	if err := decoder.Decode(&msg); err != nil {
		return ErrDecode
	}
	if msg.IPAddr.IsValid() && mappedIp {
		return ErrIpAlreadyMapped
	}
	if msg.CDN && mappedCdn {
		return ErrCDNAlreadyMapped
	}
	var cert *x509.Certificate
	if msg.Certificate != nil {
		if addedCert {
			return ErrCertAlreadyAdded
		}
		var err error
		cert, err = x509.ParseCertificate(msg.Certificate)
		if err != nil || !checkCertificateValidity(cert) {
			return ErrCertInvalid
		}
	}
	result := executor.RunSetUp(msg.IPAddr, cert, msg.CDN)
	if result.Success() {
		mappedIp = mappedIp || msg.IPAddr.IsValid()
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
	revertIp := msg.IPAddr && mappedIp
	revertCert := msg.Certificate && addedCert
	revertCdn := msg.CDN && mappedCdn
	if !revertIp && !revertCert && !revertCdn {
		return common.ErrSuccess
	}
	result := executor.RunRevert(revertIp, revertCert, revertCdn, true)
	if result.Success() {
		mappedCdn = mappedCdn && !revertCdn
		mappedIp = mappedIp && !revertIp
		addedCert = addedCert && !revertCert
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

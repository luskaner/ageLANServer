package internal

import (
	"crypto/x509"
	"encoding/gob"
	"net"

	"github.com/luskaner/ageLANServer/common"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/luskaner/ageLANServer/launcher-common/executor"
)

var mappedIps = false
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
	return cert != nil
}

func handleSetUp(decoder *gob.Decoder) int {
	var msg launcherCommon.ConfigAdminIpcSetupCommand
	if err := decoder.Decode(&msg); err != nil {
		return ErrDecode
	}
	if len(msg.IP) > 0 && mappedIps {
		return ErrIpsAlreadyMapped
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
	result := executor.RunSetUp(msg.GameId, msg.IP, cert, msg.CDN)
	if result.Success() {
		mappedIps = mappedIps || len(msg.IP) > 0
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
	if !revertIps && !revertCert {
		return common.ErrSuccess
	}
	result := executor.RunRevert(revertIps, revertCert, true)
	if result.Success() {
		mappedIps = mappedIps && !revertIps
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

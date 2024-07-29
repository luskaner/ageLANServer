package internal

import (
	"common"
	"crypto/x509"
	"encoding/gob"
	"fmt"
	"github.com/Microsoft/go-winio"
	"golang.org/x/sys/windows"
	launcherCommon "launcher-common"
	"launcher-common/executor"
	"net"
)

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
			fmt.Println("Error decoding action: ", err)
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
				fmt.Println("Error closing connection: ", err)
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
	if cert.Subject.CommonName != common.Domain {
		return false
	}
	if len(cert.DNSNames) != 1 || cert.DNSNames[0] != common.Domain {
		return false
	}
	return true
}

func checkIps(ips []net.IP) bool {
	return len(ips) < 10
}

func userSid() (err error, sid string) {
	err, token := executor.GetCurrentProcessToken()
	if err != nil {
		return
	}
	defer func(token windows.Token) {
		_ = token.Close()
	}(token)

	tokenUser, err := token.GetTokenUser()
	if err != nil {
		return
	}

	sid = tokenUser.User.Sid.String()
	return
}

func handleSetUp(decoder *gob.Decoder) int {
	var msg launcherCommon.ConfigAdminIpcSetupCommand
	if err := decoder.Decode(&msg); err != nil {
		return ErrDecode
	}
	if len(msg.IPs) > 0 && mappedIps {
		return ErrIpsAlreadyMapped
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
	result := executor.RunSetUp(msg.IPs, cert)
	if result.Success() {
		mappedIps = mappedIps || len(msg.IPs) > 0
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
		mappedIps = !revertIps
		addedCert = !revertCert
	}
	return result.ExitCode
}

func RunIpcServer() (errorCode int) {
	pipePath := launcherCommon.ConfigAdminIpcPipe
	_, sid := userSid()
	pc := &winio.PipeConfig{
		InputBufferSize:    1024,
		OutputBufferSize:   1,
		SecurityDescriptor: fmt.Sprintf("D:P(A;;GA;;;%s)", sid),
		MessageMode:        true,
	}

	l, err := winio.ListenPipe(pipePath, pc)
	if err != nil {
		fmt.Println("Error creating pipe: ", err)
		errorCode = ErrCreatePipe
	}
	defer func(l net.Listener) {
		_ = l.Close()
	}(l)

	var conn net.Conn
	for {
		conn, err = l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err)
			continue
		}
		if handleClient(conn) {
			fmt.Println("Client requested exit")
			break
		}
	}
	return
}

package internal

import (
	"crypto/x509"
	"encoding/gob"
	"io"
	"net"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executor/exec"
	"github.com/luskaner/ageLANServer/common/logger"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/luskaner/ageLANServer/launcher-common/executor"
)

var mappedIps = false
var addedCert = false

func handleClient(logRoot string, c net.Conn) (exit bool) {
	exit = false
	decoder := gob.NewDecoder(c)
	encoder := gob.NewEncoder(c)
	var action byte
	var err error

	for !exit {
		if err = decoder.Decode(&action); err != nil {
			commonLogger.Println("Could not decode action:", err)
			str := "-> ErrDecode: "
			if err = encoder.Encode(ErrDecode); err != nil {
				str += err.Error()
			} else {
				str += "OK"
			}
			commonLogger.Println(str)
			return
		}

		var exitCode = ErrNonExistingAction

		switch action {
		case launcherCommon.ConfigAdminIpcRevert:
			str := "<- ConfigAdminIpcRevert: "
			if err = encoder.Encode(common.ErrSuccess); err != nil {
				str += err.Error()
			} else {
				str += "OK"
			}
			commonLogger.Println(str)
			exitCode = handleRevert(logRoot, decoder)
		case launcherCommon.ConfigAdminIpcSetup:
			str := "<- ConfigAdminIpcSetup: "
			if err = encoder.Encode(common.ErrSuccess); err != nil {
				str += err.Error()
			} else {
				str += "OK"
			}
			commonLogger.Println(str)
			exitCode = handleSetUp(logRoot, decoder)
		case launcherCommon.ConfigAdminIpcExit:
			str := "<- ConfigAdminIpcExit: "
			err = c.Close()
			if err != nil {
				str += err.Error()
				exitCode = ErrConnectionClosing
			} else {
				str += "OK"
				exit = true
				exitCode = common.ErrSuccess
			}
			commonLogger.Println(str)
		}

		_ = encoder.Encode(exitCode)
	}

	return
}

func checkCertificateValidity(cert *x509.Certificate) bool {
	return cert != nil
}

func handleSetUp(logRoot string, decoder *gob.Decoder) int {
	var msg launcherCommon.ConfigAdminIpcSetupCommand
	commonLogger.Println("<- ConfigAdminIpcSetupCommand")
	if err := decoder.Decode(&msg); err != nil {
		commonLogger.Println("Could not decode command:", err)
		return ErrDecode
	}
	commonLogger.Printf("<- %v\n", msg)
	if len(msg.IP) > 0 && mappedIps {
		commonLogger.Println("IPs already mapped")
		return ErrIpsAlreadyMapped
	}
	var cert *x509.Certificate
	if msg.Certificate != nil {
		if addedCert {
			commonLogger.Println("certificate already added")
			return ErrCertAlreadyAdded
		}
		str := "Parsing certificate: "
		var err error
		cert, err = x509.ParseCertificate(msg.Certificate)
		if err != nil || !checkCertificateValidity(cert) {
			if err != nil {
				str += err.Error()
			} else {
				str += "invalid"
			}
			return ErrCertInvalid
		} else {
			str += "OK"
		}
		commonLogger.Println(str)
	} else {
		commonLogger.Println("No certificate")
	}
	var suffix string
	if cert != nil {
		suffix = "_cert"
	} else {
		suffix = "_hosts"
	}
	var result *exec.Result
	if buffErr := commonLogger.FileLogger.Buffer("config-admin_setup"+suffix, func(writer io.Writer) {
		result = executor.RunSetUp(msg.GameId, msg.IP, cert, msg.CDN, logRoot, writer, func(options exec.Options) {
			if writer != nil {
				commonLogger.Println("run config admin setup", options.String())
			}
		})
	}); buffErr != nil {
		return common.ErrFileLog
	}
	if result.Success() {
		mappedIps = mappedIps || len(msg.IP) > 0
		addedCert = addedCert || cert != nil
	}
	return result.ExitCode
}

func handleRevert(logRoot string, decoder *gob.Decoder) int {
	var msg launcherCommon.ConfigAdminIpcRevertCommand
	commonLogger.Println("<- ConfigAdminIpcRevertCommand")
	if err := decoder.Decode(&msg); err != nil {
		commonLogger.Println("Could not decode command:", err)
		return ErrDecode
	}
	commonLogger.Printf("<- %v\n", msg)
	revertIps := msg.IPs && mappedIps
	revertCert := msg.Certificate && addedCert
	if !revertIps && !revertCert {
		commonLogger.Println("Everything is already reverted.")
		return common.ErrSuccess
	}
	var result *exec.Result
	if buffErr := commonLogger.FileLogger.Buffer("config-admin_revert", func(writer io.Writer) {
		result = executor.RunRevert(revertIps, revertCert, true, logRoot, writer, func(options exec.Options) {
			if writer != nil {
				commonLogger.Println("run config admin revert", options.String())
			}
		})
	}); buffErr != nil {
		return common.ErrFileLog
	}
	if result.Success() {
		mappedIps = mappedIps && !revertIps
		addedCert = addedCert && !revertCert
	}
	return result.ExitCode
}

func RunIpcServer(logRoot string) (errorCode int) {
	l, err := SetupIpcServer()
	if err != nil {
		commonLogger.Printf("Could not listen to IPC: %v\n", err)
		errorCode = ErrListen
		return
	}
	defer func(l net.Listener) {
		_ = l.Close()
		RevertIpcServer()
	}(l)

	var conn net.Conn
	for {
		commonLogger.Println("Waiting for connection...")
		conn, err = l.Accept()
		if err != nil {
			commonLogger.Printf("Could not accept connection: %v\n", err)
			continue
		}
		commonLogger.Println("Accepted connection: ", conn.RemoteAddr().String())
		if handleClient(logRoot, conn) {
			break
		}
	}
	return
}

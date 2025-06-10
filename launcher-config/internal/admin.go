package internal

import (
	"crypto/x509"
	"encoding/gob"
	"fmt"
	"github.com/luskaner/ageLANServer/common"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/luskaner/ageLANServer/launcher-common/cert"
	"github.com/luskaner/ageLANServer/launcher-common/executor"
	"github.com/luskaner/ageLANServer/launcher-common/executor/exec"
	"net"
	"time"
)

var ipc net.Conn = nil
var encoder *gob.Encoder = nil
var decoder *gob.Decoder = nil

func RunSetUp(mapIp string, addCertData []byte, mapCDN bool) (err error, exitCode int) {
	exitCode = common.ErrGeneral
	ip := net.ParseIP(mapIp)
	if ipc != nil {
		return runSetUpAgent(ip, addCertData, mapCDN)
	} else {
		var certificate *x509.Certificate
		if addCertData != nil {
			certificate = cert.BytesToCertificate(addCertData)
			if certificate == nil {
				exitCode = ErrUserCertAddParse
				return
			}
		}
		result := executor.RunSetUp(ip, certificate, mapCDN)
		err, exitCode = result.Err, result.ExitCode
	}
	return
}

func RunRevert(unmapIP bool, removeCert bool, unmapCDN bool, failfast bool) (err error, exitCode int) {
	if ipc != nil {
		return runRevertAgent(unmapIP, removeCert, unmapCDN)
	}
	result := executor.RunRevert(unmapIP, removeCert, unmapCDN, failfast)
	err, exitCode = result.Err, result.ExitCode
	return
}

func StopAgentIfNeeded() (err error) {
	if ipc != nil {
		err = encoder.Encode(launcherCommon.ConfigAdminIpcExit)
		if err != nil {
			return
		}
		err = ipc.Close()
		if err != nil {
			return
		}
		encoder = nil
		decoder = nil
		ipc = nil
	}
	return
}

func ConnectAgentIfNeededWithRetries(retryUntilSuccess bool) bool {
	var ok bool
	for i := 0; i < 30; i++ {
		ok = ConnectAgentIfNeeded() == nil
		if retryUntilSuccess == ok {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}

func ConnectAgentIfNeeded() (err error) {
	if ipc != nil {
		return
	}
	var conn net.Conn
	conn, err = DialIPC()
	if err != nil {
		return
	}
	ipc = conn
	encoder = gob.NewEncoder(ipc)
	decoder = gob.NewDecoder(ipc)
	return
}

func StartAgentIfNeeded() (result *exec.Result) {
	if ipc != nil {
		return
	}
	fmt.Println("Starting 'agent'...")
	preAgentStart()
	file := common.GetExeFileName(true, common.LauncherConfigAdminAgent)
	result = exec.Options{File: file, AsAdmin: true, Pid: true}.Exec()
	if result.Success() {
		postAgentStart(file)
	}
	return
}

func runRevertAgent(unmapIP bool, removeCert bool, unmapCDN bool) (err error, exitCode int) {
	if err = encoder.Encode(launcherCommon.ConfigAdminIpcRevert); err != nil {
		return
	}

	if err = decoder.Decode(&exitCode); err != nil || exitCode != common.ErrSuccess {
		return
	}

	if err = encoder.Encode(launcherCommon.ConfigAdminIpcRevertCommand{IP: unmapIP, Certificate: removeCert, CDN: unmapCDN}); err != nil {
		return
	}

	if err = decoder.Decode(&exitCode); err != nil {
		return
	}

	return
}

func runSetUpAgent(mapIp net.IP, certificate []byte, mapCDN bool) (err error, exitCode int) {
	if err = encoder.Encode(launcherCommon.ConfigAdminIpcSetup); err != nil {
		return
	}

	if err = decoder.Decode(&exitCode); err != nil || exitCode != common.ErrSuccess {
		return
	}

	if err = encoder.Encode(launcherCommon.ConfigAdminIpcSetupCommand{IP: mapIp, Certificate: certificate, CDN: mapCDN}); err != nil {
		return
	}

	if err = decoder.Decode(&exitCode); err != nil {
		return
	}

	return
}

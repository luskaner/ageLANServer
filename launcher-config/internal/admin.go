package internal

import (
	"crypto/x509"
	"encoding/gob"
	"fmt"
	"net"
	"time"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executor/exec"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/luskaner/ageLANServer/launcher-common/cert"
	"github.com/luskaner/ageLANServer/launcher-common/cmd"
	"github.com/luskaner/ageLANServer/launcher-common/executor"
)

var ipc net.Conn = nil
var encoder *gob.Encoder = nil
var decoder *gob.Decoder = nil

func RunSetUp(ipToMap net.IP, addCertData []byte, mapCDN bool) (err error, exitCode int) {
	exitCode = common.ErrGeneral
	if ipc != nil {
		return runSetUpAgent(ipToMap, addCertData, mapCDN)
	} else {
		var certificate *x509.Certificate
		if addCertData != nil {
			certificate = cert.BytesToCertificate(addCertData)
			if certificate == nil {
				exitCode = ErrUserCertAddParse
				return
			}
		}
		result := executor.RunSetUp(cmd.GameId, ipToMap, certificate, mapCDN)
		err, exitCode = result.Err, result.ExitCode
	}
	return
}

func RunRevert(unmapIPs bool, removeCert bool, failfast bool) (err error, exitCode int) {
	if ipc != nil {
		return runRevertAgent(unmapIPs, removeCert)
	}
	result := executor.RunRevert(unmapIPs, removeCert, failfast)
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

func runRevertAgent(unmapIPs bool, removeCert bool) (err error, exitCode int) {
	if err = encoder.Encode(launcherCommon.ConfigAdminIpcRevert); err != nil {
		return
	}

	if err = decoder.Decode(&exitCode); err != nil || exitCode != common.ErrSuccess {
		return
	}

	if err = encoder.Encode(launcherCommon.ConfigAdminIpcRevertCommand{IPs: unmapIPs, Certificate: removeCert}); err != nil {
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

	if err = encoder.Encode(launcherCommon.ConfigAdminIpcSetupCommand{GameId: cmd.GameId, IP: mapIp, Certificate: certificate, CDN: mapCDN}); err != nil {
		return
	}

	if err = decoder.Decode(&exitCode); err != nil {
		return
	}

	return
}

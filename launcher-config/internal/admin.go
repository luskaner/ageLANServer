package internal

import (
	"crypto/x509"
	"encoding/gob"
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/aoe2DELanServer/common"
	launcherCommon "github.com/luskaner/aoe2DELanServer/launcher-common"
	"github.com/luskaner/aoe2DELanServer/launcher-common/cert"
	"github.com/luskaner/aoe2DELanServer/launcher-common/executor"
	"github.com/luskaner/aoe2DELanServer/launcher-common/executor/exec"
	"net"
	"time"
)

var ipc net.Conn = nil
var encoder *gob.Encoder = nil
var decoder *gob.Decoder = nil

func RunSetUp(mapIps mapset.Set[string], addCertData []byte, mapCDN bool) (err error, exitCode int) {
	exitCode = common.ErrGeneral
	if mapIps.Cardinality() > 9 {
		exitCode = launcherCommon.ErrIpMapAddTooMany
		return
	}
	ips := make([]net.IP, mapIps.Cardinality())
	i := 0
	for ip := range mapIps.Iter() {
		ips[i] = net.ParseIP(ip)
		if ips[i] == nil {
			return
		}
		i++
	}
	if ipc != nil {
		return runSetUpAgent(ips, addCertData, mapCDN)
	} else {
		var certificate *x509.Certificate
		if addCertData != nil {
			certificate = cert.BytesToCertificate(addCertData)
			if certificate == nil {
				exitCode = ErrUserCertAddParse
				return
			}
		}
		result := executor.RunSetUp(ips, certificate, mapCDN)
		err, exitCode = result.Err, result.ExitCode
	}
	return
}

func RunRevert(unmapIPs bool, removeCert bool, unmapCDN bool, failfast bool) (err error, exitCode int) {
	if ipc != nil {
		return runRevertAgent(unmapIPs, removeCert, unmapCDN)
	}
	result := executor.RunRevert(unmapIPs, removeCert, unmapCDN, failfast)
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
	conn, err = net.Dial("unix", launcherCommon.ConfigAdminIpcName())
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
	fmt.Println("Starting agent...")
	preAgentStart()
	file := common.GetExeFileName(true, common.LauncherConfigAdminAgent)
	result = exec.Options{File: file, AsAdmin: true, Pid: true}.Exec()
	if result.Success() {
		postAgentStart(file)
	}
	return
}

func runRevertAgent(unmapIPs bool, removeCert bool, unmapCDN bool) (err error, exitCode int) {
	if err = encoder.Encode(launcherCommon.ConfigAdminIpcRevert); err != nil {
		return
	}

	if err = decoder.Decode(&exitCode); err != nil || exitCode != common.ErrSuccess {
		return
	}

	if err = encoder.Encode(launcherCommon.ConfigAdminIpcRevertCommand{IPs: unmapIPs, Certificate: removeCert, CDN: unmapCDN}); err != nil {
		return
	}

	if err = decoder.Decode(&exitCode); err != nil {
		return
	}

	return
}

func runSetUpAgent(mapIps []net.IP, certificate []byte, mapCDN bool) (err error, exitCode int) {
	if err = encoder.Encode(launcherCommon.ConfigAdminIpcSetup); err != nil {
		return
	}

	if err = decoder.Decode(&exitCode); err != nil || exitCode != common.ErrSuccess {
		return
	}

	if err = encoder.Encode(launcherCommon.ConfigAdminIpcSetupCommand{IPs: mapIps, Certificate: certificate, CDN: mapCDN}); err != nil {
		return
	}

	if err = decoder.Decode(&exitCode); err != nil {
		return
	}

	return
}

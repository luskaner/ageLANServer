package internal

import (
	"crypto/x509"
	"encoding/gob"
	"github.com/Microsoft/go-winio"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/aoe2DELanServer/common"
	launcherCommon "github.com/luskaner/aoe2DELanServer/launcher-common"
	"github.com/luskaner/aoe2DELanServer/launcher-common/executor"
	"net"
	"time"
)

var pipe net.Conn = nil
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
	if pipe != nil {
		return runSetUpAgent(ips, addCertData, mapCDN)
	} else {
		var certificate *x509.Certificate
		if addCertData != nil {
			certificate = launcherCommon.BytesToCertificate(addCertData)
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
	if pipe != nil {
		return runRevertAgent(unmapIPs, removeCert, unmapCDN)
	}
	result := executor.RunRevert(unmapIPs, removeCert, unmapCDN, failfast)
	err, exitCode = result.Err, result.ExitCode
	return
}

func StopAgentIfNeeded() (err error) {
	if pipe != nil {
		err = encoder.Encode(launcherCommon.ConfigAdminIpcExit)
		if err != nil {
			return
		}
		err = pipe.Close()
		if err != nil {
			return
		}
		encoder = nil
		decoder = nil
		pipe = nil
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
	if pipe != nil {
		return
	}
	var conn net.Conn
	conn, err = winio.DialPipe(launcherCommon.ConfigAdminIpcPipe, nil)
	if err != nil {
		return
	}
	pipe = conn
	encoder = gob.NewEncoder(pipe)
	decoder = gob.NewDecoder(pipe)
	return
}

func StartAgentIfNeeded() (result *executor.ExecResult) {
	if pipe != nil {
		return
	}
	result = executor.ExecOptions{File: common.GetExeFileName(true, common.LauncherConfigAdminAgent), AsAdmin: true, Pid: true}.Exec()
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

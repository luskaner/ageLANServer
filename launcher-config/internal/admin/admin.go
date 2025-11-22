package admin

import (
	"crypto/x509"
	"encoding/gob"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executables"
	"github.com/luskaner/ageLANServer/common/executor/exec"
	"github.com/luskaner/ageLANServer/common/logger"
	"github.com/luskaner/ageLANServer/launcher-common/cert"
	"github.com/luskaner/ageLANServer/launcher-common/cmd"
	"github.com/luskaner/ageLANServer/launcher-common/executor"
	commonIpc "github.com/luskaner/ageLANServer/launcher-common/ipc"
	"github.com/luskaner/ageLANServer/launcher-config/internal"
)

var ipc net.Conn = nil
var encoder *gob.Encoder = nil
var decoder *gob.Decoder = nil

func RunSetUp(logRoot string, ipToMap net.IP, addCertData []byte) (err error, exitCode int) {
	exitCode = common.ErrGeneral
	if ipc != nil {
		return runSetUpAgent(ipToMap, addCertData)
	} else {
		var certificate *x509.Certificate
		if addCertData != nil {
			certificate = cert.BytesToCertificate(addCertData)
			if certificate == nil {
				exitCode = internal.ErrUserCertAddParse
				return
			}
		}
		var result *exec.Result
		var file *commonLogger.Root
		if logRoot != "" {
			if err, file = commonLogger.NewFile(logRoot, "", true); err != nil {
				exitCode = common.ErrFileLog
				return
			}
		}
		var suffix string
		if len(addCertData) > 0 {
			suffix = "_cert"
		} else {
			suffix = "_hosts"
		}
		if bufferErr := file.Buffer("config-admin_setup"+suffix, func(writer io.Writer) {
			result = executor.RunSetUp(cmd.GameId, ipToMap, certificate, file.Folder(), writer, func(options exec.Options) {
				if writer != nil {
					options.Stdout = writer
					options.Stderr = writer
				}
			})
		}); bufferErr == nil {
			err, exitCode = result.Err, result.ExitCode
		} else {
			err = bufferErr
			exitCode = common.ErrFileLog
		}
	}
	return
}

func RunRevert(logRoot string, unmapIPs bool, removeCert bool, failfast bool) (err error, exitCode int) {
	if ipc != nil {
		return runRevertAgent(unmapIPs, removeCert)
	}
	var result *exec.Result
	var file *commonLogger.Root
	if logRoot != "" {
		if err, file = commonLogger.NewFile(logRoot, "", true); err != nil {
			exitCode = common.ErrFileLog
			return
		}
	}
	if bufferErr := file.Buffer("config-admin_revert", func(writer io.Writer) {
		result = executor.RunRevert(unmapIPs, removeCert, failfast, file.Folder(), writer, func(options exec.Options) {
			if writer != nil {
				options.Stdout = writer
				options.Stderr = writer
			}
		})
	}); bufferErr == nil {
		err, exitCode = result.Err, result.ExitCode
	} else {
		err = bufferErr
		exitCode = common.ErrFileLog
	}
	return
}

func StopAgentIfNeeded() (err error) {
	commonLogger.Println("Stopping agent")
	if ipc != nil {
		str := "-> Exit: "
		err = encoder.Encode(commonIpc.Exit)
		if err != nil {
			commonLogger.Println(str + "Could not encode")
			return
		}
		commonLogger.Println(str + "OK")
		str = "Closing connection: "
		err = ipc.Close()
		if err != nil {
			commonLogger.Println(str + "Could not close")
			return
		}
		commonLogger.Println(str + "OK")
		encoder = nil
		decoder = nil
		ipc = nil
	} else {
		commonLogger.Println("Already stopped")
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
	commonLogger.Println("Connecting to agent")
	if ipc != nil {
		commonLogger.Println("Already connected")
		return
	}
	var conn net.Conn
	conn, err = DialIPC()
	if err != nil {
		return
	}
	commonLogger.Println("Connected")
	ipc = conn
	encoder = gob.NewEncoder(ipc)
	decoder = gob.NewDecoder(ipc)
	return
}

func StartAgentIfNeeded() (result *exec.Result) {
	commonLogger.Println("Starting agent")
	if ipc != nil {
		commonLogger.Println("Already started")
		return
	}
	preAgentStart()
	file := executables.Filename(true, executables.LauncherConfigAdminAgent)
	options := exec.Options{File: file, AsAdmin: true, Pid: true}
	if internal.Logger != nil {
		options.Args = []string{internal.Logger.Folder()}
		commonLogger.Println("start config-admin-agent:", options.String())
	} else {
		options.Args = []string{"-"}
	}
	result = options.Exec()
	if result.Success() {
		postAgentStart(file)
	}
	return
}

func runRevertAgent(unmapIPs bool, removeCert bool) (err error, exitCode int) {
	str := "-> Revert: "
	if err = encoder.Encode(commonIpc.Revert); err != nil {
		commonLogger.Println(str + "Could not encode")
		return
	} else {
		commonLogger.Println(str + "OK")
	}
	str = "<- Exit Code: "
	if err = decoder.Decode(&exitCode); err != nil || exitCode != common.ErrSuccess {
		if err != nil {
			commonLogger.Println(str + "Could not decode")
		} else {
			commonLogger.Println(str + strconv.Itoa(exitCode))
		}
		return
	}
	commonLogger.Println(str + strconv.Itoa(exitCode))
	data := commonIpc.RevertCommand{IPs: unmapIPs, Certificate: removeCert}
	str = fmt.Sprintf("-> %v: ", data)
	if err = encoder.Encode(data); err != nil {
		commonLogger.Println(str + "Could not encode")
		return
	}
	commonLogger.Println(str + "OK")
	str = "<- Exit Code: "
	if err = decoder.Decode(&exitCode); err != nil {
		commonLogger.Println(str + "Could not decode")
		return
	}
	commonLogger.Println(str + strconv.Itoa(exitCode))
	return
}

func runSetUpAgent(mapIp net.IP, certificate []byte) (err error, exitCode int) {
	str := "-> Setup: "
	if err = encoder.Encode(commonIpc.Setup); err != nil {
		commonLogger.Println(str + "Could not decode")
		return
	}
	commonLogger.Println(str + "OK")
	str = "<- Exit Code: "
	if err = decoder.Decode(&exitCode); err != nil || exitCode != common.ErrSuccess {
		if err != nil {
			commonLogger.Println(str + "Could not decode")
		} else {
			commonLogger.Println(str + strconv.Itoa(exitCode))
		}
		return
	}
	commonLogger.Println(str + strconv.Itoa(exitCode))
	data := commonIpc.SetupCommand{GameId: cmd.GameId, IP: mapIp, Certificate: certificate}
	str = fmt.Sprintf("-> %v: ", data)
	if err = encoder.Encode(data); err != nil {
		commonLogger.Println(str + "Could not encode")
		return
	}
	commonLogger.Println(str + "OK")
	str = "<- Exit Code: "
	if err = decoder.Decode(&exitCode); err != nil {
		commonLogger.Println(str + "Could not decode")
		return
	}
	commonLogger.Println(str + strconv.Itoa(exitCode))
	return
}

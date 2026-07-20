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
	commonProcess "github.com/luskaner/ageLANServer/common/process"
	"github.com/luskaner/ageLANServer/launcher-common/executor"
	commonIpc "github.com/luskaner/ageLANServer/launcher-common/ipc"
	"github.com/luskaner/ageLANServer/launcher-config/internal"
)

var ipc net.Conn = nil
var encoder *gob.Encoder = nil
var decoder *gob.Decoder = nil

func RunSetUp(gameId string, logRoot string, ipToMap net.IP, macOsExclusiveMappings bool, addCertData []byte) (err error, exitCode int) {
	exitCode = common.ErrGeneral
	if ipc != nil {
		return runSetUpAgent(gameId, ipToMap, macOsExclusiveMappings, addCertData)
	}

	var certificate *x509.Certificate
	if addCertData != nil {
		certificate = common.BytesToCertificate(addCertData)
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
		result = executor.RunSetUp(gameId, ipToMap, macOsExclusiveMappings, certificate, file.Folder(), writer, func(options exec.Options) {
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

func RunFlushCache(logRoot string, ips bool, certs bool) (err error, exitCode int) {
	if ipc != nil {
		return fmt.Errorf("cannot flush cache if agent is already started"), internal.ErrAgentAlreadyStarted
	}
	var result *exec.Result
	var file *commonLogger.Root
	if logRoot != "" {
		if err, file = commonLogger.NewFile(logRoot, "", true); err != nil {
			exitCode = common.ErrFileLog
			return
		}
	}
	if bufferErr := file.Buffer("config-admin_flushCache", func(writer io.Writer) {
		_, result = executor.RunFlushCache(ips, certs, file.Folder(), writer, func(options exec.Options) {
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

func StopAgentIfNeeded() bool {
	agentConnected := ConnectAgentIfNeeded() == nil
	exeFileName := executables.NativeFileName(true, executables.LauncherConfigAdminAgent)
	if !agentConnected {
		if _, proc, err := commonProcess.Process(exeFileName); err == nil && proc == nil {
			return true
		}
	}
	var stoppedAgent bool
	commonLogger.Println("Trying to stop 'config-admin-agent'.")
	if err := stopAgentIfNeeded(); err == nil {
		if ConnectAgentIfNeededWithRetries(false) {
			commonLogger.Println("Stopped 'config-admin-agent'")
			stoppedAgent = true
		} else {
			commonLogger.Println("Failed to stop 'config-admin-agent'")
		}
	} else {
		commonLogger.Println("Failed to trying stopping 'config-admin-agent'")
		commonLogger.Println(err)
	}
	if !stoppedAgent {
		if pid, proc, err := commonProcess.Process(exeFileName); err == nil && proc != nil {
			if err = commonProcess.KillPidProc(pid, proc); err == nil {
				commonLogger.Println("Successfully killed 'config-admin-agent'.")
				stoppedAgent = true
			} else {
				commonLogger.Println("Failed to kill 'config-admin-agent'")
				commonLogger.Println(err)
			}
		}
	}
	return stoppedAgent
}

func stopAgentIfNeeded() (err error) {
	commonLogger.Println("Stopping agent")
	if ipc != nil {
		str := "-> Exit: "
		err = encoder.Encode(commonIpc.Exit)
		if err != nil {
			commonLogger.Println(str + "Could not encode")
			return
		}
		commonLogger.Println(str + "OK")
		clearIPCState()
	} else {
		commonLogger.Println("Already stopped")
	}
	return
}

func ConnectAgentIfNeededWithRetries(retryUntilSuccess bool) bool {
	var prePostFn func()
	if retryUntilSuccess {
		prePostFn = func() {}
	} else {
		prePostFn = clearIPCState
	}
	for range 30 {
		prePostFn()
		if (ConnectAgentIfNeeded() == nil) == retryUntilSuccess {
			return true
		}
		prePostFn()
		time.Sleep(100 * time.Millisecond)
	}
	return false
}

func clearIPCState() {
	if ipc != nil {
		_ = ipc.Close()
	}
	encoder = nil
	decoder = nil
	ipc = nil
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

func StartAgent(flushIPs bool, flushCerts bool) (result *exec.Result) {
	commonLogger.Println("Starting agent")
	var file string
	var logRoot string
	if internal.Logger != nil {
		logRoot = internal.Logger.Folder()
	}
	file, result = executor.RunFlushCacheAgent(flushIPs, flushCerts, logRoot, nil, func(options exec.Options) {
		commonLogger.Println("start config-admin-agent:", options.String())
	})
	if result.Success() {
		if !postAgentStart(result.Pid, file) {
			result.Err = fmt.Errorf("agent process failed to start")
		}
	}
	return
}

func sendAgent(commandType byte, commandName string, commandFn func() any) (err error, exitCode int) {
	str := fmt.Sprintf("-> %s: ", commandName)
	if err = encoder.Encode(commandType); err != nil {
		commonLogger.Println(str + "Could not encode")
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
	data := commandFn()
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

func runRevertAgent(unmapIPs bool, removeCert bool) (err error, exitCode int) {
	return sendAgent(
		commonIpc.Revert,
		"Revert",
		func() any {
			return commonIpc.RevertCommand{IPs: unmapIPs, Certificate: removeCert}
		},
	)
}

func runSetUpAgent(gameId string, mapIp net.IP, macOsExclusiveMappings bool, certificate []byte) (err error, exitCode int) {
	return sendAgent(
		commonIpc.Setup,
		"Setup",
		func() any {
			return commonIpc.SetupCommand{GameId: gameId, IP: mapIp, MacOsExclusiveMappings: macOsExclusiveMappings, Certificate: certificate}
		},
	)
}

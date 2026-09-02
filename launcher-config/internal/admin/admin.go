package admin

import (
	"crypto/x509"
	"encoding/gob"
	"fmt"
	"io"
	"net"
	"os"
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

// deps groups the external effect points used by the admin client so tests can
// inject fakes via NewAdmin instead of mutating package globals.
type deps struct {
	bytesToCertificate func([]byte) *x509.Certificate
	newFile            func(root string, gameId string, finalRoot bool) (error, *commonLogger.Root)
	runSetUp           func(gameId string, ip net.IP, macOsExclusiveMappings bool, certificate *x509.Certificate, logRoot string, out io.Writer, optionsFn func(*exec.Options)) *exec.Result
	runRevert          func(ips bool, certs bool, failfast bool, logRoot string, out io.Writer, optionsFn func(*exec.Options)) *exec.Result
	runFlushCache      func(ips bool, certs bool, logRoot string, out io.Writer, optionsFn func(*exec.Options)) (string, *exec.Result)
	runFlushCacheAgent func(ips bool, certs bool, logRoot string, out io.Writer, optionsFn func(*exec.Options)) (string, *exec.Result)
	process            func(name string) (string, *os.Process, error)
	killPidProc        func(pid string, proc *os.Process) error
	dialIPC            func() (net.Conn, error)
	postAgentStart     func(pid uint32, file string) bool
	nativeFileName     func(bin bool, name string) string
	sleep              func(time.Duration)
	getLoggerFolder    func() string
}

func defaultDeps() deps {
	return deps{
		bytesToCertificate: common.BytesToCertificate,
		newFile:            commonLogger.NewFile,
		runSetUp:           executor.RunSetUp,
		runRevert:          executor.RunRevert,
		runFlushCache:      executor.RunFlushCache,
		runFlushCacheAgent: executor.RunFlushCacheAgent,
		process:            commonProcess.Process,
		killPidProc:        commonProcess.KillPidProc,
		dialIPC:            DialIPC,
		postAgentStart:     postAgentStart,
		nativeFileName:     executables.NativeFileName,
		sleep:              time.Sleep,
		getLoggerFolder: func() string {
			if internal.Logger != nil {
				return internal.Logger.Folder()
			}
			return ""
		},
	}
}

// Admin is the launcher-config-admin client. It holds both its injectable
// dependencies and its IPC connection state, so tests can construct isolated
// instances without mutating package globals.
type Admin struct {
	deps deps
	ipc  net.Conn
	enc  *gob.Encoder
	dec  *gob.Decoder
}

// NewAdmin returns an Admin using the given deps. Prefer DefaultDeps() plus
// field overrides in tests.
func NewAdmin(d deps) *Admin {
	return &Admin{deps: d}
}

// DefaultDeps returns the production dependencies for an Admin.
func DefaultDeps() deps { return defaultDeps() }

// Default is the process-wide Admin used by the package-level convenience
// functions, mirroring the http.DefaultClient idiom.
var Default = NewAdmin(defaultDeps())

func (a *Admin) RunSetUp(gameId string, logRoot string, ipToMap net.IP, macOsExclusiveMappings bool, addCertData []byte) (err error, exitCode int) {
	exitCode = common.ErrGeneral
	if a.ipc != nil {
		return a.runSetUpAgent(gameId, ipToMap, macOsExclusiveMappings, addCertData)
	}

	var certificate *x509.Certificate
	if addCertData != nil {
		certificate = a.deps.bytesToCertificate(addCertData)
		if certificate == nil {
			exitCode = internal.ErrUserCertAddParse
			return
		}
	}
	var result *exec.Result
	var file *commonLogger.Root
	if logRoot != "" {
		if err, file = a.deps.newFile(logRoot, "", true); err != nil {
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
		result = a.deps.runSetUp(gameId, ipToMap, macOsExclusiveMappings, certificate, file.Folder(), writer, func(options *exec.Options) {
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

func (a *Admin) RunRevert(logRoot string, unmapIPs bool, removeCert bool, failfast bool) (err error, exitCode int) {
	if a.ipc != nil {
		return a.runRevertAgent(unmapIPs, removeCert)
	}
	var result *exec.Result
	var file *commonLogger.Root
	if logRoot != "" {
		if err, file = a.deps.newFile(logRoot, "", true); err != nil {
			exitCode = common.ErrFileLog
			return
		}
	}
	if bufferErr := file.Buffer("config-admin_revert", func(writer io.Writer) {
		result = a.deps.runRevert(unmapIPs, removeCert, failfast, file.Folder(), writer, func(options *exec.Options) {
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

func (a *Admin) RunFlushCache(logRoot string, ips bool, certs bool) (err error, exitCode int) {
	if a.ipc != nil {
		return fmt.Errorf("cannot flush cache if agent is already started"), internal.ErrAgentAlreadyStarted
	}
	var result *exec.Result
	var file *commonLogger.Root
	if logRoot != "" {
		if err, file = a.deps.newFile(logRoot, "", true); err != nil {
			exitCode = common.ErrFileLog
			return
		}
	}
	if bufferErr := file.Buffer("config-admin_flushCache", func(writer io.Writer) {
		_, result = a.deps.runFlushCache(ips, certs, file.Folder(), writer, func(options *exec.Options) {
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

func (a *Admin) StopAgentIfNeeded() bool {
	agentConnected := a.ConnectAgentIfNeeded() == nil
	exeFileName := a.deps.nativeFileName(true, executables.LauncherConfigAdminAgent)
	if !agentConnected {
		if _, proc, err := a.deps.process(exeFileName); err == nil && proc == nil {
			return true
		}
	}
	commonLogger.Println("Trying to stop 'config-admin-agent'.")
	if err := a.stopAgentIfNeeded(); err == nil {
		for range 30 {
			if _, proc, err := a.deps.process(exeFileName); err == nil && proc == nil {
				commonLogger.Println("Stopped 'config-admin-agent'")
				return true
			}
			a.deps.sleep(100 * time.Millisecond)
		}
		commonLogger.Println("Failed to stop 'config-admin-agent'")
	} else {
		commonLogger.Println("Failed to trying stopping 'config-admin-agent'")
		commonLogger.Println(err)
	}
	if pid, proc, err := a.deps.process(exeFileName); err == nil && proc != nil {
		if err = a.deps.killPidProc(pid, proc); err == nil {
			commonLogger.Println("Successfully killed 'config-admin-agent'.")
			return true
		}
		commonLogger.Println("Failed to kill 'config-admin-agent'")
		commonLogger.Println(err)
	}
	return false
}

func (a *Admin) stopAgentIfNeeded() (err error) {
	commonLogger.Println("Stopping agent")
	if a.ipc != nil {
		str := "-> Exit: "
		err = a.enc.Encode(commonIpc.Exit)
		if err != nil {
			commonLogger.Println(str + "Could not encode")
			return
		}
		commonLogger.Println(str + "OK")
		a.clearIPCState()
	} else {
		commonLogger.Println("Already stopped")
	}
	return
}

func (a *Admin) ConnectAgentIfNeededWithRetries() bool {
	for range 30 {
		if a.ConnectAgentIfNeeded() == nil {
			return true
		}
		a.deps.sleep(100 * time.Millisecond)
	}
	return false
}

func (a *Admin) clearIPCState() {
	if a.ipc != nil {
		_ = a.ipc.Close()
	}
	a.enc = nil
	a.dec = nil
	a.ipc = nil
}

func (a *Admin) ConnectAgentIfNeeded() (err error) {
	commonLogger.Println("Connecting to agent")
	if a.ipc != nil {
		commonLogger.Println("Already connected")
		return
	}
	var conn net.Conn
	conn, err = a.deps.dialIPC()
	if err != nil {
		return
	}
	commonLogger.Println("Connected")
	a.ipc = conn
	a.enc = gob.NewEncoder(a.ipc)
	a.dec = gob.NewDecoder(a.ipc)
	return
}

func (a *Admin) StartAgent(flushIPs bool, flushCerts bool) (result *exec.Result) {
	commonLogger.Println("Starting agent")
	var file string
	logRoot := a.deps.getLoggerFolder()
	file, result = a.deps.runFlushCacheAgent(flushIPs, flushCerts, logRoot, nil, func(options *exec.Options) {
		commonLogger.Println("start config-admin-agent:", options.String())
	})
	if result.Success() {
		if !a.deps.postAgentStart(result.Pid, file) {
			result.Err = fmt.Errorf("agent process failed to start")
		}
	}
	return
}

func (a *Admin) sendAgent(commandType byte, commandName string, commandFn func() any) (err error, exitCode int) {
	str := fmt.Sprintf("-> %s: ", commandName)
	if err = a.enc.Encode(commandType); err != nil {
		commonLogger.Println(str + "Could not encode")
		return
	}
	commonLogger.Println(str + "OK")
	str = "<- Exit Code: "
	if err = a.dec.Decode(&exitCode); err != nil || exitCode != common.ErrSuccess {
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
	if err = a.enc.Encode(data); err != nil {
		commonLogger.Println(str + "Could not encode")
		return
	}
	commonLogger.Println(str + "OK")
	str = "<- Exit Code: "
	if err = a.dec.Decode(&exitCode); err != nil {
		commonLogger.Println(str + "Could not decode")
		return
	}
	commonLogger.Println(str + strconv.Itoa(exitCode))
	return
}

func (a *Admin) runRevertAgent(unmapIPs bool, removeCert bool) (err error, exitCode int) {
	return a.sendAgent(
		commonIpc.Revert,
		"Revert",
		func() any {
			return commonIpc.RevertCommand{IPs: unmapIPs, Certificate: removeCert}
		},
	)
}

func (a *Admin) runSetUpAgent(gameId string, mapIp net.IP, macOsExclusiveMappings bool, certificate []byte) (err error, exitCode int) {
	return a.sendAgent(
		commonIpc.Setup,
		"Setup",
		func() any {
			return commonIpc.SetupCommand{GameId: gameId, IP: mapIp, MacOsExclusiveMappings: macOsExclusiveMappings, Certificate: certificate}
		},
	)
}

// Package-level wrappers for backward compatibility. They delegate to Default.

func RunSetUp(gameId string, logRoot string, ipToMap net.IP, macOsExclusiveMappings bool, addCertData []byte) (err error, exitCode int) {
	return Default.RunSetUp(gameId, logRoot, ipToMap, macOsExclusiveMappings, addCertData)
}

func RunRevert(logRoot string, unmapIPs bool, removeCert bool, failfast bool) (err error, exitCode int) {
	return Default.RunRevert(logRoot, unmapIPs, removeCert, failfast)
}

func RunFlushCache(logRoot string, ips bool, certs bool) (err error, exitCode int) {
	return Default.RunFlushCache(logRoot, ips, certs)
}

func StopAgentIfNeeded() bool {
	return Default.StopAgentIfNeeded()
}

func ConnectAgentIfNeededWithRetries() bool {
	return Default.ConnectAgentIfNeededWithRetries()
}

func ConnectAgentIfNeeded() (err error) {
	return Default.ConnectAgentIfNeeded()
}

func StartAgent(flushIPs bool, flushCerts bool) (result *exec.Result) {
	return Default.StartAgent(flushIPs, flushCerts)
}

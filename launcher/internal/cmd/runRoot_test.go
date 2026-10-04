package cmd

import (
	"bytes"
	"crypto/x509"
	"errors"
	"io"
	"net"
	"net/netip"
	"os"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/spf13/pflag"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/cmd/bsManager"
	cmdServer "github.com/luskaner/ageLANServer/common/cmd/server"
	"github.com/luskaner/ageLANServer/common/executables"
	commonExecutor "github.com/luskaner/ageLANServer/common/executor/exec"
	"github.com/luskaner/ageLANServer/common/fileLock"
	"github.com/luskaner/ageLANServer/common/game/executor/base"
	"github.com/luskaner/ageLANServer/common/game/executor/custom"
	commonServer "github.com/luskaner/ageLANServer/common/server"
	"github.com/luskaner/ageLANServer/common/uuid"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/luskaner/ageLANServer/launcher-common/launcher"
	"github.com/luskaner/ageLANServer/launcher-common/launcher/executor"
	"github.com/luskaner/ageLANServer/launcher-common/launcher/server"
	"github.com/luskaner/ageLANServer/launcher/internal/dialog"
)

type fakeFileInfo struct{ isDir bool }

func (f fakeFileInfo) Name() string       { return "fake" }
func (f fakeFileInfo) Size() int64        { return 0 }
func (f fakeFileInfo) Mode() os.FileMode  { return 0 }
func (f fakeFileInfo) ModTime() time.Time { return time.Time{} }
func (f fakeFileInfo) IsDir() bool        { return f.isDir }
func (f fakeFileInfo) Sys() any           { return nil }

type fakeExecutor struct {
	path string
}

func (f fakeExecutor) Do(args []string, optionsFn func(commonExecutor.Options)) *commonExecutor.Result {
	return &commonExecutor.Result{}
}
func (f fakeExecutor) GameProcesses() (bool, bool, bool) { return false, false, false }
func (f fakeExecutor) String() string                    { return "fake" }
func (f fakeExecutor) Path() string {
	if f.path != "" {
		return f.path
	}
	return "fakePath"
}

type fakePidLocker struct {
	lockErr   error
	unlockErr error
}

func (f *fakePidLocker) Lock() error   { return f.lockErr }
func (f *fakePidLocker) Unlock() error { return f.unlockErr }

var _ fileLock.Locker = (*fakePidLocker)(nil)

// fakeDialog is the backend runRoot tests run against: no test may open a real
// graphical window, and none may block reading stdin, which is what the real
// backends do while waiting for an answer.
type fakeDialog struct {
	confirmCalls int
	confirm      bool
}

func (f *fakeDialog) Name() string { return "fake" }

func (f *fakeDialog) SelectServer([]dialog.ServerCandidate, io.Reader) (int, bool) {
	return 0, true
}

func (f *fakeDialog) ListCandidates([]dialog.ServerCandidate) {}

func (f *fakeDialog) ConfirmStartServer(string, io.Reader) bool {
	f.confirmCalls++
	return f.confirm
}

// captureStdout redirects os.Stdout for the duration of fn and returns what was
// written to it.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	orig := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = orig }()

	done := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.String()
	}()

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("close pipe: %v", err)
	}
	return <-done
}

func TestRunRootInvalidGame(t *testing.T) {
	oldGameId, oldCfgFile, oldGameCfgFile := gameId, cfgFile, gameCfgFile
	defer func() { gameId, cfgFile, gameCfgFile = oldGameId, oldCfgFile, oldGameCfgFile }()

	gameId = ""
	cfgFile = ""
	gameCfgFile = ""

	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	_, exitCode := runRoot(fs)
	if exitCode != common.ErrSyntax {
		t.Errorf("expected exit code %d for missing gameId, got %d", common.ErrSyntax, exitCode)
	}
}

func TestRunRootPidLockError(t *testing.T) {
	origGameId, origCfgFile, origGameCfgFile := gameId, cfgFile, gameCfgFile
	origNewPidLock := newPidLockFn
	defer func() {
		gameId, cfgFile, gameCfgFile = origGameId, origCfgFile, origGameCfgFile
		newPidLockFn = origNewPidLock
	}()

	gameId = "aoe2"
	cfgFile = ""
	gameCfgFile = ""
	newPidLockFn = func() fileLock.Locker { return &fakePidLocker{lockErr: errors.New("pid lock")} }

	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	_, exitCode := runRoot(fs)
	if exitCode != common.ErrPidLock {
		t.Errorf("expected exit code %d on pid lock error, got %d", common.ErrPidLock, exitCode)
	}
}

func validLauncherConfig() *launcher.Configuration {
	c := &launcher.Configuration{}
	c.Config.Certificate.CanTrustInPc = "local"
	c.Config.CanBroadcastBattleServer = "auto"
	c.Config.Dialog = "auto"
	c.Server.Start = "auto"
	c.Server.Stop = "auto"
	c.Server.BattleServerManager.Run = "true"
	c.Client.Isolation.Metadata = "required"
	c.Client.Isolation.Profiles = "required"
	c.Server.Executable.Path = "auto"
	c.Server.Executable.Args = []string{"-e", "{Game}", "--id", "{Id}"}
	c.Server.BattleServerManager.Executable.Args = []string{"-e", "{Game}", "-r"}
	c.Server.BattleServerManager.Executable.Path = "auto"
	c.Config.SetupCommand = []string{}
	c.Config.RevertCommand = []string{}
	c.Client.Executable.Path = "auto"
	c.Client.Isolation.Path = "auto"
	return c
}

func TestRunRootUnsupportedGame(t *testing.T) {
	origGameId, origCfgFile, origGameCfgFile := gameId, cfgFile, gameCfgFile
	origNewPidLock := newPidLockFn
	origInitConfig := initConfigFn
	origOpenMainLog := openMainLogFn
	origIsAdmin := isAdminFn
	origGameSupported := gameSupportedGamesContainsOneFn
	defer func() {
		gameId, cfgFile, gameCfgFile = origGameId, origCfgFile, origGameCfgFile
		newPidLockFn = origNewPidLock
		initConfigFn = origInitConfig
		openMainLogFn = origOpenMainLog
		isAdminFn = origIsAdmin
		gameSupportedGamesContainsOneFn = origGameSupported
	}()

	gameId = "aoe2"
	cfgFile = ""
	gameCfgFile = ""
	newPidLockFn = func() fileLock.Locker { return &fakePidLocker{} }
	initConfigFn = func(fs *pflag.FlagSet) *launcher.Configuration {
		return validLauncherConfig()
	}
	openMainLogFn = func(gameID string) error { return nil }
	isAdminFn = func() bool { return false }
	gameSupportedGamesContainsOneFn = func(id string) bool { return false }

	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	_, exitCode := runRoot(fs)
	if exitCode != launcherCommon.ErrInvalidGame {
		t.Errorf("expected exit code %d for unsupported game, got %d", launcherCommon.ErrInvalidGame, exitCode)
	}
}

func TestRunRootValidationFailures(t *testing.T) {
	tests := []struct {
		name     string
		mutate   func(c *launcher.Configuration)
		wantCode int
	}{
		{"invalid canTrustCertificate", func(c *launcher.Configuration) { c.Config.Certificate.CanTrustInPc = "bad" }, launcher.ErrInvalidCanTrustCertificate},
		{"invalid serverStart", func(c *launcher.Configuration) { c.Server.Start = "bad" }, launcher.ErrInvalidServerStart},
		{"invalid serverStop", func(c *launcher.Configuration) { c.Server.Stop = "bad" }, launcher.ErrInvalidServerStop},
		{"invalid battleServerManagerRun", func(c *launcher.Configuration) { c.Server.BattleServerManager.Run = "bad" }, launcher.ErrInvalidServerBattleServerManagerRun},
		{"invalid isolateMetadata", func(c *launcher.Configuration) { c.Client.Isolation.Metadata = "bad" }, launcher.ErrInvalidIsolateMetadata},
		{"invalid isolateProfiles", func(c *launcher.Configuration) { c.Client.Isolation.Profiles = "bad" }, launcher.ErrInvalidIsolateProfiles},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origGameId, origCfgFile, origGameCfgFile := gameId, cfgFile, gameCfgFile
			origNewPidLock := newPidLockFn
			origInitConfig := initConfigFn
			origOpenMainLog := openMainLogFn
			origIsAdmin := isAdminFn
			origGameSupported := gameSupportedGamesContainsOneFn
			defer func() {
				gameId, cfgFile, gameCfgFile = origGameId, origCfgFile, origGameCfgFile
				newPidLockFn = origNewPidLock
				initConfigFn = origInitConfig
				openMainLogFn = origOpenMainLog
				isAdminFn = origIsAdmin
				gameSupportedGamesContainsOneFn = origGameSupported
			}()
			gameId = "aoe2"
			cfgFile = ""
			gameCfgFile = ""
			newPidLockFn = func() fileLock.Locker { return &fakePidLocker{} }
			initConfigFn = func(fs *pflag.FlagSet) *launcher.Configuration {
				c := validLauncherConfig()
				tt.mutate(c)
				return c
			}
			openMainLogFn = func(gameID string) error { return nil }
			isAdminFn = func() bool { return false }
			gameSupportedGamesContainsOneFn = func(id string) bool { return true }
			fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
			_, exitCode := runRoot(fs)
			if exitCode != tt.wantCode {
				t.Errorf("%s: expected exit code %d, got %d", tt.name, tt.wantCode, exitCode)
			}
		})
	}
}

func TestRunRootOpenFileLogError(t *testing.T) {
	origGameId, origCfgFile, origGameCfgFile := gameId, cfgFile, gameCfgFile
	origNewPidLock := newPidLockFn
	origInitConfig := initConfigFn
	origOpenMainLog := openMainLogFn
	defer func() {
		gameId, cfgFile, gameCfgFile = origGameId, origCfgFile, origGameCfgFile
		newPidLockFn = origNewPidLock
		initConfigFn = origInitConfig
		openMainLogFn = origOpenMainLog
	}()
	gameId = "aoe2"
	cfgFile = ""
	gameCfgFile = ""
	newPidLockFn = func() fileLock.Locker { return &fakePidLocker{} }
	initConfigFn = func(fs *pflag.FlagSet) *launcher.Configuration {
		return validLauncherConfig()
	}
	openMainLogFn = func(gameID string) error { return errors.New("open fail") }
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	_, exitCode := runRoot(fs)
	if exitCode != common.ErrFileLog {
		t.Errorf("expected exit code %d for file log error, got %d", common.ErrFileLog, exitCode)
	}
}

func TestRunRootServerArgsParseFailure(t *testing.T) {
	origGameId, origCfgFile, origGameCfgFile := gameId, cfgFile, gameCfgFile
	origNewPidLock := newPidLockFn
	origInitConfig := initConfigFn
	origOpenMainLog := openMainLogFn
	origIsAdmin := isAdminFn
	origGameSupported := gameSupportedGamesContainsOneFn
	origParseArgs := parseCommandArgsFn
	defer func() {
		gameId, cfgFile, gameCfgFile = origGameId, origCfgFile, origGameCfgFile
		newPidLockFn = origNewPidLock
		initConfigFn = origInitConfig
		openMainLogFn = origOpenMainLog
		isAdminFn = origIsAdmin
		gameSupportedGamesContainsOneFn = origGameSupported
		parseCommandArgsFn = origParseArgs
	}()
	gameId = "aoe2"
	cfgFile = ""
	gameCfgFile = ""
	newPidLockFn = func() fileLock.Locker { return &fakePidLocker{} }
	initConfigFn = func(fs *pflag.FlagSet) *launcher.Configuration {
		c := validLauncherConfig()
		c.Server.Args = []string{"bad-args"}
		return c
	}
	openMainLogFn = func(gameID string) error { return nil }
	isAdminFn = func() bool { return false }
	gameSupportedGamesContainsOneFn = func(id string) bool { return true }
	parseCommandArgsFn = func(args []string, values map[string]string) ([]string, error) {
		return nil, errors.New("parse fail")
	}
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	_, exitCode := runRoot(fs)
	if exitCode != launcher.ErrInvalidServerArgs {
		t.Errorf("expected exit code %d for server args parse fail, got %d", launcher.ErrInvalidServerArgs, exitCode)
	}
}

func TestRunRootSetupCommandParseFailure(t *testing.T) {
	origGameId, origCfgFile, origGameCfgFile := gameId, cfgFile, gameCfgFile
	origNewPidLock := newPidLockFn
	origInitConfig := initConfigFn
	origOpenMainLog := openMainLogFn
	origIsAdmin := isAdminFn
	origGameSupported := gameSupportedGamesContainsOneFn
	origParseArgs := parseCommandArgsFn
	defer func() {
		gameId, cfgFile, gameCfgFile = origGameId, origCfgFile, origGameCfgFile
		newPidLockFn = origNewPidLock
		initConfigFn = origInitConfig
		openMainLogFn = origOpenMainLog
		isAdminFn = origIsAdmin
		gameSupportedGamesContainsOneFn = origGameSupported
		parseCommandArgsFn = origParseArgs
	}()
	gameId = "aoe2"
	cfgFile = ""
	gameCfgFile = ""
	newPidLockFn = func() fileLock.Locker { return &fakePidLocker{} }
	initConfigFn = func(fs *pflag.FlagSet) *launcher.Configuration {
		c := validLauncherConfig()
		c.Config.SetupCommand = []string{"setup", "bad"}
		return c
	}
	openMainLogFn = func(gameID string) error { return nil }
	isAdminFn = func() bool { return false }
	gameSupportedGamesContainsOneFn = func(id string) bool { return true }
	// First call for server args should succeed, second for battleServerManager should succeed, third for setup should fail
	callCount := 0
	origFn := parseCommandArgsFn
	parseCommandArgsFn = func(args []string, values map[string]string) ([]string, error) {
		callCount++
		if callCount == 3 {
			return nil, errors.New("setup parse fail")
		}
		return origFn(args, values)
	}
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	_, exitCode := runRoot(fs)
	if exitCode != launcher.ErrInvalidSetupCommand {
		t.Errorf("expected exit code %d for setup command parse fail, got %d", launcher.ErrInvalidSetupCommand, exitCode)
	}
}

func TestRunRootInvalidIsolationPath(t *testing.T) {
	origGameId, origCfgFile, origGameCfgFile := gameId, cfgFile, gameCfgFile
	origNewPidLock := newPidLockFn
	origInitConfig := initConfigFn
	origOpenMainLog := openMainLogFn
	origIsAdmin := isAdminFn
	origGameSupported := gameSupportedGamesContainsOneFn
	origParsePath := commonParsePathFn
	defer func() {
		gameId, cfgFile, gameCfgFile = origGameId, origCfgFile, origGameCfgFile
		newPidLockFn = origNewPidLock
		initConfigFn = origInitConfig
		openMainLogFn = origOpenMainLog
		isAdminFn = origIsAdmin
		gameSupportedGamesContainsOneFn = origGameSupported
		commonParsePathFn = origParsePath
	}()
	gameId = "aoe2"
	newPidLockFn = func() fileLock.Locker { return &fakePidLocker{} }
	initConfigFn = func(fs *pflag.FlagSet) *launcher.Configuration {
		c := validLauncherConfig()
		c.Client.Isolation.Path = "custom/path"
		return c
	}
	openMainLogFn = func(gameID string) error { return nil }
	isAdminFn = func() bool { return false }
	gameSupportedGamesContainsOneFn = func(id string) bool { return true }
	commonParsePathFn = func(slice []string, _ map[string]string) (os.FileInfo, string, error) {
		return nil, "", errors.New("bad path")
	}
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	_, exitCode := runRoot(fs)
	if exitCode != launcher.ErrInvalidIsolationPath {
		t.Errorf("expected %d for invalid isolation path, got %d", launcher.ErrInvalidIsolationPath, exitCode)
	}
}

func TestRunRootInvalidServerExecutable(t *testing.T) {
	origGameId, origCfgFile, origGameCfgFile := gameId, cfgFile, gameCfgFile
	origNewPidLock := newPidLockFn
	origInitConfig := initConfigFn
	origOpenMainLog := openMainLogFn
	origIsAdmin := isAdminFn
	origGameSupported := gameSupportedGamesContainsOneFn
	origParsePath := commonParsePathFn
	defer func() {
		gameId, cfgFile, gameCfgFile = origGameId, origCfgFile, origGameCfgFile
		newPidLockFn = origNewPidLock
		initConfigFn = origInitConfig
		openMainLogFn = origOpenMainLog
		isAdminFn = origIsAdmin
		gameSupportedGamesContainsOneFn = origGameSupported
		commonParsePathFn = origParsePath
	}()
	gameId = "aoe2"
	newPidLockFn = func() fileLock.Locker { return &fakePidLocker{} }
	initConfigFn = func(fs *pflag.FlagSet) *launcher.Configuration {
		c := validLauncherConfig()
		c.Server.Executable.Path = "bad/server.exe"
		return c
	}
	openMainLogFn = func(gameID string) error { return nil }
	isAdminFn = func() bool { return false }
	gameSupportedGamesContainsOneFn = func(id string) bool { return true }
	commonParsePathFn = func(slice []string, _ map[string]string) (os.FileInfo, string, error) {
		// isolation path success (auto -> not called), server path fails
		if len(slice) > 0 && slice[0] == "bad/server.exe" {
			return nil, "", errors.New("bad")
		}
		return fakeFileInfo{isDir: false}, "ok", nil
	}
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	_, exitCode := runRoot(fs)
	if exitCode != launcher.ErrInvalidServerPath {
		t.Errorf("expected %d for invalid server path, got %d", launcher.ErrInvalidServerPath, exitCode)
	}
}

func TestRunRootGameLauncherNotFound(t *testing.T) {
	origGameId, origCfgFile, origGameCfgFile := gameId, cfgFile, gameCfgFile
	origNewPidLock := newPidLockFn
	origInitConfig := initConfigFn
	origOpenMainLog := openMainLogFn
	origIsAdmin := isAdminFn
	origGameSupported := gameSupportedGamesContainsOneFn
	origMakeExec := makeExecFn
	defer func() {
		gameId, cfgFile, gameCfgFile = origGameId, origCfgFile, origGameCfgFile
		newPidLockFn = origNewPidLock
		initConfigFn = origInitConfig
		openMainLogFn = origOpenMainLog
		isAdminFn = origIsAdmin
		gameSupportedGamesContainsOneFn = origGameSupported
		makeExecFn = origMakeExec
	}()
	gameId = "aoe2"
	newPidLockFn = func() fileLock.Locker { return &fakePidLocker{} }
	initConfigFn = func(fs *pflag.FlagSet) *launcher.Configuration {
		return validLauncherConfig()
	}
	openMainLogFn = func(gameID string) error { return nil }
	isAdminFn = func() bool { return false }
	gameSupportedGamesContainsOneFn = func(id string) bool { return true }
	makeExecFn = func(gameId, clientExecutable string) base.Executor {
		return nil
	}
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	_, exitCode := runRoot(fs)
	if exitCode != launcher.ErrGameLauncherNotFound {
		t.Errorf("expected %d for game launcher not found, got %d", launcher.ErrGameLauncherNotFound, exitCode)
	}
}

func TestRunRootGameAlreadyRunning(t *testing.T) {
	origGameId, origCfgFile, origGameCfgFile := gameId, cfgFile, gameCfgFile
	origNewPidLock := newPidLockFn
	origInitConfig := initConfigFn
	origOpenMainLog := openMainLogFn
	origIsAdmin := isAdminFn
	origGameSupported := gameSupportedGamesContainsOneFn
	origMakeExec := makeExecFn
	origIsolationPath := configIsolationPathFn
	origGameRunning := gameRunningFn
	origProcess := commonProcessProcessFn
	defer func() {
		gameId, cfgFile, gameCfgFile = origGameId, origCfgFile, origGameCfgFile
		newPidLockFn = origNewPidLock
		initConfigFn = origInitConfig
		openMainLogFn = origOpenMainLog
		isAdminFn = origIsAdmin
		gameSupportedGamesContainsOneFn = origGameSupported
		makeExecFn = origMakeExec
		configIsolationPathFn = origIsolationPath
		gameRunningFn = origGameRunning
		commonProcessProcessFn = origProcess
	}()
	gameId = "age1"
	newPidLockFn = func() fileLock.Locker { return &fakePidLocker{} }
	initConfigFn = func(fs *pflag.FlagSet) *launcher.Configuration {
		c := validLauncherConfig()
		c.Client.Executable.Path = "auto"
		return c
	}
	openMainLogFn = func(gameID string) error { return nil }
	isAdminFn = func() bool { return false }
	gameSupportedGamesContainsOneFn = func(id string) bool { return true }
	makeExecFn = func(gameId, clientExecutable string) base.Executor {
		return fakeExecutor{}
	}
	configIsolationPathFn = func(exec base.Executor) string { return t.TempDir() }
	commonProcessProcessFn = func(s string) (string, *os.Process, error) { return "", nil, nil }
	gameRunningFn = func() bool { return true }
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	_, exitCode := runRoot(fs)
	if exitCode != launcher.ErrGameAlreadyRunning {
		t.Errorf("expected %d for game already running, got %d", launcher.ErrGameAlreadyRunning, exitCode)
	}
}

func TestRunRootConfigRevertBufferError(t *testing.T) {
	origGameId, origCfgFile, origGameCfgFile := gameId, cfgFile, gameCfgFile
	origNewPidLock := newPidLockFn
	origInitConfig := initConfigFn
	origOpenMainLog := openMainLogFn
	origIsAdmin := isAdminFn
	origGameSupported := gameSupportedGamesContainsOneFn
	origMakeExec := makeExecFn
	origIsolationPath := configIsolationPathFn
	origGameRunning := gameRunningFn
	origProcess := commonProcessProcessFn
	origBuffer := commonLoggerFileLoggerBufferFn
	origKillAgent := configKillAgentFn
	defer func() {
		gameId, cfgFile, gameCfgFile = origGameId, origCfgFile, origGameCfgFile
		newPidLockFn = origNewPidLock
		initConfigFn = origInitConfig
		openMainLogFn = origOpenMainLog
		isAdminFn = origIsAdmin
		gameSupportedGamesContainsOneFn = origGameSupported
		makeExecFn = origMakeExec
		configIsolationPathFn = origIsolationPath
		gameRunningFn = origGameRunning
		commonProcessProcessFn = origProcess
		commonLoggerFileLoggerBufferFn = origBuffer
		configKillAgentFn = origKillAgent
	}()
	gameId = "age1"
	newPidLockFn = func() fileLock.Locker { return &fakePidLocker{} }
	initConfigFn = func(fs *pflag.FlagSet) *launcher.Configuration {
		return validLauncherConfig()
	}
	openMainLogFn = func(gameID string) error { return nil }
	isAdminFn = func() bool { return false }
	gameSupportedGamesContainsOneFn = func(id string) bool { return true }
	makeExecFn = func(gameId, clientExecutable string) base.Executor { return fakeExecutor{} }
	configIsolationPathFn = func(exec base.Executor) string { return t.TempDir() }
	commonProcessProcessFn = func(s string) (string, *os.Process, error) { return "", nil, nil }
	gameRunningFn = func() bool { return false }
	configKillAgentFn = func() {}
	commonLoggerFileLoggerBufferFn = func(name string, fn func(io.Writer)) error { return errors.New("buffer fail") }
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	_, exitCode := runRoot(fs)
	if exitCode != common.ErrFileLog {
		t.Errorf("expected %d for buffer error, got %d", common.ErrFileLog, exitCode)
	}
}

type runRootOverrides struct {
	gameId                            string
	cfg                               func() *launcher.Configuration
	isAdmin                           bool
	gameSupported                     bool
	makeExec                          base.Executor
	isolationPath                     string
	processFn                         func(string) (string, *os.Process, error)
	gameRunning                       bool
	openLog                           error
	bufferFn                          func(string, func(io.Writer)) error
	killAgent                         func()
	newPidLock                        func() fileLock.Locker
	parseCommandArgsFn                func([]string, map[string]string) ([]string, error)
	resolveIsolateValueFnVal          func(string, bool) bool
	configSetGameIdFnVal              func(string)
	configIsolationPathFnVal          func(base.Executor) string
	commonParsePathFnVal              func([]string, map[string]string) (os.FileInfo, string, error)
	commonEnhancedViperFnVal          func(string) []string
	configNativeMacOsGameFnVal        func(base.Executor, bool) bool
	configBattleServerRequiredFnVal   func(base.Executor) bool
	newConfigFlushCacheOptionsFnVal   func(bool, string, bool, bool) *executor.ConfigFlushCacheOptions
	discoverServersFnVal              func(string, bool, mapset.Set[netip.Addr], mapset.Set[uint16]) (uuid.UUID, net.IP)
	netipParseAddrFnVal               func(string) (netip.Addr, error)
	serverFilterServerIPsFnVal        func(uuid.UUID, string, string, mapset.Set[netip.Addr]) (uuid.UUID, []server.MesuredIpAddress, *commonServer.AnnounceMessageDataSupportedLatest)
	serverGetExecutablePathFnVal      func(string) string
	serverGenerateCertsFnVal          func(string, bool) int
	configRunBattleServerManagerFnVal func(string, *pflag.FlagSet, *bsManager.StartValues, bool) int
	bsManagerStartFlagSetFnVal        func([]string) (*bsManager.StartValues, *pflag.FlagSet)
	configStartServerFnVal            func(string, *pflag.FlagSet, *cmdServer.Values, bool) (int, string)
	serverReadCACertFnVal             func(string) *x509.Certificate
	configMapHostsFnVal               func(string, string, bool, bool, bool) int
	configAddCertFnVal                func(string, uuid.UUID, *x509.Certificate, string, bool, bool) int
	configIsolateUserDataFnVal        func(bool, bool, string) int
	configAddCACertToGameFnVal        func(string, uuid.UUID, *x509.Certificate, string, string, bool, bool) int
	configLaunchAgentAndGameFnVal     func(base.Executor, custom.Exec, []string, string, string, string) int
	uuidParseFnVal                    func(string) (uuid.UUID, error)
	uuidMustParseFnVal                func(string) uuid.UUID
	uuidNilFnVal                      func() uuid.UUID
	executablesNativeFileNameFnVal    func(bool, string) string
	configRunSetupCommandFnVal        func([]string) *commonExecutor.Result
	dnsConnectivityFnVal              func() bool
	configRunStopAgentFnVal           func() *commonExecutor.Result
	waitForProcessFnVal               func(*os.Process, *time.Duration) bool
	dialogNewFnVal                    func(string) dialog.Resolution
}

func applyOverrides(t *testing.T, o runRootOverrides) func() {
	t.Helper()
	origGameId, origCfgFile, origGameCfgFile := gameId, cfgFile, gameCfgFile
	origNewPidLock := newPidLockFn
	origInitConfig := initConfigFn
	origOpenMainLog := openMainLogFn
	origIsAdmin := isAdminFn
	origGameSupported := gameSupportedGamesContainsOneFn
	origMakeExec := makeExecFn
	origIsolationPath := configIsolationPathFn
	origGameRunning := gameRunningFn
	origProcess := commonProcessProcessFn
	origBuffer := commonLoggerFileLoggerBufferFn
	origKillAgent := configKillAgentFn
	origParseCommandArgs := parseCommandArgsFn
	origResolveIsolate := resolveIsolateValueFn
	origConfigSetGameId := configSetGameIdFn
	origCommonParsePath := commonParsePathFn
	origCommonEnhancedViper := commonEnhancedViperFn
	origConfigNativeMacOsGame := configNativeMacOsGameFn
	origConfigBattleServerRequired := configBattleServerRequiredFn
	origNewConfigFlushCacheOptions := newConfigFlushCacheOptionsFn
	origDiscoverServers := discoverServersFn
	origNetipParseAddr := netipParseAddrFn
	origServerFilterServerIPs := serverFilterServerIPsFn
	origServerGetExecutablePath := serverGetExecutablePathFn
	origServerGenerateCerts := serverGenerateCertsFn
	origConfigRunBattleServerManager := configRunBattleServerManagerFn
	origBsManagerStartFlagSet := bsManagerStartFlagSetFn
	origConfigStartServer := configStartServerFn
	origServerReadCACert := serverReadCACertFn
	origConfigMapHosts := configMapHostsFn
	origConfigAddCert := configAddCertFn
	origConfigIsolateUserData := configIsolateUserDataFn
	origConfigAddCACertToGame := configAddCACertToGameFn
	origConfigLaunchAgentAndGame := configLaunchAgentAndGameFn
	origUuidParse := uuidParseFn
	origUuidMustParse := uuidMustParseFn
	origUuidNil := uuidNilFn
	origExecutablesNativeFileName := executablesNativeFileNameFn
	origConfigRunSetupCommand := configRunSetupCommandFn
	origDNSConnectivity := dnsConnectivityFn
	origRunStopAgent := configRunStopAgentFn
	origWaitForProcess := commonProcessWaitForProcessFn
	origDialogNew := dialogNewFn
	t.Cleanup(dialog.Reset)

	gameId = o.gameId
	cfgFile = ""
	gameCfgFile = ""
	if o.newPidLock != nil {
		newPidLockFn = o.newPidLock
	} else {
		newPidLockFn = func() fileLock.Locker { return &fakePidLocker{} }
	}
	if o.cfg != nil {
		initConfigFn = func(fs *pflag.FlagSet) *launcher.Configuration { return o.cfg() }
	} else {
		initConfigFn = func(fs *pflag.FlagSet) *launcher.Configuration { return validLauncherConfig() }
	}
	if o.openLog != nil {
		openMainLogFn = func(gameID string) error { return o.openLog }
	} else {
		openMainLogFn = func(gameID string) error { return nil }
	}
	isAdminFn = func() bool { return o.isAdmin }
	gameSupportedGamesContainsOneFn = func(id string) bool { return o.gameSupported }
	if o.makeExec != nil {
		makeExecFn = func(g, c string) base.Executor { return o.makeExec }
	} else {
		makeExecFn = func(g, c string) base.Executor { return fakeExecutor{} }
	}
	if o.isolationPath != "" {
		configIsolationPathFn = func(exec base.Executor) string { return o.isolationPath }
	} else {
		configIsolationPathFn = func(exec base.Executor) string { return t.TempDir() }
	}
	if o.processFn != nil {
		commonProcessProcessFn = o.processFn
	} else {
		commonProcessProcessFn = func(s string) (string, *os.Process, error) { return "", nil, nil }
	}
	gameRunningFn = func() bool { return o.gameRunning }
	if o.bufferFn != nil {
		commonLoggerFileLoggerBufferFn = o.bufferFn
	} else {
		commonLoggerFileLoggerBufferFn = func(name string, fn func(io.Writer)) error { return nil }
	}
	if o.killAgent != nil {
		configKillAgentFn = o.killAgent
	} else {
		configKillAgentFn = func() {}
	}
	if o.parseCommandArgsFn != nil {
		parseCommandArgsFn = o.parseCommandArgsFn
	} else {
		parseCommandArgsFn = defaultParseCommandArgs
	}
	if o.resolveIsolateValueFnVal != nil {
		resolveIsolateValueFn = o.resolveIsolateValueFnVal
	} else {
		resolveIsolateValueFn = func(v string, b bool) bool { return v == "true" || (v == "required" && b) }
	}
	if o.configSetGameIdFnVal != nil {
		configSetGameIdFn = o.configSetGameIdFnVal
	} else {
		configSetGameIdFn = func(s string) {}
	}
	if o.commonParsePathFnVal != nil {
		commonParsePathFn = o.commonParsePathFnVal
	} else {
		commonParsePathFn = func(s []string, m map[string]string) (os.FileInfo, string, error) {
			return fakeFileInfo{isDir: true}, t.TempDir(), nil
		}
	}
	if o.commonEnhancedViperFnVal != nil {
		commonEnhancedViperFn = o.commonEnhancedViperFnVal
	} else {
		commonEnhancedViperFn = func(s string) []string { return []string{s} }
	}
	if o.configNativeMacOsGameFnVal != nil {
		configNativeMacOsGameFn = o.configNativeMacOsGameFnVal
	} else {
		configNativeMacOsGameFn = func(e base.Executor, b bool) bool { return false }
	}
	if o.configBattleServerRequiredFnVal != nil {
		configBattleServerRequiredFn = o.configBattleServerRequiredFnVal
	} else {
		configBattleServerRequiredFn = func(e base.Executor) bool { return false }
	}
	if o.newConfigFlushCacheOptionsFnVal != nil {
		newConfigFlushCacheOptionsFn = o.newConfigFlushCacheOptionsFnVal
	} else {
		newConfigFlushCacheOptionsFn = executor.NewConfigFlushCacheOptions
	}
	if o.discoverServersFnVal != nil {
		discoverServersFn = o.discoverServersFnVal
	} else {
		discoverServersFn = func(string, bool, mapset.Set[netip.Addr], mapset.Set[uint16]) (uuid.UUID, net.IP) {
			return uuid.Nil(), nil
		}
	}
	if o.netipParseAddrFnVal != nil {
		netipParseAddrFn = o.netipParseAddrFnVal
	} else {
		netipParseAddrFn = netip.ParseAddr
	}
	if o.serverFilterServerIPsFnVal != nil {
		serverFilterServerIPsFn = o.serverFilterServerIPsFnVal
	} else {
		serverFilterServerIPsFn = func(uuid.UUID, string, string, mapset.Set[netip.Addr]) (uuid.UUID, []server.MesuredIpAddress, *commonServer.AnnounceMessageDataSupportedLatest) {
			return uuid.Nil(), nil, nil
		}
	}
	if o.serverGetExecutablePathFnVal != nil {
		serverGetExecutablePathFn = o.serverGetExecutablePathFnVal
	} else {
		serverGetExecutablePathFn = func(s string) string { return s }
	}
	if o.serverGenerateCertsFnVal != nil {
		serverGenerateCertsFn = o.serverGenerateCertsFnVal
	} else {
		serverGenerateCertsFn = func(s string, b bool) int { return common.ErrSuccess }
	}
	if o.configRunBattleServerManagerFnVal != nil {
		configRunBattleServerManagerFn = o.configRunBattleServerManagerFnVal
	} else {
		configRunBattleServerManagerFn = func(s string, fs *pflag.FlagSet, v *bsManager.StartValues, b bool) int { return common.ErrSuccess }
	}
	if o.bsManagerStartFlagSetFnVal != nil {
		bsManagerStartFlagSetFn = o.bsManagerStartFlagSetFnVal
	} else {
		bsManagerStartFlagSetFn = bsManager.StartFlagSet
	}
	if o.configStartServerFnVal != nil {
		configStartServerFn = o.configStartServerFnVal
	} else {
		configStartServerFn = func(s string, fs *pflag.FlagSet, v *cmdServer.Values, b bool) (int, string) {
			return common.ErrSuccess, "127.0.0.1"
		}
	}
	if o.serverReadCACertFnVal != nil {
		serverReadCACertFn = o.serverReadCACertFnVal
	} else {
		serverReadCACertFn = func(s string) *x509.Certificate { return &x509.Certificate{} }
	}
	if o.configMapHostsFnVal != nil {
		configMapHostsFn = o.configMapHostsFnVal
	} else {
		configMapHostsFn = func(s1, s2 string, b1, b2, b3 bool) int { return common.ErrSuccess }
	}
	if o.configAddCertFnVal != nil {
		configAddCertFn = o.configAddCertFnVal
	} else {
		configAddCertFn = func(s1 string, u uuid.UUID, c *x509.Certificate, s2 string, b1, b2 bool) int {
			return common.ErrSuccess
		}
	}
	if o.configIsolateUserDataFnVal != nil {
		configIsolateUserDataFn = o.configIsolateUserDataFnVal
	} else {
		configIsolateUserDataFn = func(b1, b2 bool, s string) int { return common.ErrSuccess }
	}
	if o.configAddCACertToGameFnVal != nil {
		configAddCACertToGameFn = o.configAddCACertToGameFnVal
	} else {
		configAddCACertToGameFn = func(s1 string, u uuid.UUID, c *x509.Certificate, s2, s3 string, b1, b2 bool) int {
			return common.ErrSuccess
		}
	}
	if o.configLaunchAgentAndGameFnVal != nil {
		configLaunchAgentAndGameFn = o.configLaunchAgentAndGameFnVal
	} else {
		configLaunchAgentAndGameFn = func(e base.Executor, ce custom.Exec, as []string, s1, s2, s3 string) int { return common.ErrSuccess }
	}
	if o.uuidParseFnVal != nil {
		uuidParseFn = o.uuidParseFnVal
	} else {
		uuidParseFn = uuid.Parse
	}
	if o.uuidMustParseFnVal != nil {
		uuidMustParseFn = o.uuidMustParseFnVal
	} else {
		uuidMustParseFn = uuid.MustParse
	}
	if o.uuidNilFnVal != nil {
		uuidNilFn = o.uuidNilFnVal
	} else {
		uuidNilFn = uuid.Nil
	}
	if o.executablesNativeFileNameFnVal != nil {
		executablesNativeFileNameFn = o.executablesNativeFileNameFnVal
	} else {
		executablesNativeFileNameFn = func(b bool, s string) string { return s }
	}
	if o.configRunSetupCommandFnVal != nil {
		configRunSetupCommandFn = o.configRunSetupCommandFnVal
	} else {
		configRunSetupCommandFn = func(s []string) *commonExecutor.Result { return &commonExecutor.Result{} }
	}
	if o.dnsConnectivityFnVal != nil {
		dnsConnectivityFn = o.dnsConnectivityFnVal
	} else {
		dnsConnectivityFn = func() bool { return false }
	}
	if o.configRunStopAgentFnVal != nil {
		configRunStopAgentFn = o.configRunStopAgentFnVal
	} else {
		configRunStopAgentFn = func() *commonExecutor.Result { return &commonExecutor.Result{} }
	}
	if o.waitForProcessFnVal != nil {
		commonProcessWaitForProcessFn = o.waitForProcessFnVal
	} else {
		commonProcessWaitForProcessFn = func(*os.Process, *time.Duration) bool { return true }
	}
	if o.dialogNewFnVal != nil {
		dialogNewFn = o.dialogNewFnVal
	} else {
		// Confirm by default so the tests keep reaching the server start path,
		// which is what they were written for.
		dialogNewFn = func(string) dialog.Resolution {
			return dialog.Resolution{Dialog: &fakeDialog{confirm: true}, Name: "fake"}
		}
	}

	return func() {
		dialog.Reset()
		gameId, cfgFile, gameCfgFile = origGameId, origCfgFile, origGameCfgFile
		newPidLockFn = origNewPidLock
		initConfigFn = origInitConfig
		openMainLogFn = origOpenMainLog
		isAdminFn = origIsAdmin
		gameSupportedGamesContainsOneFn = origGameSupported
		makeExecFn = origMakeExec
		configIsolationPathFn = origIsolationPath
		gameRunningFn = origGameRunning
		commonProcessProcessFn = origProcess
		commonLoggerFileLoggerBufferFn = origBuffer
		configKillAgentFn = origKillAgent
		parseCommandArgsFn = origParseCommandArgs
		resolveIsolateValueFn = origResolveIsolate
		configSetGameIdFn = origConfigSetGameId
		commonParsePathFn = origCommonParsePath
		commonEnhancedViperFn = origCommonEnhancedViper
		configNativeMacOsGameFn = origConfigNativeMacOsGame
		configBattleServerRequiredFn = origConfigBattleServerRequired
		newConfigFlushCacheOptionsFn = origNewConfigFlushCacheOptions
		discoverServersFn = origDiscoverServers
		netipParseAddrFn = origNetipParseAddr
		serverFilterServerIPsFn = origServerFilterServerIPs
		serverGetExecutablePathFn = origServerGetExecutablePath
		serverGenerateCertsFn = origServerGenerateCerts
		configRunBattleServerManagerFn = origConfigRunBattleServerManager
		bsManagerStartFlagSetFn = origBsManagerStartFlagSet
		configStartServerFn = origConfigStartServer
		serverReadCACertFn = origServerReadCACert
		configMapHostsFn = origConfigMapHosts
		configAddCertFn = origConfigAddCert
		configIsolateUserDataFn = origConfigIsolateUserData
		configAddCACertToGameFn = origConfigAddCACertToGame
		configLaunchAgentAndGameFn = origConfigLaunchAgentAndGame
		uuidParseFn = origUuidParse
		uuidMustParseFn = origUuidMustParse
		uuidNilFn = origUuidNil
		executablesNativeFileNameFn = origExecutablesNativeFileName
		configRunSetupCommandFn = origConfigRunSetupCommand
		dnsConnectivityFn = origDNSConnectivity
		configRunStopAgentFn = origRunStopAgent
		commonProcessWaitForProcessFn = origWaitForProcess
		dialogNewFn = origDialogNew
	}
}

func defaultParseCommandArgs(args []string, values map[string]string) ([]string, error) {
	result := make([]string, len(args))
	for i, arg := range args {
		for k, v := range values {
			arg = strings.ReplaceAll(arg, "{"+k+"}", v)
		}
		result[i] = arg
	}
	return result, nil
}

func TestRunRootFlushCacheError(t *testing.T) {
	restore := applyOverrides(t, runRootOverrides{
		gameId:        "age2",
		isAdmin:       false,
		gameSupported: true,
		newConfigFlushCacheOptionsFnVal: func(canAddHost bool, canTrust string, customHostFile, customCertFile bool) *executor.ConfigFlushCacheOptions {
			if canAddHost || canTrust != "false" {
				return &executor.ConfigFlushCacheOptions{}
			}
			return nil
		},
	})
	defer restore()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	_, exitCode := runRoot(fs)
	if exitCode == common.ErrSuccess {
		t.Errorf("expected non-success for flush cache with empty FlushCacheValues, got %d", exitCode)
	}
}

func TestRunRootMulticastInvalid(t *testing.T) {
	restore := applyOverrides(t, runRootOverrides{
		gameId:        "age2",
		isAdmin:       false,
		gameSupported: true,
		cfg: func() *launcher.Configuration {
			c := validLauncherConfig()
			c.Server.AnnounceMulticastGroups = []string{"not-a-multicast"}
			return c
		},
	})
	defer restore()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	_, exitCode := runRoot(fs)
	if exitCode != launcher.ErrAnnouncementMulticastGroup {
		t.Errorf("expected %d for invalid multicast group, got %d", launcher.ErrAnnouncementMulticastGroup, exitCode)
	}
}

func TestRunRootServerFoundByDiscovery(t *testing.T) {
	discoveredUUID := uuid.New()
	restore := applyOverrides(t, runRootOverrides{
		gameId:        "age2",
		isAdmin:       false,
		gameSupported: true,
		discoverServersFnVal: func(gameTitle string, single bool, mc mapset.Set[netip.Addr], ports mapset.Set[uint16]) (uuid.UUID, net.IP) {
			return discoveredUUID, net.ParseIP("192.168.1.100")
		},
		serverReadCACertFnVal: func(host string) *x509.Certificate {
			return &x509.Certificate{}
		},
		configLaunchAgentAndGameFnVal: func(e base.Executor, ce custom.Exec, as []string, s1, s2, s3 string) int {
			return common.ErrSuccess
		},
	})
	defer restore()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	_, exitCode := runRoot(fs)
	if exitCode != common.ErrSuccess {
		t.Errorf("expected success for server found by discovery, got %d", exitCode)
	}
}

func TestRunRootServerHostEmpty(t *testing.T) {
	restore := applyOverrides(t, runRootOverrides{
		gameId:        "age2",
		isAdmin:       false,
		gameSupported: true,
		cfg: func() *launcher.Configuration {
			c := validLauncherConfig()
			c.Server.Start = "false"
			c.Server.Host = ""
			return c
		},
	})
	defer restore()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	_, exitCode := runRoot(fs)
	if exitCode != launcher.ErrInvalidServerHost {
		t.Errorf("expected %d for empty serverHost, got %d", launcher.ErrInvalidServerHost, exitCode)
	}
}

func TestRunRootServerHostIPv6(t *testing.T) {
	restore := applyOverrides(t, runRootOverrides{
		gameId:        "age2",
		isAdmin:       false,
		gameSupported: true,
		cfg: func() *launcher.Configuration {
			c := validLauncherConfig()
			c.Server.Start = "false"
			c.Server.Host = "::1"
			return c
		},
	})
	defer restore()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	_, exitCode := runRoot(fs)
	if exitCode != launcher.ErrInvalidServerHost {
		t.Errorf("expected %d for IPv6 serverHost, got %d", launcher.ErrInvalidServerHost, exitCode)
	}
}

func TestRunRootServerHostResolutionFailure(t *testing.T) {
	restore := applyOverrides(t, runRootOverrides{
		gameId:        "age2",
		isAdmin:       false,
		gameSupported: true,
		cfg: func() *launcher.Configuration {
			c := validLauncherConfig()
			c.Server.Start = "false"
			c.Server.Host = "192.168.1.50"
			return c
		},
		serverFilterServerIPsFnVal: func(id uuid.UUID, name, game string, addrs mapset.Set[netip.Addr]) (uuid.UUID, []server.MesuredIpAddress, *commonServer.AnnounceMessageDataSupportedLatest) {
			return uuid.Nil(), nil, nil
		},
	})
	defer restore()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	_, exitCode := runRoot(fs)
	if exitCode != launcher.ErrInvalidServerHost {
		t.Errorf("expected %d for server host resolution failure, got %d", launcher.ErrInvalidServerHost, exitCode)
	}
}

func TestRunRootServerExecutableNotFound(t *testing.T) {
	restore := applyOverrides(t, runRootOverrides{
		gameId:                       "age2",
		isAdmin:                      false,
		gameSupported:                true,
		serverGetExecutablePathFnVal: func(s string) string { return "" },
	})
	defer restore()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	_, exitCode := runRoot(fs)
	if exitCode != launcher.ErrServerExecutable {
		t.Errorf("expected %d for server executable not found, got %d", launcher.ErrServerExecutable, exitCode)
	}
}

func TestRunRootReadCertFailure(t *testing.T) {
	restore := applyOverrides(t, runRootOverrides{
		gameId:                "age2",
		isAdmin:               false,
		gameSupported:         true,
		serverReadCACertFnVal: func(host string) *x509.Certificate { return nil },
	})
	defer restore()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	_, exitCode := runRoot(fs)
	if exitCode != launcher.ErrReadCert {
		t.Errorf("expected %d for read cert failure, got %d", launcher.ErrReadCert, exitCode)
	}
}

func TestRunRootMapHostsFailure(t *testing.T) {
	restore := applyOverrides(t, runRootOverrides{
		gameId:        "age2",
		isAdmin:       false,
		gameSupported: true,
		configMapHostsFnVal: func(s1, s2 string, b1, b2, b3 bool) int {
			return common.ErrGeneral
		},
	})
	defer restore()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	_, exitCode := runRoot(fs)
	if exitCode != common.ErrGeneral {
		t.Errorf("expected %d for map hosts failure, got %d", common.ErrGeneral, exitCode)
	}
}

func TestRunRootAddCertFailure(t *testing.T) {
	restore := applyOverrides(t, runRootOverrides{
		gameId:        "age2",
		isAdmin:       false,
		gameSupported: true,
		configAddCertFnVal: func(s1 string, u uuid.UUID, c *x509.Certificate, s2 string, b1, b2 bool) int {
			return common.ErrGeneral
		},
	})
	defer restore()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	_, exitCode := runRoot(fs)
	if exitCode != common.ErrGeneral {
		t.Errorf("expected %d for add cert failure, got %d", common.ErrGeneral, exitCode)
	}
}

func TestRunRootIsolateUserDataFailure(t *testing.T) {
	restore := applyOverrides(t, runRootOverrides{
		gameId:        "age2",
		isAdmin:       false,
		gameSupported: true,
		configIsolateUserDataFnVal: func(b1, b2 bool, s string) int {
			return common.ErrGeneral
		},
	})
	defer restore()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	_, exitCode := runRoot(fs)
	if exitCode != common.ErrGeneral {
		t.Errorf("expected %d for isolate user data failure, got %d", common.ErrGeneral, exitCode)
	}
}

func TestRunRootStartServerFailure(t *testing.T) {
	restore := applyOverrides(t, runRootOverrides{
		gameId:        "age2",
		isAdmin:       false,
		gameSupported: true,
		configStartServerFnVal: func(s string, fs *pflag.FlagSet, v *cmdServer.Values, b bool) (int, string) {
			return common.ErrGeneral, ""
		},
	})
	defer restore()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	_, exitCode := runRoot(fs)
	if exitCode != common.ErrGeneral {
		t.Errorf("expected %d for start server failure, got %d", common.ErrGeneral, exitCode)
	}
}

func TestRunRootBattleServerManagerParseFailure(t *testing.T) {
	callCount := 0
	restore := applyOverrides(t, runRootOverrides{
		gameId:        "age2",
		isAdmin:       false,
		gameSupported: true,
		cfg: func() *launcher.Configuration {
			c := validLauncherConfig()
			c.Server.BattleServerManager.Run = "true"
			return c
		},
		parseCommandArgsFn: func(args []string, values map[string]string) ([]string, error) {
			callCount++
			if callCount == 2 {
				return nil, errors.New("bs manager parse fail")
			}
			return defaultParseCommandArgs(args, values)
		},
	})
	defer restore()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	_, exitCode := runRoot(fs)
	if exitCode != launcher.ErrInvalidServerBattleServerManagerArgs {
		t.Errorf("expected %d for battle server manager parse failure, got %d", launcher.ErrInvalidServerBattleServerManagerArgs, exitCode)
	}
}

func TestRunRootLaunchAgentSuccess(t *testing.T) {
	restore := applyOverrides(t, runRootOverrides{
		gameId:        "age1",
		isAdmin:       false,
		gameSupported: true,
		configLaunchAgentAndGameFnVal: func(e base.Executor, ce custom.Exec, as []string, s1, s2, s3 string) int {
			return common.ErrSuccess
		},
	})
	defer restore()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	_, exitCode := runRoot(fs)
	if exitCode != common.ErrSuccess {
		t.Errorf("expected success for launch agent, got %d", exitCode)
	}
}

func TestRunRootServerFoundWithFilter(t *testing.T) {
	discoveredUUID := uuid.New()
	restore := applyOverrides(t, runRootOverrides{
		gameId:        "age2",
		isAdmin:       false,
		gameSupported: true,
		cfg: func() *launcher.Configuration {
			c := validLauncherConfig()
			c.Server.Start = "false"
			c.Server.Host = "192.168.1.50"
			return c
		},
		serverFilterServerIPsFnVal: func(id uuid.UUID, name, game string, addrs mapset.Set[netip.Addr]) (uuid.UUID, []server.MesuredIpAddress, *commonServer.AnnounceMessageDataSupportedLatest) {
			fakeData := &commonServer.AnnounceMessageDataSupportedLatest{}
			return discoveredUUID, []server.MesuredIpAddress{{Ip: net.ParseIP("192.168.1.50")}}, fakeData
		},
		serverReadCACertFnVal: func(host string) *x509.Certificate {
			return &x509.Certificate{}
		},
		configLaunchAgentAndGameFnVal: func(e base.Executor, ce custom.Exec, as []string, s1, s2, s3 string) int {
			return common.ErrSuccess
		},
	})
	defer restore()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	_, exitCode := runRoot(fs)
	if exitCode != common.ErrSuccess {
		t.Errorf("expected success for server found with filter, got %d", exitCode)
	}
}

func TestRunRootServerNotFoundNoServerHost(t *testing.T) {
	restore := applyOverrides(t, runRootOverrides{
		gameId:        "age2",
		isAdmin:       false,
		gameSupported: true,
		discoverServersFnVal: func(gameTitle string, single bool, mc mapset.Set[netip.Addr], ports mapset.Set[uint16]) (uuid.UUID, net.IP) {
			return uuid.Nil(), nil
		},
		serverGetExecutablePathFnVal: func(s string) string { return s },
		serverReadCACertFnVal: func(host string) *x509.Certificate {
			return &x509.Certificate{}
		},
		configLaunchAgentAndGameFnVal: func(e base.Executor, ce custom.Exec, as []string, s1, s2, s3 string) int {
			return common.ErrSuccess
		},
	})
	defer restore()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	_, exitCode := runRoot(fs)
	if exitCode != common.ErrSuccess {
		t.Errorf("expected success when starting server after no discovery, got %d", exitCode)
	}
}

func TestRunRootCanUseInternetDisabledByConfig(t *testing.T) {
	restore := applyOverrides(t, runRootOverrides{
		gameId:        "age2",
		isAdmin:       false,
		gameSupported: true,
		cfg: func() *launcher.Configuration {
			c := validLauncherConfig()
			c.Config.CanUseInternet = false
			return c
		},
		dnsConnectivityFnVal: func() bool { return true },
	})
	defer restore()
	origInternet := launcher.CanUseInternet
	defer func() {
		launcher.CanUseInternet = origInternet
		common.SetUseInternet(origInternet)
	}()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	_, exitCode := runRoot(fs)
	if exitCode != common.ErrSuccess {
		t.Fatalf("expected success, got %d", exitCode)
	}
	if launcher.CanUseInternet {
		t.Error("launcher.CanUseInternet should be false when config disables internet")
	}
}

func TestRunRootCanUseInternetProbeWhenConnectivity(t *testing.T) {
	restore := applyOverrides(t, runRootOverrides{
		gameId:        "age2",
		isAdmin:       false,
		gameSupported: true,
		cfg: func() *launcher.Configuration {
			c := validLauncherConfig()
			c.Config.CanUseInternet = true
			return c
		},
		dnsConnectivityFnVal: func() bool { return true },
	})
	defer restore()
	origInternet := launcher.CanUseInternet
	defer func() {
		launcher.CanUseInternet = origInternet
		common.SetUseInternet(origInternet)
	}()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	_, exitCode := runRoot(fs)
	if exitCode != common.ErrSuccess {
		t.Fatalf("expected success, got %d", exitCode)
	}
	if !launcher.CanUseInternet {
		t.Error("launcher.CanUseInternet should be true when connectivity probe succeeds")
	}
}

func TestRunRootCanUseInternetProbeNoConnectivity(t *testing.T) {
	restore := applyOverrides(t, runRootOverrides{
		gameId:        "age2",
		isAdmin:       false,
		gameSupported: true,
		cfg: func() *launcher.Configuration {
			c := validLauncherConfig()
			c.Config.CanUseInternet = true
			return c
		},
		dnsConnectivityFnVal: func() bool { return false },
	})
	defer restore()
	origInternet := launcher.CanUseInternet
	defer func() {
		launcher.CanUseInternet = origInternet
		common.SetUseInternet(origInternet)
	}()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	_, exitCode := runRoot(fs)
	if exitCode != common.ErrSuccess {
		t.Fatalf("expected success, got %d", exitCode)
	}
	if launcher.CanUseInternet {
		t.Error("launcher.CanUseInternet should be false when connectivity probe fails")
	}
}

func TestRunRootCanTrustCertificateAuto(t *testing.T) {
	restore := applyOverrides(t, runRootOverrides{
		gameId:        "age2",
		isAdmin:       false,
		gameSupported: true,
		cfg: func() *launcher.Configuration {
			c := validLauncherConfig()
			c.Config.Certificate.CanTrustInPc = "auto"
			return c
		},
	})
	defer restore()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	_, exitCode := runRoot(fs)
	if exitCode != common.ErrSuccess {
		t.Errorf("expected success for canTrustCertificate auto, got %d", exitCode)
	}
}

// Regression: the launcher used to wait 10s for a leftover 'config-admin-agent'
// to exit on its own, which never happens because nothing asks it to. It then
// reused the stale agent, which still held the previous session's mapped-ips and
// certificate state, so the next setUp failed with "already mapped". It must ask
// the agent to stop instead.
func TestRunRootStopsLeftoverConfigAdminAgent(t *testing.T) {
	stopAgentCalls := 0
	restore := applyOverrides(t, runRootOverrides{
		gameId:        "age2",
		isAdmin:       false,
		gameSupported: true,
		cfg:           validLauncherConfig,
		// Only the config-admin-agent looks alive; agent.exe must not.
		processFn: func(name string) (string, *os.Process, error) {
			if strings.Contains(name, executables.LauncherConfigAdminAgent) {
				return "", &os.Process{Pid: 4242}, nil
			}
			return "", nil, nil
		},
		configRunStopAgentFnVal: func() *commonExecutor.Result {
			stopAgentCalls++
			return &commonExecutor.Result{ExitCode: common.ErrSuccess}
		},
	})
	defer restore()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	if _, exitCode := runRoot(fs); exitCode != common.ErrSuccess {
		t.Fatalf("stopping a leftover agent must not fail the launch, got %d", exitCode)
	}
	if stopAgentCalls != 1 {
		t.Errorf("stopAgent called %d times, want exactly 1", stopAgentCalls)
	}
}

// The old code waited on the leftover process, which is the behaviour this
// replaced. It must not wait at all for config-admin-agent, or every launch
// after a leak stalls before doing anything.
func TestRunRootDoesNotWaitForLeftoverConfigAdminAgent(t *testing.T) {
	restore := applyOverrides(t, runRootOverrides{
		gameId:        "age2",
		isAdmin:       false,
		gameSupported: true,
		cfg:           validLauncherConfig,
		processFn: func(name string) (string, *os.Process, error) {
			if strings.Contains(name, executables.LauncherConfigAdminAgent) {
				return "", &os.Process{Pid: 4242}, nil
			}
			return "", nil, nil
		},
		configRunStopAgentFnVal: func() *commonExecutor.Result {
			return &commonExecutor.Result{ExitCode: common.ErrSuccess}
		},
		waitForProcessFnVal: func(proc *os.Process, d *time.Duration) bool {
			t.Errorf("waited for pid %d; the leftover agent must be asked to stop, not waited on", proc.Pid)
			return true
		},
	})
	defer restore()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	if _, exitCode := runRoot(fs); exitCode != common.ErrSuccess {
		t.Fatalf("got %d", exitCode)
	}
}

// No agent running means nothing to stop and no child process spawned.
func TestRunRootNoStopAgentWhenNoneRunning(t *testing.T) {
	stopAgentCalls := 0
	restore := applyOverrides(t, runRootOverrides{
		gameId:        "age2",
		isAdmin:       false,
		gameSupported: true,
		cfg:           validLauncherConfig,
		processFn:     func(string) (string, *os.Process, error) { return "", nil, nil },
		configRunStopAgentFnVal: func() *commonExecutor.Result {
			stopAgentCalls++
			return &commonExecutor.Result{ExitCode: common.ErrSuccess}
		},
	})
	defer restore()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	if _, exitCode := runRoot(fs); exitCode != common.ErrSuccess {
		t.Fatalf("got %d", exitCode)
	}
	if stopAgentCalls != 0 {
		t.Errorf("stopAgent called %d times with no agent running, want 0", stopAgentCalls)
	}
}

// Regression: a failed stop must be reported but must not abort the launch. The
// launcher's job at that point is to clean up, not to refuse to start.
func TestRunRootSurvivesFailedStopAgent(t *testing.T) {
	restore := applyOverrides(t, runRootOverrides{
		gameId:        "age2",
		isAdmin:       false,
		gameSupported: true,
		cfg:           validLauncherConfig,
		processFn: func(name string) (string, *os.Process, error) {
			if strings.Contains(name, executables.LauncherConfigAdminAgent) {
				return "", &os.Process{Pid: 4242}, nil
			}
			return "", nil, nil
		},
		configRunStopAgentFnVal: func() *commonExecutor.Result {
			return &commonExecutor.Result{Err: errors.New("stop failed"), ExitCode: 1}
		},
	})
	defer restore()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	if _, exitCode := runRoot(fs); exitCode != common.ErrSuccess {
		t.Fatalf("a failed agent stop must not fail the launch, got %d", exitCode)
	}
}

// Regression: the signal handler and the deferred cleanup both used to run
// config.Revert() and lock.Unlock() with no guard, so a Ctrl+C arriving while
// the main path was finishing spawned two config.exe reverts over one hosts
// lock and two stop-agent clients on one named pipe, and os.Exit could truncate
// a revert still in flight. Teardown must happen at most once, and the second
// caller must wait rather than race past.
func TestRunRootTeardownRunsOnceUnderSignal(t *testing.T) {
	var sigs chan<- os.Signal
	unlocked := make(chan struct{}, 4)

	restore := applyOverrides(t, runRootOverrides{
		gameId:        "age2",
		isAdmin:       false,
		gameSupported: true,
		cfg:           validLauncherConfig,
		newPidLock: func() fileLock.Locker {
			return &fakePidLocker{}
		},
	})
	defer restore()

	// Own the pid lock so Unlock is observable.
	locker := &countingLocker{onUnlock: func() { unlocked <- struct{}{} }}
	origPidLock := newPidLockFn
	newPidLockFn = func() fileLock.Locker { return locker }
	t.Cleanup(func() { newPidLockFn = origPidLock })

	origSignal := signalNotifyFn
	origStop := signalStopFn
	signalNotifyFn = func(c chan<- os.Signal, _ ...os.Signal) { sigs = c }
	signalStopFn = func(chan<- os.Signal) {}
	t.Cleanup(func() { signalNotifyFn = origSignal; signalStopFn = origStop })

	// Hold runRoot inside the "a previous agent is still running" wait so the
	// signal lands while the main path is still running, which is the race.
	waiting := make(chan struct{})
	release := make(chan struct{})
	origWait := commonProcessWaitForProcessFn
	commonProcessWaitForProcessFn = func(*os.Process, *time.Duration) bool {
		close(waiting)
		<-release
		return true
	}
	t.Cleanup(func() { commonProcessWaitForProcessFn = origWait })

	origProcess := commonProcessProcessFn
	commonProcessProcessFn = func(name string) (string, *os.Process, error) {
		if strings.Contains(name, executables.LauncherAgent) {
			return "", &os.Process{Pid: 1}, nil
		}
		return "", nil, nil
	}
	t.Cleanup(func() { commonProcessProcessFn = origProcess })

	done := make(chan struct{})
	go func() {
		defer close(done)
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		_, _ = runRoot(fs)
	}()

	<-waiting
	sigs <- syscall.SIGINT
	// Wait for the signal path to finish its teardown before letting the main
	// path return, so both definitely reach teardown.
	select {
	case <-unlocked:
	case <-time.After(10 * time.Second):
		t.Fatal("signal path never completed teardown")
	}
	close(release)
	<-done

	if n := locker.calls(); n != 1 {
		t.Errorf("lock unlocked %d times, want exactly 1", n)
	}
}

// countingLocker records Unlock calls.
type countingLocker struct {
	fakePidLocker
	calls32  int32
	onUnlock func()
}

func (c *countingLocker) Unlock() error {
	atomic.AddInt32(&c.calls32, 1)
	if c.onUnlock != nil {
		c.onUnlock()
	}
	return nil
}

func (c *countingLocker) calls() int32 { return atomic.LoadInt32(&c.calls32) }

// An invalid dialog mode must abort before anything is started or reverted on
// the user's behalf.
func TestRunRootInvalidDialogValue(t *testing.T) {
	startServerCalls := 0
	restore := applyOverrides(t, runRootOverrides{
		gameId:        "age2",
		isAdmin:       false,
		gameSupported: true,
		cfg: func() *launcher.Configuration {
			c := validLauncherConfig()
			c.Config.Dialog = "xxx"
			return c
		},
		configStartServerFnVal: func(string, *pflag.FlagSet, *cmdServer.Values, bool) (int, string) {
			startServerCalls++
			return common.ErrSuccess, "127.0.0.1"
		},
	})
	defer restore()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	if _, exitCode := runRoot(fs); exitCode != launcher.ErrInvalidDialog {
		t.Fatalf("got %d, want %d", exitCode, launcher.ErrInvalidDialog)
	}
	if startServerCalls != 0 {
		t.Errorf("the server was started %d times with an invalid dialog mode", startServerCalls)
	}
}

// A "true" mode on a system without graphical dialogs must warn and continue on
// the console, never abort.
func TestRunRootUsesConsoleWhenDialogsUnavailable(t *testing.T) {
	const reason = "Graphical dialogs are not available in this system, using the console instead."
	restore := applyOverrides(t, runRootOverrides{
		gameId:        "age2",
		isAdmin:       false,
		gameSupported: true,
		cfg: func() *launcher.Configuration {
			c := validLauncherConfig()
			c.Config.Dialog = dialog.ModeTrue
			return c
		},
		dialogNewFnVal: func(mode string) dialog.Resolution {
			if mode != dialog.ModeTrue {
				t.Errorf("dialog.New called with %q, want %q", mode, dialog.ModeTrue)
			}
			return dialog.Resolution{Dialog: &fakeDialog{confirm: true}, Name: "console", Reason: reason}
		},
	})
	defer restore()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	out := captureStdout(t, func() {
		if _, exitCode := runRoot(fs); exitCode != common.ErrSuccess {
			t.Errorf("an unavailable dialog must not abort the launch, got %d", exitCode)
		}
	})
	// The backend name is not printed any more: a console dialog is visible, and a
	// graphical one is too. Only the fallback matters, because it means the
	// configured choice was not honoured, and it has to reach the user.
	if strings.Contains(out, "Dialog backend") {
		t.Errorf("the backend line is gone, got:\n%s", out)
	}
	if !strings.Contains(out, reason) {
		t.Errorf("output missing %q, got:\n%s", reason, out)
	}
}

// Regression: the old prompt only had "press enter", so there was no way to
// decline. A declined confirmation must abort with its own exit code, and
// nothing may be started on the user's behalf.
func TestRunRootServerStartCanceled(t *testing.T) {
	dlg := &fakeDialog{confirm: false}
	startServerCalls := 0
	restore := applyOverrides(t, runRootOverrides{
		gameId:        "age2",
		isAdmin:       false,
		gameSupported: true,
		dialogNewFnVal: func(string) dialog.Resolution {
			return dialog.Resolution{Dialog: dlg, Name: "fake"}
		},
		configStartServerFnVal: func(string, *pflag.FlagSet, *cmdServer.Values, bool) (int, string) {
			startServerCalls++
			return common.ErrSuccess, "127.0.0.1"
		},
	})
	defer restore()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	if _, exitCode := runRoot(fs); exitCode != launcher.ErrServerStartCanceled {
		t.Fatalf("got %d, want %d", exitCode, launcher.ErrServerStartCanceled)
	}
	if dlg.confirmCalls != 1 {
		t.Errorf("the confirmation was asked %d times, want 1", dlg.confirmCalls)
	}
	if startServerCalls != 0 {
		t.Errorf("the server was started %d times after the user canceled", startServerCalls)
	}
}

func TestRunRootServerStartConfirmed(t *testing.T) {
	dlg := &fakeDialog{confirm: true}
	var gotValues *cmdServer.Values
	restore := applyOverrides(t, runRootOverrides{
		gameId:        "age2",
		isAdmin:       false,
		gameSupported: true,
		dialogNewFnVal: func(string) dialog.Resolution {
			return dialog.Resolution{Dialog: dlg, Name: "fake"}
		},
		configStartServerFnVal: func(_ string, _ *pflag.FlagSet, v *cmdServer.Values, b bool) (int, string) {
			gotValues = v
			return common.ErrSuccess, "127.0.0.1"
		},
	})
	defer restore()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	if _, exitCode := runRoot(fs); exitCode != common.ErrSuccess {
		t.Fatalf("got %d, want success", exitCode)
	}
	if dlg.confirmCalls != 1 {
		t.Errorf("the confirmation was asked %d times, want 1", dlg.confirmCalls)
	}
	if gotValues == nil {
		t.Fatal("the server was never started")
	}
	if len(gotValues.GameIds) != 1 || gotValues.GameIds[0] != "age2" {
		t.Errorf("server started with game ids %v, want [age2]", gotValues.GameIds)
	}
}

// The dialog double answers false, so a call would surface as an unexpected
// exit code rather than as a silent false positive.
func TestRunRootServerStartWithoutConfirmationSkipsDialog(t *testing.T) {
	dlg := &fakeDialog{confirm: false}
	restore := applyOverrides(t, runRootOverrides{
		gameId:        "age2",
		isAdmin:       false,
		gameSupported: true,
		cfg: func() *launcher.Configuration {
			c := validLauncherConfig()
			c.Server.StartWithoutConfirmation = true
			return c
		},
		dialogNewFnVal: func(string) dialog.Resolution {
			return dialog.Resolution{Dialog: dlg, Name: "fake"}
		},
	})
	defer restore()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	if _, exitCode := runRoot(fs); exitCode != common.ErrSuccess {
		t.Fatalf("got %d, want success without asking", exitCode)
	}
	if dlg.confirmCalls != 0 {
		t.Errorf("the confirmation was asked %d times, want 0", dlg.confirmCalls)
	}
}

// The header used to announce the config files and the game three times over: in
// the summary at the top, in a line of its own where they were loaded, and once
// more as "Game age2." when the game was about to be looked for. One place is
// enough, and the phases have to be visible as headings.
func TestRunRootHeaderHasNoDuplicates(t *testing.T) {
	restore := applyOverrides(t, runRootOverrides{
		gameId:        "age2",
		isAdmin:       false,
		gameSupported: true,
		dialogNewFnVal: func(string) dialog.Resolution {
			return dialog.Resolution{Dialog: &fakeDialog{confirm: true}, Name: "console"}
		},
	})
	defer restore()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	out := captureStdout(t, func() {
		if _, exitCode := runRoot(fs); exitCode != common.ErrSuccess {
			t.Errorf("got %d, want success", exitCode)
		}
	})
	for _, gone := range []string{
		"Using main config file",
		"Using game config file",
		"Game age2",
		"Dialog backend",
	} {
		if strings.Contains(out, gone) {
			t.Errorf("%q is printed more than once, got:\n%s", gone, out)
		}
	}
	for _, want := range []string{"Configuration", "Initial teardown", "Execution"} {
		if !strings.Contains(out, want) {
			t.Errorf("the %q heading is missing, got:\n%s", want, out)
		}
	}
	// The teardown of what the previous run left behind always happens, and it is
	// its own phase before Execution. The teardown of this run's own changes must
	// not add a second one: on a clean run it has nothing to undo, and a heading
	// over nothing is worse than no heading. The two are named apart, so one can
	// never be mistaken for the other.
	if n := strings.Count(out, "teardown"); n != 1 {
		t.Errorf("a teardown heading appears %d times, want 1, got:\n%s", n, out)
	}
	if strings.Contains(out, "Final teardown") {
		t.Errorf("a clean run must not announce a final teardown, got:\n%s", out)
	}
	// And the phases have to be in the order they happen in.
	configuration := strings.Index(out, "Configuration")
	teardown := strings.Index(out, "Initial teardown")
	execution := strings.Index(out, "Execution")
	if !(configuration < teardown && teardown < execution) {
		t.Errorf("phases out of order, got:\n%s", out)
	}
	// The game is named exactly once, in the summary.
	if n := strings.Count(out, "age2"); n != 1 {
		t.Errorf("the game is named %d times, want 1, got:\n%s", n, out)
	}
}

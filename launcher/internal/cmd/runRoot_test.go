package cmd

import (
	"crypto/x509"
	"errors"
	"io"
	"net"
	"net/netip"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/spf13/pflag"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/cmd/bsManager"
	cmdServer "github.com/luskaner/ageLANServer/common/cmd/server"
	commonExecutor "github.com/luskaner/ageLANServer/common/executor/exec"
	"github.com/luskaner/ageLANServer/common/fileLock"
	"github.com/luskaner/ageLANServer/common/game/executor/base"
	"github.com/luskaner/ageLANServer/common/game/executor/custom"
	commonServer "github.com/luskaner/ageLANServer/common/server"
	"github.com/luskaner/ageLANServer/common/uuid"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/luskaner/ageLANServer/launcher/internal"
	"github.com/luskaner/ageLANServer/launcher/internal/executor"
	"github.com/luskaner/ageLANServer/launcher/internal/server"
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

func TestValidationCanTrustCertificateInvalid(t *testing.T) {
	exitCode := validateCanTrustCertificate("invalid-value")
	if exitCode != internal.ErrInvalidCanTrustCertificate {
		t.Errorf("expected exit code %d, got %d", internal.ErrInvalidCanTrustCertificate, exitCode)
	}
}

func TestValidationCanTrustCertificateValid(t *testing.T) {
	for _, val := range []string{"auto", "false", "local"} {
		exitCode := validateCanTrustCertificate(val)
		if exitCode != common.ErrSuccess {
			t.Errorf("for canTrustCertificate=%s, expected success, got %d", val, exitCode)
		}
	}
	if runtime.GOOS == "darwin" {
		if exitCode := validateCanTrustCertificate("user"); exitCode != common.ErrSuccess {
			t.Errorf("expected success for 'user' on darwin, got %d", exitCode)
		}
	}
	if exitCode := validateCanTrustCertificate("local"); exitCode != common.ErrSuccess {
		t.Errorf("expected success for 'local', got %d", exitCode)
	}
}

func TestValidationServerStartInvalid(t *testing.T) {
	exitCode := validateServerStartValue("invalid")
	if exitCode != internal.ErrInvalidServerStart {
		t.Errorf("expected exit code %d, got %d", internal.ErrInvalidServerStart, exitCode)
	}
}

func TestValidationServerStopInvalid(t *testing.T) {
	for _, stop := range []string{"auto", "true", "false"} {
		if exitCode := validateServerStopValue(stop, false); exitCode != common.ErrSuccess {
			t.Errorf("for serverStop=%s (non-admin), expected success, got %d", stop, exitCode)
		}
	}
	exitCode := validateServerStopValue("invalid-value", true)
	if exitCode != internal.ErrInvalidServerStop {
		t.Errorf("expected exit code %d, got %d", internal.ErrInvalidServerStop, exitCode)
	}
}

func TestValidationCanBroadcastBattleServer(t *testing.T) {
	if ec := validateCanBroadcastBattleServer("auto"); ec != common.ErrSuccess {
		t.Errorf("expected success for auto, got %d", ec)
	}
	if ec := validateCanBroadcastBattleServer("false"); ec != common.ErrSuccess {
		t.Errorf("expected success for false, got %d", ec)
	}
	if ec := validateCanBroadcastBattleServer("true"); ec != internal.ErrInvalidCanBroadcastBattleServer {
		t.Errorf("expected invalid for true, got %d", ec)
	}
	if ec := validateCanBroadcastBattleServer("bad"); ec != internal.ErrInvalidCanBroadcastBattleServer {
		t.Errorf("expected invalid for bad, got %d", ec)
	}
}

func TestValidationRequiredTrueFalse(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		field    string
		wantCode int
	}{
		{"valid true", "true", "Server.BattleServerManager.Run", common.ErrSuccess},
		{"valid required", "required", "Client.Isolation.Metadata", common.ErrSuccess},
		{"invalid", "bad", "Server.BattleServerManager.Run", internal.ErrInvalidServerBattleServerManagerRun},
		{"invalid metadata", "bad", "Client.Isolation.Metadata", internal.ErrInvalidIsolateMetadata},
		{"invalid profiles", "bad", "Client.Isolation.Profiles", internal.ErrInvalidIsolateProfiles},
	}
	for _, tt := range tests {
		ec := validateRequiredTrueFalse(tt.value, tt.field, requiredTrueFalseValues)
		if ec != tt.wantCode {
			t.Errorf("%s: expected %d, got %d", tt.name, tt.wantCode, ec)
		}
	}
}

func validLauncherConfig() *internal.Configuration {
	c := &internal.Configuration{}
	c.Config.Certificate.CanTrustInPc = "local"
	c.Config.CanBroadcastBattleServer = "auto"
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
	initConfigFn = func(fs *pflag.FlagSet) *internal.Configuration {
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
		mutate   func(c *internal.Configuration)
		wantCode int
	}{
		{"invalid canTrustCertificate", func(c *internal.Configuration) { c.Config.Certificate.CanTrustInPc = "bad" }, internal.ErrInvalidCanTrustCertificate},
		{"invalid serverStart", func(c *internal.Configuration) { c.Server.Start = "bad" }, internal.ErrInvalidServerStart},
		{"invalid serverStop", func(c *internal.Configuration) { c.Server.Stop = "bad" }, internal.ErrInvalidServerStop},
		{"invalid battleServerManagerRun", func(c *internal.Configuration) { c.Server.BattleServerManager.Run = "bad" }, internal.ErrInvalidServerBattleServerManagerRun},
		{"invalid isolateMetadata", func(c *internal.Configuration) { c.Client.Isolation.Metadata = "bad" }, internal.ErrInvalidIsolateMetadata},
		{"invalid isolateProfiles", func(c *internal.Configuration) { c.Client.Isolation.Profiles = "bad" }, internal.ErrInvalidIsolateProfiles},
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
			initConfigFn = func(fs *pflag.FlagSet) *internal.Configuration {
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
	initConfigFn = func(fs *pflag.FlagSet) *internal.Configuration {
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
	initConfigFn = func(fs *pflag.FlagSet) *internal.Configuration {
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
	if exitCode != internal.ErrInvalidServerArgs {
		t.Errorf("expected exit code %d for server args parse fail, got %d", internal.ErrInvalidServerArgs, exitCode)
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
	initConfigFn = func(fs *pflag.FlagSet) *internal.Configuration {
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
	if exitCode != internal.ErrInvalidSetupCommand {
		t.Errorf("expected exit code %d for setup command parse fail, got %d", internal.ErrInvalidSetupCommand, exitCode)
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
	initConfigFn = func(fs *pflag.FlagSet) *internal.Configuration {
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
	if exitCode != internal.ErrInvalidIsolationPath {
		t.Errorf("expected %d for invalid isolation path, got %d", internal.ErrInvalidIsolationPath, exitCode)
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
	initConfigFn = func(fs *pflag.FlagSet) *internal.Configuration {
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
	if exitCode != internal.ErrInvalidServerPath {
		t.Errorf("expected %d for invalid server path, got %d", internal.ErrInvalidServerPath, exitCode)
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
	initConfigFn = func(fs *pflag.FlagSet) *internal.Configuration {
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
	if exitCode != internal.ErrGameLauncherNotFound {
		t.Errorf("expected %d for game launcher not found, got %d", internal.ErrGameLauncherNotFound, exitCode)
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
	initConfigFn = func(fs *pflag.FlagSet) *internal.Configuration {
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
	if exitCode != internal.ErrGameAlreadyRunning {
		t.Errorf("expected %d for game already running, got %d", internal.ErrGameAlreadyRunning, exitCode)
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
	initConfigFn = func(fs *pflag.FlagSet) *internal.Configuration {
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
	gameId                    string
	cfg                       func() *internal.Configuration
	isAdmin                   bool
	gameSupported             bool
	makeExec                  base.Executor
	isolationPath             string
	processFn                 func(string) (string, *os.Process, error)
	gameRunning               bool
	openLog                   error
	bufferFn                  func(string, func(io.Writer)) error
	killAgent                 func()
	newPidLock                func() fileLock.Locker
	parseCommandArgsFn        func([]string, map[string]string) ([]string, error)
	resolveIsolateValueFnVal  func(string, bool) bool
	configSetGameIdFnVal      func(string)
	configIsolationPathFnVal  func(base.Executor) string
	commonParsePathFnVal     func([]string, map[string]string) (os.FileInfo, string, error)
	commonEnhancedViperFnVal func(string) []string
	configNativeMacOsGameFnVal func(base.Executor, bool) bool
	configBattleServerRequiredFnVal func(base.Executor) bool
	newConfigFlushCacheOptionsFnVal func(bool, string, bool, bool) *executor.ConfigFlushCacheOptions
	discoverServersFnVal     func(string, bool, mapset.Set[netip.Addr], mapset.Set[uint16]) (uuid.UUID, net.IP)
	netipParseAddrFnVal      func(string) (netip.Addr, error)
	serverFilterServerIPsFnVal func(uuid.UUID, string, string, mapset.Set[netip.Addr]) (uuid.UUID, []server.MesuredIpAddress, *commonServer.AnnounceMessageDataSupportedLatest)
	serverGetExecutablePathFnVal func(string) string
	serverGenerateCertsFnVal func(string, bool) int
	configRunBattleServerManagerFnVal func(string, *pflag.FlagSet, *bsManager.StartValues, bool) int
	bsManagerStartFlagSetFnVal func([]string) (*bsManager.StartValues, *pflag.FlagSet)
	configStartServerFnVal   func(string, *pflag.FlagSet, *cmdServer.Values, bool) (int, string)
	serverReadCACertFnVal    func(string) *x509.Certificate
	configMapHostsFnVal      func(string, string, bool, bool, bool) int
	configAddCertFnVal       func(string, uuid.UUID, *x509.Certificate, string, bool, bool) int
	configIsolateUserDataFnVal func(bool, bool, string) int
	configAddCACertToGameFnVal func(string, uuid.UUID, *x509.Certificate, string, string, bool, bool) int
	configLaunchAgentAndGameFnVal func(base.Executor, custom.Exec, []string, string, string, string) int
	uuidParseFnVal           func(string) (uuid.UUID, error)
	uuidMustParseFnVal       func(string) uuid.UUID
	uuidNilFnVal             func() uuid.UUID
	executablesNativeFileNameFnVal func(bool, string) string
	configRunSetupCommandFnVal func([]string) *commonExecutor.Result
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

	gameId = o.gameId
	cfgFile = ""
	gameCfgFile = ""
	if o.newPidLock != nil {
		newPidLockFn = o.newPidLock
	} else {
		newPidLockFn = func() fileLock.Locker { return &fakePidLocker{} }
	}
	if o.cfg != nil {
		initConfigFn = func(fs *pflag.FlagSet) *internal.Configuration { return o.cfg() }
	} else {
		initConfigFn = func(fs *pflag.FlagSet) *internal.Configuration { return validLauncherConfig() }
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
		discoverServersFn = func(string, bool, mapset.Set[netip.Addr], mapset.Set[uint16]) (uuid.UUID, net.IP) { return uuid.Nil(), nil }
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
		configStartServerFn = func(s string, fs *pflag.FlagSet, v *cmdServer.Values, b bool) (int, string) { return common.ErrSuccess, "127.0.0.1" }
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
		configAddCertFn = func(s1 string, u uuid.UUID, c *x509.Certificate, s2 string, b1, b2 bool) int { return common.ErrSuccess }
	}
	if o.configIsolateUserDataFnVal != nil {
		configIsolateUserDataFn = o.configIsolateUserDataFnVal
	} else {
		configIsolateUserDataFn = func(b1, b2 bool, s string) int { return common.ErrSuccess }
	}
	if o.configAddCACertToGameFnVal != nil {
		configAddCACertToGameFn = o.configAddCACertToGameFnVal
	} else {
		configAddCACertToGameFn = func(s1 string, u uuid.UUID, c *x509.Certificate, s2, s3 string, b1, b2 bool) int { return common.ErrSuccess }
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

	return func() {
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
		gameId:       "age2",
		isAdmin:      false,
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
		gameId:       "age2",
		isAdmin:      false,
		gameSupported: true,
		cfg: func() *internal.Configuration {
			c := validLauncherConfig()
			c.Server.AnnounceMulticastGroups = []string{"not-a-multicast"}
			return c
		},
	})
	defer restore()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	_, exitCode := runRoot(fs)
	if exitCode != internal.ErrAnnouncementMulticastGroup {
		t.Errorf("expected %d for invalid multicast group, got %d", internal.ErrAnnouncementMulticastGroup, exitCode)
	}
}

func TestRunRootServerFoundByDiscovery(t *testing.T) {
	discoveredUUID := uuid.New()
	restore := applyOverrides(t, runRootOverrides{
		gameId:       "age2",
		isAdmin:      false,
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
		gameId:       "age2",
		isAdmin:      false,
		gameSupported: true,
		cfg: func() *internal.Configuration {
			c := validLauncherConfig()
			c.Server.Start = "false"
			c.Server.Host = ""
			return c
		},
	})
	defer restore()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	_, exitCode := runRoot(fs)
	if exitCode != internal.ErrInvalidServerHost {
		t.Errorf("expected %d for empty serverHost, got %d", internal.ErrInvalidServerHost, exitCode)
	}
}

func TestRunRootServerHostIPv6(t *testing.T) {
	restore := applyOverrides(t, runRootOverrides{
		gameId:       "age2",
		isAdmin:      false,
		gameSupported: true,
		cfg: func() *internal.Configuration {
			c := validLauncherConfig()
			c.Server.Start = "false"
			c.Server.Host = "::1"
			return c
		},
	})
	defer restore()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	_, exitCode := runRoot(fs)
	if exitCode != internal.ErrInvalidServerHost {
		t.Errorf("expected %d for IPv6 serverHost, got %d", internal.ErrInvalidServerHost, exitCode)
	}
}

func TestRunRootServerHostResolutionFailure(t *testing.T) {
	restore := applyOverrides(t, runRootOverrides{
		gameId:       "age2",
		isAdmin:      false,
		gameSupported: true,
		cfg: func() *internal.Configuration {
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
	if exitCode != internal.ErrInvalidServerHost {
		t.Errorf("expected %d for server host resolution failure, got %d", internal.ErrInvalidServerHost, exitCode)
	}
}

func TestRunRootServerExecutableNotFound(t *testing.T) {
	restore := applyOverrides(t, runRootOverrides{
		gameId:       "age2",
		isAdmin:      false,
		gameSupported: true,
		serverGetExecutablePathFnVal: func(s string) string { return "" },
	})
	defer restore()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	_, exitCode := runRoot(fs)
	if exitCode != internal.ErrServerExecutable {
		t.Errorf("expected %d for server executable not found, got %d", internal.ErrServerExecutable, exitCode)
	}
}

func TestRunRootReadCertFailure(t *testing.T) {
	restore := applyOverrides(t, runRootOverrides{
		gameId:       "age2",
		isAdmin:      false,
		gameSupported: true,
		serverReadCACertFnVal: func(host string) *x509.Certificate { return nil },
	})
	defer restore()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	_, exitCode := runRoot(fs)
	if exitCode != internal.ErrReadCert {
		t.Errorf("expected %d for read cert failure, got %d", internal.ErrReadCert, exitCode)
	}
}

func TestRunRootMapHostsFailure(t *testing.T) {
	restore := applyOverrides(t, runRootOverrides{
		gameId:       "age2",
		isAdmin:      false,
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
		gameId:       "age2",
		isAdmin:      false,
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
		gameId:       "age2",
		isAdmin:      false,
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
		gameId:       "age2",
		isAdmin:      false,
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
		gameId:       "age2",
		isAdmin:      false,
		gameSupported: true,
		cfg: func() *internal.Configuration {
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
	if exitCode != internal.ErrInvalidServerBattleServerManagerArgs {
		t.Errorf("expected %d for battle server manager parse failure, got %d", internal.ErrInvalidServerBattleServerManagerArgs, exitCode)
	}
}

func TestRunRootLaunchAgentSuccess(t *testing.T) {
	restore := applyOverrides(t, runRootOverrides{
		gameId:       "age1",
		isAdmin:      false,
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
		gameId:       "age2",
		isAdmin:      false,
		gameSupported: true,
		cfg: func() *internal.Configuration {
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
		gameId:       "age2",
		isAdmin:      false,
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

func TestRunRootCanTrustCertificateAuto(t *testing.T) {
	restore := applyOverrides(t, runRootOverrides{
		gameId:       "age2",
		isAdmin:      false,
		gameSupported: true,
		cfg: func() *internal.Configuration {
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

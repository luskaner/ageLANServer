package cmd

import (
	"errors"
	"fmt"
	"io"
	"net/netip"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/luskaner/ageLANServer/common/uuid"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/knadh/koanf/parsers/toml/v2"
	"github.com/knadh/koanf/v2"
	"github.com/luskaner/ageLANServer/common/cmd/bsManager"
	cmdServer "github.com/luskaner/ageLANServer/common/cmd/server"
	"github.com/luskaner/ageLANServer/common/game"
	"github.com/luskaner/ageLANServer/common/game/cert"
	gameExecutor "github.com/luskaner/ageLANServer/common/game/executor"
	"github.com/luskaner/ageLANServer/common/game/executor/custom"
	"github.com/spf13/pflag"

	"github.com/luskaner/ageLANServer/common"
	commonCmd "github.com/luskaner/ageLANServer/common/cmd"
	"github.com/luskaner/ageLANServer/common/executables"
	commonExecutor "github.com/luskaner/ageLANServer/common/executor"
	"github.com/luskaner/ageLANServer/common/executor/exec"
	"github.com/luskaner/ageLANServer/common/fileLock"
	commonLogger "github.com/luskaner/ageLANServer/common/logger"
	"github.com/luskaner/ageLANServer/common/paths"
	commonProcess "github.com/luskaner/ageLANServer/common/process"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/luskaner/ageLANServer/launcher-common/launcher"
	"github.com/luskaner/ageLANServer/launcher-common/launcher/cmdUtils"
	"github.com/luskaner/ageLANServer/launcher-common/launcher/cmdUtils/logger"
	"github.com/luskaner/ageLANServer/launcher-common/launcher/executor"
	"github.com/luskaner/ageLANServer/launcher-common/launcher/server"
	"github.com/luskaner/ageLANServer/launcher-common/ui"
	"github.com/luskaner/ageLANServer/launcher/internal/dialog"
)

const autoValue = "auto"
const trueValue = "true"
const falseValue = "false"

var configPaths = []string{paths.ResourcesDir, "."}
var config = &cmdUtils.Config{}

var (
	Version      string
	cfgFile      string
	gameCfgFile  string
	gameId       string
	output       string
	filesToPrint []string
	// usedConfigFile is the main config file initConfig ended up loading, kept
	// next to filesToPrint so the session summary can name it.
	usedConfigFile                 string
	autoTrueFalseValues            = mapset.NewThreadUnsafeSet[string](autoValue, trueValue, falseValue)
	canTrustCertificateValues      = mapset.NewThreadUnsafeSet[string](autoValue, falseValue, "user", "local")
	canBroadcastBattleServerValues = mapset.NewThreadUnsafeSet[string](autoValue, falseValue)
	requiredTrueFalseValues        = mapset.NewThreadUnsafeSet[string](trueValue, falseValue, "required")
)

var (
	initConfigFn      = initConfig
	newPidLockFn      = func() fileLock.Locker { return &fileLock.PidLock{} }
	isAdminFn         = func() bool { return commonExecutor.IsAdmin() }
	chdirToExeFn      = common.ChdirToExe
	openMainLogFn     = logger.OpenMainFileLog
	printFileFn       = logger.PrintFile
	writeFileLogFn    = logger.WriteFileLog
	dnsConnectivityFn = common.DNSConnectivity
)

var (
	gameSupportedGamesContainsOneFn = func(gameId string) bool { return game.SupportedGames.ContainsOne(gameId) }
	parseCommandArgsFn              = cmdUtils.ParseCommandArgs
	resolveIsolateValueFn           = cmdUtils.ResolveIsolateValue
	configSetGameIdFn               = config.SetGameId
	configIsolationPathFn           = config.IsolationPath
	configGamePathToGameCertPathFn  = config.GamePathToGameCertPath
	configNativeMacOsGameFn         = config.NativeMacOsGame
	configBattleServerRequiredFn    = config.BattleServerRequired
	makeExecFn                      = gameExecutor.MakeExec
	commonParsePathFn               = common.ParsePath
	commonEnhancedViperFn           = common.EnhancedViperStringToStringSlice
	executablesFindPathFn           = executables.FindPath
	commonProcessProcessFn          = commonProcess.Process
	commonProcessWaitForProcessFn   = commonProcess.WaitForProcess
	gameRunningFn                   = cmdUtils.GameRunning
	configKillAgentFn               = config.KillAgent
	launcherCommonConfigRevertFn    = launcherCommon.ConfigRevert
	executorRunRevertFn             = executor.RunRevert
	commonLoggerFileLoggerBufferFn  = func(name string, fn func(io.Writer)) error {
		if commonLogger.FileLogger == nil {
			return nil
		}
		return commonLogger.FileLogger.Buffer(name, fn)
	}
	newConfigFlushCacheOptionsFn    = executor.NewConfigFlushCacheOptions
	executablesNativeFileNameFn     = executables.NativeFileName
	configRunSetupCommandFn         = config.RunSetupCommand
	netipParseAddrFn                = netip.ParseAddr
	discoverServersFn               = cmdUtils.DiscoverServersAndSelectBestIpAddr
	serverGetExecutablePathFn       = server.GetExecutablePath
	serverGenerateCertsFn           = server.GenerateServerCertificates
	configRunBattleServerManagerFn  = config.RunBattleServerManager
	configStartServerFn             = config.StartServer
	serverReadCACertFn              = server.ReadCACertificateFromServer
	configMapHostsFn                = config.MapHosts
	configAddCertFn                 = config.AddCert
	configIsolateUserDataFn         = config.IsolateUserData
	configAddCACertToGameFn         = config.AddCACertToGame
	configLaunchAgentAndGameFn      = config.LaunchAgentAndGame
	configRunStopAgentFn            = config.RunStopAgent
	serverFilterServerIPsFn         = server.FilterServerIPs
	bsManagerStartFlagSetFn         = bsManager.StartFlagSet
	commonHostOrIpToIpsFn           = common.HostOrIpToIps
	commonStringSliceToNetIPSliceFn = common.StringSliceToNetIPSlice
	commonNetIPSliceToNetIPSetFn    = common.NetIPSliceToNetIPSet
	uuidParseFn                     = uuid.Parse
	uuidMustParseFn                 = uuid.MustParse
	uuidNilFn                       = uuid.Nil
	signalNotifyFn                  = signal.Notify
	osExitFn                        = os.Exit
	// runRoot reinstalls the dialog with dialog.Set, so a test cannot inject a
	// fake Dialog through dialog.Set alone: it has to own the whole backend.
	dialogNewFn = dialog.New
)

func Execute() (err error, exitCode int) {
	singleFs := commonCmd.NewSingleFlagSet(runRoot, Version)
	fs := singleFs.Fs()
	fs.StringVar(&cfgFile, "config", "", fmt.Sprintf(`config file (default config.toml in %s directories)`, strings.Join(configPaths, ", ")))
	fs.StringVar(&gameCfgFile, "gameConfig", "", fmt.Sprintf(`Game config file (default config.game.toml in %s directories)`, strings.Join(configPaths, ", ")))
	fs.StringP("dialog", "d", autoValue, `Whether to ask the interactive questions (which server to use, and whether to start one) in a graphical window instead of in the console, "auto" uses graphical dialogs if they are available in the system, "false" always asks in the console. It always falls back to the console if graphical dialogs are unavailable.`)
	// Registered here rather than in common/cmd/addDefaultFlags because that one
	// is shared with the modules outside this change's scope.
	fs.StringVar(&output, "output", autoValue, `How much to decorate the console output with, "auto" picks whatever the terminal can show without turning characters into boxes, "color" asks for colour but still degrades to plain text if the console cannot do it, "ascii" guarantees plain ASCII text. It also reads the AGE_LANSERVER_OUTPUT environment variable.`)
	fs.Bool("log", false, "Whether to log more info to a file. Enable it for errors.")
	fs.Bool("internet", true, "Whether or not the launcher may use the internet to look up official domains. If false, only the statically known domains will be used for the hosts file, certificates and logs. If true and there is no internet connectivity the launcher will still work with the statically known domains.")
	fs.StringP("canAddHost", "t", "true", "Add a local dns entry if it's needed to connect to the server with the official domain. Including to avoid receiving that it's on maintenance. Ignored if clientExeArgs contains '{HostFilePath}'. Will require admin privileges.")
	canTrustCertificateStr := `Trust the certificate of the server if needed. "false"`
	if runtime.GOOS != "linux" {
		canTrustCertificateStr += `, "user"`
	}
	canTrustCertificateStr += ` or local (will require admin privileges). Ignored if clientExeArgs contains '{CertFilePath}'.`
	fs.StringP("canTrustCertificate", "c", "local", canTrustCertificateStr)
	if runtime.GOOS == "windows" {
		fs.StringP("canBroadcastBattleServer", "b", "auto", `Whether or not to broadcast the game BattleServer to all interfaces in LAN (not just the most priority one)`)
	}
	var pathNamesInfo string
	if runtime.GOOS == "windows" {
		pathNamesInfo += " Path names need to use double backslashes within single quotes or be within double quotes."
	}
	commonCmd.GameVarCommand(fs, &gameId)
	fs.StringP("isolateMetadata", "m", "required", "Isolate the metadata cache of the game, otherwise, it will be shared. Not compatible with AoE:DE. If 'required' it will resolve to 'true' if using the official launcher, 'false' otherwise.")
	fs.StringP("isolateProfiles", "p", "required", "Isolate the user's profile of the game, otherwise, it will be shared. If 'required' it will resolve to 'true' if using the official launcher, 'false' otherwise.")
	fs.String("setupCommand", "", `Executable to run (including arguments) to run first after the "Setting up..." line. The command must return a 0 exit code to continue. If you need to keep it running spawn a new separate process. You may use environment variables.`+pathNamesInfo)
	fs.String("revertCommand", "", `Executable to run (including arguments) to run after setupCommand, game has exited and everything has been reverted. It may run before if there is an error. You may use environment variables.`+pathNamesInfo)
	fs.StringP("serverStart", "a", "auto", `Start the server if needed, "auto" will start a server if one is not already running, "true" (will start a server regardless if one is already running), "false" (will require an already running server).`)
	fs.StringP("serverStop", "o", "auto", `Stop the server if started, "auto" will stop the server if one was started, "false" (will not stop the server regardless if one was started), "true" (will not stop the server even if it was started).`)
	fs.StringSliceP("serverAnnouncePorts", "n", []string{strconv.Itoa(common.AnnouncePort)}, `Announce ports to listen to. If not including the default port, default configured 'servers' will not get discovered.`)
	fs.StringSliceP("serverAnnounceMulticastGroups", "g", []string{common.AnnounceMulticastGroup}, `Announce multicast groups to join. If not including the default group, default configured 'servers' will not get discovered via Multicast.`)
	fs.StringP("server", "s", "", `Hostname of the server to connect to. If not absent, serverStart will be assumed to be false. Ignored otherwise`)
	fs.Bool("serverSingleAutoSelect", false, `Auto-select the server when a single one is discovered.`)
	serverExe := executables.NativeFileName(false, executables.Server)
	fs.StringP("serverPath", "z", "auto", fmt.Sprintf(`The executable path of the server, "auto", will be try to execute in this order "./%s/%s", "../%s" and finally "../%s/%s", otherwise set the path (relative or absolute).`, executables.Server, serverExe, serverExe, executables.Server, serverExe))
	fs.StringP("serverPathArgs", "r", "", `The arguments to pass to the server executable if starting it. Execute the server help flag for available arguments. You may use environment variables.`+pathNamesInfo)
	clientExeTip := `The type of game client or the path. "auto" will use `
	if runtime.GOOS != "darwin" {
		clientExeTip += "Steam"
	}
	unixClientExeTip := `Steam (CrossOver) and then the Steam (Wine) one if found`
	if runtime.GOOS == "windows" {
		clientExeTip += ` and then the Xbox one if found`
	}
	if runtime.GOOS == "linux" {
		clientExeTip += `, `
	}
	if runtime.GOOS != "windows" {
		clientExeTip += unixClientExeTip
	}
	clientExeTip += `. Use a path to the game launcher,`
	if runtime.GOOS != "darwin" {
		clientExeTip += ` "steam"`
	}
	if runtime.GOOS == "linux" {
		clientExeTip += `,`
	}
	if runtime.GOOS != "windows" {
		clientExeTip += ` "steam_crossover" or "steam_wine"`
	}
	if runtime.GOOS == "windows" {
		clientExeTip += ` or "msstore"`
	}
	clientExeTip += " to use the default launcher."
	fs.StringP("clientExe", "l", "auto", clientExeTip)
	fs.StringP("clientExeArgs", "i", "", "The arguments to pass to the client launcher if it is custom. You may use environment variables and '{HostFilePath}'/'{CertFilePath}' replacement variables."+pathNamesInfo)

	// Default values & bindings will be handled in initConfig
	return singleFs.Execute()
}

func runRoot(fs *pflag.FlagSet) (err error, exitCode int) {
	// validate required flags
	if gameId == "" {
		return errors.New("required flag 'game' not set"), common.ErrSyntax
	}
	// The explicit flag gets the last word over AGE_LANSERVER_OUTPUT, and this is
	// the first point where the parsed flags are available.
	ui.ApplyOverride(output)

	lock := newPidLockFn()
	if err = lock.Lock(); err != nil {
		logger.Fail("Failed to lock pid file. Kill process launcher if it is running in your task manager.")
		logger.Detail("%s", err.Error())
		exitCode = common.ErrPidLock
		return
	}
	cfg := initConfigFn(fs)
	logger.LogEnabled = cfg.Config.Log
	if err = openMainLogFn(gameId); err != nil {
		logger.Fail("Failed to open file log")
		logger.Detail("%s", err.Error())
		exitCode = common.ErrFileLog
		return
	}
	if !cfg.Config.CanUseInternet {
		launcher.CanUseInternet = false
		logger.Info("Internet usage is disabled via config.")
	} else {
		launcher.CanUseInternet = dnsConnectivityFn()
	}
	if !launcher.CanUseInternet {
		logger.Warn("No internet connectivity, some features will fallback gracefully.")
	}
	common.SetUseInternet(launcher.CanUseInternet)
	for _, fileToPrint := range filesToPrint {
		printFileFn("config", fileToPrint)
	}
	var atomicExitCode atomic.Int32
	atomicExitCode.Store(int32(common.ErrSuccess))
	// Teardown runs at most once and never concurrently. The signal handler
	// below and this deferred cleanup could both decide to tear down; unguarded
	// they raced, spawning two config.exe reverts over one hosts lock and two
	// stop-agent clients on one named pipe, which is one way an elevated
	// config-admin-agent outlived the launcher. The mutex also makes the
	// second caller wait instead of returning, so the signal handler's
	// os.Exit cannot truncate a revert that is still in flight.
	var teardownMutex sync.Mutex
	teardownRan := false
	teardown := func(force bool) {
		teardownMutex.Lock()
		defer teardownMutex.Unlock()
		if teardownRan {
			return
		}
		teardownRan = true
		// A taskbar button left half filled after the program is gone is worse than
		// no button at all, and this is the one path every exit goes through.
		ui.ClearProgress()
		// force comes from the signal handler, where the user asked to stop
		// and the exit code says nothing about how far setup got.
		//
		// The heading is printed only when there is something to undo. On a clean run
		// each step reverted its own changes as it finished, so all that is left here
		// is closing a log and releasing a lock: a heading over that announces work
		// that is not happening, which is worse than no heading at all.
		if force || atomicExitCode.Load() != int32(common.ErrSuccess) {
			if config.HasTeardownWork() {
				ui.Section("Final teardown")
			}
			config.Revert()
		}
		logger.WriteFileLog(gameId, "before exit")
		commonLogger.CloseFileLog()
		_ = lock.Unlock()
	}
	defer func() {
		if r := recover(); r != nil {
			logger.Fail("%s", r)
			logger.Detail("%s", string(debug.Stack()))
			atomicExitCode.Store(int32(common.ErrGeneral))
		}
		teardown(false)
		exitCode = int(atomicExitCode.Load())
	}()
	if ec := validateDialogValue(cfg.Config.Dialog); ec != common.ErrSuccess {
		atomicExitCode.Store(int32(ec))
		return
	}
	dialogResolution := dialogNewFn(cfg.Config.Dialog)
	dialog.SetOutput(dialog.Output{Println: logger.Println, Printf: logger.Printf})
	dialog.Set(dialogResolution.Dialog)
	// Registered after the teardown defer so, by LIFO, it runs before it.
	defer dialog.Reset()
	// ui.Banner prints nothing when the console width is unknown, so a redirected
	// run or a pipe never gets a banner in the middle of its output.
	ui.Banner(executables.Launcher, Version)
	// One place for the three facts the run depends on. They used to be announced
	// once here and once again further down, which is how a mismatch between the
	// two copies of the same fact looked like a bug.
	ui.Section("Configuration")
	ui.KV(0, "main config file", orNone(usedConfigFile))
	ui.KV(0, "game config file", orNone(gameCfgFile))
	ui.KV(0, "game", gameId)
	// Printed only when the console could show more than it is showing, because
	// the fallback is a code page and not a decision: without this line the ASCII
	// markers look like the intended output rather than like what they are.
	if hint := ui.UpgradeHint(); hint != "" {
		logger.Info("%s", hint)
	}
	// The backend in use is not worth a line of its own: a graphical dialog is
	// visible, and a console one answers on the terminal. Only the fallback is
	// news, because it means the configured choice was not honoured.
	if dialogResolution.Reason != "" {
		logger.Warn("%s", dialogResolution.Reason)
	}
	writeFileLogFn(gameId, "start")
	isAdmin := isAdminFn()
	canTrustCertificate := cfg.Config.Certificate.CanTrustInPc
	if ec := validateCanTrustCertificate(canTrustCertificate); ec != common.ErrSuccess {
		atomicExitCode.Store(int32(ec))
		return
	}
	canBroadcastBattleServer := "false"
	if runtime.GOOS == "windows" && (gameId != game.AoM && gameId != game.AoE4) {
		canBroadcastBattleServer = cfg.Config.CanBroadcastBattleServer
		if ec := validateCanBroadcastBattleServer(canBroadcastBattleServer); ec != common.ErrSuccess {
			atomicExitCode.Store(int32(ec))
			return
		}
	}
	serverStart := cfg.Server.Start
	if ec := validateServerStartValue(serverStart); ec != common.ErrSuccess {
		atomicExitCode.Store(int32(ec))
		return
	}
	serverStop := cfg.Server.Stop
	if ec := validateServerStopValue(serverStop, runtime.GOOS != "windows" && isAdmin); ec != common.ErrSuccess {
		atomicExitCode.Store(int32(ec))
		return
	}
	battleServerManagerRun := cfg.Server.BattleServerManager.Run
	if ec := validateRequiredTrueFalse(battleServerManagerRun, "Server.BattleServerManager.Run", requiredTrueFalseValues); ec != common.ErrSuccess {
		atomicExitCode.Store(int32(ec))
		return
	}
	isolateMetadataStr := cfg.Client.Isolation.Metadata
	if ec := validateRequiredTrueFalse(isolateMetadataStr, "Client.Isolation.Metadata", requiredTrueFalseValues); ec != common.ErrSuccess {
		atomicExitCode.Store(int32(ec))
		return
	}
	isolateProfilesStr := cfg.Client.Isolation.Profiles
	if ec := validateRequiredTrueFalse(isolateProfilesStr, "Client.Isolation.Profiles", requiredTrueFalseValues); ec != common.ErrSuccess {
		atomicExitCode.Store(int32(ec))
		return
	}
	if !gameSupportedGamesContainsOneFn(gameId) {
		logger.Fail("Invalid game type")
		atomicExitCode.Store(int32(launcherCommon.ErrInvalidGame))
		return
	}
	configSetGameIdFn(gameId)
	serverValues := map[string]string{
		"Game": gameId,
		"Id":   uuid.New().String(),
	}
	var serverArgsValues *cmdServer.Values
	var serverFlags *pflag.FlagSet
	var serverArgs []string
	if serverArgs, err = parseCommandArgsFn(cfg.Server.Args, serverValues); err == nil {
		var serverSingleFlagSet *commonCmd.SingleFlagSet
		serverArgsValues, serverSingleFlagSet = cmdServer.SingleFlagSet("", nil, nil)
		serverFlags = serverSingleFlagSet.Fs()
		if err = serverFlags.Parse(serverArgs); err != nil {
			logger.Fail("Failed to parse server executable arguments")
			atomicExitCode.Store(int32(launcher.ErrInvalidServerArgs))
			return
		}
		if _, err = uuidParseFn(serverArgsValues.Id); err != nil {
			logger.Fail("You must provide a valid UUID for the server ID using the --id argument in server executable arguments")
			atomicExitCode.Store(int32(launcher.ErrInvalidServerArgs))
			return
		}
	} else {
		logger.Fail("Failed to parse server executable arguments")
		atomicExitCode.Store(int32(launcher.ErrInvalidServerArgs))
		return
	}
	var battleServerManagerArgs []string
	battleServerManagerArgs, err = parseCommandArgsFn(
		cfg.Server.BattleServerManager.Args,
		serverValues,
	)
	if err != nil {
		logger.Fail("Failed to parse battle-server-manager executable arguments")
		atomicExitCode.Store(int32(launcher.ErrInvalidServerBattleServerManagerArgs))
		return
	}
	var setupCommand []string
	setupCommand, err = parseCommandArgsFn(cfg.Config.SetupCommand, nil)
	if err != nil {
		logger.Fail("Failed to parse setup command")
		atomicExitCode.Store(int32(launcher.ErrInvalidSetupCommand))
		return
	}
	var revertCommand []string
	revertCommand, err = parseCommandArgsFn(cfg.Config.RevertCommand, nil)
	if err != nil {
		logger.Fail("Failed to parse revert command")
		atomicExitCode.Store(int32(launcher.ErrInvalidRevertCommand))
		return
	}
	canAddHost := cfg.Config.CanAddHost
	clientExecutable := cfg.Client.Executable.Path
	if clientExecutable == "steam" && runtime.GOOS == "darwin" && gameId != game.AoE2 {
		logger.Fail("Only AoE 2: DE is supported on 'steam'. Use 'steam_crossover' or 'steam_wine' instead.")
		atomicExitCode.Store(int32(launcher.ErrGameUnsupportedLauncherCombo))
		return
	}
	var clientExecutableOfficial bool
	if clientExecutable == "auto" || clientExecutable == "steam" {
		clientExecutableOfficial = true
	} else if runtime.GOOS == "windows" {
		clientExecutableOfficial = clientExecutable == "msstore"
	} else if clientExecutable == "steam_wine" || clientExecutable == "steam_crossover" {
		clientExecutableOfficial = true
	}
	var isolateMetadata bool
	if gameId != game.AoE1 {
		isolateMetadata = resolveIsolateValueFn(isolateMetadataStr, clientExecutableOfficial)
	}
	isolateProfiles := resolveIsolateValueFn(isolateProfilesStr, clientExecutableOfficial)
	isolation := isolateMetadata || isolateProfiles
	var isolationPath string
	if isolation {
		if cfg.Client.Isolation.Path != "auto" {
			var isolationDir os.FileInfo
			if isolationDir, isolationPath, err = commonParsePathFn(commonEnhancedViperFn(cfg.Client.Isolation.Path), nil); err != nil || !isolationDir.IsDir() {
				logger.Fail("Invalid isolation path")
				atomicExitCode.Store(int32(launcher.ErrInvalidIsolationPath))
				return
			}
			logger.SetBasePath(isolationPath)
			logger.WriteFileLog(gameId, "post isolation path")
		} else if runtime.GOOS != "windows" && !clientExecutableOfficial {
			logger.Fail("You must set the Client.Isolation.Path as you are using a custom launcher with isolation.")
			atomicExitCode.Store(int32(launcher.ErrInvalidIsolationPath))
			return
		}
	}
	var serverExecutable string
	if serverExecutable = cfg.Server.Executable.Path; serverExecutable != "auto" {
		var serverFile os.FileInfo
		if serverFile, serverExecutable, err = commonParsePathFn(commonEnhancedViperFn(cfg.Server.Executable.Path), nil); err != nil || serverFile.IsDir() {
			logger.Fail("Invalid server executable")
			atomicExitCode.Store(int32(launcher.ErrInvalidServerPath))
			return
		}
	}
	var battleServerManagerExecutable string
	if battleServerManagerExecutable = cfg.Server.BattleServerManager.Executable.Path; battleServerManagerExecutable != "auto" {
		var battleServerManagerFile os.FileInfo
		if battleServerManagerFile, battleServerManagerExecutable, err = commonParsePathFn(commonEnhancedViperFn(cfg.Server.BattleServerManager.Executable.Path), nil); err != nil || battleServerManagerFile.IsDir() {
			logger.Fail("Invalid battle-server-manager executable")
			atomicExitCode.Store(int32(launcher.ErrInvalidClientPath))
			return
		}
	}
	if !clientExecutableOfficial {
		var clientFile os.FileInfo
		if clientFile, clientExecutable, err = commonParsePathFn(commonEnhancedViperFn(cfg.Client.Executable.Path), nil); err != nil || clientFile.IsDir() {
			logger.Fail("Invalid client executable")
			atomicExitCode.Store(int32(launcher.ErrInvalidClientPath))
			return
		}
	} else if !isolateProfiles || (gameId != game.AoE1 && !isolateMetadata) {
		logger.Fail("Isolating profiles and metadata is a must when using an official launcher.")
		atomicExitCode.Store(int32(launcher.ErrRequiredIsolation))
		return
	} else {
		logger.Warn("Make sure you disable the cloud saves in the launcher settings to avoid issues.")
	}

	if isAdmin {
		logger.Warn("Running as administrator, this is not recommended for security reasons. It will request isolated admin privileges if/when needed.")
		if runtime.GOOS != "windows" {
			logger.Detail("It can also cause issues and restrict the functionality.")
		}
	}

	serverHost := cfg.Server.Host

	if clientExecutable == "msstore" && gameId == game.AoM {
		logger.Fail("The Microsoft Store (Xbox) version is not supported on this game.")
		atomicExitCode.Store(int32(launcher.ErrGameUnsupportedLauncherCombo))
		return
	}
	configSetGameIdFn(gameId)
	// Finding the game is the first thing that takes long enough to look broken,
	// so it gets an in place line and the terminal's own progress indicator rather
	// than a "Step" the reader has to trust is still running.
	gameSearch := ui.Start("Looking for the game...")
	gameProgress := ui.BeginProgress()
	var gamePath string
	executer := makeExecFn(gameId, clientExecutable)
	if executer != nil {
		gameProgress.Done()
		gameSearch.Done("Game found on %s.", executer.String())
	} else {
		gameProgress.Fail()
		gameSearch.Fail("Game not found.")
		atomicExitCode.Store(int32(launcher.ErrGameLauncherNotFound))
		return
	}
	if isolation && isolationPath == "" {
		if isolationPath = configIsolationPathFn(executer); isolationPath == "" {
			logger.Fail("Failed to auto retrieve isolation path")
			atomicExitCode.Store(int32(launcher.ErrInvalidIsolationPath))
			return
		}
		logger.SetBasePath(isolationPath)
		logger.WriteFileLog(gameId, "post isolation path")
	}
	var customExecutor custom.Exec
	var ok bool
	if customExecutor, ok = executer.(custom.Exec); ok {
		if cert.HasCA(gameId) {
			var clientFile os.FileInfo
			var clientPath string
			if clientFile, clientPath, err = commonParsePathFn(commonEnhancedViperFn(cfg.Client.Path), nil); err != nil || !clientFile.IsDir() {
				logger.Fail("Invalid client path")
				atomicExitCode.Store(int32(launcher.ErrInvalidClientPath))
				return
			}
			gamePath = clientPath
		}
	} else if cert.HasCA(gameId) {
		gamePath = executer.(game.Locatable).Path()
	}
	var gameCaCertPath string
	if gamePath != "" {
		_, caCert := cert.NewCA(gameId, configGamePathToGameCertPathFn(executer, gamePath))
		gameCaCertPath = caCert.OriginalPath()
		if commonLogger.FileLogger != nil {
			logger.SetCacert(&caCert)
		}
	}
	macOsExclusiveMappings := configNativeMacOsGameFn(executer, true)
	if commonLogger.FileLogger != nil {
		logger.SetMacOsExclusiveMappings(macOsExclusiveMappings)
	}
	sigs := make(chan os.Signal, 1)
	signalNotifyFn(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		_, sigOk := <-sigs
		if sigOk {
			teardown(true)
			osExitFn(int(atomicExitCode.Load()))
		}
	}()
	agentWaitDuration := time.Minute
	agent := executablesNativeFileNameFn(false, executables.LauncherAgent)
	if _, proc, localErr := commonProcessProcessFn(agent); localErr == nil && proc != nil {
		logger.Step("agent is running, waiting up to %s for it to end...", agentWaitDuration)
		if !commonProcessWaitForProcessFn(proc, &agentWaitDuration) {
			logger.Warn("agent did not exit on its own.")
		}
	}
	if _, proc, localErr := commonProcessProcessFn(executablesNativeFileNameFn(false, executables.LauncherConfigAdminAgent)); localErr == nil && proc != nil {
		// It never exits on its own: only an explicit 'Exit' sent over the IPC
		// pipe stops it. This used to just wait for it, which stalled every
		// launch for the full timeout and then left the elevated process in
		// place, still holding the mapped-ips/cert state of the previous
		// session so the next setUp failed with 'already mapped'. Ask it to stop.
		logger.Step("config-admin-agent from a previous run is still active, stopping it...")
		if result := configRunStopAgentFn(); result.Success() {
			logger.Ok("config-admin-agent stopped.")
		} else {
			logger.Fail("Failed to stop config-admin-agent.")
			if result.Err != nil {
				logger.Fault("Error message: %s", result.Err.Error())
			}
			if result.ExitCode != common.ErrSuccess {
				logger.Fault("Exit code: %s", strconv.Itoa(result.ExitCode))
			}
		}
	}
	if gameRunningFn() {
		atomicExitCode.Store(int32(launcher.ErrGameAlreadyRunning))
		return
	}
	/*
		Ensure:
		* No running config-admin-agent nor agent processes
		* Any previous changes are reverted
	*/
	// This is a teardown, not a setup step: what it undoes is what the previous
	// run left behind, not anything this one has done. Labelling it as execution
	// work is what made "Cleaning up" read as part of the startup, and calling
	// both of them "Teardown" is what made the two look like a bug.
	ui.Section("Initial teardown")
	logger.Step("Cleaning up (if needed)...")
	configKillAgentFn()
	if err = commonLoggerFileLoggerBufferFn("config_revert_initial", func(writer io.Writer) {
		launcherCommon.ConfigRevert("", commonLogger.FileLogger.Folder(), false, writer, func(options *exec.Options) {
			commonLogger.Println("run config revert", options.String())
		}, executor.RunRevert)
	}); err != nil {
		atomicExitCode.Store(int32(common.ErrFileLog))
		return
	}
	if canTrustCertificate == "auto" {
		if runtime.GOOS == "darwin" {
			canTrustCertificate = "user"
		} else {
			canTrustCertificate = "local"
		}
	}
	customHostFile := slices.ContainsFunc(cfg.Client.Args, func(s string) bool {
		return strings.Contains(s, "{HostFilePath}")
	})
	customCertFile := slices.ContainsFunc(cfg.Client.Args, func(s string) bool {
		return strings.Contains(s, "{CertFilePath}")
	})
	cfgFlushCacheOpts := newConfigFlushCacheOptionsFn(canAddHost, canTrustCertificate, customHostFile, customCertFile)
	if cfgFlushCacheOpts != nil {
		if result := cfgFlushCacheOpts.RunFlushCache(); !result.Success() {
			atomicExitCode.Store(int32(launcher.ErrFlushCache))
			return
		}
	}
	if _, proc, localErr := commonProcessProcessFn(executablesNativeFileNameFn(false, executables.Server)); localErr == nil && proc != nil {
		logger.Warn("Server is already running, If you did not start it manually, kill the server process using the task manager and execute the launcher again.")
	}
	if err = commonLoggerFileLoggerBufferFn("revert_command_initial", func(writer io.Writer) {
		if err = executor.RunRevertCommand(writer, func(options *exec.Options) {
			commonLogger.Println("run revert command", options.String())
		}); err != nil {
			logger.Fail("Failed to run revert command.")
			logger.Fault("Error message: %s", err.Error())
		}
	}); err != nil {
		atomicExitCode.Store(int32(common.ErrFileLog))
		return
	}
	logger.WriteFileLog(gameId, "post initial cleanup")
	if len(revertCommand) > 0 {
		if err = launcherCommon.RevertCommandStore.Store(revertCommand); err != nil {
			logger.Fail("Failed to store revert command")
			atomicExitCode.Store(int32(launcher.ErrInvalidRevertCommand))
			return
		}
	}
	// Everything from here on is this run's own work: the previous run's leftovers
	// are gone, and what follows is what the game will actually see.
	ui.Section("Execution")
	// Setup
	logger.Step("Setting up...")
	if len(setupCommand) > 0 {
		logger.Step("Running setup command '%s' and waiting for it to exit...", cfg.Config.SetupCommand)
		result := configRunSetupCommandFn(setupCommand)
		if !result.Success() {
			if result.Err != nil {
				logger.Fault("Error: %s", result.Err)
			}
			if result.ExitCode != common.ErrSuccess {
				logger.Fault("Exit code: %d.", result.ExitCode)
			}
			atomicExitCode.Store(int32(launcher.ErrSetupCommand))
			return
		}
	}
	var serverIP string
	if serverStart == "auto" {
		announcePorts := cfg.Server.AnnouncePorts
		ports := mapset.NewThreadUnsafeSetWithSize[uint16](len(announcePorts))
		for _, portInt := range announcePorts {
			ports.Add(uint16(portInt))
		}
		multicastIPsStr := cfg.Server.AnnounceMulticastGroups
		multicastIPs := mapset.NewThreadUnsafeSetWithSize[netip.Addr](len(multicastIPsStr))
		for _, str := range multicastIPsStr {
			if IP, localErr := netipParseAddrFn(str); localErr == nil && IP.Is4() && IP.IsMulticast() {
				multicastIPs.Add(IP)
			} else {
				logger.Fail("Invalid multicast group \"%s\"", str)
				atomicExitCode.Store(int32(launcher.ErrAnnouncementMulticastGroup))
				return
			}
		}
		serverId, selectedServerIp := discoverServersFn(
			gameId,
			cfg.Server.SingleAutoSelect,
			multicastIPs,
			ports,
		)
		if serverId != uuidNilFn() {
			serverIP = selectedServerIp.String()
			serverStart = "false"
			serverArgsValues.Id = serverId.String()
			if serverStop == "auto" && (!isAdmin || runtime.GOOS == "windows") {
				serverStop = "false"
			}
		} else {
			serverStart = "true"
			if serverStop == "auto" {
				serverStop = "true"
			}
		}
	}
	if serverStart == "false" {
		if serverStop == "true" {
			logger.Warn("serverStart is false. Ignoring serverStop being true.")
		}
		if serverIP == "" {
			if serverHost == "" {
				logger.Fail("serverStart is false. serverHost must be fulfilled as it is needed to know which host to connect to.")
				atomicExitCode.Store(int32(launcher.ErrInvalidServerHost))
				return
			}
			if addr, localErr := netipParseAddrFn(serverHost); localErr == nil && addr.Is6() {
				logger.Fail("serverStart is false. serverHost must be fulfilled with a host or Ipv4 address.")
				atomicExitCode.Store(int32(launcher.ErrInvalidServerHost))
				return
			}
			if id, measuredServerIPAddrs, data := serverFilterServerIPsFn(
				uuidNilFn(),
				serverHost,
				gameId,
				commonNetIPSliceToNetIPSetFn(commonStringSliceToNetIPSliceFn(commonHostOrIpToIpsFn(serverHost))),
			); data == nil {
				logger.Fail("serverStart is false. Failed to resolve serverHost to a valid and reachable IP.")
				atomicExitCode.Store(int32(launcher.ErrInvalidServerHost))
				return
			} else {
				serverIP = measuredServerIPAddrs[0].Ip.String()
				serverArgsValues.Id = id.String()
			}
		}
	} else {
		if logRoot := commonLogger.FileLogger.Folder(); logRoot != "" {
			serverArgsValues.Log = true
			serverArgsValues.LogRoot = logRoot
			serverArgsValues.Flatlog = true
			serverArgsValues.Deterministic = true
		}
		battleServerRequired := configBattleServerRequiredFn(executer)
		if battleServerManagerRun == "false" && battleServerRequired {
			logger.Warn("This game needs a Battle Server to be started but you don't allow to start one, make sure you have one running and the server configured.")
		}
		runBattleServerManager := battleServerManagerRun == "true" || (battleServerManagerRun == "required" && battleServerRequired)
		if cfg.Server.Start == autoValue && !cfg.Server.StartWithoutConfirmation {
			str := "No servers were found, proceeding to"
			if runBattleServerManager {
				str += " start a battle server (if needed) and then"
			}
			str += " start the server."
			if !dialog.Active().ConfirmStartServer(str, os.Stdin) {
				logger.Fail("Canceled starting the server.")
				atomicExitCode.Store(int32(launcher.ErrServerStartCanceled))
				return
			}
		}
		serverExecutablePath := serverGetExecutablePathFn(serverExecutable)
		if serverExecutablePath == "" {
			logger.Fail("Cannot find server executable path. Set it manually in Server.Executable.")
			atomicExitCode.Store(int32(launcher.ErrServerExecutable))
			return
		}
		if serverExecutable != serverExecutablePath {
			logger.Ok("Found server executable path: %s", serverExecutablePath)
		}
		if ec := serverGenerateCertsFn(serverExecutablePath, canTrustCertificate != "false"); ec != common.ErrSuccess {
			atomicExitCode.Store(int32(ec))
			return
		}
		if runBattleServerManager {
			values, flags := bsManagerStartFlagSetFn(nil)
			if err = flags.Parse(battleServerManagerArgs); err != nil {
				logger.Fail("Failed to parse battle-server-manager executable arguments")
				atomicExitCode.Store(int32(launcher.ErrInvalidServerBattleServerManagerArgs))
				return
			}
			ec := configRunBattleServerManagerFn(
				battleServerManagerExecutable,
				flags,
				values,
				serverStop == "true",
			)
			if ec != common.ErrSuccess {
				atomicExitCode.Store(int32(ec))
				return
			}
		}
		var ec int
		ec, serverIP = configStartServerFn(serverExecutablePath, serverFlags, serverArgsValues, serverStop == "true")
		if ec != common.ErrSuccess {
			atomicExitCode.Store(int32(ec))
			return
		}
	}
	serverCertificate := serverReadCACertFn(serverIP)
	if serverCertificate == nil {
		logger.Fail("Failed to read certificate from %s. Try to access it with your browser and checking the certificate.", serverIP)
		atomicExitCode.Store(int32(launcher.ErrReadCert))
		return
	}
	atomicExitCode.Store(int32(configMapHostsFn(gameId, serverIP, macOsExclusiveMappings, canAddHost, customHostFile)))
	if atomicExitCode.Load() != int32(common.ErrSuccess) {
		return
	}
	logger.WriteFileLog(gameId, "post host mapping")
	atomicExitCode.Store(int32(configAddCertFn(gameId, uuidMustParseFn(serverArgsValues.Id), serverCertificate, canTrustCertificate, customCertFile, macOsExclusiveMappings)))
	if atomicExitCode.Load() != int32(common.ErrSuccess) {
		return
	}
	logger.WriteFileLog(gameId, "post add cert")
	atomicExitCode.Store(int32(configIsolateUserDataFn(isolateMetadata, isolateProfiles, isolationPath)))
	if atomicExitCode.Load() != int32(common.ErrSuccess) {
		return
	}
	logger.WriteFileLog(gameId, "post isolate user data")
	if gamePath != "" {
		atomicExitCode.Store(int32(configAddCACertToGameFn(gameId, uuidMustParseFn(serverArgsValues.Id), serverCertificate, configGamePathToGameCertPathFn(executer, gamePath), gameCaCertPath, cfg.Config.Certificate.CanTrustInGame, macOsExclusiveMappings)))
		if atomicExitCode.Load() != int32(common.ErrSuccess) {
			return
		}
		logger.WriteFileLog(gameId, "post add game cert")
	}
	atomicExitCode.Store(int32(configLaunchAgentAndGameFn(executer, customExecutor, cfg.Client.Args, canTrustCertificate, canBroadcastBattleServer, isolationPath)))
	return
}

func initConfig(fs *pflag.FlagSet) *launcher.Configuration {
	k := koanf.New(".")
	defaults := map[string]any{
		"Config.Dialog":                             autoValue,
		"Config.CanAddHost":                         "true",
		"Config.Certificate.CanTrustInPc":           "local",
		"Config.Certificate.CanTrustInGame":         true,
		"Config.CanBroadcastBattleServer":           "auto",
		"Config.CanUseInternet":                     true,
		"Config.Log":                                false,
		"Client.Isolation.Metadata":                 "required",
		"Client.Isolation.Profiles":                 "required",
		"Config.SetupCommand":                       []string{},
		"Config.RevertCommand":                      []string{},
		"Client.Executable":                         "auto",
		"Client.ExecutableArgs":                     []string{},
		"Client.Path":                               "auto",
		"Server.Start":                              "auto",
		"Server.Stop":                               "auto",
		"Server.SingleAutoSelect":                   false,
		"Server.StartWithoutConfirmation":           false,
		"Server.Executable":                         "auto",
		"Server.ExecutableArgs":                     []string{"-e", "{Game}", "--id", "{Id}"},
		"Server.Host":                               netip.IPv4Unspecified().String(),
		"Server.AnnouncePorts":                      []int{common.AnnouncePort},
		"Server.AnnounceMulticastGroups":            []string{common.AnnounceMulticastGroup},
		"Server.BattleServerManager.Run":            "true",
		"Server.BattleServerManager.Executable":     "auto",
		"Server.BattleServerManager.ExecutableArgs": []string{"-e", "{Game}", "-r"},
	}
	for g := range game.SupportedGames.Iter() {
		defaults[fmt.Sprintf("Games.%s.Hosts", g)] = []string{netip.IPv4Unspecified().String()}
	}
	bindings := map[string]string{
		"dialog":                        "Config.Dialog",
		"canAddHost":                    "Config.CanAddHost",
		"canTrustCertificate":           "Config.Certificate.CanTrustInPc",
		"canBroadcastBattleServer":      "Config.CanBroadcastBattleServer",
		"internet":                      "Config.CanUseInternet",
		"log":                           "Config.Log",
		"isolateMetadata":               "Client.Isolation.Metadata",
		"isolateProfiles":               "Client.Isolation.Profiles",
		"setupCommand":                  "Config.SetupCommand",
		"revertCommand":                 "Config.RevertCommand",
		"serverStart":                   "Server.Start",
		"serverStop":                    "Server.Stop",
		"serverSingleAutoSelect":        "Server.SingleAutoSelect",
		"serverAnnouncePorts":           "Server.AnnouncePorts",
		"serverAnnounceMulticastGroups": "Server.AnnounceMulticastGroups",
		"server":                        "Server.Host",
		"serverPath":                    "Server.Executable",
		"serverPathArgs":                "Server.ExecutableArgs",
		"clientExe":                     "Client.Executable",
		"clientExeArgs":                 "Client.ExecutableArgs",
	}
	var mainfileCandidates []string
	if cfgFile != "" {
		mainfileCandidates = append(mainfileCandidates, cfgFile)
	} else {
		for _, configPath := range configPaths {
			mainfileCandidates = append(mainfileCandidates, filepath.Join(configPath, "config.toml"))
		}
	}
	usedFile := common.LoadKoanfLayersOrExit(k, defaults, mainfileCandidates, toml.Parser(), fs, bindings, executables.Launcher, commonLogger.Println)
	usedConfigFile = usedFile
	if cfgFile != "" && usedFile == "" {
		logger.Warn("No config file found, using defaults.")
	}
	if usedFile != "" {
		// Not printed here: the summary at the top of the run already names this
		// file, and two copies of the same path is one more thing to keep in sync.
		filesToPrint = append(filesToPrint, usedFile)
	}
	var gameFileCandidates []string
	if gameCfgFile != "" {
		gameFileCandidates = append(gameFileCandidates, gameCfgFile)
	} else {
		for _, configPath := range configPaths {
			gameFileCandidates = append(gameFileCandidates, filepath.Join(configPath, fmt.Sprintf("config.%s.toml", gameId)))
		}
	}
	var err error
	if gameCfgFile, err = common.LoadKoanfLayers(k, map[string]any{}, gameFileCandidates, toml.Parser(), fs, nil, executables.Launcher); err == nil {
		filesToPrint = append(filesToPrint, gameCfgFile)
	} else {
		if _, ok := errors.AsType[*common.KoanfFileLoadError](err); !ok {
			logger.Fail("Error parsing game config file: %s:%s", gameCfgFile, err.Error())
			os.Exit(launcher.ErrGameConfigParse)
		}
	}
	var c launcher.Configuration
	if err := k.Unmarshal("", &c); err != nil {
		logger.Fail("unable to decode configuration: %v", err)
		os.Exit(common.ErrConfigParse)
	}
	return &c
}

// orNone names an optional path in the session summary.
func orNone(path string) string {
	if path == "" {
		return "none"
	}
	return path
}

func validateDialogValue(dialogMode string) (exitCode int) {
	if !autoTrueFalseValues.Contains(dialogMode) {
		logger.Fail("Invalid value for dialog (auto/true/false): %s", dialogMode)
		return launcher.ErrInvalidDialog
	}
	return common.ErrSuccess
}

func validateCanTrustCertificate(canTrustCertificate string) (exitCode int) {
	validValues := mapset.NewThreadUnsafeSet[string](autoValue, falseValue, "user", "local")
	if runtime.GOOS == "linux" {
		validValues.Remove("user")
	}
	if !validValues.Contains(canTrustCertificate) {
		logger.Fail("Invalid value for canTrustCertificate (%s): %s", strings.Join(validValues.ToSlice(), "/"), canTrustCertificate)
		return launcher.ErrInvalidCanTrustCertificate
	}
	return common.ErrSuccess
}

func validateCanBroadcastBattleServer(canBroadcastBattleServer string) (exitCode int) {
	if !canBroadcastBattleServerValues.Contains(canBroadcastBattleServer) {
		logger.Fail("Invalid value for canBroadcastBattleServer (auto/false): %s", canBroadcastBattleServer)
		return launcher.ErrInvalidCanBroadcastBattleServer
	}
	return common.ErrSuccess
}

func validateServerStartValue(serverStart string) (exitCode int) {
	if !autoTrueFalseValues.Contains(serverStart) {
		logger.Fail("Invalid value for serverStart (auto/true/false): %s", serverStart)
		return launcher.ErrInvalidServerStart
	}
	return common.ErrSuccess
}

func validateServerStopValue(serverStop string, nonWindowsAdmin bool) (exitCode int) {
	validValues := mapset.NewThreadUnsafeSet[string](autoValue, trueValue, falseValue)
	if nonWindowsAdmin {
		validValues.Remove(falseValue)
	}
	if !validValues.Contains(serverStop) {
		logger.Fail("Invalid value for serverStop (%s): %s", strings.Join(validValues.ToSlice(), "/"), serverStop)
		return launcher.ErrInvalidServerStop
	}
	return common.ErrSuccess
}

func validateRequiredTrueFalse(value string, name string, validValues mapset.Set[string]) (exitCode int) {
	if !validValues.Contains(value) {
		logger.Fail("Invalid value for %s (%s): %s", name, strings.Join(validValues.ToSlice(), "/"), value)
		switch name {
		case "Server.BattleServerManager.Run":
			return launcher.ErrInvalidServerBattleServerManagerRun
		case "Client.Isolation.Metadata":
			return launcher.ErrInvalidIsolateMetadata
		case "Client.Isolation.Profiles":
			return launcher.ErrInvalidIsolateProfiles
		}
	}
	return common.ErrSuccess
}

package cmd

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/netip"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"slices"
	"strconv"
	"strings"
	"syscall"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/cmd"
	"github.com/luskaner/ageLANServer/common/executables"
	commonExecutor "github.com/luskaner/ageLANServer/common/executor"
	"github.com/luskaner/ageLANServer/common/executor/exec"
	"github.com/luskaner/ageLANServer/common/fileLock"
	commonLogger "github.com/luskaner/ageLANServer/common/logger"
	commonProcess "github.com/luskaner/ageLANServer/common/process"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/luskaner/ageLANServer/launcher-common/cert"
	"github.com/luskaner/ageLANServer/launcher/internal"
	"github.com/luskaner/ageLANServer/launcher/internal/cmdUtils"
	"github.com/luskaner/ageLANServer/launcher/internal/cmdUtils/logger"
	"github.com/luskaner/ageLANServer/launcher/internal/executor"
	gameExecutor "github.com/luskaner/ageLANServer/launcher/internal/game/executor"
	"github.com/luskaner/ageLANServer/launcher/internal/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const autoValue = "auto"
const trueValue = "true"
const falseValue = "false"

var configPaths = []string{"resources", "."}
var config = &cmdUtils.Config{}

var (
	Version                        string
	cfgFile                        string
	gameCfgFile                    string
	gameId                         string
	autoTrueFalseValues            = mapset.NewThreadUnsafeSet[string](autoValue, trueValue, falseValue)
	canTrustCertificateValues      = mapset.NewThreadUnsafeSet[string](falseValue, "user", "local")
	canBroadcastBattleServerValues = mapset.NewThreadUnsafeSet[string](autoValue, falseValue)
	requiredTrueFalseValues        = mapset.NewThreadUnsafeSet[string](trueValue, falseValue, "required")
	rootCmd                        = &cobra.Command{
		Use:   filepath.Base(os.Args[0]),
		Short: "launcher discovers and configures AoE: DE, AoE 2: DE and AoE 3: DE, and AoM: RT to connect to the local LAN 'server'",
		Long:  "launcher discovers or starts a local LAN 'server', configures and executes the game launcher to connect to it",
		Run: func(_ *cobra.Command, _ []string) {
			lock := &fileLock.PidLock{}
			if err := lock.Lock(); err != nil {
				logger.Println("Failed to lock pid file. Kill process 'launcher' if it is running in your task manager.")
				logger.Println(err.Error())
				os.Exit(common.ErrPidLock)
			}
			if err := logger.OpenMainFileLog(gameId); err != nil {
				logger.Println("Failed to open file log")
				logger.Println(err.Error())
				os.Exit(common.ErrFileLog)
			}
			var errorCode = common.ErrSuccess
			defer func() {
				if r := recover(); r != nil {
					logger.Println(r)
					logger.Println(string(debug.Stack()))
					errorCode = common.ErrGeneral
				}
				if errorCode != common.ErrSuccess {
					config.Revert()
				}
				logger.WriteFileLog(gameId, "before exit")
				commonLogger.CloseFileLog()
				_ = lock.Unlock()
				os.Exit(errorCode)
			}()
			logger.WriteFileLog(gameId, "start")
			initConfig()
			isAdmin := commonExecutor.IsAdmin()
			canTrustCertificate := viper.GetString("Config.CanTrustCertificate")
			if runtime.GOOS != "windows" {
				canTrustCertificateValues.Remove("user")
			}
			if !canTrustCertificateValues.Contains(canTrustCertificate) {
				logger.Printf("Invalid value for canTrustCertificate (%s): %s\n", strings.Join(canTrustCertificateValues.ToSlice(), "/"), canTrustCertificate)
				errorCode = internal.ErrInvalidCanTrustCertificate
				return
			}
			canBroadcastBattleServer := "false"
			if runtime.GOOS == "windows" && gameId != common.GameAoM {
				canBroadcastBattleServer = viper.GetString("Config.CanBroadcastBattleServer")
				if !canBroadcastBattleServerValues.Contains(canBroadcastBattleServer) {
					logger.Printf("Invalid value for canBroadcastBattleServer (auto/false): %s\n", canBroadcastBattleServer)
					errorCode = internal.ErrInvalidCanBroadcastBattleServer
					return
				}
			}
			serverStart := viper.GetString("Server.Start")
			if !autoTrueFalseValues.Contains(serverStart) {
				logger.Printf("Invalid value for serverStart (auto/true/false): %s\n", serverStart)
				errorCode = internal.ErrInvalidServerStart
				return
			}
			serverStop := viper.GetString("Server.Stop")
			if runtime.GOOS != "windows" && isAdmin {
				autoTrueFalseValues.Remove(falseValue)
			}
			if !autoTrueFalseValues.Contains(serverStop) {
				logger.Printf("Invalid value for serverStop (%s): %s\n", strings.Join(autoTrueFalseValues.ToSlice(), "/"), serverStop)
				errorCode = internal.ErrInvalidServerStop
				return
			}
			battleServerManagerRun := viper.GetString("Server.BattleServerManager.Run")
			if !requiredTrueFalseValues.Contains(battleServerManagerRun) {
				logger.Printf("Invalid value for Server.BattleServerManager.Run (%s): %s\n", strings.Join(requiredTrueFalseValues.ToSlice(), "/"), battleServerManagerRun)
				errorCode = internal.ErrInvalidServerBattleServerManagerRun
				return
			}
			isolateMetadataStr := viper.GetString("Config.IsolateMetadata")
			if !requiredTrueFalseValues.Contains(isolateMetadataStr) {
				logger.Printf("Invalid value for Config.IsolateMetadata (%s): %s\n", strings.Join(requiredTrueFalseValues.ToSlice(), "/"), isolateMetadataStr)
				errorCode = internal.ErrInvalidIsolateMetadata
				return
			}
			isolateProfilesStr := viper.GetString("Config.IsolateProfiles")
			if !requiredTrueFalseValues.Contains(isolateProfilesStr) {
				logger.Printf("Invalid value for Config.IsolateProfiles (%s): %s\n", strings.Join(requiredTrueFalseValues.ToSlice(), "/"), isolateProfilesStr)
				errorCode = internal.ErrInvalidIsolateProfiles
				return
			}
			if !common.SupportedGames.ContainsOne(gameId) {
				logger.Println("Invalid game type")
				errorCode = launcherCommon.ErrInvalidGame
				return
			}
			config.SetGameId(gameId)
			serverValues := map[string]string{
				"Game": gameId,
				"Id":   uuid.NewString(),
			}
			serverArgs, err := common.ParseCommandArgs("Server.ExecutableArgs", serverValues, true)
			serverId := uuid.Nil
			if err == nil {
				// Find the actual ID in case the user missed it or passed another one
				for i, arg := range serverArgs {
					if arg == "--id" && i+1 < len(serverArgs) {
						if id, err := uuid.Parse(serverArgs[i+1]); err == nil {
							serverId = id
						}
						break
					}
				}
				if serverId == uuid.Nil {
					logger.Println("You must provide a valid UUID for the server ID using the '--id' argument in 'server' executable arguments")
					errorCode = internal.ErrInvalidServerArgs
					return
				}
			} else {
				logger.Println("Failed to parse 'server' executable arguments")
				errorCode = internal.ErrInvalidServerArgs
				return
			}
			var battleServerManagerArgs []string
			battleServerManagerArgs, err = common.ParseCommandArgs(
				"Server.BattleServerManager.ExecutableArgs",
				serverValues,
				true,
			)
			if err != nil {
				logger.Println("Failed to parse 'battle-server-manager' executable arguments")
				errorCode = internal.ErrInvalidServerBattleServerManagerArgs
				return
			}
			var setupCommand []string
			setupCommand, err = common.ParseCommandArgs("Config.SetupCommand", nil, true)
			if err != nil {
				logger.Println("Failed to parse setup command")
				errorCode = internal.ErrInvalidSetupCommand
				return
			}
			var revertCommand []string
			revertCommand, err = common.ParseCommandArgs("Config.RevertCommand", nil, true)
			if err != nil {
				logger.Println("Failed to parse revert command")
				errorCode = internal.ErrInvalidRevertCommand
				return
			}
			canAddHost := viper.GetBool("Config.CanAddHost")
			var clientExecutable string
			var clientExecutableOfficial bool
			if clientExecutable = viper.GetString("Client.Executable"); clientExecutable == "auto" || clientExecutable == "steam" || clientExecutable == "msstore" {
				clientExecutableOfficial = true
			}
			var isolateMetadata bool
			if gameId != common.GameAoE1 {
				isolateMetadata = cmdUtils.ResolveIsolateValue(isolateMetadataStr, clientExecutableOfficial)
			}
			isolateProfiles := cmdUtils.ResolveIsolateValue(isolateProfilesStr, clientExecutableOfficial)
			var serverExecutable string
			if serverExecutable = viper.GetString("Server.Executable"); serverExecutable != "auto" {
				var serverFile os.FileInfo
				if serverFile, serverExecutable, err = common.ParsePath(viper.GetStringSlice("Server.Executable"), nil); err != nil || serverFile.IsDir() {
					logger.Println("Invalid 'server' executable")
					errorCode = internal.ErrInvalidServerPath
					return
				}
			}
			var battleServerManagerExecutable string
			if battleServerManagerExecutable = viper.GetString("Server.BattleServerManager.Executable"); battleServerManagerExecutable != "auto" {
				var battleServerManagerFile os.FileInfo
				if battleServerManagerFile, battleServerManagerExecutable, err = common.ParsePath(viper.GetStringSlice("Server.BattleServerManager.Executable"), nil); err != nil || battleServerManagerFile.IsDir() {
					logger.Println("Invalid 'battle-server-manager' executable")
					errorCode = internal.ErrInvalidClientPath
					return
				}
			}
			if !clientExecutableOfficial {
				var clientFile os.FileInfo
				if clientFile, clientExecutable, err = common.ParsePath(viper.GetStringSlice("Client.Executable"), nil); err != nil || clientFile.IsDir() {
					logger.Println("Invalid client executable")
					errorCode = internal.ErrInvalidClientPath
					return
				}
			} else if !isolateProfiles || (gameId != common.GameAoE && !isolateMetadata) {
				logger.Println("Isolating profiles and metadata is a must when using an official launcher.")
				errorCode = internal.ErrRequiredIsolation
				return
			} else {
				logger.Println("Make sure you disable the cloud saves in the launcher settings to avoid issues.")
			}

			serverHost := viper.GetString("Server.Host")

			logger.Printf("Game %s.\n", gameId)
			if clientExecutable == "msstore" && gameId == common.GameAoM {
				logger.Println("The Microsoft Store (Xbox) version of AoM: RT is not supported.")
				errorCode = internal.ErrGameUnsupportedLauncherCombo
				return
			}

			logger.Println("Looking for the game...")
			var gamePath string
			executer := gameExecutor.MakeExec(gameId, clientExecutable)
			var customExecutor gameExecutor.CustomExec
			switch executer.(type) {
			case gameExecutor.SteamExec:
				logger.Println("Game found on Steam.")
				if gameId != common.GameAoE1 {
					gamePath = executer.(gameExecutor.SteamExec).GamePath()
				}
			case gameExecutor.XboxExec:
				logger.Println("Game found on Xbox.")
				if gameId != common.GameAoE1 {
					gamePath = executer.(gameExecutor.XboxExec).GamePath()
				}
			case gameExecutor.CustomExec:
				customExecutor = executer.(gameExecutor.CustomExec)
				logger.Println("Game found on custom path.")
				if runtime.GOOS == "linux" {
					if isolateMetadata {
						logger.Println("Isolating metadata is not supported.")
						isolateMetadata = false
					}
					if isolateProfiles {
						logger.Println("Isolating profiles is not supported.")
						isolateProfiles = false
					}
				}
				if gameId != common.GameAoE1 {
					if clientFile, clientPath, err := common.ParsePath(viper.GetStringSlice("Client.Path"), nil); err != nil || !clientFile.IsDir() {
						logger.Println("Invalid client path")
						errorCode = internal.ErrInvalidClientPath
						return
					} else {
						gamePath = clientPath
					}
				}
			default:
				logger.Println("Game not found.")
				errorCode = internal.ErrGameLauncherNotFound
				return
			}
			if gamePath != "" && commonLogger.FileLogger != nil {
				caCert := cert.NewCA(gameId, gamePath)
				logger.Cacert = &caCert
			}

			if isAdmin {
				logger.Println("Running as administrator, this is not recommended for security reasons. It will request isolated admin privileges if/when needed.")
				if runtime.GOOS != "windows" {
					logger.Println(" It can also cause issues and restrict the functionality.")
				}
			}

			if runtime.GOOS != "windows" && isAdmin && (clientExecutable == "auto" || clientExecutable == "steam") {
				logger.Println("Steam cannot be run as administrator. Either run this as a normal user o set Client.Executable to a custom launcher.")
				errorCode = internal.ErrSteamRoot
				return
			}

			if cmdUtils.GameRunning() {
				errorCode = internal.ErrGameAlreadyRunning
				return
			}

			config.SetGameId(gameId)

			sigs := make(chan os.Signal, 1)
			signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
			go func() {
				_, ok := <-sigs
				if ok {
					config.Revert()
					commonLogger.CloseFileLog()
					_ = lock.Unlock()
					os.Exit(errorCode)
				}
			}()
			/*
				Ensure:
				* No running config-admin-agent nor agent processes
				* Any previous changes are reverted
			*/
			logger.Println("Cleaning up (if needed)...")
			config.KillAgent()
			if err = commonLogger.FileLogger.Buffer("config_revert_initial", func(writer io.Writer) {
				launcherCommon.ConfigRevert(gameId, commonLogger.FileLogger.Folder(), false, writer, func(options exec.Options) {
					commonLogger.Println("run config revert", options.String())
				}, executor.RunRevert)
			}); err != nil {
				errorCode = common.ErrFileLog
				return
			}
			var proc *os.Process
			_, proc, err = commonProcess.Process(executables.Filename(false, executables.Server))
			if err == nil && proc != nil {
				logger.Println("'Server' is already running, If you did not start it manually, kill the 'server' process using the task manager and execute the 'launcher' again.")
			}
			if err = commonLogger.FileLogger.Buffer("revert_command_initial", func(writer io.Writer) {
				if err = executor.RunRevertCommand(writer, func(options exec.Options) {
					commonLogger.Println("run revert command", options.String())
				}); err != nil {
					logger.Println("Failed to run revert command.")
					logger.Println("Error message: " + err.Error())
				}
			}); err != nil {
				errorCode = common.ErrFileLog
				return
			}
			logger.WriteFileLog(gameId, "post initial cleanup")
			if len(revertCommand) > 0 {
				if err := launcherCommon.RevertCommandStore.Store(revertCommand); err != nil {
					logger.Println("Failed to store revert command")
					errorCode = internal.ErrInvalidRevertCommand
					return
				}
			}
			// Setup
			logger.Println("Setting up...")
			if len(setupCommand) > 0 {
				logger.Printf("Running setup command '%s' and waiting for it to exit...\n", viper.GetString("Config.SetupCommand"))
				result := config.RunSetupCommand(setupCommand)
				if !result.Success() {
					if result.Err != nil {
						logger.Printf("Error: %s\n", result.Err)
					}
					if result.ExitCode != common.ErrSuccess {
						logger.Printf(`Exit code: %d.`+"\n", result.ExitCode)
					}
					errorCode = internal.ErrSetupCommand
					return
				}
			}
			var serverIP string
			if serverStart == "auto" {
				announcePorts := viper.GetIntSlice("Server.AnnouncePorts")
				ports := mapset.NewThreadUnsafeSetWithSize[uint16](len(announcePorts))
				for _, portInt := range announcePorts {
					ports.Add(uint16(portInt))
				}
				multicastIPsStr := viper.GetStringSlice("Server.AnnounceMulticastGroups")
				multicastIPs := mapset.NewThreadUnsafeSetWithSize[netip.Addr](len(multicastIPsStr))
				for _, str := range multicastIPsStr {
					if IP, err := netip.ParseAddr(str); err == nil && IP.Is4() && IP.IsMulticast() {
						multicastIPs.Add(IP)
					} else {
						logger.Printf("Invalid multicast group \"%s\"\n", str)
						errorCode = internal.ErrAnnouncementMulticastGroup
						return
					}
				}
				var selectedServerIp net.IP
				serverId, selectedServerIp = cmdUtils.DiscoverServersAndSelectBestIpAddr(
					gameId,
					viper.GetBool("Server.SingleAutoSelect"),
					multicastIPs,
					ports,
				)
				if serverId != uuid.Nil {
					serverIP = selectedServerIp.String()
					serverStart = "false"
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
					logger.Println("serverStart is false. Ignoring serverStop being true.")
				}
				if serverIP == "" {
					if serverHost == "" {
						logger.Println("serverStart is false. serverHost must be fulfilled as it is needed to know which host to connect to.")
						errorCode = internal.ErrInvalidServerHost
						return
					}
					if addr, err := netip.ParseAddr(serverHost); err == nil && addr.Is6() {
						logger.Println("serverStart is false. serverHost must be fulfilled with a host or Ipv4 address.")
						errorCode = internal.ErrInvalidServerHost
						return
					}
					if id, measuredServerIPAddrs, data := server.FilterServerIPs(
						uuid.Nil,
						serverHost,
						gameId,
						common.NetIPSliceToNetIPSet(common.StringSliceToNetIPSlice(common.HostOrIpToIps(serverHost))),
					); data == nil {
						logger.Println("serverStart is false. Failed to resolve serverHost to a valid and reachable IP.")
						errorCode = internal.ErrInvalidServerHost
						return
					} else {
						serverIP = measuredServerIPAddrs[0].Ip.String()
						serverId = id
					}
				}
			} else {
				if logRoot := commonLogger.FileLogger.Folder(); logRoot != "" {
					if !slices.Contains(serverArgs, "--log") {
						serverArgs = append(serverArgs, "--log")
					}
					if !slices.Contains(serverArgs, "--logRoot") {
						serverArgs = append(serverArgs, "--logRoot", logRoot)
					}
					if !slices.Contains(serverArgs, "--flatLog") {
						serverArgs = append(serverArgs, "--flatLog")
					}
					if !slices.Contains(serverArgs, "--deterministic") {
						serverArgs = append(serverArgs, "--deterministic")
					}
				}
				if gameId == common.GameAoM && battleServerManagerRun == "false" {
					logger.Println("AoM: RT needs a Battle Server to be started but you don't allow to start one, make sure you have one running and the server configured.")
				}
				runBattleServerManager := battleServerManagerRun == "true" || (battleServerManagerRun == "required" && gameId == common.GameAoM)
				if viper.GetString("Server.Start") == "auto" {
					str := "No 'server's were found, proceeding to"
					if runBattleServerManager {
						str += " start a battle server (if needed) and then"
					}
					logger.Println(str + " start the 'server'. Press enter to continue...")
					_, _ = bufio.NewReader(os.Stdin).ReadBytes('\n')
				}
				serverExecutablePath := server.GetExecutablePath(serverExecutable)
				if serverExecutablePath == "" {
					logger.Println("Cannot find 'server' executable path. Set it manually in Server.Executable.")
					errorCode = internal.ErrServerExecutable
					return
				}
				if serverExecutable != serverExecutablePath {
					logger.Println("Found 'server' executable path:", serverExecutablePath)
				}
				if errorCode = server.GenerateServerCertificates(serverExecutablePath, canTrustCertificate != "false"); errorCode != common.ErrSuccess {
					return
				}
				if runBattleServerManager {
					errorCode = config.RunBattleServerManager(
						battleServerManagerExecutable,
						battleServerManagerArgs,
						serverStop == "true",
					)
					if errorCode != common.ErrSuccess {
						return
					}
				}
				errorCode, serverIP = config.StartServer(serverExecutablePath, serverArgs, serverStop == "true", serverId)
				if errorCode != common.ErrSuccess {
					return
				}
			}
			serverCertificate := server.ReadCACertificateFromServer(serverIP)
			if serverCertificate == nil {
				logger.Println("Failed to read certificate from " + serverIP + ". Try to access it with your browser and checking the certificate.")
				errorCode = internal.ErrReadCert
				return
			}
			errorCode = config.MapHosts(gameId, serverIP, canAddHost, slices.ContainsFunc(viper.GetStringSlice("Client.ExecutableArgs"), func(s string) bool {
				return strings.Contains(s, "{HostFilePath}")
			}))
			if errorCode != common.ErrSuccess {
				return
			}
			logger.WriteFileLog(gameId, "post host mapping")
			errorCode = config.AddCert(gameId, serverId, serverCertificate, canTrustCertificate, slices.ContainsFunc(viper.GetStringSlice("Client.ExecutableArgs"), func(s string) bool {
				return strings.Contains(s, "{CertFilePath}")
			}))
			if errorCode != common.ErrSuccess {
				return
			}
			logger.WriteFileLog(gameId, "post add cert")
			errorCode = config.IsolateUserData(isolateMetadata, isolateProfiles)
			if errorCode != common.ErrSuccess {
				return
			}
			logger.WriteFileLog(gameId, "post isolate user data")
			if gamePath != "" {
				errorCode = config.AddCACertToGame(gameId, serverCertificate, gamePath)
				if errorCode != common.ErrSuccess {
					return
				}
				logger.WriteFileLog(gameId, "post add game cert")
			}
			errorCode = config.LaunchAgentAndGame(executer, customExecutor, canTrustCertificate, canBroadcastBattleServer)
		},
	}
)

func Execute() error {
	rootCmd.Version = Version
	rootCmd.Flags().StringVar(&cfgFile, "config", "", fmt.Sprintf(`config file (default config.toml in %s directories)`, strings.Join(configPaths, ", ")))
	rootCmd.Flags().StringVar(&gameCfgFile, "gameConfig", "", fmt.Sprintf(`Game config file (default config.game.toml in %s directories)`, strings.Join(configPaths, ", ")))
	rootCmd.Flags().Bool("log", false, "Whether to log more info to a file. Enable it for errors.")
	rootCmd.Flags().StringP("canAddHost", "t", "true", "Add a local dns entry if it's needed to connect to the 'server' with the official domain. Including to avoid receiving that it's on maintenance. Ignored if 'clientExeArgs' contains '{HostFilePath}'. Will require admin privileges.")
	canTrustCertificateStr := `Trust the certificate of the 'server' if needed. "false"`
	if runtime.GOOS == "windows" {
		canTrustCertificateStr += `, "user"`
	}
	canTrustCertificateStr += ` or local (will require admin privileges). Ignored if 'clientExeArgs' contains '{CertFilePath}'.`
	rootCmd.Flags().StringP("canTrustCertificate", "c", "local", canTrustCertificateStr)
	if runtime.GOOS == "windows" {
		rootCmd.Flags().StringP("canBroadcastBattleServer", "b", "auto", `Whether or not to broadcast the game BattleServer to all interfaces in LAN (not just the most priority one)`)
	}
	var pathNamesInfo string
	if runtime.GOOS == "windows" {
		pathNamesInfo += " Path names need to use double backslashes within single quotes or be within double quotes."
	}
	cmd.GameVarCommand(rootCmd.Flags(), &gameId)
	if err := rootCmd.MarkFlagRequired("game"); err != nil {
		panic(err)
	}
	var suffixIsolate string
	if runtime.GOOS == "linux" {
		suffixIsolate = " Unsupported when using a custom launcher."
	}
	rootCmd.Flags().StringP("isolateMetadata", "m", "required", "Isolate the metadata cache of the game, otherwise, it will be shared. Not compatible with AoE:DE. If 'required' it will resolve to 'true' if using the official launcher, 'false' otherwise."+suffixIsolate)
	rootCmd.Flags().StringP("isolateProfiles", "p", "required", "Isolate the users profile of the game, otherwise, it will be shared. If 'required' it will resolve to 'true' if using the official launcher, 'false' otherwise."+suffixIsolate)
	rootCmd.Flags().String("setupCommand", "", `Executable to run (including arguments) to run first after the "Setting up..." line. The command must return a 0 exit code to continue. If you need to keep it running spawn a new separate process. You may use environment variables.`+pathNamesInfo)
	rootCmd.Flags().String("revertCommand", "", `Executable to run (including arguments) to run after setupCommand, game has exited and everything has been reverted. It may run before if there is an error. You may use environment variables.`+pathNamesInfo)
	rootCmd.Flags().StringP("serverStart", "a", "auto", `Start the 'server' if needed, "auto" will start a 'server' if one is not already running, "true" (will start a 'server' regardless if one is already running), "false" (will require an already running 'server').`)
	rootCmd.Flags().StringP("serverStop", "o", "auto", `Stop the 'server' if started, "auto" will stop the 'server' if one was started, "false" (will not stop the 'server' regardless if one was started), "true" (will not stop the 'server' even if it was started).`)
	rootCmd.Flags().StringSliceP("serverAnnouncePorts", "n", []string{strconv.Itoa(common.AnnouncePort)}, `Announce ports to listen to. If not including the default port, default configured 'servers' will not get discovered.`)
	rootCmd.Flags().StringSliceP("serverAnnounceMulticastGroups", "g", []string{common.AnnounceMulticastGroup}, `Announce multicast groups to join. If not including the default group, default configured 'servers' will not get discovered via Multicast.`)
	rootCmd.Flags().StringP("server", "s", "", `Hostname of the 'server' to connect to. If not absent, serverStart will be assumed to be false. Ignored otherwise`)
	rootCmd.Flags().Bool("serverSingleAutoSelect", false, `Auto-select the server when a single one is discovered.`)
	serverExe := executables.Filename(false, executables.Server)
	rootCmd.Flags().StringP("serverPath", "z", "auto", fmt.Sprintf(`The executable path of the 'server', "auto", will be try to execute in this order "./%s/%s", "../%s" and finally "../%s/%s", otherwise set the path (relative or absolute).`, executables.Server, serverExe, serverExe, executables.Server, serverExe))
	rootCmd.Flags().StringP("serverPathArgs", "r", "", `The arguments to pass to the 'server' executable if starting it. Execute the 'server' help flag for available arguments. You may use environment variables.`+pathNamesInfo)
	clientExeTip := `The type of game client or the path. "auto" will use Steam`
	if runtime.GOOS == "windows" {
		clientExeTip += ` and then the Xbox one if found`
	}
	clientExeTip += `. Use a path to the game launcher`
	if runtime.GOOS == "windows" {
		clientExeTip += ","
	} else {
		clientExeTip += " or"
	}
	clientExeTip += ` "steam"`
	if runtime.GOOS == "windows" {
		clientExeTip += `or "msstore"`
	}
	clientExeTip += " to use the default launcher."
	if runtime.GOOS == "linux" {
		clientExeTip += " If using a custom launcher, the isolation of the metadata and profiles will be disabled."
	}
	rootCmd.Flags().StringP("clientExe", "l", "auto", clientExeTip)
	rootCmd.Flags().StringP("clientExeArgs", "i", "", "The arguments to pass to the client launcher if it is custom. You may use environment variables and '{HostFilePath}'/'{CertFilePath}' replacement variables."+pathNamesInfo)
	if err := viper.BindPFlag("Config.CanAddHost", rootCmd.Flags().Lookup("canAddHost")); err != nil {
		return err
	}
	if err := viper.BindPFlag("Config.CanTrustCertificate", rootCmd.Flags().Lookup("canTrustCertificate")); err != nil {
		return err
	}
	if runtime.GOOS == "windows" {
		if err := viper.BindPFlag("Config.CanBroadcastBattleServer", rootCmd.Flags().Lookup("canBroadcastBattleServer")); err != nil {
			return err
		}
	}
	if err := viper.BindPFlag("Config.Log", rootCmd.Flags().Lookup("log")); err != nil {
		return err
	}
	if err := viper.BindPFlag("Config.IsolateMetadata", rootCmd.Flags().Lookup("isolateMetadata")); err != nil {
		return err
	}
	if err := viper.BindPFlag("Config.IsolateProfiles", rootCmd.Flags().Lookup("isolateProfiles")); err != nil {
		return err
	}
	if err := viper.BindPFlag("Config.SetupCommand", rootCmd.Flags().Lookup("setupCommand")); err != nil {
		return err
	}
	if err := viper.BindPFlag("Config.RevertCommand", rootCmd.Flags().Lookup("revertCommand")); err != nil {
		return err
	}
	if err := viper.BindPFlag("Server.Start", rootCmd.Flags().Lookup("serverStart")); err != nil {
		return err
	}
	if err := viper.BindPFlag("Server.Stop", rootCmd.Flags().Lookup("serverStop")); err != nil {
		return err
	}
	if err := viper.BindPFlag("Server.SingleAutoSelect", rootCmd.Flags().Lookup("serverSingleAutoSelect")); err != nil {
		return err
	}
	if err := viper.BindPFlag("Server.AnnouncePorts", rootCmd.Flags().Lookup("serverAnnouncePorts")); err != nil {
		return err
	}
	if err := viper.BindPFlag("Server.AnnounceMulticastGroups", rootCmd.Flags().Lookup("serverAnnounceMulticastGroups")); err != nil {
		return err
	}
	if err := viper.BindPFlag("Server.Host", rootCmd.Flags().Lookup("server")); err != nil {
		return err
	}
	if err := viper.BindPFlag("Server.Executable", rootCmd.Flags().Lookup("serverPath")); err != nil {
		return err
	}
	if err := viper.BindPFlag("Server.ExecutableArgs", rootCmd.Flags().Lookup("serverPathArgs")); err != nil {
		return err
	}
	if err := viper.BindPFlag("Client.Executable", rootCmd.Flags().Lookup("clientExe")); err != nil {
		return err
	}
	if err := viper.BindPFlag("Client.ExecutableArgs", rootCmd.Flags().Lookup("clientExeArgs")); err != nil {
		return err
	}
	return rootCmd.Execute()
}

func initConfig() {
	for _, configPath := range configPaths {
		viper.AddConfigPath(configPath)
	}
	viper.SetConfigType("toml")
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("config")
	}
	if err := viper.MergeInConfig(); err == nil {
		logger.Println("Using main config file:", viper.ConfigFileUsed())
	}
	if gameCfgFile != "" {
		viper.SetConfigFile(gameCfgFile)
	} else {
		viper.SetConfigName(fmt.Sprintf("config.%s", gameId))
	}
	if err := viper.MergeInConfig(); err == nil {
		logger.Println("Using game config file:", viper.ConfigFileUsed())
	}
	viper.AutomaticEnv()
}

package cmd

import (
	"bufio"
	"fmt"
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
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/cmd"
	commonExecutor "github.com/luskaner/ageLANServer/common/executor"
	"github.com/luskaner/ageLANServer/common/pidLock"
	commonProcess "github.com/luskaner/ageLANServer/common/process"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/luskaner/ageLANServer/launcher/internal"
	"github.com/luskaner/ageLANServer/launcher/internal/cmdUtils"
	"github.com/luskaner/ageLANServer/launcher/internal/executor"
	"github.com/luskaner/ageLANServer/launcher/internal/game"
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
	Version                            string
	cfgFile                            string
	gameCfgFile                        string
	gameId                             string
	autoTrueFalseValues                = mapset.NewThreadUnsafeSet[string](autoValue, trueValue, falseValue)
	canTrustCertificateValues          = mapset.NewThreadUnsafeSet[string](falseValue, "user", "local")
	canBroadcastBattleServerValues     = mapset.NewThreadUnsafeSet[string](autoValue, falseValue)
	serverBattleServerManagerRunValues = mapset.NewThreadUnsafeSet[string](trueValue, falseValue, "required")
	rootCmd                            = &cobra.Command{
		Use:   filepath.Base(os.Args[0]),
		Short: "launcher discovers and configures AoE: DE, AoE 2: DE and AoE 3: DE, and AoM: RT to connect to the local LAN 'server'",
		Long:  "launcher discovers or starts a local LAN 'server', configures and executes the game launcher to connect to it",
		Run: func(_ *cobra.Command, _ []string) {
			lock := &pidLock.Lock{}
			if err := lock.Lock(); err != nil {
				fmt.Println("Failed to lock pid file. Kill process 'launcher' if it is running in your task manager.")
				fmt.Println(err.Error())
				os.Exit(common.ErrPidLock)
			}
			var errorCode = common.ErrSuccess
			defer func() {
				_ = lock.Unlock()
				os.Exit(errorCode)
			}()
			initConfig()
			isAdmin := commonExecutor.IsAdmin()
			canTrustCertificate := viper.GetString("Config.CanTrustCertificate")
			if runtime.GOOS != "windows" {
				canTrustCertificateValues.Remove("user")
			}
			if !canTrustCertificateValues.Contains(canTrustCertificate) {
				fmt.Printf("Invalid value for canTrustCertificate (%s): %s\n", strings.Join(canTrustCertificateValues.ToSlice(), "/"), canTrustCertificate)
				errorCode = internal.ErrInvalidCanTrustCertificate
				return
			}
			canBroadcastBattleServer := "false"
			if runtime.GOOS == "windows" && gameId != common.GameAoM {
				canBroadcastBattleServer = viper.GetString("Config.CanBroadcastBattleServer")
				if !canBroadcastBattleServerValues.Contains(canBroadcastBattleServer) {
					fmt.Printf("Invalid value for canBroadcastBattleServer (auto/false): %s\n", canBroadcastBattleServer)
					errorCode = internal.ErrInvalidCanBroadcastBattleServer
					return
				}
			}
			serverStart := viper.GetString("Server.Start")
			if !autoTrueFalseValues.Contains(serverStart) {
				fmt.Printf("Invalid value for serverStart (auto/true/false): %s\n", serverStart)
				errorCode = internal.ErrInvalidServerStart
				return
			}
			serverStop := viper.GetString("Server.Stop")
			if runtime.GOOS != "windows" && isAdmin {
				autoTrueFalseValues.Remove(falseValue)
			}
			if !autoTrueFalseValues.Contains(serverStop) {
				fmt.Printf("Invalid value for serverStop (%s): %s\n", strings.Join(autoTrueFalseValues.ToSlice(), "/"), serverStop)
				errorCode = internal.ErrInvalidServerStop
				return
			}
			battleServerManagerRun := viper.GetString("Server.BattleServerManager.Run")
			if !serverBattleServerManagerRunValues.Contains(battleServerManagerRun) {
				fmt.Printf("Invalid value for Server.BattleServerManager.Run (%s): %s\n", strings.Join(serverBattleServerManagerRunValues.ToSlice(), "/"), battleServerManagerRun)
				errorCode = internal.ErrInvalidServerBattleServerManagerRun
				return
			}
			if !common.SupportedGames.ContainsOne(gameId) {
				fmt.Println("Invalid game type")
				errorCode = launcherCommon.ErrInvalidGame
				return
			}
			config.SetGameId(gameId)
			serverValues := map[string]string{
				"Game": gameId,
			}
			serverArgs, err := common.ParseCommandArgs("Server.ExecutableArgs", serverValues, true)
			if err != nil {
				fmt.Println("Failed to parse 'server' executable arguments")
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
				fmt.Println("Failed to parse 'battle-server-manager' executable arguments")
				errorCode = internal.ErrInvalidServerBattleServerManagerArgs
				return
			}
			var setupCommand []string
			setupCommand, err = common.ParseCommandArgs("Config.SetupCommand", nil, true)
			if err != nil {
				fmt.Println("Failed to parse setup command")
				errorCode = internal.ErrInvalidSetupCommand
				return
			}
			var revertCommand []string
			revertCommand, err = common.ParseCommandArgs("Config.RevertCommand", nil, true)
			if err != nil {
				fmt.Println("Failed to parse revert command")
				errorCode = internal.ErrInvalidRevertCommand
				return
			}
			canAddHost := viper.GetBool("Config.CanAddHost")
			var isolateMetadata bool
			if gameId != common.GameAoE1 {
				isolateMetadata = viper.GetBool("Config.IsolateMetadata")
			}
			isolateProfiles := viper.GetBool("Config.IsolateProfiles")
			var serverExecutable string
			if serverExecutable = viper.GetString("Server.Executable"); serverExecutable != "auto" {
				var serverFile os.FileInfo
				if serverFile, serverExecutable, err = common.ParsePath(viper.GetStringSlice("Server.Executable"), nil); err != nil || serverFile.IsDir() {
					fmt.Println("Invalid 'server' executable")
					errorCode = internal.ErrInvalidServerPath
					return
				}
			}
			var battleServerManagerExecutable string
			if battleServerManagerExecutable = viper.GetString("Server.BattleServerManager.Executable"); battleServerManagerExecutable != "auto" {
				var battleServerManagerFile os.FileInfo
				if battleServerManagerFile, battleServerManagerExecutable, err = common.ParsePath(viper.GetStringSlice("Server.BattleServerManager.Executable"), nil); err != nil || battleServerManagerFile.IsDir() {
					fmt.Println("Invalid 'battle-server-manager' executable")
					errorCode = internal.ErrInvalidClientPath
					return
				}
			}
			var clientExecutable string
			if clientExecutable = viper.GetString("Client.Executable"); clientExecutable != "auto" && clientExecutable != "steam" && clientExecutable != "msstore" {
				var clientFile os.FileInfo
				if clientFile, clientExecutable, err = common.ParsePath(viper.GetStringSlice("Client.Executable"), nil); err != nil || clientFile.IsDir() {
					fmt.Println("Invalid client executable")
					errorCode = internal.ErrInvalidClientPath
					return
				}
			}

			serverHost := viper.GetString("Server.Host")

			fmt.Printf("Game %s.\n", gameId)
			if clientExecutable == "msstore" && gameId == common.GameAoM {
				fmt.Println("The Microsoft Store (Xbox) version of AoM: RT is not supported.")
				errorCode = internal.ErrGameUnsupportedLauncherCombo
				return
			}

			fmt.Println("Looking for the game...")
			var gamePath string
			executer := game.MakeExecutor(gameId, clientExecutable)
			var customExecutor game.CustomExecutor
			switch executer.(type) {
			case game.SteamExecutor:
				fmt.Println("Game found on Steam.")
				if gameId != common.GameAoE1 {
					gamePath = executer.(game.SteamExecutor).GamePath()
				}
			case game.XboxExecutor:
				fmt.Println("Game found on Xbox.")
				if gameId != common.GameAoE1 {
					gamePath = executer.(game.XboxExecutor).GamePath()
				}
			case game.CustomExecutor:
				customExecutor = executer.(game.CustomExecutor)
				fmt.Println("Game found on custom path.")
				if runtime.GOOS == "linux" {
					if isolateMetadata {
						fmt.Println("Isolating metadata is not supported.")
						isolateMetadata = false
					}
					if isolateProfiles {
						fmt.Println("Isolating profiles is not supported.")
						isolateProfiles = false
					}
				}
				if gameId != common.GameAoE1 {
					if clientFile, clientPath, err := common.ParsePath(viper.GetStringSlice("Client.Path"), nil); err != nil || !clientFile.IsDir() {
						fmt.Println("Invalid client path")
						errorCode = internal.ErrInvalidClientPath
						return
					} else {
						gamePath = clientPath
					}
				}
			default:
				fmt.Println("Game not found.")
				errorCode = internal.ErrGameLauncherNotFound
				return
			}

			if isAdmin {
				fmt.Println("Running as administrator, this is not recommended for security reasons. It will request isolated admin privileges if/when needed.")
				if runtime.GOOS != "windows" {
					fmt.Println(" It can also cause issues and restrict the functionality.")
				}
			}

			if runtime.GOOS != "windows" && isAdmin && (clientExecutable == "auto" || clientExecutable == "steam") {
				fmt.Println("Steam cannot be run as administrator. Either run this as a normal user o set Client.Executable to a custom launcher.")
				errorCode = internal.ErrSteamRoot
				return
			}

			if cmdUtils.GameRunning() {
				errorCode = internal.ErrGameAlreadyRunning
				return
			}

			config.SetGameId(gameId)

			defer func() {
				if r := recover(); r != nil {
					fmt.Println(r)
					fmt.Println(string(debug.Stack()))
					errorCode = common.ErrGeneral
				}
				if errorCode != common.ErrSuccess {
					config.Revert()
				}
			}()
			sigs := make(chan os.Signal, 1)
			signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
			go func() {
				_, ok := <-sigs
				if ok {
					config.Revert()
					_ = lock.Unlock()
					os.Exit(errorCode)
				}
			}()
			/*
				Ensure:
				* No running config-admin-agent nor agent processes
				* Any previous changes are reverted
			*/
			fmt.Println("Cleaning up (if needed)...")
			config.KillAgent()
			launcherCommon.ConfigRevert(gameId, false, executor.RunRevert)
			var proc *os.Process
			_, proc, err = commonProcess.Process(common.GetExeFileName(false, common.Server))
			if err == nil && proc != nil {
				fmt.Println("'Server' is already running, If you did not start it manually, kill the 'server' process using the task manager and execute the 'launcher' again.")
			}
			if err = executor.RunRevertCommand(); err != nil {
				fmt.Println("Failed to run revert command.")
				fmt.Println("Error message: " + err.Error())
			}
			if len(revertCommand) > 0 {
				if err := launcherCommon.RevertCommandStore.Store(revertCommand); err != nil {
					fmt.Println("Failed to store revert command")
					errorCode = internal.ErrInvalidRevertCommand
					return
				}
			}
			// Setup
			fmt.Println("Setting up...")
			if len(setupCommand) > 0 {
				fmt.Printf("Running setup command '%s' and waiting for it to exit...\n", viper.GetString("Config.SetupCommand"))
				result := config.RunSetupCommand(setupCommand)
				if !result.Success() {
					if result.Err != nil {
						fmt.Printf("Error: %s\n", result.Err)
					}
					if result.ExitCode != common.ErrSuccess {
						fmt.Printf(`Exit code: %d.`+"\n", result.ExitCode)
					}
					errorCode = internal.ErrSetupCommand
					return
				}
			}
			if serverStart == "auto" {
				announcePorts := viper.GetIntSlice("Server.AnnouncePorts")
				portsStr := make([]string, len(announcePorts))
				for i, portInt := range announcePorts {
					portsStr[i] = strconv.Itoa(portInt)
				}
				multicastIPsStr := viper.GetStringSlice("Server.AnnounceMulticastGroups")
				multicastIPs := make([]net.IP, len(multicastIPsStr))
				for i, str := range multicastIPsStr {
					if IP := net.ParseIP(str); IP.To4() != nil && IP.IsMulticast() {
						multicastIPs[i] = IP
					} else {
						fmt.Printf("Invalid multicast group \"%s\"\n", str)
						errorCode = internal.ErrAnnouncementMulticastGroup
						return
					}
				}
				fmt.Printf("Waiting 15 seconds for 'server' announcements on LAN on port(s) %s (we are v. %d), you might need to allow 'launcher' in the firewall...\n", strings.Join(portsStr, ", "), common.AnnounceVersionLatest)
				errorCode, selectedServerIp := cmdUtils.ListenToServerAnnouncementsAndSelectBestIp(gameId, multicastIPs, announcePorts)
				if errorCode != common.ErrSuccess {
					return
				} else if selectedServerIp != "" {
					serverHost = selectedServerIp
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
			var serverIP string
			if serverStart == "false" {
				if serverStop == "true" {
					fmt.Println("serverStart is false. Ignoring serverStop being true.")
				}
				if serverHost == "" {
					fmt.Println("serverStart is false. serverHost must be fulfilled as it is needed to know which host to connect to.")
					errorCode = internal.ErrInvalidServerHost
					return
				}
				if addr, err := netip.ParseAddr(serverHost); err == nil && addr.Is6() {
					fmt.Println("serverStart is false. serverHost must be fulfilled with a host or Ipv4 address.")
					errorCode = internal.ErrInvalidServerHost
					return
				}
				if !server.CheckConnectionFromServer(serverHost, true) {
					fmt.Println("serverStart is false. " + serverHost + " must be reachable. Review the host is correct, the 'server' is started and you can connect to TCP port 443 (HTTPS).")
					errorCode = internal.ErrInvalidServerStart
					return
				}
				var ok bool
				if ok, serverIP = cmdUtils.SelectBestServerIp(common.HostOrIpToIpsSet(serverHost).ToSlice()); !ok {
					fmt.Println("serverStart is false. Failed to resolve serverHost to a valid and reachable IP.")
				}
			} else {
				if gameId == common.GameAoM && battleServerManagerRun == "false" {
					fmt.Println("AoM: RT needs a Battle Server to be started but you don't allow to start one, make sure you have one running and the server configured.")
				}
				runBattleServerManager := battleServerManagerRun == "true" || (battleServerManagerRun == "required" && gameId == common.GameAoM)
				if viper.GetString("Server.Start") == "auto" {
					fmt.Print("No 'server's were found, proceeding to")
					if runBattleServerManager {
						fmt.Print(" start a battle server (if needed) and then")
					}
					fmt.Println(" start the 'server'.")
					_, _ = bufio.NewReader(os.Stdin).ReadBytes('\n')
				}
				serverExecutablePath := server.GetExecutablePath(serverExecutable)
				if serverExecutablePath == "" {
					fmt.Println("Cannot find 'server' executable path. Set it manually in Server.Executable.")
					errorCode = internal.ErrServerExecutable
					return
				}
				if serverExecutable != serverExecutablePath {
					fmt.Println("Found 'server' executable path:", serverExecutablePath)
				}
				if errorCode = server.GenerateServerCertificates(serverExecutablePath, canTrustCertificate != "false"); errorCode != common.ErrSuccess {
					return
				}
				if runBattleServerManager {
					errorCode = config.RunBattleServerManager(
						gameId,
						battleServerManagerExecutable,
						battleServerManagerArgs,
						serverStop == "true",
					)
					if errorCode != common.ErrSuccess {
						return
					}
				}
				errorCode, serverIP = config.StartServer(serverExecutablePath, serverArgs, serverStop == "true")
				if errorCode != common.ErrSuccess {
					return
				}
			}
			serverCertificate := server.ReadCACertificateFromServer(serverIP)
			if serverCertificate == nil {
				fmt.Println("Failed to read certificate from " + serverIP + ". Try to access it with your browser and checking the certificate.")
				errorCode = internal.ErrReadCert
				return
			}
			errorCode = config.MapHosts(gameId, serverIP, canAddHost, slices.ContainsFunc(viper.GetStringSlice("Client.ExecutableArgs"), func(s string) bool {
				return strings.Contains(s, "{HostFilePath}")
			}))
			if errorCode != common.ErrSuccess {
				return
			}
			errorCode = config.AddCert(gameId, serverCertificate, canTrustCertificate, slices.ContainsFunc(viper.GetStringSlice("Client.ExecutableArgs"), func(s string) bool {
				return strings.Contains(s, "{CertFilePath}")
			}))
			if errorCode != common.ErrSuccess {
				return
			}
			errorCode = config.IsolateUserData(isolateMetadata, isolateProfiles)
			if errorCode != common.ErrSuccess {
				return
			}
			if gamePath != "" {
				errorCode = config.AddCACertToGame(gameId, serverCertificate, gamePath)
				if errorCode != common.ErrSuccess {
					return
				}
			}
			errorCode = config.LaunchAgentAndGame(executer, customExecutor, canTrustCertificate, canBroadcastBattleServer)
		},
	}
)

func Execute() error {
	rootCmd.Version = Version
	rootCmd.Flags().StringVar(&cfgFile, "config", "", fmt.Sprintf(`config file (default config.toml in %s directories)`, strings.Join(configPaths, ", ")))
	rootCmd.Flags().StringVar(&gameCfgFile, "gameConfig", "", fmt.Sprintf(`Game config file (default config.game.toml in %s directories)`, strings.Join(configPaths, ", ")))
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
	rootCmd.Flags().StringP("isolateMetadata", "m", "true", "Isolate the metadata cache of the game, otherwise, it will be shared. Not compatible with AoE:DE."+suffixIsolate)
	rootCmd.Flags().BoolP("isolateProfiles", "p", false, "(Experimental) Isolate the users profile of the game, otherwise, it will be shared."+suffixIsolate)
	rootCmd.Flags().String("setupCommand", "", `Executable to run (including arguments) to run first after the "Setting up..." line. The command must return a 0 exit code to continue. If you need to keep it running spawn a new separate process. You may use environment variables.`+pathNamesInfo)
	rootCmd.Flags().String("revertCommand", "", `Executable to run (including arguments) to run after setupCommand, game has exited and everything has been reverted. It may run before if there is an error. You may use environment variables.`+pathNamesInfo)
	rootCmd.Flags().StringP("serverStart", "a", "auto", `Start the 'server' if needed, "auto" will start a 'server' if one is not already running, "true" (will start a 'server' regardless if one is already running), "false" (will require an already running 'server').`)
	rootCmd.Flags().StringP("serverStop", "o", "auto", `Stop the 'server' if started, "auto" will stop the 'server' if one was started, "false" (will not stop the 'server' regardless if one was started), "true" (will not stop the 'server' even if it was started).`)
	rootCmd.Flags().StringSliceP("serverAnnouncePorts", "n", []string{strconv.Itoa(common.AnnouncePort)}, `Announce ports to listen to. If not including the default port, default configured 'servers' will not get discovered.`)
	rootCmd.Flags().StringSliceP("serverAnnounceMulticastGroups", "g", []string{"239.31.97.8"}, `Announce multicast groups to join. If not including the default group, default configured 'servers' will not get discovered via Multicast.`)
	rootCmd.Flags().StringP("server", "s", "", `Hostname of the 'server' to connect to. If not absent, serverStart will be assumed to be false. Ignored otherwise`)
	serverExe := common.GetExeFileName(false, common.Server)
	rootCmd.Flags().StringP("serverPath", "z", "auto", fmt.Sprintf(`The executable path of the 'server', "auto", will be try to execute in this order "./%s/%s", "../%s" and finally "../%s/%s", otherwise set the path (relative or absolute).`, common.Server, serverExe, serverExe, common.Server, serverExe))
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
		fmt.Println("Using main config file:", viper.ConfigFileUsed())
	}
	if gameCfgFile != "" {
		viper.SetConfigFile(gameCfgFile)
	} else {
		viper.SetConfigName(fmt.Sprintf("config.%s", gameId))
	}
	if err := viper.MergeInConfig(); err == nil {
		fmt.Println("Using game config file:", viper.ConfigFileUsed())
	}
	viper.AutomaticEnv()
}

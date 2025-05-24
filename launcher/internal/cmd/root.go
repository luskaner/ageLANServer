package cmd

import (
	"bufio"
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/cmd"
	commonExecutor "github.com/luskaner/ageLANServer/common/executor"
	"github.com/luskaner/ageLANServer/common/pidLock"
	commonProcess "github.com/luskaner/ageLANServer/common/process"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/luskaner/ageLANServer/launcher/internal"
	"github.com/luskaner/ageLANServer/launcher/internal/cmdUtils"
	"github.com/luskaner/ageLANServer/launcher/internal/cmdUtils/parse"
	"github.com/luskaner/ageLANServer/launcher/internal/executor"
	"github.com/luskaner/ageLANServer/launcher/internal/game"
	"github.com/luskaner/ageLANServer/launcher/internal/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	autoTrueFalseValues            = mapset.NewThreadUnsafeSet[string](autoValue, trueValue, falseValue)
	canTrustCertificateValues      = mapset.NewThreadUnsafeSet[string](falseValue, "user", "local")
	canBroadcastBattleServerValues = mapset.NewThreadUnsafeSet[string](autoValue, falseValue)
	rootCmd                        = &cobra.Command{
		Use:   filepath.Base(os.Args[0]),
		Short: "launcher discovers and configures AoE 1, AoE 2 and AoE 3 (all DE) to connect to the local LAN 'server'",
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
			if runtime.GOOS == "windows" {
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
			gameId := viper.GetString("Game")
			if !common.SupportedGames.ContainsOne(gameId) {
				fmt.Println("Invalid game type")
				errorCode = launcherCommon.ErrInvalidGame
				return
			}
			config.SetGameId(gameId)
			serverValues := map[string]string{
				"Game": gameId,
			}
			serverArgs, err := parse.CommandArgs(viper.GetStringSlice("Server.ExecutableArgs"), serverValues)
			if err != nil {
				fmt.Println("Failed to parse 'server' executable arguments")
				errorCode = internal.ErrInvalidServerArgs
				return
			}
			var setupCommand []string
			setupCommand, err = parse.CommandArgs(viper.GetStringSlice("Config.SetupCommand"), nil)
			if err != nil {
				fmt.Println("Failed to parse setup command")
				errorCode = internal.ErrInvalidSetupCommand
				return
			}
			var revertCommand []string
			revertCommand, err = parse.CommandArgs(viper.GetStringSlice("Config.RevertCommand"), nil)
			if err != nil {
				fmt.Println("Failed to parse revert command")
				errorCode = internal.ErrInvalidRevertCommand
				return
			}
			canAddHost := viper.GetBool("Config.CanAddHost")
			clientExecutable := viper.GetString("Client.Executable")
			var isolateMetadata bool
			if gameId != common.GameAoE1 {
				isolateMetadata = viper.GetBool("Config.Isolation.Metadata")
			}
			isolateProfiles := viper.GetBool("Config.Isolation.Profiles")
			var isolateWindowsUserProfilePath string
			if runtime.GOOS == "linux" {
				isolateWindowsUserProfilePathTemp := viper.GetString("Config.Isolation.WindowsUserProfilePath")
				if isolateWindowsUserProfilePathTemp != "auto" {
					if windowsIsolateUserProfilePathTempParsed, err := parse.Executable(isolateWindowsUserProfilePathTemp, nil); err == nil {
						isolateWindowsUserProfilePath = windowsIsolateUserProfilePathTempParsed
					} else {
						isolateWindowsUserProfilePath = isolateWindowsUserProfilePathTemp
					}
				} else if clientExecutable != "auto" && clientExecutable != "steam" && (isolateMetadata || isolateProfiles) {
					fmt.Println("You need to set a custom user profile path when enabling some isolation and using a custom launcher.")
					errorCode = internal.ErrInvalidIsolationWindowsUserProfilePath
					return
				}
			}
			serverExecutable := viper.GetString("Server.Executable")
			serverHost := viper.GetString("Server.Host")

			fmt.Printf("Game %s.\n", gameId)
			fmt.Println("Looking for the game...")
			executer := game.MakeExecutor(gameId, clientExecutable)
			var customExecutor game.CustomExecutor
			switch executer.(type) {
			case game.SteamExecutor:
				fmt.Println("Game found on Steam.")
			case game.XboxExecutor:
				fmt.Println("Game found on Xbox.")
			case game.CustomExecutor:
				customExecutor = executer.(game.CustomExecutor)
				fmt.Println("Game found on custom path.")
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
			if !config.KillAgent() {
				errorCode = common.ErrGeneral
				return
			}
			if !launcherCommon.ConfigRevert(gameId, false, executor.RunRevert) {
				errorCode = common.ErrGeneral
				return
			}
			_, _, err = commonProcess.Process(common.GetExeFileName(false, common.Server))
			if err == nil {
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
				if ok, serverIP = cmdUtils.SelectBestServerIp(launcherCommon.HostOrIpToIps(serverHost).ToSlice()); !ok {
					fmt.Println("serverStart is false. Failed to resolve serverHost to a valid and reachable IP.")
				}
			} else {
				if viper.GetString("Server.Start") == "auto" {
					fmt.Println("No 'server's were found, proceeding to start one, press the Enter key to continue...")
					_, _ = bufio.NewReader(os.Stdin).ReadBytes('\n')
				}
				errorCode, serverIP = config.StartServer(serverExecutable, serverArgs, serverStop == "true", canTrustCertificate != "false")
				if errorCode != common.ErrSuccess {
					return
				}
			}
			serverCertificate := server.ReadCertificateFromServer(serverIP)
			if serverCertificate == nil {
				fmt.Println("Failed to read certificate from " + serverIP + ". Try to access it with your browser and checking the certificate.")
				errorCode = internal.ErrReadCert
				return
			}
			errorCode = config.MapHosts(serverIP, canAddHost, slices.ContainsFunc(viper.GetStringSlice("Client.ExecutableArgs"), func(s string) bool {
				return strings.Contains(s, "{HostFilePath}")
			}))
			if errorCode != common.ErrSuccess {
				return
			}
			errorCode = config.AddCert(serverCertificate, canTrustCertificate, slices.ContainsFunc(viper.GetStringSlice("Client.ExecutableArgs"), func(s string) bool {
				return strings.Contains(s, "{CertFilePath}")
			}))
			if errorCode != common.ErrSuccess {
				return
			}
			errorCode = config.IsolateUserData(isolateWindowsUserProfilePath, isolateMetadata, isolateProfiles)
			if errorCode != common.ErrSuccess {
				return
			}
			errorCode = config.LaunchAgentAndGame(executer, customExecutor, viper.GetStringSlice("Client.ExecutableArgs"), canTrustCertificate, canBroadcastBattleServer)
		},
	}
)

func Execute() error {
	cobra.OnInitialize(initConfig)
	rootCmd.Version = Version
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", fmt.Sprintf(`config file (default config.toml in %s directories)`, strings.Join(configPaths, ", ")))
	rootCmd.PersistentFlags().StringVar(&gameCfgFile, "gameConfig", "", fmt.Sprintf(`Game config file (default config.game.toml in %s directories)`, strings.Join(configPaths, ", ")))
	rootCmd.PersistentFlags().StringP("canAddHost", "t", "true", "Add a local dns entry if it's needed to connect to the 'server' with the official domain. Including to avoid receiving that it's on maintenance. Ignored if 'clientExeArgs' contains '{HostFilePath}'. Will require admin privileges.")
	canTrustCertificateStr := `Trust the certificate of the 'server' if needed. "false"`
	if runtime.GOOS == "windows" {
		canTrustCertificateStr += `, "user"`
	}
	canTrustCertificateStr += ` or local (will require admin privileges). Ignored if 'clientExeArgs' contains '{CertFilePath}'.`
	rootCmd.PersistentFlags().StringP("canTrustCertificate", "c", "local", canTrustCertificateStr)
	if runtime.GOOS == "windows" {
		rootCmd.PersistentFlags().StringP("canBroadcastBattleServer", "b", "auto", `Whether or not to broadcast the game BattleServer to all interfaces in LAN (not just the most priority one)`)
	}
	var pathNamesInfo string
	if runtime.GOOS == "windows" {
		pathNamesInfo += " Path names need to use double backslashes within single quotes or be within double quotes."
	}
	cmd.GameCommand(rootCmd.PersistentFlags())
	var suffixIsolate string
	if runtime.GOOS == "linux" {
		suffixIsolate = " When using a custom launcher 'windowsUserProfilePath' must be set."
	}
	if runtime.GOOS == "linux" {
		rootCmd.PersistentFlags().StringP("windowsUserProfilePath", "s", "auto", "Windows User Profile Path. Only relevant when using the 'isolateMetadata' or 'isolateProfiles' options. Must be set if using a custom launcher.")
	}
	rootCmd.PersistentFlags().StringP("isolateMetadata", "m", "true", "Isolate the metadata cache of the game, otherwise, it will be shared. Not compatible with AoE:DE."+suffixIsolate)
	rootCmd.PersistentFlags().BoolP("isolateProfiles", "p", false, "(Experimental) Isolate the users profile of the game, otherwise, it will be shared."+suffixIsolate)
	rootCmd.PersistentFlags().String("setupCommand", "", `Executable to run (including arguments) to run first after the "Setting up..." line. The command must return a 0 exit code to continue. If you need to keep it running spawn a new separate process. You may use environment variables.`+pathNamesInfo)
	rootCmd.PersistentFlags().String("revertCommand", "", `Executable to run (including arguments) to run after setupCommand, game has exited and everything has been reverted. It may run before if there is an error. You may use environment variables.`+pathNamesInfo)
	rootCmd.PersistentFlags().StringP("serverStart", "a", "auto", `Start the 'server' if needed, "auto" will start a 'server' if one is not already running, "true" (will start a 'server' regardless if one is already running), "false" (will require an already running 'server').`)
	rootCmd.PersistentFlags().StringP("serverStop", "o", "auto", `Stop the 'server' if started, "auto" will stop the 'server' if one was started, "false" (will not stop the 'server' regardless if one was started), "true" (will not stop the 'server' even if it was started).`)
	rootCmd.PersistentFlags().StringSliceP("serverAnnouncePorts", "n", []string{strconv.Itoa(common.AnnouncePort)}, `Announce ports to listen to. If not including the default port, default configured 'servers' will not get discovered.`)
	rootCmd.PersistentFlags().StringSliceP("serverAnnounceMulticastGroups", "g", []string{"239.31.97.8"}, `Announce multicast groups to join. If not including the default group, default configured 'servers' will not get discovered via Multicast.`)
	rootCmd.PersistentFlags().StringP("server", "s", "", `Hostname of the 'server' to connect to. If not absent, serverStart will be assumed to be false. Ignored otherwise`)
	serverExe := common.GetExeFileName(false, common.Server)
	rootCmd.PersistentFlags().StringP("serverPath", "z", "auto", fmt.Sprintf(`The executable path of the 'server', "auto", will be try to execute in this order "./%s/%s", "../%s" and finally "../%s/%s", otherwise set the path (relative or absolute).`, common.Server, serverExe, serverExe, common.Server, serverExe))
	rootCmd.PersistentFlags().StringP("serverPathArgs", "r", "", `The arguments to pass to the 'server' executable if starting it. Execute the 'server' help flag for available arguments. You may use environment variables.`+pathNamesInfo)
	clientExeTip := `The type of game client or the path. `
	if runtime.GOOS == "windows" {
		clientExeTip += `"auto" will use Steam and then the Xbox one if found`
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
		clientExeTip += " If using a custom launcher, you need to specify the user profile path if you want isolation."
	}
	rootCmd.PersistentFlags().StringP("clientExe", "l", "auto", clientExeTip)
	rootCmd.PersistentFlags().StringP("clientExeArgs", "i", "", "The arguments to pass to the client launcher if it is custom. You may use environment variables and '{HostFilePath}'/'{CertFilePath}' replacement variables."+pathNamesInfo)
	if err := viper.BindPFlag("Config.CanAddHost", rootCmd.PersistentFlags().Lookup("canAddHost")); err != nil {
		return err
	}
	if err := viper.BindPFlag("Config.CanTrustCertificate", rootCmd.PersistentFlags().Lookup("canTrustCertificate")); err != nil {
		return err
	}
	if runtime.GOOS == "windows" {
		if err := viper.BindPFlag("Config.CanBroadcastBattleServer", rootCmd.PersistentFlags().Lookup("canBroadcastBattleServer")); err != nil {
			return err
		}
	}
	if runtime.GOOS == "linux" {
		if err := viper.BindPFlag("Config.Isolation.WindowsUserProfilePath", rootCmd.PersistentFlags().Lookup("windowsUserProfilePath")); err != nil {
			return err
		}
	}
	if err := viper.BindPFlag("Config.Isolation.Metadata", rootCmd.PersistentFlags().Lookup("isolateMetadata")); err != nil {
		return err
	}
	if err := viper.BindPFlag("Config.Isolation.Profiles", rootCmd.PersistentFlags().Lookup("isolateProfiles")); err != nil {
		return err
	}
	if err := viper.BindPFlag("Config.SetupCommand", rootCmd.PersistentFlags().Lookup("setupCommand")); err != nil {
		return err
	}
	if err := viper.BindPFlag("Config.RevertCommand", rootCmd.PersistentFlags().Lookup("revertCommand")); err != nil {
		return err
	}
	if err := viper.BindPFlag("Server.Start", rootCmd.PersistentFlags().Lookup("serverStart")); err != nil {
		return err
	}
	if err := viper.BindPFlag("Server.Stop", rootCmd.PersistentFlags().Lookup("serverStop")); err != nil {
		return err
	}
	if err := viper.BindPFlag("Server.AnnouncePorts", rootCmd.PersistentFlags().Lookup("serverAnnouncePorts")); err != nil {
		return err
	}
	if err := viper.BindPFlag("Server.AnnounceMulticastGroups", rootCmd.PersistentFlags().Lookup("serverAnnounceMulticastGroups")); err != nil {
		return err
	}
	if err := viper.BindPFlag("Server.Host", rootCmd.PersistentFlags().Lookup("server")); err != nil {
		return err
	}
	if err := viper.BindPFlag("Server.Executable", rootCmd.PersistentFlags().Lookup("serverPath")); err != nil {
		return err
	}
	if err := viper.BindPFlag("Server.ExecutableArgs", rootCmd.PersistentFlags().Lookup("serverPathArgs")); err != nil {
		return err
	}
	if err := viper.BindPFlag("Client.Executable", rootCmd.PersistentFlags().Lookup("clientExe")); err != nil {
		return err
	}
	if err := viper.BindPFlag("Client.ExecutableArgs", rootCmd.PersistentFlags().Lookup("clientExeArgs")); err != nil {
		return err
	}
	if err := viper.BindPFlag("Game", rootCmd.PersistentFlags().Lookup("game")); err != nil {
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
		viper.SetConfigName("config.game")
	}
	if err := viper.MergeInConfig(); err == nil {
		fmt.Println("Using game config file:", viper.ConfigFileUsed())
	}
	viper.AutomaticEnv()
}

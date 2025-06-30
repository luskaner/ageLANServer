package cmd

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/cmd"
	commonCmd "github.com/luskaner/ageLANServer/common/cmd"
	cfg "github.com/luskaner/ageLANServer/common/config"
	"github.com/luskaner/ageLANServer/common/config/launcher"
	"github.com/luskaner/ageLANServer/common/config/launcher/parse"
	commonExecutor "github.com/luskaner/ageLANServer/common/executor"
	"github.com/luskaner/ageLANServer/common/pidLock"
	commonProcess "github.com/luskaner/ageLANServer/common/process"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/luskaner/ageLANServer/launcher/internal"
	"github.com/luskaner/ageLANServer/launcher/internal/cmdUtils"
	"github.com/luskaner/ageLANServer/launcher/internal/cmdUtils/printer"
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
	"strings"
	"syscall"
)

var configPaths = []string{"resources", "."}
var config = &cmdUtils.Config{}
var parameters = launcher.Config{}
var v = viper.New()
var lock *pidLock.Lock

var (
	Version  string
	cfgFiles []string
	rootCmd  = &cobra.Command{
		Use:   filepath.Base(os.Args[0]),
		Short: "launcher discovers and configures AoE 1, AoE 2 and AoE 3 (all DE) to connect to the local LAN 'server'",
		Long:  "launcher discovers or starts a local LAN 'server', configures and executes the game launcher to connect to it",
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			if err := launcher.Unmarshal(v, &parameters); err != nil {
				return err
			}
			err, validate := launcher.Validator()
			if err != nil {
				return err
			}
			if err := validate.Struct(&parameters); err != nil {
				return err
			}
			config.SetGameTitle(parameters.GameTitle)
			if parameters.Client.Launcher == "path" {
				parameters.Client.PathCommand[0] = parse.CommandArgs([]string{parameters.Client.PathCommand[0]}, nil)[0]
			}
			if parameters.Server.Mode != "connect" && len(parameters.Server.Run.Command) > 0 && parameters.Server.Run.Command[0] != "" {
				if err = validate.Var(parameters.Server.Run.Command[0], "file"); err != nil {
					return err
				}
			}
			parameters.SetupCommand = parse.CommandArgs(parameters.SetupCommand, nil)
			parameters.RevertCommand = parse.CommandArgs(parameters.RevertCommand, nil)
			lock = &pidLock.Lock{}
			if err := lock.Lock(); err != nil {
				return fmt.Errorf("failed to lock PID file: %w", err)
			}
			if cmdUtils.GameRunning() {
				return fmt.Errorf("an age game is already running")
			}
			printer.EnableDebug = parameters.Debug
			return nil
		},
		Run: func(_ *cobra.Command, _ []string) {
			var errorCode = common.ErrSuccess
			defer func() {
				_ = lock.Unlock()
				os.Exit(errorCode)
			}()
			isAdmin := commonExecutor.IsAdmin()
			storeToAddHosts := parameters.StoreToAddHost
			storeToAddCertificate := parameters.StoreToAddCertificate
			var rebroadcastIPs []net.IP
			if runtime.GOOS == "windows" && !parameters.RebroadcastBattleServer.Disable {
				rebroadcastIPs = game.RebroadcastIPs(
					parameters.RebroadcastBattleServer.IPs,
					parameters.RebroadcastBattleServer.Interfaces,
				)
			}
			var serverStart string
			var serverStop string
			switch parameters.Server.Mode {
			case "connect":
				serverStart = "false"
				serverStop = "false"
			case "run":
				serverStart = "true"
				if parameters.Server.Run.NoStop {
					serverStop = "false"
				} else {
					serverStop = "true"
				}
			}

			var isolateMetadata bool
			if config.GameTitle() != common.AoE1 {
				isolateMetadata = !parameters.Isolation.NoMetadata
			}
			var serverExecutable string
			if len(parameters.Server.Run.Command) > 0 {
				serverExecutable = parameters.Server.Run.Command[0]
			}
			var serverArgs []string
			if len(parameters.Server.Run.Command) > 1 {
				serverArgs = parameters.Server.Run.Command[1:]
			}
			serverHost := parameters.Server.Connect.Host

			printer.Print(
				printer.Search,
				"",
				printer.T(`Looking for `),
				printer.TS(string(config.GameTitle()), printer.LiteralStyle),
				printer.T(` game... `),
			)
			var clientExecutable string
			if parameters.Client.Launcher == "path" {
				clientExecutable = parameters.Client.PathCommand[0]
			}
			var clientArgs []string
			if len(parameters.Client.PathCommand) > 1 {
				clientArgs = parameters.Client.PathCommand[1:]
			}
			executer := game.MakeExecutor(config.GameTitle(), parameters.Client.Launcher, clientExecutable)
			var customExecutor game.CustomExecutor
			switch executer.(type) {
			case game.SteamExecutor:
				printer.Println(
					printer.Success,
					printer.T(`found on `),
					printer.TS(`Steam`, printer.LiteralStyle),
					printer.T(`.`),
				)
			case game.XboxExecutor:
				printer.Println(
					printer.Success,
					printer.T(`found on `),
					printer.TS(`Xbox`, printer.LiteralStyle),
					printer.T(`.`),
				)
			case game.CustomExecutor:
				customExecutor = executer.(game.CustomExecutor)
				printer.Println(
					printer.Success,
					printer.T(`found on `),
					printer.TS(customExecutor.Executable, printer.FilePathStyle),
					printer.T(`.`),
				)
			default:
				printer.PrintSimpln(
					printer.Error,
					`not found.`,
				)
				errorCode = internal.ErrGameLauncherNotFound
				return
			}

			if isAdmin {
				printer.PrintSimpln(printer.Warning, "Running as administrator, this is not recommended for security reasons. It will request isolated admin privileges if/when needed.")
				if runtime.GOOS != "windows" {
					printer.PrintSimpln(printer.Warning, "It can also cause issues and restrict the functionality.")
				}
			}

			if runtime.GOOS != "windows" && isAdmin && (clientExecutable == "" || clientExecutable == "steam") {
				printer.Println(
					printer.Error,
					printer.TS(`Steam`, printer.LiteralStyle),
					printer.T(` cannot be run as administrator. Either run this as a normal user or set `),
					printer.TS("Client.Executable", printer.OptionStyle),
					printer.T(` to a path.`),
				)
				errorCode = internal.ErrSteamRoot
				return
			}

			defer func() {
				if r := recover(); r != nil {
					printer.PrintSimpln(printer.Error, fmt.Sprint(r))
					printer.PrintSimpln(printer.Debug, string(debug.Stack()))
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
			printer.PrintSimpln(printer.Clean, "Cleaning up...")
			if !config.KillAgent() {
				errorCode = common.ErrGeneral
				return
			}
			if !launcherCommon.ConfigRevert(config.GameTitle(), false, executor.RunRevert, printer.ConfigRevertPrinter()) {
				errorCode = common.ErrGeneral
				return
			}
			if err := executor.RunRevertCommand(); err != nil {
				printer.Println(
					printer.Error,
					printer.T("Failed to run the last execution "),
					printer.TS("RevertCommand", printer.OptionStyle),
					printer.T("."),
				)
				printer.Println(
					printer.Debug,
					printer.T("Error message:"),
					printer.T(err.Error()),
				)
			}
			if _, _, err := commonProcess.Process(common.GetExeFileName(false, common.Server)); err == nil {
				printer.Println(
					printer.Info,
					printer.TS("Server", printer.ComponentStyle),
					printer.T(" is already running."),
				)
			}
			if mismatchHosts := cmdUtils.InternalExternalDnsMismatch(); !mismatchHosts.IsEmpty() {
				printer.Println(
					printer.Warning,
					printer.T("Host(s) "),
					printer.TS(strings.Join(mismatchHosts.ToSlice(), ", "), printer.LiteralStyle),
					printer.T(" already mapped. This can cause issues in the configuration."),
				)
			}
			// Setup
			if len(parameters.SetupCommand) > 0 {
				fmt.Print(printer.Gen(
					printer.Execute,
					"",
					printer.T(`Running `),
					printer.TS("SetupCommand", printer.OptionStyle),
					printer.T(` and waiting for it...`),
				))
				result := config.RunSetupCommand(parameters.SetupCommand)
				if !result.Success() {
					printer.PrintFailedResultError(result)
					errorCode = internal.ErrSetupCommand
					return
				} else {
					printer.PrintSucceeded()
					if len(parameters.RevertCommand) > 0 {
						if err := launcherCommon.RevertCommandStore.Store(parameters.RevertCommand); err != nil {
							printer.Println(
								printer.Error,
								printer.T("Failed to store the "),
								printer.TS("RevertCommand", printer.OptionStyle),
							)
							errorCode = internal.ErrInvalidRevertCommand
							return
						}
					}
				}
			}
			serverId := uuid.Nil
			if serverStart == "" {
				var selectedServerIp string
				errorCode, serverId, selectedServerIp = cmdUtils.DiscoverServersAndSelectBestIp(
					!parameters.Server.Query.NoBroadcast,
					config.GameTitle(),
					parameters.Server.Query.MulticastGroups,
					parameters.Server.Query.Ports,
				)
				if errorCode != common.ErrSuccess {
					return
				} else if selectedServerIp != "" {
					serverHost = selectedServerIp
					serverStart = "false"
					if serverStop == "" && (!isAdmin || runtime.GOOS == "windows") {
						serverStop = "false"
					}
				} else {
					serverStart = "true"
					if serverStop == "" {
						serverStop = "true"
					}
				}
			}
			var serverIP string
			if serverStart == "false" {
				if serverStop == "true" {
					printer.Println(
						printer.Info,
						printer.TS("Server.Start", printer.OptionStyle),
						printer.T(" is "),
						printer.TS("false", printer.LiteralStyle),
						printer.T(". Ignoring "),
						printer.TS("Server.Stop", printer.OptionStyle),
						printer.T(" being "),
						printer.TS("true", printer.LiteralStyle),
					)
				}
				if serverHost == "" {
					printer.Println(
						printer.Error,
						printer.TS("Server.Start", printer.OptionStyle),
						printer.T(" is "),
						printer.TS("false", printer.LiteralStyle),
						printer.T(". "),
						printer.TS("Server.Host", printer.OptionStyle),
						printer.T(" must be fulfilled to know which host to connect to."),
					)
					errorCode = internal.ErrInvalidServerHost
					return
				}
				if addr, err := netip.ParseAddr(serverHost); err == nil && addr.Is6() {
					printer.Println(
						printer.Error,
						printer.TS("Server.Start", printer.OptionStyle),
						printer.T(" is "),
						printer.TS("false", printer.LiteralStyle),
						printer.T(". "),
						printer.TS("Server.Host", printer.OptionStyle),
						printer.T(" must be fulfilled with a host or IPv4 address, not an IPv6 address."),
					)
					errorCode = internal.ErrInvalidServerHost
					return
				}
				if !server.CheckConnectionFromServer(serverHost, true) {
					printer.Println(
						printer.Error,
						printer.TS("Server.Start", printer.OptionStyle),
						printer.T(" is "),
						printer.TS("false", printer.LiteralStyle),
						printer.T(". "),
						printer.TS(serverHost, printer.LiteralStyle),
						printer.T(" must be reachable. Review the host is correct, the "),
						printer.TS("server", printer.ComponentStyle),
						printer.T(" is started and you can connect to TCP port 443 (HTTPS)."),
					)
					errorCode = internal.ErrInvalidServerStart
					return
				}
				if serverIPs, data := server.FilterServerIPs(serverId, config.GameTitle(), launcherCommon.HostOrIpToIps(serverHost)); data == nil {
					printer.Println(
						printer.Error,
						printer.TS("Server.Start", printer.OptionStyle),
						printer.T(" is "),
						printer.TS("false", printer.LiteralStyle),
						printer.T(". Failed to connect to "),
						printer.TS(serverHost, printer.LiteralStyle),
						printer.T("."),
					)
					errorCode = internal.ErrInvalidServerStart
					return
				} else {
					serverIP = serverIPs[0].Ip
				}
			} else {
				errorCode, serverIP = config.StartServer(serverExecutable, serverArgs, serverStop == "true", serverId, storeToAddCertificate != "false")
				if errorCode != common.ErrSuccess {
					return
				}
			}
			serverCertificate := server.ReadCertificateFromServer(serverIP)
			if serverCertificate == nil {
				printer.Println(
					printer.Error,
					printer.T("Failed to read the certificate from "),
					printer.TS(serverIP, printer.LiteralStyle),
					printer.T(". Try to access it with your browser and checking the certificate."),
				)
				errorCode = internal.ErrReadCert
				return
			}
			errorCode = config.MapHosts(serverIP, storeToAddHosts)
			if errorCode != common.ErrSuccess {
				return
			}
			errorCode = config.AddCert(serverId, serverCertificate, storeToAddCertificate)
			if errorCode != common.ErrSuccess {
				return
			}
			var executableArgs []string
			executableArgs = config.ParseGameArguments(clientArgs)
			errorCode = config.IsolateUserData(parameters.Isolation.WindowsUserProfilePath, isolateMetadata)
			if errorCode != common.ErrSuccess {
				return
			}
			errorCode = config.LaunchAgent(executer, rebroadcastIPs)
			if errorCode != common.ErrSuccess {
				return
			}
			errorCode = config.LaunchGame(executer, customExecutor, executableArgs)
			if errorCode == common.ErrSuccess {
				printer.PrintSimpln(
					printer.AllDone,
					"All done!",
				)
			}
		},
	}
)

func Execute() error {
	launcher.SetDefaults(v)
	cobra.OnInitialize(initConfig)
	rootCmd.Version = Version
	rootCmd.PersistentFlags().StringSliceVar(&cfgFiles, "config", []string{}, fmt.Sprintf(`Config files. Default config.toml in %s directories`, strings.Join(configPaths, ", ")))
	rootCmd.PersistentFlags().Bool("debug", v.GetBool("Debug"), `Whether to show debug information in the console.`)
	storeToAddHostStr := "Where to add the host entry if it's needed to connect to the 'server' with the official domain. Including to avoid receiving that it's on maintenance. 'local' for the system-wide store (needs admin), '' (empty) to not add any hosts and 'tmp' for a temporary file."
	if runtime.GOOS == "linux" {
		storeToAddHostStr += ` It follows the Windows format in this case.`
	}
	rootCmd.PersistentFlags().StringP("storeToAddHost", "t", v.GetString("StoreToAddHost"), storeToAddHostStr)
	storeToAddCertificateStr := `Trust the certificate of the 'server' if needed. "false"`
	if runtime.GOOS == "windows" {
		storeToAddCertificateStr += `, "user"`
	}
	storeToAddCertificateStr += ` or "local" (will require admin privileges).`
	rootCmd.PersistentFlags().StringP("storeToAddCertificate", "c", v.GetString("StoreToAddCertificate"), storeToAddCertificateStr)
	if runtime.GOOS == "windows" {
		rootCmd.PersistentFlags().BoolP("disableRebroadcastBattleServer", "b", v.GetBool("RebroadcastBattleServer.Disable"), `Whether to disable rebroadcasting the BattleServer to select interfaces. If not 'true', all if "Interfaces" and "IPs" are unspecified.`)
	}
	var pathNamesInfo string
	if runtime.GOOS == "windows" {
		pathNamesInfo += " Path names need to use double backslashes within single quotes or be within double quotes."
	}
	cmd.GameCommand(rootCmd.PersistentFlags())
	var suffixIsolate string
	if runtime.GOOS == "linux" {
		suffixIsolate = ` When using 'path' in "Client.Launcher" and "Isolation.NoMetadata" is 'false'', "Isolation.WindowsUserProfilePath" must be set.`
	}
	rootCmd.PersistentFlags().BoolP("noIsolateMetadata", "m", v.GetBool("Isolation.NoMetadata"), "Do not isolate the metadata cache of the game. Always 'true' in AoE: DE."+suffixIsolate)
	rootCmd.PersistentFlags().BoolP("serverRunNoStop", "a", v.GetBool("Server.Run.NoStop"), `Do not stop (and show window) if the server is run when the game exists.`)
	rootCmd.PersistentFlags().UintSliceP("serverQueryPorts", "n", v.Get("Server.Query.Ports").([]uint), `Ports to query for servers. If not including the default port, default configured 'servers' will not get discovered.`)
	rootCmd.PersistentFlags().IPSliceP("serverQueryMulticastGroups", "g", v.Get("Server.Query.MulticastGroups").([]net.IP), `Multicast Groups to query for servers. If not including the default group, default configured 'servers' will not get discovered via Multicast.`)
	rootCmd.PersistentFlags().Bool("serverQueryNoBroadcast", v.GetBool("Server.Query.NoBroadcast"), `Do not broadcast the query to select network interfaces when querying for servers.`)
	rootCmd.PersistentFlags().StringP("server", "s", v.GetString("Server.Connect.Host"), `Hostname or IPv4 of the 'server' when 'Server.Mode' is 'connect'.`)
	rootCmd.PersistentFlags().StringP("serverMode", "r", v.GetString("Server.Mode"), `'connect' for a pre-defined server, 'run' to execute a new server or empty to both query for servers and have the option to 'run'.`)
	clientExeTip := `The type of game launcher. If not set, it will use Steam`
	if runtime.GOOS == "windows" {
		clientExeTip += ` and then the Xbox one if found`
	}
	clientExeTip += `. Specify "steam"`
	if runtime.GOOS == "windows" {
		clientExeTip += ` or "msstore"`
	}
	clientExeTip += ` to use the default launcher, or "path", but then you need to specify "Client.PathCommand".`
	if runtime.GOOS == "linux" {
		clientExeTip += ` If using "path", you need to specify "Isolation.WindowsUserProfilePath".`
	}
	rootCmd.PersistentFlags().StringP("clientLauncher", "l", v.GetString("Client.Launcher"), clientExeTip)
	if err := v.BindPFlag("GameTitle", rootCmd.PersistentFlags().Lookup(commonCmd.Name)); err != nil {
		return err
	}
	if err := v.BindPFlag("Debug", rootCmd.PersistentFlags().Lookup("debug")); err != nil {
		return err
	}
	if err := v.BindPFlag("StoreToAddHost", rootCmd.PersistentFlags().Lookup("storeToAddHost")); err != nil {
		return err
	}
	if err := v.BindPFlag("StoreToAddCertificate", rootCmd.PersistentFlags().Lookup("storeToAddCertificate")); err != nil {
		return err
	}
	if runtime.GOOS == "windows" {
		if err := v.BindPFlag("RebroadcastBattleServer.Disable", rootCmd.PersistentFlags().Lookup("disableRebroadcastBattleServer")); err != nil {
			return err
		}
	}
	if err := v.BindPFlag("Isolation.NoMetadata", rootCmd.PersistentFlags().Lookup("noIsolateMetadata")); err != nil {
		return err
	}
	if err := v.BindPFlag("Server.Run.NoStop", rootCmd.PersistentFlags().Lookup("serverRunNoStop")); err != nil {
		return err
	}
	if err := v.BindPFlag("Server.Query.Ports", rootCmd.PersistentFlags().Lookup("serverQueryPorts")); err != nil {
		return err
	}
	if err := v.BindPFlag("Server.Query.MulticastGroups", rootCmd.PersistentFlags().Lookup("serverQueryMulticastGroups")); err != nil {
		return err
	}
	if err := v.BindPFlag("Server.Query.NoBroadcast", rootCmd.PersistentFlags().Lookup("serverQueryNoBroadcast")); err != nil {
		return err
	}
	if err := v.BindPFlag("Server.Connect.Host", rootCmd.PersistentFlags().Lookup("server")); err != nil {
		return err
	}
	if err := v.BindPFlag("Server.Mode", rootCmd.PersistentFlags().Lookup("serverMode")); err != nil {
		return err
	}
	return rootCmd.Execute()
}

func initConfig() {
	cfg.InitConfig(
		v,
		configPaths,
		cfgFiles,
		"launcher",
		func(path string) {
			printer.Println(
				printer.Debug,
				printer.T("Using config file: "),
				printer.TS(path, printer.FilePathStyle),
			)
		},
	)
}

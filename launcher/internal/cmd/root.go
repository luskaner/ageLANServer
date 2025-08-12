package cmd

import (
	"fmt"
	"github.com/charmbracelet/huh"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/luskaner/ageLANServer/common"
	commonCmd "github.com/luskaner/ageLANServer/common/cmd"
	cfg "github.com/luskaner/ageLANServer/common/config"
	"github.com/luskaner/ageLANServer/common/config/launcher"
	commonExecutor "github.com/luskaner/ageLANServer/common/executor"
	"github.com/luskaner/ageLANServer/common/pidLock"
	commonProcess "github.com/luskaner/ageLANServer/common/process"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/luskaner/ageLANServer/launcher-common/cmd"
	"github.com/luskaner/ageLANServer/launcher/internal"
	"github.com/luskaner/ageLANServer/launcher/internal/cmdUtils"
	"github.com/luskaner/ageLANServer/launcher/internal/cmdUtils/printer"
	"github.com/luskaner/ageLANServer/launcher/internal/executor"
	"github.com/luskaner/ageLANServer/launcher/internal/game"
	"github.com/luskaner/ageLANServer/launcher/internal/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
		Short: "launcher discovers and configures AoE 1, AoE 2 and AoE 3 (all DE) to connect to the local LAN server",
		Long:  "launcher discovers or starts a local LAN server, configures and executes the game launcher to connect to it",
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
			config.SetGameTitle(parameters.Client.GameTitle)
			config.SetIPProtocol(parameters.Server.IPProtocol)
			if parameters.Server.Mode == common.ModeQueryOrRun &&
				parameters.Server.IPProtocol.IPv6() &&
				!parameters.Server.Query.IPv6.DisableLinkLocal {
				parameters.Server.Query.IPv6.MulticastGroups.Add(netip.IPv6LinkLocalAllNodes())
			}
			if parameters.Client.Launcher == common.ClientLauncherPath {
				parameters.Client.PathCommand[0] = cmdUtils.CommandArgs([]string{parameters.Client.PathCommand[0]}, nil)[0]
			}
			if parameters.Server.Mode != common.ModeConnect && len(parameters.Server.Run.Command) > 0 && parameters.Server.Run.Command[0] != "" {
				if err = validate.Var(parameters.Server.Run.Command[0], "file"); err != nil {
					return err
				}
			}
			parameters.SetupCommand = cmdUtils.CommandArgs(parameters.SetupCommand, nil)
			parameters.RevertCommand = cmdUtils.CommandArgs(parameters.RevertCommand, nil)
			lock = &pidLock.Lock{}
			if err := lock.Lock(); err != nil {
				return fmt.Errorf("failed to lock PID file: %w, make sure the launcher is not already running", err)
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
			if !cmdUtils.GamesRunning().IsEmpty() {
				confirm := true
				if err := huh.NewConfirm().
					Description("The launcher making configuration changes can interfere with it.").
					Title("An age game is already running, continue anyway?").
					Affirmative("Continue").
					Negative("Exit").
					Value(&confirm).
					WithTheme(huh.ThemeBase()).
					Run(); err == nil {
					if !confirm {
						return
					}
				}
			}
			rebroadcastIPAddrs := mapset.NewThreadUnsafeSet[netip.Addr]()
			if runtime.GOOS == "windows" && !parameters.Client.RebroadcastBattleServer.Disable {
				if !parameters.Client.RebroadcastBattleServer.IPAddrs.IsEmpty() && !parameters.Client.RebroadcastBattleServer.Interfaces.IsEmpty() {
					printer.Println(
						printer.Warning,
						printer.T("Specifying both "),
						printer.TS("RebroadcastBattleServer.IPAddrs", printer.OptionStyle),
						printer.T(" and "),
						printer.TS("RebroadcastBattleServer.Interfaces", printer.OptionStyle),
						printer.T(", only "),
						printer.TS("RebroadcastBattleServer.IPAddrs", printer.OptionStyle),
						printer.T(" will be used."),
					)
				}
				rebroadcastIPAddrs = game.RebroadcastIPAddrs(
					parameters.Client.RebroadcastBattleServer.IPAddrs,
					parameters.Client.RebroadcastBattleServer.Interfaces,
				)
			}
			var serverStart cmdUtils.TriStateBool
			var serverStop cmdUtils.TriStateBool
			switch parameters.Server.Mode {
			case common.ModeConnect:
				serverStart.Set(false)
				serverStop.Set(false)
			case common.ModeRun:
				serverStart.Set(true)
				serverStop.Set(!parameters.Server.Run.NoStop)
			}

			var isolateMetadata bool
			if config.GameTitle() != common.AoE1 {
				isolateMetadata = !parameters.Client.Isolation.NoMetadata
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
				printer.TS(parameters.Client.GameTitle.Description(), printer.LiteralStyle),
				printer.T(`... `),
			)
			var clientExecutable string
			if parameters.Client.Launcher == common.ClientLauncherPath {
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
			if mismatchHosts := cmdUtils.InternalExternalDnsMismatch(
				config.IPProtocol().IPv4(),
				config.IPProtocol().IPv6(),
			); !mismatchHosts.IsEmpty() {
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
			if serverStart.Unset() {
				var selectedServerIpAddr netip.Addr
				errorCode, serverId, selectedServerIpAddr = cmdUtils.DiscoverServersAndSelectBestIpAddr(
					config.GameTitle(),
					parameters.Server.Query.IPv4.MulticastGroups,
					parameters.Server.Query.IPv4.Ports,
					parameters.Server.Query.IPv6.MulticastGroups,
					parameters.Server.Query.IPv6.Ports,
					!parameters.Server.Query.IPv4.DisableBroadcast,
					config.IPProtocol(),
				)
				if errorCode != common.ErrSuccess {
					return
				} else if selectedServerIpAddr.IsValid() {
					serverHost = selectedServerIpAddr.String()
					serverStart.Set(false)
					if serverStop.Unset() && (!isAdmin || runtime.GOOS == "windows") {
						serverStop.Set(false)
					}
				} else {
					serverStart.Set(true)
					if serverStop.Unset() {
						serverStop.Set(true)
					}
				}
			}
			var serverIPAddr netip.Addr
			if serverStart.False() {
				if !server.CheckConnectionFromServer(serverHost, config.IPProtocol(), true) {
					printer.Println(
						printer.Error,
						printer.TS("Server.Mode", printer.OptionStyle),
						printer.T(" is "),
						printer.TS("connect", printer.LiteralStyle),
						printer.T(". "),
						printer.TS(serverHost, printer.LiteralStyle),
						printer.T(" must be reachable. Review the host is correct, the "),
						printer.TS("server", printer.ComponentStyle),
						printer.T(" is started and you can connect to TCP port 443 (HTTPS)."),
					)
					errorCode = internal.ErrInvalidServerStart
					return
				}
				if measuredServerIPAddrs, data := server.FilterServerIPs(
					serverId,
					serverHost,
					config.GameTitle(),
					launcherCommon.AddrToIpAddrs(serverHost, config.IPProtocol().IPv4(), config.IPProtocol().IPv6()),
				); data == nil {
					printer.Println(
						printer.Error,
						printer.TS("Server.Mode", printer.OptionStyle),
						printer.T(" is "),
						printer.TS("connect", printer.LiteralStyle),
						printer.T(". Failed to connect to "),
						printer.TS(serverHost, printer.LiteralStyle),
						printer.T("."),
					)
					errorCode = internal.ErrInvalidServerStart
					return
				} else {
					serverIPAddr = measuredServerIPAddrs[0].IpAddr
				}
			} else {
				errorCode, serverIPAddr = config.StartServer(serverExecutable, serverArgs, serverStop.True(), serverId, !parameters.StoreToAddCertificate.IsNone())
				if errorCode != common.ErrSuccess {
					return
				}
			}
			serverCertificate := server.ReadCertificateFromServer(serverIPAddr.String(), config.IPProtocol())
			if serverCertificate == nil {
				printer.Println(
					printer.Error,
					printer.T("Failed to read the certificate from "),
					printer.TS(serverIPAddr.String(), printer.LiteralStyle),
					printer.T(". Try to access it with your browser and checking the certificate."),
				)
				errorCode = internal.ErrReadCert
				return
			}
			errorCode = config.MapHosts(serverIPAddr, parameters.StoreToAddHost)
			if errorCode != common.ErrSuccess {
				return
			}
			errorCode = config.AddCert(serverId, serverCertificate, parameters.StoreToAddCertificate)
			if errorCode != common.ErrSuccess {
				return
			}
			var executableArgs []string
			executableArgs = config.ParseGameArguments(clientArgs)
			errorCode = config.IsolateUserData(parameters.Client.Isolation.WindowsUserProfilePath, isolateMetadata)
			if errorCode != common.ErrSuccess {
				return
			}
			errorCode = config.LaunchAgent(executer, rebroadcastIPAddrs)
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
	var storeToAddHost common.LauncherStore
	rootCmd.PersistentFlags().VarP(&storeToAddHost, "storeToAddHost", "t", storeToAddHostStr)
	storeToAddCertificateStr := `Trust the certificate of the 'server' if needed. "false"`
	if runtime.GOOS == "windows" {
		storeToAddCertificateStr += `, "user"`
	}
	storeToAddCertificateStr += ` or "local" (will require admin privileges).`
	var storeToAddCertificate common.LauncherUserStore
	rootCmd.PersistentFlags().VarP(&storeToAddCertificate, "storeToAddCertificate", "c", storeToAddCertificateStr)
	if runtime.GOOS == "windows" {
		rootCmd.PersistentFlags().BoolP("disableRebroadcastBattleServer", "b", v.GetBool("RebroadcastBattleServer.Disable"), `Whether to disable rebroadcasting the BattleServer to select interfaces. If not 'true', all if "Interfaces" and "IPAddrs" are unspecified.`)
	}
	var pathNamesInfo string
	if runtime.GOOS == "windows" {
		pathNamesInfo += " Path names need to use double backslashes within single quotes or be within double quotes."
	}
	var gameTitle common.GameTitle
	cmd.GameVarCommand(rootCmd.PersistentFlags(), &gameTitle)
	var suffixIsolate string
	if runtime.GOOS == "linux" {
		suffixIsolate = ` When using 'path' in "Client.Launcher" and "Isolation.NoMetadata" is 'false'', "Isolation.WindowsUserProfilePath" must be set.`
	}
	rootCmd.PersistentFlags().BoolP("noIsolateMetadata", "m", v.GetBool("Isolation.NoMetadata"), "Do not isolate the metadata cache of the game. Always 'true' in AoE: DE."+suffixIsolate)
	rootCmd.PersistentFlags().BoolP("serverRunNoStop", "a", v.GetBool("Server.Run.NoStop"), `Do not stop (and show window) if the server is run when the game exists.`)
	rootCmd.PersistentFlags().StringP("server", "s", v.GetString("Server.Connect.Host"), `Hostname or IPv4 of the 'server' when 'Server.Mode' is 'connect'.`)
	var mode common.LauncherServerMode
	rootCmd.PersistentFlags().VarP(&mode, "serverMode", "r", `'connect' for a pre-defined server, 'run' to execute a new server or empty to both query for servers and have the option to 'run'.`)
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
	if err := v.BindPFlag("Client.GameTitle", rootCmd.PersistentFlags().Lookup(commonCmd.Name)); err != nil {
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
		if err := v.BindPFlag("Client.RebroadcastBattleServer.Disable", rootCmd.PersistentFlags().Lookup("disableRebroadcastBattleServer")); err != nil {
			return err
		}
	}
	if err := v.BindPFlag("Client.Isolation.NoMetadata", rootCmd.PersistentFlags().Lookup("noIsolateMetadata")); err != nil {
		return err
	}
	if err := v.BindPFlag("Server.Run.NoStop", rootCmd.PersistentFlags().Lookup("serverRunNoStop")); err != nil {
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

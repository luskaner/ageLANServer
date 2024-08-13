package cmd

import (
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/inconshreveable/mousetrap"
	"github.com/luskaner/aoe2DELanServer/common"
	"github.com/luskaner/aoe2DELanServer/common/pidLock"
	commonExecutor "github.com/luskaner/aoe2DELanServer/launcher-common/executor"
	"github.com/luskaner/aoe2DELanServer/launcher/internal"
	"github.com/luskaner/aoe2DELanServer/launcher/internal/cmdUtils"
	"github.com/luskaner/aoe2DELanServer/launcher/internal/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/sys/windows"
	"mvdan.cc/sh/v3/shell"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const autoValue = "auto"
const trueValue = "true"
const falseValue = "false"

var configPaths = []string{"resources", "."}
var config = &cmdUtils.Config{}

func parseCommandArgs(name string) (args []string, err error) {
	args, err = shell.Fields(viper.GetString(name), nil)
	return
}

var (
	Version                        string
	cfgFile                        string
	autoTrueFalseValues            = mapset.NewSet[string](autoValue, trueValue, falseValue)
	canTrustCertificateValues      = mapset.NewSet[string](falseValue, "user", "local")
	canBroadcastBattleServerValues = mapset.NewSet[string](autoValue, falseValue)
	rootCmd                        = &cobra.Command{
		Use:   filepath.Base(os.Args[0]),
		Short: "launcher discovers and configures AoE 2:DE to connect to the local LAN server",
		Long:  "launcher discovers or starts a local LAN server, optionally isolates the user data, configures the local DNS server, HTTPS certificate and finally launches the game launcher",
		Run: func(_ *cobra.Command, _ []string) {
			lock := &pidLock.Lock{}
			if err := lock.Lock(); err != nil {
				fmt.Println("Failed to lock pid file. You may try checking if the process in PID file exists (and killing it).")
				os.Exit(common.ErrPidLock)
			}
			errorMayBeConfig := false
			var errorCode = common.ErrSuccess
			defer func() {
				_ = lock.Unlock()
				os.Exit(errorCode)
			}()
			canTrustCertificate := viper.GetString("Config.CanTrustCertificate")
			if !canTrustCertificateValues.Contains(canTrustCertificate) {
				fmt.Printf("Invalid value for canTrustCertificate (local/user/false): %s\n", canTrustCertificate)
				errorCode = internal.ErrInvalidCanTrustCertificate
				return
			}
			canBroadcastBattleServer := viper.GetString("Config.CanBroadcastBattleServer")
			if !canBroadcastBattleServerValues.Contains(canBroadcastBattleServer) {
				fmt.Printf("Invalid value for canBroadcastBattleServer (auto/false): %s\n", canBroadcastBattleServer)
				errorCode = internal.ErrInvalidCanBroadcastBattleServer
				return
			}
			serverStart := viper.GetString("Server.Start")
			if !autoTrueFalseValues.Contains(serverStart) {
				fmt.Printf("Invalid value for serverStart (auto/true/false): %s\n", serverStart)
				errorCode = internal.ErrInvalidServerStart
				return
			}
			serverStop := viper.GetString("Server.Stop")
			if !autoTrueFalseValues.Contains(serverStop) {
				fmt.Printf("Invalid value for serverStop (auto/true/false): %s\n", serverStop)
				errorCode = internal.ErrInvalidServerStop
				return
			}
			serverArgs, err := parseCommandArgs("Server.ExecutableArgs")
			if err != nil {
				fmt.Println("Failed to parse server executable arguments")
				errorCode = internal.ErrInvalidServerArgs
				return
			}
			var clientArgs []string
			clientArgs, err = parseCommandArgs("Client.ExecutableArgs")
			if err != nil {
				fmt.Println("Failed to parse server executable arguments")
				errorCode = internal.ErrInvalidClientArgs
				return
			}
			canAddHost := viper.GetBool("Config.CanAddHost")
			isolateMetadata := viper.GetBool("Config.IsolateMetadata")
			isolateProfiles := viper.GetBool("Config.IsolateProfiles")
			serverExecutable := viper.GetString("Server.Executable")
			clientExecutable := viper.GetString("Client.Executable")
			serverHost := viper.GetString("Server.Host")
			isAdmin := commonExecutor.IsAdmin()

			if isAdmin {
				fmt.Println("Running as administrator, this is not recommended for security reasons. It will request isolated admin privileges if/when needed.")
			}

			sigs := make(chan os.Signal, 1)
			signal.Notify(sigs, windows.SIGINT, windows.SIGTERM)
			go func() {
				_, ok := <-sigs
				if ok {
					if config.AgentStarted() {
						config.KillAgent()
						config.Revert()
					}
					_ = lock.Unlock()
					os.Exit(errorCode)
				}
			}()

			defer func() {
				if errorCode == common.ErrSuccess {
					fmt.Print("Program finished successfully")
					if mousetrap.StartedByExplorer() {
						fmt.Println(", closing in 10 seconds...")
						time.Sleep(10 * time.Second)
					}
				} else {
					config.Revert()
					fmt.Print("Program finished with errors")
					if errorMayBeConfig {
						fmt.Print(", you may try running \"cleanup.bat\" as regular user")
					}
					if mousetrap.StartedByExplorer() {
						fmt.Println(", press the Enter key to exit...")
						_, _ = fmt.Scanln()
					}
				}
			}()
			if cmdUtils.GameRunning() {
				errorCode = internal.ErrGameAlreadyRunning
				return
			}
			// Setup
			fmt.Println("Setting up...")
			if serverStart == "auto" {
				announcePorts := viper.GetStringSlice("Server.AnnouncePorts")
				portsInt := make([]int, len(announcePorts))
				for i, str := range announcePorts {
					if portInt, err := strconv.Atoi(str); err == nil {
						portsInt[i] = portInt
					} else {
						fmt.Printf(`Invalid announce port "%s"\n`, str)
						errorCode = internal.ErrAnnouncementPort
						return
					}
				}
				fmt.Printf("Waiting 15 seconds for server announcements on LAN on port(s) %s (we are v. %d), you might need to allow 'launcher.exe' in the firewall...\n", strings.Join(announcePorts, ", "), common.AnnounceVersionLatest)
				errorCode, selectedServerHost := cmdUtils.ListenToServerAnnouncementsAndSelect(portsInt)
				if errorCode != common.ErrSuccess {
					return
				} else if selectedServerHost != "" {
					serverHost = selectedServerHost
					serverStart = "false"
					if serverStop == "auto" {
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
					fmt.Println("serverStart is false. Ignoring serverStop being true.")
				}
				if serverHost == "" {
					fmt.Println("serverStart is false. serverHost must be fulfilled as it is needed to know which host to connect to.")
					errorCode = internal.ErrInvalidServerHost
					return
				}
				if !server.CheckConnectionFromServer(serverHost, true) {
					fmt.Println("serverStart is false. " + serverHost + " must be reachable. Review the host is correct, the server is started and you can connect to TCP port 443 (HTTPS).")
					errorCode = internal.ErrInvalidServerStart
					errorMayBeConfig = true
					return
				}
			} else {
				errorCode, serverHost = config.StartServer(serverExecutable, serverArgs, serverStop == "true", canTrustCertificate != "false")
				if errorCode != common.ErrSuccess {
					return
				}
			}
			errorCode = config.MapHosts(serverHost, canAddHost)
			if errorCode != common.ErrSuccess {
				errorMayBeConfig = true
				return
			}
			errorCode = config.AddCert(canTrustCertificate)
			if errorCode != common.ErrSuccess {
				errorMayBeConfig = true
				return
			}
			errorCode = config.IsolateUserData(isolateMetadata, isolateProfiles)
			if errorCode != common.ErrSuccess {
				errorMayBeConfig = true
				return
			}
			errorCode = config.LaunchAgentAndGame(clientExecutable, clientArgs, canTrustCertificate, canBroadcastBattleServer)
		},
	}
)

func Execute() error {
	cobra.OnInitialize(initConfig)
	rootCmd.Version = Version
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", fmt.Sprintf(`config file (default config.ini in %s directories)`, strings.Join(configPaths, ", ")))
	rootCmd.PersistentFlags().BoolP("canAddHost", "t", true, "Add a local dns entry if it's needed to connect to the server with the official domain. Will require admin privileges.")
	rootCmd.PersistentFlags().StringP("canTrustCertificate", "c", "local", `Trust the certificate of the server if needed. "false", "user" or "local" (will require admin privileges)`)
	rootCmd.PersistentFlags().StringP("canBroadcastBattleServer", "b", "auto", `Whether or not to broadcast the game BattleServer to all interfaces in LAN (not just the most priority one)`)
	rootCmd.PersistentFlags().BoolP("isolateMetadata", "m", true, "Isolate the metadata cache of the game, otherwise, it will be shared.")
	rootCmd.PersistentFlags().BoolP("isolateProfiles", "p", false, "(Experimental) Isolate the users profile of the game, otherwise, it will be shared.")
	rootCmd.PersistentFlags().StringP("serverStart", "a", "auto", `Start the server if needed, "auto" will start a server if one is not already running, "true" (will start a server regardless if one is already running), "false" (will require an already running server).`)
	rootCmd.PersistentFlags().StringP("serverStop", "o", "auto", `Stop the server if started, "auto" will stop the server if one was started, "false" (will not stop the server regardless if one was started), "true" (will not stop the server even if it was started).`)
	rootCmd.PersistentFlags().StringSliceP("serverAnnouncePorts", "n", []string{strconv.Itoa(common.AnnouncePort)}, `Announce ports to listen to. If not including the default port, default configured servers will not get discovered.`)
	rootCmd.PersistentFlags().StringP("server", "s", "", `Hostname of the server to connect to. If not absent, serverStart will be assumed to be false. Ignored otherwise`)
	serverExe := common.GetExeFileName(false, common.Server)
	rootCmd.PersistentFlags().StringP("serverPath", "e", "auto", fmt.Sprintf(`The executable path of the server, "auto", will be try to execute in this order ".\%s", ".\%s\%s", "..\%s" and finally "..\%s\%s", otherwise set the path (relative or absolute).`, serverExe, common.Server, serverExe, serverExe, common.Server, serverExe))
	rootCmd.PersistentFlags().StringP("serverPathArgs", "r", "[]string{}", `The arguments to pass to the server executable if starting it. Execute the server help flag for available arguments. Path names need to use double backslashes or be within single quotes.`)
	rootCmd.PersistentFlags().StringP("clientExe", "l", "auto", `The type of game client or the path. "auto" will use the Steam and then the Microsoft Store one if found. Use a path to the game launcher, "steam" or "msstore" to use the default launcher.`)
	rootCmd.PersistentFlags().StringP("clientExeArgs", "i", "", `The arguments to pass to the client launcher if it is custom. Path names need to use double backslashes or be within single quotes.`)
	if err := viper.BindPFlag("Config.CanAddHost", rootCmd.PersistentFlags().Lookup("canAddHost")); err != nil {
		return err
	}
	if err := viper.BindPFlag("Config.CanTrustCertificate", rootCmd.PersistentFlags().Lookup("canTrustCertificate")); err != nil {
		return err
	}
	if err := viper.BindPFlag("Config.CanBroadcastBattleServer", rootCmd.PersistentFlags().Lookup("canBroadcastBattleServer")); err != nil {
		return err
	}
	if err := viper.BindPFlag("Config.IsolateMetadata", rootCmd.PersistentFlags().Lookup("isolateMetadata")); err != nil {
		return err
	}
	if err := viper.BindPFlag("Config.IsolateProfiles", rootCmd.PersistentFlags().Lookup("isolateProfiles")); err != nil {
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
	return rootCmd.Execute()
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		for _, configPath := range configPaths {
			viper.AddConfigPath(configPath)
		}
		viper.SetConfigType("ini")
		viper.SetConfigName("config")
	}
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

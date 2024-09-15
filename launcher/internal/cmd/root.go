package cmd

import (
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/inconshreveable/mousetrap"
	"github.com/luskaner/aoe2DELanServer/common"
	"github.com/luskaner/aoe2DELanServer/common/pidLock"
	commonExecutor "github.com/luskaner/aoe2DELanServer/launcher-common/executor/exec"
	"github.com/luskaner/aoe2DELanServer/launcher/internal"
	"github.com/luskaner/aoe2DELanServer/launcher/internal/cmdUtils"
	"github.com/luskaner/aoe2DELanServer/launcher/internal/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"mvdan.cc/sh/v3/shell"
	"net/netip"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const autoValue = "auto"
const trueValue = "true"
const falseValue = "false"

var reWinToLinVar = regexp.MustCompile(`%(\w+)%`)
var configPaths = []string{"resources", "."}
var config = &cmdUtils.Config{}

func parseCommandArgs(name string) (args []string, err error) {
	cmd := viper.GetString(name)
	if runtime.GOOS == "windows" {
		cmd = reWinToLinVar.ReplaceAllString(cmd, `$$$1`)
	}
	args, err = shell.Fields(cmd, nil)
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
				fmt.Println(err.Error())
				os.Exit(common.ErrPidLock)
			}
			isAdmin := commonExecutor.IsAdmin()
			errorMayBeConfig := false
			var errorCode = common.ErrSuccess
			defer func() {
				_ = lock.Unlock()
				os.Exit(errorCode)
			}()
			canTrustCertificate := viper.GetString("Config.CanTrustCertificate")
			if runtime.GOOS != "windows" && runtime.GOOS != "darwin" {
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
			var setupCommand []string
			setupCommand, err = parseCommandArgs("Config.SetupCommand")
			if err != nil {
				fmt.Println("Failed to parse setup command")
				errorCode = internal.ErrInvalidSetupCommand
				return
			}
			var revertCommand []string
			revertCommand, err = parseCommandArgs("Config.RevertCommand")
			if err != nil {
				fmt.Println("Failed to parse revert command")
				errorCode = internal.ErrInvalidRevertCommand
				return
			}
			config.SetRevertCommand(revertCommand)
			canAddHost := viper.GetBool("Config.CanAddHost")
			isolateMetadata := viper.GetBool("Config.IsolateMetadata")
			isolateProfiles := viper.GetBool("Config.IsolateProfiles")
			serverExecutable := viper.GetString("Server.Executable")
			clientExecutable := viper.GetString("Client.Executable")
			serverHost := viper.GetString("Server.Host")

			if isAdmin {
				fmt.Println("Running as administrator, this is not recommended for security reasons. It will request isolated admin privileges if/when needed.")
				if runtime.GOOS != "windows" {
					fmt.Println("It can also cause issues and restrict the functionality.")
				}
			}

			if runtime.GOOS != "windows" && isAdmin && (clientExecutable == "auto" || clientExecutable == "steam") {
				fmt.Println("Steam cannot be run as administrator. Either run this as a normal user o set Client.Executable to a custom launcher.")
				errorCode = internal.ErrSteamRoot
				return
			}

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
						fmt.Print(", you may try running \"cleanup\" as regular user")
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
			alreadySelectedIp := false
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
				fmt.Printf("Waiting 15 seconds for server announcements on LAN on port(s) %s (we are v. %d), you might need to allow 'launcher' in the firewall...\n", strings.Join(announcePorts, ", "), common.AnnounceVersionLatest)
				errorCode, selectedServerIp := cmdUtils.ListenToServerAnnouncementsAndSelectBestIp(portsInt)
				if errorCode != common.ErrSuccess {
					return
				} else if selectedServerIp != "" {
					serverHost = selectedServerIp
					alreadySelectedIp = true
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
			errorCode = config.MapHosts(serverHost, canAddHost, alreadySelectedIp)
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
	rootCmd.PersistentFlags().BoolP("canAddHost", "t", true, "Add a local dns entry if it's needed to connect to the server with the official domain. Including to avoid receiving that it's on maintenance. Will require admin privileges.")
	canTrustCertificateStr := `Trust the certificate of the server if needed. "false"`
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		canTrustCertificateStr += `, "user"`
	}
	canTrustCertificateStr += ` or local (will require admin privileges)`
	rootCmd.PersistentFlags().StringP("canTrustCertificate", "c", "local", canTrustCertificateStr)
	if runtime.GOOS == "windows" {
		rootCmd.PersistentFlags().StringP("canBroadcastBattleServer", "b", "auto", `Whether or not to broadcast the game BattleServer to all interfaces in LAN (not just the most priority one)`)
	}
	var pathNamesInfo string
	if runtime.GOOS == "windows" {
		pathNamesInfo += " Path names need to use double backslashes or be within single quotes."
	}
	rootCmd.PersistentFlags().BoolP("isolateMetadata", "m", true, "Isolate the metadata cache of the game, otherwise, it will be shared.")
	rootCmd.PersistentFlags().BoolP("isolateProfiles", "p", false, "(Experimental) Isolate the users profile of the game, otherwise, it will be shared.")
	rootCmd.PersistentFlags().String("setupCommand", "", `Executable to run (including arguments) to run first after the "Setting up..." line. The command must return a 0 exit code to continue. If you need to keep it running spawn a new separate process. You may use environment variables.`+pathNamesInfo)
	rootCmd.PersistentFlags().String("revertCommand", "", `Executable to run (including arguments) to run after setupCommand, game has exited and everything has been reverted. It may run before if there is an error. You may use environment variables.`+pathNamesInfo)
	rootCmd.PersistentFlags().StringP("serverStart", "a", "auto", `Start the server if needed, "auto" will start a server if one is not already running, "true" (will start a server regardless if one is already running), "false" (will require an already running server).`)
	rootCmd.PersistentFlags().StringP("serverStop", "o", "auto", `Stop the server if started, "auto" will stop the server if one was started, "false" (will not stop the server regardless if one was started), "true" (will not stop the server even if it was started).`)
	rootCmd.PersistentFlags().StringSliceP("serverAnnouncePorts", "n", []string{strconv.Itoa(common.AnnouncePort)}, `Announce ports to listen to. If not including the default port, default configured servers will not get discovered.`)
	rootCmd.PersistentFlags().StringP("server", "s", "", `Hostname of the server to connect to. If not absent, serverStart will be assumed to be false. Ignored otherwise`)
	serverExe := common.GetExeFileName(false, common.Server)
	rootCmd.PersistentFlags().StringP("serverPath", "e", "auto", fmt.Sprintf(`The executable path of the server, "auto", will be try to execute in this order "./%s", "./%s/%s", "../%s" and finally "../%s/%s", otherwise set the path (relative or absolute).`, serverExe, common.Server, serverExe, serverExe, common.Server, serverExe))
	rootCmd.PersistentFlags().StringP("serverPathArgs", "r", "", `The arguments to pass to the server executable if starting it. Execute the server help flag for available arguments. You may use environment variables.`+pathNamesInfo)
	clientExeTip := `The type of game client or the path. "auto" will use Steam`
	if runtime.GOOS == "windows" {
		clientExeTip += ` and then the Microsoft Store one if found`
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
	rootCmd.PersistentFlags().StringP("clientExe", "l", "auto", clientExeTip)
	rootCmd.PersistentFlags().StringP("clientExeArgs", "i", "", "The arguments to pass to the client launcher if it is custom. You may use environment variables."+pathNamesInfo)
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
	if err := viper.BindPFlag("Config.IsolateMetadata", rootCmd.PersistentFlags().Lookup("isolateMetadata")); err != nil {
		return err
	}
	if err := viper.BindPFlag("Config.IsolateProfiles", rootCmd.PersistentFlags().Lookup("isolateProfiles")); err != nil {
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

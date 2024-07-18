package cmd

import (
	"common"
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/sys/windows"
	commonExecutor "launcher-common/executor"
	"launcher/internal"
	"launcher/internal/cmd"
	"launcher/internal/server"
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
var config = &cmd.Config{}

var (
	Version                   string
	cfgFile                   string
	autoTrueFalseValues       = mapset.NewSet[string](autoValue, trueValue, falseValue)
	canTrustCertificateValues = mapset.NewSet[string](falseValue, "user", "local")
	rootCmd                   = &cobra.Command{
		Use:   filepath.Base(os.Args[0]),
		Short: "launcher discovers and configures AoE 2:DE to connect to the local LAN server",
		Long:  "launcher discovers or starts a local LAN server, optionally isolates the user data, configures the local DNS server, HTTPS certificate and finally launches the game launcher",
		Run: func(_ *cobra.Command, _ []string) {
			var errorCode = common.ErrSuccess
			defer func() {
				os.Exit(errorCode)
			}()
			canTrustCertificate := viper.GetString("Config.CanTrustCertificate")
			if !canTrustCertificateValues.Contains(canTrustCertificate) {
				fmt.Printf("Invalid value for canTrustCertificate (local/user/false): %s\n", canTrustCertificate)
				errorCode = internal.ErrInvalidCanTrustCertificate
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

			var watcherPid uint32 = 0

			sigs := make(chan os.Signal, 1)
			signal.Notify(sigs, windows.SIGINT, windows.SIGTERM)
			go func() {
				_, ok := <-sigs
				if ok {
					if watcherPid > 0 {
						config.KillWatcher()
						config.Revert()
					}
					os.Exit(errorCode)
				}
			}()
			defer func() {
				if errorCode == common.ErrSuccess {
					fmt.Println("Program finished succesfully, closing in 10 seconds...")
					time.Sleep(10 * time.Second)
				} else {
					config.Revert()
					fmt.Println("Program finished with errors you may try running \"cleanup.bat\" as administrator, press the Enter key to exit...")
					_, _ = fmt.Scanln()
				}
			}()
			if cmd.GameRunning() {
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
				fmt.Printf("Waiting 15 seconds for server announcements on LAN on port(s) %s (we are v. %d)...\n", strings.Join(announcePorts, ", "), common.AnnounceVersionLatest)
				errorCode, serverHost := cmd.ListenToServerAnnouncementsAndSelect(portsInt)
				if errorCode != common.ErrSuccess {
					return
				} else if serverHost != "" {
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
					return
				}
			} else {
				errorCode = config.StartServer(serverExecutable, serverHost, serverStop == "true", canTrustCertificate != "false")
				if errorCode != common.ErrSuccess {
					return
				}
			}
			errorCode = config.MapHosts(serverHost, canAddHost)
			if errorCode != common.ErrSuccess {
				return
			}
			errorCode = config.AddCert(canTrustCertificate)
			if errorCode != common.ErrSuccess {
				return
			}
			errorCode = config.IsolateUserData(isolateMetadata, isolateProfiles)
			if errorCode != common.ErrSuccess {
				return
			}
			errorCode = config.LaunchWatcherAndGame(clientExecutable, canTrustCertificate)
		},
	}
)

func Execute() error {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", fmt.Sprintf(`config file (default config.ini in %s directories)`, strings.Join(configPaths, ", ")))
	rootCmd.PersistentFlags().BoolP("canAddHost", "t", true, "Add a local dns entry if it's needed to connect to the server with the official domain. Will require admin privileges.")
	rootCmd.PersistentFlags().StringP("canTrustCertificate", "c", "local", `Trust the certificate of the server if needed. "false", "user" or "local" (will require admin privileges)`)
	rootCmd.PersistentFlags().BoolP("isolateMetadata", "m", true, "Isolate the metadata cache of the game, otherwise, it will be shared.")
	rootCmd.PersistentFlags().BoolP("isolateProfiles", "p", true, "Isolate the users profile of the game, otherwise, it will be shared.")
	rootCmd.PersistentFlags().StringP("serverStart", "a", "auto", `Start the server if needed, "auto" will start a server if one is not already running, "true" (will start a server regardless if one is already running), "false" (will require an already running server).`)
	rootCmd.PersistentFlags().StringP("serverStop", "o", "auto", `Stop the server if started, "auto" will stop the server if one was started, "false" (will not stop the server regardless if one was started), "true" (will not stop the server even if it was started).`)
	rootCmd.PersistentFlags().StringSliceP("serverAnnouncePorts", "n", []string{strconv.Itoa(common.AnnouncePort)}, `Announce ports to listen to. If not including the default port, default configured servers will not get discovered.`)
	rootCmd.PersistentFlags().StringP("server", "s", "", `Hostname of the server to connect to. If not absent, serverStart will be assumed to be false. Ignored otherwise`)
	serverExe := common.GetExeFileName(common.Server)
	serverScript := common.GetScriptFileName(common.Server)
	rootCmd.PersistentFlags().StringP("serverPath", "e", "auto", fmt.Sprintf(`The executable path of the server, "auto", will be try to execute in this order ".\%s", ".\%s", "..\%s", "..\%s", "..\%s\%s" and finally "..\%s\%s", otherwise set the path (relative or absolute).`, serverScript, serverExe, serverScript, serverExe, common.Server, serverScript, common.Server, serverExe))
	rootCmd.PersistentFlags().StringArrayP("serverPathArgs", "r", []string{}, `The arguments to pass to the server executable if starting it. Can be set multiple times. See the server for available arguments.`)
	rootCmd.PersistentFlags().StringP("clientExe", "l", "auto", `The type of game client or the path. "auto" will use the Steam and then the Microsoft Store one if found. Use a path to the game launcher, "steam" or "msstore" to use the default launcher.`)
	rootCmd.PersistentFlags().StringArrayP("clientExeArgs", "i", []string{}, `The arguments to pass to the client launcher if it is custom. Can be set multiple times.`)
	rootCmd.PersistentFlags().StringVarP(&Version, "version", "v", Version, "Version")
	if err := viper.BindPFlag("Config.CanAddHost", rootCmd.PersistentFlags().Lookup("canAddHost")); err != nil {
		return err
	}
	if err := viper.BindPFlag("Config.CanTrustCertificate", rootCmd.PersistentFlags().Lookup("canTrustCertificate")); err != nil {
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

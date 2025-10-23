package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/netip"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/cmd"
	"github.com/luskaner/ageLANServer/common/executor"
	"github.com/luskaner/ageLANServer/common/pidLock"
	"github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/ip"
	"github.com/luskaner/ageLANServer/server/internal/models/initializer"
	"github.com/luskaner/ageLANServer/server/internal/routes/router"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configPaths = []string{path.Join("resources", "config"), "."}
var id string

var (
	Version string
	cfgFile string
	rootCmd = &cobra.Command{
		Use:   filepath.Base(os.Args[0]),
		Short: "server is a service for LAN features in AoE: DE, AoE 2: DE, AoE 3: DE and AoM: RT.",
		Run: func(_ *cobra.Command, _ []string) {
			lock := &pidLock.Lock{}
			if err := lock.Lock(); err != nil {
				fmt.Println("Failed to lock pid file. Kill process 'server' if it is running in your task manager.")
				fmt.Println(err.Error())
				os.Exit(common.ErrPidLock)
			}
			var err error
			if internal.Id, err = uuid.Parse(id); err != nil {
				fmt.Println("Invalid server instance ID")
				_ = lock.Unlock()
				os.Exit(internal.ErrInvalidId)
			}
			fmt.Println("Server instance ID:", internal.Id)
			if viper.GetBool("GeneratePlatformUserId") {
				fmt.Println("Generating platform User ID, this should only be used as a last resort and the custom launcher should be properly configured instead.")
			}
			gameSet := mapset.NewThreadUnsafeSet[string](viper.GetStringSlice("Games.Enabled")...)
			if gameSet.IsEmpty() {
				fmt.Println("No games specified")
				_ = lock.Unlock()
				os.Exit(internal.ErrGames)
			}
			for game := range gameSet.Iter() {
				if !common.SupportedGames.ContainsOne(game) {
					fmt.Println("Invalid game specified:", game)
					_ = lock.Unlock()
					os.Exit(internal.ErrGames)
				}
			}
			if executor.IsAdmin() {
				fmt.Println("Running as administrator, this is not recommended for security reasons.")
				if runtime.GOOS == "linux" {
					fmt.Println(fmt.Sprintf("If the issue is that you cannot listen on the port, then run `sudo setcap CAP_NET_BIND_SERVICE=+eip '%s'`, before re-running the 'server'", os.Args[0]))
				}
			}
			certificatePairFolder := common.CertificatePairFolder(os.Args[0])
			if certificatePairFolder == "" {
				fmt.Println("Failed to determine certificate pair folder")
				_ = lock.Unlock()
				os.Exit(internal.ErrCertDirectory)
			}
			multicastGroups := mapset.NewThreadUnsafeSet[netip.Addr]()
			multicast := viper.GetBool("Announcement.Multicast")
			if multicast {
				multicastIP, err := netip.ParseAddr(viper.GetString("Announcement.MulticastGroup"))
				if err != nil || !multicastIP.Is4() || !multicastIP.IsMulticast() {
					fmt.Println("Invalid multicast IP")
					if err != nil {
						fmt.Println(err.Error())
					}
					_ = lock.Unlock()
					os.Exit(internal.ErrMulticastGroup)
				}
				multicastGroups.Add(multicastIP)
			}
			broadcast := viper.GetBool("Announcement.Broadcast")
			announcePort := viper.GetInt("Announcement.Port")
			internal.AnnounceMessageData = make(map[string]common.AnnounceMessageData002, gameSet.Cardinality())
			logToConsole := viper.GetBool("LogToConsole")
			customLogger := log.New(&internal.CustomWriter{OriginalWriter: os.Stderr}, "", log.LstdFlags)
			var servers []*http.Server
			stop := make(chan os.Signal, 1)
			signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
			for gameId := range gameSet.Iter() {
				fmt.Printf("Game %s:\n", gameId)
				hosts := viper.GetStringSlice(fmt.Sprintf("Games.%s.Hosts", gameId))
				addrs := ip.ResolveHosts(mapset.NewThreadUnsafeSet[string](hosts...))
				if addrs.IsEmpty() {
					fmt.Println("\tFailed to resolve host (or it was an IPv6 address)")
					_ = lock.Unlock()
					os.Exit(internal.ErrResolveHost)
				}
				initializer.InitializeGame(gameId)
				var writer io.Writer
				if logToConsole {
					writer = os.Stdout
				} else {
					err := os.MkdirAll(filepath.Join("logs", gameId), 0755)
					if err != nil {
						fmt.Println("\tFailed to create logs directory")
						_ = lock.Unlock()
						os.Exit(internal.ErrCreateLogsDir)
					}
					t := time.Now()
					fileName := fmt.Sprintf("logs/%s/access_log_%d-%02d-%02dT%02d-%02d-%02d.txt", gameId, t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
					file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0644)
					if err != nil {
						fmt.Println("\tFailed to create log file")
						_ = lock.Unlock()
						os.Exit(internal.ErrCreateLogFile)
					}
					writer = file
				}
				internal.AnnounceMessageData[gameId] = internal.AnnounceMessageDataLatest{
					GameTitle: gameId,
					Version:   Version,
				}
				general := &router.General{Writer: writer}
				mux := general.InitializeRoutes(gameId, router.HostMiddleware(gameId, writer))
				mux = router.TitleMiddleware(gameId, mux)
				for addr := range addrs.Iter() {
					var certFile string
					var keyFile string
					if common.SelfSignedCertGame(gameId) {
						certFile = filepath.Join(certificatePairFolder, common.SelfSignedCert)
						keyFile = filepath.Join(certificatePairFolder, common.SelfSignedKey)
					} else {
						certFile = filepath.Join(certificatePairFolder, common.Cert)
						keyFile = filepath.Join(certificatePairFolder, common.Key)
					}
					var listenConns []*net.UDPConn
					if broadcast || multicast {
						var err error
						err, listenConns = ip.QueryConnections(addr, multicastGroups, announcePort)
						if err != nil {
							fmt.Println("\tFailed to listen to UDP connections for address", addr.String())
							_ = lock.Unlock()
							os.Exit(internal.ErrAnnounce)
						}
					}
					server := &http.Server{
						Addr:         addr.String() + ":443",
						Handler:      mux,
						ErrorLog:     customLogger,
						IdleTimeout:  time.Second * 30,
						ReadTimeout:  time.Second * 5,
						WriteTimeout: time.Second * 30,
					}

					fmt.Println("\tListening on " + server.Addr)
					go func() {
						if len(listenConns) > 0 {
							for _, conn := range listenConns {
								fmt.Printf(
									"\tListening for query connections on %s\n",
									conn.LocalAddr(),
								)
							}
							ip.ListenQueryConnections(listenConns)
						}
						err := server.ListenAndServeTLS(certFile, keyFile)
						if err != nil && !errors.Is(err, http.ErrServerClosed) {
							fmt.Println("\tFailed to start 'server'")
							fmt.Printf("%s\n", err)
							os.Exit(internal.ErrStartServer)
						}
					}()
					servers = append(servers, server)
				}
			}

			<-stop

			fmt.Println("'Servers' are shutting down...")

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			for _, server := range servers {
				if err := server.Shutdown(ctx); err != nil {
					fmt.Printf("'Server' %s forced to shutdown: %v\n", server.Addr, err)
				}

				fmt.Println("'Server'", server.Addr, "stopped")
			}

			_ = lock.Unlock()
		},
	}
)

func Execute() error {
	cobra.OnInitialize(initConfig)
	rootCmd.Version = Version
	rootCmd.Flags().StringVar(&cfgFile, "config", "", fmt.Sprintf(`config file (default config.toml in %s directories)`, strings.Join(configPaths, ", ")))
	rootCmd.Flags().StringP("announce", "a", "true", "Announce 'server' in LAN. Disabling this will not allow launchers to discover it and will require specifying the host")
	rootCmd.Flags().IntP("announcePort", "p", common.AnnouncePort, "Port to announce to. If changed, the 'launcher's will need to specify the port in Server.AnnouncePorts")
	rootCmd.Flags().StringP("announceMulticast", "m", "true", "Whether to announce the 'server' using Multicast.")
	rootCmd.Flags().BoolP("announceBroadcast", "b", false, "Whether to announce the 'server' using Broadcast.")
	rootCmd.Flags().StringP("announceMulticastGroup", "i", common.AnnounceMulticastGroup, "Whether to announce the 'server' using Multicast or Broadcast.")
	cmd.GamesCommand(rootCmd.Flags())
	rootCmd.Flags().BoolP("logToConsole", "l", false, "Log the requests to the console (stdout) or not.")
	rootCmd.Flags().BoolP("generatePlatformUserId", "g", false, "Generate the Platform User Id based on the user's IP.")
	rootCmd.Flags().StringVar(&id, "id", uuid.NewString(), "Server instance ID to identify it.")

	if err := viper.BindPFlag("Announcement.Enabled", rootCmd.Flags().Lookup("announce")); err != nil {
		return err
	}
	if err := viper.BindPFlag("Announcement.Port", rootCmd.Flags().Lookup("announcePort")); err != nil {
		return err
	}
	if err := viper.BindPFlag("Announcement.Broadcast", rootCmd.Flags().Lookup("announceBroadcast")); err != nil {
		return err
	}
	if err := viper.BindPFlag("Announcement.Multicast", rootCmd.Flags().Lookup("announceMulticast")); err != nil {
		return err
	}
	if err := viper.BindPFlag("Announcement.MulticastGroup", rootCmd.Flags().Lookup("announceMulticastGroup")); err != nil {
		return err
	}
	if err := viper.BindPFlag("Games.Enabled", rootCmd.Flags().Lookup("games")); err != nil {
		return err
	}
	if err := viper.BindPFlag("LogToConsole", rootCmd.Flags().Lookup("logToConsole")); err != nil {
		return err
	}
	if err := viper.BindPFlag("GeneratePlatformUserId", rootCmd.Flags().Lookup("generatePlatformUserId")); err != nil {
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
		viper.SetConfigType("toml")
		viper.SetConfigName("config")
	}
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

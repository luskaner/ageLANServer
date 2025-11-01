package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/cmd"
	"github.com/luskaner/ageLANServer/common/executor"
	commonLogger "github.com/luskaner/ageLANServer/common/logger"
	"github.com/luskaner/ageLANServer/common/pidLock"
	"github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/ip"
	"github.com/luskaner/ageLANServer/server/internal/logger"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/models/initializer"
	"github.com/luskaner/ageLANServer/server/internal/routes/router"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configPaths = []string{path.Join("resources", "config"), "."}
var id string
var logRoot string
var flatLog bool

var (
	Version string
	cfgFile string
	rootCmd = &cobra.Command{
		Use:   filepath.Base(os.Args[0]),
		Short: "server is a service for multiplayer features in AoE: DE, AoE 2: DE, AoE 3: DE and AoM: RT.",
		Run: func(_ *cobra.Command, _ []string) {
			lock := &pidLock.Lock{}
			exitCode := common.ErrSuccess
			if err := lock.Lock(); err != nil {
				logger.Println("Failed to lock pid file. Kill process 'server' if it is running in your task manager.")
				logger.Println(err.Error())
				commonLogger.CloseFileLog()
				os.Exit(common.ErrPidLock)
			}
			commonLogger.Initialize(nil)
			if id == "" {
				id = uuid.NewString()
			}
			if logRoot == "" {
				logRoot = commonLogger.LogRootDate("")
			}
			if err := logger.OpenMainFileLog(logRoot); err != nil {
				logger.Printf("Failed to open main log file: %v", err)
				os.Exit(common.ErrFileLog)
			}
			var seed uint64
			if commonLogger.FileLogger == nil {
				seed = uint64(time.Now().UnixNano())
			}
			internal.InitializeRng(seed)
			var files []*os.File
			defer func() {
				if r := recover(); r != nil {
					logger.Println(r)
					logger.Println(string(debug.Stack()))
					exitCode = common.ErrGeneral
				}
				commonLogger.CloseFileLog()
				for _, file := range files {
					_ = file.Close()
				}
				_ = lock.Unlock()
				os.Exit(exitCode)
			}()
			var err error
			if internal.Id, err = uuid.Parse(id); err != nil {
				logger.Println("Invalid server instance ID")
				exitCode = internal.ErrInvalidId
				return
			}
			logger.Println("Server instance ID:", internal.Id)
			if viper.GetBool("GeneratePlatformUserId") {
				logger.Println("Generating platform User ID, this should only be used as a last resort and the custom launcher should be properly configured instead.")
			}
			gameSet := mapset.NewThreadUnsafeSet[string](viper.GetStringSlice("Games.Enabled")...)
			if gameSet.IsEmpty() {
				logger.Println("No games specified")
				exitCode = internal.ErrGames
				return
			}
			for game := range gameSet.Iter() {
				if !common.SupportedGames.ContainsOne(game) {
					logger.Println("Invalid game specified:", game)
					exitCode = internal.ErrGames
					return
				}
			}
			if executor.IsAdmin() {
				logger.Println("Running as administrator, this is not recommended for security reasons.")
				if runtime.GOOS == "linux" {
					logger.Println(fmt.Sprintf("If the issue is that you cannot listen on the port, then run `sudo setcap CAP_NET_BIND_SERVICE=+eip '%s'`, before re-running the 'server'", os.Args[0]))
				}
			}
			certificatePairFolder := common.CertificatePairFolder(os.Args[0])
			if certificatePairFolder == "" {
				logger.Println("Failed to determine certificate pair folder")
				exitCode = internal.ErrCertDirectory
				return
			}
			multicastGroups := mapset.NewThreadUnsafeSet[netip.Addr]()
			multicast := viper.GetBool("Announcement.Multicast")
			if multicast {
				multicastIP, err := netip.ParseAddr(viper.GetString("Announcement.MulticastGroup"))
				if err != nil || !multicastIP.Is4() || !multicastIP.IsMulticast() {
					logger.Println("Invalid multicast IP")
					if err != nil {
						logger.Println(err.Error())
					}
					exitCode = internal.ErrMulticastGroup
					return
				}
				multicastGroups.Add(multicastIP)
			}
			broadcast := viper.GetBool("Announcement.Broadcast")
			announcePort := viper.GetInt("Announcement.Port")
			internal.AnnounceMessageData = make(map[string]common.AnnounceMessageData002, gameSet.Cardinality())
			customLogger := log.New(&internal.CustomWriter{OriginalWriter: os.Stderr}, "", log.LstdFlags)
			var servers []*http.Server
			internal.InitializeStopSignal()
			for gameId := range gameSet.Iter() {
				logger.Printf("Game %s:\n", gameId)
				hosts := viper.GetStringSlice(fmt.Sprintf("Games.%s.Hosts", gameId))
				addrs := ip.ResolveHosts(mapset.NewThreadUnsafeSet[string](hosts...))
				if addrs.IsEmpty() {
					logger.Println("\tFailed to resolve host (or it was an IPv6 address)")
					exitCode = internal.ErrResolveHost
					return
				}
				if err = initializer.InitializeGame(gameId); err != nil {
					logger.Printf("\tFailed to initialize game: %v\n", err)
					exitCode = internal.ErrGame
					return
				}
				if battlesServers, ok := models.BattleServers[gameId]; ok && len(battlesServers) > 0 {
					logger.Println("\tBattle Servers:")
					for _, battleServer := range battlesServers {
						logger.Println("\t\t" + battleServer.String())
					}
				}
				var writer io.Writer
				var root *commonLogger.Root
				gameLogRoot := logRoot
				var filePrefix string
				if flatLog {
					filePrefix = fmt.Sprintf("%s_", gameId)
				} else {
					gameLogRoot = filepath.Join(gameLogRoot, gameId)
				}
				if commonLogger.FileLogger == nil {
					writer = os.Stdout
				} else if err, root = commonLogger.NewFile(gameLogRoot, "", true); err != nil {
					logger.Printf("\tFailed to prepare log folder: %v\n", err)
					exitCode = internal.ErrCreateLogFile
					return
				} else if f, err := root.Open(filePrefix + "access_log"); err != nil {
					logger.Printf("\tFailed to open access log file: %v\n", err)
					exitCode = internal.ErrCreateLogFile
					return
				} else {
					files = append(files, f)
					writer = f
				}
				internal.AnnounceMessageData[gameId] = internal.AnnounceMessageDataLatest{
					GameTitle: gameId,
					Version:   Version,
				}
				general := &router.General{Writer: writer}
				mux := general.InitializeRoutes(gameId, router.HostMiddleware(gameId, writer))
				mux = router.TitleMiddleware(gameId, mux)
				if root != nil {
					if f, err := root.Open(filePrefix + "communication_log"); err != nil {
						logger.Printf("\tFailed to open communication log file: %v\n", err)
						exitCode = internal.ErrCreateLogFile
					} else {
						files = append(files, f)
						opts := &slog.HandlerOptions{
							ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
								if len(groups) == 0 {
									switch a.Key {
									case slog.TimeKey, slog.LevelKey:
										return slog.Attr{}
									}
								}
								return a
							},
						}
						handler := slog.NewJSONHandler(f, opts)
						slog.SetDefault(slog.New(handler))
						mux = router.NewLoggingMiddleware(mux, time.Now().UTC())
					}
				}
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
						err, listenConns = ip.QueryConnections(addr, multicastGroups, announcePort)
						if err != nil {
							logger.Println("\tFailed to listen to UDP connections for address", addr.String())
							exitCode = internal.ErrAnnounce
							return
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

					logger.Println("\tListening on " + server.Addr)
					go func() {
						if len(listenConns) > 0 {
							for _, conn := range listenConns {
								logger.Printf(
									"\tListening for query connections on %s\n",
									conn.LocalAddr(),
								)
							}
							ip.ListenQueryConnections(listenConns)
						}
						err := server.ListenAndServeTLS(certFile, keyFile)
						if err != nil && !errors.Is(err, http.ErrServerClosed) {
							logger.Println("\tFailed to start 'server'")
							logger.Printf("%s\n", err)
							exitCode = internal.ErrStartServer
							return
						}
					}()
					servers = append(servers, server)
				}
			}

			<-internal.StopSignal

			logger.Println("'Servers' are shutting down...")

			var wg sync.WaitGroup
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			for _, server := range servers {
				wg.Add(1)
				go func(s *http.Server) {
					defer wg.Done()
					if err := s.Shutdown(ctx); err != nil {
						fmt.Printf("'Server' %s forced to shutdown: %v\n", s.Addr, err)
					}
					logger.Println("'Server'", s.Addr, "stopped")
				}(server)
			}
			wg.Wait()
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
	rootCmd.Flags().Bool("log", false, "Whether to log more info to a file. Enable it for errors.")
	rootCmd.Flags().BoolVar(&flatLog, "flatLog", false, "Whether to log in a flat structure in --logRoot. Only applicable if --log is passed.")
	cmd.GamesCommand(rootCmd.Flags())
	cmd.LogRootCommand(rootCmd.Flags(), &logRoot)
	rootCmd.Flags().BoolP("generatePlatformUserId", "g", false, "Generate the Platform User Id based on the user's IP.")
	rootCmd.Flags().StringVar(&id, "id", "", "Server instance ID to identify it.")
	if err := viper.BindPFlag("Config.Log", rootCmd.Flags().Lookup("log")); err != nil {
		return err
	}
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
		logger.Println("Using config file:", viper.ConfigFileUsed())
	}
}

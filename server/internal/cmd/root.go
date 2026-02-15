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
	"github.com/luskaner/ageLANServer/common/fileLock"
	commonLogger "github.com/luskaner/ageLANServer/common/logger"
	"github.com/luskaner/ageLANServer/common/paths"
	"github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/ip"
	"github.com/luskaner/ageLANServer/server/internal/logger"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/models/initializer"
	"github.com/luskaner/ageLANServer/server/internal/routes/router"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var v = viper.New()

var configPaths = []string{paths.ConfigsPath, "."}
var id string
var logRoot string
var flatLog bool
var deterministic bool
var cfgFile string

var (
	Version string
	rootCmd = &cobra.Command{
		Use:   filepath.Base(os.Args[0]),
		Short: "server is a service for multiplayer features in AoE: DE, AoE 2: DE, AoE 3: DE, AoE 4: AE and AoM: RT.",
		Run: func(_ *cobra.Command, _ []string) {
			lock := &fileLock.PidLock{}
			exitCode := common.ErrSuccess
			if err := lock.Lock(); err != nil {
				logger.Println("Failed to lock pid file. Kill process 'server' if it is running in your task manager.")
				logger.Println(err.Error())
				commonLogger.CloseFileLog()
				os.Exit(common.ErrPidLock)
			}
			cfg := initConfig()
			commonLogger.Initialize(nil)
			if logRoot == "" {
				logRoot = commonLogger.LogRootDate("")
			}
			if err := logger.OpenMainFileLog(logRoot, cfg.Log); err != nil {
				logger.Printf("Failed to open main log file: %v", err)
				os.Exit(common.ErrFileLog)
			}
			logger.PrintFile("config", v.ConfigFileUsed())
			var seed uint64
			if !deterministic {
				seed = uint64(time.Now().UnixNano())
			}
			internal.InitializeRng(seed)
			if id == "" {
				id = uuid.NewString()
			}
			var closables []io.Closer
			defer func() {
				if r := recover(); r != nil {
					logger.Println(r)
					logger.Println(string(debug.Stack()))
					exitCode = common.ErrGeneral
				}
				commonLogger.CloseFileLog()
				for _, file := range closables {
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
			if cfg.GeneratePlatformUserId {
				logger.Println("Generating platform User ID, this should only be used as a last resort and the custom launcher should be properly configured instead.")
			}
			gameSet := mapset.NewThreadUnsafeSet[string](cfg.Games.Enabled...)
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
			announceEnabled := cfg.Announcement.Enabled
			multicastGroups := mapset.NewThreadUnsafeSet[netip.Addr]()
			if announceEnabled && cfg.Announcement.Multicast {
				multicastIP, err := netip.ParseAddr(cfg.Announcement.MulticastGroup)
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
			announcePort := cfg.Announcement.Port
			internal.AnnounceMessageData = make(map[string]common.AnnounceMessageData002, gameSet.Cardinality())
			internal.GeneratePlatformUserId = cfg.GeneratePlatformUserId
			var servers []*http.Server
			internal.InitializeStopSignal()
			for gameId := range gameSet.Iter() {
				logger.Printf("Game %s:\n", gameId)
				hosts := cfg.GetGameHosts(gameId)
				addrs := ip.ResolveHosts(mapset.NewThreadUnsafeSet[string](hosts...))
				if addrs.IsEmpty() {
					logger.Println("\tFailed to resolve host (or it was an IPv6 address)")
					exitCode = internal.ErrResolveHost
					return
				}
				if err = initializer.InitializeGame(gameId, cfg.GetGameBattleServers(gameId)); err != nil {
					logger.Printf("\tFailed to initialize game: %v\n", err)
					exitCode = internal.ErrGame
					return
				}
				if battlesServers, ok := models.BattleServersStore[gameId]; ok && len(battlesServers) > 0 {
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
				customLoggerWriters := []io.Writer{os.Stderr}
				if commonLogger.FileLogger == nil {
					writer = os.Stdout
				} else {
					customLoggerWriters = append(customLoggerWriters, &commonLogger.Buf)
					if err, root = commonLogger.NewFile(gameLogRoot, "", true); err != nil {
						logger.Printf("\tFailed to prepare log folder: %v\n", err)
						exitCode = internal.ErrCreateLogFile
						return
					} else if f, err := root.Open(filePrefix + "access_log"); err != nil {
						logger.Printf("\tFailed to open access log file: %v\n", err)
						exitCode = internal.ErrCreateLogFile
						return
					} else {
						closables = append(closables, f)
						writer = f
					}
				}
				customLogger := log.New(
					&internal.CustomWriter{OriginalWriter: io.MultiWriter(customLoggerWriters...)},
					"|SERVER| ",
					log.Ltime|log.Lmicroseconds,
				)
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
						closables = append(closables, logger.NewBuffer(f))
						mux = router.NewLoggingMiddleware(mux)
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
					if announceEnabled {
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
				wg.Go(func() {
					if err := server.Shutdown(ctx); err != nil {
						fmt.Printf("'Server' %s forced to shutdown: %v\n", server.Addr, err)
					}
					logger.Println("'Server'", server.Addr, "stopped")
				})
			}
			wg.Wait()
		},
	}
)

func Execute() error {
	rootCmd.Version = Version
	rootCmd.Flags().StringVar(&cfgFile, "config", "", fmt.Sprintf(`config file (default config.toml in %s directories)`, strings.Join(configPaths, ", ")))
	rootCmd.Flags().StringP("announce", "a", "true", "Respond to discove 'server' in LAN. Disabling this will not allow launchers to discover it and will require specifying the host")
	rootCmd.Flags().IntP("announcePort", "p", common.AnnouncePort, "Port to respond to discovery requests. If changed, the 'launcher's will need to specify the port in Server.AnnouncePorts")
	rootCmd.Flags().StringP("announceMulticast", "m", "true", "Whether to respond to discovery queries using Multicast.")
	rootCmd.Flags().StringP("announceMulticastGroup", "i", common.AnnounceMulticastGroup, "Multicast address to respond to discovery queries if 'announce' is enabled.")
	rootCmd.Flags().Bool("log", false, "Whether to log more info to a file. Enable it for errors.")
	rootCmd.Flags().BoolVar(&flatLog, "flatLog", false, "Whether to log in a flat structure in --logRoot. Only applicable if --log is passed.")
	rootCmd.Flags().BoolVar(&deterministic, "deterministic", false, "Whether to be as deterministic as possible.")
	cmd.GamesCommand(rootCmd.Flags())
	cmd.LogRootCommand(rootCmd.Flags(), &logRoot)
	rootCmd.Flags().BoolP("generatePlatformUserId", "g", false, "Generate the Platform User Id based on the user's IP.")
	rootCmd.Flags().StringVar(&id, "id", "", "Server instance ID to identify it.")
	// Default Values
	// General
	v.SetDefault("Log", false)
	v.SetDefault("GeneratePlatformUserId", false)
	// Announcement
	v.SetDefault("Announcement.Enabled", true)
	v.SetDefault("Announcement.Multicast", true)
	v.SetDefault("Announcement.MulticastGroup", common.AnnounceMulticastGroup)
	v.SetDefault("Announcement.Port", common.AnnouncePort)
	// Games
	v.SetDefault("Games.Enabled", []string{})
	for game := range common.SupportedGames.Iter() {
		v.SetDefault(fmt.Sprintf("Games.%s.Hosts", game), []string{netip.IPv4Unspecified().String()})
	}
	// Bindings
	if err := v.BindPFlag("Log", rootCmd.Flags().Lookup("log")); err != nil {
		return err
	}
	if err := v.BindPFlag("Announcement.Enabled", rootCmd.Flags().Lookup("announce")); err != nil {
		return err
	}
	if err := v.BindPFlag("Announcement.Port", rootCmd.Flags().Lookup("announcePort")); err != nil {
		return err
	}
	if err := v.BindPFlag("Announcement.Multicast", rootCmd.Flags().Lookup("announceMulticast")); err != nil {
		return err
	}
	if err := v.BindPFlag("Announcement.MulticastGroup", rootCmd.Flags().Lookup("announceMulticastGroup")); err != nil {
		return err
	}
	if err := v.BindPFlag("Games.Enabled", rootCmd.Flags().Lookup("games")); err != nil {
		return err
	}
	if err := v.BindPFlag("GeneratePlatformUserId", rootCmd.Flags().Lookup("generatePlatformUserId")); err != nil {
		return err
	}
	return rootCmd.Execute()
}

func initConfig() *internal.Configuration {
	for _, configPath := range configPaths {
		v.AddConfigPath(configPath)
	}
	v.SetConfigType("toml")
	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
	} else {
		v.SetConfigName("config")
	}
	v.AutomaticEnv()
	if err := v.ReadInConfig(); err == nil {
		logger.Println("Using config file:", v.ConfigFileUsed())
	} else {
		if _, ok := errors.AsType[viper.ConfigFileNotFoundError](err); !ok {
			logger.Println("Error parsing config file:", v.ConfigFileUsed()+":", err.Error())
			os.Exit(common.ErrConfigParse)
		}
	}
	var c *internal.Configuration
	if err := v.Unmarshal(&c); err != nil {
		logger.Printf("unable to decode configuration: %v\n", err)
		os.Exit(common.ErrConfigParse)
	}
	return c
}

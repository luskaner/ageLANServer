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
	"sync"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/knadh/koanf/parsers/toml/v2"
	"github.com/knadh/koanf/v2"
	"github.com/luskaner/ageLANServer/common/cmd/server"
	"github.com/luskaner/ageLANServer/common/executables"
	"github.com/luskaner/ageLANServer/common/game"
	"github.com/spf13/pflag"

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
)

var configPaths = []string{paths.ConfigsPath, "."}
var values *server.Values

var (
	Version              string
	authenticationValues = mapset.NewThreadUnsafeSet[string]("required", "cached", "adaptive", "disabled")
)

func Execute() (err error, exitCode int) {
	var singleFs *cmd.SingleFlagSet
	values, singleFs = server.SingleFlagSet(Version, configPaths, runRoot)
	return singleFs.Execute()
}

func runRoot(fs *pflag.FlagSet) (err error, exitCode int) {
	lock := &fileLock.PidLock{}
	if err = lock.Lock(); err != nil {
		logger.Println("Failed to lock pid file. Kill process 'server' if it is running in your task manager.")
		logger.Println(err.Error())
		commonLogger.CloseFileLog()
		exitCode = common.ErrPidLock
		return
	}
	cfg, usedFile := initConfig(fs)
	commonLogger.Initialize(nil)
	if values.LogRoot == "" {
		values.LogRoot = commonLogger.LogRootDate("")
	}
	if err = logger.OpenMainFileLog(values.LogRoot, cfg.Log); err != nil {
		logger.Printf("Failed to open main log file: %v", err)
		exitCode = common.ErrFileLog
		return
	}
	if usedFile != "" {
		logger.PrintFile("config", usedFile)
	}
	internal.Connectivity = common.DNSConnectivity()
	models.CacheNetworkInterfaces()
	if !internal.Connectivity {
		logger.Println("No internet connectivity, some features will fallback gracefully.")
	}
	if !authenticationValues.ContainsOne(cfg.Authentication) {
		logger.Printf("Invalid authentication value: %s", cfg.Authentication)
		exitCode = internal.ErrInvalidAuthentication
		return
	} else if cfg.Authentication == "required" && !internal.Connectivity {
		logger.Println("Authentication is set to 'required' but there is no internet connectivity, which is required for authentication. Change the authentication method or fix the connectivity.")
		exitCode = internal.ErrInvalidAuthentication
		return
	}
	if cfg.Authentication == "adaptive" {
		if internal.Connectivity {
			cfg.Authentication = "cached"
		} else {
			cfg.Authentication = "disabled"
		}
		logger.Printf("Adaptive authentication resolved to '%s' based on connectivity\n", cfg.Authentication)
	}
	if cfg.Authentication == "disabled" {
		logger.Println("Authentication is disabled, you are responsible that users access it legally.")
	} else if cfg.GeneratePlatformUserId {
		logger.Println("Generating a platform User ID is not compatible with the Authentication resolving to a value other than 'disabled'.")
		exitCode = internal.ErrInvalidAuthentication
		return
	}
	internal.Authentication = cfg.Authentication
	var seed uint64
	if !values.Deterministic {
		seed = uint64(time.Now().UnixNano())
	}
	internal.InitializeRng(seed)
	if values.Id == "" {
		values.Id = uuid.NewString()
	}
	var closables []io.Closer
	defer func() {
		if r := recover(); r != nil {
			logger.Println(r)
			logger.Println(string(debug.Stack()))
			exitCode = common.ErrGeneral
		}
		commonLogger.CloseFileLog()
		for _, f := range closables {
			_ = f.Close()
		}
		_ = lock.Unlock()
	}()
	if internal.Id, err = uuid.Parse(values.Id); err != nil {
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
	for g := range gameSet.Iter() {
		if !game.SupportedGames.ContainsOne(g) {
			logger.Println("Invalid game specified:", g)
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
		var multicastIP netip.Addr
		multicastIP, err = netip.ParseAddr(cfg.Announcement.MulticastGroup)
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
		gameLogRoot := values.LogRoot
		var filePrefix string
		if values.Flatlog {
			filePrefix = fmt.Sprintf("%s_", gameId)
		} else {
			gameLogRoot = filepath.Join(gameLogRoot, gameId)
		}
		customLoggerWriters := []io.Writer{os.Stderr}
		if commonLogger.FileLogger == nil {
			writer = os.Stdout
		} else {
			customLoggerWriters = append(customLoggerWriters, &commonLogger.Buf)
			var f *os.File
			if err, root = commonLogger.NewFile(gameLogRoot, "", true); err != nil {
				logger.Printf("\tFailed to prepare log folder: %v\n", err)
				exitCode = internal.ErrCreateLogFile
				return
			} else if f, err = root.Open(filePrefix + "access_log"); err != nil {
				logger.Printf("\tFailed to open access log file: %v\n", err)
				exitCode = internal.ErrCreateLogFile
				return
			}
			closables = append(closables, f)
			writer = f
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
			var f *os.File
			if f, err = root.Open(filePrefix + "communication_log"); err != nil {
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
			s := &http.Server{
				Addr:         addr.String() + ":443",
				Handler:      mux,
				ErrorLog:     customLogger,
				IdleTimeout:  time.Second * 30,
				ReadTimeout:  time.Second * 5,
				WriteTimeout: time.Second * 30,
			}

			logger.Println("\tListening on " + s.Addr)
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
				err = s.ListenAndServeTLS(certFile, keyFile)
				if err != nil && !errors.Is(err, http.ErrServerClosed) {
					logger.Println("\tFailed to start 'server'")
					logger.Printf("%s\n", err)
					exitCode = internal.ErrStartServer
					return
				}
			}()
			servers = append(servers, s)
		}
	}

	<-internal.StopSignal

	logger.Println("'Servers' are shutting down...")

	var wg sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	for _, s := range servers {
		wg.Go(func() {
			if err = s.Shutdown(ctx); err != nil {
				fmt.Printf("'Server' %s forced to shutdown: %v\n", s.Addr, err)
			}
			logger.Println("'Server'", s.Addr, "stopped")
		})
	}
	wg.Wait()
	return
}

func initConfig(fs *pflag.FlagSet) (*internal.Configuration, string) {
	k := koanf.New(".")
	defaults := map[string]any{
		"Log":                         false,
		"GeneratePlatformUserId":      false,
		"Authentication":              "disabled",
		"Announcement.Enabled":        true,
		"Announcement.Multicast":      true,
		"Announcement.MulticastGroup": common.AnnounceMulticastGroup,
		"Announcement.Port":           common.AnnouncePort,
		"Games.Enabled":               []string{},
	}
	for g := range game.SupportedGames.Iter() {
		defaults[fmt.Sprintf("Games.%s.Hosts", g)] = []string{netip.IPv4Unspecified().String()}
	}
	bindings := map[string]string{
		"log":                    "Log",
		"generatePlatformUserId": "GeneratePlatformUserId",
		"authentication":         "Authentication",
		"announce":               "Announcement.Enabled",
		"announceMulticast":      "Announcement.Multicast",
		"announceMulticastGroup": "Announcement.MulticastGroup",
		"announcePort":           "Announcement.Port",
		cmd.GamesIdentifier:      "Games.Enabled",
	}
	var fileCandidates []string
	if values.CfgFile != "" {
		fileCandidates = append(fileCandidates, values.CfgFile)
	} else {
		for _, configPath := range configPaths {
			fileCandidates = append(fileCandidates, filepath.Join(configPath, "config.toml"))
		}
	}

	usedFile := common.LoadKoanfLayersOrExit(k, defaults, fileCandidates, toml.Parser(), fs, bindings, executables.Server, commonLogger.Println)
	if values.CfgFile != "" && usedFile == "" {
		logger.Println("No config file found, using defaults.")
	}
	if usedFile != "" {
		logger.Println("Using config file:", usedFile)
	}

	var c internal.Configuration
	if err := k.Unmarshal("", &c); err != nil {
		logger.Printf("unable to decode configuration: %v\n", err)
		os.Exit(common.ErrConfigParse)
	}
	return &c, usedFile
}

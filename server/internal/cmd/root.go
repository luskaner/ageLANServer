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
	"github.com/knadh/koanf/parsers/toml/v2"
	"github.com/knadh/koanf/v2"
	"github.com/luskaner/ageLANServer/common/executables"
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
var id string
var logRoot string
var flatLog bool
var deterministic bool
var cfgFile string

var (
	Version              string
	authenticationValues = mapset.NewThreadUnsafeSet[string]("required", "cached", "adaptive", "disabled")
)

func Execute() error {
	singleFs := cmd.NewSingleFlagSet(runRoot, Version)
	fs := singleFs.Fs()
	fs.StringVar(&cfgFile, "config", "", fmt.Sprintf(`config file (default config.toml in %s directories)`, strings.Join(configPaths, ", ")))
	fs.StringP("announce", "a", "true", "Respond to discove 'server' in LAN. Disabling this will not allow launchers to discover it and will require specifying the host")
	fs.IntP("announcePort", "p", common.AnnouncePort, "Port to respond to discovery requests. If changed, the 'launcher's will need to specify the port in Server.AnnouncePorts")
	fs.StringP("announceMulticast", "m", "true", "Whether to respond to discovery queries using Multicast.")
	fs.StringP("announceMulticastGroup", "i", common.AnnounceMulticastGroup, "Multicast address to respond to discovery queries if 'announce' is enabled.")
	fs.Bool("log", false, "Whether to log more info to a file. Enable it for errors.")
	fs.BoolVar(&flatLog, "flatLog", false, "Whether to log in a flat structure in --logRoot. Only applicable if --log is passed.")
	fs.BoolVar(&deterministic, "deterministic", false, "Whether to be as deterministic as possible.")
	cmd.GamesCommand(fs)
	cmd.LogRootCommand(fs, &logRoot)
	fs.BoolP("generatePlatformUserId", "g", false, "Generate the Platform User Id based on the user's IP.")
	fs.StringVar(&id, "id", "", "Server instance ID to identify it.")
	return singleFs.Execute()
}

func runRoot(fs *pflag.FlagSet) error {
	lock := &fileLock.PidLock{}
	exitCode := common.ErrSuccess
	if err := lock.Lock(); err != nil {
		logger.Println("Failed to lock pid file. Kill process 'server' if it is running in your task manager.")
		logger.Println(err.Error())
		commonLogger.CloseFileLog()
		os.Exit(common.ErrPidLock)
	}
	cfg, usedFile := initConfig(fs)
	commonLogger.Initialize(nil)
	if logRoot == "" {
		logRoot = commonLogger.LogRootDate("")
	}
	if err := logger.OpenMainFileLog(logRoot, cfg.Log); err != nil {
		logger.Printf("Failed to open main log file: %v", err)
		os.Exit(common.ErrFileLog)
	}
	if usedFile != "" {
		logger.PrintFile("config", usedFile)
	}
	internal.Connectivity = common.DNSConnectivity()
	if !internal.Connectivity {
		logger.Println("No internet connectivity, some features will fallback gracefully.")
	}
	if !authenticationValues.ContainsOne(cfg.Authentication) {
		logger.Printf("Invalid authentication value: %s", cfg.Authentication)
		os.Exit(internal.ErrInvalidAuthentication)
	} else if cfg.Authentication == "required" && !internal.Connectivity {
		logger.Println("Authentication is set to 'required' but there is no internet connectivity, which is required for authentication. Change the authentication method or fix the connectivity.")
		os.Exit(internal.ErrInvalidAuthentication)
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
		os.Exit(internal.ErrInvalidAuthentication)
	}
	internal.Authentication = cfg.Authentication
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
		for _, f := range closables {
			_ = f.Close()
		}
		_ = lock.Unlock()
		os.Exit(exitCode)
	}()
	var err error
	if internal.Id, err = uuid.Parse(id); err != nil {
		logger.Println("Invalid server instance ID")
		exitCode = internal.ErrInvalidId
		return nil
	}
	logger.Println("Server instance ID:", internal.Id)
	if cfg.GeneratePlatformUserId {
		logger.Println("Generating platform User ID, this should only be used as a last resort and the custom launcher should be properly configured instead.")
	}
	gameSet := mapset.NewThreadUnsafeSet[string](cfg.Games.Enabled...)
	if gameSet.IsEmpty() {
		logger.Println("No games specified")
		exitCode = internal.ErrGames
		return nil
	}
	for game := range gameSet.Iter() {
		if !common.SupportedGames.ContainsOne(game) {
			logger.Println("Invalid game specified:", game)
			exitCode = internal.ErrGames
			return nil
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
		return nil
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
			return nil
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
			return nil
		}
		if err = initializer.InitializeGame(gameId, cfg.GetGameBattleServers(gameId)); err != nil {
			logger.Printf("\tFailed to initialize game: %v\n", err)
			exitCode = internal.ErrGame
			return nil
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
				return nil
			} else if f, err := root.Open(filePrefix + "access_log"); err != nil {
				logger.Printf("\tFailed to open access log file: %v\n", err)
				exitCode = internal.ErrCreateLogFile
				return nil
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
					return nil
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
	return nil
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
	for game := range common.SupportedGames.Iter() {
		defaults[fmt.Sprintf("Games.%s.Hosts", game)] = []string{netip.IPv4Unspecified().String()}
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
	if cfgFile != "" {
		fileCandidates = append(fileCandidates, cfgFile)
	} else {
		for _, configPath := range configPaths {
			fileCandidates = append(fileCandidates, filepath.Join(configPath, "config.toml"))
		}
	}

	usedFile, err := common.LoadKoanfLayers(k, defaults, fileCandidates, toml.Parser(), fs, bindings, executables.Server)
	if err != nil {
		if fileErr, ok := errors.AsType[*common.KoanfFileLoadError](err); ok {
			logger.Println("Error parsing config file:", fileErr.Path+":", fileErr.Err.Error())
		} else {
			logger.Println("Error loading config:", err.Error())
		}
		os.Exit(common.ErrConfigParse)
	}
	if cfgFile != "" && usedFile == "" {
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

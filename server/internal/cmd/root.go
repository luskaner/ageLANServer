package cmd

import (
	"context"
	"errors"
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/gorilla/handlers"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/cmd"
	commonCmd "github.com/luskaner/ageLANServer/common/cmd"
	cfg "github.com/luskaner/ageLANServer/common/config"
	"github.com/luskaner/ageLANServer/common/config/server"
	"github.com/luskaner/ageLANServer/common/config/shared"
	"github.com/luskaner/ageLANServer/common/executor"
	"github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/ip"
	"github.com/luskaner/ageLANServer/server/internal/middleware"
	"github.com/luskaner/ageLANServer/server/internal/models/initializer"
	"github.com/luskaner/ageLANServer/server/internal/routes"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

var configPaths = []string{path.Join("resources", "config"), "."}
var parameters = server.Config{}
var v = viper.New()

var (
	Version  string
	cfgFiles []string
	rootCmd  = &cobra.Command{
		Use:   filepath.Base(os.Args[0]),
		Short: "server is a service for LAN features in AoE: DE, AoE 2: DE and AoE 3: DE.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := server.Unmarshal(v, &parameters); err != nil {
				return err
			}
			err, validate := server.Validator()
			if err != nil {
				return err
			}
			if err := validate.Struct(&parameters); err != nil {
				return err
			}
			internal.Id, _ = uuid.Parse(parameters.Id)
			internal.GeneratePlatformUserId = parameters.GeneratePlatformUserId
			return nil
		},
		Run: func(cmd *cobra.Command, _ []string) {
			fmt.Printf("GameTitles: %v\n", parameters.GameTitles)
			if executor.IsAdmin() {
				fmt.Println("Running as administrator, this is not recommended for security reasons.")
				if runtime.GOOS == "linux" {
					fmt.Println(fmt.Sprintf("If the issue is that you cannot listen on the port, then run `sudo setcap CAP_NET_BIND_SERVICE=+eip '%s'`, before re-running the 'server'", os.Args[0]))
				}
			}
			var hostAddrs []net.IP
			if len(parameters.Listen.Hosts) > 0 {
				if hostAddrs = ip.ResolveHosts(parameters.Listen.Hosts); len(hostAddrs) == 0 {
					fmt.Println("Failed to resolve hosts.")
					os.Exit(internal.ErrResolveHost)
				}
			}
			var addrs []net.IP
			if addrs = shared.FilterNetworks(nil, hostAddrs, parameters.Listen.Interfaces, true); len(addrs) == 0 {
				fmt.Println("No addresses to bind to.")
				os.Exit(internal.ErrNoAddrs)
			}
			gameTitleSet := mapset.NewSet[common.GameTitle](parameters.GameTitles...)
			mux := http.NewServeMux()
			initializer.InitializeGames(gameTitleSet)
			routes.Initialize(mux, gameTitleSet)
			gameMux := middleware.GameMiddleware(gameTitleSet, mux)
			sessionMux := middleware.SessionMiddleware(gameMux)
			handler := sessionMux
			if parameters.Logging.Enabled {
				var writer io.Writer
				if parameters.Logging.Console {
					writer = os.Stdout
				} else {
					err := os.MkdirAll("logs", 0755)
					if err != nil {
						fmt.Println("Failed to create logs directory")
						os.Exit(internal.ErrCreateLogsDir)
					}
					t := time.Now()
					fileName := fmt.Sprintf("logs/access_log_%d-%02d-%02dT%02d-%02d-%02d.txt", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
					file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0644)
					if err != nil {
						fmt.Println("Failed to create log file")
						os.Exit(internal.ErrCreateLogFile)
					}
					writer = file
				}
				handler = handlers.LoggingHandler(writer, sessionMux)
			}
			certificatePairFolder := common.CertificatePairFolder(os.Args[0])
			if certificatePairFolder == "" {
				fmt.Println("Failed to determine certificate pair folder")
				os.Exit(internal.ErrCertDirectory)
			}
			stop := make(chan os.Signal, 1)
			signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
			certFile := filepath.Join(certificatePairFolder, common.Cert)
			keyFile := filepath.Join(certificatePairFolder, common.Key)
			var servers []*http.Server
			customLogger := log.New(&internal.CustomWriter{OriginalWriter: os.Stderr}, "", log.LstdFlags)
			var multicastIP net.IP
			var announcePort int
			gameTitles := make([]string, len(parameters.GameTitles))
			for i, gameTitle := range parameters.GameTitles {
				gameTitles[i] = string(gameTitle)
			}
			internal.AnnounceMessageData = internal.AnnounceMessageDataLatest{
				AnnounceMessageData001: common.AnnounceMessageData001{
					GameTitles: gameTitles,
				},
				Version: Version,
			}
			if !parameters.Announcement.Disabled {
				announcePort = int(parameters.Announcement.Port)
				multicastIP = parameters.Announcement.MulticastGroup
			}
			fmt.Println("ID:", parameters.Id)
			if parameters.GeneratePlatformUserId {
				fmt.Println("Generating platform User ID, this should only be used as a last resort and the custom launcher should be properly configured instead.")
			}
			for _, addr := range addrs {
				server := &http.Server{
					Addr:        addr.String() + ":443",
					Handler:     handler,
					ErrorLog:    customLogger,
					IdleTimeout: time.Second * 20,
				}

				fmt.Println("Listening on " + server.Addr)
				go func() {
					if announcePort != 0 {
						if ip.Announce(
							net.ParseIP(addr.String()),
							multicastIP,
							announcePort,
						) {
							fmt.Printf("Responding to discovery requests on port %d (including using Multicast on group %s)\n", announcePort, multicastIP)
						} else {
							fmt.Println("Failed to respond to discovery requests on port ", announcePort)
						}
					}
					err := server.ListenAndServeTLS(certFile, keyFile)
					if err != nil && !errors.Is(err, http.ErrServerClosed) {
						fmt.Println("Failed to start 'server'")
						fmt.Println(err)
						os.Exit(internal.ErrStartServer)
					}
				}()
				servers = append(servers, server)
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
		},
	}
)

func Execute() error {
	server.SetDefaults(v)
	cobra.OnInitialize(initConfig)
	rootCmd.Version = Version
	rootCmd.PersistentFlags().StringSliceVar(&cfgFiles, "config", []string{}, fmt.Sprintf(`config file (default config.toml in %s directories)`, strings.Join(configPaths, ", ")))
	rootCmd.PersistentFlags().BoolP("noAnnounce", "a", v.GetBool("Announcement.Disabled"), "Disable responding to discovery queries in LAN. This will make 'launcher's unable to discover it.")
	rootCmd.PersistentFlags().UintP("announcePort", "p", v.GetUint("Announcement.Port"), "Port to respond discovery requests to. If changed, the 'launcher's will need to specify the port.")
	rootCmd.PersistentFlags().IPP("announceMulticastGroup", "i", v.Get("Announcement.MulticastGroup").(net.IP), "Multicast group to respond discovery requests to.")
	cmd.GamesCommand(rootCmd.PersistentFlags())
	rootCmd.PersistentFlags().StringArrayP("host", "n", v.GetStringSlice("Listen.Hosts"), "The host the 'server' will bind to. Can be set multiple times.")
	rootCmd.PersistentFlags().BoolP("log", "o", v.GetBool("Logging.Enabled"), "Log requests.")
	rootCmd.PersistentFlags().BoolP("logToConsole", "l", v.GetBool("Logging.Console"), "Log the requests to the terminal (stdout) instead of a file. Depends on 'Logging.Enabled' being 'true'.")
	rootCmd.PersistentFlags().BoolP("generatePlatformUserId", "g", v.GetBool("GeneratePlatformUserId"), "Generate the Platform User Id to avoid issues with users sharing the same.")
	rootCmd.PersistentFlags().StringP("id", "d", v.GetString("Id"), "ID to identify this server. Sent in '/test' and announce data.")
	if err := v.BindPFlag("Id", rootCmd.PersistentFlags().Lookup("id")); err != nil {
		return err
	}
	if err := v.BindPFlag("Announcement.Disabled", rootCmd.PersistentFlags().Lookup("noAnnounce")); err != nil {
		return err
	}
	if err := v.BindPFlag("Announcement.Port", rootCmd.PersistentFlags().Lookup("announcePort")); err != nil {
		return err
	}
	if err := v.BindPFlag("Announcement.MulticastGroup", rootCmd.PersistentFlags().Lookup("announceMulticastGroup")); err != nil {
		return err
	}
	if err := v.BindPFlag("Listen.Hosts", rootCmd.PersistentFlags().Lookup("host")); err != nil {
		return err
	}
	if err := v.BindPFlag("GameTitles", rootCmd.PersistentFlags().Lookup(commonCmd.Names)); err != nil {
		return err
	}
	if err := v.BindPFlag("Logging.Console", rootCmd.PersistentFlags().Lookup("logToConsole")); err != nil {
		return err
	}
	if err := v.BindPFlag("Logging.Enabled", rootCmd.PersistentFlags().Lookup("log")); err != nil {
		return err
	}
	if err := v.BindPFlag("GeneratePlatformUserId", rootCmd.PersistentFlags().Lookup("generatePlatformUserId")); err != nil {
		return err
	}
	return rootCmd.Execute()
}

func initConfig() {
	cfg.InitConfig(
		v,
		configPaths,
		cfgFiles,
		common.Server,
		func(path string) {
			fmt.Println("Using config file:", path)
		},
	)
}

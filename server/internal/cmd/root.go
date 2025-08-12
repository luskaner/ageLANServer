package cmd

import (
	"context"
	"errors"
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/gorilla/handlers"
	"github.com/luskaner/ageLANServer/common"
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
	"net/netip"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
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
			if parameters.Network.Listen.Hosts.Values.IsEmpty() {
				if parameters.Network.IPProtocol != common.IPvDual && parameters.Network.IPProtocol.IPv4() {
					parameters.Network.Listen.Hosts.Values.Add(netip.IPv4Unspecified().String())
				}
				if parameters.Network.IPProtocol.IPv6() {
					parameters.Network.Listen.Hosts.Values.Add(netip.IPv6Unspecified().String())
				}
			}
			if parameters.Network.Listen.Port == 0 {
				if parameters.Network.Listen.DisableHttps {
					parameters.Network.Listen.Port = 80
				} else {
					parameters.Network.Listen.Port = 443
				}
			}
			err, validate := server.Validator()
			if err != nil {
				return err
			}
			if err := validate.Struct(&parameters); err != nil {
				return err
			}
			internal.Id = parameters.Id
			internal.GeneratePlatformUserId = parameters.GeneratePlatformUserId
			return nil
		},
		Run: func(cmd *cobra.Command, _ []string) {
			fmt.Printf("GameTitles: %v\n", parameters.GameTitles)
			if parameters.Network.IPProtocol.IPv6() {
				fmt.Println("There might be issues with IPv6.")
			}
			if !parameters.Network.Listen.Hosts.Values.IsEmpty() && !parameters.Network.Listen.Interfaces.IsEmpty() {
				fmt.Println("Setting both 'Network.Listen.Hosts.Values' and 'Network.Listen.Interfaces' is not supported, only the 'Hosts.Values' will be used for binding.")
			}
			if parameters.Network.Listen.DisableHttps {
				fmt.Println("Will only listen on HTTP, you will require an HTTPS frontend with some kind of redirection to make it work.")
			} else if parameters.Network.Listen.Port != 443 {
				fmt.Println("Not listening on the default HTTPS port, you will require some kind of port redirection to make it work.")
			}
			if executor.IsAdmin() {
				fmt.Println("Running as administrator, this is not recommended for security reasons.")
				if runtime.GOOS == "linux" {
					fmt.Println(fmt.Sprintf("If the issue is that you cannot listen on the port, then run `sudo setcap CAP_NET_BIND_SERVICE=+eip '%s'`, before re-running the 'server'", os.Args[0]))
				}
			}
			var ipAddrs mapset.Set[netip.Addr]
			if !parameters.Network.Listen.Hosts.Values.IsEmpty() {
				if ipAddrs = ip.ResolveHosts(
					parameters.Network.Listen.Hosts.Values,
					parameters.Network.Listen.Hosts.UseOnlyFirstResolvedIP,
					parameters.Network.IPProtocol.IPv4(),
					parameters.Network.IPProtocol.IPv6(),
				); ipAddrs.IsEmpty() {
					fmt.Println("Failed to resolve hosts.")
					os.Exit(internal.ErrResolveHost)
				}
			} else if !parameters.Network.Listen.Interfaces.IsEmpty() {
				if ipAddrs = shared.FilterNetworks(
					nil,
					parameters.Network.Listen.Interfaces,
					parameters.Network.IPProtocol.IPv4(),
					parameters.Network.IPProtocol.IPv6(),
					false,
				); ipAddrs.IsEmpty() {
					fmt.Println("No addresses to bind to.")
					os.Exit(internal.ErrNoAddrs)
				}
			}
			mux := http.NewServeMux()
			initializer.InitializeGames(parameters.GameTitles)
			routes.Initialize(mux, parameters.GameTitles)
			gameMux := middleware.GameMiddleware(parameters.GameTitles, mux)
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
			var serve func(ln *net.Listener, s *http.Server) error
			if !parameters.Network.Listen.DisableHttps {
				certificatePairFolder := common.CertificatePairFolder(os.Args[0])
				if certificatePairFolder == "" {
					fmt.Println("Failed to determine certificate pair folder")
					os.Exit(internal.ErrCertDirectory)
				}
				certFile := filepath.Join(certificatePairFolder, common.Cert)
				keyFile := filepath.Join(certificatePairFolder, common.Key)
				serve = func(ln *net.Listener, s *http.Server) error {
					return s.ServeTLS(*ln, certFile, keyFile)
				}
			} else {
				serve = func(ln *net.Listener, s *http.Server) error {
					return s.Serve(*ln)
				}
			}
			stop := make(chan os.Signal, 1)
			signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
			var servers []*http.Server
			customLogger := log.New(&internal.CustomWriter{OriginalWriter: os.Stderr}, "", log.LstdFlags)
			var gameTitles []string
			for gameTitle := range parameters.GameTitles.Iter() {
				gameTitles = append(gameTitles, string(gameTitle))
			}
			internal.AnnounceMessageData = internal.AnnounceMessageDataLatest{
				AnnounceMessageData001: common.AnnounceMessageData001{
					GameTitles: gameTitles,
				},
				Version: Version,
			}
			fmt.Println("ID:", parameters.Id)
			if parameters.GeneratePlatformUserId {
				fmt.Println("Generating platform User ID, this should only be used as a last resort and the custom launcher should be properly configured instead.")
			}
			for ipAddr := range ipAddrs.Iter() {
				server := &http.Server{
					Addr:        net.JoinHostPort(ipAddr.String(), strconv.Itoa(int(parameters.Network.Listen.Port))),
					Handler:     handler,
					ErrorLog:    customLogger,
					IdleTimeout: time.Second * 20,
				}
				network := "tcp"
				var isIPv4, isIPv6 bool
				if ipAddr.Is4() {
					isIPv4 = true
					if parameters.Network.IPProtocol != common.IPvDual {
						network += "4"
					}
				} else {
					isIPv6 = true
					if parameters.Network.IPProtocol != common.IPvDual {
						network += "6"
					}
				}
				fmt.Println("Listening on " + server.Addr)
				ln, err := net.Listen(network, server.Addr)
				if err != nil {
					fmt.Println("Failed to listen.")
					fmt.Println(err)
					os.Exit(internal.ErrStartServer)
				}
				var queryConnections []*net.UDPConn
				if !parameters.Network.Announcement.Disabled {
					if isIPv4 {
						if err, queryConnections = ip.QueryConnections(
							ipAddr,
							mapset.NewThreadUnsafeSet[netip.Addr](parameters.Network.Announcement.IPv4.MulticastGroup),
							int(parameters.Network.Announcement.IPv4.Port),
							true,
							parameters.Network.IPProtocol == common.IPvDual,
						); err != nil {
							fmt.Println("Failed to get query connections.")
							fmt.Println(err)
							os.Exit(internal.ErrQueryServer)
						}
					} else if isIPv6 {
						multicastGroups := mapset.NewThreadUnsafeSet[netip.Addr](
							parameters.Network.Announcement.IPv6.MulticastGroup,
						)
						if !parameters.Network.Announcement.IPv6.DisableLinkLocal {
							multicastGroups.Add(netip.IPv6LinkLocalAllNodes())
						}
						if err, queryConnections = ip.QueryConnections(
							ipAddr,
							multicastGroups,
							int(parameters.Network.Announcement.IPv6.Port),
							false,
							parameters.Network.IPProtocol == common.IPvDual,
						); err != nil {
							fmt.Println("Failed to get query connections.")
							fmt.Println(err)
							os.Exit(internal.ErrQueryServer)
						}
					}
				}
				if len(queryConnections) > 0 {
					for _, conn := range queryConnections {
						fmt.Printf(
							"Listening for query connections on %s\n",
							conn.LocalAddr(),
						)
					}
					ip.ListenQueryConnections(queryConnections)
				}
				go func() {
					if err = serve(&ln, server); err != nil && !errors.Is(err, http.ErrServerClosed) {
						fmt.Println("Failed to serve.")
						fmt.Println(err)
						os.Exit(internal.ErrStartServer)
					}
				}()
				servers = append(servers, server)
			}

			<-stop

			fmt.Println("Servers are shutting down...")

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			for _, server := range servers {
				if err := server.Shutdown(ctx); err != nil {
					fmt.Printf("Server %s forced to shutdown: %v\n", server.Addr, err)
				}

				fmt.Println("Server", server.Addr, "stopped")
			}
		},
	}
)

func Execute() error {
	server.SetDefaults(v)
	cobra.OnInitialize(initConfig)
	rootCmd.Version = Version
	rootCmd.PersistentFlags().StringSliceVar(&cfgFiles, "config", []string{}, fmt.Sprintf(`config file (default config.toml in %s directories)`, strings.Join(configPaths, ", ")))
	rootCmd.PersistentFlags().BoolP("noAnnounce", "a", v.GetBool("Network.Announcement.Disabled"), "Disable responding to discovery queries in LAN. This will make 'launcher's unable to discover it.")
	rootCmd.PersistentFlags().String("ipProtocol", v.GetString("Network.IPProtocol"), "IP versions of the server. 'v4' for IPv4, 'v6' for IPv6, 'v4+v6' for separate IPv4 and IPv6 support and '' (empty) for dual stack support.")
	var gameTitles GameTitleValues
	GamesCommand(rootCmd.PersistentFlags(), &gameTitles)
	rootCmd.PersistentFlags().Bool("noHttps", v.GetBool("Network.Listen.DisableHttps"), "Use HTTP instead of HTTPS.")
	rootCmd.PersistentFlags().String("port", v.GetString("Network.Listen.Port"), "The port the 'server' will listen to.")
	rootCmd.PersistentFlags().StringArrayP("host", "n", []string{}, "The host the 'server' will bind to. Can be set multiple times.")
	rootCmd.PersistentFlags().BoolP("log", "o", v.GetBool("Logging.Enabled"), "Log requests.")
	rootCmd.PersistentFlags().BoolP("logToConsole", "l", v.GetBool("Logging.Console"), "Log the requests to the terminal (stdout) instead of a file. Depends on 'Logging.Enabled' being 'true'.")
	rootCmd.PersistentFlags().BoolP("generatePlatformUserId", "g", v.GetBool("GeneratePlatformUserId"), "Generate the Platform User Id to avoid issues with users sharing the same.")
	rootCmd.PersistentFlags().StringP("id", "d", v.GetString("Id"), "ID to identify this server. Sent in '/test' and announce data.")
	if err := v.BindPFlag("Id", rootCmd.PersistentFlags().Lookup("id")); err != nil {
		return err
	}
	if err := v.BindPFlag("Network.IPProtocol", rootCmd.PersistentFlags().Lookup("ipProtocol")); err != nil {
		return err
	}
	if err := v.BindPFlag("Network.Announcement.Disabled", rootCmd.PersistentFlags().Lookup("noAnnounce")); err != nil {
		return err
	}
	if err := v.BindPFlag("Network.Listen.Port", rootCmd.PersistentFlags().Lookup("port")); err != nil {
		return err
	}
	if err := v.BindPFlag("Network.Listen.DisableHttps", rootCmd.PersistentFlags().Lookup("noHttps")); err != nil {
		return err
	}
	if err := v.BindPFlag("Network.Listen.Hosts.Values", rootCmd.PersistentFlags().Lookup("host")); err != nil {
		return err
	}
	if err := v.BindPFlag("GameTitles", rootCmd.PersistentFlags().Lookup(Names)); err != nil {
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

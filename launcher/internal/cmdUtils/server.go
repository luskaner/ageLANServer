package cmdUtils

import (
	"context"
	"fmt"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/luskaner/ageLANServer/common"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	commonExecutor "github.com/luskaner/ageLANServer/launcher-common/executor/exec"
	"github.com/luskaner/ageLANServer/launcher/internal"
	"github.com/luskaner/ageLANServer/launcher/internal/cmdUtils/parse"
	"github.com/luskaner/ageLANServer/launcher/internal/cmdUtils/printer"
	"github.com/luskaner/ageLANServer/launcher/internal/server"
	"net"
	"os"
	"slices"
	"sort"
	"strings"
	"time"
)

type processedServer struct {
	server.MesuredIpAddress
	id          uuid.UUID
	description string
}

func (s *processedServer) Option() huh.Option[uuid.UUID] {
	return huh.Option[uuid.UUID]{
		Key:   s.description,
		Value: s.id,
	}
}

func processedServers(gameId string, servers map[uuid.UUID]*server.AnnounceMessage) []*processedServer {
	var processed []*processedServer
	for serverId, data := range servers {
		measuredIPs, internalData := server.FilterServerIPs(serverId, gameId, data.Ips)
		if internalData == nil {
			continue
		}
		bestAddress := measuredIPs[0]
		var bestHostsSlice []string
		bestHosts := launcherCommon.IpToHosts(bestAddress.Ip)
		var alternativeIpSlice []string
		var alternativeHostsSlice []string
		alternativeHosts := mapset.NewThreadUnsafeSet[string]()
		for _, alternativeAddress := range measuredIPs[1:] {
			alternativeHosts.Append(launcherCommon.IpToHosts(alternativeAddress.Ip).Difference(bestHosts).ToSlice()...)
			alternativeIpSlice = append(alternativeIpSlice, alternativeAddress.Ip)
		}
		sort.Strings(alternativeIpSlice)
		if !alternativeHosts.IsEmpty() {
			alternativeHostsSlice = alternativeHosts.ToSlice()
			sort.Strings(alternativeHostsSlice)
		}
		if !bestHosts.IsEmpty() {
			bestHostsSlice = bestHosts.ToSlice()
			sort.Strings(bestHostsSlice)
		}
		description := lipgloss.NewStyle().Bold(true).Render(bestAddress.Ip)
		if len(alternativeIpSlice) > 1 {
			description += ", "
			description += strings.Join(alternativeIpSlice, ", ")
		}
		if len(bestHostsSlice) > 0 || len(alternativeHostsSlice) > 0 {
			description += " ("
			for i, host := range bestHostsSlice {
				if i > 0 {
					description += ", "
				}
				description += lipgloss.NewStyle().Bold(true).Render(host)
			}
			if len(alternativeHostsSlice) > 0 {
				if len(bestHostsSlice) > 0 {
					description += ", "
				}
				description += strings.Join(alternativeHostsSlice, ", ")
			}
			description += ")"
		}
		description += printer.Gen("", "", printer.TS(" | ", printer.SeparatorStyle)) + printer.Gen(
			printer.Speed,
			"",
			printer.TS(
				fmt.Sprintf("%d ms", bestAddress.Latency.Truncate(time.Millisecond).Milliseconds()),
				printer.LiteralStyle,
			),
		)
		description += printer.Gen("", "", printer.TS(" | ", printer.SeparatorStyle)) + "📦 " + internalData.Version
		processed = append(processed, &processedServer{
			id:               serverId,
			MesuredIpAddress: bestAddress,
			description:      description,
		})
	}
	slices.SortStableFunc(processed, func(a, b *processedServer) int {
		return int(a.Latency - b.Latency)
	})
	return processed
}

func listenServerProgressUI(ctx context.Context, cancel context.CancelFunc) {
	defer cancel()
	_ = spinner.New().
		Style(lipgloss.NewStyle()).
		TitleStyle(lipgloss.NewStyle()).
		Title(
			printer.Gen(
				"",
				"",
				printer.T("Querying "),
				printer.TS("server", printer.ComponentStyle),
				printer.T("s, you might need to allow the "),
				printer.TS("launcher", printer.ComponentStyle),
				printer.T(" in the firewall..."),
			),
		).
		Context(ctx).
		Run()
}

func DiscoverServersAndSelectBestIp(broadcast bool, gameId string, startServerId uuid.UUID, multicastIPs []net.IP, ports []int) (errorCode int, id uuid.UUID, ip string) {
	id = startServerId
	servers := make(map[uuid.UUID]*server.AnnounceMessage)
	ctx, cancel := context.WithTimeout(context.Background(), server.AnnounceQuery)
	go listenServerProgressUI(ctx, cancel)
	server.QueryServers(ctx, multicastIPs, ports, broadcast, servers)
	if len(servers) > 0 {
		serverCtx, serverCancel := context.WithCancel(context.Background())
		var spinnerError error
		go func() {
			spinnerError = spinner.New().
				Style(lipgloss.NewStyle()).
				TitleStyle(lipgloss.NewStyle()).
				Title("Processing server results...").
				Context(serverCtx).
				Run()
		}()
		procServers := processedServers(gameId, servers)
		serverCancel()
		if spinnerError != nil {
			return
		}
		if len(procServers) == 1 {
			confirm := true
			if err := huh.NewConfirm().
				Description(procServers[0].description).
				Title("A single server was found, connect to it?").
				Affirmative("Connect").
				Negative("Host instead").
				Value(&confirm).
				WithTheme(huh.ThemeBase()).
				Run(); err == nil {
				if confirm {
					ip = procServers[0].Ip
					id = procServers[0].id
				}
			} else {
				errorCode = internal.ErrServerStart
			}
		} else {
			serverOptions := make([]huh.Option[uuid.UUID], len(procServers))
			serverIdToIp := make(map[uuid.UUID]string, len(procServers))
			for i, procServer := range procServers {
				serverOptions[i] = procServer.Option()
				serverIdToIp[procServer.id] = procServer.Ip
			}
			serverOptions = append(serverOptions, huh.Option[uuid.UUID]{
				Key:   "Host a server",
				Value: uuid.Nil,
			})
			id = serverOptions[0].Value
			selectable := huh.NewSelect[uuid.UUID]().
				Title("Select a server:").
				Options(
					serverOptions...,
				).
				Value(&id).
				WithTheme(huh.ThemeBase())
			if err := selectable.Run(); err == nil {
				if id != uuid.Nil {
					ip = serverIdToIp[id]
				}
			} else {
				errorCode = internal.ErrServerStart
			}
		}
	} else {
		confirm := true
		if err := huh.NewConfirm().
			Title("No server was found, host instead?").
			Affirmative("Host").
			Negative("Exit").
			Value(&confirm).
			WithTheme(huh.ThemeBase()).
			Run(); err == nil {
			if !confirm {
				errorCode = internal.ErrServerStartDeclined
			}
		} else {
			errorCode = internal.ErrServerStart
		}
	}
	return
}

func (c *Config) StartServer(executable string, args []string, stop bool, serverId uuid.UUID, canTrustCertificate bool) (errorCode int, ip string) {
	var serverExecutablePath string
	var err error
	if executable == "auto" {
		printer.Print(
			printer.Search,
			"",
			printer.T(`Looking for `),
			printer.TS("server", printer.ComponentStyle),
			printer.T(`... `),
		)
		serverExecutablePath = server.ResolveExecutablePath()
		if serverExecutablePath == "" {
			printer.PrintSimpln(
				printer.Error,
				`not found.`,
			)
			errorCode = internal.ErrServerExecutable
			return
		} else {
			printer.Println(
				printer.Success,
				printer.T(`found at: `),
				printer.TS(serverExecutablePath, printer.FilePathStyle),
			)
		}
	} else if serverExecutablePath, err = parse.Executable(executable, nil); err != nil {
		printer.Println(
			printer.Error,
			printer.T(`Could not parse the `),
			printer.TS("Server.Executable", printer.OptionStyle),
			printer.T(`.`),
		)
		errorCode = internal.ErrServerExecutable
		return
	} else if f, err := os.Stat(serverExecutablePath); err != nil || f.IsDir() {
		printer.Println(
			printer.Error,
			printer.T(`Could not find the `),
			printer.TS("Server.Executable", printer.OptionStyle),
			printer.T(`.`),
		)
		errorCode = internal.ErrServerExecutable
		return
	}

	if exists, certificateFolder, cert := common.CertificatePair(serverExecutablePath); !exists || server.CertificateSoonExpired(cert) {
		if !canTrustCertificate {
			printer.Println(
				printer.Error,
				printer.TS("Server.Start", printer.ComponentStyle),
				printer.T(" is "),
				printer.TS("true", printer.OptionStyle),
				printer.T(" but "),
				printer.TS("Config.CanTrustCertificate", printer.OptionStyle),
				printer.T(" is "),
				printer.TS("false", printer.OptionStyle),
				printer.T("."),
				printer.T(" Certificate pair is missing or soon to be expired."),
			)
			errorCode = internal.ErrServerCertMissingExpired
			return
		}
		if certificateFolder == "" {
			printer.Println(
				printer.Error,
				printer.T("Cannot find "),
				printer.TS("certificates", printer.FilePathStyle),
				printer.T(" folder of the "),
				printer.TS("server", printer.ComponentStyle),
				printer.T("."),
			)
			errorCode = internal.ErrServerCertDirectory
			return
		}
		if result := server.GenerateCertificatePair(certificateFolder); !result.Success() {
			printer.Println(
				printer.Error,
				printer.T("Failed to generate certificate pair for the "),
				printer.TS("server", printer.ComponentStyle),
				printer.T("."),
			)
			printer.PrintFailedResultError(result)
			errorCode = internal.ErrServerCertCreate
			return
		}
	}
	fmt.Print(
		printer.Gen(
			printer.Execute,
			"",
			printer.T("Starting "),
			printer.TS("server", printer.ComponentStyle),
			printer.T(", authorize it in firewall if needed... "),
		),
	)
	var stopStr string
	if stop {
		stopStr = "true"
	} else {
		stopStr = "false"
	}
	var result *commonExecutor.Result
	result, ip = server.StartServer(c.gameId, serverId, stopStr, serverExecutablePath, args)
	if result.Success() {
		printer.PrintSucceeded()
		if stop {
			c.SetServerPid(result.Pid)
		}
	} else {
		printer.PrintFailedResultError(result)
		errorCode = internal.ErrServerStart
	}
	return
}

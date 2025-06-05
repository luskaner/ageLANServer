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
	"sort"
	"strconv"
	"strings"
)

type processedServer struct {
	ip          string
	description string
}

func (s *processedServer) Option() huh.Option[string] {
	return huh.Option[string]{
		Value: s.description,
		Key:   s.ip,
	}
}

func SelectBestServerIp(ips []string) (ok bool, ip string) {
	var successIps []net.IP

	for _, curIp := range ips {
		if server.LanServer(curIp, true) {
			parsedIp := net.ParseIP(curIp)
			if parsedIp.IsLoopback() {
				return true, curIp
			}
			successIps = append(successIps, parsedIp.To4())
		}
	}

	countSuccessIps := len(successIps)
	if countSuccessIps == 0 {
		return
	}

	ok = true
	ip = successIps[0].String()
	interfaces, err := net.Interfaces()

	if err != nil {
		return
	}

	var addrs []net.Addr
	for _, i := range interfaces {
		addrs, err = i.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			v, addrOk := addr.(*net.IPNet)
			if !addrOk {
				continue
			}

			for _, curIp := range successIps {
				if v.Contains(curIp) {
					ip = curIp.String()
					return
				}
			}
		}
	}

	return
}

func processedServers(gameId string, servers map[uuid.UUID]*common.AnnounceMessage) []*processedServer {
	var processed []*processedServer
	for _, data := range servers {
		// Announce version 0 means not using the new domain.
		if data.Version < common.AnnounceVersion1 {
			continue
		}
		// Check the server is running with game support
		announceData := data.Data.(common.AnnounceMessageData001)
		gameIdSet := mapset.NewThreadUnsafeSet[string](announceData.GameIds...)
		if !gameIdSet.ContainsOne(gameId) {
			continue
		}
		// Check we can connect to the server
		ok, bestIp := SelectBestServerIp(data.Ips.ToSlice())
		if !ok {
			continue
		}
		// Announce version 1 could be using the new domain or the old, we need to check it
		serverCert := server.ReadCertificateFromServer(bestIp)
		if serverCert == nil || serverCert.Subject.CommonName != common.Name {
			continue
		}
		var bestHostsSlice []string
		bestHosts := launcherCommon.IpToHosts(bestIp)
		var alternativeIpSlice []string
		var alternativeHostsSlice []string
		alternativeHosts := mapset.NewThreadUnsafeSet[string]()
		for foundIp := range data.Ips.Iter() {
			if foundIp != bestIp {
				alternativeHosts.Append(launcherCommon.IpToHosts(foundIp).Difference(bestHosts).ToSlice()...)
				alternativeIpSlice = append(alternativeIpSlice, foundIp)
			}
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
		description := lipgloss.NewStyle().Bold(true).Render(bestIp)
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
		processed = append(processed, &processedServer{
			ip:          bestIp,
			description: description,
		})
	}
	return processed
}

func ListenToServerAnnouncementsAndSelectBestIp(ctx context.Context, ctxCancel context.CancelFunc, gameId string, multicastIPs []net.IP, ports []int) (errorCode int, ip string) {
	defer ctxCancel()
	errorCode = common.ErrSuccess
	servers := server.LanServersAnnounced(ctx, multicastIPs, ports)
	if servers == nil {
		ctxCancel()
		printer.Println(
			printer.Error,
			printer.T("Could not listen to "),
			printer.TS("server", printer.ComponentStyle),
			printer.T(" announcements. Maybe the UDP port "),
			printer.TS(strconv.Itoa(common.AnnouncePort), printer.LiteralStyle),
			printer.T(" is blocked or already in use."),
		)
		errorCode = internal.ErrListenServerAnnouncements
		return
	}
	ctxCancel()
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
			var confirm bool
			if err := huh.NewConfirm().
				Description(procServers[0].description).
				Title("A single server was found, connect to it?").
				Affirmative("Connect").
				Negative("Host instead").
				Value(&confirm).
				WithTheme(huh.ThemeBase()).
				Run(); err == nil {
				if confirm {
					ip = procServers[0].ip
				}
			}
		} else {
			var serverSelected string
			serverOptions := make([]huh.Option[string], 0, len(procServers))
			for i, procServer := range procServers {
				serverOptions[i] = procServer.Option()
			}
			serverOptions = append(serverOptions, huh.Option[string]{
				Value: "Host a server",
				Key:   "host",
			})
			if err := huh.NewSelect[string]().
				Title("Select a server").
				Options(
					serverOptions...,
				).
				Value(&serverSelected).
				Run(); err == nil {
				if serverSelected != "host" {
					ip = serverSelected
				}
			}
		}
	} else {
		var confirm bool
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
		}
	}
	return
}

func (c *Config) StartServer(executable string, args []string, stop bool, canTrustCertificate bool) (errorCode int, ip string) {
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
	result, ip = server.StartServer(stopStr, serverExecutablePath, args, SelectBestServerIp)
	if result.Success() {
		printer.PrintSucceeded()
		if stop {
			c.SetServerExe(serverExecutablePath)
		}
	} else {
		printer.PrintFailedResultError(result)
		errorCode = internal.ErrServerStart
	}
	return
}

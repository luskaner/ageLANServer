package cmdUtils

import (
	"fmt"
	"io"
	"net"
	"net/netip"
	"os"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/luskaner/ageLANServer/common/uuid"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common"
	cmdServer "github.com/luskaner/ageLANServer/common/cmd/server"
	commonExecutor "github.com/luskaner/ageLANServer/common/executor/exec"
	commonLogger "github.com/luskaner/ageLANServer/common/logger"
	"github.com/luskaner/ageLANServer/launcher-common/launcher"
	"github.com/luskaner/ageLANServer/launcher-common/ui"
	"github.com/luskaner/ageLANServer/launcher/internal/cmdUtils/logger"
	"github.com/luskaner/ageLANServer/launcher/internal/dialog"
	"github.com/luskaner/ageLANServer/launcher/internal/server"
	"github.com/spf13/pflag"
)

type processedServer struct {
	server.MesuredIpAddress
	id          uuid.UUID
	description string
	// label is a compact version of description, without the alternative IPs and
	// hostnames, for dialogs with little horizontal room.
	label string
}

func processedServers(gameTitle string, servers map[uuid.UUID]*server.AnnounceMessage) []*processedServer {
	var processed []*processedServer
	for serverId, data := range servers {
		_, measuredIPs, internalData := server.FilterServerIPs(serverId, "", gameTitle, data.IpAddrs)
		if internalData == nil {
			continue
		}
		bestAddress := measuredIPs[0]
		var bestHostsSlice []string
		bestHosts := common.IpToHosts(bestAddress.Ip.String())
		var alternativeIpSlice []string
		var alternativeHostsSlice []string
		alternativeHosts := mapset.NewThreadUnsafeSet[string]()
		for _, alternativeAddress := range measuredIPs[1:] {
			alternativeHosts.Append(common.IpToHosts(alternativeAddress.Ip.String()).Difference(bestHosts).ToSlice()...)
			alternativeIpSlice = append(alternativeIpSlice, alternativeAddress.Ip.String())
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
		var sb strings.Builder
		sb.WriteString(bestAddress.Ip.String())
		if len(alternativeIpSlice) > 0 {
			sb.WriteString(", ")
			sb.WriteString(strings.Join(alternativeIpSlice, ", "))
		}
		if len(bestHostsSlice) > 0 || len(alternativeHostsSlice) > 0 {
			sb.WriteString(" (")
			if len(bestHostsSlice) > 0 {
				sb.WriteString(strings.Join(bestHostsSlice, ", "))
			}
			if len(alternativeHostsSlice) > 0 {
				if len(bestHostsSlice) > 0 {
					sb.WriteString(", ")
				}
				sb.WriteString(strings.Join(alternativeHostsSlice, ", "))
			}
			sb.WriteString(")")
		}
		latencyMs := bestAddress.Latency.Truncate(time.Millisecond).Milliseconds()
		_, _ = fmt.Fprintf(&sb, " - %d ms (%s)", latencyMs, internalData.Version)
		processed = append(processed, &processedServer{
			id:               serverId,
			MesuredIpAddress: bestAddress,
			description:      sb.String(),
			// Only what decides the choice, in an order that survives a narrow
			// list: the address that will be used, the latency, the version.
			label: fmt.Sprintf("%s - %d ms (%s)",
				bestAddress.Ip.String(), latencyMs, internalData.Version,
			),
		})
	}
	slices.SortStableFunc(processed, func(a, b *processedServer) int {
		return int(a.Latency - b.Latency)
	})
	return processed
}

func DiscoverServersAndSelectBestIpAddr(gameTitle string, singleAutoSelect bool, multicastGroups mapset.Set[netip.Addr], targetPorts mapset.Set[uint16]) (id uuid.UUID, ip net.IP) {
	id = uuid.Nil()
	servers := make(map[uuid.UUID]*server.AnnounceMessage)
	// The search takes a couple of seconds and prints nothing until it is over,
	// so it gets an in place line and the terminal's own progress indicator: the
	// last one is the only indicator still visible once the console has scrolled or
	// the window is in the background.
	search := ui.Start("Looking for servers...")
	bar := ui.BeginProgress()
	server.QueryServersWithProgress(multicastGroups, targetPorts, servers, func(round, rounds, found int) {
		// Percentages, not fractions: the indicator has one scale and it is 0 to
		// 100. A round of zero would divide by zero, and it never happens, but the
		// clamp is the difference between a bar and a panic.
		bar.Set(round * 100 / max(rounds, 1))
	})
	bar.Done()
	// The firewall hint is the one thing the reader can act on, so it is printed
	// only when nothing answered, where it is news, and not on every run.
	if len(servers) == 0 {
		search.Info("No servers found. You might need to allow the launcher in the firewall.")
	} else {
		// The numbered list that follows is the outcome; two lines saying the same
		// thing would be noise.
		search.Stop()
	}
	if len(servers) > 0 {
		if procServers := processedServers(gameTitle, servers); len(procServers) > 0 {
			idx, ok := selectDiscoveredServer(procServers, singleAutoSelect, os.Stdin)
			if i := usableServerIndex(idx, ok, len(procServers)); i >= 0 {
				selectedServer := procServers[i]
				ip = selectedServer.Ip
				id = selectedServer.id
			}
		}
	}
	return
}

// selectDiscoveredServer resolves which of the processed servers to use. It
// returns the 0-based index into procServers and false when the user declined
// to pick one, in which case the caller falls back to starting its own server.
func selectDiscoveredServer(procServers []*processedServer, singleAutoSelect bool, stdin io.Reader) (int, bool) {
	candidates := make([]dialog.ServerCandidate, len(procServers))
	for i, procServer := range procServers {
		candidates[i] = dialog.ServerCandidate{
			Description: procServer.description,
			Label:       procServer.label,
		}
	}
	if singleAutoSelect && len(procServers) == 1 {
		// Auto-selecting still lists the candidate first: that is what the
		// console has always done before this shortcut, and the backend that
		// would have rendered the list is not rendering anything now.
		dialog.Active().ListCandidates(candidates)
		// Left undecorated on purpose: select_server_test.go pins this line byte
		// for byte, and a marker there would buy nothing the numbered list above
		// does not already give.
		logger.Println("Auto-selecting the only found server.")
		return 0, true
	}
	return dialog.Active().SelectServer(candidates, stdin)
}

// usableServerIndex turns the dialog answer into a safe index into procServers.
// It returns -1 when the user declined, or when a backend reports an index that
// is out of range, so the caller starts its own server instead of indexing out
// of bounds.
func usableServerIndex(idx int, ok bool, procCount int) int {
	if !ok || idx < 0 || idx >= procCount {
		return -1
	}
	return idx
}

func (c *Config) StartServer(executable string, flags *pflag.FlagSet, values *cmdServer.Values, stop bool) (exitCode int, ip string) {
	if !launcher.CanUseInternet {
		values.CanUseInternet = false
	}
	logger.Step("Starting server, authorize it in firewall if needed...")
	var stopStr string
	if stop {
		stopStr = "true"
	} else {
		stopStr = "false"
	}
	var result *commonExecutor.Result
	var serverExe string
	result, serverExe, ip = server.StartServer(c.gameId, stopStr, executable, flags, values, func(options commonExecutor.Options) {
		commonLogger.Println("start server", options.String())
	})
	if result.Success() {
		logger.Ok("Server started.")
		if stop {
			c.serverExe = serverExe
		}
	} else {
		logger.Fail("Could not start server.")
		exitCode = launcher.ErrServerStart
		if result != nil {
			if result.Err != nil {
				logger.Fault("Error message: %s", result.Err.Error())
			}
			if result.ExitCode != common.ErrSuccess {
				logger.Fault("Exit code: %d.", result.ExitCode)
			}
		} else {
			logger.Detail("Try running the server manually.")
		}
	}
	return
}

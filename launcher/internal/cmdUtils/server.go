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
	"github.com/luskaner/ageLANServer/launcher/internal"
	"github.com/luskaner/ageLANServer/launcher/internal/cmdUtils/logger"
	"github.com/luskaner/ageLANServer/launcher/internal/server"
	"github.com/spf13/pflag"
)

type processedServer struct {
	server.MesuredIpAddress
	id          uuid.UUID
	description string
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
		_, _ = fmt.Fprintf(&sb, " - %d ms (%s)",
			bestAddress.Latency.Truncate(time.Millisecond).Milliseconds(),
			internalData.Version,
		)
		processed = append(processed, &processedServer{
			id:               serverId,
			MesuredIpAddress: bestAddress,
			description:      sb.String(),
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
	logger.Println("Looking for 'server's, you might need to allow the 'launcher' in the firewall...")
	server.QueryServers(multicastGroups, targetPorts, servers)
	if len(servers) > 0 {
		if procServers := processedServers(gameTitle, servers); len(procServers) > 0 {
			idx := selectServerIndex(len(procServers), singleAutoSelect, os.Stdin)
			if idx >= 0 {
				selectedServer := procServers[idx]
				ip = selectedServer.Ip
				id = selectedServer.id
			}
		}
	}
	return
}

// selectServerIndex asks the user to pick one of the discovered servers.
// Returns the 0-based index, or -1 when reading fails (e.g. stdin exhausted),
// in which case the caller should fall back to starting its own server.
func selectServerIndex(procCount int, singleAutoSelect bool, reader io.Reader) int {
	for {
		commonLogger.Printf("Found %d 'server'(s):\n", procCount)
		if singleAutoSelect && procCount == 1 {
			commonLogger.Println("Auto-selecting the only found 'server'.")
			return 0
		}
		commonLogger.Printf("Enter the number of the 'server' (1-%d): ", procCount)
		var option int
		if _, err := fmt.Fscan(reader, &option); err != nil {
			// Stdin exhausted or broken: we can never get a valid answer,
			// so retrying would spin forever printing the list.
			commonLogger.Println("Could not read selection from input.")
			return -1
		}
		if option < 1 || option > procCount {
			commonLogger.Println("Invalid option. Please enter a number from the list.")
			continue
		}
		return option - 1
	}
}

func (c *Config) StartServer(executable string, flags *pflag.FlagSet, values *cmdServer.Values, stop bool) (exitCode int, ip string) {
	logger.Println("Starting 'server', authorize it in firewall if needed...")
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
		logger.Println("'Server' started.")
		if stop {
			c.serverExe = serverExe
		}
	} else {
		logger.Println("Could not start 'server'.")
		exitCode = internal.ErrServerStart
		if result != nil {
			if result.Err != nil {
				logger.Println("Error message: " + result.Err.Error())
			}
			if result.ExitCode != common.ErrSuccess {
				logger.Printf(`Exit code: %d.`+"\n", result.ExitCode)
			}
		} else {
			logger.Println("Try running the 'server' manually.")
		}
	}
	return
}

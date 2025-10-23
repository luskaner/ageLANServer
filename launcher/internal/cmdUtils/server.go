package cmdUtils

import (
	"fmt"
	"net"
	"net/netip"
	"slices"
	"sort"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/luskaner/ageLANServer/common"
	commonExecutor "github.com/luskaner/ageLANServer/common/executor/exec"
	"github.com/luskaner/ageLANServer/launcher/internal"
	"github.com/luskaner/ageLANServer/launcher/internal/server"
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
		description := bestAddress.Ip.String()
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
				description += host
			}
			if len(alternativeHostsSlice) > 0 {
				if len(bestHostsSlice) > 0 {
					description += ", "
				}
				description += strings.Join(alternativeHostsSlice, ", ")
			}
			description += ")"
		}
		description += fmt.Sprintf(" - %d ms", bestAddress.Latency.Truncate(time.Millisecond).Milliseconds())
		description += fmt.Sprintf(" (%s)", internalData.Version)
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

func DiscoverServersAndSelectBestIpAddr(gameTitle string, multicastGroups mapset.Set[netip.Addr], targetPorts mapset.Set[uint16]) (id uuid.UUID, ip net.IP) {
	id = uuid.Nil
	servers := make(map[uuid.UUID]*server.AnnounceMessage)
	fmt.Println("Searching for 'server's, you might need to allow the 'launcher' in the firewall...")
	server.QueryServers(multicastGroups, targetPorts, servers)
	if len(servers) > 0 {
		if procServers := processedServers(gameTitle, servers); len(procServers) > 0 {
			var option int
			for {
				fmt.Println("Found the following 'server's:")
				for i := range procServers {
					fmt.Printf("%d. %s\n", i+1, procServers[i].description)
				}
				fmt.Printf("Enter the number of the 'server' (1-%d): ", len(procServers))
				_, err := fmt.Scan(&option)
				if err != nil || option < 1 || option > len(procServers) {
					fmt.Println("Invalid (or error reading) option. Please enter a number from the list.")
					continue
				}
				selectedServer := procServers[option-1]
				ip = selectedServer.Ip
				id = selectedServer.id
				break
			}
		}
	}
	return
}

func (c *Config) StartServer(executable string, args []string, stop bool, serverId uuid.UUID) (errorCode int, ip string) {
	fmt.Println("Starting 'server', authorize it in firewall if needed...")
	var stopStr string
	if stop {
		stopStr = "true"
	} else {
		stopStr = "false"
	}
	var result *commonExecutor.Result
	var serverExe string
	result, serverExe, ip = server.StartServer(c.gameId, stopStr, executable, args, serverId, func(options commonExecutor.Options) {
		LogPrintln("start server", options.String())
	})
	if result.Success() {
		fmt.Println("'Server' started.")
		if stop {
			c.serverExe = serverExe
		}
	} else {
		fmt.Println("Could not start 'server'.")
		errorCode = internal.ErrServerStart
		if result != nil {
			if result.Err != nil {
				fmt.Println("Error message: " + result.Err.Error())
			}
			if result.ExitCode != common.ErrSuccess {
				fmt.Printf(`Exit code: %d.`+"\n", result.ExitCode)
			}
		} else {
			fmt.Println("Try running the 'server' manually.")
		}
	}
	return
}

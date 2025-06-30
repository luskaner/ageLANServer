package game

import (
	battleServerBroadcast "github.com/luskaner/ageLANServer/battle-server-broadcast"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/config/shared"
	"net"
	"slices"
)

func RebroadcastIPs(filterIPs []net.IP, filterInterfaces []shared.Interface) (IPs []net.IP) {
	mostPriority, restInterfaces, err := battleServerBroadcast.RetrieveBsInterfaceAddresses()
	if err != nil || mostPriority == nil || len(restInterfaces) == 0 {
		return
	}
	var availableInterfaces map[*net.Interface][]*net.IPNet
	availableInterfaces, err = common.IPv4RunningNetworkInterfaces()
	if err != nil {
		return
	}
	for iff, nets := range availableInterfaces {
		// Remove interfaces that are not in restInterfaces
		availableInterfaces[iff] = slices.DeleteFunc(
			nets,
			func(n *net.IPNet) bool {
				return !slices.Equal(nets, restInterfaces)
			},
		)
		if len(availableInterfaces[iff]) == 0 {
			delete(availableInterfaces, iff)
		}
	}
	return shared.FilterNetworks(availableInterfaces, filterIPs, filterInterfaces, false)
}

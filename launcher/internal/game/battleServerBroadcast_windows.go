package game

import (
	mapset "github.com/deckarep/golang-set/v2"
	battleServerBroadcast "github.com/luskaner/ageLANServer/battle-server-broadcast"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/config/shared"
	"net"
	"net/netip"
	"slices"
)

func RebroadcastIPAddrs(filterIPAddrs mapset.Set[netip.Addr], filterInterfaces mapset.Set[shared.Interface]) (IPAddrs mapset.Set[netip.Addr]) {
	mostPriority, restInterfaces, err := battleServerBroadcast.RetrieveBsInterfaceAddresses()
	if err != nil || mostPriority == nil || len(restInterfaces) == 0 {
		return
	}
	var availableInterfaces map[*net.Interface][]*net.IPNet
	availableInterfaces, err = common.RunningNetworkInterfaces(true, false, false)
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
	if !filterIPAddrs.IsEmpty() {
		IPAddrs = mapset.NewThreadUnsafeSet[netip.Addr]()
		for _, nets := range availableInterfaces {
			for _, n := range nets {
				for filterIpAddr := range filterIPAddrs.Iter() {
					if ipAddr := common.NetIPToNetIPAddr(n.IP); filterIpAddr == ipAddr {
						IPAddrs.Add(ipAddr)
					}
				}
			}
		}
	} else if !filterInterfaces.IsEmpty() {
		return shared.FilterNetworks(
			availableInterfaces,
			filterInterfaces,
			true,
			false,
			false,
		)
	}
	return
}
